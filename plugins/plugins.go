package plugins

import (
	"asdf/config"
	gitPlugin "asdf/plugins/git"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
)

const dataDirPlugins = "plugins"
const invalidPluginNameMsg = "'%q' is invalid. Name may only contain lowercase letters, numbers, '_', and '-'"
const pluginAlreadyExists = "plugin named %q already added"

type Plugin struct {
	Name string
	Dir  string
	Ref  string
	URL  string
}

func List(config config.Config, urls, refs bool) (plugins []Plugin, err error) {
	pluginsDir := DataDirectory(config.DataDir)
	files, err := os.ReadDir(pluginsDir)
	if err != nil {
		return plugins, err
	}

	for _, file := range files {
		if file.IsDir() {
			if refs || urls {
				var url string
				var refString string
				location := filepath.Join(pluginsDir, file.Name())
				plugin := gitPlugin.NewGitPlugin(location)

				// TODO: Improve these error messages
				if err != nil {
					return plugins, err
				}

				if refs {
					refString, err = plugin.Head()

					if err != nil {
						return plugins, err
					}
				}

				if urls {
					url, err = plugin.RemoteURL()

					if err != nil {
						return plugins, err
					}
				}

				plugins = append(plugins, Plugin{
					Name: file.Name(),
					Dir:  location,
					URL:  url,
					Ref:  refString,
				})
			} else {
				plugins = append(plugins, Plugin{
					Name: file.Name(),
					Dir:  filepath.Join(pluginsDir, file.Name()),
				})
			}
		}
	}

	return plugins, nil
}

func Add(config config.Config, pluginName, pluginURL string) error {
	err := validatePluginName(pluginName)

	if err != nil {
		return err
	}

	exists, err := PluginExists(config.DataDir, pluginName)

	if err != nil {
		return fmt.Errorf("unable to check if plugin already exists: %w", err)
	}

	if exists {
		return fmt.Errorf(pluginAlreadyExists, pluginName)
	}

	pluginDir := PluginDirectory(config.DataDir, pluginName)

	if err != nil {
		return fmt.Errorf("unable to create plugin directory: %w", err)
	}

	return gitPlugin.NewGitPlugin(pluginDir).Clone(pluginURL)
}

func Remove(config config.Config, pluginName string) error {
	err := validatePluginName(pluginName)

	if err != nil {
		return err
	}

	exists, err := PluginExists(config.DataDir, pluginName)

	if err != nil {
		return fmt.Errorf("unable to check if plugin exists: %w", err)
	}

	if !exists {
		return fmt.Errorf("no such plugin: %s", pluginName)
	}

	pluginDir := PluginDirectory(config.DataDir, pluginName)

	return os.RemoveAll(pluginDir)
}

func Update(config config.Config, pluginName, ref string) (string, error) {
	exists, err := PluginExists(config.DataDir, pluginName)

	if err != nil {
		return "", fmt.Errorf("unable to check if plugin exists: %w", err)
	}

	if !exists {
		return "", fmt.Errorf("no such plugin: %s", pluginName)
	}

	pluginDir := PluginDirectory(config.DataDir, pluginName)

	plugin := gitPlugin.NewGitPlugin(pluginDir)

	return plugin.Update(ref)
}

func PluginExists(dataDir, pluginName string) (bool, error) {
	pluginDir := PluginDirectory(dataDir, pluginName)
	return directoryExists(pluginDir)
}

func directoryExists(dir string) (bool, error) {
	fileInfo, err := os.Stat(dir)
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}

	if err != nil {
		return false, err
	}

	return fileInfo.IsDir(), nil
}

func PluginDirectory(dataDir, pluginName string) string {
	return filepath.Join(DataDirectory(dataDir), pluginName)
}

func DataDirectory(dataDir string) string {
	return filepath.Join(dataDir, dataDirPlugins)
}

func validatePluginName(name string) error {
	match, err := regexp.MatchString("^[[:lower:][:digit:]_-]+$", name)
	if err != nil {
		return err
	}

	if !match {
		return fmt.Errorf(invalidPluginNameMsg, name)
	}

	return nil
}
