package groomba

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/stretchr/testify/assert"
)

func CheckTestInitError(err error, msg ...string) {
	msg = append([]string{"Failed to initialize test"}, msg...)
	CheckIfError(err, msg...)
}

type MockAuthenticator struct {
	auth transport.AuthMethod
}

func (ma *MockAuthenticator) Get() transport.AuthMethod {
	return ma.auth
}

func InitTest() {
	// cleanup dirs from previous tests
	os.RemoveAll("testdata/src")
	os.RemoveAll("testdata/dst")

	// create source repo
	err := os.MkdirAll("testdata/src", 0755)
	if err != nil {
		log.Fatal("failed to setup for test")
	}
	err = os.Chdir("testdata/src")
	if err != nil {
		log.Fatal("failed to setup for test")
	}

	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	nowDate := today.Format(time.RFC3339)
	zeroDate := today.AddDate(0, 0, -20).Format(time.RFC3339)
	staleDate := today.AddDate(0, 0, -19).Format(time.RFC3339)
	freshDate := today.AddDate(0, 0, -5).Format(time.RFC3339)
	gitCommands := [][]string{
		{zeroDate, "init"},
		{zeroDate, "config user.email 'test@user.com'"},
		{zeroDate, "config user.name 'Test User'"},
		{zeroDate, fmt.Sprintf("commit --allow-empty -am Initial_commit --date \"%v\"", zeroDate)},
		{staleDate, "checkout -b IsStale"},
		{staleDate, fmt.Sprintf("commit --allow-empty -am Stale_commit --date \"%v\"", staleDate)},
		{staleDate, "checkout -b stale/IsStale1"}, // Ensure "already up-to-date" is not returned as error
		{staleDate, "checkout -b IsStale2"},
		{freshDate, "checkout -b IsFresh"},
		{freshDate, fmt.Sprintf("commit --allow-empty -am Fresh_commit --date \"%v\"", freshDate)},
		{freshDate, "checkout -b IsFresh2"},
		{nowDate, "checkout IsStale"},
		{nowDate, "checkout -b StaleCommitFreshCommitter"},
		{nowDate, "commit --allow-empty -am Stale_commit_2 --date 2020-01-02"},
		{nowDate, "rebase HEAD~1"},
		{nowDate, "checkout master"},
	}
	for _, value := range gitCommands {
		committerDate, args := value[0], value[1]
		cmd := exec.Command("git", strings.Split(args, " ")...)
		cmd.Env = append(os.Environ(),
			fmt.Sprintf("GIT_COMMITTER_DATE=\"%v\"", committerDate),
		)
		err := cmd.Run()
		CheckTestInitError(err, "git", args)
	}
	err = os.Chdir("../..")
	if err != nil {
		log.Fatal("failed to setup for test")
	}

	// create cloned repo
	err = exec.Command("git", "clone", "testdata/src", "testdata/dst").Run()
	CheckTestInitError(err)
}

func ExampleGroomba_PrintBranchesGroupbyAuthor() {
	InitTest()

	cfg, _ := GetConfig(".")
	repo, _ := git.PlainOpen("testdata/dst")
	g := Groomba{cfg: cfg, repo: repo, auth: &MockAuthenticator{}}

	fb, _ := g.FilterBranches(time.Now())
	_ = g.PrintBranchesGroupbyAuthor(fb)
	// Output:
	// Test:
	//     - name: refs/remotes/origin/IsStale
	//       age: 19d
	//     - name: refs/remotes/origin/IsStale2
	//       age: 19d
}

func TestGroomba(t *testing.T) {
	InitTest()

	os.Setenv("GROOMBA_PREFIX", "stale/")
	cfg, _ := GetConfig(".")
	repo, _ := git.PlainOpen("testdata/dst")
	g := Groomba{cfg: cfg, repo: repo, auth: &MockAuthenticator{}}
	t.Run("main branch should be static", func(t *testing.T) {
		a := assert.New(t)
		a.Equal(true, g.IsStaticBranch("refs/remotes/origin/main"))
		a.Equal(true, g.IsStaticBranch("refs/remotes/origin/master"))
	})

	today := time.Now()
	fb, _ := g.FilterBranches(today)
	t.Run("Only stale branches should be detected", func(t *testing.T) {
		a := assert.New(t)

		a.Equal(2, len(fb))
		actual := fb[0].Name().Short()
		a.Equal("origin/IsStale", actual)
	})

	err := g.MoveStaleBranches(fb)
	assert.Nil(t, err)

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

	t.Run("fresh branch with stale commit date should not be removed from origin", func(t *testing.T) {
		a := assert.New(t)
		_, err := upstream.Reference("refs/heads/StaleCommitFreshCommitter", false)
		a.Nil(err)
	})

	t.Run("stale branch should be removed from origin", func(t *testing.T) {
		a := assert.New(t)
		_, err := upstream.Reference("refs/heads/IsStale", false)
		a.NotNil(err)
		if err != nil {
			a.Equal("reference not found", err.Error())
		}
	})

	t.Run("stale branch should be renamed at origin", func(t *testing.T) {
		a := assert.New(t)
		_, err := upstream.Reference("refs/heads/stale/IsStale", false)
		a.Nil(err)
	})

	t.Run("origin should have exactly 7 branches", func(t *testing.T) {
		a := assert.New(t)
		count := 0
		b, _ := upstream.Branches()
		err := b.ForEach(func(ref *plumbing.Reference) error {
			count++
			return nil
		})
		a.Nil(err)
		a.Equal(7, count)
	})
}

