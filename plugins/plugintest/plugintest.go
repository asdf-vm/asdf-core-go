package plugintest

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func InstallMockPluginRepo(dataDir, name string) (string, error) {
	// Because the legacy dummy plugin directory is relative to the root of this
	// project I cannot use the usual testing functions to locate it. To
	// determine the location of it we compute the module root, which also
	// happens to be the root of the repo.
	modRootDir, err := moduleRoot()
	if err != nil {
		return "", err
	}

	location := dataDir + "/repo-" + name

	// Then we specify the path to the dummy plugin relative to the module root
	err = runCmd("cp", "-r", filepath.Join(modRootDir, "test/fixtures/dummy_plugin"), location)
	if err != nil {
		return location, err
	}

	// Definitely some opportunities to refactor here. This code might be
	// simplified by switching to the Go git library
	err = runCmd("git", "-C", location, "init", "-q")
	if err != nil {
		return location, err
	}

	err = runCmd("git", "-C", location, "config", "user.name", "\"Test\"")
	if err != nil {
		return location, err
	}

	err = runCmd("git", "-C", location, "config", "user.email", "\"test@example.com\"")
	if err != nil {
		return location, err
	}

	err = runCmd("git", "-C", location, "add", "-A")
	if err != nil {
		return location, err
	}

	err = runCmd("git", "-C", location, "commit", "-q", "-m", fmt.Sprintf("\"asdf %s plugin init\"", name))
	if err != nil {
		return location, err
	}

	err = runCmd("touch", filepath.Join(location, "README.md"))
	if err != nil {
		return location, err
	}

	err = runCmd("git", "-C", location, "add", "-A")
	if err != nil {
		return location, err
	}

	err = runCmd("git", "-C", location, "commit", "-q", "-m", fmt.Sprintf("\"asdf %s plugin readme \"", name))
	if err != nil {
		return location, err
	}

	// kind of ugly but I want a remote with a valid path so I use the same
	// location as the remote. Probably should refactor
	err = runCmd("git", "-C", location, "remote", "add", "origin", location)
	if err != nil {
		return location, err
	}

	return location, err
}

func moduleRoot() (string, error) {
	currentDir, err := os.Getwd()

	if err != nil {
		return "", err
	}

	return findModuleRoot(currentDir), nil
}

// Taken from https://github.com/golang/go/blob/9e3b1d53a012e98cfd02de2de8b1bd53522464d4/src/cmd/go/internal/modload/init.go#L1504C1-L1522C2 because that function is in an internal module
// and I can't rely on it.
func findModuleRoot(dir string) (roots string) {
	if dir == "" {
		panic("dir not set")
	}
	dir = filepath.Clean(dir)

	// Look for enclosing go.mod.
	for {
		if fi, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil && !fi.IsDir() {
			return dir
		}
		d := filepath.Dir(dir)
		if d == dir {
			break
		}
		dir = d
	}
	return ""
}

// helper function to make running commands easier
func runCmd(cmdName string, args ...string) error {
	cmd := exec.Command(cmdName, args...)

	// Capture stdout and stderr
	var stdout strings.Builder
	var stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	if err != nil {
		// If command fails print both stderr and stdout
		fmt.Println("stdout:", stdout.String())
		fmt.Println("stderr:", stderr.String())
		return err
	}

	return nil
}