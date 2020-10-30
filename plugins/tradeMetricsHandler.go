package metrics

// TODO DS Rework handler to accumulate trades, not just count.
type tradeMetricsHandler struct {
	numTrades int
}

func (h *tradeMetricsHandler) Reset() {
	h.numTrades = 0
}

func (h *tradeMetricsHandler) Add(numNewTrades int) {
	h.numTrades += numNewTrades
}

func (h *tradeMetricsHandler) Get() int {
	return h.numTrades
}
