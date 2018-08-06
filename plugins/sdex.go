package plugins

import (
	"log"
	"strconv"

	"github.com/lightyeario/kelp/support/utils"
	"github.com/pkg/errors"
	"github.com/stellar/go/build"
	"github.com/stellar/go/clients/horizon"
)

const baseReserve = 0.5

// SDEX helps with building and submitting transactions to the Stellar network
type SDEX struct {
	API                        *horizon.Client
	SourceAccount              string
	TradingAccount             string
	SourceSeed                 string
	TradingSeed                string
	Network                    build.Network
	FractionalReserveMagnifier int8
	operationalBuffer          float64
	simMode                    bool

	// uninitialized
	seqNum            uint64
	reloadSeqNum      bool
	cachedXlmExposure *float64
}

// MakeSDEX is a factory method for SDEX
func MakeSDEX(
	api *horizon.Client,
	sourceSeed string,
	tradingSeed string,
	sourceAccount string,
	tradingAccount string,
	network build.Network,
	fractionalReserveMagnifier int8,
	operationalBuffer float64,
	simMode bool,
) *SDEX {
	sdex := &SDEX{
		API:                        api,
		SourceSeed:                 sourceSeed,
		TradingSeed:                tradingSeed,
		SourceAccount:              sourceAccount,
		TradingAccount:             tradingAccount,
		Network:                    network,
		FractionalReserveMagnifier: fractionalReserveMagnifier,
		operationalBuffer:          operationalBuffer,
		simMode:                    simMode,
	}

	log.Printf("Using network passphrase: %s\n", sdex.Network.Passphrase)

	if sdex.SourceAccount == "" {
		sdex.SourceAccount = sdex.TradingAccount
		sdex.SourceSeed = sdex.TradingSeed
		log.Println("No Source Account Set")
	}
	sdex.reloadSeqNum = true

	return sdex
}

func (sdex *SDEX) incrementSeqNum() {
	if sdex.reloadSeqNum {
		log.Println("reloading sequence number")
		seqNum, err := sdex.API.SequenceForAccount(sdex.SourceAccount)
		if err != nil {
			log.Printf("error getting seq num: %s\n", err)
			return
		}
		sdex.seqNum = uint64(seqNum)
		sdex.reloadSeqNum = false
	}
	sdex.seqNum++
}

// DeleteAllOffers is a helper that accumulates delete operations for the passed in offers
func (sdex *SDEX) DeleteAllOffers(offers []horizon.Offer) []build.TransactionMutator {
	ops := []build.TransactionMutator{}
	for _, offer := range offers {
		op := sdex.DeleteOffer(offer)
		ops = append(ops, &op)
	}
	return ops
}

// DeleteOffer returns the op that needs to be submitted to the network in order to delete the passed in offer
func (sdex *SDEX) DeleteOffer(offer horizon.Offer) build.ManageOfferBuilder {
	rate := build.Rate{
		Selling: utils.Asset2Asset(offer.Selling),
		Buying:  utils.Asset2Asset(offer.Buying),
		Price:   build.Price(offer.Price),
	}

	if sdex.SourceAccount == sdex.TradingAccount {
		return build.ManageOffer(false, build.Amount("0"), rate, build.OfferID(offer.ID))
	}
	return build.ManageOffer(false, build.Amount("0"), rate, build.OfferID(offer.ID), build.SourceAccount{AddressOrSeed: sdex.TradingAccount})
}

// ModifyBuyOffer modifies a buy offer
func (sdex *SDEX) ModifyBuyOffer(offer horizon.Offer, price float64, amount float64) *build.ManageOfferBuilder {
	return sdex.ModifySellOffer(offer, 1/price, amount*price)
}

// ModifySellOffer modifies a sell offer
func (sdex *SDEX) ModifySellOffer(offer horizon.Offer, price float64, amount float64) *build.ManageOfferBuilder {
	return sdex.createModifySellOffer(&offer, offer.Selling, offer.Buying, price, amount)
}

// CreateSellOffer creates a sell offer
func (sdex *SDEX) CreateSellOffer(base horizon.Asset, counter horizon.Asset, price float64, amount float64) *build.ManageOfferBuilder {
	if amount > 0 {
		return sdex.createModifySellOffer(nil, base, counter, price, amount)
	}
	log.Println("error: cannot place sell order, zero amount")
	return nil
}

// ParseOfferAmount is a convenience method to parse an offer amount
func (sdex *SDEX) ParseOfferAmount(amt string) (float64, error) {
	offerAmt, err := strconv.ParseFloat(amt, 64)
	if err != nil {
		log.Printf("error parsing offer amount: %s\n", err)
		return -1, err
	}
	return offerAmt, nil
}

func (sdex *SDEX) minReserve(subentries int32) float64 {
	return float64(2+subentries) * baseReserve
}

func (sdex *SDEX) lumenBalance() (float64, float64, error) {
	account, err := sdex.API.LoadAccount(sdex.TradingAccount)
	if err != nil {
		log.Printf("error loading account to fetch lumen balance: %s\n", err)
		return -1, -1, nil
	}

	for _, balance := range account.Balances {
		if balance.Asset.Type == utils.Native {
			b, e := strconv.ParseFloat(balance.Balance, 64)
			if e != nil {
				log.Printf("error parsing native balance: %s\n", e)
			}
			return b, sdex.minReserve(account.SubentryCount), e
		}
	}
	return -1, -1, errors.New("could not find a native lumen balance")
}

