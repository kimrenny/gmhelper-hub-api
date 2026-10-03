package geometry

import (
	"strings"
	"testing"
)

// Test 1 — Explicit target extraction
func TestExtractInputFacts_ExplicitTarget(t *testing.T) {
	payload := `{
		"triangle_1": {
			"points": [{"label": "A"}, {"label": "B"}, {"label": "C"}],
			"lines": {"AB": 5.0, "BC": 5.0, "AC": 6.0},
			"angles": {}
		},
		"target": "Find BH"
	}`

	facts, summary, err := ExtractInputFacts(payload)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if facts.ExplicitTarget != "Find BH" {
		t.Errorf("expected ExplicitTarget 'Find BH', got '%s'", facts.ExplicitTarget)
	}

	if !strings.Contains(summary, "Explicit problem goal") || !strings.Contains(summary, "Find BH") {
		t.Errorf("expected summary to contain explicit goal, got:\n%s", summary)
	}
}

// Test 2 — Additional conditions extraction
func TestExtractInputFacts_AdditionalConditions(t *testing.T) {
	payload := `{
		"triangle_1": {
			"points": [{"label": "A"}, {"label": "B"}, {"label": "C"}],
			"lines": {"AB": 5.0, "BC": 5.0, "AC": 6.0},
			"angles": {}
		},
		"additionalConditions": [
			"BH is perpendicular to AC",
			"  AB = BC  ",
			""
		],
		"target": "Find BH"
	}`

	facts, summary, err := ExtractInputFacts(payload)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(facts.AdditionalConditions) != 2 {
		t.Fatalf("expected 2 non-empty conditions, got %d: %v", len(facts.AdditionalConditions), facts.AdditionalConditions)
	}

	if facts.AdditionalConditions[0] != "BH is perpendicular to AC" {
		t.Errorf("expected 'BH is perpendicular to AC', got '%s'", facts.AdditionalConditions[0])
	}
	if facts.AdditionalConditions[1] != "AB = BC" {
		t.Errorf("expected 'AB = BC', got '%s'", facts.AdditionalConditions[1])
	}

	if !strings.Contains(summary, "Additional geometric constraints") || !strings.Contains(summary, "BH is perpendicular to AC") {
		t.Errorf("expected summary to contain conditions, got:\n%s", summary)
	}

	// Legacy "conditions" field test
	legacyPayload := `{
		"triangle_1": {
			"points": [{"label": "A"}, {"label": "B"}, {"label": "C"}],
			"lines": {"AB": 5.0}
		},
		"conditions": ["Point M is midpoint of AB"]
	}`
	legacyFacts, _, err := ExtractInputFacts(legacyPayload)
	if err != nil {
		t.Fatalf("unexpected error in legacy conditions: %v", err)
	}
	if len(legacyFacts.AdditionalConditions) != 1 || legacyFacts.AdditionalConditions[0] != "Point M is midpoint of AB" {
		t.Errorf("expected legacy condition to be extracted, got: %v", legacyFacts.AdditionalConditions)
	}
}

// Test 3 — Prompt propagation
func TestBuildPrompt_MetadataPropagation(t *testing.T) {
	payload := `{
		"triangle_1": {
			"points": [{"label": "A"}, {"label": "B"}, {"label": "C"}],
			"lines": {"AB": 5.0, "BC": 5.0, "AC": 6.0},
			"angles": {}
		},
		"additionalConditions": [
			"BH is perpendicular to AC"
		],
		"target": "Find BH"
	}`

	facts, summary, err := ExtractInputFacts(payload)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	prompt := BuildPrompt(facts, summary)

	if !strings.Contains(prompt, "Explicit Problem Goal (AUTHORITATIVE") {
		t.Errorf("prompt missing explicit problem goal header")
	}
	if !strings.Contains(prompt, "Find BH") {
		t.Errorf("prompt missing explicit target text 'Find BH'")
	}
	if !strings.Contains(prompt, "Additional Geometric Constraints:") {
		t.Errorf("prompt missing additional geometric constraints header")
	}
	if !strings.Contains(prompt, "- BH is perpendicular to AC") {
		t.Errorf("prompt missing condition text 'BH is perpendicular to AC'")
	}
	if !strings.Contains(prompt, `Solve the explicitly specified problem goal ("Find BH")`) {
		t.Errorf("prompt missing authoritative requirement for explicit goal")
	}
}

