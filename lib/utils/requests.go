package utils

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func GetAllowedPackageVersions(pkg *Package) ([]string, error) {
	if pkg.Name == "" {
		return nil, nil
	}
	url := fmt.Sprintf("https://pypi.org/pypi/%s/json", pkg.Name)

	// Perform HTTP GET request
	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("An error occurred: %v\n", err)
		return nil, err
	}
	defer resp.Body.Close() // Ensure the response body is closed

	var packageInfo map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&packageInfo)
	if err != nil {
		fmt.Printf("Error parsing JSON response: %v\n%v\n", err, packageInfo)
		return nil, err
	}

	// Extract the "releases" map
	releases, ok := packageInfo["releases"].(map[string]interface{})
	if !ok {
		fmt.Println("Error: Package was not found.")
		return nil, fmt.Errorf("Package with specified version was not found.")
	}

	// Collect the versions (keys of the "releases" map)
	var versions []string
	for version := range releases {
		versions = append(versions, version)
	}

	return versions, nil
}
