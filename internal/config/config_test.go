package config

import (
	"os"
	"testing"
	"time"
)

func TestConfig_LoadDefaults(t *testing.T) {
	os.Unsetenv("GRPC_PORT")
	os.Unsetenv("PORT")
	os.Unsetenv("GEMINI_API_KEY")
	os.Unsetenv("GEMINI_MODEL")
	os.Unsetenv("GEMINI_TIMEOUT_SECONDS")
	os.Unsetenv("GEMINI_TEMPERATURE")
	os.Unsetenv("GEMINI_MAX_OUTPUT_TOKENS")
	os.Unsetenv("GEMINI_RESPONSE_MIME_TYPE")

	cfg := Load()

	if cfg.GRPCPort != "50051" {
		t.Errorf("expected GRPCPort 50051, got %s", cfg.GRPCPort)
	}
	if cfg.GeminiModel != "gemini-2.5-flash" {
		t.Errorf("expected GeminiModel gemini-2.5-flash, got %s", cfg.GeminiModel)
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

	cfg.GeminiAPIKey = "test-key"
	if err := cfg.Validate(); err != nil {
		t.Errorf("unexpected error for valid config: %v", err)
	}
}
