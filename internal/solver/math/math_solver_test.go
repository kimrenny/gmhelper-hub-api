package math

import (
	"context"
	"errors"
	"strings"
	"testing"

	"gmhelper.solution-hub/internal/gemini"
	"gmhelper.solution-hub/internal/solver"
)

func TestMathSolver_Solve_Success(t *testing.T) {
	expectedResponse := `{"steps":[{"step":1,"text":"solve equation"}],"answer":"x=2"}`
	mockGemini := gemini.NewMockClient(expectedResponse, nil)
	s := NewMathSolver(mockGemini)

	task := solver.Task{
		TaskID:      "task-math-1",
		ProblemType: "math",
		Payload:     `{"data":"2x=4"}`,
		UserID:      "user-1",
	}

	result, err := s.Solve(context.Background(), task)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !result.Success {
		t.Errorf("expected Success true, got false")
	}

	if result.RawOutput != expectedResponse {
		t.Errorf("expected RawOutput '%s', got '%s'", expectedResponse, result.RawOutput)
	}

	if !strings.Contains(mockGemini.LastPrompt, "2x=4") {
		t.Errorf("prompt should contain task payload, got: %s", mockGemini.LastPrompt)
	}

	if !strings.Contains(mockGemini.LastPrompt, "math") {
		t.Errorf("prompt should identify problem type, got: %s", mockGemini.LastPrompt)
	}
}

func TestMathSolver_Solve_GeminiError(t *testing.T) {
	mockGemini := gemini.NewMockClient("", errors.New("api rate limit exceeded"))
	s := NewMathSolver(mockGemini)

	task := solver.Task{
		TaskID:      "task-math-2",
		ProblemType: "math",
		Payload:     `{"data":"x^2=9"}`,
	}

	result, err := s.Solve(context.Background(), task)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	if result.Success {
		t.Errorf("expected Success false, got true")
	}

	if !strings.Contains(err.Error(), "rate limit") {
		t.Errorf("expected rate limit error, got: %v", err)
	}
}

func TestMathSolver_Solve_EmptyPayload(t *testing.T) {
	mockGemini := gemini.NewMockClient("ok", nil)
	s := NewMathSolver(mockGemini)

	task := solver.Task{
		TaskID:      "task-math-3",
		ProblemType: "math",
		Payload:     "",
	}

	_, err := s.Solve(context.Background(), task)
	if err == nil {
		t.Fatalf("expected error for empty payload, got nil")
	}
}
