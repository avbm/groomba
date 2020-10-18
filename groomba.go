package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"

	"gopkg.in/yaml.v3"
)

const (
	STALE_AGE_THRESHOLD = 10
)

var STATIC_BRANCHES = []string{"main", "mainline", "production", "staging", "canary", "govcloud", "release", "master", "gcp_staging", "gcp_beta", "gcp_production"}

// CheckIfError should be used to naively panics if an error is not nil.
func CheckIfError(err error, prefix ...string) {
	if err == nil {
		return
	}

	fmt.Printf("\x1b[31;1m%s\x1b[0m\n", fmt.Sprintf("ERROR: %s %s", prefix, err))
	os.Exit(1)
}

func isStaticBranch(name string) bool {
	for _, b := range STATIC_BRANCHES {
		if fmt.Sprintf("refs/remotes/origin/%s", b) == name {
			return true
		}
	}
	return false
}

func filterBranches(repo *git.Repository, threshold int, referenceDate time.Time) []*plumbing.Reference {
	branchList, err := repo.References() //Branches()
	CheckIfError(err)

	filteredBranches := []*plumbing.Reference{}
	branchList.ForEach(func(ref *plumbing.Reference) error {
		// fmt.Println(ref)
		// fmt.Println(ref.Name().String())
		if ref.Type() == plumbing.HashReference && ref.Name().IsRemote() &&
			!isStaticBranch(ref.Name().String()) &&
			!strings.HasPrefix(ref.Name().String(), "refs/remotes/origin/revert") &&
			!strings.HasPrefix(ref.Name().String(), "refs/remotes/origin/cherry-pick") &&
			!strings.HasPrefix(ref.Name().String(), "refs/remotes/origin/stale") {

			commit, err := repo.CommitObject(ref.Hash())
			CheckIfError(err)

			t, err := time.ParseDuration(fmt.Sprintf("%dh", threshold*24))
			CheckIfError(err)
			if referenceDate.Sub(commit.Author.When) > t {
				filteredBranches = append(filteredBranches, ref)
			}
		}
		// fmt.Println(ref.Name().String(), commit.Author.Name, commit.Author.When)
		return nil
	})
	return filteredBranches
}

func printBranchesGroupbyAuthor(repo *git.Repository, branches []*plumbing.Reference) {
	type Branch struct {
		Name string
		Age  string
	}
	authors := make(map[string][]*Branch)
	for _, ref := range branches {
		commit, err := repo.CommitObject(ref.Hash())
		CheckIfError(err)

		b := &Branch{
			Name: ref.Name().String(),
			Age:  fmt.Sprintf("%dd", int64(time.Since(commit.Author.When).Hours()/24)),
		}
		if _, ok := authors[commit.Author.Name]; ok {
			authors[commit.Author.Name] = append(authors[commit.Author.Name], b)
		} else {
			authors[commit.Author.Name] = []*Branch{b}
		}

	}

	a, err := yaml.Marshal(authors)
	CheckIfError(err)

	fmt.Println(string(a))
}

func moveBranch(repo *git.Repository, ref *plumbing.Reference) error {
	refName := ref.Name().Short()[7:]
	newRefName := "stale/" + refName
	fmt.Printf("INFO: Copy %s to %s\n", refName, newRefName)
	renameSpec := config.RefSpec(fmt.Sprintf("refs/remotes/origin/%s:refs/heads/%s", refName, newRefName))
	err := repo.Push(&git.PushOptions{
		RemoteName: "origin",
		RefSpecs:   []config.RefSpec{renameSpec},
	})

	if err != nil {
		return err
	}
	fmt.Printf("INFO: Delete %s\n", refName)
	deleteSpec := config.RefSpec(fmt.Sprintf(":refs/heads/%s", refName))
	err = repo.Push(&git.PushOptions{
		RemoteName: "origin",
		RefSpecs:   []config.RefSpec{deleteSpec},
	})
	return err
}

func moveStaleBranches(repo *git.Repository, branches []*plumbing.Reference) error {
	for _, ref := range branches {
		fmt.Printf("INFO: moving branch %s\n", ref.Name().Short())
		err := moveBranch(repo, ref)
		if err != nil {
			return err
		}
	}
	return nil
}

func main() {
	repo, _ := git.PlainOpen(".")
	repo.Fetch(&git.FetchOptions{
		RemoteName: "origin",
		RefSpecs:   []config.RefSpec{"refs/remotes/origin"},
		Depth:      1,
	})
	fb := filterBranches(repo, STALE_AGE_THRESHOLD, time.Now())
	printBranchesGroupbyAuthor(repo, fb)
	if len(fb) > 10 {
		fb = fb[:10]
	}
	// err := moveStaleBranches(repo, fb)
	// CheckIfError(err)
}
