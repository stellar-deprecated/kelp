package plugins

import (
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"time"

	hProtocol "github.com/stellar/go/protocols/horizon"
	"github.com/stellar/go/txnbuild"
	"github.com/stellar/kelp/database"
	"github.com/stellar/kelp/model"
	"github.com/stellar/kelp/support/postgresdb"
	"github.com/stellar/kelp/support/utils"
)

// volumeFilterConfig ensures that any one constraint that is hit will result in deleting all offers and pausing until limits are no longer constrained
type volumeFilterConfig struct {
	sellBaseAssetCapInBaseUnits  *float64
	sellBaseAssetCapInQuoteUnits *float64
	// buyBaseAssetCapInBaseUnits   *float64
	// buyBaseAssetCapInQuoteUnits  *float64
}

type volumeFilter struct {
	baseAsset  hProtocol.Asset
	quoteAsset hProtocol.Asset
	marketID   string
	config     *volumeFilterConfig
	db         *sql.DB
}

// MakeVolumeFilterConfig makes the config for the volume filter
func MakeVolumeFilterConfig(
	sellBaseAssetCapInBaseUnits string,
	sellBaseAssetCapInQuoteUnits string,
	// buyBaseAssetCapInBaseUnits string,
	// buyBaseAssetCapInQuoteUnits string,
) (*volumeFilterConfig, error) {
	sellBaseBase, e := utils.ParseMaybeFloat(sellBaseAssetCapInBaseUnits)
	if e != nil {
		return nil, fmt.Errorf("could not parse sellBaseAssetCapInBaseUnits: %s", e)
	}
	sellBaseQuote, e := utils.ParseMaybeFloat(sellBaseAssetCapInQuoteUnits)
	if e != nil {
		return nil, fmt.Errorf("could not parse sellBaseAssetCapInQuoteUnits: %s", e)
	}
	// buyBaseBase, e := utils.ParseMaybeFloat(buyBaseAssetCapInBaseUnits)
	// if e != nil {
	// 	return nil, fmt.Errorf("could not parse buyBaseAssetCapInBaseUnits: %s", e)
	// }
	// buyBaseQuote, e := utils.ParseMaybeFloat(buyBaseAssetCapInQuoteUnits)
	// if e != nil {
	// 	return nil, fmt.Errorf("could not parse buyBaseAssetCapInQuoteUnits: %s", e)
	// }

	return &volumeFilterConfig{
		sellBaseAssetCapInBaseUnits:  sellBaseBase,
		sellBaseAssetCapInQuoteUnits: sellBaseQuote,
		// buyBaseAssetCapInBaseUnits:   buyBaseBase,
		// buyBaseAssetCapInQuoteUnits:  buyBaseQuote,
	}, nil
}

// MakeFilterVolume makes a submit filter that limits orders placed based on the daily volume traded
func MakeFilterVolume(
	exchangeName string,
	tradingPair *model.TradingPair,
	assetDisplayFn model.AssetDisplayFn,
	baseAsset hProtocol.Asset,
	quoteAsset hProtocol.Asset,
	config *volumeFilterConfig,
	db *sql.DB,
) (SubmitFilter, error) {
	if db == nil {
		return nil, fmt.Errorf("the provided db should be non-nil")
	}

	// use assetDisplayFn to make baseAssetString and quoteAssetString because it is issuer independent for non-sdex exchanges keeping a consistent marketID
	baseAssetString, e := assetDisplayFn(tradingPair.Base)
	if e != nil {
		return nil, fmt.Errorf("could not convert base asset (%s) from trading pair via the passed in assetDisplayFn: %s", string(tradingPair.Base), e)
	}
	quoteAssetString, e := assetDisplayFn(tradingPair.Quote)
	if e != nil {
		return nil, fmt.Errorf("could not convert quote asset (%s) from trading pair via the passed in assetDisplayFn: %s", string(tradingPair.Quote), e)
	}
	marketID := makeMarketID(exchangeName, baseAssetString, quoteAssetString)

	return &volumeFilter{
		baseAsset:  baseAsset,
		quoteAsset: quoteAsset,
		marketID:   marketID,
		config:     config,
		db:         db,
	}, nil
}

var _ SubmitFilter = &volumeFilter{}

func (f *volumeFilter) Apply(ops []txnbuild.Operation, sellingOffers []hProtocol.Offer, buyingOffers []hProtocol.Offer) ([]txnbuild.Operation, error) {
	if f.config.isEmpty() {
		log.Printf("the volumeFilterConfig was empty so not running through the volumeFilter\n")
		return ops, nil
	}

	dateString := time.Now().UTC().Format(postgresdb.DateFormatString)
	// TODO do for buying base and also for a flipped marketID
	dailyValuesBaseSold, e := f.dailyValuesByDate(f.marketID, dateString, "sell")
	if e != nil {
		return nil, fmt.Errorf("could not load dailyValuesByDate for today (%s): %s", dateString, e)
	}

	log.Printf("dailyValuesByDate for today (%s): baseSoldUnits = %.8f %s, quoteCostUnits = %.8f %s (config = %+v)\n",
		dateString, dailyValuesBaseSold.baseVol, utils.Asset2String(f.baseAsset), dailyValuesBaseSold.quoteVol, utils.Asset2String(f.quoteAsset), f.config)

	// daily on-the-books
	dailyOTB := &volumeFilterConfig{
		sellBaseAssetCapInBaseUnits:  &dailyValuesBaseSold.baseVol,
		sellBaseAssetCapInQuoteUnits: &dailyValuesBaseSold.quoteVol,
	}
	// daily to-be-booked starts out as empty and accumulates the values of the operations
	dailyTBB := &volumeFilterConfig{}

	innerFn := func(op *txnbuild.ManageSellOffer) (*txnbuild.ManageSellOffer, bool, error) {
		return f.volumeFilterFn(dailyOTB, dailyTBB, op)
	}
	ops, e = filterOps(ops, innerFn)
	if e != nil {
		return nil, fmt.Errorf("could not apply filter: %s", e)
	}
	return ops, nil
}

