package ipam

import (
	"github.com/lugnut42/openipam/internal/config"
)

var cfg *config.Config

// SetConfig sets the global configuration
func SetConfig(config *config.Config) {
	cfg = config
}

// GetConfig returns the global configuration
func GetConfig() *config.Config {
	return cfg
}
