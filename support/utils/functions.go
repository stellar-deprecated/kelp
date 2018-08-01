package utils

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"math/big"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"github.com/stellar/go/build"
	"github.com/stellar/go/clients/horizon"
	"github.com/stellar/go/keypair"
	"github.com/stellar/go/support/config"
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

// FloatEquals returns true if the two floats are equal within the epsilon of error (avoids mismatched equality because of floating point noise)
func FloatEquals(f1 float64, f2 float64, epsilon float64) bool {
	return math.Abs(f1-f2) < epsilon
}

// GetPrice gets the price from an offer
func GetPrice(offer horizon.Offer) float64 {
	if int64(offer.PriceR.D) == 0 {
		return 0.0
	}
	return PriceAsFloat(big.NewRat(int64(offer.PriceR.N), int64(offer.PriceR.D)).FloatString(10))
}

// GetInvertedPrice gets the inverted price from an offer
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

// Asset2Asset2 converts a build.Asset to a horizon.Asset
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

// String2Asset converts a code:issuer to a horizon.Asset
func String2Asset(code string, issuer string) horizon.Asset {
	if code == "XLM" {
		return Asset2Asset2(build.NativeAsset())
	}
	return Asset2Asset2(build.CreditAsset(code, issuer))
}

// LoadAllOffers loads all the offers for a given account
func LoadAllOffers(account string, api *horizon.Client) (offersRet []horizon.Offer, err error) {
	// get what orders are outstanding now
	offersPage, err := api.LoadAccountOffers(account)
	if err != nil {
		log.Printf("Can't load offers: %s\n", err)
		return
	}
	offersRet = offersPage.Embedded.Records
	for len(offersPage.Embedded.Records) > 0 {
		offersPage, err = api.LoadAccountOffers(account, horizon.At(offersPage.Links.Next.Href))
		if err != nil {
			log.Printf("Can't load offers: %s\n", err)
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
			log.Fatalf("config error: %v\n", cause)
		default:
			log.Fatal(e)
		}
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

// GetJSON is a helper method to get json from a URL
func GetJSON(client http.Client, url string, target interface{}) error {
	r, err := client.Get(url)
	if err != nil {
		return err
	}
	defer r.Body.Close()

	return json.NewDecoder(r.Body).Decode(target)
}

// GetSortedKeys gets the keys of the map after sorting
func GetSortedKeys(m map[string]string) []string {
	keys := []string{}
	for name := range m {
		keys = append(keys, name)
	}
	sort.Strings(keys)
	return keys
}
