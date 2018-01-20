package assets

// Asset is typed and enlists the allowed assets that are understood by the bot
type Asset string

// this is the list of assets understood by the bot.
// This string can be converted by the specific exchange adapter as is needed by the exchange's API
const (
	XLM Asset = "XLM"
	BTC Asset = "BTC"
	USD Asset = "USD"
)
