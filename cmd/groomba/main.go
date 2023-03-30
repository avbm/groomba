package main

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
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"

	"github.com/apex/log"
	"github.com/apex/log/handlers/cli"
	"github.com/avbm/groomba"
	"github.com/avbm/groomba/auth"
)

func main() {
	log.SetHandler(cli.Default)
	cfg, err := groomba.GetConfig(".")
	groomba.CheckIfError(err, "failed to get configs")

	repo, err := git.PlainOpen(".")
	groomba.CheckIfError(err, "failed to open repository")

	a, err := auth.NewAuth(cfg.Auth)
	groomba.CheckIfError(err, "failed to initialize auth")

	err = repo.Fetch(&git.FetchOptions{
		RemoteName: "origin",
		//RefSpecs:   []config.RefSpec{"refs/remotes/origin"},
		RefSpecs: []config.RefSpec{"+refs/heads/*:refs/remotes/origin/*"},
		Depth:    1,
		Auth:     a.Get(),
	})
	groomba.CheckIfError(err, "failed to fetch references from upstream")

	g := groomba.NewGroomba(cfg, repo, a)

	fb, err := g.FilterBranches(time.Now())
	groomba.CheckIfError(err, "failed to filter stale branches")

	err = g.PrintBranchesGroupbyAuthor(fb)
	groomba.CheckIfError(err, "failed to print branches by author")

	err = g.MoveStaleBranches(fb)
	groomba.CheckIfError(err, "failed to move stale branches")
}
