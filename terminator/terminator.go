package main

import (
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/lightyeario/kelp/support"
	"github.com/stellar/go/build"
	"github.com/stellar/go/clients/horizon"
	"github.com/stellar/go/support/log"
)

const terminatorKey = "term"

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

type botKey struct {
	assetACode   string
	assetAIssuer string
	assetBCode   string
	assetBIssuer string
	lastUpdated  int64
}

func (t *Terminator) run() {
	account, e := t.api.LoadAccount(t.tradingAccount)
	if e != nil {
		log.Error(e)
		return
	}

	// m is a map of hashes to botKey(s)
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
		key := fmt.Sprintf("%s/%s/%s/%s/%d", bk.assetACode, bk.assetAIssuer, bk.assetBCode, bk.assetBIssuer, bk.lastUpdated)
		log.Info("working on bot with key: ", key)

		assetA := convertToAsset(bk.assetACode, bk.assetAIssuer)
		assetB := convertToAsset(bk.assetBCode, bk.assetBIssuer)
		inactiveSellOffers, inactiveBuyOffers := kelp.FilterOffers(offers, assetA, assetB)
		t.deleteOffers(inactiveSellOffers, inactiveBuyOffers, hash)
	}
}

func convertToAsset(code string, issuer string) horizon.Asset {
	if code == "native" {
		return kelp.Asset2Asset2(build.NativeAsset())
	}
	return kelp.Asset2Asset2(build.CreditAsset(code, issuer))
}

// deleteOffers deletes passed in offers along with the data for the passed in hash
func (t *Terminator) deleteOffers(sellOffers []horizon.Offer, buyOffers []horizon.Offer, hash string) {
	ops := []build.TransactionMutator{}
	ops = append(ops, t.txb.DeleteAllOffers(sellOffers)...)
	ops = append(ops, t.txb.DeleteAllOffers(buyOffers)...)
	numOffers := len(ops)

	ops = append(ops, build.ClearData(hash+"/0", build.SourceAccount{AddressOrSeed: t.tradingAccount}))
	ops = append(ops, build.ClearData(hash+"/1", build.SourceAccount{AddressOrSeed: t.tradingAccount}))
	ops = append(ops, build.ClearData(hash+"/2", build.SourceAccount{AddressOrSeed: t.tradingAccount}))
	ops = append(ops, build.ClearData(hash+"/3", build.SourceAccount{AddressOrSeed: t.tradingAccount}))
	ops = append(ops, build.ClearData(hash+"/4", build.SourceAccount{AddressOrSeed: t.tradingAccount}))

	log.Info(fmt.Sprintf("deleting %d offers and 5 data entries", numOffers))
	if len(ops) > 0 {
		e := t.txb.SubmitOps(ops)
		if e != nil {
			log.Error(e)
			return
		}
	}
}

// excludeActiveBots filters out bots that have a lastUpdated timestamp that is greater than or equal to cutoffMillis
func excludeActiveBots(m map[string]botKey, cutoffMillis int64) map[string]botKey {
	filtered := make(map[string]botKey)
	for k, v := range m {
		if v.lastUpdated < cutoffMillis {
			filtered[k] = v
		}
	}
	return filtered
}

func reconstructBotMap(data map[string]string) (map[string]botKey, error) {
	m := make(map[string]botKey)
	for k, v := range data {
		if k == terminatorKey {
			continue
		}

		keyParts := strings.Split(k, "/")
		hash := keyParts[0]
		botKeyPart := keyParts[1]

		currentBotKey, ok := m[hash]
		if !ok {
			currentBotKey = botKey{}
		}
		e := updateBotKey(&currentBotKey, botKeyPart, v)
		if e != nil {
			return nil, e
		}
		m[hash] = currentBotKey
	}
	return m, nil
}

func updateBotKey(currentBotKey *botKey, botKeyPart string, value string) error {
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
		currentBotKey.assetACode = string(decoded)
	case "2":
		currentBotKey.assetAIssuer = string(decoded)
	case "3":
		currentBotKey.assetBCode = string(decoded)
	case "4":
		currentBotKey.assetBIssuer = string(decoded)
	}
	return nil
}
