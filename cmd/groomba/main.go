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
)

func main() {
	log.SetHandler(cli.Default)
	cfg, err := groomba.GetConfig(".")
	groomba.CheckIfError(err)

	repo, _ := git.PlainOpen(".")
	repo.Fetch(&git.FetchOptions{
		RemoteName: "origin",
		RefSpecs:   []config.RefSpec{"refs/remotes/origin"},
		Depth:      1,
	})

	g := groomba.NewGroomba(cfg, repo)

	fb, err := g.FilterBranches(time.Now())
	groomba.CheckIfError(err)

	err = g.PrintBranchesGroupbyAuthor(fb)
	groomba.CheckIfError(err)

	err = g.MoveStaleBranches(fb)
	groomba.CheckIfError(err)
}
