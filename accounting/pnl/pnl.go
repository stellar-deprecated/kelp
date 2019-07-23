/*
Computes the current balance valued in XLM for the given accounts

Makes some assumptions (only works for balanced bot accounts):
1. only 2 assets per account
2. one asset is always native XLM
3. price of the non-native asset is available on coinmarketCap
*/
package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/stellar/go/clients/horizonclient"
	hProtocol "github.com/stellar/go/protocols/horizon"
	"github.com/stellar/kelp/api"
	"github.com/stellar/kelp/plugins"
)

const native = "native"

func usage(errCode int) {
	log.Println("Usage:")
	file := filepath.Base(os.Args[0])
	log.Println(file + " address/cmcRef [address/cmcRef ...]")
	log.Println()
	log.Println("cmcRef is the given name for the non-native asset for coinmarketcap, XLM is 'stellar'")
	os.Exit(errCode)
}

func main() {
	addressCmcPairs := os.Args[1:]
	if len(addressCmcPairs) == 0 {
		usage(1)
	}

	sum := 0.0
	for _, acp := range addressCmcPairs {
		arr := strings.Split(acp, "/")
		if len(arr) != 2 {
			usage(2)
		}
		sum += getTotalNativeValue(arr[0], arr[1])
	}

	log.Println("total value in lumens:", sum, "XLM")
}

func getTotalNativeValue(address string, cmcRef string) float64 {
	client := horizonclient.DefaultPublicNetClient
	account := loadAccount(client, address)

	nativeBal := 0.0
	cryptoBal := 0.0
	cryptoNativeBal := 0.0
	for _, b := range account.Balances {
		bal, e := strconv.ParseFloat(b.Balance, 64)
		if e != nil {
			log.Fatal(e)
		}

		if b.Asset.Type == native {
			nativeBal = bal
		} else {
			cryptoBal = bal

			pf, e := makeCmcFeed("stellar")
			if e != nil {
				log.Fatal(e)
			}
			nativePriceInUSD, e := pf.GetPrice()
			if e != nil {
				log.Fatal(e)
			}

			pf, e = makeCmcFeed(cmcRef)
			if e != nil {
				log.Fatal(e)
			}
			cryptoPriceInUSD, e := pf.GetPrice()
			if e != nil {
				log.Fatal(e)
			}

			cryptoNativeBal = bal * cryptoPriceInUSD / nativePriceInUSD
		}
	}
	totalNativeValue := nativeBal + cryptoNativeBal

	log.Printf("%s: native (%.7f) + %s (%.7f) = %.7f XLM\n", address, nativeBal, cmcRef, cryptoBal, totalNativeValue)
	return totalNativeValue
}

func loadAccount(client *horizonclient.Client, address string) hProtocol.Account {
	acctReq := horizonclient.AccountRequest{AccountID: address}
	account, e := client.AccountDetail(acctReq)
	if e != nil {
		switch t := e.(type) {
		case *horizonclient.Error:
			log.Fatal(t.Problem)
		default:
			log.Fatal(e)
		}
	}
	return account
}

func makeCmcFeed(cmcRef string) (api.PriceFeed, error) {
	url := fmt.Sprintf("https://api.coinmarketcap.com/v1/ticker/%s/", cmcRef)
	return plugins.MakePriceFeed("crypto", url)
}
