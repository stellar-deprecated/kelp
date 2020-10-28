package plugins

import (
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"

	"github.com/nikhilsaraf/go-tools/multithreading"

	"github.com/stellar/go/build"
	hProtocol "github.com/stellar/go/protocols/horizon"
	"github.com/stellar/go/txnbuild"
	"github.com/stellar/kelp/api"
	"github.com/stellar/kelp/kelpdb"
	"github.com/stellar/kelp/model"
	"github.com/stellar/kelp/queries"
	"github.com/stellar/kelp/support/toml"
	"github.com/stellar/kelp/support/utils"
)

// numOrdersBufferMinVolumeFilter is the number of extra orders we want to fetch from the exchange in addition to the configured OrderbookDepth
// to allow us to account for any skipped orders because of min base volume requirements
const numOrdersBufferMinVolumeFilter = 3
const debugLogOffersOrders = true

const maxOrderbookDepth int32 = 50

// mirrorConfig contains the configuration params for this strategy
type mirrorConfig struct {
	Exchange       string `valid:"-" toml:"EXCHANGE"`
	ExchangeBase   string `valid:"-" toml:"EXCHANGE_BASE"`
	ExchangeQuote  string `valid:"-" toml:"EXCHANGE_QUOTE"`
	OrderbookDepth int    `valid:"-" toml:"ORDERBOOK_DEPTH"`
	// Deprecated: use BID_VOLUME_DIVIDE_BY and ASK_VOLUME_DIVIDE_BY instead
	VolumeDivideByDeprecated *float64 `valid:"-" toml:"VOLUME_DIVIDE_BY" deprecated:"true"`
	BidVolumeDivideBy        *float64 `valid:"-" toml:"BID_VOLUME_DIVIDE_BY"`
	AskVolumeDivideBy        *float64 `valid:"-" toml:"ASK_VOLUME_DIVIDE_BY"`
	MaxOrderBaseCap          *float64 `valid:"-" toml:"MAX_ORDER_BASE_CAP"` // use a pointer here so we don't need to special case 0.0 everywhere and a nil value is clearly not user-entered
	PerLevelSpread           float64  `valid:"-" toml:"PER_LEVEL_SPREAD"`
	PricePrecisionOverride   *int8    `valid:"-" toml:"PRICE_PRECISION_OVERRIDE"`
	VolumePrecisionOverride  *int8    `valid:"-" toml:"VOLUME_PRECISION_OVERRIDE"`
	// Deprecated: use MIN_BASE_VOLUME_OVERRIDE instead
	MinBaseVolumeDeprecated                   *float64                 `valid:"-" toml:"MIN_BASE_VOLUME" deprecated:"true"`
	MinBaseVolumeOverride                     *float64                 `valid:"-" toml:"MIN_BASE_VOLUME_OVERRIDE"`
	MinQuoteVolumeOverride                    *float64                 `valid:"-" toml:"MIN_QUOTE_VOLUME_OVERRIDE"`
	OffsetTrades                              bool                     `valid:"-" toml:"OFFSET_TRADES"`
	BackingDbOverrideAccountID                string                   `valid:"-" toml:"BACKING_DB_OVERRIDE__ACCOUNT_ID"`
	BackingFillTrackerLastTradeCursorOverride string                   `valid:"-" toml:"BACKING_FILL_TRACKER_LAST_TRADE_CURSOR_OVERRIDE"`
	ExchangeAPIKeys                           toml.ExchangeAPIKeysToml `valid:"-" toml:"EXCHANGE_API_KEYS"`
	ExchangeParams                            toml.ExchangeParamsToml  `valid:"-" toml:"EXCHANGE_PARAMS"`
	ExchangeHeaders                           toml.ExchangeHeadersToml `valid:"-" toml:"EXCHANGE_HEADERS"`
}

// String impl.
func (c mirrorConfig) String() string {
	return utils.StructString(c, 0, map[string]func(interface{}) interface{}{
		"EXCHANGE_API_KEYS": utils.Hide,
		"EXCHANGE_PARAMS":   utils.Hide,
		"EXCHANGE_HEADERS":  utils.Hide,
	})
}

// assetSurplus holds information about how many units of an asset needs to be offset on the exchange
// negative values mean we have eagerly offset an asset, likely because of minBaseVolume requirements of the backingExchange
type assetSurplus struct {
	total     *model.Number // total value in base asset units that are pending to be offset
	committed *model.Number // base asset units that are already committed to being offset
}

// makeAssetSurplus is a factory method
func makeAssetSurplus() *assetSurplus {
	return &assetSurplus{
		total:     model.NumberConstants.Zero,
		committed: model.NumberConstants.Zero,
	}
}

// mirrorStrategy is a strategy to mirror the orderbook of a given exchange
type mirrorStrategy struct {
	sdex                                  *SDEX
	ieif                                  *IEIF
	baseAsset                             *hProtocol.Asset
	quoteAsset                            *hProtocol.Asset
	primaryConstraints                    *model.OrderConstraints
	marketID                              string
	backingPair                           *model.TradingPair
	backingConstraints                    *model.OrderConstraints
	backingMarketID                       string
	backingFillTracker                    api.FillTracker
	strategyMirrorTradeTriggerExistsQuery *queries.StrategyMirrorTradeTriggerExists
	orderbookDepth                        int
	perLevelSpread                        float64
	bidVolumeDivideBy                     float64
	askVolumeDivideBy                     float64
	maybeMaxOrderBaseCap                  *float64 // using a nil value makes it clear whether this value exists or not
	exchange                              api.Exchange
	offsetTrades                          bool
	mutex                                 *sync.Mutex
	baseSurplus                           map[model.OrderAction]*assetSurplus // baseSurplus keeps track of any surplus we have of the base asset that needs to be offset on the backing exchange
	db                                    *sql.DB

	// uninitialized
	sellOnPrimaryBalanceCoordinator *balanceCoordinator
	buyOnPrimaryBalanceCoordinator  *balanceCoordinator
}

// ensure this implements api.Strategy
var _ api.Strategy = &mirrorStrategy{}

// ensure this implements api.FillHandler
var _ api.FillHandler = &mirrorStrategy{}

