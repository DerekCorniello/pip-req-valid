package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"os"
	"os/exec"
	"strings"

	"github.com/DerekCorniello/pip-req-valid/input"
	"github.com/DerekCorniello/pip-req-valid/output"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
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

func HandleRequest(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	contentType := request.Headers["Content-Type"]
	if !strings.HasPrefix(contentType, "multipart/form-data") {
		return events.APIGatewayProxyResponse{
			StatusCode: 400,
			Body:       "Expected multipart/form-data",
		}, nil
	}

	fileContent, err := parseMultipartForm(request.Body, request.Headers["Content-Type"])
	if err != nil {
		log.Println("Error parsing form data:", err)
		return events.APIGatewayProxyResponse{
			StatusCode: 400,
			Body:       "Error parsing file content",
		}, nil
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

	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Body:       string(jsonResponse),
		Headers:    map[string]string{"Content-Type": "application/json"},
	}, nil
}

func parseMultipartForm(body, contentType string) ([]byte, error) {
	// Find the boundary in the content-type header
	// Content-Type looks like "multipart/form-data; boundary=---boundary"
	parts := strings.Split(contentType, "boundary=")
	if len(parts) < 2 {
		return nil, fmt.Errorf("Invalid content type, boundary not found")
	}
	boundary := parts[1]

	// create a reader to read from the multipart form body
	reader := multipart.NewReader(strings.NewReader(body), boundary)

	form, err := reader.ReadForm(10 << 20) // 10 MB limit for the form data
	if err != nil {
		return nil, err
	}

	files, ok := form.File["file"]
	if !ok || len(files) == 0 {
		return nil, fmt.Errorf("no file found in form data")
	}

	file, err := files[0].Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open the file: %v", err)
	}
	defer file.Close()

	fileContent, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read the file: %v", err)
	}

	return fileContent, nil
}

func main() {
	// Start the Lambda handler
	lambda.Start(HandleRequest)
}
