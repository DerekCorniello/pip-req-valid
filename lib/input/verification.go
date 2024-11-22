package input

import (
	"fmt"

	utils "github.com/DerekCorniello/pip-req-valid/utils"
)

func VerifyPackage(pkg utils.Package) bool {
    // for packages that need latest versions...
	if pkg.VersionSpecs[0] == "latest" {
		// as long as he name of the package exists, this is ok
		versions, err := utils.GetAllowedPackageVersions(&pkg)
		if err != nil {
			fmt.Printf("An error occurred trying to get package versions: %v")
			return false
		}

		return len(versions) != 0
	}

    // parse out all of the package info

    return false
}