// Test 4 — Explicit target protects against unrelated generated target
func TestValidateTargetCompatibility_RejectsUnrelatedTarget(t *testing.T) {
	explicitTarget := "Find BH"

	unrelatedTarget := Target{
		Descriptions: []string{"Calculate the area of triangle ABC"},
		Variables:    []string{"area"},
	}

	err := ValidateTargetCompatibility(explicitTarget, unrelatedTarget)
	if err == nil {
		t.Fatalf("expected error when generated target only calculates area for explicit target 'Find BH', got nil")
	}

	if !strings.Contains(err.Error(), "missing referenced geometric element") && !strings.Contains(err.Error(), "BH") {
		t.Errorf("expected error mentioning missing element BH, got: %v", err)
	}

	// Full validator check with expectedFacts
	expectedFacts := sampleExpectedFacts()
	expectedFacts.ExplicitTarget = "Find BH"

	// validGeometryJSON modified so that target is purely area
	areaOnlyJSON := strings.Replace(
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

	_, valErr := ValidateGeometryResult(areaOnlyJSON, &expectedFacts)
	if valErr == nil {
		t.Fatalf("expected ValidateGeometryResult to reject unrelated area target when explicit target is 'Find BH'")
	}
	if !strings.Contains(valErr.Error(), "target validation failed") {
		t.Errorf("expected target validation failed error, got: %v", valErr)
	}
}

// Test 5 — Matching/compatible target
func TestValidateTargetCompatibility_AcceptsCompatibleTarget(t *testing.T) {
	expectedFacts := sampleExpectedFacts()
	expectedFacts.ExplicitTarget = "Find BH"

	// validGeometryJSON has target containing BH
	res, err := ValidateGeometryResult(validGeometryJSON(), &expectedFacts)
	if err != nil {
		t.Fatalf("expected validation success for compatible target, got: %v", err)
	}
	if res == nil {
		t.Fatalf("expected non-nil result")
	}

	// Single target variable BH
	singleBHJSON := strings.Replace(
		validGeometryJSON(),
		`"variables": ["BH", "S", "P"]`,
		`"variables": ["BH"]`,
		1,
	)
	resSingle, errSingle := ValidateGeometryResult(singleBHJSON, &expectedFacts)
	if errSingle != nil {
		t.Fatalf("expected validation success for single BH variable target, got: %v", errSingle)
	}
	if resSingle == nil {
		t.Fatalf("expected non-nil result for single BH variable")
	}
	// Test natural language phrasing variants for target BH
	phrasings := []string{
		"Find BH",
		"Find the length of BH",
		"Calculate the altitude BH",
		"Determine BH",
		"Calculate BH",
		"Compute the length of segment BH",
		"Знайти висоту BH",
		"Обчислити BH",
	}

	targetWithBH := Target{
		Descriptions: []string{"Find the length of the altitude BH perpendicular to AC"},
		Variables:    []string{"BH"},
	}

	for _, phrasing := range phrasings {
		if err := ValidateTargetCompatibility(phrasing, targetWithBH); err != nil {
			t.Errorf("expected phrasing '%s' to be accepted as compatible with BH target, got error: %v", phrasing, err)
		}
	}
}

// Test 6 — No explicit target
func TestExtractInputFacts_NoExplicitTarget(t *testing.T) {
	payload := `{
		"triangle_1": {
			"points": [{"label": "A"}, {"label": "B"}, {"label": "C"}],
			"lines": {"AB": 5.0, "BC": 5.0, "AC": 6.0},
			"angles": {}
		}
	}`

	facts, summary, err := ExtractInputFacts(payload)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if facts.ExplicitTarget != "" {
		t.Errorf("expected empty ExplicitTarget, got '%s'", facts.ExplicitTarget)
	}
	if len(facts.AdditionalConditions) != 0 {
		t.Errorf("expected empty AdditionalConditions, got %v", facts.AdditionalConditions)
	}

	prompt := BuildPrompt(facts, summary)
	if strings.Contains(prompt, "Explicit Problem Goal") {
		t.Errorf("prompt should not contain Explicit Problem Goal section when absent")
	}
	if strings.Contains(prompt, "Additional Geometric Constraints") {
		t.Errorf("prompt should not contain Additional Geometric Constraints section when absent")
	}

	// Validation without explicit target should accept standard valid result
	res, valErr := ValidateGeometryResult(validGeometryJSON(), &facts)
	if valErr != nil {
		t.Fatalf("expected validation to pass when no explicit target is set, got: %v", valErr)
	}
	if res == nil {
		t.Fatalf("expected non-nil result")
	}
}

// Test 7 — Empty/null metadata
func TestExtractInputFacts_EmptyNullMetadata(t *testing.T) {
	testCases := []struct {
		name    string
		payload string
	}{
		{
			name: "empty strings and empty array",
			payload: `{
				"triangle_1": {
					"points": [{"label": "A"}, {"label": "B"}, {"label": "C"}],
					"lines": {"AB": 5.0}
				},
				"target": "   ",
				"additionalConditions": []
			}`,
		},
		{
			name: "null target and null additionalConditions",
			payload: `{
				"triangle_1": {
					"points": [{"label": "A"}, {"label": "B"}, {"label": "C"}],
					"lines": {"AB": 5.0}
				},
				"target": null,
				"additionalConditions": null
			}`,
		},
		{
			name: "non-string entries in additionalConditions",
			payload: `{
				"triangle_1": {
					"points": [{"label": "A"}, {"label": "B"}, {"label": "C"}],
					"lines": {"AB": 5.0}
				},
				"target": 123,
				"additionalConditions": [null, 456, "   ", true, {"foo": "bar"}]
			}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			facts, _, err := ExtractInputFacts(tc.payload)
			if err != nil {
				t.Fatalf("unexpected error for %s: %v", tc.name, err)
			}
			if facts.ExplicitTarget != "" {
				t.Errorf("expected empty ExplicitTarget for %s, got: '%s'", tc.name, facts.ExplicitTarget)
			}
			if len(facts.AdditionalConditions) != 0 {
				t.Errorf("expected empty AdditionalConditions for %s, got: %v", tc.name, facts.AdditionalConditions)
			}
		})
	}
}

// Test 8 — Multiple figures
func TestExtractInputFacts_MultipleFigures(t *testing.T) {
	payload := `{
		"triangle_1": {
			"points": [{"label": "A"}, {"label": "B"}, {"label": "C"}],
			"lines": {"AB": 5.0, "BC": 5.0, "AC": 6.0}
		},
		"circle_1": {
			"points": [{"label": "O"}]
		},
		"additionalConditions": [
			"Circle O is incircle of triangle ABC"
		],
		"target": "Find radius of incircle"
	}`

	facts, summary, err := ExtractInputFacts(payload)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(facts.Figures) != 2 {
		t.Fatalf("expected 2 figures, got %d", len(facts.Figures))
	}
	if facts.ExplicitTarget != "Find radius of incircle" {
		t.Errorf("expected explicit target 'Find radius of incircle', got '%s'", facts.ExplicitTarget)
	}
	if len(facts.AdditionalConditions) != 1 || facts.AdditionalConditions[0] != "Circle O is incircle of triangle ABC" {
		t.Errorf("expected additional condition 'Circle O is incircle of triangle ABC', got: %v", facts.AdditionalConditions)
	}

	if !strings.Contains(summary, "triangle_1") || !strings.Contains(summary, "circle_1") {
		t.Errorf("summary should contain both figures, got:\n%s", summary)
	}
}

// Test 9 — Explicit language field extraction
func TestExtractInputFacts_LanguageExplicit(t *testing.T) {
	testCases := []struct {
		name         string
		payload      string
		expectedLang string
	}{
		{
			name: "explicit language ru",
			payload: `{
				"triangle_1": {
					"points": [{"label": "A"}, {"label": "B"}, {"label": "C"}],
					"lines": {"AB": 5.0}
				},
				"language": "ru"
			}`,
			expectedLang: "ru",
		},
		{
			name: "explicit lang uk",
			payload: `{
				"triangle_1": {
					"points": [{"label": "A"}, {"label": "B"}, {"label": "C"}],
					"lines": {"AB": 5.0}
				},
				"lang": "uk"
			}`,
			expectedLang: "uk",
		},
		{
			name: "explicit locale de",
			payload: `{
				"triangle_1": {
					"points": [{"label": "A"}, {"label": "B"}, {"label": "C"}],
					"lines": {"AB": 5.0}
				},
				"locale": "de"
			}`,
			expectedLang: "de",
		},
		{
			name: "inside task map language fr",
			payload: `{
				"task": {
					"triangle_1": {
						"points": [{"label": "A"}, {"label": "B"}, {"label": "C"}],
						"lines": {"AB": 5.0}
					},
					"language": "fr"
				}
			}`,
			expectedLang: "fr",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			facts, _, err := ExtractInputFacts(tc.payload)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if facts.Language != tc.expectedLang {
				t.Errorf("expected Language '%s', got '%s'", tc.expectedLang, facts.Language)
			}
		})
	}
}

// Test 10 — Language auto-detection from Cyrillic / Ukrainian / German text
func TestExtractInputFacts_LanguageAutoDetect(t *testing.T) {
	testCases := []struct {
		name         string
		payload      string
		expectedLang string
	}{
		{
			name: "ukrainian text with special character і",
			payload: `{
				"triangle_1": {
					"points": [{"label": "A"}, {"label": "B"}, {"label": "C"}],
					"lines": {"AB": 5.0}
				},
				"target": "Знайти висоту BH"
			}`,
			expectedLang: "uk",
		},
		{
			name: "russian text with special character ы",
			payload: `{
				"triangle_1": {
					"points": [{"label": "A"}, {"label": "B"}, {"label": "C"}],
					"lines": {"AB": 5.0}
				},
				"target": "Найти высоту BH"
			}`,
			expectedLang: "ru",
		},
		{
			name: "german text with Berechne",
			payload: `{
				"triangle_1": {
					"points": [{"label": "A"}, {"label": "B"}, {"label": "C"}],
					"lines": {"AB": 5.0}
				},
				"target": "Berechne die Höhe BH"
			}`,
			expectedLang: "de",
		},
		{
			name: "english text default",
			payload: `{
				"triangle_1": {
					"points": [{"label": "A"}, {"label": "B"}, {"label": "C"}],
					"lines": {"AB": 5.0}
				},
				"target": "Find BH"
			}`,
			expectedLang: "en",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			facts, _, err := ExtractInputFacts(tc.payload)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if facts.Language != tc.expectedLang {
				t.Errorf("expected Language '%s', got '%s'", tc.expectedLang, facts.Language)
			}
		})
	}
}

// Test 11 — Prompt contains language requirements
func TestBuildPrompt_LanguageInstruction(t *testing.T) {
	factsRu := InputFacts{
		Figures: []Figure{
			{ID: "triangle_1", Type: "triangle", Vertices: []string{"A", "B", "C"}},
		},
		Lengths:        map[string]float64{"AB": 5, "BC": 5, "AC": 6},
		ExplicitTarget: "Найти высоту BH",
		Language:       "ru",
	}

	promptRu := BuildPrompt(factsRu, formatFactsSummary(factsRu))
	if !strings.Contains(promptRu, "Russian (Русский)") {
		t.Errorf("expected promptRu to contain 'Russian (Русский)', got:\n%s", promptRu)
	}
	if !strings.Contains(promptRu, "All human-readable natural language text MUST be written in Russian (Русский)") {
		t.Errorf("expected promptRu to contain language instruction")
	}

	factsUk := InputFacts{
		Figures: []Figure{
			{ID: "triangle_1", Type: "triangle", Vertices: []string{"A", "B", "C"}},
		},
		Lengths:        map[string]float64{"AB": 5, "BC": 5, "AC": 6},
		ExplicitTarget: "Знайти висоту BH",
		Language:       "uk",
	}

	promptUk := BuildPrompt(factsUk, formatFactsSummary(factsUk))
	if !strings.Contains(promptUk, "Ukrainian (Українська)") {
		t.Errorf("expected promptUk to contain 'Ukrainian (Українська)', got:\n%s", promptUk)
	}
}

// Test A — No figure (JSON and pure text payload)
func TestExtractInputFacts_NoFigure(t *testing.T) {
	// JSON format without figures
	jsonPayload := `{
		"problem": "In triangle ABC, AB = 5, BC = 5, AC = 6. Find the altitude BH to AC.",
		"additionalConditions": [
			"BH is perpendicular to AC"
		],
		"target": "Find BH",
		"language": "en"
	}`

	facts, summary, err := ExtractInputFacts(jsonPayload)
	if err != nil {
		t.Fatalf("unexpected error for figureless geometry payload: %v", err)
	}

	if len(facts.Figures) != 0 {
		t.Errorf("expected 0 figures, got %d", len(facts.Figures))
	}
	if facts.ExplicitTarget != "Find BH" {
		t.Errorf("expected ExplicitTarget 'Find BH', got '%s'", facts.ExplicitTarget)
	}
	if facts.Problem != "In triangle ABC, AB = 5, BC = 5, AC = 6. Find the altitude BH to AC." {
		t.Errorf("unexpected Problem: '%s'", facts.Problem)
	}
	if len(facts.AdditionalConditions) != 1 || facts.AdditionalConditions[0] != "BH is perpendicular to AC" {
		t.Errorf("unexpected AdditionalConditions: %v", facts.AdditionalConditions)
	}
	if facts.Lengths["AB"] != 5.0 || facts.Lengths["BC"] != 5.0 || facts.Lengths["AC"] != 6.0 {
		t.Errorf("expected extracted side lengths AB=5, BC=5, AC=6, got %v", facts.Lengths)
	}

	prompt := BuildPrompt(facts, summary)
	if !strings.Contains(prompt, "Diagram: None supplied") {
		t.Errorf("expected prompt to indicate no diagram supplied, got:\n%s", prompt)
	}
	if !strings.Contains(prompt, "Find BH") {
		t.Errorf("expected prompt to contain target 'Find BH'")
	}

	// Plain text format
	textPayload := "In triangle ABC, AB = 5, BC = 5, AC = 6. BH is perpendicular to AC. Find BH."
	factsText, summaryText, errText := ExtractInputFacts(textPayload)
	if errText != nil {
		t.Fatalf("unexpected error for plain text payload: %v", errText)
	}
	if len(factsText.Figures) != 0 {
		t.Errorf("expected 0 figures for plain text, got %d", len(factsText.Figures))
	}
	if factsText.Lengths["AB"] != 5.0 || factsText.Lengths["BC"] != 5.0 || factsText.Lengths["AC"] != 6.0 {
		t.Errorf("expected extracted side lengths from plain text, got %v", factsText.Lengths)
	}
	if !strings.Contains(summaryText, "Diagram: None supplied") {
		t.Errorf("expected summaryText to indicate no diagram supplied")
	}
}

// Test B — Text overrides drawing measurements
func TestExtractInputFacts_TextOverridesDrawing(t *testing.T) {
	payload := `{
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

	facts, summary, err := ExtractInputFacts(payload)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Textual values (5, 5, 6) must override canvas line values (7, 7, 8)
	if facts.Lengths["AB"] != 5.0 {
		t.Errorf("expected textual value AB=5.0 to override drawing 7.0, got %v", facts.Lengths["AB"])
	}
	if facts.Lengths["BC"] != 5.0 {
		t.Errorf("expected textual value BC=5.0 to override drawing 7.0, got %v", facts.Lengths["BC"])
	}
	if facts.Lengths["AC"] != 6.0 {
		t.Errorf("expected textual value AC=6.0 to override drawing 8.0, got %v", facts.Lengths["AC"])
	}

	if !strings.Contains(summary, "AB = 5") {
		t.Errorf("summary should contain textual length AB = 5, got:\n%s", summary)
	}
	if strings.Contains(summary, "AB = 7") {
		t.Errorf("summary should NOT contain drawing length AB = 7")
	}

	prompt := BuildPrompt(facts, summary)
	if !strings.Contains(prompt, "ALWAYS use the textual value") {
		t.Errorf("prompt should include instruction that textual value wins over drawing")
	}
}

// Test C — Schematic coordinates are not mathematical measurements
func TestExtractInputFacts_SchematicCoordinates(t *testing.T) {
	payload := `{
		"triangle_1": {
			"points": [
				{"label": "A", "x": 0, "y": 0},
				{"label": "B", "x": 1000, "y": 0},
				{"label": "C", "x": 500, "y": 800}
			]
		},
		"problem": "In triangle ABC, AB = 5, BC = 5, AC = 6. Find BH.",
		"target": "Find BH"
	}`

	facts, summary, err := ExtractInputFacts(payload)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Mathematical length from text is AB = 5, not pixel distance 1000
	if facts.Lengths["AB"] != 5.0 {
		t.Errorf("expected mathematical length AB = 5, got %v", facts.Lengths["AB"])
	}

	prompt := BuildPrompt(facts, summary)
	if !strings.Contains(prompt, "Never infer an exact length, angle, ratio, or other mathematical measurement from canvas/pixel coordinates") {
		t.Errorf("prompt should instruct not to infer measurements from pixel coordinates")
	}
}

// Test D — Target without figure
func TestExtractInputFacts_TargetWithoutFigure(t *testing.T) {
	payload := `{
		"problem": "In triangle ABC, AB = 5, BC = 5, AC = 6. BH is perpendicular to AC.",
		"target": "Find BH"
	}`

	facts, _, err := ExtractInputFacts(payload)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if facts.ExplicitTarget != "Find BH" {
		t.Errorf("expected ExplicitTarget 'Find BH', got '%s'", facts.ExplicitTarget)
	}
	if len(facts.Figures) != 0 {
		t.Errorf("expected 0 figures, got %d", len(facts.Figures))
	}
}

// Test E — Additional conditions without figure
func TestExtractInputFacts_AdditionalConditionsWithoutFigure(t *testing.T) {
	payload := `{
		"problem": "In triangle ABC, AB = 5, BC = 5, AC = 6.",
		"additionalConditions": [
			"BH is perpendicular to AC",
			"H is on AC"
		],
		"target": "Find BH"
	}`

	facts, summary, err := ExtractInputFacts(payload)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(facts.AdditionalConditions) != 2 {
		t.Fatalf("expected 2 conditions, got %d", len(facts.AdditionalConditions))
	}
	if facts.AdditionalConditions[0] != "BH is perpendicular to AC" {
		t.Errorf("unexpected condition 0: %s", facts.AdditionalConditions[0])
	}
	if facts.AdditionalConditions[1] != "H is on AC" {
		t.Errorf("unexpected condition 1: %s", facts.AdditionalConditions[1])
	}

	prompt := BuildPrompt(facts, summary)
	if !strings.Contains(prompt, "BH is perpendicular to AC") || !strings.Contains(prompt, "H is on AC") {
		t.Errorf("prompt missing conditions")
	}
}

