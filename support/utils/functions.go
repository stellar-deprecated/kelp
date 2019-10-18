package utils

import (
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/stellar/go/clients/horizonclient"
	"github.com/stellar/go/keypair"
	"github.com/stellar/go/network"
	hProtocol "github.com/stellar/go/protocols/horizon"
	"github.com/stellar/go/protocols/horizon/base"
	"github.com/stellar/go/txnbuild"
)

// Common Utilities needed by various bots

// Native is the string representing the type for the native lumen asset
const Native = "native"

// NativeAsset represents the native asset
var NativeAsset = hProtocol.Asset{Type: Native}

// SdexPrecision defines the number of decimals used in SDEX
const SdexPrecision int8 = 7

// ByPrice implements sort.Interface for []horizon.Offer based on the price
type ByPrice []hProtocol.Offer

func (a ByPrice) Len() int      { return len(a) }
func (a ByPrice) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ByPrice) Less(i, j int) bool {
	return PriceAsFloat(a[i].Price) < PriceAsFloat(a[j].Price)
}

// PriceAsFloat converts a string price to a float price
func PriceAsFloat(price string) float64 {
	p, err := strconv.ParseFloat(price, 64)
	if err != nil {
		log.Printf("Error parsing price: %s | %s\n", price, err)
		return 0
	}
	return p
}

// AmountStringAsFloat converts a string amount to a float amount
func AmountStringAsFloat(amount string) float64 {
	if amount == "" {
		return 0
	}
	p, err := strconv.ParseFloat(amount, 64)
	if err != nil {
		log.Printf("Error parsing amount: %s | %s\n", amount, err)
		return 0
	}
	return p
}

// ParseOfferAmount is a convenience method to parse an offer amount
func ParseOfferAmount(amt string) (float64, error) {
	offerAmt, e := strconv.ParseFloat(amt, 64)
	if e != nil {
		log.Printf("error parsing offer amount: %s\n", e)
		return -1, e
	}
	return offerAmt, nil
}

// GetPrice gets the price from an offer
func GetPrice(offer hProtocol.Offer) float64 {
	if int64(offer.PriceR.D) == 0 {
		return 0.0
	}
	return PriceAsFloat(big.NewRat(int64(offer.PriceR.N), int64(offer.PriceR.D)).FloatString(10))
}

// GetInvertedPrice gets the inverted price from an offer
func GetInvertedPrice(offer hProtocol.Offer) float64 {
	if int64(offer.PriceR.N) == 0 {
		return 0.0
	}
	return PriceAsFloat(big.NewRat(int64(offer.PriceR.D), int64(offer.PriceR.N)).FloatString(10))
}

// Asset2Asset converts a horizon.Asset to a txnbuild.Asset.
func Asset2Asset(Asset hProtocol.Asset) txnbuild.Asset {
	if Asset.Type == Native {
		return txnbuild.NativeAsset{}
	}
	return txnbuild.CreditAsset{Code: Asset.Code, Issuer: Asset.Issuer}
}

// Asset2Asset2 converts a txnbuild.Asset to a horizon.Asset.
func Asset2Asset2(Asset txnbuild.Asset) hProtocol.Asset {
	a := hProtocol.Asset{}

	a.Code = Asset.GetCode()
	a.Issuer = Asset.GetIssuer()
	if Asset.IsNative() {
		a.Type = Native
	} else if len(a.Code) > 4 {
		a.Type = "credit_alphanum12"
	} else {
		a.Type = "credit_alphanum4"
	}
	return a
}

// Asset2String converts a horizon.Asset to a string representation, using "native" for the native XLM
func Asset2String(asset hProtocol.Asset) string {
	if asset.Type == Native {
		return Native
	}
	return fmt.Sprintf("%s:%s", asset.Code, asset.Issuer)
}

// Asset2CodeString extracts the code out of a horizon.Asset
func Asset2CodeString(asset hProtocol.Asset) string {
	if asset.Type == Native {
		return "XLM"
	}
	return asset.Code
}

