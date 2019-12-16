package postgresdb

import "fmt"

// Config takes in the information needed to connect to a postgres database
type Config struct {
	Host      string `toml:"HOST"`
	Port      uint16 `toml:"PORT"`
	DbName    string `toml:"DB_NAME"`
	User      string `toml:"USER"`
	Password  string `toml:"PASSWORD"`
	SSLEnable bool   `toml:"SSL_ENABLE"`
}

// GetHost returns the host of the database after defaulting if needed
func (c *Config) GetHost() string {
	return defaultStringValue(c.Host, "localhost")
}

// GetPort returns the port of the database after defaulting if needed
func (c *Config) GetPort() uint16 {
	if c.Port == 0 {
		return 5432
	}
	return c.Port
}

// GetDbName returns the name of the database after defaulting if needed
func (c *Config) GetDbName() string {
	return defaultStringValue(c.DbName, "kelp")
}

// GetUser returns the username to the database, no defaulting
func (c *Config) GetUser() string {
	return c.User
}

// GetPassword returns the password to the database, no defaulting
func (c *Config) GetPassword() string {
	return c.Password
}

// GetSSLMode returns the sslmode string to the database after converting the input boolean value
func (c *Config) GetSSLMode() string {
	if c.SSLEnable {
		return "enable"
	}
	return "disable"
}

func defaultStringValue(actual string, def string) string {
	if actual == "" {
		return def
	}
	return actual
}

// MakeConnectStringWithoutDB returns the string to be used when connecting to a postgres instance without a database
func (c *Config) MakeConnectStringWithoutDB() string {
	s := fmt.Sprintf("host=%s port=%d sslmode=%s", c.GetHost(), c.GetPort(), c.GetSSLMode())
	if c.GetUser() != "" {
		s = fmt.Sprintf("%s user=%s", s, c.GetUser())
	}
	if c.GetPassword() != "" {
		s = fmt.Sprintf("%s password=%s", s, c.GetPassword())
	}
	return s
}

// MakeConnectString returns the string to be used to open this db
func (c *Config) MakeConnectString() string {
	return fmt.Sprintf("%s dbname=%s", c.MakeConnectStringWithoutDB(), c.GetDbName())
}