func convertDeprecatedMirrorConfigValues(config *mirrorConfig) {
	if config.MinBaseVolumeOverride != nil && config.MinBaseVolumeDeprecated != nil {
		log.Printf("deprecation warning: cannot set both '%s' (deprecated) and '%s' in the mirror strategy config, using value from '%s'\n", "MIN_BASE_VOLUME", "MIN_BASE_VOLUME_OVERRIDE", "MIN_BASE_VOLUME_OVERRIDE")
	} else if config.MinBaseVolumeDeprecated != nil {
		log.Printf("deprecation warning: '%s' is deprecated, use the field '%s' in the mirror strategy config instead, see sample_mirror.cfg as an example\n", "MIN_BASE_VOLUME", "MIN_BASE_VOLUME_OVERRIDE")
	}
	if config.MinBaseVolumeOverride == nil {
		config.MinBaseVolumeOverride = config.MinBaseVolumeDeprecated
	}

	if (config.BidVolumeDivideBy != nil || config.AskVolumeDivideBy != nil) && config.VolumeDivideByDeprecated != nil {
		log.Printf("deprecation warning: cannot set both '%s' (deprecated) and ('%s' / '%s') in the mirror strategy config, overriding with values set from '%s' and '%s'\n", "VOLUME_DIVIDE_BY", "BID_VOLUME_DIVIDE_BY", "ASK_VOLUME_DIVIDE_BY", "BID_VOLUME_DIVIDE_BY", "ASK_VOLUME_DIVIDE_BY")
	} else if config.VolumeDivideByDeprecated != nil {
		log.Printf("deprecation warning: '%s' is deprecated, use the fields '%s' and '%s' in the mirror strategy config instead, see sample_mirror.cfg as an example\n", "VOLUME_DIVIDE_BY", "BID_VOLUME_DIVIDE_BY", "ASK_VOLUME_DIVIDE_BY")
	}
	// if only one is specified, we will use the deprecated value for the unspecified value right now
	if config.BidVolumeDivideBy == nil {
		config.BidVolumeDivideBy = config.VolumeDivideByDeprecated
	}
	if config.AskVolumeDivideBy == nil {
		config.AskVolumeDivideBy = config.VolumeDivideByDeprecated
	}
}

