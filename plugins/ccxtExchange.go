package plugins

import (
	"fmt"

	"github.com/lightyeario/kelp/api"
	"github.com/lightyeario/kelp/model"
	"github.com/lightyeario/kelp/support/sdk"
	"github.com/lightyeario/kelp/support/utils"
)

// ensure that ccxtExchange conforms to the Exchange interface
var _ api.Exchange = ccxtExchange{}

// ccxtExchange is the implementation for the CCXT REST library that supports many exchanges (https://github.com/franz-see/ccxt-rest, https://github.com/ccxt/ccxt/)
type ccxtExchange struct {
	assetConverter *model.AssetConverter
	delimiter      string
	api            *sdk.Ccxt
	precision      int8
}

// makeCcxtExchange is a factory method to make an exchange using the CCXT interface
func makeCcxtExchange(ccxtBaseURL string, exchangeName string) (api.Exchange, error) {
	c, e := sdk.MakeInitializedCcxtExchange(ccxtBaseURL, exchangeName)
	if e != nil {
		return nil, fmt.Errorf("error making a ccxt exchange: %s", e)
	}

	return ccxtExchange{
		assetConverter: model.CcxtAssetConverter,
		delimiter:      "/",
		api:            c,
		precision:      utils.SdexPrecision,
	}, nil
}

// GetTickerPrice impl.
func (c ccxtExchange) GetTickerPrice(pairs []model.TradingPair) (map[model.TradingPair]api.Ticker, error) {
	pairsMap, e := model.TradingPairs2Strings(c.assetConverter, c.delimiter, pairs)
	if e != nil {
		return nil, e
	}

	priceResult := map[model.TradingPair]api.Ticker{}
	for _, p := range pairs {
		tickerMap, e := c.api.FetchTicker(pairsMap[p])
		if e != nil {
			return nil, fmt.Errorf("error while fetching ticker price for trading pair %s: %s", pairsMap[p], e)
		}

		priceResult[p] = api.Ticker{
			AskPrice: model.NumberFromFloat(tickerMap["ask"].(float64), c.precision),
			BidPrice: model.NumberFromFloat(tickerMap["bid"].(float64), c.precision),
		}
	}

	return priceResult, nil
}

// GetAssetConverter impl
func (c ccxtExchange) GetAssetConverter() *model.AssetConverter {
	return c.assetConverter
}

// GetAccountBalances impl
func (c ccxtExchange) GetAccountBalances(assetList []model.Asset) (map[model.Asset]model.Number, error) {
	// TODO implement
	return nil, nil
}

// GetPrecision impl
func (c ccxtExchange) GetPrecision() int8 {
	// TODO implement
	return utils.SdexPrecision
}

// GetOrderBook impl
func (c ccxtExchange) GetOrderBook(pair *model.TradingPair, maxCount int32) (*model.OrderBook, error) {
	// TODO implement
	return nil, nil
}

// GetTrades impl
func (c ccxtExchange) GetTrades(pair *model.TradingPair, maybeCursor interface{}) (*api.TradesResult, error) {
	// TODO implement
	return nil, nil
}

// GetTradeHistory impl
func (c ccxtExchange) GetTradeHistory(maybeCursorStart interface{}, maybeCursorEnd interface{}) (*api.TradeHistoryResult, error) {
	// TODO implement
	return nil, nil
}

// GetOpenOrders impl
func (c ccxtExchange) GetOpenOrders() (map[model.TradingPair][]model.OpenOrder, error) {
	// TODO implement
	return nil, nil
}

// AddOrder impl
func (c ccxtExchange) AddOrder(order *model.Order) (*model.TransactionID, error) {
	// TODO implement
	return nil, nil
}

// CancelOrder impl
func (c ccxtExchange) CancelOrder(txID *model.TransactionID) (model.CancelOrderResult, error) {
	// TODO implement
	return model.CancelResultCancelSuccessful, nil
}

// PrepareDeposit impl
func (c ccxtExchange) PrepareDeposit(asset model.Asset, amount *model.Number) (*api.PrepareDepositResult, error) {
	// TODO implement
	return nil, nil
}

// GetWithdrawInfo impl
func (c ccxtExchange) GetWithdrawInfo(asset model.Asset, amountToWithdraw *model.Number, address string) (*api.WithdrawInfo, error) {
	// TODO implement
	return nil, nil
}

// WithdrawFunds impl
func (c ccxtExchange) WithdrawFunds(
	asset model.Asset,
	amountToWithdraw *model.Number,
	address string,
) (*api.WithdrawFunds, error) {
	// TODO implement
	return nil, nil
}
