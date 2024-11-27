package input

import (
	"bufio"
	"bytes"
	"fmt"
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

	// comments can trail actual commands, split it here
	if strings.Contains(line, "#") {
		line = strings.Split(line, "#")[0]
	}

	// handles any of the reference or constraints tags, will
	// print a message here to tell user to run the other file
	// as well.
	if strings.HasPrefix(line, "-") {
		fmt.Printf("Parsed an input with a tag reference to another file: %v. Please run the file through the tool following this run.\n", line)
		return utils.Package{Name: line, VersionSpecs: []string{"local"}}, nil
	}

	// handles external packages
	if strings.Contains(line, "http") {
		re := regexp.MustCompile(`http.*`)
		matches := re.FindStringSubmatch(line)
		if matches == nil {
			return utils.Package{},
				fmt.Errorf("Error in regex operation finding http url in '%v'", line)
		}
		return utils.Package{Name: matches[0], VersionSpecs: []string{"latest", "url"}}, nil
	}

	// handles local refs
	if strings.HasPrefix(line, ".") ||
		strings.HasPrefix(line, "/") ||
		strings.HasPrefix(line, "..") ||
		strings.HasSuffix(line, ".whl") {

		fmt.Printf("Cannot verify local file: %v\n", line)
		return utils.Package{Name: line, VersionSpecs: []string{"local"}}, nil
	}

	// regex to match name and optional extras [extras]
	// this should handle all of the other __stuff__
	re := regexp.MustCompile(`^([a-zA-Z0-9_\-]+)(\[[^\]]*\])?`)
	matches := re.FindStringSubmatch(line)
	if matches == nil {
		return utils.Package{Name: "invalid"}, fmt.Errorf("invalid format: '%s'", line)
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

func ParseFile(fileContent []byte) ([]utils.Package, []error) {

	reader := bytes.NewReader(fileContent)
	scanner := bufio.NewScanner(reader)
	scanner.Split(bufio.ScanLines)

	var packageStrings []string

	for scanner.Scan() {
		packageStrings = append(packageStrings, scanner.Text())
	}

	var packageList []utils.Package
	var errList []error
	var wg sync.WaitGroup
	for _, pkg := range packageStrings {
		wg.Add(1)
		currPkg, err := parseLine(pkg, &wg)
		// we don't need empty package names, those are comments or tag reqs
		// or it is an errored package that will be handled
		if currPkg.Name != "" {
			if err != nil {
				errList = append(errList, fmt.Errorf("An error occurred parsing package: %v", err))
			}
			packageList = append(packageList, currPkg)
		}
	}

	wg.Wait()

	return packageList, errList
}

func VerifyPackages(packages []utils.Package) ([]utils.Package, []utils.Package) {
	var verifiedPackages, invalidPackages []utils.Package
	for _, pkg := range packages {
		if VerifyPackage(pkg) {
			verifiedPackages = append(verifiedPackages, pkg)
		} else {
			invalidPackages = append(invalidPackages, pkg)
		}
	}
	return verifiedPackages, invalidPackages
}
