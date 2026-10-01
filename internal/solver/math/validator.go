package math

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strings"
)

// MaxResultSizeBytes defines the maximum allowed size for Gemini response payloads (64 KB).
const MaxResultSizeBytes = 64 * 1024

var (
	htmlTagRegex = regexp.MustCompile(`(?i)<\s*/?\s*(script|html|body|head|div|span|p|a|iframe|style|table|tr|td|th|ul|ol|li|img|form|input|button|h[1-6]|b|i|strong|em|applet|object|embed)\b[^>]*>`)

	dangerousTeXCommands = []string{
		`\def`, `\let`, `\write`, `\input`, `\catcode`,
		`\openin`, `\openout`, `\newwrite`, `\csname`, `\endcsname`,
		`\newcount`, `\newdimen`, `\newskip`, `\newmuskip`, `\newtoks`,
	}
)

// ValidateMathResult treats raw Gemini output as untrusted and strictly validates it
// against the canonical Math contract specification. It rejects unknown fields,
// enforces strict status semantics, verifies step sequentiality, and checks LaTeX safety.
func ValidateMathResult(raw string) (*MathResult, error) {
	if len(raw) == 0 {
		return nil, errors.New("empty gemini output")
	}

	if len(raw) > MaxResultSizeBytes {
		return nil, fmt.Errorf("gemini response size %d bytes exceeds maximum limit of %d bytes", len(raw), MaxResultSizeBytes)
	}

	trimmed := strings.TrimSpace(raw)

	// Reject markdown code blocks
	if strings.Contains(trimmed, "```") {
		return nil, errors.New("gemini response contains markdown code fences")
	}

	// Reject raw HTML markup
	if htmlTagRegex.MatchString(trimmed) {
		return nil, errors.New("gemini response contains forbidden HTML elements")
	}

	// Strict JSON decoding with unknown field rejection
	var res MathResult
	dec := json.NewDecoder(strings.NewReader(trimmed))
	dec.DisallowUnknownFields()

	if err := dec.Decode(&res); err != nil {
		return nil, fmt.Errorf("failed to parse gemini output as JSON: %w", err)
	}

	// Ensure no trailing tokens after root JSON object
	var trailing json.RawMessage
	if err := dec.Decode(&trailing); err != io.EOF {
		return nil, errors.New("gemini output contains unexpected trailing data")
	}

	// Validate problemType
	if res.ProblemType != "math" {
		return nil, fmt.Errorf("invalid problemType: expected 'math', got '%s'", res.ProblemType)
	}

	// Validate status
	if res.Status != "completed" {
		return nil, fmt.Errorf("invalid status: expected 'completed', got '%s'", res.Status)
	}

	// Validate canonical required strings for status "completed"
	if strings.TrimSpace(res.Problem) == "" {
		return nil, errors.New("problem field cannot be empty")
	}
	if strings.TrimSpace(res.LatexProblem) == "" {
		return nil, errors.New("latexProblem field cannot be empty")
	}
	if strings.TrimSpace(res.FinalAnswer) == "" {
		return nil, errors.New("finalAnswer field cannot be empty")
	}
	if strings.TrimSpace(res.LatexAnswer) == "" {
		return nil, errors.New("latexAnswer field cannot be empty")
	}
	if strings.TrimSpace(res.CompositeLatex) == "" {
		return nil, errors.New("compositeLatex field cannot be empty")
	}

	// Validate LaTeX fields
	if err := validateLaTeX(res.LatexProblem, "latexProblem"); err != nil {
		return nil, err
	}
	if err := validateLaTeX(res.LatexAnswer, "latexAnswer"); err != nil {
		return nil, err
	}
	if err := validateLaTeX(res.CompositeLatex, "compositeLatex"); err != nil {
		return nil, err
	}

	// Validate steps
	if len(res.Steps) == 0 {
		return nil, errors.New("steps array cannot be empty for completed solution")
	}

	for i, step := range res.Steps {
		expectedStepNum := i + 1
		if step.StepNumber != expectedStepNum {
			return nil, fmt.Errorf("invalid stepNumber at step index %d: expected %d, got %d", i, expectedStepNum, step.StepNumber)
		}
		if strings.TrimSpace(step.Title) == "" {
			return nil, fmt.Errorf("step %d title cannot be empty", step.StepNumber)
		}
		if strings.TrimSpace(step.Explanation) == "" {
			return nil, fmt.Errorf("step %d explanation cannot be empty", step.StepNumber)
		}
		if strings.TrimSpace(step.LatexFormula) == "" {
			return nil, fmt.Errorf("step %d latexFormula cannot be empty", step.StepNumber)
		}
		if err := validateLaTeX(step.LatexFormula, fmt.Sprintf("steps[%d].latexFormula", step.StepNumber)); err != nil {
			return nil, err
		}
	}

	return &res, nil
}

// validateLaTeX checks for balanced braces and dangerous control sequences in LaTeX expressions.
func validateLaTeX(latex string, fieldName string) error {
	if err := checkBalancedBraces(latex); err != nil {
		return fmt.Errorf("invalid LaTeX in field '%s': %w", fieldName, err)
	}

	if err := checkDangerousTeX(latex); err != nil {
		return fmt.Errorf("dangerous LaTeX command in field '%s': %w", fieldName, err)
	}

	return nil
}

// checkBalancedBraces verifies that all unescaped opening braces '{' have a matching closing brace '}'.
func checkBalancedBraces(s string) error {
	depth := 0
	inEscape := false

	for i := 0; i < len(s); i++ {
		ch := s[i]
		if inEscape {
			inEscape = false
			continue
		}
		if ch == '\\' {
			inEscape = true
			continue
		}
		if ch == '{' {
			depth++
		} else if ch == '}' {
			depth--
			if depth < 0 {
				return errors.New("unbalanced closing brace '}'")
			}
		}
	}

	if depth != 0 {
		return errors.New("unbalanced opening brace '{'")
	}

	return nil
}

// checkDangerousTeX verifies that the LaTeX expression does not contain hazardous macro/primitive definitions.
func checkDangerousTeX(s string) error {
	for _, cmd := range dangerousTeXCommands {
		if strings.Contains(s, cmd) {
			return fmt.Errorf("forbidden control sequence '%s' detected", cmd)
		}
	}
	return nil
}
