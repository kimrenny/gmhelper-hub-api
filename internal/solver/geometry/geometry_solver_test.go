package geometry

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"gmhelper.solution-hub/internal/gemini"
	"gmhelper.solution-hub/internal/solver"
)

func canvasPayload() string {
	return `{
  "triangle_1": {
    "points": [
      {"label": "A"},
      {"label": "B"},
      {"label": "C"}
    ],
    "lines": {
      "AB": 5.0,
      "BC": 5.0,
      "AC": 6.0
    },
    "angles": {}
  }
}`
}

func gmhelperWrappedPayload() string {
	return `{
  "task": {
    "triangle_1": {
      "points": [
        {"label": "A"},
        {"label": "B"},
        {"label": "C"}
      ],
      "lines": {
        "AB": 5.0,
        "BC": 5.0,
        "AC": 6.0
      },
      "angles": {}
    }
  },
  "given": "ABC – triangle\nAB = BC = 5\nAC = 6\n",
  "solution": {},
  "answer": "..."
}`
}

func TestGeometrySolver_Solve_CanvasPayload_Success(t *testing.T) {
	mockGemini := gemini.NewMockClient(validGeometryJSON(), nil)
	s := NewGeometrySolver(mockGemini)

	task := solver.Task{
		TaskID:      "task-geo-1",
		ProblemType: "geometry",
		Payload:     canvasPayload(),
		UserID:      "user-geo-1",
	}

	result, err := s.Solve(context.Background(), task)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !result.Success {
		t.Errorf("expected Success true, got false")
	}

	if result.TaskID != "task-geo-1" {
		t.Errorf("expected TaskID 'task-geo-1', got '%s'", result.TaskID)
	}

	if result.ProblemType != "geometry" {
		t.Errorf("expected ProblemType 'geometry', got '%s'", result.ProblemType)
	}

	var parsed GeometryResult
	if err := json.Unmarshal([]byte(result.RawOutput), &parsed); err != nil {
		t.Fatalf("failed to unmarshal normalized output JSON: %v", err)
	}
	if parsed.FinalAnswer != "Altitude BH = 4, Area S = 12, Perimeter P = 16" {
		t.Errorf("unexpected FinalAnswer: %s", parsed.FinalAnswer)
	}

	if !strings.Contains(mockGemini.LastPrompt, "AB = 5") {
		t.Errorf("prompt should contain extracted side lengths, got: %s", mockGemini.LastPrompt)
	}
}

func TestGeometrySolver_Solve_WrappedPayload_Success(t *testing.T) {
	mockGemini := gemini.NewMockClient(validGeometryJSON(), nil)
	s := NewGeometrySolver(mockGemini)

	task := solver.Task{
		TaskID:      "task-geo-wrapped",
		ProblemType: "geometry",
		Payload:     gmhelperWrappedPayload(),
	}

	result, err := s.Solve(context.Background(), task)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !result.Success {
		t.Errorf("expected Success true, got false")
	}
}

