package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestConfig_LoadDefaults(t *testing.T) {
	// Temporarily clear environment variables
	os.Unsetenv("GRPC_PORT")
	os.Unsetenv("PORT")
	os.Unsetenv("GEMINI_API_KEY")
	os.Unsetenv("GEMINI_MODEL")
	os.Unsetenv("GEMINI_TIMEOUT_SECONDS")
	os.Unsetenv("GEMINI_TEMPERATURE")
	os.Unsetenv("GEMINI_MAX_OUTPUT_TOKENS")
	os.Unsetenv("GEMINI_RESPONSE_MIME_TYPE")

	// Call Load with a nonexistent file so local .env doesn't interfere with defaults test
	cfg := Load()

	if cfg.GRPCPort != "50051" {
		t.Errorf("expected GRPCPort 50051, got %s", cfg.GRPCPort)
	}
	if cfg.GeminiModel != "gemini-3.5-flash-lite" {
		t.Errorf("expected GeminiModel gemini-3.5-flash-lite, got %s", cfg.GeminiModel)
	}
	if cfg.GeminiTimeout != 30*time.Second {
		t.Errorf("expected GeminiTimeout 30s, got %v", cfg.GeminiTimeout)
	}
	if cfg.GeminiTemp != 0.2 {
		t.Errorf("expected GeminiTemp 0.2, got %v", cfg.GeminiTemp)
	}
	if cfg.GeminiMaxTokens != 4096 {
		t.Errorf("expected GeminiMaxTokens 4096, got %d", cfg.GeminiMaxTokens)
	}
	if cfg.ResponseMIMEType != "application/json" {
		t.Errorf("expected ResponseMIMEType application/json, got %s", cfg.ResponseMIMEType)
	}
}

func TestConfig_Validate(t *testing.T) {
	cfg := &Config{
		GeminiAPIKey: "",
		GeminiModel:  "gemini-2.5-flash",
	}
	if err := cfg.Validate(); err == nil {
		t.Errorf("expected error for empty API key, got nil")
	}

	cfg.GeminiAPIKey = "dummy-test-key"
	if err := cfg.Validate(); err != nil {
		t.Errorf("unexpected error for valid config: %v", err)
	}

	// Verify error message does not leak key or secrets
	cfg.GeminiModel = ""
	err := cfg.Validate()
	if err == nil || strings.Contains(err.Error(), "dummy-test-key") {
		t.Errorf("error message leaked key or was nil: %v", err)
	}
}

func TestConfig_LoadDotEnv(t *testing.T) {
	tempDir := t.TempDir()
	envPath := filepath.Join(tempDir, ".env.test")

	content := `# Comment line
GRPC_PORT=50099
GEMINI_MODEL="gemini-1.5-pro"
GEMINI_TIMEOUT_SECONDS='45'
GEMINI_TEMPERATURE=0.7
GEMINI_MAX_OUTPUT_TOKENS=8192
GEMINI_RESPONSE_MIME_TYPE=application/json
`
	if err := os.WriteFile(envPath, []byte(content), 0600); err != nil {
		t.Fatalf("failed to write test .env file: %v", err)
	}

	os.Unsetenv("GRPC_PORT")
	os.Unsetenv("GEMINI_MODEL")
	os.Unsetenv("GEMINI_TIMEOUT_SECONDS")
	os.Unsetenv("GEMINI_TEMPERATURE")
	os.Unsetenv("GEMINI_MAX_OUTPUT_TOKENS")
	os.Unsetenv("GEMINI_RESPONSE_MIME_TYPE")

	// Set one environment variable beforehand to verify process environment precedence
	os.Setenv("GRPC_PORT", "55555")

	loadDotEnv(envPath)

	if os.Getenv("GRPC_PORT") != "55555" {
		t.Errorf("expected process environment to take precedence, got %s", os.Getenv("GRPC_PORT"))
	}
	if os.Getenv("GEMINI_MODEL") != "gemini-1.5-pro" {
		t.Errorf("expected gemini-1.5-pro, got %s", os.Getenv("GEMINI_MODEL"))
	}
	if os.Getenv("GEMINI_TIMEOUT_SECONDS") != "45" {
		t.Errorf("expected 45, got %s", os.Getenv("GEMINI_TIMEOUT_SECONDS"))
	}
	if os.Getenv("GEMINI_TEMPERATURE") != "0.7" {
		t.Errorf("expected 0.7, got %s", os.Getenv("GEMINI_TEMPERATURE"))
	}
}

func TestConfig_EnvExampleTemplate(t *testing.T) {
	// Locate .env.example relative to repo root
	examplePaths := []string{
		"../../.env.example",
		"../.env.example",
		".env.example",
	}

	var data []byte
	var err error
	for _, p := range examplePaths {
		data, err = os.ReadFile(p)
		if err == nil {
			break
		}
	}

	if err != nil {
		t.Fatalf("failed to find or read .env.example: %v", err)
	}

	content := string(data)
	requiredVars := []string{
		"GEMINI_API_KEY",
		"GEMINI_MODEL",
		"GEMINI_TIMEOUT_SECONDS",
		"GEMINI_TEMPERATURE",
		"GEMINI_MAX_OUTPUT_TOKENS",
		"GEMINI_RESPONSE_MIME_TYPE",
		"GRPC_PORT",
	}

	for _, v := range requiredVars {
		if !strings.Contains(content, v) {
			t.Errorf(".env.example missing required variable: %s", v)
		}
	}
}
