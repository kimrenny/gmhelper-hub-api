package config

import (
	"errors"
	"os"
	"strconv"
	"time"
)

type Config struct {
	GRPCPort         string
	GeminiAPIKey     string
	GeminiModel      string
	GeminiTimeout    time.Duration
	GeminiTemp       float32
	GeminiMaxTokens  int32
	ResponseMIMEType string
}

func Load() *Config {
	loadDotEnv()

	port := os.Getenv("GRPC_PORT")
	if port == "" {
		port = os.Getenv("PORT")
	}
	if port == "" {
		port = "50051"
	}

	apiKey := os.Getenv("GEMINI_API_KEY")

	model := os.Getenv("GEMINI_MODEL")
	if model == "" {
		model = "gemini-3.5-flash-lite"
	}

	timeoutSec := 30
	if secStr := os.Getenv("GEMINI_TIMEOUT_SECONDS"); secStr != "" {
		if sec, err := strconv.Atoi(secStr); err == nil && sec > 0 {
			timeoutSec = sec
		}
	}

	var temp float32 = 0.2
	if tempStr := os.Getenv("GEMINI_TEMPERATURE"); tempStr != "" {
		if t, err := strconv.ParseFloat(tempStr, 32); err == nil && t >= 0 && t <= 2.0 {
			temp = float32(t)
		}
	}

	var maxTokens int32 = 4096
	if maxStr := os.Getenv("GEMINI_MAX_OUTPUT_TOKENS"); maxStr != "" {
		if m, err := strconv.ParseInt(maxStr, 10, 32); err == nil && m > 0 {
			maxTokens = int32(m)
		}
	}

	mimeType := os.Getenv("GEMINI_RESPONSE_MIME_TYPE")
	if mimeType == "" {
		mimeType = "application/json"
	}

	return &Config{
		GRPCPort:         port,
		GeminiAPIKey:     apiKey,
		GeminiModel:      model,
		GeminiTimeout:    time.Duration(timeoutSec) * time.Second,
		GeminiTemp:       temp,
		GeminiMaxTokens:  maxTokens,
		ResponseMIMEType: mimeType,
	}
}

func (c *Config) Validate() error {
	if c.GeminiAPIKey == "" {
		return errors.New("GEMINI_API_KEY is required but not set")
	}
	if c.GeminiModel == "" {
		return errors.New("GEMINI_MODEL is required but not set")
	}
	return nil
}
