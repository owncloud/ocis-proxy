package policy

import (
	"net/url"
	"reflect"
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

func Decoder(srcT reflect.Type, tgtT reflect.Type, val interface{}) (interface{}, error) {
	if srcT.Name() == "string" && tgtT.Name() == "URL" {
		return url.Parse(val.(string))
	}

	if srcT.Name() == "" && tgtT.Name() == "PolicyStrategy" {
		return val, nil
	}

	return val, nil

}

/*
func defaultPolicies() []Policy {
	return []Policy{
		Policy{
			Name: "reva",
			Routes: []Route{
				{
					Endpoint: "/",
					Backend:  "http://localhost:9100",
				},
				{
					Endpoint: "/.well-known/",
					Backend:  "http://localhost:9130",
				},
				{
					Endpoint: "/konnect/",
					Backend:  "http://localhost:9130",
				},
				{
					Endpoint: "/signin/",
					Backend:  "http://localhost:9130",
				},
				{
					Endpoint: "/ocs/",
					Backend:  "http://localhost:9140",
				},
				{
					Endpoint: "/remote.php/",
					Backend:  "http://localhost:9140",
				},
				{
					Endpoint: "/dav/",
					Backend:  "http://localhost:9140",
				},
				{
					Endpoint: "/webdav/",
					Backend:  "http://localhost:9140",
				},
			},
		},
		{
			Name: "oc10",
			Routes: []Route{
				{
					Endpoint: "/",
					Backend:  "http://localhost:9100",
				},
				{
					Endpoint: "/.well-known/",
					Backend:  "http://localhost:9130",
				},
				{
					Endpoint: "/konnect/",
					Backend:  "http://localhost:9130",
				},
				{
					Endpoint: "/signin/",
					Backend:  "http://localhost:9130",
				},
				{
					Endpoint:    "/ocs/",
					Backend:     "https://demo.owncloud.com",
					ApacheVHost: true,
				},
				{
					Endpoint:    "/remote.php/",
					Backend:     "https://demo.owncloud.com",
					ApacheVHost: true,
				},
				{
					Endpoint:    "/dav/",
					Backend:     "https://demo.owncloud.com",
					ApacheVHost: true,
				},
				{
					Endpoint:    "/webdav/",
					Backend:     "https://demo.owncloud.com",
					ApacheVHost: true,
				},
			},
		},
	}
}

*/
