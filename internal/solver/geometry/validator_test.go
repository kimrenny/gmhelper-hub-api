package geometry

import (
	"strings"
	"testing"
)

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

func sampleExpectedFacts() InputFacts {
	return InputFacts{
		Figures: []Figure{
			{
				ID:       "triangle_1",
				Type:     "triangle",
				Vertices: []string{"A", "B", "C"},
			},
		},
		Lengths: map[string]float64{
			"AB": 5.0,
			"BC": 5.0,
			"AC": 6.0,
		},
		Angles: map[string]float64{},
	}
}

func TestValidateGeometryResult_Success(t *testing.T) {
	expected := sampleExpectedFacts()
	res, err := ValidateGeometryResult(validGeometryJSON(), &expected)
	if err != nil {
		t.Fatalf("expected valid result, got error: %v", err)
	}

	if res.ProblemType != "geometry" {
		t.Errorf("expected problemType 'geometry', got '%s'", res.ProblemType)
	}
	if res.Status != "completed" {
		t.Errorf("expected status 'completed', got '%s'", res.Status)
	}
	if len(res.Steps) != 4 {
		t.Errorf("expected 4 steps, got %d", len(res.Steps))
	}
	if res.DerivedFacts.Metrics["area"] != 12.0 {
		t.Errorf("expected area 12.0, got %v", res.DerivedFacts.Metrics["area"])
	}
	if res.DerivedFacts.Metrics["perimeter"] != 16.0 {
		t.Errorf("expected perimeter 16.0, got %v", res.DerivedFacts.Metrics["perimeter"])
	}
}

func TestValidateGeometryResult_UnknownFields(t *testing.T) {
	// Top-level unknown field
	topLevelUnknown := strings.Replace(validGeometryJSON(), `"status": "completed"`, `"status": "completed", "unknownField": "fail"`, 1)
	if _, err := ValidateGeometryResult(topLevelUnknown, nil); err == nil || !strings.Contains(err.Error(), "unknown field") {
		t.Errorf("expected top-level unknown field rejection, got: %v", err)
	}

	// Nested unknown field inside target
	nestedTargetUnknown := strings.Replace(validGeometryJSON(), `"variables": ["BH", "S", "P"]`, `"variables": ["BH", "S", "P"], "extraTarget": 123`, 1)
	if _, err := ValidateGeometryResult(nestedTargetUnknown, nil); err == nil || !strings.Contains(err.Error(), "unknown field") {
		t.Errorf("expected nested target unknown field rejection, got: %v", err)
	}

	// Nested unknown field inside steps[0]
	stepUnknown := strings.Replace(validGeometryJSON(), `"title": "Construct Altitude and Determine Segment Lengths"`, `"title": "Construct Altitude", "badStepField": true`, 1)
	if _, err := ValidateGeometryResult(stepUnknown, nil); err == nil || !strings.Contains(err.Error(), "unknown field") {
		t.Errorf("expected step unknown field rejection, got: %v", err)
	}

	// Extra error field
	extraError := strings.Replace(validGeometryJSON(), `"finalAnswer":`, `"error": "some error", "finalAnswer":`, 1)
	if _, err := ValidateGeometryResult(extraError, nil); err == nil || !strings.Contains(err.Error(), "unknown field") {
		t.Errorf("expected extra error field rejection, got: %v", err)
	}
}

func TestValidateGeometryResult_TrailingData(t *testing.T) {
	trailing := validGeometryJSON() + ` {"extra": true}`
	if _, err := ValidateGeometryResult(trailing, nil); err == nil || !strings.Contains(err.Error(), "unexpected trailing data") {
		t.Errorf("expected trailing data rejection, got: %v", err)
	}
}

func TestValidateGeometryResult_ContractFields(t *testing.T) {
	// Invalid problemType
	invalidType := strings.Replace(validGeometryJSON(), `"problemType": "geometry"`, `"problemType": "math"`, 1)
	if _, err := ValidateGeometryResult(invalidType, nil); err == nil || !strings.Contains(err.Error(), "expected 'geometry'") {
		t.Errorf("expected invalid problemType rejection, got: %v", err)
	}

	// Invalid status
	invalidStatus := strings.Replace(validGeometryJSON(), `"status": "completed"`, `"status": "unsolvable"`, 1)
	if _, err := ValidateGeometryResult(invalidStatus, nil); err == nil || !strings.Contains(err.Error(), "invalid status") {
		t.Errorf("expected invalid status rejection, got: %v", err)
	}

	// Empty problemStatement
	emptyStatement := strings.Replace(validGeometryJSON(), `"problemStatement": "In isosceles triangle ABC with side lengths AB = 5, BC = 5, and base AC = 6, find the altitude BH to base AC, the area S, and the perimeter P."`, `"problemStatement": "  "`, 1)
	if _, err := ValidateGeometryResult(emptyStatement, nil); err == nil || !strings.Contains(err.Error(), "problemStatement") {
		t.Errorf("expected empty problemStatement rejection, got: %v", err)
	}

	// Empty finalAnswer
	emptyAnswer := strings.Replace(validGeometryJSON(), `"finalAnswer": "Altitude BH = 4, Area S = 12, Perimeter P = 16"`, `"finalAnswer": ""`, 1)
	if _, err := ValidateGeometryResult(emptyAnswer, nil); err == nil || !strings.Contains(err.Error(), "finalAnswer") {
		t.Errorf("expected empty finalAnswer rejection, got: %v", err)
	}
}

