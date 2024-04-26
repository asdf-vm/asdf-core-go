package plugins

import (
	"asdf/config"
	"asdf/plugins/plugintest"
	"os"
	"strings"
	"testing"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/stretchr/testify/assert"
)

// TODO: Switch to local repo so tests don't go over the network
const (
	testRepo       = "https://github.com/Stratus3D/asdf-lua"
	testPluginName = "lua"
)

func TestList(t *testing.T) {
	testDataDir := t.TempDir()
	conf := config.Config{DataDir: testDataDir}
	testRepo, err := plugintest.InstallMockPluginRepo(testDataDir, testPluginName)
	assert.Nil(t, err)

	err = Add(conf, testPluginName, testRepo)
	assert.Nil(t, err)

	t.Run("when urls and refs are set to false returns plugin names", func(t *testing.T) {
		plugins, err := List(conf, false, false)
		assert.Nil(t, err)

		plugin := plugins[0]
		assert.Equal(t, "lua", plugin.Name)
		assert.NotZero(t, plugin.Dir)
		assert.Zero(t, plugin.URL)
		assert.Zero(t, plugin.Ref)
	})

	t.Run("when urls is set to true returns plugins with repo urls set", func(t *testing.T) {
		plugins, err := List(conf, true, false)
		assert.Nil(t, err)

		plugin := plugins[0]
		assert.Equal(t, "lua", plugin.Name)
		assert.NotZero(t, plugin.Dir)
		assert.Zero(t, plugin.Ref)
		assert.NotZero(t, plugin.URL)
	})

	t.Run("when refs is set to true returns plugins with current repo refs set", func(t *testing.T) {
		plugins, err := List(conf, false, true)
		assert.Nil(t, err)

		plugin := plugins[0]
		assert.Equal(t, "lua", plugin.Name)
		assert.NotZero(t, plugin.Dir)
		assert.NotZero(t, plugin.Ref)
		assert.Zero(t, plugin.URL)
	})

	t.Run("when refs and urls are both set to true returns plugins with both set", func(t *testing.T) {
		plugins, err := List(conf, true, true)
		assert.Nil(t, err)

		plugin := plugins[0]
		assert.Equal(t, "lua", plugin.Name)
		assert.NotZero(t, plugin.Dir)
		assert.NotZero(t, plugin.Ref)
		assert.NotZero(t, plugin.URL)
	})
}

func TestAdd(t *testing.T) {
	testDataDir := t.TempDir()

	t.Run("when given an invalid plugin name prints an error", func(t *testing.T) {
		var invalids = []string{"plugin^name", "plugin%name", "plugin name", "PLUGIN_NAME"}

		for _, invalid := range invalids {
			t.Run(invalid, func(t *testing.T) {
				err := Add(config.Config{}, invalid, testRepo)

				expectedErrMsg := "is invalid. Name may only contain lowercase letters, numbers, '_', and '-'"
				if !strings.Contains(err.Error(), expectedErrMsg) {
					t.Errorf("Expected an error with message %v", expectedErrMsg)
				}
			})
		}
	})

	t.Run("when plugin with same name already exists prints an error", func(t *testing.T) {
		conf := config.Config{DataDir: testDataDir}

		// Add plugin
		err := Add(conf, testPluginName, testRepo)

		if err != nil {
			t.Fatal("Expected to be able to add plugin")
		}

		// Add it again to trigger error
		err = Add(conf, testPluginName, testRepo)

		if err == nil {
			t.Fatal("expected error got nil")
		}

		expectedErrMsg := "plugin named \"lua\" already added"
		if !strings.Contains(err.Error(), expectedErrMsg) {
			t.Errorf("Expected an error with message %v", expectedErrMsg)
		}
	})

	t.Run("when plugin name is valid but URL is invalid prints an error", func(t *testing.T) {
		conf := config.Config{DataDir: testDataDir}

		err := Add(conf, "foo", "foobar")

		assert.ErrorContains(t, err, "unable to clone plugin: repository not found")
	})

	t.Run("when plugin name and URL are valid installs plugin", func(t *testing.T) {
		testDataDir := t.TempDir()
		conf := config.Config{DataDir: testDataDir}

		err := Add(conf, testPluginName, testRepo)

		assert.Nil(t, err, "Expected to be able to add plugin")

		// Assert plugin directory contains Git repo with bin directory
		pluginDir := PluginDirectory(testDataDir, testPluginName)

		_, err = os.ReadDir(pluginDir + "/.git")
		assert.Nil(t, err)

		entries, err := os.ReadDir(pluginDir + "/bin")
		assert.Nil(t, err)
		assert.Equal(t, 5, len(entries))
	})
}

func TestRemove(t *testing.T) {
	testDataDir := t.TempDir()
	conf := config.Config{DataDir: testDataDir}

	err := Add(conf, testPluginName, testRepo)
	assert.Nil(t, err)

	t.Run("returns error when plugin with name does not exist", func(t *testing.T) {
		err := Remove(conf, "nonexistant")
		assert.NotNil(t, err)
		assert.ErrorContains(t, err, "no such plugin")
	})

	t.Run("returns error when invalid plugin name is given", func(t *testing.T) {
		err := Remove(conf, "foo/bar/baz")
		assert.NotNil(t, err)
		expectedErrMsg := "is invalid. Name may only contain lowercase letters, numbers, '_', and '-'"
		assert.ErrorContains(t, err, expectedErrMsg)
	})

	t.Run("removes plugin when passed name of installed plugin", func(t *testing.T) {
		err := Remove(conf, testPluginName)
		assert.Nil(t, err)

		pluginDir := PluginDirectory(testDataDir, testPluginName)
		_, err = os.Stat(pluginDir)
		assert.NotNil(t, err)
		assert.True(t, os.IsNotExist(err))
	})
}

