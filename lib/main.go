package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"github.com/DerekCorniello/pip-req-valid/input"
	"github.com/DerekCorniello/pip-req-valid/output"
)

func RunDockerInstall(requirements []byte) (string, error) {
	// save the file temporarily
	tmpFile, err := os.CreateTemp("", "requirements-*.txt")
	if err != nil {
		return "", fmt.Errorf("could not create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	// write the requirements.txt content to the temporary file
	_, err = tmpFile.Write(requirements)
	if err != nil {
		return "", fmt.Errorf("could not write to temp file: %v", err)
	}

	cmd := exec.Command("docker", "run", "--rm", "-v", fmt.Sprintf("%s:/app/requirements.txt", tmpFile.Name()), "python:3.11-slim", "pip", "install", "--no-cache-dir", "-r", "/app/requirements.txt")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("docker install failed: %v\n%s", err, string(output))
	}

	return string(output), nil
}

func handleFileUpload(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "https://reqinspect.com")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")

	// Check if the content type is multipart/form-data
	contentType := r.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "multipart/form-data") {
		http.Error(w, "Expected multipart/form-data", http.StatusBadRequest)
		return
	}

	fileContent, err := parseMultipartForm(r)
	if err != nil {
		log.Println("Error parsing form data:", err)
		http.Error(w, "Error parsing file content", http.StatusBadRequest)
		return
	}

	pkgs, errs := input.ParseFile(fileContent)

	errList := []string{}
	for _, err := range errs {
		errList = append(errList, err.Error())
	}

	verPkgs, invPkgs, details := input.VerifyPackages(pkgs)

	response := map[string]interface{}{
		"prettyOutput": output.GetPrettyOutput(verPkgs, invPkgs, errs), // formatted output
		"details":      strings.Join(details, "\n"),                    // details of the process
		"errors":       strings.Join(errList, "\n"),                    // errors occurred during processing
	}

	jsonResponse, _ := json.Marshal(response)

	w.WriteHeader(http.StatusOK)
	w.Write(jsonResponse)
}

func parseMultipartForm(r *http.Request) ([]byte, error) {
	err := r.ParseMultipartForm(10 << 20) // 10 MB limit
	if err != nil {
		return nil, fmt.Errorf("failed to parse form: %v", err)
	}

	files := r.MultipartForm.File["file"]
	if len(files) == 0 {
		return nil, fmt.Errorf("no file found in form data")
	}

	file, err := files[0].Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()

	fileContent, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read file content: %v", err)
	}

	return fileContent, nil
}

func main() {
	// Set up the HTTP server
	http.HandleFunc("/upload", handleFileUpload)

	err := http.ListenAndServe(":443", nil)

	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}
