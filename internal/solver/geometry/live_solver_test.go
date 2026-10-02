package geometry

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"gmhelper.solution-hub/internal/config"
	"gmhelper.solution-hub/internal/gemini"
	"gmhelper.solution-hub/internal/solver"
)

func TestGeometrySolver_LiveSolve(t *testing.T) {
	if os.Getenv("RUN_LIVE_GEMINI_TESTS") != "1" && os.Getenv("RUN_LIVE_GEMINI_TESTS") != "true" {
		t.Skip("skipping live Gemini test: set RUN_LIVE_GEMINI_TESTS=1 to enable")
	}

	cfg := config.Load()
	if cfg.GeminiAPIKey == "" || cfg.GeminiAPIKey == "your_gemini_api_key_here" {
		t.Skip("skipping live test: Gemini API key not configured")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()

	client, err := gemini.NewGenAIClient(ctx, cfg)
	if err != nil {
		t.Fatalf("failed to initialize live Gemini client: %v", err)
	}

	solverInstance := NewGeometrySolver(client)

	livePayload := `{
  "triangle_1": {
    "tool": "triangle",
    "path": [
      { "x": 100, "y": 100 },
      { "x": 300, "y": 100 },
      { "x": 200, "y": 50 }
    ],
    "lines": {
      "AB": 5,
      "BC": 5,
      "AC": 6
    },
    "angles": {
      "ABC": 60
    },
    "elements": [],
    "points": [
      {
        "x": 100,
        "y": 100,
        "label": "A",
        "attachedToFigure": "triangle_1"
      },
      {
        "x": 300,
        "y": 100,
        "label": "B",
        "attachedToFigure": "triangle_1"
      },
      {
        "x": 200,
        "y": 50,
        "label": "C",
        "attachedToFigure": "triangle_1"
      }
    ]
  },
  "additionalConditions": [
    "BH is perpendicular to AC"
  ],
  "target": "Find BH"
}`

	task := solver.Task{
		TaskID:      "live-test-target-bh",
		ProblemType: "geometry",
		Payload:     livePayload,
		UserID:      "test-user",
	}

	res, err := solverInstance.Solve(ctx, task)
	if err != nil {
		t.Fatalf("Live solve failed: %v", err)
	}

	if !res.Success {
		t.Fatalf("Live solve returned failure")
	}

	var parsed GeometryResult
	if err := json.Unmarshal([]byte(res.RawOutput), &parsed); err != nil {
		t.Fatalf("Failed to parse live result JSON: %v", err)
	}

	fmt.Printf("=== LIVE SOLVE RESULT ===\n")
	fmt.Printf("ProblemStatement: %s\n", parsed.ProblemStatement)
	fmt.Printf("Target Descriptions: %v\n", parsed.Target.Descriptions)
	fmt.Printf("Target Variables: %v\n", parsed.Target.Variables)
	fmt.Printf("FinalAnswer: %s\n", parsed.FinalAnswer)
	fmt.Printf("LatexAnswer: %s\n", parsed.LatexAnswer)
	fmt.Printf("DerivedFacts.Lengths: %v\n", parsed.DerivedFacts.Lengths)
	fmt.Printf("DerivedFacts.Metrics: %v\n", parsed.DerivedFacts.Metrics)
	fmt.Printf("=========================\n")

	// Verify that the generated solution actually solved BH
	solvedBH := false
	for _, v := range parsed.Target.Variables {
		if v == "BH" || v == "bh" || v == "h" || v == "h_b" {
			solvedBH = true
		}
	}
	for _, d := range parsed.Target.Descriptions {
		if containsEntityWord(strings.ToUpper(d), "BH") {
			solvedBH = true
		}
	}

	if !solvedBH {
		t.Errorf("Live solve did not set target to BH: descriptions=%v, variables=%v", parsed.Target.Descriptions, parsed.Target.Variables)
	}
}
