package plugins

import (
	"fmt"
	"log"
	"strings"

	hProtocol "github.com/stellar/go/protocols/horizon"
	"github.com/stellar/kelp/api"
	"github.com/stellar/kelp/support/utils"
)

// Liabilities represents the "committed" units of an asset on both the buy and sell sides
type Liabilities struct {
	Buying  float64 // affects how much more can be bought
	Selling float64 // affects how much more can be sold
}

// IEIF is the module that allows us to ensure that orders are always "Immediately Executable In Full"
type IEIF struct {
	// explicitly calculate liabilities here for now, we can switch over to using the values from Horizon once the protocol change has taken effect
	cachedLiabilities map[hProtocol.Asset]Liabilities

	// TODO 2 streamline requests instead of caching
	// cache balances to avoid redundant requests
	cachedBalances map[hProtocol.Asset]api.Balance

	isTradingSdex bool

	// TODO this is a hack because the logic to fetch balances is in the exchange, maybe take in an api.Account interface
	// TODO this is a hack because the logic to fetch offers is in the exchange, maybe take in api.GetOpenOrders() as an interface
	// TODO 1 this should not be horizon specific
	exchangeShim api.ExchangeShim
}

// SetExchangeShim is a hack, TODO remove this hack
func (ieif *IEIF) SetExchangeShim(exchangeShim api.ExchangeShim) {
	ieif.exchangeShim = exchangeShim
}

// MakeIEIF factory method
func MakeIEIF(isTradingSdex bool) *IEIF {
	return &IEIF{
		cachedLiabilities: map[hProtocol.Asset]Liabilities{},
		cachedBalances:    map[hProtocol.Asset]api.Balance{},
		isTradingSdex:     isTradingSdex,
	}
}

// AddLiabilities updates the cached liabilities, units are in their respective assets
func (ieif *IEIF) AddLiabilities(selling hProtocol.Asset, buying hProtocol.Asset, incrementalSell float64, incrementalBuy float64, incrementalNativeAmountRaw float64) {
	ieif.cachedLiabilities[selling] = Liabilities{
		Selling: ieif.cachedLiabilities[selling].Selling + incrementalSell,
		Buying:  ieif.cachedLiabilities[selling].Buying,
	}
	ieif.cachedLiabilities[buying] = Liabilities{
		Selling: ieif.cachedLiabilities[buying].Selling,
		Buying:  ieif.cachedLiabilities[buying].Buying + incrementalBuy,
	}
	ieif.cachedLiabilities[utils.NativeAsset] = Liabilities{
		Selling: ieif.cachedLiabilities[utils.NativeAsset].Selling + incrementalNativeAmountRaw,
		Buying:  ieif.cachedLiabilities[utils.NativeAsset].Buying,
	}
}

// RecomputeAndLogCachedLiabilities clears the cached liabilities and recomputes from the network before logging
func (ieif *IEIF) RecomputeAndLogCachedLiabilities(assetBase hProtocol.Asset, assetQuote hProtocol.Asset) {
	ieif.cachedLiabilities = map[hProtocol.Asset]Liabilities{}
	// reset cached balances too so we fetch fresh balances
	ieif.ResetCachedBalances()
	ieif.LogAllLiabilities(assetBase, assetQuote)
}

// ResetCachedLiabilities resets the cache to include only the two assets passed in
func (ieif *IEIF) ResetCachedLiabilities(assetBase hProtocol.Asset, assetQuote hProtocol.Asset) error {
	ieif.cachedLiabilities = map[hProtocol.Asset]Liabilities{}

	// re-compute the liabilities
	offers, e := ieif.exchangeShim.LoadOffersHack()
	if e != nil {
		return fmt.Errorf("cannot load offers when trying to reset cached liabilities: %s", e)
	}
	baseLiabilities, basePairLiabilities, e := ieif.pairLiabilities(offers, assetBase, assetQuote)
	if e != nil {
		return fmt.Errorf("could not get pairLiabilities for base asset: %s", e)
	}
	quoteLiabilities, quotePairLiabilities, e := ieif.pairLiabilities(offers, assetQuote, assetBase)
	if e != nil {
		return fmt.Errorf("could not get pairLiabilities for quote asset: %s", e)
	}

	// delete liability amounts related to all offers (filter on only those offers involving **both** assets in case the account is used by multiple bots)
	ieif.cachedLiabilities[assetBase] = Liabilities{
		Buying:  baseLiabilities.Buying - basePairLiabilities.Buying,
		Selling: baseLiabilities.Selling - basePairLiabilities.Selling,
	}
	ieif.cachedLiabilities[assetQuote] = Liabilities{
		Buying:  quoteLiabilities.Buying - quotePairLiabilities.Buying,
		Selling: quoteLiabilities.Selling - quotePairLiabilities.Selling,
	}
	return nil
}

