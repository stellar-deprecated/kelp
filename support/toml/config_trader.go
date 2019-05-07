package toml

import "github.com/stellar/kelp/api"

// ExchangeAPIKeysToml is the toml representation of ExchangeAPIKeys
type ExchangeAPIKeysToml []struct {
	Key    string `valid:"-" toml:"KEY"`
	Secret string `valid:"-" toml:"SECRET"`
}

// ToExchangeAPIKeys converts object
func (t *ExchangeAPIKeysToml) ToExchangeAPIKeys() []api.ExchangeAPIKey {
	apiKeys := []api.ExchangeAPIKey{}
	for _, apiKey := range *t {
		apiKeys = append(apiKeys, api.ExchangeAPIKey{
			Key:    apiKey.Key,
			Secret: apiKey.Secret,
		})
	}
	return apiKeys
}
