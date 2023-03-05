package groomba

import (
	"fmt"
	"os"
	"testing"

	"github.com/avbm/groomba/auth"
	"github.com/stretchr/testify/assert"
)

func TestConfig(t *testing.T) {
	cfg, err := GetConfig(".")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	t.Run("Default configs should load correctly", func(t *testing.T) {
		a := assert.New(t)
		a.Equal(auth.DefaultAuth, cfg.Auth)
		a.Equal(false, cfg.Clobber)
		a.Equal(false, cfg.DryRun)
		a.Equal(uint8(4), cfg.MaxConcurrency)
		a.Equal("stale/", cfg.Prefix)
		a.Equal(14, cfg.StaleAgeThreshold)
		a.Equal([]string{"main", "master", "production"}, cfg.StaticBranches)
	})

	cfg, err = GetConfig("testdata")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	t.Run("Configs from .groomba.yaml should override defaults", func(t *testing.T) {
		a := assert.New(t)
		a.Equal(auth.SSHAgentAuth, cfg.Auth)
		a.Equal(true, cfg.Clobber)
		a.Equal(true, cfg.DryRun)
		a.Equal(uint8(10), cfg.MaxConcurrency)
		a.Equal("zzz_", cfg.Prefix)
		a.Equal(10, cfg.StaleAgeThreshold)
		a.Equal([]string{"main", "teststatic"}, cfg.StaticBranches)
	})

	os.Setenv("GROOMBA_AUTH", "env-auth")
	os.Setenv("GROOMBA_CLOBBER", "false")
	os.Setenv("GROOMBA_DRY_RUN", "false")
	os.Setenv("GROOMBA_MAX_CONCURRENCY", "2")
	os.Setenv("GROOMBA_PREFIX", "zzx/")
	os.Setenv("GROOMBA_STALE_AGE_THRESHOLD", "7")
	os.Setenv("GROOMBA_STATIC_BRANCHES", "main,master")
	cfg, err = GetConfig("testdata")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	t.Run("Configs from Environment should override .groomba.yaml and defaults", func(t *testing.T) {
		a := assert.New(t)
		a.Equal("env-auth", string(cfg.Auth))
		a.Equal(false, cfg.Clobber)
		a.Equal(false, cfg.DryRun)
		a.Equal(uint8(2), cfg.MaxConcurrency)
		a.Equal("zzx/", cfg.Prefix)
		a.Equal(7, cfg.StaleAgeThreshold)
		a.Equal([]string{"main", "master"}, cfg.StaticBranches)
	})

	os.Setenv("GROOMBA_MAX_CONCURRENCY", "0")
	cfg, err = GetConfig("testdata")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	t.Run("Ensure if MaxConcurrency is set to 0 its overridden to 1", func(t *testing.T) {
		a := assert.New(t)
		a.Equal(uint8(1), cfg.MaxConcurrency)
	})

}
