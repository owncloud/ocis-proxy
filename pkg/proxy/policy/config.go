package policy

// Policy enables us to use multiple directors.
type Policy struct {
	Name   Name    `mapstructure:"name"`
	Routes []Route `mapstructure:"routes"`
}

// Route define forwarding routes
type Route struct {
	Endpoint    Endpoint `mapstructure:"endpoint"`
	Backend     string   `mapstructure:"backend"`
	ApacheVHost bool     `mapstructure:"apache-vhost"`
}
