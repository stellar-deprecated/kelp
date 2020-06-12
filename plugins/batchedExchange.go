package plugins

import (
	"fmt"
	"log"
	"math"
	"math/rand"
	"reflect"
	"strconv"
	"time"

	"github.com/stellar/go/build"
	hProtocol "github.com/stellar/go/protocols/horizon"
	"github.com/stellar/go/txnbuild"
	"github.com/stellar/kelp/api"
	"github.com/stellar/kelp/model"
	"github.com/stellar/kelp/support/utils"
)

// largePrecision is a large precision value for in-memory calculations
const largePrecision = 10

// BatchedExchange accumulates instructions that can be read out and processed in a batch-style later
type BatchedExchange struct {
	commands        []Command
	inner           api.Exchange
	simMode         bool
	baseAsset       hProtocol.Asset
	quoteAsset      hProtocol.Asset
	tradingAccount  string
	orderID2OfferID map[string]int64
	offerID2OrderID map[int64]string
}

var _ api.ExchangeShim = BatchedExchange{}

// MakeBatchedExchange factory
func MakeBatchedExchange(
	inner api.Exchange,
	simMode bool,
	baseAsset hProtocol.Asset,
	quoteAsset hProtocol.Asset,
	tradingAccount string,
) *BatchedExchange {
	return &BatchedExchange{
		commands:        []Command{},
		inner:           inner,
		simMode:         simMode,
		baseAsset:       baseAsset,
		quoteAsset:      quoteAsset,
		tradingAccount:  tradingAccount,
		orderID2OfferID: map[string]int64{},
		offerID2OrderID: map[int64]string{},
	}
}

// Operation represents a type of operation
type Operation int8

// type of Operations
const (
	OpAdd Operation = iota
	OpCancel
)

// Command struct allows us to follow the Command pattern
type Command struct {
	op     Operation
	add    *model.Order
	cancel *model.OpenOrder
}

// GetOp returns the Operation
func (c *Command) GetOp() Operation {
	return c.op
}

// GetAdd returns the add op
func (c *Command) GetAdd() (*model.Order, error) {
	if c.add == nil {
		return nil, fmt.Errorf("add op does not exist")
	}
	return c.add, nil
}

// GetCancel returns the cancel op
func (c *Command) GetCancel() (*model.OpenOrder, error) {
	if c.cancel == nil {
		return nil, fmt.Errorf("cancel op does not exist")
	}
	return c.cancel, nil
}

// MakeCommandAdd impl
func MakeCommandAdd(order *model.Order) Command {
	return Command{
		op:  OpAdd,
		add: order,
	}
}

// MakeCommandCancel impl
func MakeCommandCancel(openOrder *model.OpenOrder) Command {
	return Command{
		op:     OpCancel,
		cancel: openOrder,
	}
}

// GetBalanceHack impl
func (b BatchedExchange) GetBalanceHack(asset hProtocol.Asset) (*api.Balance, error) {
	modelAsset := model.FromHorizonAsset(asset)
	balances, e := b.GetAccountBalances([]interface{}{modelAsset})
	if e != nil {
		return nil, fmt.Errorf("error fetching balances in GetBalanceHack: %s", e)
	}

	if v, ok := balances[modelAsset]; ok {
		return &api.Balance{
			Balance: v.AsFloat(),
			Trust:   math.MaxFloat64,
			Reserve: 0.0,
		}, nil
	}
	return nil, fmt.Errorf("asset was missing in GetBalanceHack result: %s", utils.Asset2String(asset))
}

// LoadOffersHack impl
func (b BatchedExchange) LoadOffersHack() ([]hProtocol.Offer, error) {
	pair := &model.TradingPair{
		Base:  model.FromHorizonAsset(b.baseAsset),
		Quote: model.FromHorizonAsset(b.quoteAsset),
	}
	openOrders, e := b.GetOpenOrders([]*model.TradingPair{pair})
	if e != nil {
		return nil, fmt.Errorf("error fetching open orders in LoadOffersHack: %s", e)
	}

	offers := []hProtocol.Offer{}
	for i, v := range openOrders {
		var offers1 []hProtocol.Offer
		offers1, e = b.OpenOrders2Offers(v, b.baseAsset, b.quoteAsset, b.tradingAccount)
		if e != nil {
			return nil, fmt.Errorf("error converting open orders to offers in iteration %v in LoadOffersHack: %s", i, e)
		}
		offers = append(offers, offers1...)
	}
	return offers, nil
}

