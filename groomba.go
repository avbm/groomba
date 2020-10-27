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

// Groomba base type to store config
type Groomba struct {
	cfg *Config
}

// CheckIfError should be used to naively panics if an error is not nil.
func CheckIfError(err error, prefix ...string) {
	if err == nil {
		return
	}

	fmt.Printf("\x1b[31;1m%s\x1b[0m\n", fmt.Sprintf("ERROR: %s %s", prefix, err))
	os.Exit(1)
}

func (g Groomba) isStaticBranch(name string) bool {
	for _, b := range g.cfg.StaticBranches {
		if fmt.Sprintf("refs/remotes/origin/%s", b) == name {
			return true
		}
	}
	return false
}

func (g Groomba) filterBranches(repo *git.Repository, referenceDate time.Time) ([]*plumbing.Reference, error) {
	branchList, err := repo.References() //Branches()
	if err != nil {
		return nil, err
	}

	filteredBranches := []*plumbing.Reference{}
	branchList.ForEach(func(ref *plumbing.Reference) error {
		if ref.Type() == plumbing.HashReference && ref.Name().IsRemote() &&
			!g.isStaticBranch(ref.Name().String()) &&
			!strings.HasPrefix(ref.Name().String(), "refs/remotes/origin/revert") &&
			!strings.HasPrefix(ref.Name().String(), "refs/remotes/origin/cherry-pick") &&
			!strings.HasPrefix(ref.Name().String(), "refs/remotes/origin/stale") {

			commit, err := repo.CommitObject(ref.Hash())
			if err != nil {
				fmt.Printf("WARN: failed to read reference: %s, err: %s\n", ref, err)
			}

			t, err := time.ParseDuration(fmt.Sprintf("%dh", g.cfg.StaleAgeThreshold*24))
			if err != nil {
				fmt.Printf("WARN: failed to calculate age for ref: %s, err: %s\n", ref, err)
			}
			if referenceDate.Sub(commit.Author.When) > t {
				filteredBranches = append(filteredBranches, ref)
			}
		}
		return nil
	})
	return filteredBranches, nil
}

func (g Groomba) printBranchesGroupbyAuthor(repo *git.Repository, branches []*plumbing.Reference) error {
	type Branch struct {
		Name string
		Age  string
	}
	authors := make(map[string][]*Branch)
	for _, ref := range branches {
		commit, err := repo.CommitObject(ref.Hash())
		if err != nil {
			return err
		}

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
	if err != nil {
		return err
	}

	fmt.Println(string(a))
	return nil
}

func (g Groomba) moveBranch(repo *git.Repository, ref *plumbing.Reference) error {
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

func (g Groomba) moveStaleBranches(repo *git.Repository, branches []*plumbing.Reference) error {
	for _, ref := range branches {
		fmt.Printf("INFO: moving branch %s\n", ref.Name().Short())
		err := g.moveBranch(repo, ref)
		if err != nil {
			return err
		}
	}
	return nil
}

func main() {
	cfg, err := getConfig(".")
	CheckIfError(err)

	g := Groomba{cfg: cfg}
	repo, _ := git.PlainOpen(".")
	repo.Fetch(&git.FetchOptions{
		RemoteName: "origin",
		RefSpecs:   []config.RefSpec{"refs/remotes/origin"},
		Depth:      1,
	})
	fb, err := g.filterBranches(repo, time.Now())
	CheckIfError(err)

	err = g.printBranchesGroupbyAuthor(repo, fb)
	CheckIfError(err)

	err = g.moveStaleBranches(repo, fb)
	CheckIfError(err)
}
