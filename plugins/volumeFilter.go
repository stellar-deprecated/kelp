package plugins

import (
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	hProtocol "github.com/stellar/go/protocols/horizon"
	"github.com/stellar/go/txnbuild"
	"github.com/stellar/kelp/model"
	"github.com/stellar/kelp/queries"
	"github.com/stellar/kelp/support/postgresdb"
	"github.com/stellar/kelp/support/utils"
)

type volumeFilterMode string

// type of volumeFilterMode
const (
	volumeFilterModeExact  volumeFilterMode = "exact"
	volumeFilterModeIgnore volumeFilterMode = "ignore"
)

func parseVolumeFilterMode(mode string) (volumeFilterMode, error) {
	if mode == string(volumeFilterModeExact) {
		return volumeFilterModeExact, nil
	} else if mode == string(volumeFilterModeIgnore) {
		return volumeFilterModeIgnore, nil
	}
	return volumeFilterModeExact, fmt.Errorf("invalid input mode '%s'", mode)
}

// VolumeFilterConfig ensures that any one constraint that is hit will result in deleting all offers and pausing until limits are no longer constrained
type VolumeFilterConfig struct {
	BaseAssetCapInBaseUnits  *float64
	BaseAssetCapInQuoteUnits *float64
	action                   queries.DailyVolumeAction
	mode                     volumeFilterMode
	additionalMarketIDs      []string
	optionalAccountIDs       []string
}

type limitParameters struct {
	baseAssetCapInBaseUnits  *float64
	baseAssetCapInQuoteUnits *float64
	mode                     volumeFilterMode
}

type volumeFilter struct {
	name                   string
	configValue            string
	baseAsset              hProtocol.Asset
	quoteAsset             hProtocol.Asset
	config                 *VolumeFilterConfig
	dailyVolumeByDateQuery *queries.DailyVolumeByDate
}

// makeFilterVolume makes a submit filter that limits orders placed based on the daily volume traded
func makeFilterVolume(
	configValue string,
	exchangeName string,
	tradingPair *model.TradingPair,
	assetDisplayFn model.AssetDisplayFn,
	baseAsset hProtocol.Asset,
	quoteAsset hProtocol.Asset,
	db *sql.DB,
	config *VolumeFilterConfig,
) (SubmitFilter, error) {
	// use assetDisplayFn to make baseAssetString and quoteAssetString because it is issuer independent for non-sdex exchanges keeping a consistent marketID
	baseAssetString, e := assetDisplayFn(tradingPair.Base)
	if e != nil {
		return nil, fmt.Errorf("could not convert base asset (%s) from trading pair via the passed in assetDisplayFn: %s", string(tradingPair.Base), e)
	}
	quoteAssetString, e := assetDisplayFn(tradingPair.Quote)
	if e != nil {
		return nil, fmt.Errorf("could not convert quote asset (%s) from trading pair via the passed in assetDisplayFn: %s", string(tradingPair.Quote), e)
	}

	marketID := MakeMarketID(exchangeName, baseAssetString, quoteAssetString)
	marketIDs := utils.Dedupe(append([]string{marketID}, config.additionalMarketIDs...))
	dailyVolumeByDateQuery, e := queries.MakeDailyVolumeByDateForMarketIdsAction(db, marketIDs, config.action, config.optionalAccountIDs)
	if e != nil {
		return nil, fmt.Errorf("could not make daily volume by date Query: %s", e)
	}

	// TODO DS Validate the config, to have exactly one asset cap defined; a valid mode; non-nil market IDs; and non-nil optional account IDs.

	return &volumeFilter{
		name:                   "volumeFilter",
		configValue:            configValue,
		baseAsset:              baseAsset,
		quoteAsset:             quoteAsset,
		config:                 config,
		dailyVolumeByDateQuery: dailyVolumeByDateQuery,
	}, nil
}

var _ SubmitFilter = &volumeFilter{}

// Validate ensures validity
func (c *VolumeFilterConfig) Validate() error {
	if c.isEmpty() {
		return fmt.Errorf("the volumeFilterConfig was empty")
	}
	return nil
}

