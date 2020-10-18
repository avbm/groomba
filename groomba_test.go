package main

import (
	"os"
	"os/exec"
	"strings"
	"testing"
)

func CheckTestInitError(err error) {
	CheckIfError(err, "Failed to initialize tests")
}

func TestInit(t *testing.T) {
	// cleanup dirs from previous tests
	os.RemoveAll("testdata/src")
	os.RemoveAll("testdata/dst")

	// create source repo
	os.MkdirAll("testdata/src", 0755)
	os.Chdir("testdata/src")
	gitCommands := []string{
		"init",
		"commit --allow-empty -am Initial_commit --date 2020-01-01",
		"checkout -b IsStale",
		"commit --allow-empty -am Stale_commit --date 2020-01-02",
		"checkout -b IsFresh",
		"commit --allow-empty -am Fresh_commit --date 2020-01-15",
	}
	for _, cmd := range gitCommands {
		err := exec.Command("git", strings.Split(cmd, " ")...).Run()
		CheckTestInitError(err)
	}
	os.Chdir("../..")

	// create cloned repo
	err := exec.Command("git", "clone", "testdata/src", "testdata/dst").Run()
	CheckTestInitError(err)
}

func TestGroomba(t *testing.T) {
	TestInit(t)
}
