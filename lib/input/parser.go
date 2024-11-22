package input

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
	"sync"

	utils "github.com/DerekCorniello/pip-req-valid/utils"
)

/*
requirements.txt example:

# This is a comment, to show how #-prefixed lines are ignored.
# It is possible to specify requirements as plain names.
pytest
pytest-cov
beautifulsoup4

# The syntax supported here is the same as that of requirement specifiers.
docopt == 0.6.1
requests [security] >= 2.8.1, == 2.8.* ; python_version < "2.7"
urllib3 @ https://github.com/urllib3/urllib3/archive/refs/tags/1.26.8.zip

# It is possible to refer to other requirement files or constraints files.
-r other-requirements.txt
-c constraints.txt

# It is possible to refer to specific local distribution paths.
./downloads/numpy-1.9.2-cp34-none-win32.whl

# It is possible to refer to URLs.
http://wxpython.org/Phoenix/snapshot-builds/wxPython_Phoenix-3.0.3.dev1820+49a8884-cp34-none-win_amd64.whl
*/

func parseLine(line string, wg *sync.WaitGroup) (utils.Package, error) {

	defer wg.Done()

	line = strings.TrimSpace(line)
	// handles comment and empty lines
	if line == "" || strings.HasPrefix(line, "#") {
		return utils.Package{}, nil
	}

	// handles any of the reference or constraints tags, will
	// print a message here to tell user to run the other file
	// as well.
	if strings.HasPrefix(line, "-") {
		// TODO: should it just parse this file as well?
		fmt.Printf(`Parsed an input with a tag reference to another
                    file: %v. Please run the file through the tool
                    following this run through.`)
		return utils.Package{}, nil
	}

	// handles external packages
	if strings.HasPrefix(line, "http") {
		// TODO: May be an issue, can we find versions on external sites?
		return utils.Package{}, nil
	}

	// handles local refs
	if strings.HasPrefix(line, "./") ||
		strings.HasPrefix(line, "/") ||
		strings.HasSuffix(line, ".whl") {
		// TODO: May be an issue, can we find versions of local files?
		return utils.Package{}, nil
	}

	// regex to match name and optional extras [extras]
	// this should handle all of the other __stuff__
	re := regexp.MustCompile(`^([a-zA-Z0-9_\-]+)(\[[^\]]*\])?`)
	matches := re.FindStringSubmatch(line)
	if matches == nil {
		return utils.Package{}, fmt.Errorf("invalid format: '%s'", line)
	}

	name := matches[1]
	extras := strings.Trim(matches[2], "[]")
	remaining := strings.TrimSpace(line[len(matches[0]):])
	var versionSpecs []string
	var envMarker string

	// Check for environment marker (split on ;)
	if strings.Contains(remaining, ";") {
		parts := strings.SplitN(remaining, ";", 2)
		remaining = strings.TrimSpace(parts[0])
		envMarker = strings.TrimSpace(parts[1])
	}

	// Split version specifiers by commas
	if remaining != "" {
		versionSpecs = strings.Split(remaining, ",")
		for i := range versionSpecs {
			versionSpecs[i] = strings.TrimSpace(versionSpecs[i])
		}
	}

	return utils.Package{
		Name:         name,
		Extras:       extras,
		VersionSpecs: versionSpecs,
		EnvMarker:    envMarker,
	}, nil

}

func ParseFile(filePath string) ([]utils.Package, error) {
	var packages []utils.Package
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Printf("Failed to parse pip file: %v", err)
		return packages, fmt.Errorf("Failed to parse pip file: %v", err)
	}

	// scan the document and separate by newlines
	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)
	var packageStrings []string

	for scanner.Scan() {
		packageStrings = append(packageStrings, scanner.Text())
	}

	var packageList []utils.Package
	var wg sync.WaitGroup
	for _, pkg := range packageStrings {
		wg.Add(1)
		currPkg, err := parseLine(pkg, &wg)
		if err != nil {
			return packages, fmt.Errorf("An error occurred parsing package: %v", err)
		}
		packageList = append(packageList, currPkg)
	}

	wg.Wait()

	return packages, nil
}
