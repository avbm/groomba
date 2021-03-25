package groomba

/*
   Copyright 2020 Amod Mulay

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

import (
	"fmt"

	"github.com/spf13/viper"
)

// Config stores the configuration for Groomba
type Config struct {
	DryRun            bool     `yaml:"dry_run" toml:"dry_run"`
	Prefix            string   `yaml:"prefix" toml:"prefix"`
	StaleAgeThreshold int      `yaml:"stale_age_threshold" toml:"stale_age_threshold"`
	StaticBranches    []string `yaml:"static_branches" toml:"static_branches"`
}

func GetConfig(configPath string) (*Config, error) {
	viper.SetConfigName(".groomba")
	viper.AddConfigPath(configPath) // should be "." except for tests

	viper.RegisterAlias("DryRun", "dry_run")
	viper.SetDefault("stale_age_threshold", 14)
	viper.RegisterAlias("StaleAgeThreshold", "stale_age_threshold")
	viper.SetDefault("static_branches", []string{"main", "master", "production"})
	viper.RegisterAlias("StaticBranches", "static_branches")
	viper.SetDefault("prefix", "stale/")

	if err := viper.BindEnv("stale_age_threshold", "GROOMBA_STALE_AGE_THRESHOLD"); err != nil {
		return nil, fmt.Errorf("getConfig: failed to bind env stale_age_threshold: %s", err)
	}
	if err := viper.BindEnv("static_branches", "GROOMBA_STATIC_BRANCHES"); err != nil {
		return nil, fmt.Errorf("getConfig: failed to bind env static_branches: %s", err)
	}
	if err := viper.BindEnv("prefix", "GROOMBA_PREFIX"); err != nil {
		return nil, fmt.Errorf("getConfig: failed to bind env prefix: %s", err)
	}
	if err := viper.BindEnv("dry_run", "GROOMBA_DRY_RUN"); err != nil {
		return nil, fmt.Errorf("getConfig: failed to bind env dry_run: %s", err)
	}

	err := viper.ReadInConfig()
	if err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("getConfig: failed to read in config: %s", err)
		}
	}

	// fmt.Printf("DEBUG: %v\n", viper.AllSettings())
	var cfg Config
	err = viper.Unmarshal(&cfg)
	if err != nil {
		return nil, fmt.Errorf("getConfig: failed to unmarshal config: %s", err)
	}

	return &cfg, nil
}