func TestGroombaDryRun(t *testing.T) {
	InitTest()

	os.Setenv("GROOMBA_DRY_RUN", "true")
	cfg, _ := GetConfig(".")
	repo, _ := git.PlainOpen("testdata/dst")
	g := Groomba{cfg: cfg, repo: repo, auth: &MockAuthenticator{}}
	t.Run("main branch should be static", func(t *testing.T) {
		a := assert.New(t)
		a.Equal(true, g.IsStaticBranch("refs/remotes/origin/main"))
		a.Equal(true, g.IsStaticBranch("refs/remotes/origin/master"))
	})

	today := time.Now()
	fb, _ := g.FilterBranches(today)
	t.Run("Only stale branches should be detected", func(t *testing.T) {
		a := assert.New(t)

		a.Equal(2, len(fb))
		actual := fb[0].Name().Short()
		a.Equal("origin/IsStale", actual)
	})

	err := g.MoveStaleBranches(fb)
	assert.Nil(t, err)

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

	t.Run("fresh branch with stale commit date should not be removed from origin", func(t *testing.T) {
		a := assert.New(t)
		_, err := upstream.Reference("refs/heads/StaleCommitFreshCommitter", false)
		a.Nil(err)
	})

	t.Run("stale branch should not be removed from origin in dry_run mode", func(t *testing.T) {
		a := assert.New(t)
		_, err := upstream.Reference("refs/heads/IsStale", false)
		a.Nil(err)
	})

	t.Run("stale branch should not be renamed at origin in dry_run mode", func(t *testing.T) {
		a := assert.New(t)
		_, err := upstream.Reference("refs/heads/stale/IsStale", false)
		a.NotNil(err)
		if err != nil {
			a.Equal("reference not found", err.Error())
		}
	})

	t.Run("origin should have exactly 7 branches", func(t *testing.T) {
		a := assert.New(t)
		count := 0
		b, _ := upstream.Branches()
		err := b.ForEach(func(ref *plumbing.Reference) error {
			count++
			return nil
		})
		a.Nil(err)
		a.Equal(7, count)
	})
}

func TestGroombaPrefix(t *testing.T) {
	InitTest()

	os.Setenv("GROOMBA_PREFIX", "zzz/")
	os.Setenv("GROOMBA_DRY_RUN", "false")
	cfg, _ := GetConfig(".")
	repo, _ := git.PlainOpen("testdata/dst")
	g := Groomba{cfg: cfg, repo: repo, auth: &MockAuthenticator{}}
	t.Run("main branch should be static", func(t *testing.T) {
		a := assert.New(t)
		a.Equal(true, g.IsStaticBranch("refs/remotes/origin/main"))
		a.Equal(true, g.IsStaticBranch("refs/remotes/origin/master"))
	})

	today := time.Now()
	fb, _ := g.FilterBranches(today)
	t.Run("Only stale branches should be detected", func(t *testing.T) {
		a := assert.New(t)

		a.Equal(3, len(fb)) // Expected count is 3 here unlike 2 above since stale/IsStale is also detected here
		actual := fb[0].Name().Short()
		a.Equal("origin/IsStale", actual)
	})

	err := g.MoveStaleBranches(fb)
	assert.Nil(t, err)

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

	t.Run("fresh branch with stale commit date should not be removed from origin", func(t *testing.T) {
		a := assert.New(t)
		_, err := upstream.Reference("refs/heads/StaleCommitFreshCommitter", false)
		a.Nil(err)
	})

	t.Run("stale branch should be removed from origin", func(t *testing.T) {
		a := assert.New(t)
		_, err := upstream.Reference("refs/heads/IsStale", false)
		a.NotNil(err)
		if err != nil {
			a.Equal("reference not found", err.Error())
		}
	})

	t.Run("stale branch should be renamed at origin with correct prefix", func(t *testing.T) {
		a := assert.New(t)
		_, err := upstream.Reference("refs/heads/zzz/IsStale", false)
		a.Nil(err)
	})

	t.Run("origin should have exactly 7 branches", func(t *testing.T) {
		a := assert.New(t)
		count := 0
		b, _ := upstream.Branches()
		err = b.ForEach(func(ref *plumbing.Reference) error {
			count++
			return nil
		})
		a.Nil(err)
		a.Equal(7, count)
	})
}

