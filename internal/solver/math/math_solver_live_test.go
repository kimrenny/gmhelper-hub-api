package math

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"testing"
	"time"

	"gmhelper.solution-hub/internal/config"
	"gmhelper.solution-hub/internal/gemini"
	"gmhelper.solution-hub/internal/solver"
)

func TestMathSolver_LiveGeminiSmokeTest(t *testing.T) {
	if os.Getenv("RUN_LIVE_GEMINI_TESTS") != "1" && os.Getenv("RUN_LIVE_GEMINI_TESTS") != "true" {
		t.Skip("skipping live Gemini test: set RUN_LIVE_GEMINI_TESTS=1 to enable")
	}

	cfg := config.Load()
	if cfg.GeminiAPIKey == "" || cfg.GeminiAPIKey == "your_gemini_api_key_here" {
		t.Skip("skipping live test: GEMINI_API_KEY is not set in environment or .env")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()

	client, err := gemini.NewGenAIClient(ctx, cfg)
	if err != nil {
		t.Fatalf("failed to initialize real Gemini client: %v", err)
	}

	s := NewMathSolver(client)

	task := solver.Task{
		TaskID:      "live-smoke-test-1",
		ProblemType: "math",
		Payload:     `{"data": "2x + 5 = 15"}`,
		UserID:      "smoke-test-user",
	}

	result, err := s.Solve(ctx, task)
	if err != nil {
		t.Fatalf("real Gemini solver request failed: %v", err)
	}

	if !result.Success {
		t.Fatalf("expected solver success, got false")
	}

	var mathResult MathResult
	if err := json.Unmarshal([]byte(result.RawOutput), &mathResult); err != nil {
		t.Fatalf("failed to parse validated solver JSON output: %v", err)
	}

	if mathResult.ProblemType != "math" {
		t.Errorf("expected problemType 'math', got '%s'", mathResult.ProblemType)
	}

	if mathResult.Status != "completed" {
		t.Errorf("expected status 'completed', got '%s'", mathResult.Status)
	}

	if strings.TrimSpace(mathResult.CompositeLatex) == "" {
		t.Errorf("expected non-empty compositeLatex")
	}

	if !strings.Contains(mathResult.FinalAnswer, "5") {
		t.Errorf("expected final answer to contain '5', got: '%s'", mathResult.FinalAnswer)
	}

	if !strings.Contains(mathResult.LatexAnswer, "5") {
		t.Errorf("expected LaTeX answer to contain '5', got: '%s'", mathResult.LatexAnswer)
	}

	if len(mathResult.Steps) == 0 {
		t.Errorf("expected at least 1 step, got 0")
	}

	// Verify no HTML tags in any textual or formula fields
	if htmlTagRegex.MatchString(result.RawOutput) {
		t.Errorf("expected no HTML tags in structured solution output, got: %s", result.RawOutput)
	}

	t.Logf("Live Gemini test succeeded!")
	t.Logf("Model used: %s", cfg.GeminiModel)
	t.Logf("Final Answer: %s", mathResult.FinalAnswer)
	t.Logf("LaTeX Answer: %s", mathResult.LatexAnswer)
	t.Logf("Composite LaTeX: %s", mathResult.CompositeLatex)
	t.Logf("Steps count: %d", len(mathResult.Steps))
}

func TestMathSolver_LiveSolve_Russian(t *testing.T) {
	if os.Getenv("RUN_LIVE_GEMINI_TESTS") != "1" && os.Getenv("RUN_LIVE_GEMINI_TESTS") != "true" {
		t.Skip("skipping live Gemini test: set RUN_LIVE_GEMINI_TESTS=1 to enable")
	}

	cfg := config.Load()
	if cfg.GeminiAPIKey == "" || cfg.GeminiAPIKey == "your_gemini_api_key_here" {
		t.Skip("skipping live test: GEMINI_API_KEY is not set in environment or .env")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()

	client, err := gemini.NewGenAIClient(ctx, cfg)
	if err != nil {
		t.Fatalf("failed to initialize real Gemini client: %v", err)
	}

	s := NewMathSolver(client)

	task := solver.Task{
		TaskID:      "live-math-ru-1",
		ProblemType: "math",
		Payload:     `{"data": "3x + 7 = 22", "language": "ru"}`,
		UserID:      "test-user-ru",
	}

	result, err := s.Solve(ctx, task)
	if err != nil {
		t.Fatalf("real Gemini solver request failed: %v", err)
	}

	if !result.Success {
		t.Fatalf("expected solver success, got false")
	}

	var mathResult MathResult
	if err := json.Unmarshal([]byte(result.RawOutput), &mathResult); err != nil {
		t.Fatalf("failed to parse validated solver JSON output: %v", err)
	}

	if !strings.Contains(mathResult.FinalAnswer, "5") {
		t.Errorf("expected final answer to contain '5', got: '%s'", mathResult.FinalAnswer)
	}

	if len(mathResult.Steps) == 0 {
		t.Fatalf("expected steps, got 0")
	}

	// Verify Russian text is generated
	hasRussian := detectRussian(mathResult.Steps[0].Title) || detectRussian(mathResult.Steps[0].Explanation) || detectRussian(mathResult.FinalAnswer)
	if !hasRussian {
		t.Errorf("expected Russian text in steps or final answer, got: title=%s, explanation=%s, finalAnswer=%s", mathResult.Steps[0].Title, mathResult.Steps[0].Explanation, mathResult.FinalAnswer)
	}
}

func TestMathSolver_LiveSolve_Ukrainian(t *testing.T) {
	if os.Getenv("RUN_LIVE_GEMINI_TESTS") != "1" && os.Getenv("RUN_LIVE_GEMINI_TESTS") != "true" {
		t.Skip("skipping live Gemini test: set RUN_LIVE_GEMINI_TESTS=1 to enable")
	}

	cfg := config.Load()
	if cfg.GeminiAPIKey == "" || cfg.GeminiAPIKey == "your_gemini_api_key_here" {
		t.Skip("skipping live test: GEMINI_API_KEY is not set in environment or .env")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()

	client, err := gemini.NewGenAIClient(ctx, cfg)
	if err != nil {
		t.Fatalf("failed to initialize real Gemini client: %v", err)
	}

	s := NewMathSolver(client)

	task := solver.Task{
		TaskID:      "live-math-uk-1",
		ProblemType: "math",
		Payload:     `{"data": "3x + 7 = 22", "language": "uk"}`,
		UserID:      "test-user-uk",
	}

	result, err := s.Solve(ctx, task)
	if err != nil {
		t.Fatalf("real Gemini solver request failed: %v", err)
	}

	if !result.Success {
		t.Fatalf("expected solver success, got false")
	}

	var mathResult MathResult
	if err := json.Unmarshal([]byte(result.RawOutput), &mathResult); err != nil {
		t.Fatalf("failed to parse validated solver JSON output: %v", err)
	}

	if !strings.Contains(mathResult.FinalAnswer, "5") {
		t.Errorf("expected final answer to contain '5', got: '%s'", mathResult.FinalAnswer)
	}

	if len(mathResult.Steps) == 0 {
		t.Fatalf("expected steps, got 0")
	}

	// Verify Ukrainian text is generated
	hasUkrainian := detectUkrainian(mathResult.Steps[0].Title) || detectUkrainian(mathResult.Steps[0].Explanation) || detectUkrainian(mathResult.FinalAnswer)
	if !hasUkrainian {
		t.Errorf("expected Ukrainian text in steps or final answer, got: title=%s, explanation=%s, finalAnswer=%s", mathResult.Steps[0].Title, mathResult.Steps[0].Explanation, mathResult.FinalAnswer)
	}
}
