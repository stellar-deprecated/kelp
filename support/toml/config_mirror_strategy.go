package toml

import "github.com/stellar/kelp/api"

// ExchangeParamsToml is the toml representation of ExchangeParams
type ExchangeParamsToml []struct {
	Param string      `valid:"-" toml:"PARAM"`
	Value interface{} `valid:"-" toml:"VALUE"`
}

// ToExchangeParams converts object
func (t *ExchangeParamsToml) ToExchangeParams() []api.ExchangeParam {
	exchangeParams := []api.ExchangeParam{}
	for _, param := range *t {
		exchangeParams = append(exchangeParams, api.ExchangeParam{
			Param: param.Param,
			Value: param.Value,
		})
	}
	return exchangeParams
}

// ExchangeHeadersToml is the toml representation of ExchangeHeaders
type ExchangeHeadersToml []struct {
	Header string `valid:"-" toml:"HEADER"`
	Value  string `valid:"-" toml:"VALUE"`
}

// ToExchangeHeaders converts object
func (t *ExchangeHeadersToml) ToExchangeHeaders() []api.ExchangeHeader {
	apiHeaders := []api.ExchangeHeader{}
	for _, header := range *t {
		apiHeaders = append(apiHeaders, api.ExchangeHeader{
			Header: header.Header,
			Value:  header.Value,
		})
	}
	return apiHeaders
}