func TestValidateGeometryResult_StepsValidation(t *testing.T) {
	// Empty steps array
	noSteps := strings.Replace(validGeometryJSON(), `"steps": [
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
  ]`, `"steps": []`, 1)
	if _, err := ValidateGeometryResult(noSteps, nil); err == nil || !strings.Contains(err.Error(), "steps array cannot be empty") {
		t.Errorf("expected empty steps rejection, got: %v", err)
	}

	// Non-sequential step numbers (1, 3)
	badSeq := strings.Replace(validGeometryJSON(), `"stepNumber": 2`, `"stepNumber": 3`, 1)
	if _, err := ValidateGeometryResult(badSeq, nil); err == nil || !strings.Contains(err.Error(), "invalid stepNumber") {
		t.Errorf("expected non-sequential stepNumber rejection, got: %v", err)
	}

	// Starting at step 2
	startTwo := strings.Replace(validGeometryJSON(), `"stepNumber": 1`, `"stepNumber": 2`, 1)
	if _, err := ValidateGeometryResult(startTwo, nil); err == nil || !strings.Contains(err.Error(), "invalid stepNumber") {
		t.Errorf("expected step starting at 2 rejection, got: %v", err)
	}
}

func TestValidateGeometryResult_InputFactParity(t *testing.T) {
	expected := sampleExpectedFacts()

	// 1. Changed side length: AB changed from 5.0 to 4.0
	alteredLength := strings.Replace(validGeometryJSON(), `"AB": 5.0`, `"AB": 4.0`, 1)
	if _, err := ValidateGeometryResult(alteredLength, &expected); err == nil || !strings.Contains(err.Error(), "input length mismatch") {
		t.Errorf("expected input length mismatch error, got: %v", err)
	}

	// 2. Missing side length: AC missing
	missingLength := strings.Replace(validGeometryJSON(), `"AC": 6.0`, `"XX": 6.0`, 1)
	if _, err := ValidateGeometryResult(missingLength, &expected); err == nil || !strings.Contains(err.Error(), "missing input length") {
		t.Errorf("expected missing input length error, got: %v", err)
	}

	// 3. Unexpected additional side length in inputFacts
	extraLength := strings.Replace(validGeometryJSON(), `"AC": 6.0`, `"AC": 6.0, "AD": 10.0`, 1)
	if _, err := ValidateGeometryResult(extraLength, &expected); err == nil || !strings.Contains(err.Error(), "unexpected input length") {
		t.Errorf("expected unexpected input length rejection, got: %v", err)
	}

	// 4. Unexpected additional angle in inputFacts
	extraAngle := strings.Replace(validGeometryJSON(), `"angles": {}`, `"angles": {"XYZ": 45.0}`, 1)
	if _, err := ValidateGeometryResult(extraAngle, &expected); err == nil || !strings.Contains(err.Error(), "unexpected input angle") {
		t.Errorf("expected unexpected input angle rejection, got: %v", err)
	}

	// 5. Unexpected additional figure in inputFacts
	extraFigure := strings.Replace(validGeometryJSON(), `"figures": [`, `"figures": [{"id": "rect_2", "type": "rectangle", "vertices": ["D","E","F","G"]}, `, 1)
	if _, err := ValidateGeometryResult(extraFigure, &expected); err == nil || !strings.Contains(err.Error(), "figure count mismatch") {
		t.Errorf("expected unexpected figure rejection, got: %v", err)
	}

	// 6. Changed angle when expected angle is set
	expectedWithAngle := sampleExpectedFacts()
	expectedWithAngle.Angles["ABC"] = 73.74
	validWithInputAngle := strings.Replace(validGeometryJSON(), `"angles": {}`, `"angles": {"ABC": 73.74}`, 1)
	if _, err := ValidateGeometryResult(validWithInputAngle, &expectedWithAngle); err != nil {
		t.Errorf("expected valid with matching angle, got: %v", err)
	}

	alteredAngle := strings.Replace(validGeometryJSON(), `"angles": {}`, `"angles": {"ABC": 60.0}`, 1)
	if _, err := ValidateGeometryResult(alteredAngle, &expectedWithAngle); err == nil || !strings.Contains(err.Error(), "input angle mismatch") {
		t.Errorf("expected input angle mismatch error, got: %v", err)
	}
}