// makeMirrorStrategy is a factory method
func makeMirrorStrategy(
	sdex *SDEX,
	ieif *IEIF,
	pair *model.TradingPair,
	baseAsset *hProtocol.Asset,
	quoteAsset *hProtocol.Asset,
	marketID string,
	config *mirrorConfig,
	db *sql.DB,
	simMode bool,
) (api.Strategy, error) {
	convertDeprecatedMirrorConfigValues(config)
	var bidVolumeDivideBy float64
	var askVolumeDivideBy float64
	if config.BidVolumeDivideBy == nil {
		bidVolumeDivideBy = 1.0
	} else {
		bidVolumeDivideBy = *config.BidVolumeDivideBy
	}
	if config.AskVolumeDivideBy == nil {
		askVolumeDivideBy = 1.0
	} else {
		askVolumeDivideBy = *config.AskVolumeDivideBy
	}
	if bidVolumeDivideBy == -1.0 && askVolumeDivideBy == -1.0 {
		utils.PrintErrorHintf("both BID_VOLUME_DIVIDE_BY and ASK_VOLUME_DIVIDE_BY cannot be -1.0")
		return nil, fmt.Errorf("invalid mirror strategy config file, cannot set both BID_VOLUME_DIVIDE_BY and ASK_VOLUME_DIVIDE_BY to -1.0")
	}
	if bidVolumeDivideBy != -1.0 && bidVolumeDivideBy <= 0 {
		utils.PrintErrorHintf("need to set a valid value for BID_VOLUME_DIVIDE_BY, needs to be -1.0 or > 0")
		return nil, fmt.Errorf("invalid mirror strategy config file, BID_VOLUME_DIVIDE_BY needs to be -1.0 or > 0")
	}
	if askVolumeDivideBy != -1.0 && askVolumeDivideBy <= 0 {
		utils.PrintErrorHintf("need to set a valid value for ASK_VOLUME_DIVIDE_BY, needs to be -1.0 or > 0")
		return nil, fmt.Errorf("invalid mirror strategy config file, ASK_VOLUME_DIVIDE_BY needs to be -1.0 or > 0")
	}

	var exchange api.Exchange
	var e error
	var strategyMirrorTradeTriggerExistsQuery *queries.StrategyMirrorTradeTriggerExists
	if config.OffsetTrades {
		if db == nil {
			return nil, fmt.Errorf("db should not be nil when OffsetTrades is enabled")
		}

		exchangeAPIKeys := config.ExchangeAPIKeys.ToExchangeAPIKeys()
		exchangeParams := config.ExchangeParams.ToExchangeParams()
		exchangeHeaders := config.ExchangeHeaders.ToExchangeHeaders()
		exchange, e = MakeTradingExchange(config.Exchange, exchangeAPIKeys, exchangeParams, exchangeHeaders, simMode)
		if e != nil {
			return nil, e
		}

		if config.MinBaseVolumeOverride != nil && *config.MinBaseVolumeOverride <= 0.0 {
			return nil, fmt.Errorf("need to specify positive MIN_BASE_VOLUME_OVERRIDE config param in mirror strategy config file")
		}
		if config.MinQuoteVolumeOverride != nil && *config.MinQuoteVolumeOverride <= 0.0 {
			return nil, fmt.Errorf("need to specify positive MIN_QUOTE_VOLUME_OVERRIDE config param in mirror strategy config file")
		}
		if config.VolumePrecisionOverride != nil && *config.VolumePrecisionOverride < 0 {
			return nil, fmt.Errorf("need to specify non-negative VOLUME_PRECISION_OVERRIDE config param in mirror strategy config file")
		}
		if config.PricePrecisionOverride != nil && *config.PricePrecisionOverride < 0 {
			return nil, fmt.Errorf("need to specify non-negative PRICE_PRECISION_OVERRIDE config param in mirror strategy config file")
		}
		if config.BackingDbOverrideAccountID == "" {
			utils.PrintErrorHintf("BACKING_DB_OVERRIDE__ACCOUNT_ID needs to be set in the mirror strategy config file when OFFSET_TRADES is enabled so we can assign an account_id to trades that are fetched from the backing exchange before writing them in the db")
			return nil, fmt.Errorf("invalid mirror strategy config file, need to set BACKING_DB_OVERRIDE__ACCOUNT_ID")
		}

		strategyMirrorTradeTriggerExistsQuery, e = queries.MakeStrategyMirrorTradeTriggerExists(db, marketID)
		if e != nil {
			return nil, fmt.Errorf("unable to create strategyMirrorTradeTriggerExistsQuery: %s", e)
		}
	} else {
		exchange, e = MakeExchange(config.Exchange, simMode)
		if e != nil {
			return nil, e
		}
	}

	// we have two sets of (tradingPair, orderConstraints): the primaryExchange and the backingExchange
	primaryConstraints := sdex.GetOrderConstraints(pair)
	// backingPair is taken from the mirror strategy config not from the passed in trading pair
	backingPair := &model.TradingPair{
		Base:  exchange.GetAssetConverter().MustFromString(config.ExchangeBase),
		Quote: exchange.GetAssetConverter().MustFromString(config.ExchangeQuote),
	}

	// make fill tracker for backing exchange
	var backingFillTracker api.FillTracker
	if config.OffsetTrades {
		var backingLastCursor interface{}
		if config.BackingFillTrackerLastTradeCursorOverride == "" {
			// loads cursor by fetching from exchange
			backingLastCursor, e = exchange.GetLatestTradeCursor()
			if e != nil {
				return nil, fmt.Errorf("could not get last trade cursor from backing exchange in mirrorStrategy: %s", e)
			}
			log.Printf("set backingLastCursor from where to start tracking fills for backing exchange in mirror strategy (no override specified): %v\n", backingLastCursor)
		} else {
			// loads cursor from config file
			backingLastCursor = config.BackingFillTrackerLastTradeCursorOverride
			log.Printf("set backingLastCursor from where to start tracking fills for backing exchange in mirror strategy (used override value): %v\n", backingLastCursor)
		}
		backingFillTracker = MakeFillTracker(backingPair, multithreading.MakeThreadTracker(), exchange, 0, 0, backingLastCursor)
		backingFillTracker.RegisterHandler(MakeFillLogger())
		backingAssetDisplayFn := model.MakePassthroughAssetDisplayFn()
		if config.Exchange == "sdex" {
			return nil, fmt.Errorf("we cannot mirror trades from SDEX for now (programmer: need to create sdexAssetMap to inject into the backingAssetDisplayFn)")
		}
		fillDBWriter := MakeFillDBWriter(db, backingAssetDisplayFn, config.Exchange, config.BackingDbOverrideAccountID)
		backingFillTracker.RegisterHandler(fillDBWriter)
	}

	// update precision overrides
	exchange.OverrideOrderConstraints(backingPair, model.MakeOrderConstraintsOverride(
		config.PricePrecisionOverride,
		config.VolumePrecisionOverride,
		nil,
		nil,
	))
	if config.MinBaseVolumeOverride != nil {
		// use updated precision overrides to convert the minBaseVolume to a model.Number
		exchange.OverrideOrderConstraints(backingPair, model.MakeOrderConstraintsOverride(
			nil,
			nil,
			model.NumberFromFloat(*config.MinBaseVolumeOverride, exchange.GetOrderConstraints(backingPair).VolumePrecision),
			nil,
		))
	}
	if config.MinQuoteVolumeOverride != nil {
		// use updated precision overrides to convert the minQuoteVolume to a model.Number
		minQuoteVolume := model.NumberFromFloat(*config.MinQuoteVolumeOverride, exchange.GetOrderConstraints(backingPair).VolumePrecision)
		exchange.OverrideOrderConstraints(backingPair, model.MakeOrderConstraintsOverride(
			nil,
			nil,
			nil,
			&minQuoteVolume,
		))
	}
	backingConstraints := exchange.GetOrderConstraints(backingPair)
	log.Printf("primaryPair='%s', primaryConstraints=%s\n", pair, primaryConstraints)
	log.Printf("backingPair='%s', backingConstraints=%s\n", backingPair, backingConstraints)
	if config.MaxOrderBaseCap != nil {
		if *config.MaxOrderBaseCap < backingConstraints.MinBaseVolume.AsFloat() {
			utils.PrintErrorHintf("MAX_ORDER_BASE_CAP (%f) cannot be less than minBaseVolume allowed on backing exchange (%s)", *config.MaxOrderBaseCap, backingConstraints.MinBaseVolume.AsString())
			return nil, fmt.Errorf("MAX_ORDER_BASE_CAP (%f) cannot be less than minBaseVolume allowed on backing exchange (%s)", *config.MaxOrderBaseCap, backingConstraints.MinBaseVolume.AsString())
		}
		if *config.MaxOrderBaseCap <= 0.0 {
			utils.PrintErrorHintf("invalid mirror strategy config file, if you set a value for MAX_ORDER_BASE_CAP it needs to be > 0.0, leaving it unset does not constrain the order size")
			return nil, fmt.Errorf("invalid mirror strategy config file, if you set a value for MAX_ORDER_BASE_CAP it needs to be > 0.0, leaving it unset does not constrain the order size")
		}
	}

	// insert into database if needed
	var backingMarketID string
	if db != nil {
		backingMarketID, e = FetchOrRegisterMarketID(db, config.Exchange, config.ExchangeBase, config.ExchangeQuote)
		if e != nil {
			return nil, fmt.Errorf("error calling FetchOrRegisterMarketID: %s", e)
		}
	}

	// trigger fill tracking on backing exchange at creation time
	if backingFillTracker != nil {
		trades, e := backingFillTracker.FillTrackSingleIteration()
		if e != nil {
			return nil, fmt.Errorf("unable to track a single iteration of fills from the backing exchange in factory method: %s", e)
		}

		log.Printf("found %d trades on first load from backing exchange\n", len(trades))
	} else {
		log.Printf("backingFillTracker was nil so not loading trades at creation time\n")
	}

	if config.OrderbookDepth > int(maxOrderbookDepth) {
		return nil, fmt.Errorf("cannot construct the mirrorStrategy, ORDERBOOK_DEPTH config param should not exceed %d", maxOrderbookDepth)
	}

	return &mirrorStrategy{
		sdex:                                  sdex,
		ieif:                                  ieif,
		baseAsset:                             baseAsset,
		quoteAsset:                            quoteAsset,
		primaryConstraints:                    primaryConstraints,
		marketID:                              marketID,
		backingPair:                           backingPair,
		backingConstraints:                    backingConstraints,
		backingMarketID:                       backingMarketID,
		backingFillTracker:                    backingFillTracker,
		strategyMirrorTradeTriggerExistsQuery: strategyMirrorTradeTriggerExistsQuery,
		orderbookDepth:                        config.OrderbookDepth,
		perLevelSpread:                        config.PerLevelSpread,
		bidVolumeDivideBy:                     bidVolumeDivideBy,
		askVolumeDivideBy:                     askVolumeDivideBy,
		maybeMaxOrderBaseCap:                  config.MaxOrderBaseCap,
		exchange:                              exchange,
		offsetTrades:                          config.OffsetTrades,
		mutex:                                 &sync.Mutex{},
		baseSurplus: map[model.OrderAction]*assetSurplus{
			model.OrderActionBuy:  makeAssetSurplus(),
			model.OrderActionSell: makeAssetSurplus(),
		},
		db: db,
	}, nil
}

// PruneExistingOffers deletes any extra offers
func (s *mirrorStrategy) PruneExistingOffers(buyingAOffers []hProtocol.Offer, sellingAOffers []hProtocol.Offer) ([]build.TransactionMutator, []hProtocol.Offer, []hProtocol.Offer) {
	return []build.TransactionMutator{}, buyingAOffers, sellingAOffers
}

