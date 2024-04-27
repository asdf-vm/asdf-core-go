package git

import (
	"fmt"

	"github.com/go-git/go-git/v5"
	gitConfig "github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
)

const remoteName = "origin"

type gitPlugin struct {
	directory string
}

type PluginOps interface {
	Clone(pluginURL string) error
	Head() (string, error)
	RemoteURL() (string, error)
	Update(ref string) (string, error)
}

func NewGitPlugin(directory string) gitPlugin {
	return gitPlugin{directory: directory}
}

func (g gitPlugin) Clone(pluginURL string) error {
	_, err := git.PlainClone(g.directory, false, &git.CloneOptions{
		URL: pluginURL,
	})

	if err != nil {
		return fmt.Errorf("unable to clone plugin: %w", err)
	}

	return nil
}

func (g gitPlugin) Head() (string, error) {
	repo, err := git.PlainOpen(g.directory)

	// TODO: Improve these error messages
	if err != nil {
		return "", err
	}

	ref, err := repo.Head()
	if err != nil {
		return "", err
	}

	return ref.Hash().String(), nil
}

func (g gitPlugin) RemoteURL() (string, error) {
	repo, err := git.PlainOpen(g.directory)

	// TODO: Improve these error messages
	if err != nil {
		return "", err
	}

	remotes, err := repo.Remotes()
	if err != nil {
		return "", err
	}

	return remotes[0].Config().URLs[0], nil
}

func (g gitPlugin) Update(ref string) (string, error) {
	repo, err := git.PlainOpen(g.directory)

	if err != nil {
		return "", fmt.Errorf("unable to open plugin: %w", err)
	}

	var checkoutOptions git.CheckoutOptions

	if ref == "" {
		// If no ref is provided checkout latest commit on current branch
		head, err := repo.Head()

		if err != nil {
			return "", err
		}

		if head.Name().IsBranch() {
			// If on a branch checkout the latest version of it from the remote
			currentBranch := head.Name()
			ref = currentBranch.String()
			checkoutOptions = git.CheckoutOptions{Branch: currentBranch, Force: true}
		} else {
			return "", fmt.Errorf("not on a branch, unable to update")
		}
	} else {
		// Checkout ref if provided
		checkoutOptions = git.CheckoutOptions{Hash: plumbing.NewHash(ref), Force: true}
	}
	err = repo.Fetch(&git.FetchOptions{RemoteName: remoteName, Force: true, RefSpecs: []gitConfig.RefSpec{
		gitConfig.RefSpec(ref + ":" + ref),
	}})

	if err != nil && err != git.NoErrAlreadyUpToDate {
		return "", err
	}

	worktree, err := repo.Worktree()
	if err != nil {
		return "", err
	}

	err = worktree.Checkout(&checkoutOptions)
	if err != nil {
		return "", err
	}

	hash, err := repo.ResolveRevision(plumbing.Revision("HEAD"))
	return hash.String(), err
}