// willOversellNative returns willOversellNative, error
func (ieif *IEIF) willOversellNative(incrementalNativeAmount float64) (bool, error) {
	nativeBalance, e := ieif.assetBalance(utils.NativeAsset)
	if e != nil {
		return false, e
	}
	// TODO don't break out into vars
	nativeBal, _, minAccountBal := nativeBalance.Balance, nativeBalance.Trust, nativeBalance.Reserve
	nativeLiabilities, e := ieif.assetLiabilities(utils.NativeAsset)
	if e != nil {
		return false, e
	}

	willOversellNative := incrementalNativeAmount > (nativeBal - minAccountBal - nativeLiabilities.Selling)
	if willOversellNative {
		log.Printf("we will oversell the native asset after considering fee and min reserves, incrementalNativeAmount = %.8f, nativeBal = %.8f, minAccountBal = %.8f, nativeLiabilities.Selling = %.8f\n",
			incrementalNativeAmount, nativeBal, minAccountBal, nativeLiabilities.Selling)
	}
	return willOversellNative, nil
}

// willOversell returns willOversell, error
func (ieif *IEIF) willOversell(asset hProtocol.Asset, amountSelling float64) (bool, error) {
	balance, e := ieif.assetBalance(asset)
	if e != nil {
		return false, e
	}
	// TODO don't break out into vars
	bal, _, minAccountBal := balance.Balance, balance.Trust, balance.Reserve
	liabilities, e := ieif.assetLiabilities(asset)
	if e != nil {
		return false, e
	}

	willOversell := amountSelling > (bal - minAccountBal - liabilities.Selling)
	if willOversell {
		log.Printf("we will oversell the asset '%s', amountSelling = %.8f, bal = %.8f, minAccountBal = %.8f, liabilities.Selling = %.8f\n",
			utils.Asset2String(asset), amountSelling, bal, minAccountBal, liabilities.Selling)
	}
	return willOversell, nil
}

// willOverbuy returns willOverbuy, error
func (ieif *IEIF) willOverbuy(asset hProtocol.Asset, amountBuying float64) (bool, error) {
	if asset.Type == utils.Native {
		// you can never overbuy the native asset
		return false, nil
	}

	balance, e := ieif.assetBalance(asset)
	if e != nil {
		return false, e
	}
	liabilities, e := ieif.assetLiabilities(asset)
	if e != nil {
		return false, e
	}

	willOverbuy := amountBuying > (balance.Trust - liabilities.Buying)
	return willOverbuy, nil
}

// LogAllLiabilities logs the liabilities for the two assets along with the native asset
func (ieif *IEIF) LogAllLiabilities(assetBase hProtocol.Asset, assetQuote hProtocol.Asset) {
	ieif.logLiabilities(assetBase, "base  ")
	ieif.logLiabilities(assetQuote, "quote ")

	if ieif.isTradingSdex && assetBase != utils.NativeAsset && assetQuote != utils.NativeAsset {
		ieif.logLiabilities(utils.NativeAsset, "native")
	}
}

func (ieif *IEIF) logLiabilities(asset hProtocol.Asset, assetStr string) {
	trimmedAssetStr := strings.TrimSpace(assetStr)
	l, e := ieif.assetLiabilities(asset)
	if e != nil {
		log.Printf("could not fetch liability for %s asset, error = %s\n", trimmedAssetStr, e)
		return
	}

	balance, e := ieif.assetBalance(asset)
	if e != nil {
		log.Printf("cannot fetch balance for %s asset, error = %s\n", trimmedAssetStr, e)
		return
	}
	// TODO don't break out into vars
	bal, trust, minAccountBal := balance.Balance, balance.Trust, balance.Reserve

	trustString := "math.MaxFloat64"
	if trust != maxLumenTrust {
		trustString = fmt.Sprintf("%.8f", trust)
	}
	log.Printf("asset=%s, balance=%.8f, trust=%s, minAccountBal=%.8f, buyingLiabilities=%.8f, sellingLiabilities=%.8f\n",
		assetStr, bal, trustString, minAccountBal, l.Buying, l.Selling)
}

