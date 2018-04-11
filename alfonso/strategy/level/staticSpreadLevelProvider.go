package level

// StaticLevel represents a layer in the orderbook defined statically
// extracted here because it's shared by strategy and sideStrategy where strategy depeneds on sideStrategy
type StaticLevel struct {
	SPREAD float64 `valid:"-"`
	AMOUNT float64 `valid:"-"`
}

// staticSpreadLevelProvider provides a fixed number of levels using a static percentage spread
type staticSpreadLevelProvider struct {
	staticLevels []StaticLevel
	amountOfBase float64
}

// ensure it implements the Provider interface
var _ Provider = &staticSpreadLevelProvider{}

// MakeStaticSpreadLevelProvider is a factory method
func MakeStaticSpreadLevelProvider(staticLevels []StaticLevel, amountOfBase float64) Provider {
	return &staticSpreadLevelProvider{
		staticLevels: staticLevels,
		amountOfBase: amountOfBase,
	}
}

// GetLevels impl.
func (p *staticSpreadLevelProvider) GetLevels(centerPrice float64) ([]Level, error) {
	levels := []Level{}
	for _, sl := range p.staticLevels {
		absoluteSpread := centerPrice * sl.SPREAD
		levels = append(levels, Level{
			// we always add here because it is only used in the context of selling so we always charge a higher price to include a spread
			targetPrice:  centerPrice + absoluteSpread,
			targetAmount: sl.AMOUNT * p.amountOfBase,
		})
	}
	return levels, nil
}
