package sdk

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/stellar/kelp/support/networking"
)

type ccxtMapper struct {
	timestamp int64
}

// makeHeaderMappingsFromNewTimestamp creates a new ccxtMapper so the timestamp can be consistent across HeaderFns and returns the required map
func makeHeaderMappingsFromNewTimestamp() map[string]networking.HeaderFnFactory {
	c := &ccxtMapper{
		timestamp: time.Now().Unix(),
	}

	return map[string]networking.HeaderFnFactory{
		"COINBASEPRO__CB-ACCESS-SIGN": networking.HeaderFnFactory(c.coinbaseSignFn),
		"TIMESTAMP":                   networking.HeaderFnFactory(c.timestampFn),
	}
}

func (c *ccxtMapper) coinbaseSignFn(base64EncodedSigningKey string) (networking.HeaderFn, error) {
	base64DecodedSigningKey, e := base64.StdEncoding.DecodeString(base64EncodedSigningKey)
	if e != nil {
		return nil, fmt.Errorf("could not decode signing key (%s): %s", base64EncodedSigningKey, e)
	}

	// return this inline method casted as a HeaderFn to work as a headerValue
	return networking.HeaderFn(func(method string, requestPath string, body string) string {
		uppercaseMethod := strings.ToUpper(method)
		payload := fmt.Sprintf("%d%s%s%s", c.timestamp, uppercaseMethod, requestPath, body)

		// sign
		mac := hmac.New(sha256.New, base64DecodedSigningKey)
		mac.Write([]byte(payload))
		signature := mac.Sum(nil)
		base64EncodedSignature := base64.StdEncoding.EncodeToString(signature)

		return base64EncodedSignature
	}), nil
}

func (c *ccxtMapper) timestampFn(_ string) (networking.HeaderFn, error) {
	return networking.HeaderFn(func(method string, requestPath string, body string) string {
		return strconv.FormatInt(c.timestamp, 10)
	}), nil
}
