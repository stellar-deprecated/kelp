package model

import (
	"fmt"
)

// TradingPair lists an ordered pair that is understood by the bot and our exchange API.
// EUR/USD = 1.25; EUR is base, USD is Quote. EUR is more valuable in this example
// USD/EUR = 0.80; USD is base, EUR is Quote. EUR is more valuable in this example
type TradingPair struct {
	// Base represents the asset that has a unit of 1 (implicit)
	Base Asset
	// Quote (or Counter) represents the asset that has its unit specified relative to the base asset
	Quote Asset
}

// MakeTradingPair is a factory method
func MakeTradingPair(base Asset, quote Asset) *TradingPair {
	return &TradingPair{
		Base:  base,
		Quote: quote,
	}
}

// String is the stringer function
func (p TradingPair) String() string {
	s, e := p.ToString(Display, "/")
	if e != nil {
		return fmt.Sprintf("<error, TradingPair: %s>", e)
	}
	return s
}

// ToString converts the trading pair to a string using the passed in assetConverterInterface
func (p TradingPair) ToString(c AssetConverterInterface, delim string) (string, error) {
	a, e := c.ToString(p.Base)
	if e != nil {
		return "", e
	}

	b, e := c.ToString(p.Quote)
	if e != nil {
		return "", e
	}

	return a + delim + b, nil
}

// TradingPairFromString makes a TradingPair out of a string
func TradingPairFromString(codeSize int8, c AssetConverterInterface, p string) (*TradingPair, error) {
	return TradingPairFromString2(codeSize, []AssetConverterInterface{c}, p)
}

// TradingPairFromString2 makes a TradingPair out of a string
func TradingPairFromString2(codeSize int8, converters []AssetConverterInterface, p string) (*TradingPair, error) {
	var base Asset
	var quote Asset
	var e error

	baseString := p[0:codeSize]
	for _, c := range converters {
		base, e = c.FromString(baseString)
		if e == nil {
			break
		}
	}
	if e != nil {
		return nil, fmt.Errorf("base asset could not be converted using any of the converters in the list of %d converters: %s", len(converters), baseString)
	}

	quoteString := p[codeSize : codeSize*2]
	for _, c := range converters {
		quote, e = c.FromString(quoteString)
		if e == nil {
			break
		}
	}
	if e != nil {
		return nil, fmt.Errorf("quote asset could not be converted using any of the converters in the list of %d converters: %s", len(converters), quoteString)
	}

	return &TradingPair{Base: base, Quote: quote}, nil
}

// TradingPairs2Strings converts the trading pairs to an array of strings
func TradingPairs2Strings(c AssetConverterInterface, delim string, pairs []TradingPair) (map[TradingPair]string, error) {
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

// TradingPairs2Strings2 converts the trading pairs to an array of strings
func TradingPairs2Strings2(c AssetConverterInterface, delim string, pairs []*TradingPair) (map[TradingPair]string, error) {
	m := map[TradingPair]string{}
	for _, p := range pairs {
		pairString, e := p.ToString(c, delim)
		if e != nil {
			return nil, e
		}
		m[*p] = pairString
	}
	return m, nil
}
