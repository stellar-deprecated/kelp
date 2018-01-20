package assets

// KrakenAssetConverter is the asset converter for the Kraken exchange
var KrakenAssetConverter = makeAssetConverter(map[Asset]string{
	XLM: "XXLM",
	BTC: "XXBT",
	USD: "ZUSD",
})
