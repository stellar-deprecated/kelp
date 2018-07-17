package utils

import (
	"strconv"

	"github.com/pkg/errors"
	"github.com/stellar/go/build"
	"github.com/stellar/go/clients/horizon"
	"github.com/stellar/go/support/log"
)

const baseReserve = 0.5

// TxButler helps with building and submitting transactions to the Stellar network
type TxButler struct {
	API                        *horizon.Client
	SourceAccount              string
	TradingAccount             string
	SourceSeed                 string
	TradingSeed                string
	Network                    build.Network
	FractionalReserveMagnifier int8
	operationalBuffer          float64
	seqNum                     uint64
	reloadSeqNum               bool

	// uninitialized
	cachedXlmExposure *float64
}

// MakeTxButler is a factory method for TxButler
func MakeTxButler(
	client *horizon.Client,
	sourceSeed string,
	tradingSeed string,
	sourceAccount string,
	tradingAccount string,
	network build.Network,
	fractionalReserveMagnifier int8,
	operationalBuffer float64,
) *TxButler {
	txb := &TxButler{
		API:                        client,
		SourceSeed:                 sourceSeed,
		TradingSeed:                tradingSeed,
		SourceAccount:              sourceAccount,
		TradingAccount:             tradingAccount,
		Network:                    network,
		FractionalReserveMagnifier: fractionalReserveMagnifier,
		operationalBuffer:          operationalBuffer,
	}

	//log.Info("init txbutler")
	log.Info("Using network passphrase: ", txb.Network.Passphrase)

	if txb.SourceAccount == "" {
		txb.SourceAccount = txb.TradingAccount
		txb.SourceSeed = txb.TradingSeed
		log.Info("No Source Account Set")
	}
	txb.reloadSeqNum = true

	return txb
}

/*
func (txb *TxButler) SetSeqNum(num string) {
	s, err := strconv.ParseUint(num, 10, 64)
	if err != nil {
		log.Info("SetSeqNum :", num, " failed: ", err)
		return
	}
	txb.seqNum = s
	txb.reloadSeqNum = false
}
*/

func (txb *TxButler) incrementSeqNum() {
	if txb.reloadSeqNum {
		log.Info("reloadSeqNum ")
		seqNum, err := txb.API.SequenceForAccount(txb.SourceAccount)
		if err != nil {
			log.Info("error getting seq num ", err)
			return
		}
		txb.seqNum = uint64(seqNum)
		txb.reloadSeqNum = false
	}
	txb.seqNum++

}

// DeleteAllOffers is a helper that accumulates delete operations for the passed in offers
func (txb *TxButler) DeleteAllOffers(offers []horizon.Offer) []build.TransactionMutator {
	ops := []build.TransactionMutator{}
	for _, offer := range offers {
		op := txb.DeleteOffer(offer)
		ops = append(ops, &op)
	}
	return ops
}

// DeleteOffer returns the op that needs to be submitted to the network in order to delete the passed in offer
func (txb *TxButler) DeleteOffer(offer horizon.Offer) build.ManageOfferBuilder {
	rate := build.Rate{
		Selling: Asset2Asset(offer.Selling),
		Buying:  Asset2Asset(offer.Buying),
		Price:   build.Price(offer.Price),
	}

	if txb.SourceAccount == txb.TradingAccount {
		return build.ManageOffer(false, build.Amount("0"), rate, build.OfferID(offer.ID))
	}
	return build.ManageOffer(false, build.Amount("0"), rate, build.OfferID(offer.ID), build.SourceAccount{AddressOrSeed: txb.TradingAccount})
}

// ModifyBuyOffer modifies a buy offer
func (txb *TxButler) ModifyBuyOffer(offer horizon.Offer, price float64, amount float64) *build.ManageOfferBuilder {
	//log.Info("modifyBuyOffer: ", offer.ID, " p:", price)
	return txb.ModifySellOffer(offer, 1/price, amount*price)
}

// ModifySellOffer modifies a sell offer
func (txb *TxButler) ModifySellOffer(offer horizon.Offer, price float64, amount float64) *build.ManageOfferBuilder {
	//log.Info("modifySellOffer: ", offer.ID, " p:", amount)
	return txb.createModifySellOffer(&offer, offer.Selling, offer.Buying, price, amount)
}

// CreateSellOffer creates a sell offer
func (txb *TxButler) CreateSellOffer(base horizon.Asset, counter horizon.Asset, price float64, amount float64) *build.ManageOfferBuilder {
	if amount > 0 {
		//log.Info("createSellOffer: ", price, amount)
		return txb.createModifySellOffer(nil, base, counter, price, amount)
	}
	log.Info("zero amount ")
	return nil
}

// ParseOfferAmount is a convenience method to parse an offer amount created by the txButler
func (txb *TxButler) ParseOfferAmount(amt string) (float64, error) {
	offerAmt, err := strconv.ParseFloat(amt, 64)
	if err != nil {
		log.Info("error parsing offer amount: ", err)
		return -1, err
	}
	return offerAmt, nil
}

func (txb *TxButler) minReserve(subentries int32) float64 {
	return float64(2+subentries) * baseReserve
}

func (txb *TxButler) lumenBalance() (float64, float64, error) {
	account, err := txb.API.LoadAccount(txb.TradingAccount)
	if err != nil {
		log.Info("error loading account to fetch lumen balance: ", err)
		return -1, -1, nil
	}

	for _, balance := range account.Balances {
		if balance.Asset.Type == Native {
			b, e := strconv.ParseFloat(balance.Balance, 64)
			if e != nil {
				log.Info("error parsing native balance: ", e)
			}
			return b, txb.minReserve(account.SubentryCount), e
		}
	}
	return -1, -1, errors.New("could not find a native lumen balance")
}

