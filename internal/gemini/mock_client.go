package gemini

import (
	"context"
	"errors"
)

type MockClient struct {
	GenerateFunc func(ctx context.Context, prompt string) (string, error)
	LastPrompt   string
	CallCount    int
}

func NewMockClient(response string, err error) *MockClient {
	return &MockClient{
		GenerateFunc: func(ctx context.Context, prompt string) (string, error) {
			if err != nil {
				return "", err
			}
			return response, nil
		},
	}
}

func (m *MockClient) Generate(ctx context.Context, prompt string) (string, error) {
	m.CallCount++
	m.LastPrompt = prompt
	if m.GenerateFunc != nil {
		return m.GenerateFunc(ctx, prompt)
	}
	return "", errors.New("mock GenerateFunc not defined")
}