// String2Asset converts a code:issuer to a horizon.Asset
func String2Asset(code string, issuer string) hProtocol.Asset {
	if code == "XLM" {
		return Asset2Asset2(txnbuild.NativeAsset{})
	}
	return Asset2Asset2(txnbuild.CreditAsset{Code: code, Issuer: issuer})
}

// LoadAllOffers loads all the offers for a given account
func LoadAllOffers(account string, api *horizonclient.Client) ([]hProtocol.Offer, error) {
	// get what orders are outstanding now
	offerReq := horizonclient.OfferRequest{
		ForAccount: account,
		Limit:      uint(200),
	}

	offersPage, e := api.Offers(offerReq)
	if e != nil {
		return []hProtocol.Offer{}, fmt.Errorf("Can't load offers: %s\n", e)
	}

	offersRet := offersPage.Embedded.Records
	for len(offersPage.Embedded.Records) > 0 {
		offersPage, e = api.NextOffersPage(offersPage)
		if e != nil {
			return []hProtocol.Offer{}, fmt.Errorf("Can't load offers: %s\n", e)
		}
		offersRet = append(offersRet, offersPage.Embedded.Records...)
	}

	return offersRet, nil
}

// FilterOffers filters out the offers into selling and buying, where sellOffers sells the sellAsset and buyOffers buys the sellAsset
func FilterOffers(offers []hProtocol.Offer, sellAsset hProtocol.Asset, buyAsset hProtocol.Asset) (sellOffers []hProtocol.Offer, buyOffers []hProtocol.Offer) {
	for _, offer := range offers {
		if offer.Selling == sellAsset {
			if offer.Buying == buyAsset {
				sellOffers = append(sellOffers, offer)
			}
		} else if offer.Selling == buyAsset {
			if offer.Buying == sellAsset {
				buyOffers = append(buyOffers, offer)
			}
		}
	}
	return
}

// ParseSecret returns the address from the secret
func ParseSecret(secret string) (*string, error) {
	if secret == "" {
		return nil, nil
	}

	sourceKP, err := keypair.Parse(secret)
	if err != nil {
		return nil, err
	}

	address := sourceKP.Address()
	return &address, nil
}

// ParseNetwork checks the horizon url and returns the test network if it contains "test"
func ParseNetwork(horizonURL string) string {
	if strings.Contains(horizonURL, "test") {
		return network.TestNetworkPassphrase
	}
	return network.PublicNetworkPassphrase
}

// GetJSON is a helper method to get json from a URL
func GetJSON(client http.Client, url string, target interface{}) error {
	r, err := client.Get(url)
	if err != nil {
		return err
	}
	defer r.Body.Close()

	return json.NewDecoder(r.Body).Decode(target)
}

// GetCreditBalance is a drop-in for the function in the GoSDK, we want it to return nil if there's no balance (as opposed to "0")
func GetCreditBalance(a hProtocol.Account, code string, issuer string) *string {
	for _, balance := range a.Balances {
		if balance.Asset.Code == code && balance.Asset.Issuer == issuer {
			return &balance.Balance
		}
	}
	return nil
}

// AssetsEqual is a convenience method to compare horizon.Asset and base.Asset because they are not type aliased
func AssetsEqual(baseAsset base.Asset, horizonAsset hProtocol.Asset) bool {
	return horizonAsset.Type == baseAsset.Type &&
		horizonAsset.Code == baseAsset.Code &&
		horizonAsset.Issuer == baseAsset.Issuer
}

// CheckFetchFloat tries to fetch and then cast the value for the provided key
func CheckFetchFloat(m map[string]interface{}, key string) (float64, error) {
	v, ok := m[key]
	if !ok {
		return 0.0, fmt.Errorf("'%s' field not in map: %v", key, m)
	}

	f, ok := v.(float64)
	if !ok {
		return 0.0, fmt.Errorf("unable to cast '%s' field to float64, value: %v", key, v)
	}

	return f, nil
}

