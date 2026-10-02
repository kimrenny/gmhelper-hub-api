package math

import (
	"strings"
	"testing"
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
      "title": "Divide both sides by 2",
      "explanation": "Divide by coefficient 2 to find x.",
      "latexFormula": "\\frac{2x}{2} = \\frac{10}{2} \\implies x = 5"
    }
  ],
  "finalAnswer": "x = 5",
  "latexAnswer": "x = 5",
  "compositeLatex": "2x + 5 = 15 \\\\\n2x = 10 \\\\\nx = 5"
}`
}

func TestValidateMathResult_Success(t *testing.T) {
	res, err := ValidateMathResult(validMathJSON())
	if err != nil {
		t.Fatalf("expected valid result, got error: %v", err)
	}

	if res.ProblemType != "math" {
		t.Errorf("expected problemType 'math', got '%s'", res.ProblemType)
	}
	if res.Status != "completed" {
		t.Errorf("expected status 'completed', got '%s'", res.Status)
	}
	if len(res.Steps) != 2 {
		t.Errorf("expected 2 steps, got %d", len(res.Steps))
	}
	if res.FinalAnswer != "x = 5" {
		t.Errorf("expected finalAnswer 'x = 5', got '%s'", res.FinalAnswer)
	}
}

func TestValidateMathResult_UnknownTopLevelField(t *testing.T) {
	jsonWithUnknown := `{
  "problemType": "math",
  "status": "completed",
  "problem": "2x + 5 = 15",
  "latexProblem": "2x + 5 = 15",
  "unknownTopLevelField": "should fail",
  "steps": [
    {
      "stepNumber": 1,
      "title": "Subtract 5 from both sides",
      "explanation": "Subtract 5 from both sides to isolate the linear term.",
      "latexFormula": "2x + 5 - 5 = 15 - 5 \\implies 2x = 10"
    }
  ],
  "finalAnswer": "x = 5",
  "latexAnswer": "x = 5",
  "compositeLatex": "2x = 10"
}`

	_, err := ValidateMathResult(jsonWithUnknown)
	if err == nil || !strings.Contains(err.Error(), "unknown field") {
		t.Errorf("expected unknown field rejection error, got: %v", err)
	}
}

func TestValidateMathResult_UnknownStepField(t *testing.T) {
	jsonWithUnknownStepField := `{
  "problemType": "math",
  "status": "completed",
  "problem": "2x + 5 = 15",
  "latexProblem": "2x + 5 = 15",
  "steps": [
    {
      "stepNumber": 1,
      "title": "Subtract 5 from both sides",
      "explanation": "Subtract 5 from both sides to isolate the linear term.",
      "latexFormula": "2x = 10",
      "unexpectedField": "must be rejected"
    }
  ],
  "finalAnswer": "x = 5",
  "latexAnswer": "x = 5",
  "compositeLatex": "2x = 10"
}`

	_, err := ValidateMathResult(jsonWithUnknownStepField)
	if err == nil || !strings.Contains(err.Error(), "unknown field") {
		t.Errorf("expected unknown step field rejection error, got: %v", err)
	}
}

func TestValidateMathResult_ExtraErrorField(t *testing.T) {
	jsonWithErrorField := strings.Replace(validMathJSON(), `"compositeLatex": "2x + 5 = 15 \\\\\n2x = 10 \\\\\nx = 5"`, `"compositeLatex": "2x + 5 = 15 \\\\\n2x = 10 \\\\\nx = 5", "error": "something"`, 1)

	_, err := ValidateMathResult(jsonWithErrorField)
	if err == nil || !strings.Contains(err.Error(), "unknown field") {
		t.Errorf("expected extra error field rejection error, got: %v", err)
	}
}

func TestValidateMathResult_StatusContract(t *testing.T) {
	// Status must be strictly "completed"
	invalidStatuses := []string{"unsolvable", "pending", "failed", "in_progress", "partial", ""}
	for _, st := range invalidStatuses {
		jsonStr := strings.Replace(validMathJSON(), `"status": "completed"`, `"status": "`+st+`"`, 1)
		_, err := ValidateMathResult(jsonStr)
		if err == nil || !strings.Contains(err.Error(), "invalid status") {
			t.Errorf("expected invalid status error for status '%s', got: %v", st, err)
		}
	}
}

func TestValidateMathResult_TrailingData(t *testing.T) {
	trailing := validMathJSON() + ` {"extra": true}`
	_, err := ValidateMathResult(trailing)
	if err == nil || !strings.Contains(err.Error(), "unexpected trailing data") {
		t.Errorf("expected trailing data error, got: %v", err)
	}
}

func TestValidateMathResult_EmptyAndOversized(t *testing.T) {
	_, err := ValidateMathResult("")
	if err == nil {
		t.Errorf("expected error on empty input, got nil")
	}

	oversized := validMathJSON() + strings.Repeat(" ", 70000)
	_, err = ValidateMathResult(oversized)
	if err == nil || !strings.Contains(err.Error(), "exceeds maximum limit") {
		t.Errorf("expected oversized error, got: %v", err)
	}
}

func TestValidateMathResult_MarkdownCodeFences(t *testing.T) {
	fenced := "```json\n" + validMathJSON() + "\n```"
	_, err := ValidateMathResult(fenced)
	if err == nil || !strings.Contains(err.Error(), "code fences") {
		t.Errorf("expected code fences error, got: %v", err)
	}
}

func TestValidateMathResult_HTMLInjection(t *testing.T) {
	htmlSnippets := []string{
		"<script>alert('pwned')</script>",
		"<div>wrapper</div>",
		"<p>paragraph explanation</p>",
		"<span>styled text</span>",
		"Line 1<br>Line 2",
		"Line 1<br/>Line 2",
		"<b>Bold title</b>",
		"<i>Italic explanation</i>",
		"<a href='http://example.com'>link</a>",
		"<strong>Important</strong>",
		"<em>Emphasis</em>",
		"<h1>Header</h1>",
	}

	for _, snippet := range htmlSnippets {
		t.Run(snippet, func(t *testing.T) {
			htmlPayload := strings.Replace(validMathJSON(), "Subtract 5 from both sides to isolate the linear term.", snippet, 1)
			_, err := ValidateMathResult(htmlPayload)
			if err == nil || !strings.Contains(err.Error(), "forbidden HTML elements") {
				t.Errorf("expected HTML rejection error for snippet '%s', got: %v", snippet, err)
			}
		})
	}
}

func TestValidateMathResult_PlainTextExplanation(t *testing.T) {
	plainText := "Subtract 5 from both sides of the equation."
	jsonStr := strings.Replace(validMathJSON(), "Subtract 5 from both sides to isolate the linear term.", plainText, 1)
	res, err := ValidateMathResult(jsonStr)
	if err != nil {
		t.Fatalf("expected valid plain text explanation to pass, got error: %v", err)
	}
	if res.Steps[0].Explanation != plainText {
		t.Errorf("expected explanation '%s', got '%s'", plainText, res.Steps[0].Explanation)
	}
}

func TestValidateMathResult_MathInequalitiesNotTreatedAsHTML(t *testing.T) {
	inequalities := []string{
		"For x < b, where y > 0",
		"When 0 < x < 5 and y > 2",
		"Given i < n and j > 0",
		"Since x < a and z > 1",
		"If p < q and r > s",
	}

	for _, ineq := range inequalities {
		t.Run(ineq, func(t *testing.T) {
			jsonStr := strings.Replace(validMathJSON(), "Subtract 5 from both sides to isolate the linear term.", ineq, 1)
			_, err := ValidateMathResult(jsonStr)
			if err != nil {
				t.Errorf("expected inequality '%s' to pass validation, got error: %v", ineq, err)
			}
		})
	}
}

func TestValidateMathResult_ValidLaTeXCommands(t *testing.T) {
	validFormulas := []struct {
		name    string
		formula string
	}{
		{"linear equation", "2x = 10"},
		{"fractions", "\\frac{2x}{2} = \\frac{10}{2}"},
		{"square roots", "\\sqrt{16} = 4"},
		{"multiplication dot", "2 \\cdot x = 10"},
		{"environment block", "\\begin{matrix} 2x & 10 \\\\ x & 5 \\end{matrix}"},
		{"inequality in LaTeX", "0 < x < 5"},
	}

	for _, tc := range validFormulas {
		t.Run(tc.name, func(t *testing.T) {
			jsonStr := strings.Replace(validMathJSON(), "2x + 5 - 5 = 15 - 5 \\implies 2x = 10", tc.formula, 1)
			_, err := ValidateMathResult(jsonStr)
			if err != nil {
				t.Errorf("expected valid formula '%s' to pass validation, got error: %v", tc.formula, err)
			}
		})
	}
}

func TestValidateMathResult_InvalidJSON(t *testing.T) {
	_, err := ValidateMathResult(`{ "problemType": "math", broken json ...`)
	if err == nil || !strings.Contains(err.Error(), "failed to parse gemini output as JSON") {
		t.Errorf("expected JSON parse error, got: %v", err)
	}
}

func TestValidateMathResult_InvalidProblemType(t *testing.T) {
	jsonStr := strings.Replace(validMathJSON(), `"problemType": "math"`, `"problemType": "geometry"`, 1)
	_, err := ValidateMathResult(jsonStr)
	if err == nil || !strings.Contains(err.Error(), "expected 'math'") {
		t.Errorf("expected invalid problemType error, got: %v", err)
	}
}

func TestValidateMathResult_MissingRequiredFields(t *testing.T) {
	tests := []struct {
		name       string
		target     string
		substitute string
		errMsg     string
	}{
		{"missing problem", `"problem": "2x + 5 = 15"`, `"problem": "   "`, "problem field cannot be empty"},
		{"missing latexProblem", `"latexProblem": "2x + 5 = 15"`, `"latexProblem": ""`, "latexProblem field cannot be empty"},
		{"missing finalAnswer", `"finalAnswer": "x = 5"`, `"finalAnswer": " "`, "finalAnswer field cannot be empty"},
		{"missing latexAnswer", `"latexAnswer": "x = 5"`, `"latexAnswer": ""`, "latexAnswer field cannot be empty"},
		{"missing compositeLatex", `"compositeLatex": "2x + 5 = 15 \\\\\n2x = 10 \\\\\nx = 5"`, `"compositeLatex": ""`, "compositeLatex field cannot be empty"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			jsonStr := strings.Replace(validMathJSON(), tc.target, tc.substitute, 1)
			_, err := ValidateMathResult(jsonStr)
			if err == nil || !strings.Contains(err.Error(), tc.errMsg) {
				t.Errorf("expected error '%s', got: %v", tc.errMsg, err)
			}
		})
	}
}

func TestValidateMathResult_StepsValidation(t *testing.T) {
	// Empty steps
	emptySteps := `{
  "problemType": "math",
  "status": "completed",
  "problem": "1+1",
  "latexProblem": "1+1",
  "steps": [],
  "finalAnswer": "2",
  "latexAnswer": "2",
  "compositeLatex": "1+1=2"
}`
	if _, err := ValidateMathResult(emptySteps); err == nil || !strings.Contains(err.Error(), "steps array cannot be empty") {
		t.Errorf("expected empty steps error, got: %v", err)
	}

	// Non-sequential step numbers (1, 3)
	badSeqSteps := strings.Replace(validMathJSON(), `"stepNumber": 2`, `"stepNumber": 3`, 1)
	if _, err := ValidateMathResult(badSeqSteps); err == nil || !strings.Contains(err.Error(), "invalid stepNumber") {
		t.Errorf("expected step sequence error, got: %v", err)
	}

	// Starting at 2 (2, 3)
	startAtTwo := strings.Replace(validMathJSON(), `"stepNumber": 1`, `"stepNumber": 2`, 1)
	if _, err := ValidateMathResult(startAtTwo); err == nil || !strings.Contains(err.Error(), "invalid stepNumber") {
		t.Errorf("expected step sequence error for start at 2, got: %v", err)
	}

	// Missing step title
	emptyTitle := strings.Replace(validMathJSON(), `"title": "Subtract 5 from both sides"`, `"title": "  "`, 1)
	if _, err := ValidateMathResult(emptyTitle); err == nil || !strings.Contains(err.Error(), "title cannot be empty") {
		t.Errorf("expected empty step title error, got: %v", err)
	}

	// Missing step explanation
	emptyExplanation := strings.Replace(validMathJSON(), `"explanation": "Subtract 5 from both sides to isolate the linear term."`, `"explanation": ""`, 1)
	if _, err := ValidateMathResult(emptyExplanation); err == nil || !strings.Contains(err.Error(), "explanation cannot be empty") {
		t.Errorf("expected empty step explanation error, got: %v", err)
	}

	// Missing step latexFormula
	emptyFormula := strings.Replace(validMathJSON(), `"latexFormula": "2x + 5 - 5 = 15 - 5 \\implies 2x = 10"`, `"latexFormula": " "`, 1)
	if _, err := ValidateMathResult(emptyFormula); err == nil || !strings.Contains(err.Error(), "latexFormula cannot be empty") {
		t.Errorf("expected empty step latexFormula error, got: %v", err)
	}
}

func TestValidateMathResult_LaTeXSafety(t *testing.T) {
	// Unbalanced opening brace
	unbalancedOpen := strings.Replace(validMathJSON(), `\\frac{2x}{2}`, `\\frac{2x{2}`, 1)
	if _, err := ValidateMathResult(unbalancedOpen); err == nil || !strings.Contains(err.Error(), "unbalanced opening brace") {
		t.Errorf("expected unbalanced open brace error, got: %v", err)
	}

	// Unbalanced closing brace
	unbalancedClose := strings.Replace(validMathJSON(), `x = 5`, `x = 5}`, 1)
	if _, err := ValidateMathResult(unbalancedClose); err == nil || !strings.Contains(err.Error(), "unbalanced closing brace") {
		t.Errorf("expected unbalanced closing brace error, got: %v", err)
	}

	// Dangerous TeX commands (JSON-escaped)
	dangerousCommands := []string{`\\def\\foo{bar}`, `\\let\\a\\b`, `\\write18{rm -rf /}`, `\\input{secret}`, `\\catcode65=13`}
	for _, cmd := range dangerousCommands {
		dangerousJSON := strings.Replace(validMathJSON(), `x = 5`, cmd, 1)
		if _, err := ValidateMathResult(dangerousJSON); err == nil || !strings.Contains(err.Error(), "forbidden control sequence") {
			t.Errorf("expected forbidden control sequence error for '%s', got: %v", cmd, err)
		}
	}
}
