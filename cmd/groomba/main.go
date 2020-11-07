package main

import (
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"

	"github.com/avbm/groomba"
)

func main() {
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
