package level

// Level represents a layer in the orderbook
type Level struct {
	targetPrice  float64
	targetAmount float64
}

// TargetPrice returns the target price for this Level
func (l Level) TargetPrice() float64 {
	return l.targetPrice
}

// TargetAmount returns the target amount for this Level
func (l Level) TargetAmount() float64 {
	return l.targetAmount
}

// Provider returns the levels for the given center price.
// Implementations should consider the spread they want to keep along with possibly having variable number of levels if needed
type Provider interface {
	GetLevels(centerPrice float64) ([]Level, error)
}
