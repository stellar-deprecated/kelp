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

// Reset sets the handler's trade counter to zero.
func (h *TradeMetricsHandler) Reset() {
	h.trades = []model.Trade{}
}

// Read stores new trades internally.
func (h *TradeMetricsHandler) Read(newTrades []model.Trade) error {
	for _, nt := range newTrades {
		h.trades = append(h.trades, nt)
	}
}

// Get returns the number of trades.
func (h *TradeMetricsHandler) Get() int {
	return len(h.trades)
}

// HandleTrade impl
func (h *TradeMetricsHandler) HandleTrade(trade model.Trade) error {
	// TODO: Add more if needed.
	h.trades = append(h.trades, trade)
	return nil
}