// PreUpdate changes the strategy's state in prepration for the update
func (s *mirrorStrategy) PreUpdate(maxAssetA float64, maxAssetB float64, trustA float64, trustB float64) error {
	// we don't care about or use balance coordinators if we are not offsetting trades
	if !s.offsetTrades {
		return nil
	}

	baseBackingBalance, quoteBackingBalance, e := s.getBackingBalances()
	if e != nil {
		return fmt.Errorf("error while fetching backing balances: %s", e)
	}

	// buyOnPrimaryBalanceCoordinator is buying on the primary exchange and selling on the backing exchange
	// primary asset being sold here is quote and backing asset being sold is base, so constrain on those
	s.buyOnPrimaryBalanceCoordinator = &balanceCoordinator{
		primaryBalance:     model.NumberFromFloat(maxAssetB, s.primaryConstraints.VolumePrecision),
		placedPrimaryUnits: model.NumberConstants.Zero,
		primaryAssetType:   "quote",
		isPrimaryBuy:       true,
		backingBalance:     baseBackingBalance,
		placedBackingUnits: model.NumberConstants.Zero,
		backingAssetType:   "base",
	}

	// sellOnPrimaryBalanceCoordinator is selling on the primary exchange and buying on the backing exchange
	// primary asset being sold here is base and backing asset being sold is quote, so constrain on those
	s.sellOnPrimaryBalanceCoordinator = &balanceCoordinator{
		primaryBalance:     model.NumberFromFloat(maxAssetA, s.primaryConstraints.VolumePrecision),
		placedPrimaryUnits: model.NumberConstants.Zero,
		primaryAssetType:   "base",
		isPrimaryBuy:       false,
		backingBalance:     quoteBackingBalance,
		placedBackingUnits: model.NumberConstants.Zero,
		backingAssetType:   "quote",
	}

	return nil
}

func (s *mirrorStrategy) getBackingBalances() (*model.Number /*baseBackingBalance*/, *model.Number /*quoteBackingBalance*/, error) {
	balanceMap, e := s.exchange.GetAccountBalances([]interface{}{s.backingPair.Base, s.backingPair.Quote})
	if e != nil {
		return nil, nil, fmt.Errorf("unable to fetch balances for assets: %s", e)
	}

	// save asset balances from backing exchange to be used when placing offers in offset mode
	baseBalance, ok := balanceMap[s.backingPair.Base]
	if !ok {
		return nil, nil, fmt.Errorf("unable to fetch balance for base asset: %s", string(s.backingPair.Base))
	}

	quoteBalance, ok := balanceMap[s.backingPair.Quote]
	if !ok {
		return nil, nil, fmt.Errorf("unable to fetch balance for quote asset: %s", string(s.backingPair.Quote))
	}

	return &baseBalance, &quoteBalance, nil
}

// UpdateWithOps builds the operations we want performed on the account
func (s *mirrorStrategy) UpdateWithOps(
	buyingAOffers []hProtocol.Offer,
	sellingAOffers []hProtocol.Offer,
) ([]build.TransactionMutator, error) {
	// we want to fetch a few extra orders to account for potentially filtering out orders that don't meet the min base volume requirements
	ordersToFetch := int32(s.orderbookDepth + numOrdersBufferMinVolumeFilter)
	ob, e := s.exchange.GetOrderBook(s.backingPair, ordersToFetch)
	if e != nil {
		return nil, e
	}

	// limit bids and asks to max 50 operations each because of Stellar's limit of 100 ops/tx
	bids := ob.Bids()
	asks := ob.Asks()
	log.Printf("backing orderbook before transformations, including %d additional buffer orders:\n", numOrdersBufferMinVolumeFilter)
	printBidsAndAsks(bids, asks)

	// we modify the bids and ask to represent the new orders to place so we reduce unnecessary memory allocations
	if s.bidVolumeDivideBy == -1.0 {
		bids = []model.Order{}
	} else {
		transformOrders(bids, (1 - s.perLevelSpread), (1.0 / s.bidVolumeDivideBy), s.maybeMaxOrderBaseCap)
		// only place orders that we can fulfill on the backing exchange, to reduce surpluses needing offsetting
		bids = filterOrdersByVolume(bids, s.backingConstraints.MinBaseVolume.AsFloat())
		if len(bids) > s.orderbookDepth {
			bids = bids[:s.orderbookDepth]
		}
	}
	if s.askVolumeDivideBy == -1.0 {
		asks = []model.Order{}
	} else {
		transformOrders(asks, (1 + s.perLevelSpread), (1.0 / s.askVolumeDivideBy), s.maybeMaxOrderBaseCap)
		// only place orders that we can fulfill on the backing exchange, to reduce surpluses needing offsetting
		asks = filterOrdersByVolume(asks, s.backingConstraints.MinBaseVolume.AsFloat())
		if len(asks) > s.orderbookDepth {
			asks = asks[:s.orderbookDepth]
		}
	}
	log.Printf("new orders to be placed (after transforming and filtering orders from backing exchange):\n")
	printBidsAndAsks(bids, asks)

	deleteBuyOps, buyOps, e := s.updateLevels(
		buyingAOffers,
		bids,
		s.sdex.ModifyBuyOffer,
		s.sdex.CreateBuyOffer,
		true,
		s.buyOnPrimaryBalanceCoordinator, // we sell on the backing exchange to offset trades that are bought on the primary exchange
	)
	if e != nil {
		return nil, e
	}
	log.Printf("num. buyOps in this update: %d\n", len(buyOps))

	deleteSellOps, sellOps, e := s.updateLevels(
		sellingAOffers,
		asks,
		s.sdex.ModifySellOffer,
		s.sdex.CreateSellOffer,
		false,
		s.sellOnPrimaryBalanceCoordinator, // we buy on the backing exchange to offset trades that are sold on the primary exchange
	)
	if e != nil {
		return nil, e
	}
	log.Printf("num. sellOps in this update: %d\n", len(sellOps))

	placeSellOpsFirst := len(bids) > 0 && len(sellingAOffers) > 0 && bids[0].Price.AsFloat() >= utils.PriceAsFloat(sellingAOffers[0].Price)
	if debugLogOffersOrders {
		if placeSellOpsFirst {
			log.Printf("---> passed placeSellOpsFirst condition where top bid (bids[0]) > first open ask offer (sellingAOffers[0]), placing sellOps first\n")
		} else {
			log.Printf("---> failed placeSellOpsFirst condition where top bid (bids[0]) > first open ask offer (sellingAOffers[0]), placing bidOps first\n")
		}
		printDebugOffersAndOps(
			buyingAOffers,
			sellingAOffers,
			deleteBuyOps,
			deleteSellOps,
			buyOps,
			sellOps,
		)
	}

	ops := []txnbuild.Operation{}
	// add both deleteOps lists first because we want to delete offers first so we "free" up our liabilities capacity to place the new/modified offers
	ops = append(ops, deleteBuyOps...)
	ops = append(ops, deleteSellOps...)
	if placeSellOpsFirst {
		ops = append(ops, sellOps...)
		ops = append(ops, buyOps...)
	} else {
		ops = append(ops, buyOps...)
		ops = append(ops, sellOps...)
	}

	return api.ConvertOperation2TM(ops), nil
}

