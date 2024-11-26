package main

import (
	"github.com/DerekCorniello/pip-req-valid/input"
	"github.com/DerekCorniello/pip-req-valid/output"
	"testing"
)

var testCases map[string]string = map[string]string{
	"tests/test1.txt":  
`Verified the following packages:
        requests, numpy, flask
No packages had errors.
No processing errors.`,
        "tests/test2.txt":  
`No verified packages.
Found 3 error packages:
        requests, numpy, flask
No processing errors.`,
	"tests/test3.txt":  
`Verified the following packages:
        requests, flask, https://github.com/username/special-package.git@v1.0.0
Found 1 error packages:
        private-package
No processing errors.`,
	"tests/test4.txt":  
`Verified the following packages:
        requests, flask, pytest, black, mypy, numpy, pandas, cryptography
No packages had errors.
No processing errors.`,
	"tests/test5.txt":  
`Verified the following packages:
        requests, flask, numpy
No packages had errors.
No processing errors.`,
	"tests/test6.txt":  
`Verified the following packages:
        requests, flask
No packages had errors.
No processing errors.`,
	"tests/test7.txt":  
`Verified the following packages:
        requests, flask, numpy
No packages had errors.
No processing errors.`,
	"tests/test8.txt":  
`Verified the following packages:
        requests, flask, pytest, black
No packages had errors.
No processing errors.`,
	"tests/test9.txt":  
`No verified packages.
No packages had errors.
No processing errors.`,
	"tests/test10.txt": 
`Verified the following packages:
        flask, numpy
Found 1 error packages:
        request
No processing errors.`,
	"tests/test11.txt": 
`Verified the following packages:
        requests, numpy, flask
No packages had errors.
No processing errors.`,
	"tests/test12.txt": 
`No verified packages.
Found 4 error packages:
        -r base-requirements.txt, -e ., ../my-local-library/, ./dist/custom_package-1.0.0-py3-none-any.whl
No processing errors.`,
}

func TestParseAndVerifyRequirements(t *testing.T) {
	for fileName, expectedOutput := range testCases {
		t.Run(fileName, func(t *testing.T) {
			pkgs, errs := input.ParseFile(fileName)
			verPkgs, invPkgs := input.VerifyPackages(pkgs)
			actualOutput := output.GetPrettyOutput(verPkgs, invPkgs, errs)
			if actualOutput != expectedOutput {
				t.Errorf("Output mismatch for %s\nExpected:\n%s\nGot:\n%s", fileName, expectedOutput, actualOutput)
			}
		})
	}
}
