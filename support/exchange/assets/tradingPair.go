package assets

// TradingPair lists an ordered pair that is understood by the bot and our exchange API
type TradingPair struct {
	AssetA Asset
	AssetB Asset
}

// ToString converts the trading pair to a string using the passed in assetConverter
func (p TradingPair) ToString(c *AssetConverter, delim string) (string, error) {
	a, e := c.ToString(p.AssetA)
	if e != nil {
		return "", e
	}

	b, e := c.ToString(p.AssetB)
	if e != nil {
		return "", e
	}

	return a + delim + b, nil
}

// TradingPairs2Strings converts the trading pairs to an array of strings
func TradingPairs2Strings(c *AssetConverter, delim string, pairs []TradingPair) (map[TradingPair]string, error) {
	m := map[TradingPair]string{}
	for _, p := range pairs {
		pairString, e := p.ToString(c, delim)
		if e != nil {
			return nil, e
		}
		m[p] = pairString
	}
	return m, nil
}
