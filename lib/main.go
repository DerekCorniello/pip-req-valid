package main

import (
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"strings"

	"github.com/DerekCorniello/pip-req-valid/input"
	"github.com/DerekCorniello/pip-req-valid/output"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

// Lambda handler to process the file
func HandleRequest(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// Check if it's multipart/form-data
	contentType := request.Headers["Content-Type"]
	if !strings.HasPrefix(contentType, "multipart/form-data") {
		return events.APIGatewayProxyResponse{
			StatusCode: 400,
			Body:       "Expected multipart/form-data",
		}, nil
	}

	// Parse the multipart form data
	fileContent, err := parseMultipartForm(request.Body, request.Headers["Content-Type"])
	if err != nil {
		log.Println("Error parsing form data:", err)
		return events.APIGatewayProxyResponse{
			StatusCode: 400,
			Body:       "Error parsing file content",
		}, nil
	}

	// Use the existing ParseFile function with the file content
	pkgs, errs := input.ParseFile(fileContent)
	verPkgs, invPkgs := input.VerifyPackages(pkgs)

	// Get the formatted output
	prettyOutput := output.GetPrettyOutput(verPkgs, invPkgs, errs)

	// Return the output as JSON (frontend expects this)
	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Body:       fmt.Sprintf("{\"output\": \"%s\"}", prettyOutput),
		Headers:    map[string]string{"Content-Type": "application/json"},
	}, nil
}

// Helper function to parse the multipart form and extract the file content
func parseMultipartForm(body, contentType string) ([]byte, error) {
	// Parse the multipart form using the boundary in the content type header
	// Extract the file content from the form-data
	// Create a new reader to handle the multipart data

	// Find the boundary in the content-type header
	// Content-Type looks like "multipart/form-data; boundary=---boundary"
	parts := strings.Split(contentType, "boundary=")
	if len(parts) < 2 {
		return nil, fmt.Errorf("Invalid content type, boundary not found")
	}
	boundary := parts[1]

	// Create a reader to read from the multipart form body
	reader := multipart.NewReader(strings.NewReader(body), boundary)

	// Read the form data
	form, err := reader.ReadForm(10 << 20) // 10 MB limit for the form data
	if err != nil {
		return nil, err
	}

	// Extract the file content from the form data
	files, ok := form.File["file"] // assuming the file is passed with the key "file"
	if !ok || len(files) == 0 {
		return nil, fmt.Errorf("no file found in form data")
	}

	// Now extract the file
	file, err := files[0].Open() // Open the first file in the list
	if err != nil {
		return nil, fmt.Errorf("failed to open the file: %v", err)
	}
	defer file.Close()

	// Read the content of the uploaded file
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
