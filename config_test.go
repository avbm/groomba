package groomba

import (
	"fmt"
	"os"
	"testing"

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
		a.Equal(false, cfg.Clobber)
		a.Equal(false, cfg.DryRun)
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
		a.Equal(true, cfg.Clobber)
		a.Equal(true, cfg.DryRun)
		a.Equal("zzz_", cfg.Prefix)
		a.Equal(10, cfg.StaleAgeThreshold)
		a.Equal([]string{"main", "teststatic"}, cfg.StaticBranches)
	})

	os.Setenv("GROOMBA_CLOBBER", "false")
	os.Setenv("GROOMBA_DRY_RUN", "false")
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
		a.Equal(false, cfg.Clobber)
		a.Equal(false, cfg.DryRun)
		a.Equal("zzx/", cfg.Prefix)
		a.Equal(7, cfg.StaleAgeThreshold)
		a.Equal([]string{"main", "master"}, cfg.StaticBranches)
	})
}
