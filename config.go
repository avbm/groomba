package main

import (
	"fmt"

	"github.com/spf13/viper"
)

// Config stores the configuration for Groomba
type Config struct {
	StaleAgeThreshold int      `yaml:"stale_age_threshold"`
	StaticBranches    []string `yaml:"static_branches"`
}

func getConfig(configPath string) (*Config, error) {
	viper.SetConfigName(".groomba")
	viper.AddConfigPath(configPath) // should be "." except for tests

	viper.SetDefault("StaleAgeThreshold", 14)
	viper.SetDefault("StaticBranches", []string{"main", "master", "production"})

	if err := viper.BindEnv("StaleAgeThreshold", "GROOMBA_STALE_AGE_THRESHOLD"); err != nil {
		return nil, fmt.Errorf("getConfig: failed to bind env StaleAgeThreshold: %s", err)
	}
	if err := viper.BindEnv("StaticBranches", "GROOMBA_STATIC_BRANCHES"); err != nil {
		return nil, fmt.Errorf("getConfig: failed to bind env StaticBranches: %s", err)
	}

	err := viper.ReadInConfig()
	if err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("getConfig: failed to read in config: %s", err)
		}
	}

	var cfg Config
	err = viper.Unmarshal(&cfg)
	if err != nil {
		return nil, fmt.Errorf("getConfig: failed to unmarshal config: %s", err)
	}

	return &cfg, nil
}