func transformOrders(orders []model.Order, priceMultiplier float64, volumeMultiplier float64, maxVolumeCap *float64) {
	for _, o := range orders {
		*o.Price = *o.Price.Scale(priceMultiplier)
		*o.Volume = *o.Volume.Scale(volumeMultiplier)
		if maxVolumeCap != nil && o.Volume.AsFloat() > *maxVolumeCap {
			*o.Volume = *model.NumberFromFloat(*maxVolumeCap, o.Volume.Precision())
		}
	}
}

func filterOrdersByVolume(orders []model.Order, minBaseVolume float64) []model.Order {
	ret := []model.Order{}
	for _, o := range orders {
		if o.Volume.AsFloat() >= minBaseVolume {
			ret = append(ret, o)
		}
	}
	return ret
}

func printBidsAndAsks(bids []model.Order, asks []model.Order) {
	log.Printf("    bids on backing exchange:\n")
	for _, o := range bids {
		log.Printf("        price=%s, amount=%s\n", o.Price.AsString(), o.Volume.AsString())
	}
	log.Printf("    asks on backing exchange:\n")
	for _, o := range asks {
		log.Printf("        price=%s, amount=%s\n", o.Price.AsString(), o.Volume.AsString())
	}
}

func printDebugOffersAndOps(
	buyingAOffers []hProtocol.Offer,
	sellingAOffers []hProtocol.Offer,
	deleteBuyOps []txnbuild.Operation,
	deleteSellOps []txnbuild.Operation,
	buyOps []txnbuild.Operation,
	sellOps []txnbuild.Operation,
) {
	log.Printf("---> buyingOffers:\n")
	for _, o := range buyingAOffers {
		price := float64(o.PriceR.N) / float64(o.PriceR.D)
		invertedPrice := 1 / price
		amountFloat, _ := strconv.ParseFloat(o.Amount, 64)
		log.Printf("--->     offerID=%d, price=%.7f, amount=%.7f\n", o.ID, invertedPrice, amountFloat*price)
	}
	log.Printf("---> sellingOffers:\n")
	for _, o := range sellingAOffers {
		log.Printf("--->     offerID=%d, price=%s, amount=%s\n", o.ID, o.Price, o.Amount)
	}
	log.Printf("---> deleteBuyOps:\n")
	for _, o := range deleteBuyOps {
		mso := o.(*txnbuild.ManageSellOffer)
		price, _ := strconv.ParseFloat(mso.Price, 64)
		invertedPrice := 1 / price
		amountFloat, _ := strconv.ParseFloat(mso.Amount, 64)
		log.Printf("--->     offerID=%d, price=%.7f, amount=%.7f\n", mso.OfferID, invertedPrice, amountFloat*price)
	}
	log.Printf("---> deleteSellOps:\n")
	for _, o := range deleteSellOps {
		mso := o.(*txnbuild.ManageSellOffer)
		log.Printf("--->     offerID=%d, price=%s, amount=%s\n", mso.OfferID, mso.Price, mso.Amount)
	}
	log.Printf("---> additional bid ops:\n")
	for _, o := range buyOps {
		mso := o.(*txnbuild.ManageSellOffer)
		price, _ := strconv.ParseFloat(mso.Price, 64)
		invertedPrice := 1 / price
		amountFloat, _ := strconv.ParseFloat(mso.Amount, 64)
		log.Printf("--->     offerID=%d, price=%.7f, amount=%.7f\n", mso.OfferID, invertedPrice, amountFloat*price)
	}
	log.Printf("---> additional ask ops:\n")
	for _, o := range sellOps {
		mso := o.(*txnbuild.ManageSellOffer)
		log.Printf("--->     offerID=%d, price=%s, amount=%s\n", mso.OfferID, mso.Price, mso.Amount)
	}
}

func (s *mirrorStrategy) updateLevels(
	oldOffers []hProtocol.Offer,
	newOrders []model.Order,
	modifyOffer func(offer hProtocol.Offer, price float64, amount float64, incrementalNativeAmountRaw float64) (*txnbuild.ManageSellOffer, error),
	createOffer func(baseAsset hProtocol.Asset, quoteAsset hProtocol.Asset, price float64, amount float64, incrementalNativeAmountRaw float64) (*txnbuild.ManageSellOffer, error),
	hackPriceInvertForBuyOrderChangeCheck bool, // needed because createBuy and modBuy inverts price so we need this for price comparison in doModifyOffer
	bc *balanceCoordinator,
) ([]txnbuild.Operation /*deleteOps*/, []txnbuild.Operation /*ops*/, error) {
	ops := []txnbuild.Operation{}
	deleteOps := []txnbuild.Operation{}
	if len(newOrders) >= len(oldOffers) {
		for i := 0; i < len(oldOffers); i++ {
			if s.offsetTrades {
				hasBackingBalance, newBaseVolume, _ := bc.checkBalance(newOrders[i].Volume, newOrders[i].Price)
				if !hasBackingBalance {
					continue
				}
				// TODO NS - don't modify existing variables
				newOrders[i].Volume = newBaseVolume
			}

			modifyOp, deleteOp, e := s.doModifyOffer(oldOffers[i], newOrders[i], modifyOffer, hackPriceInvertForBuyOrderChangeCheck)
			if e != nil {
				return nil, nil, e
			}

			if modifyOp != nil {
				ops = append(ops, modifyOp)
			}
			if deleteOp != nil {
				deleteOps = append(deleteOps, deleteOp)
			}
		}

		// create offers for remaining new bids
		for i := len(oldOffers); i < len(newOrders); i++ {
			price := newOrders[i].Price
			vol := newOrders[i].Volume
			if s.offsetTrades {
				hasBackingBalance, newBaseVol, _ := bc.checkBalance(vol, price)
				if !hasBackingBalance {
					continue
				}
				// TODO NS - don't reuse variables
				vol = newBaseVol
			}

			incrementalNativeAmountRaw := s.sdex.ComputeIncrementalNativeAmountRaw(true)
			if vol.AsFloat() < s.backingConstraints.MinBaseVolume.AsFloat() {
				log.Printf("skip level creation, baseVolume (%s) < minBaseVolume (%s) of backing exchange\n", vol.AsString(), s.backingConstraints.MinBaseVolume.AsString())
				continue
			}

			mo, e := createOffer(*s.baseAsset, *s.quoteAsset, price.AsFloat(), vol.AsFloat(), incrementalNativeAmountRaw)
			if e != nil {
				return nil, nil, e
			}
			if mo != nil {
				ops = append(ops, mo)
				// update the cached liabilities if we create a valid operation to create an offer
				if hackPriceInvertForBuyOrderChangeCheck {
					s.ieif.AddLiabilities(*s.quoteAsset, *s.baseAsset, vol.Multiply(*price).AsFloat(), vol.AsFloat(), incrementalNativeAmountRaw)
				} else {
					s.ieif.AddLiabilities(*s.baseAsset, *s.quoteAsset, vol.AsFloat(), vol.Multiply(*price).AsFloat(), incrementalNativeAmountRaw)
				}
			}
		}
	} else {
		for i := 0; i < len(newOrders); i++ {
			if s.offsetTrades {
				hasBackingBalance, newBaseVolume, _ := bc.checkBalance(newOrders[i].Volume, newOrders[i].Price)
				if !hasBackingBalance {
					continue
				}
				// TODO NS - don't modify existing variables
				newOrders[i].Volume = newBaseVolume
			}

			modifyOp, deleteOp, e := s.doModifyOffer(oldOffers[i], newOrders[i], modifyOffer, hackPriceInvertForBuyOrderChangeCheck)
			if e != nil {
				return nil, nil, e
			}
			if modifyOp != nil {
				ops = append(ops, modifyOp)
			}
			if deleteOp != nil {
				deleteOps = append(deleteOps, deleteOp)
			}
		}

		// delete remaining prior offers
		for i := len(newOrders); i < len(oldOffers); i++ {
			deleteOp := s.sdex.DeleteOffer(oldOffers[i])
			deleteOps = append(deleteOps, &deleteOp)
		}
	}

	log.Printf("returning %d deleteOps and %d ops\n", len(deleteOps), len(ops))
	return deleteOps, ops, nil
}

