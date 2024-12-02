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
	"github.com/joho/godotenv"
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

func handleRequest(writer http.ResponseWriter, reader *http.Request) {
	API_URL := os.Getenv("API_URL")
	if API_URL == "" {
		panic("No API URL Found!")
	}
	AUTH_TOKEN := os.Getenv("AUTH_TOKEN")
	if AUTH_TOKEN == "" {
		panic("No Auth Token found!")
	}

	// Set CORS headers
	writer.Header().Set("Access-Control-Allow-Origin", API_URL)
	writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")

	if reader.Method == http.MethodOptions {
		writer.WriteHeader(http.StatusOK)
		return
	}

	if reader.Method != http.MethodPost {
		http.Error(writer, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}

	authHeader := reader.Header.Get("Authorization")
	if !strings.HasPrefix(authHeader, "Bearer ") {
		http.Error(writer, "Unauthorized: Missing or incorrect Authorization header", http.StatusUnauthorized)
		return
	}

	token := strings.TrimPrefix(authHeader, "Bearer ")
	if token != AUTH_TOKEN {
		http.Error(writer, "Unauthorized: Invalid token", http.StatusUnauthorized)
		return
	}

	contentType := reader.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "multipart/form-data") {
		http.Error(writer, "Expected multipart/form-data", http.StatusBadRequest)
		return
	}

	fileContent, err := parseMultipartForm(reader)
	if err != nil {
		log.Println("Error parsing form data:", err)
		http.Error(writer, "Error parsing file content", http.StatusBadRequest)
		return
	}

	pkgs, errs := input.ParseFile(fileContent)

	errList := []string{}
	for _, err := range errs {
		errList = append(errList, err.Error())
	}

	verPkgs, invPkgs, details := input.VerifyPackages(pkgs)

	installOutput, installErr := RunDockerInstall(fileContent)
	if installErr != nil {
		errList = append(errList, installErr.Error())
	}

	response := map[string]interface{}{
		"prettyOutput":  output.GetPrettyOutput(verPkgs, invPkgs, errs), // formatted output
		"details":       strings.Join(details, "\n"),                    // details of the process
		"errors":        strings.Join(errList, "\n"),                    // errors occurred during processing
		"installOutput": installOutput,                                  // test install output
	}

	jsonResponse, err := json.Marshal(response)
	if err != nil {
		http.Error(writer, "Failed to encode response", http.StatusInternalServerError)
		return
	}

	writer.WriteHeader(http.StatusOK)
	writer.Write(jsonResponse)
}

func parseMultipartForm(r *http.Request) ([]byte, error) {
	// Parse the multipart form
	err := r.ParseMultipartForm(10 << 20) // 10 MB limit for the form data
	if err != nil {
		return nil, err
	}

	file, _, err := r.FormFile("file")
	if err != nil {
		return nil, fmt.Errorf("no file found in form data")
	}
	defer file.Close()

	fileContent, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read the file: %v", err)
	}

	return fileContent, nil
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}
	http.HandleFunc("/", handleRequest)
	port := "8080"
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
