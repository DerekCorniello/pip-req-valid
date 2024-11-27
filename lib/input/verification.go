package input

import (
	"fmt"
	"slices"
	"strings"

	"github.com/Masterminds/semver/v3"

	utils "github.com/DerekCorniello/pip-req-valid/utils"
)

func parseVersionSpecifier(spec string) (string, string, error) {
	operators := []string{"==", ">=", "<=", ">", "<", "~=", "!="}

	for _, op := range operators {
		if strings.HasPrefix(spec, op) {
			return op, strings.TrimSpace(strings.TrimPrefix(spec, op)), nil
		}
	}

	return "", "", fmt.Errorf("invalid version specifier: %s", spec)
}

func versionMatchesConstraint(version, operator, constraint string, details *[]string) bool {
	parsedVersion, err := semver.NewVersion(version)
	if err != nil {
		*details = append(*details, fmt.Sprintf("Invalid version format: %s\n", version))
		return false
	}

	parsedConstraint, err := semver.NewConstraint(fmt.Sprintf("%s %s", operator, constraint))
	if err != nil {
		*details = append(*details, fmt.Sprintf("Invalid version constraint: %s %s\n", operator, constraint))
		return false
	}

	return parsedConstraint.Check(parsedVersion)
}

func VerifyPackage(pkg utils.Package, details *[]string) bool {

	versions, err := utils.GetAllowedPackageVersions(&pkg, details)
	if err != nil {
		*details = append(*details, fmt.Sprintf("An error occurred while retrieving allowed package versions: %v\n", err))
		return false
	}

	if slices.Contains(pkg.VersionSpecs, "latest") {

		return len(versions) > 0
	}

	for _, spec := range pkg.VersionSpecs {
		if strings.HasPrefix(spec, "==") {
			targetVersion := strings.TrimPrefix(spec, "==")
			if slices.Contains(versions, targetVersion) {
				return true
			} else {
				*details = append(*details, fmt.Sprintf("Specified version '%s' not found for package '%s'.\n", targetVersion, pkg.Name))
				return false
			}
		}
	}

	for _, version := range versions {
		versionValid := true
		for _, spec := range pkg.VersionSpecs {
			op, targetVersion, parseErr := parseVersionSpecifier(spec)
			if parseErr != nil {
				*details = append(*details, fmt.Sprintf("Error parsing version specifier '%s': %v\n", spec, parseErr))
				versionValid = false
				break
			}

			if !versionMatchesConstraint(version, op, targetVersion, details) {
				versionValid = false
				break
			}
		}

		if versionValid {
			return true
		}
	}

	return false
}
