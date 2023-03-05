package groomba

/*
   Copyright 2021 Amod Mulay

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
	"strings"
	"sync"
	"time"

	"github.com/apex/log"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport"

	"gopkg.in/yaml.v3"
)

// Groomba base type to store config and other shared references
type Groomba struct {
	cfg  *Config
	repo *git.Repository
	auth Authenticator
}

// CheckIfError should be used to naively panic if an error is not nil.
func CheckIfError(err error, prefix ...string) {
	if err == nil {
		return
	}

	log.Fatalf("%s %s", prefix, err)
}

type Authenticator interface {
	Get() transport.AuthMethod
}

func NewGroomba(config *Config, repo *git.Repository, a Authenticator) Groomba {
	return Groomba{
		cfg:  config,
		repo: repo,
		auth: a,
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
				log.Warnf("failed to read reference: %s, err: %s", ref, err)
			}

			t, err := time.ParseDuration(fmt.Sprintf("%dh", g.cfg.StaleAgeThreshold*24))
			if err != nil {
				log.Warnf("failed to calculate age for ref: %s, err: %s", ref, err)
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

func (g Groomba) MoveBranch(refName string) *MoveBranchError {
	newRefName := g.cfg.Prefix + refName
	if g.cfg.DryRun {
		log.Infof("Would have moved branch %s to %s -- skipping since dry_run=true", refName, newRefName)
		return nil
	}
	log.Infof("  copy %s to %s", refName, newRefName)
	renameSpec := config.RefSpec(fmt.Sprintf("refs/remotes/origin/%s:refs/heads/%s", refName, newRefName))
	err := g.repo.Push(&git.PushOptions{
		RemoteName: "origin",
		RefSpecs:   []config.RefSpec{renameSpec},
		Force:      g.cfg.Clobber,
		Auth:       g.auth.Get(),
	})
	if err != nil && err != git.NoErrAlreadyUpToDate {
		log.Infof("  Failed to copy %s to %s with error: %s", refName, newRefName, err)
		return &MoveBranchError{branch: refName, operation: CopyBranch, err: err}
	}

	log.Infof("  delete %s", refName)
	deleteSpec := config.RefSpec(fmt.Sprintf(":refs/heads/%s", refName))
	err = g.repo.Push(&git.PushOptions{
		RemoteName: "origin",
		RefSpecs:   []config.RefSpec{deleteSpec},
		Auth:       g.auth.Get(),
	})
	if err != nil && err != git.NoErrAlreadyUpToDate {
		log.Infof("  Failed to delete %s", refName, err)
		return &MoveBranchError{branch: refName, operation: DeleteBranch, err: err}
	}

	return nil
}

func (g Groomba) MoveStaleBranches(branches []*plumbing.Reference) error {
	var wg sync.WaitGroup
	errCh := make(chan *MoveBranchError) //, len(branches))
	ch := make(chan string)

	branchNames := []string{}
	for _, ref := range branches {
		wg.Add(1)
		log.Debugf("ref: %s", ref.Name())
		branchNames = append(branchNames, ref.Name().Short()[7:])
		log.Debugf("branchNames: %s", branchNames)
	}
	log.Debugf("%s", branchNames)
	go func(branchNames []string) {
		// send branches to move to ch
		for _, ref := range branchNames {
			log.Debugf("sending ref: %s", ref)
			ch <- ref
		}
		close(ch)
	}(branchNames)
	for i := uint8(0); i < g.cfg.MaxConcurrency; i++ {
		// Create workers to move branches
		go func(ch <-chan string) {
			for refName := range ch {
				defer wg.Done()
				log.Infof("Moving branch %s", refName)
				err := g.MoveBranch(refName)
				log.Debugf("branch: %s, returned error: %s", refName, err)
				if err != nil {
					errCh <- err
				}
			}
		}(ch)
	}

	// channel to get aggregated list of errors
	errListCh := make(chan []MoveBranchError)
	defer close(errListCh)
	go func(errorListCh chan<- []MoveBranchError) {
		errList := []MoveBranchError{}
		for e := range errCh {
			log.Debugf("%s", e)
			errList = append(errList, *e)
		}
		errListCh <- errList
	}(errListCh)

	wg.Wait()
	close(errCh)

	errList := <-errListCh
	if len(errList) != 0 {
		return &MoveStaleBranchesError{errList: errList}
	}

	return nil
}