// doModifyOffer returns a new modifyOp, deleteOp, error
func (s *mirrorStrategy) doModifyOffer(
	oldOffer hProtocol.Offer,
	newOrder model.Order,
	modifyOffer func(offer hProtocol.Offer, price float64, amount float64, incrementalNativeAmountRaw float64) (*txnbuild.ManageSellOffer, error),
	hackPriceInvertForBuyOrderChangeCheck bool, // needed because createBuy and modBuy inverts price so we need this for price comparison in doModifyOffer
) (txnbuild.Operation, txnbuild.Operation, error) {
	price := newOrder.Price
	vol := newOrder.Volume
	oldPrice := model.MustNumberFromString(oldOffer.Price, s.primaryConstraints.PricePrecision)
	oldVol := model.MustNumberFromString(oldOffer.Amount, s.primaryConstraints.VolumePrecision)
	if hackPriceInvertForBuyOrderChangeCheck {
		// we want to multiply oldVol by the original oldPrice so we can get the correct oldVol, since ModifyBuyOffer multiplies price * vol
		oldVol = oldVol.Multiply(*oldPrice)
		oldPrice = model.InvertNumber(oldPrice)
	}
	epsilon := 0.0001
	incrementalNativeAmountRaw := s.sdex.ComputeIncrementalNativeAmountRaw(false)
	sameOrderParams := oldPrice.EqualsPrecisionNormalized(*price, epsilon) && oldVol.EqualsPrecisionNormalized(*vol, epsilon)
	if sameOrderParams {
		// update the cached liabilities if we keep the existing offer
		if hackPriceInvertForBuyOrderChangeCheck {
			s.ieif.AddLiabilities(oldOffer.Selling, oldOffer.Buying, oldVol.Multiply(*oldPrice).AsFloat(), oldVol.AsFloat(), incrementalNativeAmountRaw)
		} else {
			s.ieif.AddLiabilities(oldOffer.Selling, oldOffer.Buying, oldVol.AsFloat(), oldVol.Multiply(*oldPrice).AsFloat(), incrementalNativeAmountRaw)
		}
		return nil, nil, nil
	}

	// convert the precision from the backing exchange to the primary exchange
	offerPrice := model.NumberByCappingPrecision(price, s.primaryConstraints.PricePrecision)
	offerAmount := model.NumberByCappingPrecision(vol, s.primaryConstraints.VolumePrecision)
	if s.offsetTrades && offerAmount.AsFloat() < s.backingConstraints.MinBaseVolume.AsFloat() {
		log.Printf("deleting level, baseVolume (%f) on backing exchange dropped below minBaseVolume of backing exchange (%f)\n",
			offerAmount.AsFloat(), s.backingConstraints.MinBaseVolume.AsFloat())
		deleteOp := s.sdex.DeleteOffer(oldOffer)
		return nil, &deleteOp, nil
	}
	mo, e := modifyOffer(
		oldOffer,
		offerPrice.AsFloat(),
		offerAmount.AsFloat(),
		incrementalNativeAmountRaw,
	)
	if e != nil {
		return nil, nil, e
	}
	if mo != nil {
		// update the cached liabilities if we create a valid operation to modify the offer
		if hackPriceInvertForBuyOrderChangeCheck {
			s.ieif.AddLiabilities(oldOffer.Selling, oldOffer.Buying, offerAmount.Multiply(*offerPrice).AsFloat(), offerAmount.AsFloat(), incrementalNativeAmountRaw)
		} else {
			s.ieif.AddLiabilities(oldOffer.Selling, oldOffer.Buying, offerAmount.AsFloat(), offerAmount.Multiply(*offerPrice).AsFloat(), incrementalNativeAmountRaw)
		}
		return mo, nil, nil
	}

	// since mo is nil we want to delete this offer
	deleteOp := s.sdex.DeleteOffer(oldOffer)
	return nil, &deleteOp, nil
}

// PostUpdate changes the strategy's state after the update has taken place
func (s *mirrorStrategy) PostUpdate() error {
	return nil
}

// GetFillHandlers impl
func (s *mirrorStrategy) GetFillHandlers() ([]api.FillHandler, error) {
	if s.offsetTrades {
		return []api.FillHandler{s}, nil
	}
	return nil, nil
}

func (s *mirrorStrategy) baseVolumeToOffset(trade model.Trade, newOrderAction model.OrderAction) (newVolume *model.Number, ok bool) {
	uncommittedBase := s.baseSurplus[newOrderAction].total.Subtract(*s.baseSurplus[newOrderAction].committed)

	if uncommittedBase.AsFloat() < s.backingConstraints.MinBaseVolume.Scale(0.5).AsFloat() {
		log.Printf("offset-skip | tradeID=%s | tradeBaseAmt=%f | tradeQuoteAmt=%f | tradePriceQuote=%f | minBaseVolume=%f | newOrderAction=%s | baseSurplusTotal=%f | baseSurplusCommitted=%f\n",
			trade.TransactionID.String(),
			trade.Volume.AsFloat(),
			trade.Volume.Multiply(*trade.Price).AsFloat(),
			trade.Price.AsFloat(),
			s.backingConstraints.MinBaseVolume.AsFloat(),
			newOrderAction.String(),
			s.baseSurplus[newOrderAction].total.AsFloat(),
			s.baseSurplus[newOrderAction].committed.AsFloat())
		return nil, false
	}

	if uncommittedBase.AsFloat() > s.backingConstraints.MinBaseVolume.AsFloat() {
		newVolume = uncommittedBase
	} else {
		// we want to offset the MinBaseVolume and take a deficit in the baseSurplus on success
		newVolume = &s.backingConstraints.MinBaseVolume
	}
	return model.NumberByCappingPrecision(newVolume, s.backingConstraints.VolumePrecision), true
}