func (f *volumeFilter) volumeFilterFn(dailyOTB *volumeFilterConfig, dailyTBB *volumeFilterConfig, op *txnbuild.ManageSellOffer) (*txnbuild.ManageSellOffer, bool, error) {
	// delete operations should never be dropped
	if op.Amount == "0" {
		return op, true, nil
	}

	isSell, e := utils.IsSelling(f.baseAsset, f.quoteAsset, op.Selling, op.Buying)
	if e != nil {
		return nil, false, fmt.Errorf("error when running the isSelling check: %s", e)
	}

	sellPrice, e := strconv.ParseFloat(op.Price, 64)
	if e != nil {
		return nil, false, fmt.Errorf("could not convert price (%s) to float: %s", op.Price, e)
	}

	amountValueUnitsBeingSold, e := strconv.ParseFloat(op.Amount, 64)
	if e != nil {
		return nil, false, fmt.Errorf("could not convert amount (%s) to float: %s", op.Amount, e)
	}
	amountValueUnitsBeingBought := amountValueUnitsBeingSold * sellPrice

	var keep bool
	if isSell {
		var keepSellingBase bool
		var keepSellingQuote bool
		if f.config.sellBaseAssetCapInBaseUnits != nil {
			projectedSoldInBaseUnits := *dailyOTB.sellBaseAssetCapInBaseUnits + *dailyTBB.sellBaseAssetCapInBaseUnits + amountValueUnitsBeingSold
			keepSellingBase := projectedSoldInBaseUnits < *f.config.sellBaseAssetCapInBaseUnits
			log.Printf("volumeFilter: selling (base units), keep = (projectedSoldInBaseUnits) %.7f < %.7f (config.sellBaseAssetCapInBaseUnits): keepSellingBase = %v", projectedSoldInBaseUnits, *f.config.sellBaseAssetCapInBaseUnits, keepSellingBase)
		}

		if f.config.sellBaseAssetCapInQuoteUnits != nil {
			projectedSoldInQuoteUnits := *dailyOTB.sellBaseAssetCapInQuoteUnits + *dailyTBB.sellBaseAssetCapInQuoteUnits + amountValueUnitsBeingBought
			keepSellingQuote = projectedSoldInQuoteUnits < *f.config.sellBaseAssetCapInQuoteUnits
			log.Printf("volumeFilter: selling (quote units), keep = (projectedSoldInQuoteUnits) %.7f < %.7f (config.sellBaseAssetCapInQuoteUnits): keepSellingQuote = %v", projectedSoldInQuoteUnits, *f.config.sellBaseAssetCapInQuoteUnits, keepSellingQuote)
		}

		keep = keepSellingBase && keepSellingQuote
	} else {
		// TODO buying side
	}

	if keep {
		// update the dailyTBB to include the additional amounts so they can be used in the calculation of the next operation
		*dailyTBB.sellBaseAssetCapInBaseUnits += amountValueUnitsBeingSold
		*dailyTBB.sellBaseAssetCapInQuoteUnits += amountValueUnitsBeingBought
		return op, true, nil
	}

	// TODO - reduce amount in offer so we can just meet the capacity limit, instead of dropping
	// convert the offer to a dropped state
	if op.OfferID == 0 {
		// new offers can be dropped
		return nil, false, nil
	} else if op.Amount != "0" {
		// modify offers should be converted to delete offers
		opCopy := *op
		opCopy.Amount = "0"
		return &opCopy, false, nil
	}
	return nil, keep, fmt.Errorf("unable to transform manageOffer operation: offerID=%d, amount=%s, price=%.7f", op.OfferID, op.Amount, sellPrice)
}

func (c *volumeFilterConfig) isEmpty() bool {
	if c.sellBaseAssetCapInBaseUnits != nil {
		return false
	}
	if c.sellBaseAssetCapInQuoteUnits != nil {
		return false
	}
	// if buyBaseAssetCapInBaseUnits != nil {
	// 	return false
	// }
	// if buyBaseAssetCapInQuoteUnits != nil {
	// 	return false
	// }
	return true
}

// dailyValues represents any volume value which can be either bought or sold depending on the query
type dailyValues struct {
	baseVol  float64
	quoteVol float64
}

func (f *volumeFilter) dailyValuesByDate(marketID string, dateUTC string, action string) (*dailyValues, error) {
	row := f.db.QueryRow(database.SqlQueryDailyValues, marketID, dateUTC, action)

	var baseVol sql.NullFloat64
	var quoteVol sql.NullFloat64
	e := row.Scan(&baseVol, &quoteVol)
	if e != nil {
		return nil, fmt.Errorf("could not read data from SqlQueryDailyValues query: %s", e)
	}

	if !baseVol.Valid {
		return nil, fmt.Errorf("baseVol was invalid")
	}
	if !quoteVol.Valid {
		return nil, fmt.Errorf("quoteVol was invalid")
	}

	return &dailyValues{
		baseVol:  baseVol.Float64,
		quoteVol: quoteVol.Float64,
	}, nil
}
