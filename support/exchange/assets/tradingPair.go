package assets

// TradingPair lists an ordered pair that is understood by the bot and our exchange API
type TradingPair struct {
	AssetA Asset
	AssetB Asset
}
