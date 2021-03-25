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
	"os"
	"strings"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"

	"gopkg.in/yaml.v3"
)

// Groomba base type to store config and other shared references
type Groomba struct {
	cfg  *Config
	repo *git.Repository
}

// CheckIfError should be used to naively panic if an error is not nil.
func CheckIfError(err error, prefix ...string) {
	if err == nil {
		return
	}

	fmt.Printf("\x1b[31;1m%s\x1b[0m\n", fmt.Sprintf("ERROR: %s %s", prefix, err))
	os.Exit(1)
}

func NewGroomba(config *Config, repo *git.Repository) Groomba {
	return Groomba{
		cfg:  config,
		repo: repo,
	}
}

func (g Groomba) IsStaticBranch(name string) bool {
	for _, b := range g.cfg.StaticBranches {
		if fmt.Sprintf("refs/remotes/origin/%s", b) == name {
			return true
		}
	}
	return false
}

func (g Groomba) FilterBranches(referenceDate time.Time) ([]*plumbing.Reference, error) {
	branchList, err := g.repo.References() //Branches()
	if err != nil {
		return nil, err
	}

	filteredBranches := []*plumbing.Reference{}
	branchList.ForEach(func(ref *plumbing.Reference) error {
		if ref.Type() == plumbing.HashReference && ref.Name().IsRemote() &&
			!g.IsStaticBranch(ref.Name().String()) &&
			!strings.HasPrefix(ref.Name().String(), "refs/remotes/origin/revert") &&
			!strings.HasPrefix(ref.Name().String(), "refs/remotes/origin/cherry-pick") &&
			!strings.HasPrefix(ref.Name().String(), fmt.Sprintf("refs/remotes/origin/%s", g.cfg.Prefix)) {

			commit, err := g.repo.CommitObject(ref.Hash())
			if err != nil {
				fmt.Printf("WARN: failed to read reference: %s, err: %s\n", ref, err)
			}

			t, err := time.ParseDuration(fmt.Sprintf("%dh", g.cfg.StaleAgeThreshold*24))
			if err != nil {
				fmt.Printf("WARN: failed to calculate age for ref: %s, err: %s\n", ref, err)
			}
			if referenceDate.Sub(commit.Committer.When) > t {
				filteredBranches = append(filteredBranches, ref)
			}
		}
		return nil
	})
	return filteredBranches, nil
}

func (g Groomba) PrintBranchesGroupbyAuthor(branches []*plumbing.Reference) error {
	type Branch struct {
		Name string
		Age  string
	}
	authors := make(map[string][]*Branch)
	for _, ref := range branches {
		commit, err := g.repo.CommitObject(ref.Hash())
		if err != nil {
			return err
		}

		b := &Branch{
			Name: ref.Name().String(),
			Age:  fmt.Sprintf("%dd", int64(time.Since(commit.Committer.When).Hours()/24)),
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

func (g Groomba) MoveBranch(refName string) error {
	newRefName := g.cfg.Prefix + refName
	if g.cfg.DryRun {
		fmt.Printf("INFO: Would have moved branch %s to %s -- skipping since dry_run=true\n", refName, newRefName)
		return nil
	}
	fmt.Printf("INFO:   copy %s to %s\n", refName, newRefName)
	renameSpec := config.RefSpec(fmt.Sprintf("refs/remotes/origin/%s:refs/heads/%s", refName, newRefName))
	err := g.repo.Push(&git.PushOptions{
		RemoteName: "origin",
		RefSpecs:   []config.RefSpec{renameSpec},
		Force:      g.cfg.Clobber,
	})

	if err != nil {
		if err != git.NoErrAlreadyUpToDate {
			return err
		}
	}
	fmt.Printf("INFO:   delete %s\n", refName)
	deleteSpec := config.RefSpec(fmt.Sprintf(":refs/heads/%s", refName))
	err = g.repo.Push(&git.PushOptions{
		RemoteName: "origin",
		RefSpecs:   []config.RefSpec{deleteSpec},
	})
	if err != nil {
		if err != git.NoErrAlreadyUpToDate {
			return err
		}
	}
	return nil
}

func (g Groomba) MoveStaleBranches(branches []*plumbing.Reference) error {
	for _, ref := range branches {
		fmt.Printf("INFO: Moving branch %s\n", ref.Name().Short())
		refName := ref.Name().Short()[7:]
		err := g.MoveBranch(refName)
		if err != nil {
			return err
		}
	}
	return nil
}
