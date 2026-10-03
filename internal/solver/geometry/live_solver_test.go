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

func TestGeometrySolver_LiveSolve_Russian(t *testing.T) {
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
    "lines": {
      "AB": 5,
      "BC": 5,
      "AC": 6
    },
    "angles": {
      "ABC": 60
    },
    "points": [
      { "label": "A" },
      { "label": "B" },
      { "label": "C" }
    ]
  },
  "additionalConditions": [
    "BH перпендикулярна AC"
  ],
  "target": "Найти высоту BH",
  "language": "ru"
}`

	task := solver.Task{
		TaskID:      "live-test-target-bh-ru",
		ProblemType: "geometry",
		Payload:     livePayload,
		UserID:      "test-user-ru",
	}

	res, err := solverInstance.Solve(ctx, task)
	if err != nil {
		t.Fatalf("Live Russian solve failed: %v", err)
	}

	if !res.Success {
		t.Fatalf("Live Russian solve returned failure")
	}

	var parsed GeometryResult
	if err := json.Unmarshal([]byte(res.RawOutput), &parsed); err != nil {
		t.Fatalf("Failed to parse live result JSON: %v", err)
	}

	fmt.Printf("=== LIVE RUSSIAN SOLVE RESULT ===\n")
	fmt.Printf("ProblemStatement: %s\n", parsed.ProblemStatement)
	fmt.Printf("Target Descriptions: %v\n", parsed.Target.Descriptions)
	fmt.Printf("Target Variables: %v\n", parsed.Target.Variables)
	fmt.Printf("FinalAnswer: %s\n", parsed.FinalAnswer)
	fmt.Printf("LatexAnswer: %s\n", parsed.LatexAnswer)
	fmt.Printf("=================================\n")

	// Verify target solved BH
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
		t.Errorf("Live Russian solve did not set target to BH: descriptions=%v, variables=%v", parsed.Target.Descriptions, parsed.Target.Variables)
	}

	// Verify Russian text is present
	if !detectRussian(parsed.ProblemStatement) && !detectRussian(parsed.FinalAnswer) {
		t.Errorf("Expected Russian natural language text in output, got: problemStatement=%s, finalAnswer=%s", parsed.ProblemStatement, parsed.FinalAnswer)
	}
}

func TestGeometrySolver_LiveSolve_Ukrainian(t *testing.T) {
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
    "lines": {
      "AB": 5,
      "BC": 5,
      "AC": 6
    },
    "points": [
      { "label": "A" },
      { "label": "B" },
      { "label": "C" }
    ]
  },
  "additionalConditions": [
    "BH перпендикулярна до AC"
  ],
  "target": "Знайти висоту BH",
  "language": "uk"
}`

	task := solver.Task{
		TaskID:      "live-test-target-bh-uk",
		ProblemType: "geometry",
		Payload:     livePayload,
		UserID:      "test-user-uk",
	}

	res, err := solverInstance.Solve(ctx, task)
	if err != nil {
		t.Fatalf("Live Ukrainian solve failed: %v", err)
	}

	if !res.Success {
		t.Fatalf("Live Ukrainian solve returned failure")
	}

	var parsed GeometryResult
	if err := json.Unmarshal([]byte(res.RawOutput), &parsed); err != nil {
		t.Fatalf("Failed to parse live result JSON: %v", err)
	}

	fmt.Printf("=== LIVE UKRAINIAN SOLVE RESULT ===\n")
	fmt.Printf("ProblemStatement: %s\n", parsed.ProblemStatement)
	fmt.Printf("Target Descriptions: %v\n", parsed.Target.Descriptions)
	fmt.Printf("Target Variables: %v\n", parsed.Target.Variables)
	fmt.Printf("FinalAnswer: %s\n", parsed.FinalAnswer)
	fmt.Printf("LatexAnswer: %s\n", parsed.LatexAnswer)
	fmt.Printf("===================================\n")

	// Verify target solved BH
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
		t.Errorf("Live Ukrainian solve did not set target to BH: descriptions=%v, variables=%v", parsed.Target.Descriptions, parsed.Target.Variables)
	}

	// Verify Ukrainian text is present
	if !detectUkrainian(parsed.ProblemStatement) && !detectUkrainian(parsed.FinalAnswer) {
		t.Errorf("Expected Ukrainian natural language text in output, got: problemStatement=%s, finalAnswer=%s", parsed.ProblemStatement, parsed.FinalAnswer)
	}
}

