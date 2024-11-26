package main

import (
	"github.com/DerekCorniello/pip-req-valid/input"
	"github.com/DerekCorniello/pip-req-valid/output"
	"testing"
)

var testFiles []string = []string{
	"tests/test1.txt",
	"tests/test1.txt",
	"tests/test3.txt",
	"tests/test4.txt",
	"tests/test5.txt",
	"tests/test6.txt",
	"tests/test7.txt",
	"tests/test8.txt",
	"tests/test9.txt",
	"tests/test10.txt",
	"tests/test11.txt",
}

func TestParseAndVerifyRequirements(t *testing.T) {
	for _, testCase := range testFiles {
		t.Run(testCase, func(t *testing.T) {
			pkgs, errs := input.ParseFile(testCase)
			verPkgs, invPkgs := input.VerifyPackages(pkgs)
			outputMessage := output.GetPrettyOutput(verPkgs, invPkgs, errs)
			t.Logf("Output for %s:\n%s", testCase, outputMessage)
		})
	}
}
