package utils

// https://pypi.org/pypi/<package name>/<version>/json

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func GetAllowedPackageVersions(pkg *Package) ([]string, error) {
	url := fmt.Sprintf(
		"https://pypi.org/pypi/%s/json",
		pkg.Name,
	)
	resp, err := http.Get(url)
	var versions []string
	if err != nil {
		fmt.Printf("An error occurred: %v\n", err)
		return versions, err
	}
	defer resp.Body.Close() // close the network connection after reading

    // create an empty map that maps from string to an empty interface
    // allows for flexibility in the structure of the HTTP responses
	var packageInfo map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&packageInfo)

	if err != nil {
		fmt.Printf("Error parsing JSON response: %v", err)
		return versions, err
	}

	// Extract versions from the releases map
	releases, ok := packageInfo["releases"].(map[string]interface{})
	if !ok {
		fmt.Println("Error: 'releases' field is missing or malformed")
		return versions, fmt.Errorf("missing or malformed 'releases' field")
	}

	for version := range releases {
		versions = append(versions, version)
	}

	return versions, nil
}