// Case 1 — Text-only Geometry (no figure supplied)
func TestGeometrySolver_LiveSolve_TextOnly(t *testing.T) {
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

	textOnlyPayload := `{
		"problem": "In triangle ABC, AB = 5, BC = 5, AC = 6. Find the altitude BH to AC.",
		"additionalConditions": [
			"BH is perpendicular to AC"
		],
		"target": "Find BH",
		"language": "en"
	}`

	task := solver.Task{
		TaskID:      "live-test-text-only",
		ProblemType: "geometry",
		Payload:     textOnlyPayload,
		UserID:      "test-user-text-only",
	}

	res, err := solverInstance.Solve(ctx, task)
	if err != nil {
		t.Fatalf("Live text-only solve failed: %v", err)
	}

	if !res.Success {
		t.Fatalf("Live text-only solve returned failure")
	}

	var parsed GeometryResult
	if err := json.Unmarshal([]byte(res.RawOutput), &parsed); err != nil {
		t.Fatalf("Failed to parse live result JSON: %v", err)
	}

	fmt.Printf("=== LIVE TEXT-ONLY SOLVE RESULT ===\n")
	fmt.Printf("ProblemStatement: %s\n", parsed.ProblemStatement)
	fmt.Printf("Target Descriptions: %v\n", parsed.Target.Descriptions)
	fmt.Printf("Target Variables: %v\n", parsed.Target.Variables)
	fmt.Printf("FinalAnswer: %s\n", parsed.FinalAnswer)
	fmt.Printf("LatexAnswer: %s\n", parsed.LatexAnswer)
	fmt.Printf("DerivedFacts.Lengths: %v\n", parsed.DerivedFacts.Lengths)
	fmt.Printf("DerivedFacts.Metrics: %v\n", parsed.DerivedFacts.Metrics)
	fmt.Printf("===================================\n")

	// Verify that BH = 4 is found in final answer or derived lengths
	bhVal, hasBH := parsed.DerivedFacts.Lengths["BH"]
	if !hasBH {
		bhVal, hasBH = parsed.DerivedFacts.Lengths["HB"]
	}

	if hasBH && bhVal != 4.0 {
		t.Errorf("expected BH = 4.0 in derived lengths, got: %v", bhVal)
	}

	if !strings.Contains(parsed.FinalAnswer, "4") && !strings.Contains(parsed.LatexAnswer, "4") {
		t.Errorf("expected final answer to contain 4, got finalAnswer: %s, latexAnswer: %s", parsed.FinalAnswer, parsed.LatexAnswer)
	}
}

// Case 2 — Conflicting visual data (Text: AB = 5, Drawing: AB = 7)
func TestGeometrySolver_LiveSolve_ConflictingVisualData(t *testing.T) {
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

	conflictingPayload := `{
		"triangle_1": {
			"tool": "triangle",
			"path": [
				{ "x": 100, "y": 100 },
				{ "x": 800, "y": 100 },
				{ "x": 450, "y": 50 }
			],
			"lines": {
				"AB": 7,
				"BC": 7,
				"AC": 8
			},
			"points": [
				{ "label": "A" },
				{ "label": "B" },
				{ "label": "C" }
			]
		},
		"problem": "In triangle ABC, AB = 5, BC = 5, AC = 6.",
		"additionalConditions": [
			"BH is perpendicular to AC"
		],
		"target": "Find BH",
		"language": "en"
	}`

	task := solver.Task{
		TaskID:      "live-test-conflict",
		ProblemType: "geometry",
		Payload:     conflictingPayload,
		UserID:      "test-user-conflict",
	}

	res, err := solverInstance.Solve(ctx, task)
	if err != nil {
		t.Fatalf("Live conflict solve failed: %v", err)
	}

	if !res.Success {
		t.Fatalf("Live conflict solve returned failure")
	}

	var parsed GeometryResult
	if err := json.Unmarshal([]byte(res.RawOutput), &parsed); err != nil {
		t.Fatalf("Failed to parse live result JSON: %v", err)
	}

	fmt.Printf("=== LIVE CONFLICT SOLVE RESULT ===\n")
	fmt.Printf("ProblemStatement: %s\n", parsed.ProblemStatement)
	fmt.Printf("FinalAnswer: %s\n", parsed.FinalAnswer)
	fmt.Printf("LatexAnswer: %s\n", parsed.LatexAnswer)
	fmt.Printf("DerivedFacts.Lengths: %v\n", parsed.DerivedFacts.Lengths)
	fmt.Printf("==================================\n")

	// Verify that text AB = 5 was used to calculate BH = 4 (NOT drawing AB = 7)
	if !strings.Contains(parsed.FinalAnswer, "4") && !strings.Contains(parsed.LatexAnswer, "4") {
		t.Errorf("expected final answer to calculate BH = 4 using textual values, got finalAnswer: %s, latexAnswer: %s", parsed.FinalAnswer, parsed.LatexAnswer)
	}
}
