package main

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfig(t *testing.T) {
	cfg, err := getConfig(".")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	t.Run("Default configs should load correctly", func(t *testing.T) {
		a := assert.New(t)
		a.Equal(14, cfg.StaleAgeThreshold)
		a.Equal([]string{"main", "master", "production"}, cfg.StaticBranches)
	})

	cfg, err = getConfig("testdata")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	t.Run("Configs from .groomba.toml should override defaults", func(t *testing.T) {
		a := assert.New(t)
		a.Equal(10, cfg.StaleAgeThreshold)
		a.Equal([]string{"main", "teststatic"}, cfg.StaticBranches)
	})

	os.Setenv("GROOMBA_STALE_AGE_THRESHOLD", "7")
	os.Setenv("GROOMBA_STATIC_BRANCHES", "main,master")
	cfg, err = getConfig("testdata")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	t.Run("Configs from Environment should override .groomba.toml and defaults", func(t *testing.T) {
		a := assert.New(t)
		a.Equal(7, cfg.StaleAgeThreshold)
		a.Equal([]string{"main", "master"}, cfg.StaticBranches)
	})
}
