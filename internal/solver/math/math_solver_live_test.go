package math

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"gmhelper.solution-hub/internal/config"
	"gmhelper.solution-hub/internal/gemini"
	"gmhelper.solution-hub/internal/solver"
)

func TestMathSolver_LiveGeminiSmokeTest(t *testing.T) {
	cfg := config.Load()
	if cfg.GeminiAPIKey == "" {
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
