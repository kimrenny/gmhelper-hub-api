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

func createTestHandler(mockGemini gemini.Client) *SolveHandler {
	m := mathSolver.NewMathSolver(mockGemini)
	g := geoSolver.NewGeometrySolver(mockGemini)
	p := planner.NewDefaultPlanner(m, g)
	return NewSolveHandler(p)
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
	geminiResponse := `{"steps":[{"step":1,"text":"calculate area"}],"finalAnswer":"50"}`
	mockGemini := gemini.NewMockClient(geminiResponse, nil)
	h := createTestHandler(mockGemini)

	req := &pb.SolveProblemRequest{
		TaskId:      "task-geo-200",
		ProblemType: "geometry",
		Payload:     `{"rect":{"width":5,"height":10}}`,
		UserId:      "user-200",
	}

	resp, err := h.SolveProblem(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected gRPC error: %v", err)
	}

	if !resp.Success {
		t.Errorf("expected Success true, got false")
	}
	if resp.Result != geminiResponse {
		t.Errorf("expected Result '%s', got '%s'", geminiResponse, resp.Result)
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