// HandleFill impl
func (s *mirrorStrategy) HandleFill(trade model.Trade) error {
	// we should only ever have one active fill handler to avoid inconsistent R/W on baseSurplus
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// first check if this trade has already been handled
	queryResult, e := s.strategyMirrorTradeTriggerExistsQuery.QueryRow(trade.TransactionID.String())
	if e != nil {
		return fmt.Errorf("unable to fetch trade trigger for transactionID '%s': %s", trade.TransactionID.String(), e)
	}
	rowExists, ok := queryResult.(bool)
	if !ok {
		return fmt.Errorf("unable to convert result of strategyMirrorTradeTriggerExistsQuery to bool: %v (type=%T)", queryResult, queryResult)
	}
	if rowExists {
		log.Printf("trade with txid '%s' was previously handled because we have a row in the strategy_mirror_trade_triggers table with this txid, not handling again and returning\n", trade.TransactionID.String())
		return nil
	}

	newOrderAction := trade.OrderAction.Reverse()
	// increase the baseSurplus for the additional amount that needs to be offset because of the incoming trade
	s.baseSurplus[newOrderAction].total = s.baseSurplus[newOrderAction].total.Add(*trade.Volume)

	newVolume, ok := s.baseVolumeToOffset(trade, newOrderAction)
	if !ok {
		return nil
	}
	// commit the newVolume that we are trying to use so the next handler does not double-count this amount
	s.baseSurplus[newOrderAction].committed = s.baseSurplus[newOrderAction].committed.Add(*newVolume)

	newOrder := model.Order{
		Pair:        s.backingPair, // we want to offset trades on the backing exchange so use the backing exchange's trading pair
		OrderAction: newOrderAction,
		OrderType:   model.OrderTypeLimit,
		Price:       model.NumberByCappingPrecision(trade.Price, s.backingConstraints.PricePrecision),
		Volume:      newVolume,
		Timestamp:   nil,
	}
	log.Printf("offset-attempt | tradeID=%s | tradeBaseAmt=%f | tradeQuoteAmt=%f | tradePriceQuote=%f | newOrderAction=%s | baseSurplusTotal=%f | baseSurplusCommitted=%f | minBaseVolume=%f | newOrderBaseAmt=%f | newOrderQuoteAmt=%f | newOrderPriceQuote=%f\n",
		trade.TransactionID.String(),
		trade.Volume.AsFloat(),
		trade.Volume.Multiply(*trade.Price).AsFloat(),
		trade.Price.AsFloat(),
		newOrderAction.String(),
		s.baseSurplus[newOrderAction].total.AsFloat(),
		s.baseSurplus[newOrderAction].committed.AsFloat(),
		s.backingConstraints.MinBaseVolume.AsFloat(),
		newOrder.Volume.AsFloat(),
		newOrder.Volume.Multiply(*newOrder.Price).AsFloat(),
		newOrder.Price.AsFloat())

	// when offsetting trades we always submit as a taker order so use api.SubmitModeBoth
	transactionID, e := s.exchange.AddOrder(&newOrder, api.SubmitModeBoth)
	if e != nil {
		return fmt.Errorf("error when offsetting trade (newOrder=%s): %s", newOrder, e)
	}
	if transactionID == nil {
		return fmt.Errorf("error when offsetting trade (newOrder=%s): transactionID was <nil>", newOrder)
	}
	// insert into the db immediately after placing order on backing exchange
	e = s.insertTradeTrigger(trade.TransactionID.String(), transactionID.String())
	if e != nil {
		return fmt.Errorf("error when inserting trade trigger with txID=%s (newOrder=%s) (PK dupes not allowed): %s", transactionID.String(), newOrder, e)
	}

	// update the baseSurplus on success
	s.baseSurplus[newOrderAction].total = s.baseSurplus[newOrderAction].total.Subtract(*newVolume)
	s.baseSurplus[newOrderAction].committed = s.baseSurplus[newOrderAction].committed.Subtract(*newVolume)

	log.Printf("offset-success | tradeID=%s | tradeBaseAmt=%f | tradeQuoteAmt=%f | tradePriceQuote=%f | newOrderAction=%s | baseSurplusTotal=%f | baseSurplusCommitted=%f | minBaseVolume=%f | newOrderBaseAmt=%f | newOrderQuoteAmt=%f | newOrderPriceQuote=%f | transactionID=%s\n",
		trade.TransactionID.String(),
		trade.Volume.AsFloat(),
		trade.Volume.Multiply(*trade.Price).AsFloat(),
		trade.Price.AsFloat(),
		newOrderAction.String(),
		s.baseSurplus[newOrderAction].total.AsFloat(),
		s.baseSurplus[newOrderAction].committed.AsFloat(),
		s.backingConstraints.MinBaseVolume.AsFloat(),
		newOrder.Volume.AsFloat(),
		newOrder.Volume.Multiply(*newOrder.Price).AsFloat(),
		newOrder.Price.AsFloat(),
		transactionID)

	// trigger fill tracking on backing exchange
	trades, e := s.backingFillTracker.FillTrackSingleIteration()
	if e != nil {
		return fmt.Errorf("unable to track a single iteration of fills from the backing exchange: %s", e)
	}
	log.Printf("found %d trades on load from backing exchange in HandleFill\n", len(trades))

	return nil
}

func (s *mirrorStrategy) insertTradeTrigger(primaryTxID string, backingTxID string) error {
	sqlInsert := fmt.Sprintf(kelpdb.SqlStrategyMirrorTradeTriggersInsertTemplate,
		s.marketID,
		primaryTxID,
		s.backingMarketID,
		backingTxID,
	)
	_, e := s.db.Exec(sqlInsert)
	if e != nil {
		if strings.Contains(e.Error(), "duplicate key value violates unique constraint \"strategy_mirror_trade_triggers_pkey\"") {
			log.Printf("trying to reinsert trade trigger (market_id=%s, txid=%s, backing_market_id=%s, backing_txid=%s) to db, ignore and continue\n", s.marketID, primaryTxID, s.backingMarketID, backingTxID)
			return nil
		}

		// return an error on any other errors
		return fmt.Errorf("could not execute sql insert values statement (%s): %s", sqlInsert, e)
	}

	log.Printf("wrote trade trigger (market_id=%s, txid=%s, backing_market_id=%s, backing_txid=%s) to db\n", s.marketID, primaryTxID, s.backingMarketID, backingTxID)
	return nil
}

