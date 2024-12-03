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
	"time"

	"github.com/DerekCorniello/pip-req-valid/input"
	"github.com/DerekCorniello/pip-req-valid/output"

	"github.com/golang-jwt/jwt/v4"
	"github.com/joho/godotenv"
)

var jwtKey = []byte(os.Getenv("SECRET_TOKEN"))

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

func validateToken(tokenString string) (*jwt.Token, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return jwtKey, nil
	})
	if err != nil {
		return nil, err
	}

	_, ok := token.Claims.(jwt.MapClaims)

	if ok && token.Valid {
		return token, nil
	}

	return nil, fmt.Errorf("invalid token")
}

func handleRequest(writer http.ResponseWriter, reader *http.Request) {
	// Set CORS headers
	writer.Header().Set("Access-Control-Allow-Origin", "https://www.reqinspect.com")
	writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")

	// Handle preflight OPTIONS request
	if reader.Method == http.MethodOptions {
		writer.WriteHeader(http.StatusOK)
		return
	}

	if reader.Method != http.MethodPost && reader.Method != http.MethodOptions {
		http.Error(writer, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}

	auth := reader.Header.Get("Authorization")

	if !strings.HasPrefix(auth, "Bearer ") {
		http.Error(writer, "Unauthenticated Request: Missing Bearer Token", http.StatusUnauthorized)
		return
	}
	tokenString := auth[len("Bearer "):]
	_, err := validateToken(tokenString)
	if err != nil {
		http.Error(writer, fmt.Sprintf("Unauthenticated Request: %s", err.Error()), http.StatusUnauthorized)
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

func parseMultipartForm(reader *http.Request) ([]byte, error) {
	// Parse the multipart form
	err := reader.ParseMultipartForm(10 << 20) // 10 MB limit for the form data
	if err != nil {
		return nil, err
	}

	file, _, err := reader.FormFile("file")
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

func handleAuth(writer http.ResponseWriter, reader *http.Request) {
	// Set CORS headers
	writer.Header().Set("Access-Control-Allow-Origin", "https://www.reqinspect.com")
	writer.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	writer.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")

	// Handle preflight OPTIONS request
	if reader.Method == http.MethodOptions {
		writer.WriteHeader(http.StatusOK)
		return
	}

	if reader.Method != http.MethodGet && reader.Method != http.MethodOptions {
		http.Error(writer, "Only GET method is allowed", http.StatusMethodNotAllowed)
		return
	}

	contentType := reader.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "application/json") {
		http.Error(writer, "Expected application/json data", http.StatusBadRequest)
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"exp": time.Now().Add(15 * time.Second).Unix(), // Expires in 15 seconds
		"iss": "api.reqinspect.com",                    // Issuer
	})

	signedToken, err := token.SignedString(jwtKey)
	if err != nil {
		fmt.Println("Error signing token:", err)
		return
	}

	jsonResponse := map[string]interface{}{
		"token": signedToken,
	}

	responseData, err := json.Marshal(jsonResponse)
	if err != nil {
		fmt.Println("Error marshaling JSON:", err)
		return
	}

	writer.WriteHeader(http.StatusOK)
	writer.Write(responseData)
}

func main() {
	err := godotenv.Load()
	if err != nil {
		panic(fmt.Sprintf("Error loading .env!\n%v", err.Error()))
	}
	http.HandleFunc("/", handleRequest)
	http.HandleFunc("/auth/", handleAuth)
	port := "8080"
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
