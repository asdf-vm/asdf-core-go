// Package toolversions handles reading and writing tools and versions from
// asdf's .tool-versions files. It also handles parsing version strings from
// .tool-versions files and command line arguments.
package toolversions

import (
	"fmt"
	"os"
	"slices"
	"strings"
)

// ToolVersions represents a tool along with versions specified for it
type ToolVersions struct {
	Name     string
	Versions []string
}

// FindToolVersions looks up a tool version in a tool versions file and if found
// returns a slice of versions for it.
func FindToolVersions(filepath, toolName string) (versions []string, found bool, err error) {
	content, err := os.ReadFile(filepath)
	if err != nil {
		return versions, false, err
	}

	versions, found = findToolVersionsInContent(string(content), toolName)
	return versions, found, nil
}

// GetAllToolsAndVersions returns a list of all tools and associated versions
// contained in a .tool-versions file
func GetAllToolsAndVersions(filepath string) (toolVersions []ToolVersions, err error) {
	content, err := os.ReadFile(filepath)
	if err != nil {
		return toolVersions, err
	}

	toolVersions = getAllToolsAndVersionsInContent(string(content))
	return toolVersions, nil
}

// Intersect takes two slices of versions and returns a new slice containing
// only the versions found in both.
func Intersect(versions1 []string, versions2 []string) (versions []string) {
	for _, version1 := range versions1 {
		for _, version2 := range versions2 {
			if version2 == version1 {
				versions = append(versions, version1)
			}
		}
	}
	return versions
}

// Unique takes a slice of ToolVersions and returns a slice of unique tools and
// versions.
func Unique(versions []ToolVersions) (uniques []ToolVersions) {
	for _, version := range versions {
		var found bool

		for index, unique := range uniques {
			if unique.Name == version.Name {
				// Duplicate name, check versions
				for _, versionNumber := range version.Versions {
					if !slices.Contains(unique.Versions, versionNumber) {
						unique.Versions = append(unique.Versions, versionNumber)
					}
				}

				uniques[index] = unique
				found = true
				break
			}
		}

		// None with name found, add
		if !found {
			uniques = append(uniques, version)
		}
	}

	return uniques
}

// ParseFromCliArg parses a string that is passed in as an argument to one of
// the asdf subcommands. Some subcommands allow the special version `latest` to
// be used, with an optional filter string.
func ParseFromCliArg(version string) (string, string) {
	segments := strings.Split(version, ":")
	if len(segments) > 0 && segments[0] == "latest" {
		if len(segments) > 1 {
			// Must be latest with filter
			return "latest", segments[1]
		}
		return "latest", ""
	}

	return Parse(version)
}

// Parse parses a version string into versionType and version components
func Parse(version string) (string, string) {
	segments := strings.Split(version, ":")
	if len(segments) >= 1 {
		remainder := strings.Join(segments[1:], ":")
		switch segments[0] {
		case "ref":
			return "ref", remainder
		case "path":
			// This is for people who have the local source already compiled
			// Like those who work on the language, etc
			// We'll allow specifying path:/foo/bar/project in .tool-versions
			// And then use the binaries there
			return "path", remainder
		default:
			return "version", version
		}
	}

	return "version", version
}

// FormatForFS takes a versionType and version strings and generate a version
// string suitable for the file system
func FormatForFS(versionType, version string) string {
	switch versionType {
	case "ref":
		return fmt.Sprintf("ref-%s", version)
	default:
		return version
	}
}

// readLines reads all the lines in a given file
// removing spaces and comments which are marked by '#'
func readLines(content string) (lines []string) {
	for _, line := range strings.Split(content, "\n") {
		line = strings.SplitN(line, "#", 2)[0]
		line = strings.TrimSpace(line)
		if len(line) > 0 {
			lines = append(lines, line)
		}
	}
	return
}

func findToolVersionsInContent(content, toolName string) (versions []string, found bool) {
	toolVersions := getAllToolsAndVersionsInContent(content)
	for _, tool := range toolVersions {
		if tool.Name == toolName {
			return tool.Versions, true
		}
	}

	return versions, found
}

func getAllToolsAndVersionsInContent(content string) (toolVersions []ToolVersions) {
	for _, line := range readLines(content) {
		tokens := parseLine(line)
		newTool := ToolVersions{Name: tokens[0], Versions: tokens[1:]}
		toolVersions = append(toolVersions, newTool)
	}

	return toolVersions
}

func parseLine(line string) (tokens []string) {
	for _, token := range strings.Split(line, " ") {
		token = strings.TrimSpace(token)
		if len(token) > 0 {
			tokens = append(tokens, token)
		}
	}

	return tokens
}