// createModifySellOffer is the main method that handles the logic of creating or modifying an offer, note that all offers are treated as sell offers in Stellar
func (txb *TxButler) createModifySellOffer(offer *horizon.Offer, selling horizon.Asset, buying horizon.Asset, price float64, amount float64) *build.ManageOfferBuilder {
	if selling.Type == Native {
		var incrementalXlmAmount float64
		if offer != nil {
			offerAmt, err := txb.ParseOfferAmount(offer.Amount)
			if err != nil {
				log.Info(err)
				return nil
			}
			// modifying an offer will not increase the min reserve but will affect the xlm exposure
			incrementalXlmAmount = amount - offerAmt
		} else {
			// creating a new offer will incrase the min reserve on the account so add baseReserve
			incrementalXlmAmount = amount + baseReserve
		}

		// check if incrementalXlmAmount is within budget
		bal, minAccountBal, err := txb.lumenBalance()
		if err != nil {
			log.Info(err)
			return nil
		}

		xlmExposure, err := txb.xlmExposure()
		if err != nil {
			log.Info(err)
			return nil
		}

		additionalExposure := incrementalXlmAmount >= 0
		possibleTerminalExposure := ((xlmExposure + incrementalXlmAmount) / float64(txb.FractionalReserveMagnifier)) > (bal - minAccountBal - txb.operationalBuffer)
		if additionalExposure && possibleTerminalExposure {
			log.Info("not placing offer because we run the risk of running out of lumens | xlmExposure: ", xlmExposure,
				" | incrementalXlmAmount: ", incrementalXlmAmount, " | bal: ", bal, " | minAccountBal: ", minAccountBal,
				" | operationalBuffer: ", txb.operationalBuffer, " | fractionalReserveMagnifier: ", txb.FractionalReserveMagnifier)
			return nil
		}
	}

	stringPrice := strconv.FormatFloat(price, 'f', 8, 64)
	rate := build.Rate{
		Selling: Asset2Asset(selling),
		Buying:  Asset2Asset(buying),
		Price:   build.Price(stringPrice),
	}

	//log.Info("sa: ", txb.sourceAccount, " ta:", txb.tradingAccount, " r:", rate)
	mutators := []interface{}{
		rate,
		build.Amount(strconv.FormatFloat(amount, 'f', -1, 64)),
	}
	if offer != nil {
		mutators = append(mutators, build.OfferID(offer.ID))
	}
	if txb.SourceAccount != txb.TradingAccount {
		mutators = append(mutators, build.SourceAccount{AddressOrSeed: txb.TradingAccount})
	}
	result := build.ManageOffer(false, mutators...)
	return &result
}

// SubmitOps submits the passed in operations to the network asynchronously in a single transaction
func (txb *TxButler) SubmitOps(ops []build.TransactionMutator) error {
	txb.incrementSeqNum()
	muts := []build.TransactionMutator{
		build.Sequence{Sequence: txb.seqNum},
		txb.Network,
		build.SourceAccount{AddressOrSeed: txb.SourceAccount},
	}
	muts = append(muts, ops...)
	tx := build.Transaction(muts...)
	if tx.Err != nil {
		return errors.Wrap(tx.Err, "txButler SubmitOps error: ")
	}
	go txb.signAndSubmit(tx)
	return nil
}

// CreateBuyOffer creates a buy offer
func (txb *TxButler) CreateBuyOffer(base horizon.Asset, counter horizon.Asset, price float64, amount float64) *build.ManageOfferBuilder {
	//log.Info("createBuyOffer: ", price, amount)
	return txb.CreateSellOffer(counter, base, 1/price, amount*price)
}

func (txb *TxButler) signAndSubmit(tx *build.TransactionBuilder) {
	var txe build.TransactionEnvelopeBuilder
	if txb.SourceSeed != txb.TradingSeed {
		txe = tx.Sign(txb.SourceSeed, txb.TradingSeed)
	} else {
		txe = tx.Sign(txb.SourceSeed)
	}

	txeB64, err := txe.Base64()
	if err != nil {
		log.Error("", err)
		return
	}
	log.Info("tx: ", txeB64)

	resp, err := txb.API.SubmitTransaction(txeB64)
	if err != nil {
		if herr, ok := errors.Cause(err).(*horizon.Error); ok {
			var rcs *horizon.TransactionResultCodes
			rcs, err = herr.ResultCodes()
			if err != nil {
				log.Info("no rc from horizon: ", err)
				return
			}
			if rcs.TransactionCode == "tx_bad_seq" {
				log.Info("tx_bad_seq, reloading")
				txb.reloadSeqNum = true
			}

			log.Info("tx code: ", rcs.TransactionCode, " opcodes: ", rcs.OperationCodes)
		} else {
			log.Info("tx failed: ", err)
		}
		return
	}

	log.Info("", resp.Hash)
}

// ResetCachedXlmExposure resets the cache
func (txb *TxButler) ResetCachedXlmExposure() {
	txb.cachedXlmExposure = nil
}

func (txb *TxButler) xlmExposure() (float64, error) {
	if txb.cachedXlmExposure != nil {
		return *txb.cachedXlmExposure, nil
	}

	// uses all offers for this trading account to accommodate sharing by other bots
	offers, err := LoadAllOffers(txb.TradingAccount, txb.API)
	if err != nil {
		log.Info("error computing XLM exposure: ", err)
		return -1, err
	}

	var sum float64
	for _, offer := range offers {
		// only need to compute sum of selling because that's the max XLM we can give up if all our offers are taken
		if offer.Selling.Type == Native {
			offerAmt, err := txb.ParseOfferAmount(offer.Amount)
			if err != nil {
				return -1, err
			}
			sum += offerAmt
		}
	}

	txb.cachedXlmExposure = &sum
	return sum, nil
}
