package main

import (
	"encoding/base64"
	"fmt"
	"strconv"
	"time"

	"github.com/lightyeario/kelp/support/datamodel"

	"github.com/lightyeario/kelp/support"
	"github.com/stellar/go/build"
	"github.com/stellar/go/clients/horizon"
	"github.com/stellar/go/support/log"
)

const terminatorKey = "term"
const botPrefix = "b/"

// Terminator contains the logic to terminate offers
type Terminator struct {
	api                  *horizon.Client
	txb                  *kelp.TxButler
	tradingAccount       string
	tickIntervalSeconds  int32
	allowInactiveMinutes int32
}

// MakeTerminator is a factory method to make a Terminator
func MakeTerminator(
	api *horizon.Client,
	txb *kelp.TxButler,
	tradingAccount string,
	tickIntervalSeconds int32,
	allowInactiveMinutes int32,
) *Terminator {
	return &Terminator{
		api:                  api,
		txb:                  txb,
		tradingAccount:       tradingAccount,
		tickIntervalSeconds:  tickIntervalSeconds,
		allowInactiveMinutes: allowInactiveMinutes,
	}
}

// StartService starts the Terminator service
func (t *Terminator) StartService() {
	for {
		t.run()
		log.Info(fmt.Sprintf("sleeping for %d seconds...", t.tickIntervalSeconds))
		time.Sleep(time.Duration(t.tickIntervalSeconds) * time.Second)
	}
}

// botKeyPair is a pair of the datamodel.BotKey and the time the bot was last updated
type botKeyPair struct {
	dataKey     datamodel.BotKey
	lastUpdated int64
}

func (t *Terminator) run() {
	account, e := t.api.LoadAccount(t.tradingAccount)
	if e != nil {
		log.Error(e)
		return
	}

	// m is a map of hashes to botKeyPair(s)
	m, e := reconstructBotMap(account.Data)
	if e != nil {
		log.Error(e)
		return
	}
	log.Info(fmt.Sprintf("Found %d bots", len(m)))
	log.Info("bots in map: ", m)

	// compute cutoff millis
	nowMillis := time.Now().UnixNano() / 1000000
	cutoffMillis := nowMillis - (int64(t.allowInactiveMinutes) * 60 * 1000)
	log.Info("cutoff millis: ", cutoffMillis)

	// compute the inactive bots
	filtered := excludeActiveBots(m, cutoffMillis)
	log.Info(fmt.Sprintf("Found %d inactive bots", len(filtered)))
	if len(filtered) == 0 {
		// update data to reflect a successful return from terminator
		newTimestamp := time.Now().UnixNano() / 1000000
		tsMillisStr := fmt.Sprintf("%d", newTimestamp)
		ops := []build.TransactionMutator{
			build.SetData(terminatorKey, []byte(tsMillisStr), build.SourceAccount{AddressOrSeed: t.tradingAccount}),
		}

		log.Info("updating delete timestamp to ", tsMillisStr)
		e = t.txb.SubmitOps(ops)
		if e != nil {
			log.Error(e)
			return
		}
		return
	}

	// TODO 2 fix by passing in pointer directly
	offers, e := kelp.LoadAllOffers(t.tradingAccount, *t.api)
	if e != nil {
		log.Error(e)
		return
	}

	// delete the offers of inactive bots
	for hash, bk := range filtered {
		log.Info(fmt.Sprintf("working on bot with key: %s | lastUpdated: %d", bk.dataKey.Key(), bk.lastUpdated))

		assetA := convertToAsset(bk.dataKey.AssetBaseCode, bk.dataKey.AssetBaseIssuer)
		assetB := convertToAsset(bk.dataKey.AssetQuoteCode, bk.dataKey.AssetQuoteIssuer)
		inactiveSellOffers, inactiveBuyOffers := kelp.FilterOffers(offers, assetA, assetB)
		newTimestamp := time.Now().UnixNano() / 1000000
		t.deleteOffers(inactiveSellOffers, inactiveBuyOffers, hash, newTimestamp)
	}
}