// createModifySellOffer is the main method that handles the logic of creating or modifying an offer, note that all offers are treated as sell offers in Stellar
func (sdex *SDEX) createModifySellOffer(offer *horizon.Offer, selling horizon.Asset, buying horizon.Asset, price float64, amount float64) *build.ManageOfferBuilder {
	if selling.Type == utils.Native {
		var incrementalXlmAmount float64
		if offer != nil {
			offerAmt, err := sdex.ParseOfferAmount(offer.Amount)
			if err != nil {
				log.Println(err)
				return nil
			}
			// modifying an offer will not increase the min reserve but will affect the xlm exposure
			incrementalXlmAmount = amount - offerAmt
		} else {
			// creating a new offer will incrase the min reserve on the account so add baseReserve
			incrementalXlmAmount = amount + baseReserve
		}

		// check if incrementalXlmAmount is within budget
		bal, minAccountBal, err := sdex.lumenBalance()
		if err != nil {
			log.Println(err)
			return nil
		}

		xlmExposure, err := sdex.xlmExposure()
		if err != nil {
			log.Println(err)
			return nil
		}

		additionalExposure := incrementalXlmAmount >= 0
		possibleTerminalExposure := ((xlmExposure + incrementalXlmAmount) / float64(sdex.FractionalReserveMagnifier)) > (bal - minAccountBal - sdex.operationalBuffer)
		if additionalExposure && possibleTerminalExposure {
			log.Println("not placing offer because we run the risk of running out of lumens | xlmExposure:", xlmExposure,
				"| incrementalXlmAmount:", incrementalXlmAmount, "| bal:", bal, "| minAccountBal:", minAccountBal,
				"| operationalBuffer:", sdex.operationalBuffer, "| fractionalReserveMagnifier:", sdex.FractionalReserveMagnifier)
			return nil
		}
	}

	stringPrice := strconv.FormatFloat(price, 'f', 8, 64)
	rate := build.Rate{
		Selling: utils.Asset2Asset(selling),
		Buying:  utils.Asset2Asset(buying),
		Price:   build.Price(stringPrice),
	}

	mutators := []interface{}{
		rate,
		build.Amount(strconv.FormatFloat(amount, 'f', -1, 64)),
	}
	if offer != nil {
		mutators = append(mutators, build.OfferID(offer.ID))
	}
	if sdex.SourceAccount != sdex.TradingAccount {
		mutators = append(mutators, build.SourceAccount{AddressOrSeed: sdex.TradingAccount})
	}
	result := build.ManageOffer(false, mutators...)
	return &result
}

// SubmitOps submits the passed in operations to the network asynchronously in a single transaction
func (sdex *SDEX) SubmitOps(ops []build.TransactionMutator) error {
	sdex.incrementSeqNum()
	muts := []build.TransactionMutator{
		build.Sequence{Sequence: sdex.seqNum},
		sdex.Network,
		build.SourceAccount{AddressOrSeed: sdex.SourceAccount},
	}
	muts = append(muts, ops...)
	tx := build.Transaction(muts...)
	if tx.Err != nil {
		return errors.Wrap(tx.Err, "SubmitOps error: ")
	}

	// convert to xdr string
	txeB64, e := sdex.sign(tx)
	if e != nil {
		return e
	}
	log.Printf("tx XDR: %s\n", txeB64)

	// submit
	if !sdex.simMode {
		log.Println("submitting tx XDR to network")
		go sdex.submit(txeB64)
	} else {
		log.Println("not submitting tx XDR to network in simulation mode")
	}

	return nil
}

// CreateBuyOffer creates a buy offer
func (sdex *SDEX) CreateBuyOffer(base horizon.Asset, counter horizon.Asset, price float64, amount float64) *build.ManageOfferBuilder {
	return sdex.CreateSellOffer(counter, base, 1/price, amount*price)
}

func (sdex *SDEX) sign(tx *build.TransactionBuilder) (string, error) {
	var txe build.TransactionEnvelopeBuilder
	if sdex.SourceSeed != sdex.TradingSeed {
		txe = tx.Sign(sdex.SourceSeed, sdex.TradingSeed)
	} else {
		txe = tx.Sign(sdex.SourceSeed)
	}
	return txe.Base64()
}

func (sdex *SDEX) submit(txeB64 string) {
	resp, err := sdex.API.SubmitTransaction(txeB64)
	if err != nil {
		if herr, ok := errors.Cause(err).(*horizon.Error); ok {
			var rcs *horizon.TransactionResultCodes
			rcs, err = herr.ResultCodes()
			if err != nil {
				log.Printf("no rc from horizon: %s\n", err)
				return
			}
			if rcs.TransactionCode == "tx_bad_seq" {
				log.Println("tx_bad_seq, reloading")
				sdex.reloadSeqNum = true
			}

			log.Println("tx code:", rcs.TransactionCode, "opcodes:", rcs.OperationCodes)
		} else {
			log.Printf("tx failed: %s\n", err)
		}
		return
	}

	log.Printf("tx-hash: %s\n", resp.Hash)
}

// ResetCachedXlmExposure resets the cache
func (sdex *SDEX) ResetCachedXlmExposure() {
	sdex.cachedXlmExposure = nil
}

func (sdex *SDEX) xlmExposure() (float64, error) {
	if sdex.cachedXlmExposure != nil {
		return *sdex.cachedXlmExposure, nil
	}

	// uses all offers for this trading account to accommodate sharing by other bots
	offers, err := utils.LoadAllOffers(sdex.TradingAccount, sdex.API)
	if err != nil {
		log.Printf("error computing XLM exposure: %s\n", err)
		return -1, err
	}

	var sum float64
	for _, offer := range offers {
		// only need to compute sum of selling because that's the max XLM we can give up if all our offers are taken
		if offer.Selling.Type == utils.Native {
			offerAmt, err := sdex.ParseOfferAmount(offer.Amount)
			if err != nil {
				return -1, err
			}
			sum += offerAmt
		}
	}

	sdex.cachedXlmExposure = &sum
	return sum, nil
}