func InitClobberTest() {
	// cleanup dirs from previous tests
	os.RemoveAll("testdata/src")
	os.RemoveAll("testdata/dst")

	// create source repo
	err := os.MkdirAll("testdata/src", 0755)
	if err != nil {
		log.Fatal("error setting up clobber test")
	}
	err = os.Chdir("testdata/src")
	if err != nil {
		log.Fatal("error setting up clobber test")
	}
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	nowDate := today.Format(time.RFC3339)
	zeroDate := today.AddDate(0, 0, -20).Format(time.RFC3339)
	staleDate := today.AddDate(0, 0, -19).Format(time.RFC3339)
	freshDate := today.AddDate(0, 0, -5).Format(time.RFC3339)
	gitCommands := [][]string{
		{zeroDate, "init"},
		{zeroDate, "config user.email 'test@user.com'"},
		{zeroDate, "config user.name 'Test User'"},
		{zeroDate, fmt.Sprintf("commit --allow-empty -am Initial_commit --date \"%v\"", zeroDate)},
		{staleDate, "checkout -b stale/IsStale"}, // create branch to clobber
		{staleDate, fmt.Sprintf("commit --allow-empty -am Stale_commit --date \"%v\"", staleDate)},
		{staleDate, "checkout -b stale/IsStale3"}, // create second branch to clobber
		{staleDate, "checkout master"},
		{staleDate, "checkout -b IsStale"},
		{staleDate, fmt.Sprintf("commit --allow-empty -am Stale_commit2 --date \"%v\"", staleDate)},
		{staleDate, "checkout -b IsStale2"},
		{staleDate, "checkout -b IsStale3"},
		{freshDate, "checkout -b IsFresh"},
		{freshDate, fmt.Sprintf("commit --allow-empty -am Fresh_commit --date \"%v\"", freshDate)},
		{freshDate, "checkout -b IsFresh2"},
		{nowDate, "checkout IsStale"},
		{nowDate, "checkout -b StaleCommitFreshCommitter"},
		{nowDate, "commit --allow-empty -am Stale_commit_2 --date 2020-01-02"},
		{nowDate, "rebase HEAD~1"},
		{nowDate, "checkout master"},
	}
	for _, value := range gitCommands {
		committerDate, args := value[0], value[1]
		cmd := exec.Command("git", strings.Split(args, " ")...)
		cmd.Env = append(os.Environ(),
			fmt.Sprintf("GIT_COMMITTER_DATE=\"%v\"", committerDate),
		)
		err := cmd.Run()
		CheckTestInitError(err, "git", args)
	}
	err = os.Chdir("../..")
	if err != nil {
		log.Fatal("error setting up clobber test")
	}

	// create cloned repo
	err = exec.Command("git", "clone", "testdata/src", "testdata/dst").Run()
	CheckTestInitError(err)
}

func TestGroombaClobber(t *testing.T) {
	InitClobberTest()

	os.Setenv("GROOMBA_PREFIX", "stale/")
	os.Setenv("GROOMBA_CLOBBER", "false")
	cfg, _ := GetConfig(".")
	repo, _ := git.PlainOpen("testdata/dst")
	g := Groomba{cfg: cfg, repo: repo, auth: &MockAuthenticator{}}

	today := time.Now()
	fb, _ := g.FilterBranches(today)
	t.Run("Only stale branches should be detected", func(t *testing.T) {
		a := assert.New(t)
		a.Equal(3, len(fb))
		actual := fb[0].Name().Short()
		a.Equal("origin/IsStale", actual)
	})

	//upstream, _ := git.PlainOpen("testdata/src")

	t.Run("MoveBranch should fail when clobber disabled", func(t *testing.T) {
		a := assert.New(t)
		err := g.MoveBranch("IsStale")
		expectedErrMsg := "branch: IsStale failed on operation copy with error: non-fast-forward update: refs/heads/stale/IsStale"
		a.NotNil(err)
		if err != nil {
			a.Equal(expectedErrMsg, err.Error())
		}
	})

	t.Run("MoveStaleBranches should continue with failures when clobber disabled", func(t *testing.T) {
		a := assert.New(t)
		err := g.MoveStaleBranches(fb)
		expectedErrMsg := []string{"branch: IsStale failed on operation copy with error: non-fast-forward update: refs/heads/stale/IsStale",
			"branch: IsStale3 failed on operation copy with error: non-fast-forward update: refs/heads/stale/IsStale3"}
		sort.Strings(expectedErrMsg)
		a.NotNil(err)
		if err != nil {
			a.Equal(strings.Join(expectedErrMsg, "\n"), err.Error())
		}
	})

	os.Setenv("GROOMBA_CLOBBER", "true")
	cfgC, _ := GetConfig(".")
	gc := Groomba{cfg: cfgC, repo: repo, auth: &MockAuthenticator{}}
	t.Run("MoveBranch should succeed when clobber enabled", func(t *testing.T) {
		a := assert.New(t)
		err := gc.MoveBranch("IsStale")
		a.Nil(err)
	})
}