// GetOrderConstraints impl
func (b BatchedExchange) GetOrderConstraints(pair *model.TradingPair) *model.OrderConstraints {
	return b.inner.GetOrderConstraints(pair)
}

// OverrideOrderConstraints impl, can partially override values for specific pairs
func (b BatchedExchange) OverrideOrderConstraints(pair *model.TradingPair, override *model.OrderConstraintsOverride) {
	b.inner.OverrideOrderConstraints(pair, override)
}

// GetOrderBook impl
func (b BatchedExchange) GetOrderBook(pair *model.TradingPair, maxCount int32) (*model.OrderBook, error) {
	return b.inner.GetOrderBook(pair, maxCount)
}

// GetTradeHistory impl
func (b BatchedExchange) GetTradeHistory(pair model.TradingPair, maybeCursorStart interface{}, maybeCursorEnd interface{}) (*api.TradeHistoryResult, error) {
	return b.inner.GetTradeHistory(pair, maybeCursorStart, maybeCursorEnd)
}

// GetLatestTradeCursor impl
func (b BatchedExchange) GetLatestTradeCursor() (interface{}, error) {
	return b.inner.GetLatestTradeCursor()
}

// SubmitOpsSynch is the forced synchronous version of SubmitOps below (same for batchedExchange)
func (b BatchedExchange) SubmitOpsSynch(ops []build.TransactionMutator, submitMode api.SubmitMode, asyncCallback func(hash string, e error)) error {
	return b.SubmitOps(ops, submitMode, asyncCallback)
}

// SubmitOps performs any finalization or submission step needed by the exchange
func (b BatchedExchange) SubmitOps(opsOld []build.TransactionMutator, submitMode api.SubmitMode, asyncCallback func(hash string, e error)) error {
	ops := api.ConvertTM2Operation(opsOld)

	var e error
	b.commands, e = b.Ops2Commands(ops, b.baseAsset, b.quoteAsset)
	if e != nil {
		if asyncCallback != nil {
			go asyncCallback("", e)
		}
		return fmt.Errorf("could not convert ops2commands: %s | allOps = %v", e, ops)
	}

	if b.simMode {
		log.Printf("running in simulation mode so not submitting to the inner exchange\n")
		if asyncCallback != nil {
			go asyncCallback("", nil)
		}
		return nil
	}

	results := []submitResult{}
	numProcessed := 0
	for _, c := range b.commands {
		r := c.exec(b.inner, submitMode)
		if r == nil {
			// remove all processed commands
			// b.commands = b.commands[numProcessed:]
			b.logResults(results)
			e := fmt.Errorf("unrecognized operation '%v', stopped submitting", c.op)
			if asyncCallback != nil {
				go asyncCallback("", e)
			}
			return e
		}
		results = append(results, *r)
		numProcessed++
	}

	// remove all processed commands
	// b.commands = b.commands[numProcessed:]

	b.logResults(results)
	if asyncCallback != nil {
		go asyncCallback("", nil)
	}
	return nil
}

func (b BatchedExchange) logResults(results []submitResult) {
	log.Printf("Results from submitting:\n")
	for _, r := range results {
		opString := "add"
		var v interface{}
		v = r.add
		if r.op == OpCancel {
			opString = "cancel"
			v = r.cancel
		}

		errorSuffix := ""
		if r.e != nil {
			errorSuffix = fmt.Sprintf(", error=%s", r.e)
		}
		log.Printf("    submitResult[op=%s, value=%v%s]\n", opString, v, errorSuffix)
	}
}

func (c Command) exec(x api.Exchange, submitMode api.SubmitMode) *submitResult {
	switch c.op {
	case OpAdd:
		v, e := x.AddOrder(c.add, submitMode)
		return &submitResult{
			op:  c.op,
			e:   e,
			add: v,
		}
	case OpCancel:
		v, e := x.CancelOrder(model.MakeTransactionID(c.cancel.ID), *c.cancel.Pair)
		return &submitResult{
			op:     c.op,
			e:      e,
			cancel: &v,
		}
	default:
		return nil
	}
}

// GetAccountBalances impl.
func (b BatchedExchange) GetAccountBalances(assetList []interface{}) (map[interface{}]model.Number, error) {
	return b.inner.GetAccountBalances(assetList)
}

// GetOpenOrders impl.
func (b BatchedExchange) GetOpenOrders(pairs []*model.TradingPair) (map[model.TradingPair][]model.OpenOrder, error) {
	return b.inner.GetOpenOrders(pairs)
}

type submitResult struct {
	op     Operation
	e      error
	add    *model.TransactionID
	cancel *model.CancelOrderResult
}

