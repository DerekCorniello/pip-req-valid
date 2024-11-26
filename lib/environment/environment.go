package environment

import (
	"bytes"
	"errors"
	"os/exec"
	"path/filepath"
)

// InstallRequirements installs the Python dependencies listed in the given requirements.txt file.
// It takes the filename as input and returns a response string and an error if the installation fails.
func InstallRequirements(filename string) (string, error) {
	if filepath.Ext(filename) != ".txt" {
		return "", errors.New("invalid file type: must be a .txt file")
	}

	cmd := exec.Command("pip", "install", "-r", filename)

	// capture the standard output and standard error
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return stderr.String(), errors.New("failed to install requirements")
	}
	return stdout.String(), nil
}
