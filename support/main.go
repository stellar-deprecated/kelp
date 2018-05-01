package kelp

import (
	"fmt"
	"math"
	"math/big"
	"os"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"github.com/stellar/go/build"
	"github.com/stellar/go/clients/horizon"
	"github.com/stellar/go/keypair"
	"github.com/stellar/go/support/config"
	"github.com/stellar/go/support/log"
)

// Common Utilities needed by various bots

// Native is the string representing the type for the native lumen asset
const Native = "native"

// ByPrice implements sort.Interface for []horizon.Offer based on the price
type ByPrice []horizon.Offer

func (a ByPrice) Len() int      { return len(a) }
func (a ByPrice) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ByPrice) Less(i, j int) bool {
	return PriceAsFloat(a[i].Price) < PriceAsFloat(a[j].Price)
}

func PriceAsFloat(price string) float64 {
	p, err := strconv.ParseFloat(price, 64)
	if err != nil {
		log.Error("Error parsing price", price)
		return 0
	}
	return p
}

func AmountStringAsFloat(amount string) float64 {
	if amount == "" {
		return 0
	}
	p, err := strconv.ParseFloat(amount, 64)
	if err != nil {
		log.Error("Error parsing amount", amount)
		return 0
	}
	return p
}

func FloatEquals(f1 float64, f2 float64, epsilon float64) bool {
	return math.Abs(f1-f2) < epsilon
}

func GetPrice(offer horizon.Offer) float64 {
	if int64(offer.PriceR.D) == 0 {
		return 0.0
	}
	return PriceAsFloat(big.NewRat(int64(offer.PriceR.N), int64(offer.PriceR.D)).FloatString(10))
}

func GetInvertedPrice(offer horizon.Offer) float64 {
	if int64(offer.PriceR.N) == 0 {
		return 0.0
	}
	return PriceAsFloat(big.NewRat(int64(offer.PriceR.D), int64(offer.PriceR.N)).FloatString(10))
}

// Asset2Asset is a Boyz2Men cover band on the blockchain
func Asset2Asset(Asset horizon.Asset) build.Asset {
	a := build.Asset{}

	a.Code = Asset.Code
	a.Issuer = Asset.Issuer
	if Asset.Type == Native {
		a.Native = true
	}
	return a
}

func Asset2Asset2(Asset build.Asset) horizon.Asset {
	a := horizon.Asset{}

	a.Code = Asset.Code
	a.Issuer = Asset.Issuer
	if Asset.Native {
		a.Type = Native
	} else if len(a.Code) > 4 {
		a.Type = "credit_alphanum12"
	} else {
		a.Type = "credit_alphanum4"
	}
	return a
}

func String2Asset(code string, issuer string) horizon.Asset {
	if code == "XLM" {
		return Asset2Asset2(build.NativeAsset())
	} else {
		return Asset2Asset2(build.CreditAsset(code, issuer))
	}
}

func LoadAllOffers(account string, api *horizon.Client) (offersRet []horizon.Offer, err error) {
	// get what orders are outstanding now
	offersPage, err := api.LoadAccountOffers(account)
	if err != nil {
		log.Error("Can't load offers ", err)
		return
	}
	offersRet = offersPage.Embedded.Records
	for len(offersPage.Embedded.Records) > 0 {
		offersPage, err = api.LoadAccountOffers(account, horizon.At(offersPage.Links.Next.Href))
		if err != nil {
			log.Error("Can't load offers ", err)
			return
		}
		offersRet = append(offersRet, offersPage.Embedded.Records...)
	}
	return
}

// FilterOffers filters out the offers into selling and buying, where sellOffers sells the sellAsset and buyOffers buys the sellAsset
func FilterOffers(offers []horizon.Offer, sellAsset horizon.Asset, buyAsset horizon.Asset) (sellOffers []horizon.Offer, buyOffers []horizon.Offer) {
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

// CheckConfigError checks configs for errors
func CheckConfigError(cfg interface{}, e error) {
	fmt.Printf("Result: %+v\n", cfg)

	if e != nil {
		switch cause := errors.Cause(e).(type) {
		case *config.InvalidConfigError:
			log.Error("config file: ", cause)
		default:
			log.Error(e)
		}
		os.Exit(1)
	}
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
func ParseNetwork(horizonURL string) build.Network {
	if strings.Contains(horizonURL, "test") {
		return build.TestNetwork
	}
	return build.PublicNetwork
}
