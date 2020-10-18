package main

import (
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/stretchr/testify/assert"
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

	t.Run("main branch should be static", func(t *testing.T) {
		a := assert.New(t)
		a.Equal(true, isStaticBranch("refs/remotes/origin/main"))
		a.Equal(true, isStaticBranch("refs/remotes/origin/master"))
	})

	repo, _ := git.PlainOpen("testdata/dst")
	today, _ := time.Parse(time.RFC3339, "2020-01-20T00:00:00Z")
	fb := filterBranches(repo, STALE_AGE_THRESHOLD, today)
	t.Run("stale branch should be detected", func(t *testing.T) {
		a := assert.New(t)

		a.Equal(1, len(fb))
		actual := fb[0].Name().Short()
		a.Equal("origin/IsStale", actual)
	})

	moveStaleBranches(repo, fb)
	upstream, _ := git.PlainOpen("testdata/src")
	t.Run("main branch should not be removed from origin", func(t *testing.T) {
		a := assert.New(t)
		_, err := upstream.Reference("refs/heads/master", false)
		a.Nil(err)
	})

	t.Run("fresh branch should not be removed from origin", func(t *testing.T) {
		a := assert.New(t)
		_, err := upstream.Reference("refs/heads/IsFresh", false)
		a.Nil(err)
	})

	t.Run("stale branch should be removed from origin", func(t *testing.T) {
		a := assert.New(t)
		_, err := upstream.Reference("refs/heads/IsStale", false)
		a.Equal("reference not found", err.Error())
	})

	t.Run("stale branch should be renamed at origin", func(t *testing.T) {
		a := assert.New(t)
		_, err := upstream.Reference("refs/heads/stale/IsStale", false)
		a.Nil(err)
	})

	t.Run("origin should have exactly 3 branches", func(t *testing.T) {
		a := assert.New(t)
		count := 0
		b, _ := upstream.Branches()
		b.ForEach(func(ref *plumbing.Reference) error {
			count++
			return nil
		})
		a.Equal(3, count)
	})
}
