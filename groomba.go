package main

import (
	"fmt"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"

	//"github.com/go-git/go-git/v5/storage/memory"
	"log"
)

func main() {
	fmt.Println("Hello World!")
	r, _ := git.PlainOpen(".")
	// rem, _ := r.Remote("origin")
	//listBranches, err := rem.List()
	listBranches, err := r.Branches()
	if err != nil {
		log.Fatal(err)
	}
	listBranches.ForEach(func(ref *plumbing.Reference) error {
		fmt.Println(ref)
		return nil
	})
	// for _, branch := range listBranches {
	// 	fmt.Println(branch)
	// }
}