// balanceCoordinator coordinates the balances from the backing exchange with orders placed on the primary exchange
// it serves an almost identical function to the ieif module but includes logic to coordinate balances with the backing exchange
type balanceCoordinator struct {
	primaryBalance     *model.Number
	placedPrimaryUnits *model.Number
	primaryAssetType   string
	isPrimaryBuy       bool
	placedBackingUnits *model.Number
	backingBalance     *model.Number
	backingAssetType   string
}

func (b *balanceCoordinator) getPlacedPrimaryUnits() *model.Number {
	return b.placedPrimaryUnits
}

func (b *balanceCoordinator) getPlacedBackingUnits() *model.Number {
	return b.placedBackingUnits
}

// checkBalance uses a larger precision for internal calculations because of this bug: https://github.com/stellar/kelp/issues/541
// it eventually rounds back to the precision of the passed in volume
func (b *balanceCoordinator) checkBalance(vol *model.Number, price *model.Number) (bool /*hasBackingBalance*/, *model.Number /*newBaseVolume*/, *model.Number /*newQuoteVolume*/) {
	expandedPrecisionVolume := model.NumberFromFloat(vol.AsFloat(), model.InternalCalculationsPrecision)
	expandedPrecisionPrice := model.NumberFromFloat(price.AsFloat(), model.InternalCalculationsPrecision)

	// we want to constrain units on primary exchange to ensure we can mirror correctly
	additionalPrimaryUnits := expandedPrecisionVolume
	if b.isPrimaryBuy { // buying base on primary, selling base on backing
		// convert to quote units since we are selling quote on primary
		additionalPrimaryUnits = vol.MultiplyRoundTruncate(*expandedPrecisionPrice)
	}
	remainingPrimaryUnits := b.primaryBalance.Subtract(*b.placedPrimaryUnits)
	if remainingPrimaryUnits.AsFloat() < 0.0000002 {
		log.Printf("balanceCoordinator: skip level creation, not enough balance of %s asset on primary exchange: %s (needs an additional %s units)\n", b.primaryAssetType, b.primaryBalance.AsString(), additionalPrimaryUnits.AsString())
		return false, nil, nil
	}
	if additionalPrimaryUnits.AsFloat() > remainingPrimaryUnits.AsFloat() {
		log.Printf("balanceCoordinator: constraining level to %s units of %s asset because we were limited by assets on primary exchange", remainingPrimaryUnits.AsString(), b.primaryAssetType)
		additionalPrimaryUnits = remainingPrimaryUnits
	}

	// now do for backing exchange to ensure we can offset trades correctly
	additionalBackingUnits := expandedPrecisionVolume
	if !b.isPrimaryBuy { // selling base on primary, buying base on backing
		// convert to quote units since we are selling quote on backing
		additionalBackingUnits = vol.MultiplyRoundTruncate(*expandedPrecisionPrice)
	}
	remainingBackingUnits := b.backingBalance.Subtract(*b.placedBackingUnits)
	if remainingBackingUnits.AsFloat() < 0.0000002 {
		log.Printf("balanceCoordinator: skip level creation, not enough balance of %s asset on backing exchange: %s (needs an additional %s units)\n", b.backingAssetType, b.backingBalance.AsString(), additionalBackingUnits.AsString())
		return false, nil, nil
	}
	if additionalBackingUnits.AsFloat() > remainingBackingUnits.AsFloat() {
		log.Printf("balanceCoordinator: constraining level to %s units of %s asset because we were limited by assets on backing exchange", remainingBackingUnits.AsString(), b.backingAssetType)
		additionalBackingUnits = remainingBackingUnits
	}

	// take the min value of new possible units to ensure there is overlap between primary and backing
	normalizedPrimaryUnits := additionalPrimaryUnits
	normalizedBackingUnits := additionalBackingUnits
	if b.isPrimaryBuy {
		normalizedPrimaryUnits = normalizedPrimaryUnits.DivideRoundTruncate(*expandedPrecisionPrice)
	} else {
		normalizedBackingUnits = normalizedBackingUnits.DivideRoundTruncate(*expandedPrecisionPrice)
	}
	var minBaseUnitsFloat float64
	if normalizedBackingUnits.AsFloat() < normalizedPrimaryUnits.AsFloat() {
		minBaseUnitsFloat = normalizedBackingUnits.AsFloat()
		log.Printf("balanceCoordinator: using normalizedBackingUnits (%.10f) since it is smaller than normalizedPrimaryUnits (%.10f)\n", normalizedBackingUnits.AsFloat(), normalizedPrimaryUnits.AsFloat())
	} else if normalizedPrimaryUnits.AsFloat() < normalizedBackingUnits.AsFloat() {
		minBaseUnitsFloat = normalizedPrimaryUnits.AsFloat()
		log.Printf("balanceCoordinator: using normalizedPrimaryUnits (%.10f) since it is smaller than normalizedBackingUnits (%.10f)\n", normalizedPrimaryUnits.AsFloat(), normalizedBackingUnits.AsFloat())
	} else {
		minBaseUnitsFloat = normalizedPrimaryUnits.AsFloat()
		log.Printf("balanceCoordinator: nothing to constrain since normalizedPrimaryUnits and normalizedBackingUnits were equal (%.10f)\n", normalizedPrimaryUnits.AsFloat())
	}
	minBaseUnits := model.NumberFromFloat(minBaseUnitsFloat, model.InternalCalculationsPrecision)

	// convert to final precision values now
	minBaseUnitsFinalPrecision := model.NumberFromFloatRoundTruncate(minBaseUnits.AsFloat(), vol.Precision())
	minQuoteUnits := minBaseUnitsFinalPrecision.MultiplyRoundTruncate(*expandedPrecisionPrice)
	minQuoteUnitsFinalPrecision := model.NumberFromFloatRoundTruncate(minQuoteUnits.AsFloat(), vol.Precision())
	var minPrimaryUnitsFinalPrecision, minBackingUnitsFinalPrecision *model.Number
	if b.isPrimaryBuy {
		minPrimaryUnitsFinalPrecision = minQuoteUnitsFinalPrecision
		minBackingUnitsFinalPrecision = minBaseUnitsFinalPrecision
	} else {
		minPrimaryUnitsFinalPrecision = minBaseUnitsFinalPrecision
		minBackingUnitsFinalPrecision = minQuoteUnitsFinalPrecision
	}

	b.placedPrimaryUnits = b.placedPrimaryUnits.Add(*minPrimaryUnitsFinalPrecision)
	b.placedBackingUnits = b.placedBackingUnits.Add(*minBackingUnitsFinalPrecision)
	return true, minBaseUnitsFinalPrecision, minQuoteUnitsFinalPrecision
}
