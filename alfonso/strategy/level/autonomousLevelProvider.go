package level

// autonomousLevelProvider provides levels based on an exponential curve wrt. the number of assets held in the account.
// This strategy does not allow using the balance of a single asset for more strategies other than this one because
// that would require building in some trade tracking along with asset balance tracking for this strategy. The support
// for this can always be added later.
type autonomousLevelProvider struct {
	spread    float64
	maxLevels int8
}

// ensure it implements Provider
var _ Provider = &autonomousLevelProvider{}

// MakeAutonomousLevelProvider is the factory method
func MakeAutonomousLevelProvider(spread float64, maxLevels int8) Provider {
	return &autonomousLevelProvider{
		spread:    spread,
		maxLevels: maxLevels,
	}
}

// GetLevels impl.
func (p *autonomousLevelProvider) GetLevels(maxAssetBase float64, maxAssetQuote float64) ([]Level, error) {
	levels := []Level{}
	// TODO 2
	return levels, nil
}
