package main

import (
	"fmt"
	"github.com/DerekCorniello/pip-req-valid/input"
	"github.com/DerekCorniello/pip-req-valid/output"
)

func main() {
	filepath := "./requirements.txt"
	pkgs, errs := input.ParseFile(filepath)
	verPkgs, invPkgs := input.VerifyPackages(pkgs)
	fmt.Printf(output.GetPrettyOutput(verPkgs, invPkgs, errs))
}
