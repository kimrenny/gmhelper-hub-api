package math

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"gmhelper.solution-hub/internal/gemini"
	"gmhelper.solution-hub/internal/solver"
)

func TestMathSolver_Solve_Success(t *testing.T) {
	mockGemini := gemini.NewMockClient(validMathJSON(), nil)
	s := NewMathSolver(mockGemini)

	task := solver.Task{
		TaskID:      "task-math-1",
		ProblemType: "math",
		Payload:     `{"data":"2x + 5 = 15"}`,
		UserID:      "user-1",
	}

	result, err := s.Solve(context.Background(), task)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !result.Success {
		t.Errorf("expected Success true, got false")
	}

	if result.TaskID != "task-math-1" {
		t.Errorf("expected TaskID 'task-math-1', got '%s'", result.TaskID)
	}

	if result.ProblemType != "math" {
		t.Errorf("expected ProblemType 'math', got '%s'", result.ProblemType)
	}

	// Verify RawOutput is valid normalized JSON
	var parsed MathResult
	if err := json.Unmarshal([]byte(result.RawOutput), &parsed); err != nil {
		t.Fatalf("failed to unmarshal normalized output JSON: %v", err)
	}
	if parsed.FinalAnswer != "x = 5" {
		t.Errorf("expected FinalAnswer 'x = 5', got '%s'", parsed.FinalAnswer)
	}

	if !strings.Contains(mockGemini.LastPrompt, "2x + 5 = 15") {
		t.Errorf("prompt should contain extracted problem, got: %s", mockGemini.LastPrompt)
	}
}

func TestMathSolver_Solve_InvalidJSON(t *testing.T) {
	mockGemini := gemini.NewMockClient("not valid json at all", nil)
	s := NewMathSolver(mockGemini)

	task := solver.Task{
		TaskID:      "task-math-invalid-json",
		ProblemType: "math",
		Payload:     `2x + 5 = 15`,
	}

	result, err := s.Solve(context.Background(), task)
	if err == nil {
		t.Fatalf("expected error on invalid JSON, got nil")
	}

	if result.Success {
		t.Errorf("expected Success false, got true")
	}

	if !strings.Contains(err.Error(), "output validation failed") {
		t.Errorf("expected validation failure message, got: %v", err)
	}
}

func TestMathSolver_Solve_GeminiError(t *testing.T) {
	mockGemini := gemini.NewMockClient("", errors.New("api rate limit exceeded"))
	s := NewMathSolver(mockGemini)

	task := solver.Task{
		TaskID:      "task-math-err",
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
	mockGemini := gemini.NewMockClient(validMathJSON(), nil)
	s := NewMathSolver(mockGemini)

	task := solver.Task{
		TaskID:      "task-math-empty",
		ProblemType: "math",
		Payload:     "   ",
	}

	_, err := s.Solve(context.Background(), task)
	if err == nil {
		t.Fatalf("expected error for empty payload, got nil")
	}
}

func TestMathSolver_Solve_ContextCancellation(t *testing.T) {
	mockGemini := gemini.NewMockClient(validMathJSON(), nil)
	s := NewMathSolver(mockGemini)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	task := solver.Task{
		TaskID:      "task-math-cancel",
		ProblemType: "math",
		Payload:     `2x = 4`,
	}

	result, err := s.Solve(ctx, task)
	if err == nil {
		t.Fatalf("expected error for canceled context, got nil")
	}

	if result.Success {
		t.Errorf("expected Success false, got true")
	}
}

func TestMathSolver_Solve_NilClient(t *testing.T) {
	s := NewMathSolver(nil)

	task := solver.Task{
		TaskID:      "task-math-nil",
		ProblemType: "math",
		Payload:     `2x = 4`,
	}

	_, err := s.Solve(context.Background(), task)
	if err == nil {
		t.Fatalf("expected error for nil client, got nil")
	}
}

func TestExtractProblemAndLanguage(t *testing.T) {
	testCases := []struct {
		name         string
		payload      string
		expectedProb string
		expectedLang string
	}{
		{
			name:         "json with data and language ru",
			payload:      `{"data":"2x + 5 = 15", "language":"ru"}`,
			expectedProb: "2x + 5 = 15",
			expectedLang: "ru",
		},
		{
			name:         "json with problem and lang uk",
			payload:      `{"problem":"\\int x dx", "lang":"uk"}`,
			expectedProb: "\\int x dx",
			expectedLang: "uk",
		},
		{
			name:         "json with expression and locale de",
			payload:      `{"expression":"3a + 2b = 10", "locale":"de"}`,
			expectedProb: "3a + 2b = 10",
			expectedLang: "de",
		},
		{
			name:         "plain string defaults to en",
			payload:      `2x + 5 = 15`,
			expectedProb: "2x + 5 = 15",
			expectedLang: "en",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			prob, lang := ExtractProblemAndLanguage(tc.payload)
			if prob != tc.expectedProb {
				t.Errorf("expected problem '%s', got '%s'", tc.expectedProb, prob)
			}
			if lang != tc.expectedLang {
				t.Errorf("expected language '%s', got '%s'", tc.expectedLang, lang)
			}
		})
	}
}

func TestMathSolver_Solve_LanguagePropagation(t *testing.T) {
	mockGemini := gemini.NewMockClient(validMathJSON(), nil)
	s := NewMathSolver(mockGemini)

	task := solver.Task{
		TaskID:      "task-math-lang-ru",
		ProblemType: "math",
		Payload:     `{"data":"2x + 5 = 15", "language":"ru"}`,
	}

	result, err := s.Solve(context.Background(), task)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Success {
		t.Errorf("expected Success true, got false")
	}

	if !strings.Contains(mockGemini.LastPrompt, "Russian (Русский)") {
		t.Errorf("expected prompt to contain 'Russian (Русский)', got:\n%s", mockGemini.LastPrompt)
	}
}
