package metrics

import (
	"fmt"
	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-git/plumbing"
	"github.com/go-git/go-git/storage/memory"
	"github.com/go-git/go-git"
	"os"
	//"time"
)

func Checkout(repoUrl, hash string) *git.Repository {
	//PrintInBlue("git clone " + repoUrl)

	r := GetRepo(repoUrl)
	w, err := r.Worktree()
	CheckIfError(err)

	// ... checking out to commit
	//PrintInBlue("git checkout %s", hash)
	err = w.Checkout(&git.CheckoutOptions{
		Hash: plumbing.NewHash(hash),
	})
	CheckIfError(err)
	return r
}

func GetRepo(repoUrl string) *git.Repository {
	//defer helper.Duration(helper.Track("GetRepo"))

	//PrintInBlue("git clone " + repoUrl)

	var r *git.Repository
	var err error
	//if strings.HasPrefix(repoUrl, "https://github.com") {
	r, err = git.Clone(memory.NewStorage(), memfs.New(), &git.CloneOptions{
		URL: repoUrl,
	})
	CheckIfError(err)
	//} else {
	//	r, err = git.PlainOpen(repoUrl)
	//	CheckIfError(err)
	//}
	return r
}

// CheckIfError should be used to naively panics if an error is not nil.
func CheckIfError(err error) {
	if err == nil {
		return
	}

	fmt.Printf("\x1b[31;1m%s\x1b[0m\n", fmt.Sprintf("error: %s", err))
	os.Exit(1)
}
