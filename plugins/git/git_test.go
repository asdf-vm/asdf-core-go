package git

import (
	"asdf/plugins/plugintest"
	"os"
	"path/filepath"
	"testing"

	"github.com/go-git/go-git/v5"
	"github.com/stretchr/testify/assert"
)

// TODO: Switch to local repo so tests don't go over the network
const (
	testRepo       = "https://github.com/Stratus3D/asdf-lua"
	testPluginName = "lua"
)

func TestPluginDefaultBranch(t *testing.T) {
	testRepoPath, err := plugintest.InstallMockPluginRepo(t.TempDir(), testPluginName)
	assert.Nil(t, err)

	repo, err := git.PlainOpen(testRepoPath)
	assert.Nil(t, err)

	t.Run("returns default branch when remote named 'origin' exists", func(t *testing.T) {
		defaultBranch, err := pluginDefaultBranch(repo)
		assert.Nil(t, err)
		assert.Equal(t, "master", defaultBranch)
	})

	t.Run("returns error when no remote named 'origin' exists", func(t *testing.T) {
		err := repo.DeleteRemote("origin")
		assert.Nil(t, err)

		defaultBranch, err := pluginDefaultBranch(repo)
		assert.ErrorContains(t, err, "remote not found")
		assert.Equal(t, "", defaultBranch)
	})
}

func TestGitPluginClone(t *testing.T) {
	t.Run("when plugin name is valid but URL is invalid prints an error", func(t *testing.T) {
		tempDir := t.TempDir()
		directory := filepath.Join(tempDir, testPluginName)

		plugin := NewGitPlugin(directory)
		err := plugin.Clone("foobar")

		assert.ErrorContains(t, err, "unable to clone plugin: repository not found")
	})

	t.Run("clones provided Git URL to plugin directory when URL is valid", func(t *testing.T) {
		tempDir := t.TempDir()
		directory := filepath.Join(tempDir, testPluginName)

		plugin := NewGitPlugin(directory)
		err := plugin.Clone(testRepo)

		assert.Nil(t, err)

		// Assert plugin directory contains Git repo with bin directory
		_, err = os.ReadDir(directory + "/.git")
		assert.Nil(t, err)

		entries, err := os.ReadDir(directory + "/bin")
		assert.Nil(t, err)
		assert.Equal(t, 5, len(entries))
	})
}

func TestGitPluginHead(t *testing.T) {
	tempDir := t.TempDir()
	directory := filepath.Join(tempDir, testPluginName)

	plugin := NewGitPlugin(directory)

	err := plugin.Clone(testRepo)
	assert.Nil(t, err)

	head, err := plugin.Head()
	assert.Nil(t, err)
	assert.NotZero(t, head)
}

func TestGitPluginRemoteURL(t *testing.T) {
	tempDir := t.TempDir()
	directory := filepath.Join(tempDir, testPluginName)

	plugin := NewGitPlugin(directory)

	err := plugin.Clone(testRepo)
	assert.Nil(t, err)

	url, err := plugin.RemoteURL()
	assert.Nil(t, err)
	assert.NotZero(t, url)
}
