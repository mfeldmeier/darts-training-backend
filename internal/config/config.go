package config

import (
	"encoding/base64"
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port              string
	DatabaseURL       string
	Auth0Domain       string
	Auth0ClientID     string
	Auth0ClientSecret string
	JWTSecret         string
	FrontendURL       string

	// Additional fields for AuthManager
	OidcBaseURL                     string
	ClientCredentialAuthHeaderValue string
}

func LoadConfig() (*Config, error) {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		// .env file not found is not an error in production
		fmt.Println("No .env file found, using environment variables")
	}

	auth0Domain := getEnv("AUTH0_DOMAIN", "")

	config := &Config{
		Port:              getEnv("PORT", "8080"),
		DatabaseURL:       getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost/darts_training?sslmode=disable"),
		Auth0Domain:       auth0Domain,
		Auth0ClientID:     getEnv("AUTH0_CLIENT_ID", ""),
		Auth0ClientSecret: getEnv("AUTH0_CLIENT_SECRET", ""),
		JWTSecret:         getEnv("JWT_SECRET", "your-secret-key"),
		FrontendURL:       getEnv("FRONTEND_URL", "http://localhost:4200"),

		// Calculate AuthManager fields
		OidcBaseURL:                     getEnv("OIDC_BASE_URL", "https://"+auth0Domain),
		ClientCredentialAuthHeaderValue: calculateAuthHeader(getEnv("AUTH0_CLIENT_ID", ""), getEnv("AUTH0_CLIENT_SECRET", "")),
	}

	// Validate required fields
	if config.Auth0Domain == "" || config.Auth0ClientID == "" || config.Auth0ClientSecret == "" {
		return nil, fmt.Errorf("Auth0 configuration is missing. Please set AUTH0_DOMAIN, AUTH0_CLIENT_ID, and AUTH0_CLIENT_SECRET")
	}

	return config, nil
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func calculateAuthHeader(clientID, clientSecret string) string {
	if clientID == "" || clientSecret == "" {
		return ""
	}

	credentials := clientID + ":" + clientSecret
	encodedCredentials := base64.StdEncoding.EncodeToString([]byte(credentials))
	return encodedCredentials
}
