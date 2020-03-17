package policy

import (
	"net/url"
)

// Policy enables us to use multiple directors.
type Policy struct {
	Name   Name    `mapstructure:"name"`
	Routes []Route `mapstructure:"routes"`
}

// StaticPolicyConfig defines configuration for the StaticPolicyStrategy
type StaticPolicyConfig struct {
	PolicyName Name `mapstructure:"policy_name"`
}

// StrategyConfig is the generic untyped configuration for policy-strategies
type StrategyConfig struct {
	Name   string                 `mapstructure:"type"`
	Config map[string]interface{} `mapstructure:"config"`
}

// Route define forwarding routes
type Route struct {
	Endpoint    Endpoint `mapstructure:"endpoint"`
	Backend     *url.URL `mapstructure:"backend"`
	ApacheVHost bool     `mapstructure:"apache-vhost"`
}
