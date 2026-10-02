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
