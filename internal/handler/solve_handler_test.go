package handler

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"gmhelper.solution-hub/internal/gemini"
	"gmhelper.solution-hub/internal/planner"
	geoSolver "gmhelper.solution-hub/internal/solver/geometry"
	mathSolver "gmhelper.solution-hub/internal/solver/math"
	pb "gmhelper.solution-hub/proto"
)

func validMathJSON() string {
	return `{
  "problemType": "math",
  "status": "completed",
  "problem": "2x + 5 = 15",
  "latexProblem": "2x + 5 = 15",
  "steps": [
    {
      "stepNumber": 1,
      "title": "Subtract 5 from both sides",
      "explanation": "Subtract 5 from both sides to isolate the linear term.",
      "latexFormula": "2x + 5 - 5 = 15 - 5 \\implies 2x = 10"
    },
    {
      "stepNumber": 2,
      "title": "Divide by 2",
      "explanation": "Divide both sides by 2 to find x.",
      "latexFormula": "\\frac{2x}{2} = \\frac{10}{2} \\implies x = 5"
    }
  ],
  "finalAnswer": "x = 5",
  "latexAnswer": "x = 5",
  "compositeLatex": "2x + 5 = 15 \\\\\n2x = 10 \\\\\nx = 5"
}`
}

func validGeometryJSON() string {
	return `{
  "problemType": "geometry",
  "status": "completed",
  "problemStatement": "In isosceles triangle ABC with side lengths AB = 5, BC = 5, and base AC = 6, find the altitude BH to base AC, the area S, and the perimeter P.",
  "inputFacts": {
    "figures": [
      {
        "id": "triangle_1",
        "type": "triangle",
        "vertices": ["A", "B", "C"]
      }
    ],
    "lengths": {
      "AB": 5.0,
      "BC": 5.0,
      "AC": 6.0
    },
    "angles": {}
  },
  "target": {
    "descriptions": [
      "Altitude BH to base AC",
      "Area S of triangle ABC",
      "Perimeter P of triangle ABC"
    ],
    "variables": ["BH", "S", "P"]
  },
  "derivedFacts": {
    "auxiliaryConstructions": [
      {
        "type": "altitude",
        "label": "BH",
        "fromVertex": "B",
        "toSegment": "AC",
        "footPoint": "H"
      }
    ],
    "lengths": {
      "AH": 3.0,
      "HC": 3.0,
      "BH": 4.0
    },
    "angles": {
      "BAC": 53.13,
      "BCA": 53.13,
      "ABC": 73.74,
      "AHB": 90.0,
      "BHC": 90.0
    },
    "metrics": {
      "perimeter": 16.0,
      "area": 12.0
    }
  },
  "steps": [
    {
      "stepNumber": 1,
      "title": "Construct Altitude and Determine Segment Lengths",
      "explanation": "Construct altitude BH perpendicular to base AC. In isosceles triangle ABC with AB = BC, altitude BH bisects base AC. Thus, H is the midpoint of AC.",
      "latexFormula": "AH = HC = \\frac{AC}{2} = \\frac{6}{2} = 3"
    },
    {
      "stepNumber": 2,
      "title": "Apply Pythagorean Theorem in Right Triangle ABH",
      "explanation": "In right-angled triangle ABH, by the Pythagorean theorem, the square of hypotenuse AB equals the sum of the squares of legs AH and BH.",
      "latexFormula": "AB^2 = AH^2 + BH^2 \\implies 5^2 = 3^2 + BH^2 \\implies BH = \\sqrt{25 - 9} = 4"
    },
    {
      "stepNumber": 3,
      "title": "Calculate Area of Triangle ABC",
      "explanation": "The area S of a triangle is half the product of its base and corresponding altitude.",
      "latexFormula": "S = \\frac{1}{2} \\cdot AC \\cdot BH = \\frac{1}{2} \\cdot 6 \\cdot 4 = 12"
    },
    {
      "stepNumber": 4,
      "title": "Calculate Perimeter of Triangle ABC",
      "explanation": "The perimeter P is the sum of all three side lengths.",
      "latexFormula": "P = AB + BC + AC = 5 + 5 + 6 = 16"
    }
  ],
  "finalAnswer": "Altitude BH = 4, Area S = 12, Perimeter P = 16",
  "latexAnswer": "BH = 4, \\quad S = 12, \\quad P = 16"
}`
}

