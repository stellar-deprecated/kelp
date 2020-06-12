package terminator

import (
	"encoding/base64"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/stellar/go/clients/horizonclient"
	hProtocol "github.com/stellar/go/protocols/horizon"
	"github.com/stellar/go/txnbuild"
	"github.com/stellar/kelp/api"
	"github.com/stellar/kelp/model"
	"github.com/stellar/kelp/plugins"
	"github.com/stellar/kelp/support/utils"
)

const terminatorKey = "term"

// Terminator contains the logic to terminate offers
type Terminator struct {
	api                  *horizonclient.Client
	sdex                 *plugins.SDEX
	tradingAccount       string
	tickIntervalSeconds  int32
	allowInactiveMinutes int32
}

// MakeTerminator is a factory method to make a Terminator
func MakeTerminator(
	api *horizonclient.Client,
	sdex *plugins.SDEX,
	tradingAccount string,
	tickIntervalSeconds int32,
	allowInactiveMinutes int32,
) *Terminator {
	return &Terminator{
		api:                  api,
		sdex:                 sdex,
		tradingAccount:       tradingAccount,
		tickIntervalSeconds:  tickIntervalSeconds,
		allowInactiveMinutes: allowInactiveMinutes,
	}
}

// StartService starts the Terminator service
func (t *Terminator) StartService() {
	for {
		t.run()
		log.Printf("sleeping for %d seconds...\n", t.tickIntervalSeconds)
		time.Sleep(time.Duration(t.tickIntervalSeconds) * time.Second)
	}
}

// botKeyPair is a pair of the model.BotKey and the time the bot was last updated
type botKeyPair struct {
	dataKey     model.BotKey
	lastUpdated int64
}

// String impl
func (kp botKeyPair) String() string {
	return fmt.Sprintf("botKeyPair(dataKey=%v, lastUpdated=%d)", kp.dataKey, kp.lastUpdated)
}

// TODO 3 add db-based support, manage-data based support is invalid since we don't write it from trader anymore.
func (t *Terminator) run() {
	// use condition to avoid unreachable code warning
	if 1 < 2 {
		panic("need to add db-based support, manage-data based support is invalid since we don't write it from trader anymore.")
	}

	acctReq := horizonclient.AccountRequest{AccountID: t.tradingAccount}
	account, e := t.api.AccountDetail(acctReq)
	if e != nil {
		log.Println(e)
		return
	}

	// m is a map of hashes to botKeyPair(s)
	botList, e := reconstructBotList(account.Data)
	if e != nil {
		log.Println(e)
		return
	}

	// compute cutoff millis
	nowMillis := time.Now().UnixNano() / 1000000
	cutoffMillis := nowMillis - (int64(t.allowInactiveMinutes) * 60 * 1000)
	log.Printf("cutoff millis: %d\n", cutoffMillis)

	// compute the inactive bots
	inactiveBots := excludeActiveBots(botList, cutoffMillis)
	log.Printf("Found %d inactive bots\n", len(inactiveBots))
	if len(inactiveBots) == 0 {
		// update data to reflect a successful return from terminator
		newTimestamp := time.Now().UnixNano() / 1000000
		tsMillisStr := fmt.Sprintf("%d", newTimestamp)
		ops := []txnbuild.Operation{
			&txnbuild.ManageData{
				Name:          terminatorKey,
				Value:         []byte(tsMillisStr),
				SourceAccount: &txnbuild.SimpleAccount{AccountID: t.tradingAccount},
			},
		}

		log.Printf("updating delete timestamp to %s\n", tsMillisStr)
		e = t.sdex.SubmitOps(api.ConvertOperation2TM(ops), api.SubmitModeBoth, nil)
		if e != nil {
			log.Println(e)
		}
		return
	}

	offers, e := utils.LoadAllOffers(t.tradingAccount, t.api)
	if e != nil {
		log.Println(e)
		return
	}

	// delete the offers of inactive bots (don't ever use hash directly)
	for _, bk := range inactiveBots {
		log.Printf("working on bot with key: %+v\n", bk)
		assetA := convertToAsset(bk.dataKey.AssetBaseCode, bk.dataKey.AssetBaseIssuer)
		assetB := convertToAsset(bk.dataKey.AssetQuoteCode, bk.dataKey.AssetQuoteIssuer)
		inactiveSellOffers, inactiveBuyOffers := utils.FilterOffers(offers, assetA, assetB)
		newTimestamp := time.Now().UnixNano() / 1000000
		t.deleteOffers(inactiveSellOffers, inactiveBuyOffers, bk.dataKey, newTimestamp)
	}
}