func convertToAsset(code string, issuer string) horizon.Asset {
	if code == "native" {
		return kelp.Asset2Asset2(build.NativeAsset())
	}
	return kelp.Asset2Asset2(build.CreditAsset(code, issuer))
}

// deleteOffers deletes passed in offers along with the data for the passed in hash
func (t *Terminator) deleteOffers(sellOffers []horizon.Offer, buyOffers []horizon.Offer, hash string, tsMillis int64) {
	ops := []build.TransactionMutator{}
	ops = append(ops, t.txb.DeleteAllOffers(sellOffers)...)
	ops = append(ops, t.txb.DeleteAllOffers(buyOffers)...)
	numOffers := len(ops)

	// delete existing data entries
	dataKeyPrefix := botPrefix + hash
	ops = append(ops, build.ClearData(dataKeyPrefix+"/0", build.SourceAccount{AddressOrSeed: t.tradingAccount}))
	ops = append(ops, build.ClearData(dataKeyPrefix+"/1", build.SourceAccount{AddressOrSeed: t.tradingAccount}))
	ops = append(ops, build.ClearData(dataKeyPrefix+"/2", build.SourceAccount{AddressOrSeed: t.tradingAccount}))
	ops = append(ops, build.ClearData(dataKeyPrefix+"/3", build.SourceAccount{AddressOrSeed: t.tradingAccount}))
	ops = append(ops, build.ClearData(dataKeyPrefix+"/4", build.SourceAccount{AddressOrSeed: t.tradingAccount}))

	// update timestamp for terminator
	tsMillisStr := fmt.Sprintf("%d", tsMillis)
	ops = append(ops, build.SetData(terminatorKey, []byte(tsMillisStr), build.SourceAccount{AddressOrSeed: t.tradingAccount}))

	log.Info(fmt.Sprintf("deleting %d offers and 5 data entries, updating delete timestamp to %s", numOffers, tsMillisStr))
	if len(ops) > 0 {
		e := t.txb.SubmitOps(ops)
		if e != nil {
			log.Error(e)
			return
		}
	}
}

// excludeActiveBots filters out bots that have a lastUpdated timestamp that is greater than or equal to cutoffMillis
func excludeActiveBots(m map[string]botKeyPair, cutoffMillis int64) map[string]botKeyPair {
	filtered := make(map[string]botKeyPair)
	for k, v := range m {
		if v.lastUpdated < cutoffMillis {
			filtered[k] = v
		}
	}
	return filtered
}

func reconstructBotMap(data map[string]string) (map[string]botKeyPair, error) {
	m := make(map[string]botKeyPair)
	for k, v := range data {
		if !datamodel.IsBotKey(k) {
			continue
		}

		hash, botKeyPart := datamodel.SplitDataKey(k)
		currentBotKey, ok := m[hash]
		if !ok {
			currentBotKey = botKeyPair{}
		}
		e := updateBotKey(&currentBotKey, botKeyPart, v)
		if e != nil {
			return nil, e
		}
		m[hash] = currentBotKey
	}
	return m, nil
}

func updateBotKey(currentBotKey *botKeyPair, botKeyPart string, value string) error {
	decoded, e := base64.StdEncoding.DecodeString(value)
	if e != nil {
		return e
	}

	switch botKeyPart {
	case "0":
		var e error
		millisStr := string(decoded)
		currentBotKey.lastUpdated, e = strconv.ParseInt(millisStr, 10, 64)
		if e != nil {
			return e
		}
	case "1":
		currentBotKey.dataKey.AssetBaseCode = string(decoded)
	case "2":
		currentBotKey.dataKey.AssetBaseIssuer = string(decoded)
	case "3":
		currentBotKey.dataKey.AssetQuoteCode = string(decoded)
	case "4":
		currentBotKey.dataKey.AssetQuoteIssuer = string(decoded)
	}
	return nil
}
