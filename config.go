package groomba

import (
	"fmt"

	"github.com/spf13/viper"
)

// Config stores the configuration for Groomba
type Config struct {
	StaleAgeThreshold int      `yaml:"stale_age_threshold" toml:"stale_age_threshold"`
	StaticBranches    []string `yaml:"static_branches" toml:"static_branches"`
}

func GetConfig(configPath string) (*Config, error) {
	viper.SetConfigName(".groomba")
	viper.AddConfigPath(configPath) // should be "." except for tests

	viper.SetDefault("stale_age_threshold", 14)
	viper.RegisterAlias("StaleAgeThreshold", "stale_age_threshold")
	viper.SetDefault("static_branches", []string{"main", "master", "production"})
	viper.RegisterAlias("StaticBranches", "static_branches")

	if err := viper.BindEnv("stale_age_threshold", "GROOMBA_STALE_AGE_THRESHOLD"); err != nil {
		return nil, fmt.Errorf("getConfig: failed to bind env stale_age_threshold: %s", err)
	}
	if err := viper.BindEnv("static_branches", "GROOMBA_STATIC_BRANCHES"); err != nil {
		return nil, fmt.Errorf("getConfig: failed to bind env static_branches: %s", err)
	}

	err := viper.ReadInConfig()
	if err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("getConfig: failed to read in config: %s", err)
		}
	}

	fmt.Printf("DEBUG: %v\n", viper.AllSettings())
	var cfg Config
	err = viper.Unmarshal(&cfg)
	if err != nil {
		return nil, fmt.Errorf("getConfig: failed to unmarshal config: %s", err)
	}

	return &cfg, nil
}