func convertToAsset(code string, issuer string) hProtocol.Asset {
	if code == utils.Native {
		return utils.Asset2Asset2(txnbuild.NativeAsset{})
	}
	return utils.Asset2Asset2(txnbuild.CreditAsset{code, issuer})
}

// deleteOffers deletes passed in offers along with the data for the passed in hash
func (t *Terminator) deleteOffers(sellOffers []hProtocol.Offer, buyOffers []hProtocol.Offer, botKey model.BotKey, tsMillis int64) {
	ops := []txnbuild.Operation{}
	ops = append(ops, t.sdex.DeleteAllOffers(sellOffers)...)
	ops = append(ops, t.sdex.DeleteAllOffers(buyOffers)...)
	numOffers := len(ops)

	// delete existing data entries
	ops = append(ops, &txnbuild.ManageData{Name: botKey.FullKey(0),
		SourceAccount: &txnbuild.SimpleAccount{AccountID: t.tradingAccount},
	})
	ops = append(ops, &txnbuild.ManageData{Name: botKey.FullKey(1),
		SourceAccount: &txnbuild.SimpleAccount{AccountID: t.tradingAccount},
	})
	if len(botKey.AssetBaseIssuer) > 0 {
		ops = append(ops, &txnbuild.ManageData{Name: botKey.FullKey(2),
			SourceAccount: &txnbuild.SimpleAccount{AccountID: t.tradingAccount},
		})
	}
	ops = append(ops, &txnbuild.ManageData{Name: botKey.FullKey(3),
		SourceAccount: &txnbuild.SimpleAccount{AccountID: t.tradingAccount},
	})
	if len(botKey.AssetQuoteIssuer) > 0 {
		ops = append(ops, &txnbuild.ManageData{Name: botKey.FullKey(4),
			SourceAccount: &txnbuild.SimpleAccount{AccountID: t.tradingAccount},
		})
	}

	// update timestamp for terminator
	tsMillisStr := fmt.Sprintf("%d", tsMillis)
	ops = append(ops, &txnbuild.ManageData{
		Name:          terminatorKey,
		Value:         []byte(tsMillisStr),
		SourceAccount: &txnbuild.SimpleAccount{AccountID: t.tradingAccount},
	})

	log.Printf("deleting %d offers and 5 data entries, updating delete timestamp to %s\n", numOffers, tsMillisStr)
	if len(ops) > 0 {
		e := t.sdex.SubmitOps(api.ConvertOperation2TM(ops), api.SubmitModeBoth, nil)
		if e != nil {
			log.Println(e)
			return
		}
	}
}

// excludeActiveBots filters out bots that have a lastUpdated timestamp that is greater than or equal to cutoffMillis
func excludeActiveBots(botList []botKeyPair, cutoffMillis int64) []botKeyPair {
	inactive := []botKeyPair{}
	for _, v := range botList {
		if v.lastUpdated < cutoffMillis {
			inactive = append(inactive, v)
		}
	}
	return inactive
}

func reconstructBotList(data map[string]string) ([]botKeyPair, error) {
	m := make(map[string]botKeyPair)
	for k, v := range data {
		if !model.IsBotKey(k) {
			continue
		}

		hash, botKeyPart := model.SplitDataKey(k)
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

	// convert to a list
	l := []botKeyPair{}
	for _, k := range m {
		l = append(l, k)
	}

	log.Printf("Found %d bots\n", len(l))
	if len(l) > 0 {
		logLine := "bots in list:\n"
		for _, k := range l {
			logLine = logLine + fmt.Sprintf("\t%v\n", k)
		}
		log.Println(logLine)
	}

	return l, nil
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