func validGeometryCanvasPayload() string {
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

func createTestHandler(mockGemini gemini.Client) *SolveHandler {
	m := mathSolver.NewMathSolver(mockGemini)
	g := geoSolver.NewGeometrySolver(mockGemini)
	p := planner.NewDefaultPlanner(m, g)
	return NewSolveHandler(p)
}

func TestSolveHandler_MathEndToEnd_ExactMock(t *testing.T) {
	geminiResponse := validMathJSON()
	mockGemini := gemini.NewMockClient(geminiResponse, nil)
	h := createTestHandler(mockGemini)

	req := &pb.SolveProblemRequest{
		TaskId:      "test-math-1",
		ProblemType: "math",
		Payload:     "2x + 5 = 15",
		UserId:      "test-user",
	}

	resp, err := h.SolveProblem(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected gRPC error: %v", err)
	}

	if resp.TaskId != "test-math-1" {
		t.Errorf("expected TaskId 'test-math-1', got '%s'", resp.TaskId)
	}
	if !resp.Success {
		t.Errorf("expected Success true, got false")
	}
	if resp.Status != "SUCCESS" {
		t.Errorf("expected Status 'SUCCESS', got '%s'", resp.Status)
	}

	// 1. Verify prompt was received by mock Gemini
	if !strings.Contains(mockGemini.LastPrompt, "2x + 5 = 15") {
		t.Errorf("expected prompt to contain math problem, got: %s", mockGemini.LastPrompt)
	}

	// 2. Verify returned result is valid canonical JSON
	var parsed mathSolver.MathResult
	if err := json.Unmarshal([]byte(resp.Result), &parsed); err != nil {
		t.Fatalf("expected Result to be valid JSON: %v", err)
	}
	if parsed.FinalAnswer != "x = 5" {
		t.Errorf("expected FinalAnswer 'x = 5', got '%s'", parsed.FinalAnswer)
	}
	if parsed.LatexAnswer != "x = 5" {
		t.Errorf("expected LatexAnswer 'x = 5', got '%s'", parsed.LatexAnswer)
	}
	if len(parsed.Steps) != 2 {
		t.Errorf("expected 2 steps, got %d", len(parsed.Steps))
	}
}

func TestSolveHandler_GeometryEndToEnd_ExactMock(t *testing.T) {
	geminiResponse := validGeometryJSON()
	mockGemini := gemini.NewMockClient(geminiResponse, nil)
	h := createTestHandler(mockGemini)

	req := &pb.SolveProblemRequest{
		TaskId:      "test-geometry-1",
		ProblemType: "geometry",
		Payload:     validGeometryCanvasPayload(),
		UserId:      "test-user",
	}

	resp, err := h.SolveProblem(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected gRPC error: %v", err)
	}

	if resp.TaskId != "test-geometry-1" {
		t.Errorf("expected TaskId 'test-geometry-1', got '%s'", resp.TaskId)
	}
	if !resp.Success {
		t.Errorf("expected Success true, got false")
	}
	if resp.Status != "SUCCESS" {
		t.Errorf("expected Status 'SUCCESS', got '%s'", resp.Status)
	}

	// 1. Verify prompt contains original facts
	if !strings.Contains(mockGemini.LastPrompt, "AB = 5") {
		t.Errorf("expected prompt to contain side lengths, got: %s", mockGemini.LastPrompt)
	}

	// 2. Verify returned result is valid canonical JSON
	var parsed geoSolver.GeometryResult
	if err := json.Unmarshal([]byte(resp.Result), &parsed); err != nil {
		t.Fatalf("expected Result to be valid JSON: %v", err)
	}
	if parsed.FinalAnswer != "Altitude BH = 4, Area S = 12, Perimeter P = 16" {
		t.Errorf("unexpected FinalAnswer: %s", parsed.FinalAnswer)
	}
	if parsed.DerivedFacts.Metrics["area"] != 12.0 {
		t.Errorf("expected area 12.0, got %v", parsed.DerivedFacts.Metrics["area"])
	}
	if parsed.DerivedFacts.Metrics["perimeter"] != 16.0 {
		t.Errorf("expected perimeter 16.0, got %v", parsed.DerivedFacts.Metrics["perimeter"])
	}
}

func TestSolveHandler_MathSuccess(t *testing.T) {
	geminiResponse := validMathJSON()
	mockGemini := gemini.NewMockClient(geminiResponse, nil)
	h := createTestHandler(mockGemini)

	req := &pb.SolveProblemRequest{
		TaskId:      "task-math-100",
		ProblemType: "math",
		Payload:     `{"data":"2x + 5 = 15"}`,
		UserId:      "user-100",
	}

	resp, err := h.SolveProblem(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected gRPC error: %v", err)
	}

	if resp.TaskId != "task-math-100" {
		t.Errorf("expected TaskId 'task-math-100', got '%s'", resp.TaskId)
	}
	if !resp.Success {
		t.Errorf("expected Success true, got false")
	}
	if resp.Status != "SUCCESS" {
		t.Errorf("expected Status 'SUCCESS', got '%s'", resp.Status)
	}

	var parsed mathSolver.MathResult
	if err := json.Unmarshal([]byte(resp.Result), &parsed); err != nil {
		t.Fatalf("expected Result to be valid JSON: %v", err)
	}
	if parsed.FinalAnswer != "x = 5" {
		t.Errorf("expected FinalAnswer 'x = 5', got '%s'", parsed.FinalAnswer)
	}
}

func TestSolveHandler_MathInvalidGeminiOutput(t *testing.T) {
	mockGemini := gemini.NewMockClient("not valid json", nil)
	h := createTestHandler(mockGemini)

	req := &pb.SolveProblemRequest{
		TaskId:      "task-math-invalid",
		ProblemType: "math",
		Payload:     `{"data":"2x + 5 = 15"}`,
	}

	resp, err := h.SolveProblem(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected gRPC error: %v", err)
	}

	if resp.Success {
		t.Errorf("expected Success false on invalid gemini output, got true")
	}
	if resp.Status != "SOLVER_ERROR" {
		t.Errorf("expected Status 'SOLVER_ERROR', got '%s'", resp.Status)
	}
	if !strings.Contains(resp.Result, "output validation failed") {
		t.Errorf("expected validation failure error in result, got: %s", resp.Result)
	}
}

func TestSolveHandler_GeometrySuccess(t *testing.T) {
	geminiResponse := validGeometryJSON()
	mockGemini := gemini.NewMockClient(geminiResponse, nil)
	h := createTestHandler(mockGemini)

	req := &pb.SolveProblemRequest{
		TaskId:      "task-geo-200",
		ProblemType: "geometry",
		Payload:     validGeometryCanvasPayload(),
		UserId:      "user-200",
	}

	resp, err := h.SolveProblem(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected gRPC error: %v", err)
	}

	if resp.TaskId != "task-geo-200" {
		t.Errorf("expected TaskId 'task-geo-200', got '%s'", resp.TaskId)
	}
	if !resp.Success {
		t.Errorf("expected Success true, got false")
	}
	if resp.Status != "SUCCESS" {
		t.Errorf("expected Status 'SUCCESS', got '%s'", resp.Status)
	}

	var parsed geoSolver.GeometryResult
	if err := json.Unmarshal([]byte(resp.Result), &parsed); err != nil {
		t.Fatalf("expected Result to be valid JSON: %v", err)
	}
	if parsed.FinalAnswer != "Altitude BH = 4, Area S = 12, Perimeter P = 16" {
		t.Errorf("unexpected FinalAnswer: %s", parsed.FinalAnswer)
	}
}

func TestSolveHandler_GeometryInvalidGeminiOutput(t *testing.T) {
	mockGemini := gemini.NewMockClient("not valid json at all", nil)
	h := createTestHandler(mockGemini)

	req := &pb.SolveProblemRequest{
		TaskId:      "task-geo-invalid",
		ProblemType: "geometry",
		Payload:     validGeometryCanvasPayload(),
	}

	resp, err := h.SolveProblem(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected gRPC error: %v", err)
	}

	if resp.Success {
		t.Errorf("expected Success false on invalid geometry output, got true")
	}
	if resp.Status != "SOLVER_ERROR" {
		t.Errorf("expected Status 'SOLVER_ERROR', got '%s'", resp.Status)
	}
	if !strings.Contains(resp.Result, "output validation failed") {
		t.Errorf("expected validation failure in result, got: %s", resp.Result)
	}
}

func TestSolveHandler_GeometryParityFailure(t *testing.T) {
	alteredGemini := strings.Replace(validGeometryJSON(), `"AB": 5.0`, `"AB": 4.0`, 1)
	mockGemini := gemini.NewMockClient(alteredGemini, nil)
	h := createTestHandler(mockGemini)

	req := &pb.SolveProblemRequest{
		TaskId:      "task-geo-parity",
		ProblemType: "geometry",
		Payload:     validGeometryCanvasPayload(),
	}

	resp, err := h.SolveProblem(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected gRPC error: %v", err)
	}

	if resp.Success {
		t.Errorf("expected Success false on parity failure, got true")
	}
	if resp.Status != "SOLVER_ERROR" {
		t.Errorf("expected Status 'SOLVER_ERROR', got '%s'", resp.Status)
	}
	if !strings.Contains(resp.Result, "parity violation") {
		t.Errorf("expected parity violation in result, got: %s", resp.Result)
	}
}

func TestSolveHandler_UnsupportedProblemType(t *testing.T) {
	mockGemini := gemini.NewMockClient("ok", nil)
	h := createTestHandler(mockGemini)

	req := &pb.SolveProblemRequest{
		TaskId:      "task-unknown-300",
		ProblemType: "quantum_physics",
		Payload:     `{"spin":"1/2"}`,
	}

	resp, err := h.SolveProblem(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.Success {
		t.Errorf("expected Success false for unsupported type, got true")
	}
	if resp.Status != "UNSUPPORTED_PROBLEM_TYPE" {
		t.Errorf("expected Status 'UNSUPPORTED_PROBLEM_TYPE', got '%s'", resp.Status)
	}
	if mockGemini.CallCount != 0 {
		t.Errorf("Gemini should not be called for unsupported problem types, got %d calls", mockGemini.CallCount)
	}
}

func TestSolveHandler_GeminiFailure(t *testing.T) {
	mockGemini := gemini.NewMockClient("", errors.New("gemini upstream 503 unavailable"))
	h := createTestHandler(mockGemini)

	req := &pb.SolveProblemRequest{
		TaskId:      "task-math-err",
		ProblemType: "math",
		Payload:     `{"data":"1/0"}`,
	}

	resp, err := h.SolveProblem(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected gRPC error: %v", err)
	}

	if resp.Success {
		t.Errorf("expected Success false on Gemini error, got true")
	}
	if resp.Status != "SOLVER_ERROR" {
		t.Errorf("expected Status 'SOLVER_ERROR', got '%s'", resp.Status)
	}
	if !strings.Contains(resp.Result, "503") {
		t.Errorf("expected error message to contain 503, got '%s'", resp.Result)
	}
}

func TestSolveHandler_ValidationErrors(t *testing.T) {
	mockGemini := gemini.NewMockClient("ok", nil)
	h := createTestHandler(mockGemini)

	// Nil request
	resp, err := h.SolveProblem(context.Background(), nil)
	if err != nil || resp.Success || resp.Status != "INVALID_ARGUMENT" {
		t.Errorf("expected INVALID_ARGUMENT for nil request, got %v, %v", resp, err)
	}

	// Empty TaskId
	resp, err = h.SolveProblem(context.Background(), &pb.SolveProblemRequest{TaskId: "", Payload: "1+1"})
	if err != nil || resp.Success || resp.Status != "INVALID_ARGUMENT" {
		t.Errorf("expected INVALID_ARGUMENT for empty TaskId, got %v, %v", resp, err)
	}

	// Empty Payload
	resp, err = h.SolveProblem(context.Background(), &pb.SolveProblemRequest{TaskId: "123", Payload: ""})
	if err != nil || resp.Success || resp.Status != "INVALID_ARGUMENT" {
		t.Errorf("expected INVALID_ARGUMENT for empty Payload, got %v, %v", resp, err)
	}
}

func TestSolveHandler_ContextCancellation(t *testing.T) {
	mockGemini := gemini.NewMockClient("", context.Canceled)
	h := createTestHandler(mockGemini)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	req := &pb.SolveProblemRequest{
		TaskId:      "task-cancel",
		ProblemType: "math",
		Payload:     `{"data":"x=1"}`,
	}

	resp, err := h.SolveProblem(ctx, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.Success {
		t.Errorf("expected Success false on context cancellation, got true")
	}
}
