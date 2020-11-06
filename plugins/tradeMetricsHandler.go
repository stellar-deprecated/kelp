package plugins

import (
	"github.com/stellar/kelp/model"
)

// TradeMetricsHandler tracks the number of trades
type TradeMetricsHandler struct {
	trades []model.Trade
}

// MakeTradeMetricsHandler is a factory method for the TradeMetricsHandler
func MakeTradeMetricsHandler() *TradeMetricsHandler {
	return &TradeMetricsHandler{
		trades: []model.Trade{},
	}
}

// Reset sets the handler's trades to empty.
func (h *TradeMetricsHandler) Reset() {
	h.trades = []model.Trade{}
}

// GetTrades returns all stored trades.
func (h *TradeMetricsHandler) GetTrades() []model.Trade {
	return h.trades
}

// HandleFill handles a new trade
// Implements FillHandler interface
func (h *TradeMetricsHandler) HandleFill(trade model.Trade) error {
	h.trades = append(h.trades, trade)
	return nil
}
