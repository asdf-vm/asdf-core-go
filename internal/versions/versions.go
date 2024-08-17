// Package versions handles all operations pertaining to specific versions.
// Install, uninstall, etc...
package versions

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"asdf/config"
	"asdf/hook"
	"asdf/plugins"
)

const (
	systemVersion           = "system"
	latestVersion           = "latest"
	uninstallableVersionMsg = "uninstallable version: system"
	dataDirDownloads        = "downloads"
	dataDirInstalls         = "installs"
)

// UninstallableVersion is an error returned if someone tries to install the
// system version.
type UninstallableVersion struct{}

func (e UninstallableVersion) Error() string {
	return fmt.Sprint(uninstallableVersionMsg)
}

// TODO: Implement these functions
//func InstallAll() {
//}

//func InstallOne() {
//}

// InstallOneVersion installs a specific version of a specific tool
func InstallOneVersion(conf config.Config, plugin plugins.Plugin, version string, _ bool, stdOut io.Writer, stdErr io.Writer) error {
	err := plugin.Exists()
	if err != nil {
		return err
	}

	if version == systemVersion {
		return UninstallableVersion{}
	}

	if version == latestVersion {
		// TODO: Implement this
		return errors.New("not implemented")
	}

	downloadDir := downloadPath(conf, plugin, version)
	installDir := installPath(conf, plugin, version)
	versionType, version := ParseString(version)

	// Check if version already installed
	if _, err = os.Stat(installDir); !os.IsNotExist(err) {
		return fmt.Errorf("version %s of %s is already installed", version, plugin.Name)
	}

	env := map[string]string{
		"ASDF_INSTALL_TYPE":    versionType,
		"ASDF_INSTALL_VERSION": version,
		"ASDF_INSTALL_PATH":    installDir,
		"ASDF_DOWNLOAD_PATH":   downloadDir,
	}

	err = os.MkdirAll(downloadDir, 0o777)
	if err != nil {
		return fmt.Errorf("unable to create download dir: %w", err)
	}

	err = hook.RunWithOutput(conf, fmt.Sprintf("pre_asdf_download_%s", plugin.Name), []string{version}, stdOut, stdErr)
	if err != nil {
		return fmt.Errorf("failed to run pre-download hook: %w", err)
	}

	err = plugin.RunCallback("download", []string{}, env, stdOut, stdErr)
	if _, ok := err.(plugins.NoCallbackError); err != nil && !ok {
		return fmt.Errorf("failed to run download callback: %w", err)
	}

	err = hook.RunWithOutput(conf, fmt.Sprintf("pre_asdf_install_%s", plugin.Name), []string{version}, stdOut, stdErr)
	if err != nil {
		return fmt.Errorf("failed to run pre-install hook: %w", err)
	}

	err = os.MkdirAll(installDir, 0o777)
	if err != nil {
		return fmt.Errorf("unable to create install dir: %w", err)
	}

	err = plugin.RunCallback("install", []string{}, env, stdOut, stdErr)
	if err != nil {
		return fmt.Errorf("failed to run install callback: %w", err)
	}

	err = hook.RunWithOutput(conf, fmt.Sprintf("post_asdf_install_%s", plugin.Name), []string{version}, stdOut, stdErr)
	if err != nil {
		return fmt.Errorf("failed to run post-install hook: %w", err)
	}
	return nil
}

func Latest(plugin plugins.Plugin, query string) (versions []string, err error) {

	if query == "" {
		query = defaultQuery
	}

	var stdOut strings.Builder
	var stdErr strings.Builder

	err = plugin.RunCallback("latest-stable", []string{query}, map[string]string{}, &stdOut, &stdErr)
	if err != nil {
		if _, ok := err.(plugins.NoCallbackError); !ok {
			return versions, err
		}

		// compute latest version manually
		// get all versi
		//versions=$(list_all_command "$plugin_name" "$query" |
		//  grep -ivE "(^Available versions:|-src|-dev|-latest|-stm|[-\\.]rc|-milestone|-alpha|-beta|[-\\.]pre|-next|(a|b|c)[0-9]+|snapshot|master)" |
		//  sed 's/^[[:space:]]\+//' |
		//  tail -1)
		return versions, nil
	}

	// parse stdOut and return version
	versions = strings.Split(stdOut.String(), " ")
	return versions, nil
}

// ListAll returns a slice of all available versions for the tool managed by
// the given plugin by invoking the plugin's list-all callback
func ListAll(plugin plugins.Plugin) (versions []string, err error) {
	var stdout strings.Builder
	var stderr strings.Builder

	err = plugin.RunCallback("list-all", []string{}, map[string]string{}, &stdout, &stderr)
	if err != nil {
		return versions, err
	}

	versions = parseVersions(stdout.String())

	return versions, err
}

// future refactoring opportunity: this function is an exact copy of
// resolve.parseVersion
func parseVersions(rawVersions string) []string {
	var versions []string
	for _, version := range strings.Split(rawVersions, " ") {
		version = strings.TrimSpace(version)
		if len(version) > 0 {
			versions = append(versions, version)
		}
	}
	return versions
}

// ParseString parses a version string into versionType and version components
func ParseString(version string) (string, string) {
	segments := strings.Split(version, ":")
	if len(segments) >= 1 && segments[0] == "ref" {
		return "ref", strings.Join(segments[1:], ":")
	}

	return "version", version
}

func downloadPath(conf config.Config, plugin plugins.Plugin, version string) string {
	return filepath.Join(conf.DataDir, dataDirDownloads, plugin.Name, version)
}

func installPath(conf config.Config, plugin plugins.Plugin, version string) string {
	return filepath.Join(conf.DataDir, dataDirInstalls, plugin.Name, version)
}
