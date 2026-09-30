package gemini

import (
	"context"
	"errors"
	"testing"
)

func TestMockClient_Generate(t *testing.T) {
	mock := NewMockClient(`{"result": "success"}`, nil)

	ctx := context.Background()
	res, err := mock.Generate(ctx, "calculate 2+2")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if res != `{"result": "success"}` {
		t.Errorf("unexpected result: %s", res)
	}

	if mock.CallCount != 1 {
		t.Errorf("expected 1 call, got %d", mock.CallCount)
	}

	if mock.LastPrompt != "calculate 2+2" {
		t.Errorf("expected prompt 'calculate 2+2', got '%s'", mock.LastPrompt)
	}
}

func TestMockClient_Error(t *testing.T) {
	mock := NewMockClient("", errors.New("network error"))

	ctx := context.Background()
	_, err := mock.Generate(ctx, "calculate 2+2")
	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	if err.Error() != "network error" {
		t.Errorf("expected 'network error', got '%v'", err)
	}
}
