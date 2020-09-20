package main

import (
	"fmt"
	"os"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	//"github.com/go-git/go-git/v5/storage/memory"
)

// CheckIfError should be used to naively panics if an error is not nil.
func CheckIfError(err error) {
	if err == nil {
		return
	}

	fmt.Printf("\x1b[31;1m%s\x1b[0m\n", fmt.Sprintf("error: %s", err))
	os.Exit(1)
}

func main() {
	repo, _ := git.PlainOpen(".")
	// rem, _ := r.Remote("origin")
	//listBranches, err := rem.List()
	listBranches, err := repo.Branches()
	CheckIfError(err)
	listBranches.ForEach(func(ref *plumbing.Reference) error {
		// fmt.Println(ref)
		commit, err := repo.CommitObject(ref.Hash())
		CheckIfError(err)
		fmt.Println(ref.Name().String(), commit.Author.Name, commit.Author.When)
		return nil
	})
	// for _, branch := range listBranches {
	// 	fmt.Println(branch)
	// }
}
