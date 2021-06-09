package guiconfig

type Auth0Config struct {
	Auth0Enabled bool   `valid:"-" toml:"AUTH0_ENABLED" json:"auth0_enabled"`
	Domain       string `valid:"-" toml:"DOMAIN"json:"domain"`
	ClientId     string `valid:"-" toml:"CLIENT_ID"json:"client_id"`
	Audience     string `valid:"-" toml:"AUDIENCE"json:"audience"`
}

type GUIConfig struct {
	Auth0Config 		*Auth0Config `valid:"-" toml:"AUTH0" json:"auth0"`
}