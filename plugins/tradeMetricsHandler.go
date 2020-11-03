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
func (h *TradeMetricsHandler) Read(newTrades []model.Trade) {
	for _, nt := range newTrades {
		h.trades = append(h.trades, nt)
	}
}

// NumTrades returns the number of trades.
func (h *TradeMetricsHandler) NumTrades() int {
	return len(h.trades)
}

// TotalBaseVolume returns the total base volume.
func (h *TradeMetricsHandler) TotalBaseVolume() (total float64) {
	for _, t := range h.trades {
		total += t.Volume.AsFloat()
	}
	return
}