func (b BatchedExchange) genUniqueID() int64 {
	var ID int64
	for {
		ID = rand.Int63()
		log.Printf("generated unique ID = %d\n", ID)
		// should have generated a unique value
		if _, ok := b.offerID2OrderID[ID]; !ok {
			break
		}
		log.Printf("generated ID (%d) was not unique! retrying...\n", ID)
	}
	return ID
}

// OpenOrders2Offers converts...
func (b BatchedExchange) OpenOrders2Offers(orders []model.OpenOrder, baseAsset hProtocol.Asset, quoteAsset hProtocol.Asset, tradingAccount string) ([]hProtocol.Offer, error) {
	offers := []hProtocol.Offer{}
	for _, order := range orders {
		sellingAsset := baseAsset
		buyingAsset := quoteAsset
		amount := order.Volume.AsString()
		price, e := convert2Price(order.Price)
		if e != nil {
			return nil, fmt.Errorf("unable to convert order price to a ratio: %s", e)
		}
		priceString := order.Price.AsString()
		if order.OrderAction == model.OrderActionBuy {
			sellingAsset = quoteAsset
			buyingAsset = baseAsset
			// TODO need to test price and volume conversions correctly
			amount = fmt.Sprintf("%.8f", order.Volume.AsFloat()*order.Price.AsFloat())
			invertedPrice := model.InvertNumber(order.Price)
			// invert price ratio here instead of using convert2Price again since it has an overflow for XLM/BTC
			price = hProtocol.Price{
				N: price.D,
				D: price.N,
			}
			priceString = invertedPrice.AsString()
		}

		// generate an offerID for the non-numerical orderID (hoops we have to jump through because of the hacked approach to using centralized exchanges)
		var ID int64
		if v, ok := b.orderID2OfferID[order.ID]; ok {
			ID = v
		} else {
			ID = b.genUniqueID()
			b.orderID2OfferID[order.ID] = ID
			b.offerID2OrderID[ID] = order.ID
		}

		var lmt *time.Time
		if order.Timestamp != nil {
			lastModTime := time.Unix(int64(*order.Timestamp)/1000, 0)
			lmt = &lastModTime
		}
		offers = append(offers, hProtocol.Offer{
			ID:                 ID,
			Seller:             tradingAccount,
			Selling:            sellingAsset,
			Buying:             buyingAsset,
			Amount:             amount,
			PriceR:             price,
			Price:              priceString,
			LastModifiedLedger: 0, // TODO fix?
			LastModifiedTime:   lmt,
		})
	}
	return offers, nil
}

func convert2Price(number *model.Number) (hProtocol.Price, error) {
	n, d, e := number.AsRatio()
	if e != nil {
		return hProtocol.Price{}, fmt.Errorf("unable to convert2Price: %s", e)
	}
	return hProtocol.Price{
		N: n,
		D: d,
	}, nil
}

// manageOffer2Order converts a manage offer operation to a model.Order
func manageOffer2Order(mob *txnbuild.ManageSellOffer, baseAsset hProtocol.Asset, quoteAsset hProtocol.Asset, orderConstraints *model.OrderConstraints) (*model.Order, error) {
	orderAction := model.OrderActionSell

	priceFloat, e := strconv.ParseFloat(mob.Price, 64)
	if e != nil {
		return nil, fmt.Errorf("could not convert price (%s) to float: %s", mob.Price, e)
	}

	amountFloat, e := strconv.ParseFloat(mob.Amount, 64)
	if e != nil {
		return nil, fmt.Errorf("could not convert amount (%s) to float: %s", mob.Amount, e)
	}

	price := model.NumberFromFloat(priceFloat, largePrecision)
	volume := model.NumberFromFloat(amountFloat, largePrecision)
	isBuy, e := utils.AssetOnlyCodeEquals(quoteAsset, mob.Selling)
	if e != nil {
		return nil, fmt.Errorf("could not compare assets, error: %s", e)
	}
	if isBuy {
		orderAction = model.OrderActionBuy
		// TODO need to test price and volume conversions correctly
		// volume calculation needs to happen first since it uses the non-inverted price when multiplying
		volume = model.NumberFromFloat(volume.AsFloat()*price.AsFloat(), orderConstraints.VolumePrecision)
		price = model.InvertNumber(price)
	}
	volume = model.NumberByCappingPrecision(volume, orderConstraints.VolumePrecision)
	price = model.NumberByCappingPrecision(price, orderConstraints.PricePrecision)

	return &model.Order{
		Pair: &model.TradingPair{
			Base:  model.FromHorizonAsset(baseAsset),
			Quote: model.FromHorizonAsset(quoteAsset),
		},
		OrderAction: orderAction,
		OrderType:   model.OrderTypeLimit,
		Price:       price,
		Volume:      volume,
		Timestamp:   model.MakeTimestamp(time.Now().UnixNano() / int64(time.Millisecond)),
	}, nil
}

