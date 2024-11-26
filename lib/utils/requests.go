package utils

import (
	"encoding/json"
	"fmt"
	"net/http"
	"slices"
)

func GetAllowedPackageVersions(pkg *Package) ([]string, error) {
	if pkg.Name == "" {
		return nil, nil
	} else if slices.Contains(pkg.VersionSpecs, "local") {
		return nil, nil
	}
	var url string
	isURL := slices.Contains(pkg.VersionSpecs, "url")
	if isURL {
		url = pkg.Name
	} else {
		url = fmt.Sprintf("https://pypi.org/pypi/%s/json", pkg.Name)
	}

	// Perform HTTP GET request
	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("An error occurred: %v\n", err)
		return nil, err
	}
	defer resp.Body.Close() // Ensure the response body is closed

	if !isURL {

		var packageInfo map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&packageInfo)
		if err != nil {
			fmt.Printf("Error parsing JSON response: %v\n%v\n", err, packageInfo)
			return nil, err
		}

		// Extract the "releases" map
		releases, ok := packageInfo["releases"].(map[string]interface{})
		if !ok {
			fmt.Printf("Error: Package `%v` was not found.", pkg)
			return nil, fmt.Errorf("Package with specified version was not found.")
		}

		// Collect the versions (keys of the "releases" map)
		var versions []string
		for version := range releases {
			versions = append(versions, version)
		}
		return versions, nil
	}

	return []string{"latest"}, nil
}
