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

// String is the Stringer method
func (v volumeFilterMode) String() string {
	return string(v)
}

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
	additionalMarketIDs      []string // can be nil
	optionalAccountIDs       []string // can be nil
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
	// note that append(s, nil) is valid
	marketIDs := utils.Dedupe(append([]string{marketID}, config.additionalMarketIDs...))
	dailyVolumeByDateQuery, e := queries.MakeDailyVolumeByDateForMarketIdsAction(db, marketIDs, config.action, config.optionalAccountIDs)
	if e != nil {
		return nil, fmt.Errorf("could not make daily volume by date Query: %s", e)
	}

	e = config.Validate()
	if e != nil {
		return nil, fmt.Errorf("invalid config: %s", e)
	}

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
	if c.BaseAssetCapInBaseUnits != nil && c.BaseAssetCapInQuoteUnits != nil {
		return fmt.Errorf("invalid asset caps: only one asset cap can be non-nil, but both are non-nil")
	}

	if c.BaseAssetCapInBaseUnits == nil && c.BaseAssetCapInQuoteUnits == nil {
		return fmt.Errorf("invalid asset caps: only one asset cap can be non-nil, but both are nil")
	}

	if _, e := parseVolumeFilterMode(string(c.mode)); e != nil {
		return fmt.Errorf("could not parse mode: %s", e)
	}

	if _, e := queries.ParseDailyVolumeAction(string(c.action)); e != nil {
		return fmt.Errorf("could not parse action: %s", e)
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
	// TODO for flipped marketIDs
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
	dailyOTB := makeIntermediateVolumeFilterConfig(&dailyValuesBaseSold.BaseVol, &dailyValuesBaseSold.QuoteVol)
	// daily to-be-booked starts out as empty and accumulates the values of the operations
	dailyTbbBase := 0.0
	dailyTbbSellQuote := 0.0
	dailyTBB := makeIntermediateVolumeFilterConfig(&dailyTbbBase, &dailyTbbSellQuote)

	innerFn := func(op *txnbuild.ManageSellOffer) (*txnbuild.ManageSellOffer, error) {
		limitParameters := limitParameters{
			baseAssetCapInBaseUnits:  f.config.BaseAssetCapInBaseUnits,
			baseAssetCapInQuoteUnits: f.config.BaseAssetCapInQuoteUnits,
			mode:                     f.config.mode,
		}
		return volumeFilterFn(f.config.action, dailyOTB, dailyTBB, op, f.baseAsset, f.quoteAsset, limitParameters)
	}
	ops, e = filterOps(f.name, f.baseAsset, f.quoteAsset, sellingOffers, buyingOffers, ops, innerFn)
	if e != nil {
		return nil, fmt.Errorf("could not apply filter: %s", e)
	}
	return ops, nil
}

func makeIntermediateVolumeFilterConfig(baseCapBaseUnits *float64, baseCapQuoteUnits *float64) *VolumeFilterConfig {
	return &VolumeFilterConfig{
		BaseAssetCapInBaseUnits:  baseCapBaseUnits,
		BaseAssetCapInQuoteUnits: baseCapQuoteUnits,
	}
}