func TestGeometrySolver_Solve_InvalidJSON(t *testing.T) {
	mockGemini := gemini.NewMockClient("not valid json at all", nil)
	s := NewGeometrySolver(mockGemini)

	task := solver.Task{
		TaskID:      "task-geo-invalid-json",
		ProblemType: "geometry",
		Payload:     canvasPayload(),
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

func TestGeometrySolver_Solve_ParityFailure(t *testing.T) {
	// Gemini returns altered length AB = 4.0 instead of input 5.0
	alteredGemini := strings.Replace(validGeometryJSON(), `"AB": 5.0`, `"AB": 4.0`, 1)
	mockGemini := gemini.NewMockClient(alteredGemini, nil)
	s := NewGeometrySolver(mockGemini)

	task := solver.Task{
		TaskID:      "task-geo-parity-fail",
		ProblemType: "geometry",
		Payload:     canvasPayload(),
	}

	result, err := s.Solve(context.Background(), task)
	if err == nil {
		t.Fatalf("expected error for input parity failure, got nil")
	}

	if result.Success {
		t.Errorf("expected Success false, got true")
	}

	if !strings.Contains(err.Error(), "parity violation") {
		t.Errorf("expected parity violation error, got: %v", err)
	}
}

func TestGeometrySolver_Solve_GeminiError(t *testing.T) {
	mockGemini := gemini.NewMockClient("", errors.New("upstream service unavailable 503"))
	s := NewGeometrySolver(mockGemini)

	task := solver.Task{
		TaskID:      "task-geo-err",
		ProblemType: "geometry",
		Payload:     canvasPayload(),
	}

	result, err := s.Solve(context.Background(), task)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	if result.Success {
		t.Errorf("expected Success false, got true")
	}

	if !strings.Contains(err.Error(), "503") {
		t.Errorf("expected 503 error, got: %v", err)
	}
}

func TestGeometrySolver_Solve_EmptyPayload(t *testing.T) {
	mockGemini := gemini.NewMockClient(validGeometryJSON(), nil)
	s := NewGeometrySolver(mockGemini)

	task := solver.Task{
		TaskID:      "task-geo-empty",
		ProblemType: "geometry",
		Payload:     "   ",
	}

	_, err := s.Solve(context.Background(), task)
	if err == nil {
		t.Fatalf("expected error for empty payload, got nil")
	}
}

func TestGeometrySolver_Solve_ContextCancellation(t *testing.T) {
	mockGemini := gemini.NewMockClient(validGeometryJSON(), nil)
	s := NewGeometrySolver(mockGemini)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	task := solver.Task{
		TaskID:      "task-geo-cancel",
		ProblemType: "geometry",
		Payload:     canvasPayload(),
	}

	result, err := s.Solve(ctx, task)
	if err == nil {
		t.Fatalf("expected error for canceled context, got nil")
	}

	if result.Success {
		t.Errorf("expected Success false, got true")
	}
}

func TestGeometrySolver_Solve_NilClient(t *testing.T) {
	s := NewGeometrySolver(nil)

	task := solver.Task{
		TaskID:      "task-geo-nil",
		ProblemType: "geometry",
		Payload:     canvasPayload(),
	}

	_, err := s.Solve(context.Background(), task)
	if err == nil {
		t.Fatalf("expected error for nil client, got nil")
	}
}

func TestGeometrySolver_Solve_WithTargetAndConditions_Success(t *testing.T) {
	mockGemini := gemini.NewMockClient(validGeometryJSON(), nil)
	s := NewGeometrySolver(mockGemini)

	payloadWithMetadata := `{
		"triangle_1": {
			"points": [
				{"label": "A"},
				{"label": "B"},
				{"label": "C"}
			],
			"lines": {
				"AB": 5.0,
				"BC": 5.0,
				"AC": 6.0
			},
			"angles": {}
		},
		"additionalConditions": [
			"BH is perpendicular to AC"
		],
		"target": "Find BH"
	}`

	task := solver.Task{
		TaskID:      "task-geo-target-1",
		ProblemType: "geometry",
		Payload:     payloadWithMetadata,
	}

	result, err := s.Solve(context.Background(), task)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !result.Success {
		t.Errorf("expected Success true, got false")
	}

	if !strings.Contains(mockGemini.LastPrompt, "Explicit Problem Goal (AUTHORITATIVE") || !strings.Contains(mockGemini.LastPrompt, "Find BH") {
		t.Errorf("prompt should contain authoritative explicit target, got: %s", mockGemini.LastPrompt)
	}

	if !strings.Contains(mockGemini.LastPrompt, "Additional Geometric Constraints") || !strings.Contains(mockGemini.LastPrompt, "BH is perpendicular to AC") {
		t.Errorf("prompt should contain additional constraints, got: %s", mockGemini.LastPrompt)
	}

	var parsed GeometryResult
	if err := json.Unmarshal([]byte(result.RawOutput), &parsed); err != nil {
		t.Fatalf("failed to parse raw output: %v", err)
	}
	if parsed.InputFacts.ExplicitTarget != "Find BH" {
		t.Errorf("expected normalized inputFacts to have ExplicitTarget 'Find BH', got '%s'", parsed.InputFacts.ExplicitTarget)
	}
}

func TestGeometrySolver_Solve_UnrelatedTarget_ValidationFail(t *testing.T) {
	// Gemini outputs an unrelated area-only target
	unrelatedJSON := strings.Replace(
		validGeometryJSON(),
		`"target": {
    "descriptions": [
      "Altitude BH to base AC",
      "Area S of triangle ABC",
      "Perimeter P of triangle ABC"
    ],
    "variables": ["BH", "S", "P"]
  }`,
		`"target": {
    "descriptions": ["Calculate the area of triangle ABC"],
    "variables": ["area"]
  }`,
		1,
	)

	mockGemini := gemini.NewMockClient(unrelatedJSON, nil)
	s := NewGeometrySolver(mockGemini)

	payloadWithTarget := `{
		"triangle_1": {
			"points": [
				{"label": "A"},
				{"label": "B"},
				{"label": "C"}
			],
			"lines": {
				"AB": 5.0,
				"BC": 5.0,
				"AC": 6.0
			},
			"angles": {}
		},
		"target": "Find BH"
	}`

	task := solver.Task{
		TaskID:      "task-geo-unrelated-target",
		ProblemType: "geometry",
		Payload:     payloadWithTarget,
	}

	result, err := s.Solve(context.Background(), task)
	if err == nil {
		t.Fatalf("expected error due to target validation failure, got nil")
	}

	if result.Success {
		t.Errorf("expected Success false, got true")
	}

	if !strings.Contains(err.Error(), "target validation failed") {
		t.Errorf("expected error containing 'target validation failed', got: %v", err)
	}
}

func TestGeometrySolver_Solve_TextOnlyNoFigure_Success(t *testing.T) {
	mockGemini := gemini.NewMockClient(validGeometryJSON(), nil)
	s := NewGeometrySolver(mockGemini)

	textOnlyPayload := `{
		"problem": "In triangle ABC, AB = 5, BC = 5, AC = 6. Find the altitude BH to AC.",
		"additionalConditions": [
			"BH is perpendicular to AC"
		],
		"target": "Find BH",
		"language": "en"
	}`

	task := solver.Task{
		TaskID:      "task-geo-text-only",
		ProblemType: "geometry",
		Payload:     textOnlyPayload,
	}

	result, err := s.Solve(context.Background(), task)
	if err != nil {
		t.Fatalf("unexpected error for text-only geometry task: %v", err)
	}

	if !result.Success {
		t.Errorf("expected Success true, got false")
	}

	if !strings.Contains(mockGemini.LastPrompt, "Diagram: None supplied") {
		t.Errorf("prompt should indicate no diagram was supplied, got: %s", mockGemini.LastPrompt)
	}
	if !strings.Contains(mockGemini.LastPrompt, "Find BH") {
		t.Errorf("prompt should contain target 'Find BH'")
	}

	var parsed GeometryResult
	if err := json.Unmarshal([]byte(result.RawOutput), &parsed); err != nil {
		t.Fatalf("failed to unmarshal normalized output JSON: %v", err)
	}
	if parsed.FinalAnswer != "Altitude BH = 4, Area S = 12, Perimeter P = 16" {
		t.Errorf("unexpected FinalAnswer: %s", parsed.FinalAnswer)
	}
}

func TestGeometrySolver_Solve_TextOverridesDrawing_Success(t *testing.T) {
	mockGemini := gemini.NewMockClient(validGeometryJSON(), nil)
	s := NewGeometrySolver(mockGemini)

	conflictingPayload := `{
		"triangle_1": {
			"points": [
				{"label": "A", "x": 100, "y": 100},
				{"label": "B", "x": 500, "y": 100},
				{"label": "C", "x": 300, "y": 50}
			],
			"lines": {
				"AB": 7.0,
				"BC": 7.0,
				"AC": 8.0
			}
		},
		"problem": "In triangle ABC, AB = 5, BC = 5, AC = 6.",
		"additionalConditions": [
			"BH is perpendicular to AC"
		],
		"target": "Find BH"
	}`

	task := solver.Task{
		TaskID:      "task-geo-conflict",
		ProblemType: "geometry",
		Payload:     conflictingPayload,
	}

	result, err := s.Solve(context.Background(), task)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !result.Success {
		t.Errorf("expected Success true, got false")
	}

	// Verify that prompt uses AB = 5 and NOT AB = 7
	if !strings.Contains(mockGemini.LastPrompt, "AB = 5") {
		t.Errorf("prompt should contain authoritative textual length AB = 5, got: %s", mockGemini.LastPrompt)
	}
	if strings.Contains(mockGemini.LastPrompt, "AB = 7") {
		t.Errorf("prompt should NOT contain drawing length AB = 7")
	}
}
