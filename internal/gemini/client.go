package gemini

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"google.golang.org/genai"

	"gmhelper.solution-hub/internal/config"
)

type Client interface {
	Generate(ctx context.Context, prompt string) (string, error)
}

type GenAIClient struct {
	client *genai.Client
	cfg    *config.Config
}

func NewGenAIClient(ctx context.Context, cfg *config.Config) (*GenAIClient, error) {
	if cfg == nil {
		return nil, errors.New("gemini config cannot be nil")
	}
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid gemini config: %w", err)
	}

	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey: cfg.GeminiAPIKey,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create gemini client: %w", err)
	}

	return &GenAIClient{
		client: client,
		cfg:    cfg,
	}, nil
}

func (c *GenAIClient) Generate(ctx context.Context, prompt string) (string, error) {
	if strings.TrimSpace(prompt) == "" {
		return "", errors.New("prompt cannot be empty")
	}

	reqCtx, cancel := context.WithTimeout(ctx, c.cfg.GeminiTimeout)
	defer cancel()

	genConfig := &genai.GenerateContentConfig{
		Temperature:      &c.cfg.GeminiTemp,
		MaxOutputTokens:  c.cfg.GeminiMaxTokens,
		ResponseMIMEType: c.cfg.ResponseMIMEType,
	}

	resp, err := c.client.Models.GenerateContent(reqCtx, c.cfg.GeminiModel, genai.Text(prompt), genConfig)
	if err != nil {
		if errors.Is(reqCtx.Err(), context.DeadlineExceeded) {
			return "", fmt.Errorf("gemini request timed out after %v", c.cfg.GeminiTimeout)
		}
		if errors.Is(reqCtx.Err(), context.Canceled) {
			return "", errors.New("gemini request canceled")
		}
		return "", fmt.Errorf("gemini api error: %w", err)
	}

	if resp == nil {
		return "", errors.New("empty response received from gemini")
	}

	text := resp.Text()
	if text == "" {
		return "", errors.New("no content text returned by gemini model")
	}

	return text, nil
}
