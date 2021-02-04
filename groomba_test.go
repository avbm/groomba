package groomba

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/stretchr/testify/assert"
)

func CheckTestInitError(err error, msg ...string) {
	msg = append([]string{"Failed to initialize test"}, msg...)
	CheckIfError(err, msg...)
}

func TestInit(t *testing.T) {
	// cleanup dirs from previous tests
	os.RemoveAll("testdata/src")
	os.RemoveAll("testdata/dst")

	// create source repo
	os.MkdirAll("testdata/src", 0755)
	os.Chdir("testdata/src")
	now := time.Now()
	zeroDate := now.AddDate(0, 0, -20).Format(time.RFC3339)
	staleDate := now.AddDate(0, 0, -19).Format(time.RFC3339)
	freshDate := now.AddDate(0, 0, -5).Format(time.RFC3339)
	gitCommands := [][]string{
		[]string{fmt.Sprintf("GIT_COMMITTER_DATE=\"%v\"", zeroDate), "init"},
		[]string{fmt.Sprintf("GIT_COMMITTER_DATE=\"%v\"", zeroDate), "config user.email 'test@user.com'"},
		[]string{fmt.Sprintf("GIT_COMMITTER_DATE=\"%v\"", zeroDate), "config user.name 'Test User'"},
		[]string{fmt.Sprintf("GIT_COMMITTER_DATE=\"%v\"", zeroDate), fmt.Sprintf("commit --allow-empty -am Initial_commit --date \"%v\"", zeroDate)},
		[]string{fmt.Sprintf("GIT_COMMITTER_DATE=\"%v\"", staleDate), "checkout -b IsStale"},
		[]string{fmt.Sprintf("GIT_COMMITTER_DATE=\"%v\"", staleDate), fmt.Sprintf("commit --allow-empty -am Stale_commit --date \"%v\"", staleDate)},
		[]string{fmt.Sprintf("GIT_COMMITTER_DATE=\"%v\"", freshDate), "checkout -b IsFresh"},
		[]string{fmt.Sprintf("GIT_COMMITTER_DATE=\"%v\"", freshDate), fmt.Sprintf("commit --allow-empty -am Fresh_commit --date \"%v\"", freshDate)},
		[]string{fmt.Sprintf("GIT_COMMITTER_DATE=\"%v\"", now.Format(time.RFC3339)), "checkout IsStale"},
		[]string{fmt.Sprintf("GIT_COMMITTER_DATE=\"%v\"", now.Format(time.RFC3339)), "checkout -b StaleCommitFreshCommitter"},
		[]string{fmt.Sprintf("GIT_COMMITTER_DATE=\"%v\"", now.Format(time.RFC3339)), "commit --allow-empty -am Stale_commit_2 --date 2020-01-02"},
		[]string{fmt.Sprintf("GIT_COMMITTER_DATE=\"%v\"", now.Format(time.RFC3339)), "rebase HEAD~1"},
		[]string{fmt.Sprintf("GIT_COMMITTER_DATE=\"%v\"", now.Format(time.RFC3339)), "checkout master"},
	}
	for _, value := range gitCommands {
		env, args := value[0], value[1]
		cmd := exec.Command("git", strings.Split(args, " ")...)
		cmd.Env = append(os.Environ(),
			env,
		)
		err := cmd.Run()
		CheckTestInitError(err, "git", args)
	}
	os.Chdir("../..")

	// create cloned repo
	err := exec.Command("git", "clone", "testdata/src", "testdata/dst").Run()
	CheckTestInitError(err)
}

func TestGroomba(t *testing.T) {
	TestInit(t)

	cfg, _ := GetConfig(".")
	repo, _ := git.PlainOpen("testdata/dst")
	g := Groomba{cfg: cfg, repo: repo}
	t.Run("main branch should be static", func(t *testing.T) {
		a := assert.New(t)
		a.Equal(true, g.IsStaticBranch("refs/remotes/origin/main"))
		a.Equal(true, g.IsStaticBranch("refs/remotes/origin/master"))
	})

	today := time.Now()
	fb, _ := g.FilterBranches(today)
	t.Run("Only stale branch should be detected", func(t *testing.T) {
		a := assert.New(t)

		a.Equal(1, len(fb))
		actual := fb[0].Name().Short()
		a.Equal("origin/IsStale", actual)
	})

	g.MoveStaleBranches(fb)
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
		a.Equal("reference not found", err.Error())
	})

	t.Run("stale branch should be renamed at origin", func(t *testing.T) {
		a := assert.New(t)
		_, err := upstream.Reference("refs/heads/stale/IsStale", false)
		a.Nil(err)
	})

	t.Run("origin should have exactly 4 branches", func(t *testing.T) {
		a := assert.New(t)
		count := 0
		b, _ := upstream.Branches()
		b.ForEach(func(ref *plumbing.Reference) error {
			count++
			return nil
		})
		a.Equal(4, count)
	})
}
