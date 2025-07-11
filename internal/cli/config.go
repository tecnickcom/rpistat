package cli

import (
	"github.com/Vonage/gosrvlib/pkg/config"
	"github.com/Vonage/gosrvlib/pkg/validator"
)

const (
	// AppName is the name of the application executable.
	AppName = "rpistat"

	// appEnvPrefix is the prefix of the configuration environment variables.
	appEnvPrefix = "RPISTAT"

	// appShortDesc is the short description of the application.
	appShortDesc = "Web-Service to collect system usage statistics"

	// appLongDesc is the long description of the application.
	appLongDesc = "Web-Service to collect system usage statistics"

	// fieldTagName is the name of the tag containing the original JSON field name.
	fieldTagName = "mapstructure"
)

type cfgServer struct {
	Address string `mapstructure:"address" validate:"required,hostname_port"`
	Timeout int    `mapstructure:"timeout" validate:"required,min=1"`
}

type cfgServerMonitoring cfgServer

// cfgServers contains the configuration for all exposed servers.
type cfgServers struct {
	Monitoring cfgServerMonitoring `mapstructure:"monitoring" validate:"required"`
}

type cfgClientIpify struct {
	Address string `mapstructure:"address" validate:"required,url"`
	Timeout int    `mapstructure:"timeout" validate:"required,min=1"`
}

// cfgClients contains the configuration for all external clients.
type cfgClients struct {
	Ipify cfgClientIpify `mapstructure:"ipify" validate:"required"`
}

// appConfig contains the full application configuration.
type appConfig struct {
	config.BaseConfig `mapstructure:",squash" validate:"required"`

	Enabled bool       `mapstructure:"enabled"`
	Servers cfgServers `mapstructure:"servers" validate:"required"`
	Clients cfgClients `mapstructure:"clients" validate:"required"`
}

// SetDefaults sets the default configuration values in Viper.
func (c *appConfig) SetDefaults(v config.Viper) {
	v.SetDefault("enabled", true)

	v.SetDefault("servers.monitoring.address", ":65501")
	v.SetDefault("servers.monitoring.timeout", 15)

	v.SetDefault("clients.ipify.address", "https://api.ipify.org")
	v.SetDefault("clients.ipify.timeout", 1)
}

// Validate performs the validation of the configuration values.
func (c *appConfig) Validate() error {
	opts := []validator.Option{
		validator.WithFieldNameTag(fieldTagName),
		validator.WithCustomValidationTags(validator.CustomValidationTags()),
		validator.WithErrorTemplates(validator.ErrorTemplates()),
	}

	v, _ := validator.New(opts...)

	return v.ValidateStruct(c) //nolint:wrapcheck
}
