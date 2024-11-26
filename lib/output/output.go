package output

import (
	"fmt"
	"strings"

	"github.com/DerekCorniello/pip-req-valid/utils"
)

// I wanted to try enums too
type MessageType int

const (
	VerifiedPackages MessageType = iota
	ErrorPackages
	ProcessingErrors
)

// this is interesting to me, using generics, we can say that both the
// items slice and mapper function's input must be of the same type T
// that is what the [T any] means. If we use `any` on both of those,
// it could allow for mismatched types. Super cooL!
func extractStrings[T any](items []T, mapper func(T) string) []string {
	result := []string{}
	for _, item := range items {
		result = append(result, mapper(item))
	}
	return result
}

func createMessage(items []string, msgType MessageType) string {
	if len(items) == 0 {
		var typeString string
		switch msgType {
		case VerifiedPackages:
			typeString = "verified packages"
		case ErrorPackages:
			typeString = "packages had errors"
		case ProcessingErrors:
			typeString = "processing errors"
		}
		return fmt.Sprintf("No %v.", typeString)
	}

	switch msgType {
	case VerifiedPackages:
		return fmt.Sprintf(
			"Verified the following packages:\n%v",
			"\t"+strings.Join(items, ", "),
		)
	case ErrorPackages:
		return fmt.Sprintf(
			"Found %d error packages:\n%v",
			len(items),
			"\t"+strings.Join(items, ", "),
		)
	case ProcessingErrors:
		return fmt.Sprintf(
			"Encountered %d processing errors:\n%v",
			len(items),
			"\t"+strings.Join(items, ", "),
		)
	default:
		return fmt.Sprintf("Unknown message type: %v", msgType)
	}
}

func GetPrettyOutput(verifiedPackages []utils.Package,
	errorPackages []utils.Package, errs []error) string {

	// cool, we can use anonymous funcs too! I love Go
	// create a list of each of the string versions
	// of each list, and we can format into a long string
	csVerPkgs := extractStrings(verifiedPackages,
		func(pkg utils.Package) string { return pkg.Name })
	csErrPkgs := extractStrings(errorPackages,
		func(pkg utils.Package) string { return pkg.Name })
	csErrs := extractStrings(errs,
		func(err error) string { return err.Error() })

	s := fmt.Sprintf("%v\n%v\n%v", createMessage(csVerPkgs, MessageType(VerifiedPackages)),
		createMessage(csErrPkgs, MessageType(ErrorPackages)),
		createMessage(csErrs, MessageType(ProcessingErrors)))
	return s
}
