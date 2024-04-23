package git

import (
	"fmt"
	"strings"

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
		ref, err = pluginDefaultBranch(repo)

		if err != nil {
			return "", err
		}

		checkoutOptions = git.CheckoutOptions{Branch: plumbing.NewBranchReferenceName(ref), Force: true}
	} else {
		checkoutOptions = git.CheckoutOptions{Hash: plumbing.NewHash(ref), Force: true}
	}

	err = repo.Fetch(&git.FetchOptions{RemoteName: remoteName, Force: true, RefSpecs: []gitConfig.RefSpec{
		gitConfig.RefSpec(ref + ":" + ref),
	}})

	if err != nil {
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

func pluginDefaultBranch(repo *git.Repository) (ref string, err error) {
	remote, err := repo.Remote(remoteName)
	if err != nil {
		return ref, err
	}

	refs, err := remote.List(&git.ListOptions{})
	if err != nil {
		return ref, err
	}

	for _, r := range refs {
		if r.Name().IsBranch() {
			segments := strings.Split(r.Name().String(), "/")
			ref = segments[len(segments)-1]
		}
	}

	return ref, err
}
