package config

import (
	"github.com/perses/common/config"
)

const defaultPath = "/schemas"

type Config struct {
	SchemasPath string `yaml:"schemas_path,omitempty"`
}

func (c *Config) Verify() error {
	if len(c.SchemasPath) == 0 {
		c.SchemasPath = defaultPath
	}
	return nil
}

// Resolve retrieves the configuration data, either from a config file or from env
func Resolve(configFile string) (*Config, error) {
	c := &Config{}

	return c, config.NewResolver().
		SetEnvPrefix("PERSES_POC_CUELANG").
		SetConfigFile(configFile).
		Resolve(c).Verify()
}
