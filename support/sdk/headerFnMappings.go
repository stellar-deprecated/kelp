package sdk

import (
	"github.com/stellar/kelp/support/networking"
)

var ccxtHeaderFnMappings = map[string]networking.HeaderFnFactory{
	"COINBASE__CB-ACCESS-KEY": networking.HeaderFnFactory(getCoinbaseSignatureFn),
}

func getCoinbaseSignatureFn(base64EncodedSigningKey string) networking.HeaderFn {
	return nil
	// base64DecodedSigningKey := base64Decode(base64EncodedSigningKey)

	// // return this inline method casted as a HeaderFn to work as a headerValue
	// return HeaderFn(func(method string, requestPath string, body string) string {
	// 	payload := fmt.Sprintf("%s%s%s%s", time.Now(), method, requestPath, body)
	// 	signature := signMessage(payload, base64DecodedSigningKey)
	// 	base64EncodedHeader := base64Encode(signature)
	// 	return base64EncodedHeader
	// })
}
