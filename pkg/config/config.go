package config

import "github.com/owncloud/ocis-proxy/pkg/proxy/policy"

// Log defines the available logging configuration.
type Log struct {
	Level  string
	Pretty bool
	Color  bool
}

// Debug defines the available debug configuration.
type Debug struct {
	Addr   string
	Token  string
	Pprof  bool
	Zpages bool
}

// HTTP defines the available http configuration.
type HTTP struct {
	Addr      string
	Namespace string
	Root      string
}

// Tracing defines the available tracing configuration.
type Tracing struct {
	Enabled   bool
	Type      string
	Endpoint  string
	Collector string
	Service   string
}

// Asset defines the available asset configuration.
type Asset struct {
	Path string
}

// Policy enables us to use multiple directors.
type Policy struct {
	Name   string
	Routes []Route
}

// OIDC handles authentication related configuration.
type OIDC struct {
	Endpoint    string
	Realm       string
	SigningAlgs []string
	Insecure    bool
	Disabled    bool
}

// Route define forwarding routes
type Route struct {
	Endpoint    string
	Backend     string
	ApacheVHost bool
}

// Config combines all available configuration parts.
type Config struct {
	File           string
	Log            Log
	Debug          Debug
	HTTP           HTTP
	Tracing        Tracing
	Asset          Asset
	Policies       []policy.Policy
	OIDC           OIDC
	PolicyStrategy policy.StrategyConfig
}

// New initializes a new configuration
func New() *Config {
	return &Config{}
}
