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

func RunEFSInstall(requirements []byte) (string, error) {
	efsMountPath := "/requirements.txt"
	err := os.WriteFile(efsMountPath, requirements, 0644)
	if err != nil {
		return "", fmt.Errorf("could not write to EFS: %v", err)
	}

	cmd := exec.Command("pip", "install", "--no-cache-dir", "--dry-run", "-r", efsMountPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("pip install failed: %v\n%s", err, string(output))
	}

	return string(output), nil
}

func HandleRequest(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	contentType := request.Headers["Content-Type"]
	if !strings.HasPrefix(contentType, "multipart/form-data") {
		return events.APIGatewayProxyResponse{
			StatusCode: 400,
			Body:       "Expected multipart/form-data",
			Headers: map[string]string{
				"Access-Control-Allow-Origin":  "https://www.reqinspect.com",
				"Access-Control-Allow-Headers": "Content-Type",
				"Access-Control-Allow-Methods": "POST, OPTIONS",
			},
		}, nil
	}

	// Parse the file content from the multipart form data
	fileContent, err := parseMultipartForm(request.Body, request.Headers["Content-Type"])
	if err != nil {
		log.Println("Error parsing form data:", err)
		return events.APIGatewayProxyResponse{
			StatusCode: 400,
			Body:       "Error parsing file content",
			Headers: map[string]string{
				"Access-Control-Allow-Origin":  "https://www.reqinspect.com",
				"Access-Control-Allow-Headers": "Content-Type",
				"Access-Control-Allow-Methods": "POST, OPTIONS",
			},
		}, nil
	}

	pkgs, errs := input.ParseFile(fileContent)

	errList := []string{}
	for _, err := range errs {
		errList = append(errList, err.Error())
	}

	verPkgs, invPkgs, details := input.VerifyPackages(pkgs)

	installOutput, installErr := RunEFSInstall(fileContent)
	if installErr != nil {
		errList = append(errList, installErr.Error())
	}

	response := map[string]interface{}{
		"prettyOutput":  output.GetPrettyOutput(verPkgs, invPkgs, errs), // formatted output
		"details":       strings.Join(details, "\n"),                    // details of the process
		"errors":        strings.Join(errList, "\n"),                    // errors occurred during processing
		"installOutput": installOutput,                                  // test install output
	}

	jsonResponse, _ := json.Marshal(response)

	// Return the response
	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Body:       string(jsonResponse),
		Headers: map[string]string{
			"Access-Control-Allow-Origin":  "https://www.reqinspect.com",
			"Access-Control-Allow-Headers": "Content-Type",
			"Access-Control-Allow-Methods": "POST, OPTIONS",
		},
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