func TestValidateGeometryResult_AuxiliaryConstructions(t *testing.T) {
	// Unknown fromVertex "Z"
	unknownVertex := strings.Replace(validGeometryJSON(), `"fromVertex": "B"`, `"fromVertex": "Z"`, 1)
	if _, err := ValidateGeometryResult(unknownVertex, nil); err == nil || !strings.Contains(err.Error(), "unknown fromVertex 'Z'") {
		t.Errorf("expected unknown fromVertex rejection, got: %v", err)
	}

	// Invalid toSegment "XY" (neither vertex exists)
	invalidSegment := strings.Replace(validGeometryJSON(), `"toSegment": "AC"`, `"toSegment": "XY"`, 1)
	if _, err := ValidateGeometryResult(invalidSegment, nil); err == nil || !strings.Contains(err.Error(), "invalid segment vertices") {
		t.Errorf("expected invalid toSegment rejection, got: %v", err)
	}

	// Nonexistent segment with existing vertices (e.g. vertices A, B, C, D exist in disconnected elements, but AD is not an edge)
	fourVertexJSON := strings.Replace(validGeometryJSON(), `"vertices": ["A", "B", "C"]`, `"vertices": ["A", "B", "C", "D"]`, 1)
	// In 4-gon ABCD, edges are AB, BC, CD, DA. The cross-segment BD is not an edge and not in lengths
	nonexistentEdge := strings.Replace(fourVertexJSON, `"toSegment": "AC"`, `"toSegment": "BD"`, 1)
	if _, err := ValidateGeometryResult(nonexistentEdge, nil); err == nil || !strings.Contains(err.Error(), "nonexistent segment 'BD'") {
		t.Errorf("expected nonexistent segment BD rejection, got: %v", err)
	}

	// FootPoint duplicating existing input vertex "A"
	duplicateFootPoint := strings.Replace(validGeometryJSON(), `"footPoint": "H"`, `"footPoint": "A"`, 1)
	if _, err := ValidateGeometryResult(duplicateFootPoint, nil); err == nil || !strings.Contains(err.Error(), "cannot duplicate existing input vertex") {
		t.Errorf("expected duplicate footPoint rejection, got: %v", err)
	}
}

func TestValidateGeometryResult_NumericInvariants(t *testing.T) {
	// Non-positive length in derivedFacts (BH = -4.0)
	negativeLength := strings.Replace(validGeometryJSON(), `"BH": 4.0`, `"BH": -4.0`, 1)
	if _, err := ValidateGeometryResult(negativeLength, nil); err == nil || !strings.Contains(err.Error(), "must be finite and positive") {
		t.Errorf("expected negative length rejection, got: %v", err)
	}

	// Invalid angle in derivedFacts (> 180 degrees)
	invalidAngle := strings.Replace(validGeometryJSON(), `"ABC": 73.74`, `"ABC": 195.0`, 1)
	if _, err := ValidateGeometryResult(invalidAngle, nil); err == nil || !strings.Contains(err.Error(), "must be within (0, 180) degrees") {
		t.Errorf("expected invalid angle rejection, got: %v", err)
	}

	// Non-positive area metric
	zeroArea := strings.Replace(validGeometryJSON(), `"area": 12.0`, `"area": 0.0`, 1)
	if _, err := ValidateGeometryResult(zeroArea, nil); err == nil || !strings.Contains(err.Error(), "must be finite and positive") {
		t.Errorf("expected zero area metric rejection, got: %v", err)
	}
}

func TestValidateGeometryResult_LaTeXSafety(t *testing.T) {
	// Unbalanced opening brace
	unbalancedOpen := strings.Replace(validGeometryJSON(), `\\frac{AC}{2}`, `\\frac{AC{2}`, 1)
	if _, err := ValidateGeometryResult(unbalancedOpen, nil); err == nil || !strings.Contains(err.Error(), "unbalanced opening brace") {
		t.Errorf("expected unbalanced open brace rejection, got: %v", err)
	}

	// Dangerous TeX command in latexAnswer
	dangerousTeX := strings.Replace(validGeometryJSON(), `BH = 4, \\quad S = 12, \\quad P = 16`, `\\def\\foo{bar} BH = 4`, 1)
	if _, err := ValidateGeometryResult(dangerousTeX, nil); err == nil || !strings.Contains(err.Error(), "forbidden control sequence") {
		t.Errorf("expected forbidden control sequence rejection, got: %v", err)
	}

	// Markdown fences
	fenced := "```json\n" + validGeometryJSON() + "\n```"
	if _, err := ValidateGeometryResult(fenced, nil); err == nil || !strings.Contains(err.Error(), "markdown code fences") {
		t.Errorf("expected markdown code fences rejection, got: %v", err)
	}

	// Raw HTML
	htmlInjected := strings.Replace(validGeometryJSON(), "Altitude BH = 4", "<script>alert('hack')</script>", 1)
	if _, err := ValidateGeometryResult(htmlInjected, nil); err == nil || !strings.Contains(err.Error(), "HTML") {
		t.Errorf("expected HTML injection rejection, got: %v", err)
	}
}
