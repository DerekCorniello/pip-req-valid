package main

import (
	"fmt"
    "log"
    "os"
	"github.com/DerekCorniello/pip-req-valid/input"
	"github.com/DerekCorniello/pip-req-valid/output"
)

func main() {
	/* old workflow
	filepath := "./tests/test1.txt"
	pkgs, errs := input.ParseFile(filepath)
	verPkgs, invPkgs := input.VerifyPackages(pkgs)
	fmt.Printf(output.GetPrettyOutput(verPkgs, invPkgs, errs))
	*/
    filePath := "./tests/test1.txt"
	fileContent, err := os.ReadFile(filePath)
	if err != nil {
		log.Fatalf("Failed to read file: %v", err)
	}
	pkgs, errs := input.ParseFile(fileContent)
	verPkgs, invPkgs := input.VerifyPackages(pkgs)
	fmt.Printf(output.GetPrettyOutput(verPkgs, invPkgs, errs))
}
