package backend

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/stellar/go/clients/horizonclient"
	"github.com/stellar/go/keypair"
	"github.com/stellar/go/network"
	hProtocol "github.com/stellar/go/protocols/horizon"
	"github.com/stellar/go/strkey"
	"github.com/stellar/go/txnbuild"
	"github.com/stellar/kelp/gui/model2"
	"github.com/stellar/kelp/plugins"
	"github.com/stellar/kelp/support/kelpos"
	"github.com/stellar/kelp/support/toml"
	"github.com/stellar/kelp/support/utils"
	"github.com/stellar/kelp/trader"
)

type upsertBotConfigRequest struct {
	Name           string                `json:"name"`
	Strategy       string                `json:"strategy"`
	TraderConfig   trader.BotConfig      `json:"trader_config"`
	StrategyConfig plugins.BuySellConfig `json:"strategy_config"`
}

type upsertBotConfigResponse struct {
	Success bool `json:"success"`
}

type upsertBotConfigResponseErrors struct {
	Error  string                 `json:"error"`
	Fields upsertBotConfigRequest `json:"fields"`
}

func makeUpsertError(fields upsertBotConfigRequest) *upsertBotConfigResponseErrors {
	return &upsertBotConfigResponseErrors{
		Error:  "There are some errors marked in red inline",
		Fields: fields,
	}
}

func (s *APIServer) upsertBotConfig(w http.ResponseWriter, r *http.Request) {
	bodyBytes, e := ioutil.ReadAll(r.Body)
	if e != nil {
		s.writeErrorJson(w, fmt.Sprintf("error reading request input: %s", e))
		return
	}
	log.Printf("upsertBotConfig requestJson: %s\n", string(bodyBytes))

	var req upsertBotConfigRequest
	e = json.Unmarshal(bodyBytes, &req)
	if e != nil {
		s.writeErrorJson(w, fmt.Sprintf("error unmarshaling json: %s; bodyString = %s", e, string(bodyBytes)))
		return
	}

	botState, e := s.kos.QueryBotState(req.Name)
	if e != nil {
		s.writeErrorJson(w, fmt.Sprintf("error getting bot state for bot '%s': %s", req.Name, e))
		return
	}
	if botState != kelpos.BotStateStopped {
		s.writeErrorJson(w, fmt.Sprintf("bot state needs to be '%s' when upserting bot config, but was '%s'\n", kelpos.BotStateStopped, botState))
		return
	}

	// validate before init validation so we return validation errors to user instead of throwing unknown errors on init if file is invalid
	if errResp := s.validateConfigs(req); errResp != nil {
		s.writeJson(w, errResp)
		return
	}

	// init after validation so we return validation errors to user instead of throwing unknown errors on init if file is invalid
	e = req.TraderConfig.Init()
	if e != nil {
		s.writeErrorJson(w, fmt.Sprintf("error running Init() for TraderConfig: %s", e))
		return
	}

	e = s.setupOpsDirectory()
	if e != nil {
		s.writeError(w, fmt.Sprintf("error setting up ops directory: %s\n", e))
		return
	}

	filenamePair := model2.GetBotFilenames(req.Name, req.Strategy)
	traderFilePath := s.botConfigsPath.Join(filenamePair.Trader)
	botConfig := req.TraderConfig
	log.Printf("upsert bot config to file: %s\n", traderFilePath.AsString())
	e = toml.WriteFile(traderFilePath.Native(), &botConfig)
	if e != nil {
		s.writeErrorJson(w, fmt.Sprintf("error writing trader botConfig toml file for bot '%s': %s", req.Name, e))
		return
	}

	strategyFilePath := s.botConfigsPath.Join(filenamePair.Strategy)
	strategyConfig := req.StrategyConfig
	log.Printf("upsert strategy config to file: %s\n", strategyFilePath.AsString())
	e = toml.WriteFile(strategyFilePath.Native(), &strategyConfig)
	if e != nil {
		s.writeErrorJson(w, fmt.Sprintf("error writing strategy toml file for bot '%s': %s", req.Name, e))
		return
	}

	// check if we need to create new funding accounts and new trustlines
	s.reinitBotCheck(req)

	s.writeJson(w, upsertBotConfigResponse{Success: true})
}

