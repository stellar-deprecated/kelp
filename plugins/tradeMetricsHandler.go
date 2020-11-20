package plugins

import (
	"github.com/stellar/kelp/model"
)

// TradeMetricsHandler tracks the number of trades
type TradeMetricsHandler struct {
	trades []model.Trade
}

// TradeMetrics stores the computed trade-related metrics.
// TODO DS Pre-pend interval__ to all JSON fields.
type TradeMetrics struct {
	totalBaseVolume   float64 `json:"total_base_volume"`
	totalQuoteVolume  float64 `json:"total_quote_volume"`
	netBaseVolume     float64 `json:"net_base_volume"`
	netQuoteVolume    float64 `json:"net_quote_volume"`
	numTrades         float64 `json:"num_trades"`
	avgTradeSizeBase  float64 `json:"avg_trade_size_base"`
	avgTradeSizeQuote float64 `json:"avg_trade_size_quote"`
	avgTradePrice     float64 `json:"avg_trade_price"`
	vwap              float64 `json:"vwap"`
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

// ComputeTradeMetrics computes trade-related metrics.
func (h *TradeMetricsHandler) ComputeTradeMetrics(secondsSinceLastUpdate float64) TradeMetrics {
	trades := h.GetTrades()
	h.Reset()

	// Note that we don't consider fees right now, as we don't have the infrastructure to understand fee units when tracking trades.
	totalBaseVolume := 0.0
	totalQuoteVolume := 0.0
	totalPrice := 0.0
	netBaseVolume := 0.0
	netQuoteVolume := 0.0

	numTrades := float64(len(trades))
	for _, t := range trades {
		base := t.Volume.AsFloat()
		price := t.Price.AsFloat()
		quote := base * price

		totalBaseVolume += base
		totalPrice += price
		totalQuoteVolume += quote

		if t.OrderAction.IsBuy() {
			netBaseVolume += base
			netQuoteVolume -= quote
		} else {
			netBaseVolume -= base
			netQuoteVolume -= quote
		}
	}

	avgTradeSizeBase := totalBaseVolume / numTrades
	avgTradeSizeQuote := totalQuoteVolume / numTrades
	avgTradePrice := totalPrice / numTrades
	avgTradeThroughputBase := totalBaseVolume / secondsSinceLastUpdate
	avgTradeThroughputQuote := totalQuoteVolume / secondsSinceLastUpdate

	tradeMetrics := TradeMetrics{
		totalBaseVolume:   totalBaseVolume,
		totalQuoteVolume:  totalQuoteVolume,
		netBaseVolume:     netBaseVolume,
		netQuoteVolume:    netQuoteVolume,
		numTrades:         numTrades,
		avgTradeSizeBase:  avgTradeSizeBase,
		avgTradeSizeQuote: avgTradeSizeQuote,
		avgTradePrice:     avgTradePrice,
		vwap:              totalQuoteVolume / totalBaseVolume,
	}

	return tradeMetrics
}
