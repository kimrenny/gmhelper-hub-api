package geometry

import (
	"context"
	"errors"
	"strings"
	"testing"

	"gmhelper.solution-hub/internal/gemini"
	"gmhelper.solution-hub/internal/solver"
)

func TestGeometrySolver_Solve_Success(t *testing.T) {
	expectedResponse := `{"steps":[{"step":1,"text":"find triangle area"}],"answer":"S=25"}`
	mockGemini := gemini.NewMockClient(expectedResponse, nil)
	s := NewGeometrySolver(mockGemini)

	task := solver.Task{
		TaskID:      "task-geo-1",
		ProblemType: "geometry",
		Payload:     `{"triangle":{"lines":{"AB":5,"BC":10}}}`,
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

	if !strings.Contains(mockGemini.LastPrompt, "AB") {
		t.Errorf("prompt should contain task payload, got: %s", mockGemini.LastPrompt)
	}

	if !strings.Contains(mockGemini.LastPrompt, "geometry") {
		t.Errorf("prompt should identify problem type, got: %s", mockGemini.LastPrompt)
	}
}

func TestGeometrySolver_Solve_GeminiError(t *testing.T) {
	mockGemini := gemini.NewMockClient("", errors.New("timeout"))
	s := NewGeometrySolver(mockGemini)

	task := solver.Task{
		TaskID:      "task-geo-2",
		ProblemType: "geometry",
		Payload:     `{"circle":{}}`,
	}

	result, err := s.Solve(context.Background(), task)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	if result.Success {
		t.Errorf("expected Success false, got true")
	}

	if !strings.Contains(err.Error(), "timeout") {
		t.Errorf("expected timeout error, got: %v", err)
	}
}