func TestUpdate(t *testing.T) {
	testDataDir := t.TempDir()
	conf := config.Config{DataDir: testDataDir}

	err := Add(conf, testPluginName, testRepo)
	assert.Nil(t, err)

	t.Run("returns error when plugin with name does not exist", func(t *testing.T) {
		err := Update(conf, "nonexistant", "")

		assert.NotNil(t, err)
		expectedErrMsg := "no such plugin: nonexistant"
		assert.ErrorContains(t, err, expectedErrMsg)
	})

	t.Run("returns error when plugin repo does not exist", func(t *testing.T) {
		badPluginName := "badplugin"
		badRepo := PluginDirectory(testDataDir, badPluginName)
		os.MkdirAll(badRepo, 0777)

		err := Update(conf, badPluginName, "")

		assert.NotNil(t, err)
		expectedErrMsg := "unable to open plugin: repository does not exist"
		assert.ErrorContains(t, err, expectedErrMsg)
	})

	t.Run("updates plugin when plugin with name exists", func(t *testing.T) {
		err := Update(conf, testPluginName, "")
		assert.Nil(t, err)

		// TODO: Check that plugin was updated to latest
	})

	t.Run("does not return error when plugin is already updated", func(t *testing.T) {
		// update plugin twice to test already updated case
		err := Update(conf, testPluginName, "")
		assert.Nil(t, err)
		err = Update(conf, testPluginName, "")
		assert.Nil(t, err)
	})

	t.Run("updates plugin to ref when plugin with name and ref exist", func(t *testing.T) {
		ref := "foobar"
		err := Update(conf, testPluginName, ref)
		assert.Nil(t, err)

		// TODO: Check that plugin was updated to ref
	})
}

func TestPluginDefaultBranch(t *testing.T) {
	testRepoPath, err := installMockPluginRepo(t.TempDir(), testPluginName)
	assert.Nil(t, err)

	repo, err := git.PlainOpen(testRepoPath)
	assert.Nil(t, err)

	t.Run("returns default branch when remote named 'origin' exists", func(t *testing.T) {
		defaultBranch, err := PluginDefaultBranch(repo)
		assert.Nil(t, err)
		assert.Equal(t, "master", defaultBranch)
	})

	t.Run("returns error when no remote named 'origin' exists", func(t *testing.T) {
		err := repo.DeleteRemote("origin")
		assert.Nil(t, err)

		defaultBranch, err := PluginDefaultBranch(repo)
		assert.ErrorContains(t, err, "remote not found")
		assert.Equal(t, "", defaultBranch)
	})
}

func TestPluginExists(t *testing.T) {
	testDataDir := t.TempDir()
	pluginDir := PluginDirectory(testDataDir, testPluginName)
	err := os.MkdirAll(pluginDir, 0777)
	if err != nil {
		t.Errorf("got %v, expected nil", err)
	}

	t.Run("returns true when plugin exists", func(t *testing.T) {
		exists, err := PluginExists(testDataDir, testPluginName)

		if err != nil {
			t.Errorf("got %v, expected nil", err)
		}

		if exists != true {
			t.Error("got false, expected true")
		}
	})

	t.Run("returns false when plugin path is file and not dir", func(t *testing.T) {
		pluginName := "file"
		pluginDir := PluginDirectory(testDataDir, pluginName)
		err := touchFile(pluginDir)
		if err != nil {
			t.Errorf("got %v, expected nil", err)
		}

		exists, err := PluginExists(testDataDir, pluginName)

		if err != nil {
			t.Errorf("got %v, expected nil", err)
		}

		if exists != false {
			t.Error("got false, expected true")
		}
	})

	t.Run("returns false when plugin dir does not exist", func(t *testing.T) {
		exists, err := PluginExists(testDataDir, "non-existant")

		if err != nil {
			t.Errorf("got %v, expected nil", err)
		}

		if exists != false {
			t.Error("got false, expected true")
		}
	})
}

func TestPluginDirectory(t *testing.T) {
	t.Run("returns new path with plugin name as last segment", func(t *testing.T) {
		pluginDir := PluginDirectory("~/.asdf/", testPluginName)
		expected := "~/.asdf/plugins/lua"
		if pluginDir != expected {
			t.Errorf("got %v, expected %v", pluginDir, expected)
		}
	})
}

func TestValidatePluginName(t *testing.T) {
	t.Run("returns no error when plugin name is valid", func(t *testing.T) {
		err := validatePluginName(testPluginName)
		refuteError(t, err)
	})

	var invalids = []string{"plugin^name", "plugin%name", "plugin name", "PLUGIN_NAME"}

	for _, invalid := range invalids {
		t.Run(invalid, func(t *testing.T) {
			err := validatePluginName(invalid)

			if err == nil {
				t.Error("Expected an error")
			}
		})
	}
}

func refuteError(t *testing.T, err error) {
	if err != nil {
		t.Fatal("Returned unexpected error", err)
	}
}

func touchFile(name string) error {
	file, err := os.OpenFile(name, os.O_RDONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	return file.Close()
}