// String is the stringer method
func (c *VolumeFilterConfig) String() string {
	return fmt.Sprintf("VolumeFilterConfig[BaseAssetCapInBaseUnits=%s, BaseAssetCapInQuoteUnits=%s, mode=%s, action=%s, additionalMarketIDs=%v, optionalAccountIDs=%v]",
		utils.CheckedFloatPtr(c.BaseAssetCapInBaseUnits), utils.CheckedFloatPtr(c.BaseAssetCapInQuoteUnits), c.mode, c.action, c.additionalMarketIDs, c.optionalAccountIDs)
}

func (f *volumeFilter) Apply(ops []txnbuild.Operation, sellingOffers []hProtocol.Offer, buyingOffers []hProtocol.Offer) ([]txnbuild.Operation, error) {
	dateString := time.Now().UTC().Format(postgresdb.DateFormatString)
	// TODO do for buying base and also for flipped marketIDs
	queryResult, e := f.dailyVolumeByDateQuery.QueryRow(dateString)
	if e != nil {
		return nil, fmt.Errorf("could not load dailyValuesByDate for today (%s): %s", dateString, e)
	}
	dailyValuesBaseSold, ok := queryResult.(*queries.DailyVolume)
	if !ok {
		return nil, fmt.Errorf("incorrect type returned from DailyVolumeByDate query, expecting '*queries.DailyVolume' but was '%T'", queryResult)
	}

	log.Printf("dailyValuesByDate for today (%s): baseSoldUnits = %.8f %s, quoteCostUnits = %.8f %s (%s)\n",
		dateString, dailyValuesBaseSold.BaseVol, utils.Asset2String(f.baseAsset), dailyValuesBaseSold.QuoteVol, utils.Asset2String(f.quoteAsset), f.config)

	// daily on-the-books
	dailyOTB := &VolumeFilterConfig{
		BaseAssetCapInBaseUnits:  &dailyValuesBaseSold.BaseVol,
		BaseAssetCapInQuoteUnits: &dailyValuesBaseSold.QuoteVol,
	}
	// daily to-be-booked starts out as empty and accumulates the values of the operations
	dailyTbbBase := 0.0
	dailyTbbSellQuote := 0.0
	dailyTBB := &VolumeFilterConfig{
		BaseAssetCapInBaseUnits:  &dailyTbbBase,
		BaseAssetCapInQuoteUnits: &dailyTbbSellQuote,
	}

	innerFn := func(op *txnbuild.ManageSellOffer) (*txnbuild.ManageSellOffer, error) {
		limitParameters := limitParameters{
			baseAssetCapInBaseUnits:  f.config.BaseAssetCapInBaseUnits,
			baseAssetCapInQuoteUnits: f.config.BaseAssetCapInQuoteUnits,
			mode:                     f.config.mode,
		}
		return volumeFilterFn(dailyOTB, dailyTBB, op, f.baseAsset, f.quoteAsset, limitParameters)
	}
	ops, e = filterOps(f.name, f.baseAsset, f.quoteAsset, sellingOffers, buyingOffers, ops, innerFn)
	if e != nil {
		return nil, fmt.Errorf("could not apply filter: %s", e)
	}
	return ops, nil
}