func order2OpenOrder(order *model.Order, txID *model.TransactionID) *model.OpenOrder {
	return &model.OpenOrder{
		Order: *order,
		ID:    txID.String(),
		// we don't know these values so use nil
		StartTime:      nil,
		ExpireTime:     nil,
		VolumeExecuted: nil,
	}
}

// Ops2Commands converts...
func (b BatchedExchange) Ops2Commands(ops []txnbuild.Operation, baseAsset hProtocol.Asset, quoteAsset hProtocol.Asset) ([]Command, error) {
	pair := &model.TradingPair{
		Base:  model.FromHorizonAsset(baseAsset),
		Quote: model.FromHorizonAsset(quoteAsset),
	}
	return Ops2CommandsHack(ops, baseAsset, quoteAsset, b.offerID2OrderID, b.inner.GetOrderConstraints(pair))
}

// Ops2CommandsHack converts...
func Ops2CommandsHack(
	ops []txnbuild.Operation,
	baseAsset hProtocol.Asset,
	quoteAsset hProtocol.Asset,
	offerID2OrderID map[int64]string, // if map is nil then we ignore ID errors
	orderConstraints *model.OrderConstraints,
) ([]Command, error) {
	commands := []Command{}
	for _, op := range ops {
		switch manageOffer := op.(type) {
		case *txnbuild.ManageSellOffer:
			c, e := op2CommandsHack(manageOffer, baseAsset, quoteAsset, offerID2OrderID, orderConstraints)
			if e != nil {
				return nil, fmt.Errorf("unable to convert *txnbuild.ManageSellOffer to a Command: %s", e)
			}
			commands = append(commands, c...)
		default:
			return nil, fmt.Errorf("unable to recognize transaction mutator op (%s): %v", reflect.TypeOf(op), manageOffer)
		}
	}
	return commands, nil
}

// op2CommandsHack converts one op to possibly many Commands
func op2CommandsHack(
	manageOffer *txnbuild.ManageSellOffer,
	baseAsset hProtocol.Asset,
	quoteAsset hProtocol.Asset,
	offerID2OrderID map[int64]string, // if map is nil then we ignore ID errors
	orderConstraints *model.OrderConstraints,
) ([]Command, error) {
	commands := []Command{}
	order, e := manageOffer2Order(manageOffer, baseAsset, quoteAsset, orderConstraints)
	if e != nil {
		return nil, fmt.Errorf("error converting from manageOffer op to Order: %s", e)
	}

	amtFloat, e := strconv.ParseFloat(manageOffer.Amount, 64)
	if e != nil {
		return nil, fmt.Errorf("error parsing offer amount (%s): %s", manageOffer.Amount, e)
	}

	if amtFloat == 0 {
		// cancel
		// fetch real orderID here (hoops we have to jump through because of the hacked approach to using centralized exchanges)
		var orderID string
		if offerID2OrderID != nil {
			ID := manageOffer.OfferID
			var ok bool
			orderID, ok = offerID2OrderID[ID]
			if !ok {
				return nil, fmt.Errorf("there was an order that we have never seen before and did not have in the offerID2OrderID map, offerID (int): %d", ID)
			}
		} else {
			orderID = ""
		}
		txID := model.MakeTransactionID(orderID)
		openOrder := order2OpenOrder(order, txID)
		commands = append(commands, MakeCommandCancel(openOrder))
	} else if manageOffer.OfferID != 0 {
		// modify is cancel followed by create
		// -- cancel
		// fetch real orderID here (hoops we have to jump through because of the hacked approach to using centralized exchanges)
		var orderID string
		if offerID2OrderID != nil {
			ID := manageOffer.OfferID
			var ok bool
			orderID, ok = offerID2OrderID[ID]
			if !ok {
				return nil, fmt.Errorf("there was an order that we have never seen before and did not have in the offerID2OrderID map, offerID (int): %d", ID)
			}
		} else {
			orderID = ""
		}
		txID := model.MakeTransactionID(orderID)
		openOrder := order2OpenOrder(order, txID)
		commands = append(commands, MakeCommandCancel(openOrder))
		// -- create
		commands = append(commands, MakeCommandAdd(order))
	} else {
		// create
		commands = append(commands, MakeCommandAdd(order))
	}
	return commands, nil
}