func volumeFilterFn(action queries.DailyVolumeAction, dailyOTB *VolumeFilterConfig, dailyTBBAccumulator *VolumeFilterConfig, op *txnbuild.ManageSellOffer, baseAsset hProtocol.Asset, quoteAsset hProtocol.Asset, lp limitParameters) (*txnbuild.ManageSellOffer, error) {
	isFilterApplicable, e := offerSameTypeAsFilter(action, op, baseAsset, quoteAsset)
	if e != nil {
		return nil, fmt.Errorf("could not compare offer and filter: %s", e)
	}

	if !isFilterApplicable {
		// ignore filter so return op directly
		log.Printf("volumeFilter: isSell=%v, isFilterApplicable=false; keep=true", action.IsSell())
		return op, nil
	}

	// extract offer price and amount and adjust for buy offers
	offerPrice, e := strconv.ParseFloat(op.Price, 64)
	if e != nil {
		return nil, fmt.Errorf("could not convert price (%s) to float: %s", op.Price, e)
	}
	offerAmount, e := strconv.ParseFloat(op.Amount, 64)
	if e != nil {
		return nil, fmt.Errorf("could not convert amount (%s) to float: %s", op.Amount, e)
	}
	// A "buy" op has amount = sellAmount * sellPrice, and price = 1/sellPrice
	// So, we adjust the offer variables by "undoing" those adjustments
	// We can then use the same computations as sell orders on buy orders
	if action.IsBuy() {
		offerAmount = offerAmount * offerPrice
		offerPrice = 1 / offerPrice
	}

	// capPrice is used when computing amounts to sell or buy
	// it's the offer price when capping on quote, and 1.0 when capping on base
	capPrice := offerPrice
	if lp.baseAssetCapInBaseUnits != nil {
		capPrice = 1.0
	}

	// extracts from base or quote side, depending on filter
	otb, tbb, cap, e := extractAllCaps(dailyOTB, dailyTBBAccumulator, lp)
	if e != nil {
		return nil, fmt.Errorf("could not extract filter inputs from filter: %s", e)
	}

	// if projected is under the cap, update the tbb and return the original op
	projected := otb + tbb + offerAmount*capPrice
	if projected <= cap {
		dailyTBBAccumulator = updateTBB(dailyTBBAccumulator, offerAmount, offerPrice)
		log.Printf("volumeFilter: isSell=%v, offerPrice=%.10f, projected (%.10f) <= cap (%.10f); keep=true", action.IsSell(), offerPrice, projected, cap)
		return op, nil
	}

	// for ignore type of filters we want to drop the operations when the cap is exceeded
	if lp.mode == volumeFilterModeIgnore {
		log.Printf("volumeFilter: isSell=%v, offerPrice=%.10f; lp.mode=%s, keep=false", action.IsSell(), offerPrice, lp.mode.String())
		return nil, nil
	}

	// if exact mode and with remaining capacity, update the op amount and return the op otherwise return nil
	newOfferAmount := (cap - otb - tbb) / capPrice
	if newOfferAmount <= 0 {
		log.Printf("volumeFilter: isSell=%v, offerPrice=%.10f, newOfferAmount (%.10f) <= 0; keep=false", action.IsSell(), offerPrice, newOfferAmount)
		return nil, nil
	}
	dailyTBBAccumulator = updateTBB(dailyTBBAccumulator, newOfferAmount, offerPrice)
	// if we have a buy operation, we want to make sure buy ops have the same relationship between price and amount
	// to do this, we apply the same amount adjustment as `makeBuyOpAmtPrice`
	// The following conversion is done above on input:
	// sellOfferAmount = buyOfferAmount * buyOfferPrice
	// sellOfferPrice = 1 / buyOfferPrice
	//
	// Therefore we need to undo it using the following:
	// newOpAmount = newOpAmount * sellOfferPrice
	// newOpAmount => newOpAmount * 1 / buyOfferPrice
	newOpAmount := newOfferAmount
	if action.IsBuy() {
		newOpAmount = newOpAmount * offerPrice
	}
	op.Amount = fmt.Sprintf("%.7f", newOpAmount)

	log.Printf("volumeFilter: isSell=%v, offerPrice=%.10f, newOpAmount=%s; keep=true", action.IsSell(), offerPrice, op.Amount)
	return op, nil
}

func offerSameTypeAsFilter(action queries.DailyVolumeAction, op *txnbuild.ManageSellOffer, baseAsset hProtocol.Asset, quoteAsset hProtocol.Asset) (bool, error) {
	opIsSelling, e := utils.IsSelling(baseAsset, quoteAsset, op.Selling, op.Buying)
	if e != nil {
		return false, fmt.Errorf("error when running the isSelling check for offer '%+v': %s", *op, e)
	}
	isSame := opIsSelling == action.IsSell()
	log.Printf("volumeFilter: opIsSelling (%v) == filter.action.IsSell() (%v); isSame = %v", opIsSelling, action.IsSell(), isSame)
	return isSame, nil
}

// extractAllCaps will extract caps from both filters and the limit parameters
func extractAllCaps(dailyOTB *VolumeFilterConfig, dailyTBB *VolumeFilterConfig, lp limitParameters) (float64 /* otbCap */, float64 /* tbbCap */, float64 /* lpCap */, error) {
	if lp.baseAssetCapInBaseUnits != nil {
		return *dailyOTB.BaseAssetCapInBaseUnits, *dailyTBB.BaseAssetCapInBaseUnits, *lp.baseAssetCapInBaseUnits, nil
	}

	if lp.baseAssetCapInQuoteUnits != nil {
		return *dailyOTB.BaseAssetCapInQuoteUnits, *dailyTBB.BaseAssetCapInQuoteUnits, *lp.baseAssetCapInQuoteUnits, nil
	}

	// should never reach this code - means that the configs were not validated properly
	return -1, -1, -1, fmt.Errorf("found two nil filters")
}

func updateTBB(tbb *VolumeFilterConfig, amount float64, price float64) *VolumeFilterConfig {
	*tbb.BaseAssetCapInBaseUnits += amount
	*tbb.BaseAssetCapInQuoteUnits += amount * price
	return tbb
}

// String is the Stringer method
func (f *volumeFilter) String() string {
	return f.configValue
}

// isBase returns true if the filter is on the amount of the base asset sold, false otherwise
func (f *volumeFilter) isBase() bool {
	return strings.Contains(f.configValue, "/base/")
}

func (f *volumeFilter) mustGetBaseAssetCapInBaseUnits() (float64, error) {
	value := f.config.BaseAssetCapInBaseUnits
	if value == nil {
		return 0.0, fmt.Errorf("BaseAssetCapInBaseUnits is nil, config = %v", f.config)
	}
	return *value, nil
}
