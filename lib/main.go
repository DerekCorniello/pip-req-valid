package main

import (
	"context"
	"crypto/rand"
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
	"golang.org/x/time/rate"
)

var jwtKey []byte

func generateRandomKey() []byte {
	key := make([]byte, 32)
	_, err := rand.Read(key)
	if err != nil {
		panic("Failed to generate random key")
	}
	return key
}

func init() {
	token := os.Getenv("SECRET_TOKEN")
	if token == "" {
		jwtKey = generateRandomKey()
	} else {
		jwtKey = []byte(token)
	}
}

func CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("CORS middleware, method: %s, path: %s", r.Method, r.URL.Path)
		// Set the CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:5173")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			log.Printf("Handling OPTIONS in middleware")
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func RateLimitMiddleware(limiter *rate.Limiter, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !limiter.Allow() {
			http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func RunDockerInstall(requirements []byte) (string, error) {
	log.Printf("Starting RunDockerInstall")
	// save the file temporarily
	tmpFile, err := os.CreateTemp("/host_tmp", "requirements-*.txt")
	if err != nil {
		return "", fmt.Errorf("could not create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	// write the requirements.txt content to the temporary file
	_, err = tmpFile.Write(requirements)
	if err != nil {
		return "", fmt.Errorf("could not write to temp file: %v", err)
	}
	tmpFile.Close()

	hostPath := strings.Replace(tmpFile.Name(), "/host_tmp", "/tmp", 1)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "docker", "run", "--rm", "-v", fmt.Sprintf("%s:/app/requirements.txt", hostPath), "my-python-git", "sh", "-c", "mkdir -p /app && pip install --progress-bar off --disable-pip-version-check --no-cache-dir --root-user-action ignore -r /app/requirements.txt")
	output, err := cmd.CombinedOutput()
	if err != nil {
		if err == context.DeadlineExceeded {
			return "Pip install timed out after 30 seconds.", nil
		}
		return fmt.Sprintf("Pip install failed: %s", string(output)), nil
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
	log.Printf("Main request received, method: %s", reader.Method)
	// Set CORS headers
	origin := reader.Header.Get("Origin")
	if origin == "" {
		origin = "*"
	}
	writer.Header().Set("Access-Control-Allow-Origin", origin)
	writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")

	// Handle preflight OPTIONS request
	if reader.Method == http.MethodOptions {
		log.Printf("Handling OPTIONS for main request")
		writer.WriteHeader(http.StatusOK)
		return
	}

	if reader.Method != http.MethodPost && reader.Method != http.MethodOptions {
		log.Printf("Invalid method for main request: %s", reader.Method)
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
	log.Printf("Parsed multipart form, file size: %d", len(fileContent))

	pkgs, errs := input.ParseFile(fileContent)
	log.Printf("Parsed file, packages: %d, errors: %d", len(pkgs), len(errs))

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
		log.Printf("Failed to marshal JSON: %v", err)
		http.Error(writer, "Failed to encode response", http.StatusInternalServerError)
		return
	}

	log.Printf("Sending response for main request")
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusOK)
	if _, err := writer.Write(jsonResponse); err != nil {
		log.Printf("Error writing response: %v", err)
	}
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
	log.Printf("Auth request received, method: %s", reader.Method)
	// Set CORS headers
	origin := reader.Header.Get("Origin")
	if origin == "" {
		origin = "*"
	}
	writer.Header().Set("Access-Control-Allow-Origin", origin)
	writer.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")

	// Handle preflight OPTIONS request
	if reader.Method == http.MethodOptions {
		log.Printf("Handling OPTIONS for auth")
		writer.WriteHeader(http.StatusOK)
		return
	}

	if reader.Method != http.MethodGet && reader.Method != http.MethodOptions {
		log.Printf("Invalid method for auth: %s", reader.Method)
		http.Error(writer, "Only GET method is allowed", http.StatusMethodNotAllowed)
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"exp": time.Now().Add(time.Second).Unix(), // Expires in 15 seconds
		"iss": "api.reqinspect.com",               // Issuer
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

	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusOK)
	writer.Write(responseData)
}

func main() {
	godotenv.Load() // Load .env if exists, otherwise use env vars
	limiter := rate.NewLimiter(rate.Every(time.Minute), 10)

	http.Handle("/", CORSMiddleware(RateLimitMiddleware(limiter, http.HandlerFunc(handleRequest))))
	http.Handle("/auth", CORSMiddleware(RateLimitMiddleware(limiter, http.HandlerFunc(handleAuth))))
	port := "8080"
	log.Printf("Server starting on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
