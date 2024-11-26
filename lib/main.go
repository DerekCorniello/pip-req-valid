package main

import (
	"fmt"
	"github.com/DerekCorniello/pip-req-valid/input"
	"github.com/DerekCorniello/pip-req-valid/output"
)

func main() {
	filepath := "./tests/test1.txt"
	pkgs, errs := input.ParseFile(filepath)
	verPkgs, invPkgs := input.VerifyPackages(pkgs)
	fmt.Printf(output.GetPrettyOutput(verPkgs, invPkgs, errs))
}