// CheckedString returns "<nil>" if the object is nil, otherwise calls the String() function on the object
func CheckedString(v interface{}) string {
	if v == nil {
		return "<nil>"
	}
	return fmt.Sprintf("%v", v)
}

// ParseAsset returns a horizon asset a string
func ParseAsset(code string, issuer string) (*hProtocol.Asset, error) {
	if code != "XLM" && issuer == "" {
		return nil, fmt.Errorf("error: issuer can only be empty if asset is XLM")
	}

	if code == "XLM" && issuer != "" {
		return nil, fmt.Errorf("error: issuer needs to be empty if asset is XLM")
	}

	if code == "XLM" {
		asset := Asset2Asset2(txnbuild.NativeAsset{})
		return &asset, nil
	}

	asset := Asset2Asset2(txnbuild.CreditAsset{Code: code, Issuer: issuer})
	return &asset, nil
}

// AssetOnlyCodeEquals only checks the type and code of these assets, i.e. insensitive to asset issuer
func AssetOnlyCodeEquals(hAsset hProtocol.Asset, txnAsset txnbuild.Asset) (bool, error) {
	if txnAsset.IsNative() {
		return hAsset.Type == Native, nil
	} else if hAsset.Type == Native {
		return false, nil
	}

	return txnAsset.GetCode() == hAsset.Code, nil
}

// assetEqualsExact does an exact comparison of two assets
func assetEqualsExact(hAsset hProtocol.Asset, xAsset txnbuild.Asset) (bool, error) {
	if xAsset.IsNative() {
		return hAsset.Type == Native, nil
	} else if hAsset.Type == Native {
		return false, nil
	}

	return xAsset.GetCode() == hAsset.Code && xAsset.GetIssuer() == hAsset.Issuer, nil
}

// IsSelling helper method
func IsSelling(sdexBase hProtocol.Asset, sdexQuote hProtocol.Asset, selling txnbuild.Asset, buying txnbuild.Asset) (bool, error) {
	sellingBase, e := assetEqualsExact(sdexBase, selling)
	if e != nil {
		return false, fmt.Errorf("error comparing sdexBase with selling asset")
	}
	buyingQuote, e := assetEqualsExact(sdexQuote, buying)
	if e != nil {
		return false, fmt.Errorf("error comparing sdexQuote with buying asset")
	}
	if sellingBase && buyingQuote {
		return true, nil
	}

	sellingQuote, e := assetEqualsExact(sdexQuote, selling)
	if e != nil {
		return false, fmt.Errorf("error comparing sdexQuote with selling asset")
	}
	buyingBase, e := assetEqualsExact(sdexBase, buying)
	if e != nil {
		return false, fmt.Errorf("error comparing sdexBase with buying asset")
	}
	if sellingQuote && buyingBase {
		return false, nil
	}

	return false, fmt.Errorf("invalid assets, there are more than 2 distinct assets: sdexBase=%s, sdexQuote=%s, selling=%s, buying=%s", sdexBase, sdexQuote, selling, buying)
}

// Shuffle any string slice
func Shuffle(slice []string) {
	r := rand.New(rand.NewSource(time.Now().Unix()))
	for n := len(slice); n > 0; n-- {
		randIndex := r.Intn(n)
		slice[n-1], slice[randIndex] = slice[randIndex], slice[n-1]
	}
}

// SignWithSeed modifies the passed in tx with the signatures of the passed in seeds
func SignWithSeed(tx *txnbuild.Transaction, seeds ...string) error {
	for i, s := range seeds {
		kp, e := keypair.Parse(s)
		if e != nil {
			return fmt.Errorf("cannot parse seed into keypair at index %d: %s", i, e)
		}

		e = tx.Sign(kp.(*keypair.Full))
		if e != nil {
			return fmt.Errorf("cannot sign tx with keypair at index %d (pubKey: %s): %s", i, kp.Address(), e)
		}
	}

	return nil
}

// StringSet converts a string slice to a map of string to bool values to represent a Set
func StringSet(list []string) map[string]bool {
	m := map[string]bool{}
	for _, s := range list {
		m[s] = true
	}
	return m
}