// AvailableCapacity returns the buying and selling amounts available for a given asset
func (ieif *IEIF) AvailableCapacity(asset hProtocol.Asset, incrementalNativeAmountRaw float64) (*Liabilities, error) {
	l, e := ieif.assetLiabilities(asset)
	if e != nil {
		return nil, e
	}

	balance, e := ieif.assetBalance(asset)
	if e != nil {
		return nil, e
	}
	// TODO don't break out into vars
	bal, trust, minAccountBal := balance.Balance, balance.Trust, balance.Reserve

	// factor in cost of increase in minReserve and fee when calculating selling capacity of native asset
	incrementalSellingLiability := 0.0
	if asset == utils.NativeAsset {
		incrementalSellingLiability = incrementalNativeAmountRaw
	}

	return &Liabilities{
		Buying:  trust - l.Buying,
		Selling: bal - minAccountBal - l.Selling - incrementalSellingLiability,
	}, nil
}

// assetLiabilities returns the liabilities for the asset
func (ieif *IEIF) assetLiabilities(asset hProtocol.Asset) (*Liabilities, error) {
	if v, ok := ieif.cachedLiabilities[asset]; ok {
		return &v, nil
	}

	offers, e := ieif.exchangeShim.LoadOffersHack()
	if e != nil {
		assetString := utils.Asset2String(asset)
		return nil, fmt.Errorf("cannot load offers to compute liabilities for asset (%s): %s", assetString, e)
	}

	assetLiabilities, _, e := ieif._liabilities(offers, asset, asset) // pass in the same asset, we ignore the returned object anyway
	return assetLiabilities, e
}

// pairLiabilities returns the liabilities for the asset along with the pairLiabilities
func (ieif *IEIF) pairLiabilities(offers []hProtocol.Offer, asset hProtocol.Asset, otherAsset hProtocol.Asset) (*Liabilities, *Liabilities, error) {
	assetLiabilities, pairLiabilities, e := ieif._liabilities(offers, asset, otherAsset)
	return assetLiabilities, pairLiabilities, e
}

// liabilities returns the asset liabilities and pairLiabilities (non-nil only if the other asset is specified)
func (ieif *IEIF) _liabilities(offers []hProtocol.Offer, asset hProtocol.Asset, otherAsset hProtocol.Asset) (*Liabilities, *Liabilities, error) {
	// uses all offers for this trading account to accommodate sharing by other bots

	// liabilities for the asset
	liabilities := Liabilities{}
	// liabilities for the asset w.r.t. the trading pair
	pairLiabilities := Liabilities{}
	for _, offer := range offers {
		if offer.Selling == asset {
			offerAmt, e := utils.ParseOfferAmount(offer.Amount)
			if e != nil {
				return nil, nil, fmt.Errorf("unable to parse offer amount '%s': %s", offer.Amount, e)
			}
			liabilities.Selling += offerAmt

			if offer.Buying == otherAsset {
				pairLiabilities.Selling += offerAmt
			}
		} else if offer.Buying == asset {
			offerAmt, e := utils.ParseOfferAmount(offer.Amount)
			if e != nil {
				return nil, nil, fmt.Errorf("unable to parse offer amount '%s': %s", offer.Amount, e)
			}
			offerPrice, e := utils.ParseOfferAmount(offer.Price)
			if e != nil {
				return nil, nil, fmt.Errorf("unable to parse offer price '%s': %s", offer.Price, e)
			}
			buyingAmount := offerAmt * offerPrice
			liabilities.Buying += buyingAmount

			if offer.Selling == otherAsset {
				pairLiabilities.Buying += buyingAmount
			}
		}
	}

	ieif.cachedLiabilities[asset] = liabilities
	return &liabilities, &pairLiabilities, nil
}

// ResetCachedBalances resets the cached balances map
func (ieif *IEIF) ResetCachedBalances() {
	ieif.cachedBalances = map[hProtocol.Asset]api.Balance{}
}

// GetAssetBalance is the exported version of assetBalance
func (ieif *IEIF) GetAssetBalance(asset hProtocol.Asset) (*api.Balance, error) {
	return ieif.assetBalance(asset)
}

// assetBalance is a memoized version of submitX.
func (ieif *IEIF) assetBalance(asset hProtocol.Asset) (*api.Balance, error) {
	if v, ok := ieif.cachedBalances[asset]; ok {
		return &v, nil
	}

	balance, e := ieif.exchangeShim.GetBalanceHack(asset)
	if e == nil {
		ieif.cachedBalances[asset] = *balance
	}

	return balance, e
}
