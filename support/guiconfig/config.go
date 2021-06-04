package guiconfig

type Auth0 struct {
	Auth0Enabled bool   `valid:"-" toml:"AUTH0_ENABLED" json:"auth0_enabled"`
	Domain       string `valid:"-" toml:"DOMAIN"json:"domain"`
	ClientId     string `valid:"-" toml:"CLIENT_ID"json:"client_id"`
	Audience     string `valid:"-" toml:"AUDIENCE"json:"audience"`
}

type GUIConfig struct {
	Auth0 		*Auth0 `valid:"-" toml:"AUTH0" json:"auth0"`
	// -- separation -- //
	// DelegatedEnabled    bool   `valid:"-" toml:"DELEGATED_ENABLED" json:"delegated_enabled"`
	// DelegatedSigningUrl string `valid:"-" toml:"DELEGATED_SIGNING_URL" json:"delegated_signing_url"`
	// Callback            string `valid:"-" toml:"CALLBACK_URL" json:"callback_url"`
}