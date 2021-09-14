package json

import (
	"fmt"
	"github.com/tidwall/gjson"
)

type GJsonParserWrapper struct{}

func NewJsonParserWrapper() *GJsonParserWrapper {
	return &GJsonParserWrapper{}
}

func (j GJsonParserWrapper) GetRawJsonValue(json []byte, path string) (string, error) {
	value := gjson.GetBytes(json, path)

	if value.Raw == "" {
		return "", fmt.Errorf("json parser wrapper error: could not find json for path %s in %s", path, json)
	}

	return value.Raw, nil
}

func (j GJsonParserWrapper) GetNum(json []byte, path string) (float64, error) {
	value := gjson.GetBytes(json, path)
	return value.Num, nil
}