func (s *APIServer) validateConfigs(req upsertBotConfigRequest) *upsertBotConfigResponseErrors {
	hasError := false
	errResp := upsertBotConfigRequest{
		TraderConfig:   trader.BotConfig{},
		StrategyConfig: plugins.BuySellConfig{},
	}

	if _, e := strkey.Decode(strkey.VersionByteSeed, req.TraderConfig.TradingSecretSeed); e != nil {
		errResp.TraderConfig.TradingSecretSeed = "invalid Trader Secret Key"
		hasError = true
	} else {
		// only check this if it is a valid trading secret seed
		tradingAccount, e := utils.ParseSecret(req.TraderConfig.TradingSecretSeed)
		if e != nil {
			errResp.TraderConfig.TradingSecretSeed = fmt.Sprintf("unable to parse: %s", e)
			hasError = true
		} else {
			if req.TraderConfig.IssuerA == *tradingAccount {
				errResp.TraderConfig.TradingSecretSeed = "cannot trade using issuer account"
				errResp.TraderConfig.IssuerA = "cannot trade asset issued by trading account"
				hasError = true
			}
			if req.TraderConfig.IssuerB == *tradingAccount {
				errResp.TraderConfig.TradingSecretSeed = "cannot trade using issuer account"
				errResp.TraderConfig.IssuerB = "cannot trade asset issued by trading account"
				hasError = true
			}
		}
	}

	if req.TraderConfig.AssetCodeA == "" || len(req.TraderConfig.AssetCodeA) > 12 {
		errResp.TraderConfig.AssetCodeA = "1 - 12 characters"
		hasError = true
	}
	if _, e := strkey.Decode(strkey.VersionByteAccountID, req.TraderConfig.IssuerA); req.TraderConfig.AssetCodeA != "XLM" && e != nil {
		errResp.TraderConfig.IssuerA = "invalid issuer account ID"
		hasError = true
	}

	if req.TraderConfig.AssetCodeB == "" || len(req.TraderConfig.AssetCodeB) > 12 {
		errResp.TraderConfig.AssetCodeB = "1 - 12 characters"
		hasError = true
	}
	if _, e := strkey.Decode(strkey.VersionByteAccountID, req.TraderConfig.IssuerB); req.TraderConfig.AssetCodeB != "XLM" && e != nil {
		errResp.TraderConfig.IssuerB = "invalid issuer account ID"
		hasError = true
	}

	if _, e := strkey.Decode(strkey.VersionByteSeed, req.TraderConfig.SourceSecretSeed); req.TraderConfig.SourceSecretSeed != "" && e != nil {
		errResp.TraderConfig.SourceSecretSeed = "invalid Source Secret Key"
		hasError = true
	}

	if len(req.StrategyConfig.Levels) == 0 || hasNewLevel(req.StrategyConfig.Levels) {
		errResp.StrategyConfig.Levels = []plugins.StaticLevel{}
		hasError = true
	}

	if hasError {
		return makeUpsertError(errResp)
	}
	return nil
}

func hasNewLevel(levels []plugins.StaticLevel) bool {
	for _, l := range levels {
		if l.AMOUNT == 0 || l.SPREAD == 0 {
			return true
		}
	}
	return false
}

func (s *APIServer) reinitBotCheck(req upsertBotConfigRequest) {
	isTestnet := strings.Contains(req.TraderConfig.HorizonURL, "test")
	bot := &model2.Bot{
		Name:     req.Name,
		Strategy: req.Strategy,
		Running:  false,
		Test:     isTestnet,
	}

	// set bot state to initializing so it handles the update
	s.kos.RegisterBotWithStateUpsert(bot, kelpos.InitState())

	// we only want to start initializing bot once it has been created, so we only advance state if everything is completed
	go func() {
		tradingKP, e := keypair.Parse(req.TraderConfig.TradingSecretSeed)
		if e != nil {
			log.Printf("error parsing trading secret seed for bot '%s': %s\n", bot.Name, e)
			return
		}
		traderAccount, e := s.checkFundAccount(tradingKP.Address(), bot.Name)
		if e != nil {
			log.Printf("error checking and funding trader account during upsert config: %s\n", e)
			return
		}

		// add trustline for trader account if needed
		assets := []hProtocol.Asset{
			req.TraderConfig.AssetBase(),
			req.TraderConfig.AssetQuote(),
		}
		e = s.checkAddTrustline(*traderAccount, tradingKP, req.TraderConfig.TradingSecretSeed, bot.Name, isTestnet, assets)
		if e != nil {
			log.Printf("error checking and adding trustline to trader account during upsert config: %s\n", e)
			return
		}

		// fund source account if needed
		if req.TraderConfig.SourceSecretSeed != "" {
			sourceKP, e := keypair.Parse(req.TraderConfig.SourceSecretSeed)
			if e != nil {
				log.Printf("error parsing source secret seed for bot '%s': %s\n", bot.Name, e)
				return
			}
			_, e = s.checkFundAccount(sourceKP.Address(), bot.Name)
			if e != nil {
				fmt.Printf("error checking and funding source account during upsert config: %s\n", e)
				return
			}
		}

		// advance bot state
		e = s.kos.AdvanceBotState(bot.Name, kelpos.InitState())
		if e != nil {
			log.Printf("error advancing bot state after reinitializing account for bot '%s': %s\n", bot.Name, e)
			return
		}
	}()
}