func volumeFilterFn(dailyOTB *VolumeFilterConfig, dailyTBBAccumulator *VolumeFilterConfig, op *txnbuild.ManageSellOffer, baseAsset hProtocol.Asset, quoteAsset hProtocol.Asset, lp limitParameters) (*txnbuild.ManageSellOffer, error) {
	isSell, e := utils.IsSelling(baseAsset, quoteAsset, op.Selling, op.Buying)
	if e != nil {
		return nil, fmt.Errorf("error when running the isSelling check for offer '%+v': %s", *op, e)
	}

	sellPrice, e := strconv.ParseFloat(op.Price, 64)
	if e != nil {
		return nil, fmt.Errorf("could not convert price (%s) to float: %s", op.Price, e)
	}

	amountValueUnitsBeingSold, e := strconv.ParseFloat(op.Amount, 64)
	if e != nil {
		return nil, fmt.Errorf("could not convert amount (%s) to float: %s", op.Amount, e)
	}

	if isSell {
		opToReturn := op
		newAmountBeingSold := amountValueUnitsBeingSold
		var keepSellingBase bool
		var keepSellingQuote bool
		if lp.baseAssetCapInBaseUnits != nil {
			projectedSoldInBaseUnits := *dailyOTB.BaseAssetCapInBaseUnits + *dailyTBBAccumulator.BaseAssetCapInBaseUnits + amountValueUnitsBeingSold
			keepSellingBase = projectedSoldInBaseUnits <= *lp.baseAssetCapInBaseUnits
			newAmountString := ""
			if lp.mode == volumeFilterModeExact && !keepSellingBase {
				newAmount := *lp.baseAssetCapInBaseUnits - *dailyOTB.BaseAssetCapInBaseUnits - *dailyTBBAccumulator.BaseAssetCapInBaseUnits
				if newAmount > 0 {
					newAmountBeingSold = newAmount
					opToReturn.Amount = fmt.Sprintf("%.7f", newAmountBeingSold)
					keepSellingBase = true
					newAmountString = ", newAmountString = " + opToReturn.Amount
				}
			}
			log.Printf("volumeFilter:  selling (base units), price=%.8f amount=%.8f, keep = (projectedSoldInBaseUnits) %.7f <= %.7f (config.BaseAssetCapInBaseUnits): keepSellingBase = %v%s", sellPrice, amountValueUnitsBeingSold, projectedSoldInBaseUnits, *lp.baseAssetCapInBaseUnits, keepSellingBase, newAmountString)
		} else {
			keepSellingBase = true
		}

		if lp.baseAssetCapInQuoteUnits != nil {
			projectedSoldInQuoteUnits := *dailyOTB.BaseAssetCapInQuoteUnits + *dailyTBBAccumulator.BaseAssetCapInQuoteUnits + (newAmountBeingSold * sellPrice)
			keepSellingQuote = projectedSoldInQuoteUnits <= *lp.baseAssetCapInQuoteUnits
			newAmountString := ""
			if lp.mode == volumeFilterModeExact && !keepSellingQuote {
				newAmount := (*lp.baseAssetCapInQuoteUnits - *dailyOTB.BaseAssetCapInQuoteUnits - *dailyTBBAccumulator.BaseAssetCapInQuoteUnits) / sellPrice
				if newAmount > 0 {
					newAmountBeingSold = newAmount
					opToReturn.Amount = fmt.Sprintf("%.7f", newAmountBeingSold)
					keepSellingQuote = true
					newAmountString = ", newAmountString = " + opToReturn.Amount
				}
			}
			log.Printf("volumeFilter: selling (quote units), price=%.8f amount=%.8f, keep = (projectedSoldInQuoteUnits) %.7f <= %.7f (config.BaseAssetCapInQuoteUnits): keepSellingQuote = %v%s", sellPrice, amountValueUnitsBeingSold, projectedSoldInQuoteUnits, *lp.baseAssetCapInQuoteUnits, keepSellingQuote, newAmountString)
		} else {
			keepSellingQuote = true
		}

		if keepSellingBase && keepSellingQuote {
			// update the dailyTBB to include the additional amounts so they can be used in the calculation of the next operation
			*dailyTBBAccumulator.BaseAssetCapInBaseUnits += newAmountBeingSold
			*dailyTBBAccumulator.BaseAssetCapInQuoteUnits += (newAmountBeingSold * sellPrice)
			return opToReturn, nil
		}
	} else {
		// TODO buying side - we need to implement this to support buy side filters; extract common logic from the above sell side case
	}

	// we don't want to keep it so return the dropped command
	return nil, nil
}

// String is the Stringer method
func (f *volumeFilter) String() string {
	return f.configValue
}

// isBase returns true if the filter is on the amount of the base asset sold, false otherwise
func (f *volumeFilter) isSellingBase() bool {
	return strings.Contains(f.configValue, "/sell/base/")
}

func (f *volumeFilter) mustGetBaseAssetCapInBaseUnits() (float64, error) {
	value := f.config.BaseAssetCapInBaseUnits
	if value == nil {
		return 0.0, fmt.Errorf("BaseAssetCapInBaseUnits is nil, config = %v", f.config)
	}
	return *value, nil
}

func (c *VolumeFilterConfig) isEmpty() bool {
	if c.BaseAssetCapInBaseUnits != nil {
		return false
	}
	if c.BaseAssetCapInQuoteUnits != nil {
		return false
	}
	return true
}