func (s *APIServer) checkAddTrustline(account hProtocol.Account, kp keypair.KP, traderSeed string, botName string, isTestnet bool, assets []hProtocol.Asset) error {
	activeNetwork := network.PublicNetworkPassphrase
	client := s.apiPubNet
	if isTestnet {
		activeNetwork = network.TestNetworkPassphrase
		client = s.apiTestNet
	}

	address := kp.Address()
	// find trustlines to be added
	trustlines := []hProtocol.Asset{}
	for _, a := range assets {
		if a.Type == "native" {
			log.Printf("not adding a trustline for the native asset\n")
			continue
		}
		if a.Issuer == address {
			log.Printf("not adding a trustline for an asset created by this trading account\n")
			continue
		}

		found := false
		// inefficient to have a double for-loop but ok since there will only ever be two assets in the list
		for _, bal := range account.Balances {
			log.Printf("iterating through asset balance: %+v\n", bal)
			if bal.Asset.Type != "native" && bal.Asset.Code == a.Code && bal.Asset.Issuer == a.Issuer {
				found = true
				break
			}
		}
		if !found {
			trustlines = append(trustlines, a)
		} else {
			log.Printf("trustline exists for asset %+v\n", a)
		}
	}
	if len(trustlines) == 0 {
		return nil
	}

	// build txn
	accountReq := horizonclient.AccountRequest{AccountID: address}
	account, err := client.AccountDetail(accountReq)
	if err != nil {
		return fmt.Errorf("Unable to load account for %s\n: %s", address, err)
	}

	var txOps []txnbuild.Operation
	for _, a := range trustlines {
		trustOp := txnbuild.ChangeTrust{
			Line: txnbuild.CreditAsset{Code: a.Code, Issuer: a.Issuer},
		}
		txOps = append(txOps, &trustOp)
		log.Printf("added trust asset operation to transaction for asset: %+v\n", a)
	}

	tx, e := txnbuild.NewTransaction(
		txnbuild.TransactionParams{
			SourceAccount: &account,
			Operations:    txOps,
			Timebounds:    txnbuild.NewInfiniteTimeout(),
			BaseFee:       100,
			// If IncrementSequenceNum is true, NewTransaction() will call `sourceAccount.IncrementSequenceNumber()`
			// to obtain the sequence number for the transaction.
			// If IncrementSequenceNum is false, NewTransaction() will call `sourceAccount.GetSequenceNumber()`
			// to obtain the sequence number for the transaction.
			// leaving as true since that's what it was in the old sdk so we want to maintain backward compatibility and we
			// need to increment the seq number on the account somewhere to use the next seq num
			IncrementSequenceNum: true,
		},
	)
	if e != nil {
		return fmt.Errorf("cannot make tx to create trustline transaction for account %s for bot '%s': %s", address, botName, e)
	}

	kpSigner, e := keypair.Parse(traderSeed)
	if e != nil {
		return fmt.Errorf("cannot parse seed  %s required for signing: %s", traderSeed, e)
	}

	tx, e = tx.Sign(activeNetwork, kpSigner.(*keypair.Full))
	if e != nil {
		return fmt.Errorf("cannot sign trustline transaction for account %s for bot '%s': %s", address, botName, e)
	}

	txn64, e := tx.Base64()
	if e != nil {
		return fmt.Errorf("cannot convert trustline transaction to base64 for account %s for bot '%s': %s", address, botName, e)
	}

	txSuccess, e := client.SubmitTransactionXDR(txn64)
	if e != nil {
		var herr *horizonclient.Error
		switch t := e.(type) {
		case *horizonclient.Error:
			herr = t
		case horizonclient.Error:
			herr = &t
		default:
			return fmt.Errorf("error when submitting change trust transaction for address %s for bot '%s' for assets(%v): %s (%s)", address, botName, trustlines, e, txn64)
		}
		return fmt.Errorf("horizon error when submitting change trust transaction for address %s for bot '%s' for assets(%v): %s (%s)", address, botName, trustlines, *herr, txn64)
	}

	log.Printf("tx result of submitting trustline transaction for address %s for bot '%s' for assets(%v): %v (%s)\n", address, botName, trustlines, txSuccess, txn64)
	return nil
}
