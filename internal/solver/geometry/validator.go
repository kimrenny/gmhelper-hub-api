package geometry

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
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

	stopWords = map[string]bool{
		"FIND": true, "CALCULATE": true, "COMPUTE": true, "DETERMINE": true, "SOLVE": true,
		"WHAT": true, "FOR": true, "THE": true, "OF": true, "IN": true, "TO": true, "AND": true,
		"WITH": true, "FROM": true, "BASE": true, "SIDE": true, "TRIANGLE": true, "RECTANGLE": true,
		"SQUARE": true, "CIRCLE": true, "TRAPEZOID": true, "RHOMBUS": true, "PARALLELOGRAM": true,
		"FIGURE": true, "POLYGON": true, "SEGMENT": true, "LINE": true, "POINT": true, "VERTEX": true,
		"LENGTH": true, "VALUE": true, "GIVEN": true, "ITS": true, "EACH": true, "BOTH": true,
		"ЗНАЙТИ": true, "ОБЧИСЛИТИ": true, "ВИЗНАЧИТИ": true, "НАЙТИ": true, "ВЫЧИСЛИТЬ": true, "ОПРЕДЕЛИТЬ": true,
		"BERECHNE": true, "BESTIMME": true, "FINDE": true, "TROUVER": true, "CALCULER": true,
	}

	entityRegex = regexp.MustCompile(`\b[A-Za-z]{1,4}\b`)
)

// ValidateGeometryResult strictly validates raw Gemini output against the canonical Geometry
// contract specification, checking JSON structure, bidirectional input-fact parity,
// numeric invariants, auxiliary reference integrity, step sequentiality, and LaTeX safety.
func ValidateGeometryResult(raw string, expectedFacts *InputFacts) (*GeometryResult, error) {
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
	var res GeometryResult
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
	if res.ProblemType != "geometry" {
		return nil, fmt.Errorf("invalid problemType: expected 'geometry', got '%s'", res.ProblemType)
	}

	// Validate status
	if res.Status != "completed" {
		return nil, fmt.Errorf("invalid status: expected 'completed', got '%s'", res.Status)
	}

	// Validate canonical required strings
	if strings.TrimSpace(res.ProblemStatement) == "" {
		return nil, errors.New("problemStatement field cannot be empty")
	}
	if strings.TrimSpace(res.FinalAnswer) == "" {
		return nil, errors.New("finalAnswer field cannot be empty")
	}
	if strings.TrimSpace(res.LatexAnswer) == "" {
		return nil, errors.New("latexAnswer field cannot be empty")
	}

	// Validate LaTeX in latexAnswer
	if err := validateLaTeX(res.LatexAnswer, "latexAnswer"); err != nil {
		return nil, err
	}

	// Validate Target
	if len(res.Target.Descriptions) == 0 {
		return nil, errors.New("target.descriptions cannot be empty")
	}
	for i, d := range res.Target.Descriptions {
		if strings.TrimSpace(d) == "" {
			return nil, fmt.Errorf("target.descriptions[%d] cannot be empty", i)
		}
	}
	if len(res.Target.Variables) == 0 {
		return nil, errors.New("target.variables cannot be empty")
	}
	for i, v := range res.Target.Variables {
		if strings.TrimSpace(v) == "" {
			return nil, fmt.Errorf("target.variables[%d] cannot be empty", i)
		}
	}

	// Validate target compatibility against authoritative explicit target if present
	if expectedFacts != nil && strings.TrimSpace(expectedFacts.ExplicitTarget) != "" {
		if err := ValidateTargetCompatibility(expectedFacts.ExplicitTarget, res.Target); err != nil {
			return nil, fmt.Errorf("target validation failed: %w", err)
		}
	}

	// Validate DerivedFacts: ensure deductions exist
	if res.DerivedFacts.Metrics == nil {
		res.DerivedFacts.Metrics = make(map[string]float64)
	}
	if len(res.DerivedFacts.Metrics) == 0 && len(res.DerivedFacts.Lengths) == 0 && len(res.DerivedFacts.Angles) == 0 && len(res.DerivedFacts.AuxiliaryConstructions) == 0 {
		return nil, errors.New("derivedFacts cannot be empty: must contain lengths, angles, metrics, or auxiliary constructions")
	}

	// Validate Steps
	if len(res.Steps) == 0 {
		return nil, errors.New("steps array cannot be empty")
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

	// Validate Input Fact Parity against authoritative task facts (both directions)
	if expectedFacts != nil {
		if err := ValidateInputFactParity(*expectedFacts, res.InputFacts); err != nil {
			return nil, fmt.Errorf("input facts parity violation: %w", err)
		}
	}

	// Validate Numeric Invariants (lengths > 0, angles in (0, 180), finite numbers)
	if err := checkNumericInvariants(&res); err != nil {
		return nil, err
	}

	// Validate Auxiliary Construction References and Segments
	if err := checkAuxiliaryConstructions(&res); err != nil {
		return nil, err
	}

	return &res, nil
}

// ValidateInputFactParity verifies that Gemini strictly preserved all immutable user input facts
// without modification, deletion, or fabrication of additional input facts.
func ValidateInputFactParity(expected InputFacts, actual InputFacts) error {
	// 1. Verify figure parity (both directions)
	if len(actual.Figures) != len(expected.Figures) {
		return fmt.Errorf("figure count mismatch: expected %d figures, got %d", len(expected.Figures), len(actual.Figures))
	}

	for _, expFig := range expected.Figures {
		var match *Figure
		for i := range actual.Figures {
			if actual.Figures[i].ID == expFig.ID {
				match = &actual.Figures[i]
				break
			}
		}
		if match == nil {
			return fmt.Errorf("missing expected input figure '%s'", expFig.ID)
		}
		if strings.ToLower(match.Type) != strings.ToLower(expFig.Type) {
			return fmt.Errorf("input figure '%s' type mismatch: expected '%s', got '%s'", expFig.ID, expFig.Type, match.Type)
		}
		if len(expFig.Vertices) > 0 {
			if len(match.Vertices) != len(expFig.Vertices) {
				return fmt.Errorf("input figure '%s' vertex count mismatch: expected %d, got %d", expFig.ID, len(expFig.Vertices), len(match.Vertices))
			}
			for _, v := range expFig.Vertices {
				if !containsString(match.Vertices, v) {
					return fmt.Errorf("input figure '%s' missing vertex '%s'", expFig.ID, v)
				}
			}
		}
	}

	// Check that actual figures contain no unexpected additional figures
	for _, actFig := range actual.Figures {
		found := false
		for _, expFig := range expected.Figures {
			if actFig.ID == expFig.ID {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("unexpected input figure '%s' in inputFacts", actFig.ID)
		}
	}

	// 2. Verify lengths parity (both directions)
	for seg, expLen := range expected.Lengths {
		actLen, found := findSegmentLength(actual.Lengths, seg)
		if !found {
			return fmt.Errorf("missing input length for segment '%s'", seg)
		}
		if math.Abs(actLen-expLen) > 1e-4 {
			return fmt.Errorf("input length mismatch for segment '%s': expected %v, got %v", seg, expLen, actLen)
		}
	}

	// Detect unexpected fabricated lengths in actual.Lengths
	for actSeg, actLen := range actual.Lengths {
		expLen, found := findSegmentLength(expected.Lengths, actSeg)
		if !found {
			return fmt.Errorf("unexpected input length for segment '%s' in inputFacts", actSeg)
		}
		if math.Abs(actLen-expLen) > 1e-4 {
			return fmt.Errorf("input length mismatch for segment '%s': expected %v, got %v", actSeg, expLen, actLen)
		}
	}

	// 3. Verify angles parity (both directions)
	for ang, expAng := range expected.Angles {
		actAng, found := findAngleMeasure(actual.Angles, ang)
		if !found {
			return fmt.Errorf("missing input angle '%s'", ang)
		}
		if math.Abs(actAng-expAng) > 1e-2 {
			return fmt.Errorf("input angle mismatch for angle '%s': expected %v, got %v", ang, expAng, actAng)
		}
	}

	// Detect unexpected fabricated angles in actual.Angles
	for actAng, actMeasure := range actual.Angles {
		expMeasure, found := findAngleMeasure(expected.Angles, actAng)
		if !found {
			return fmt.Errorf("unexpected input angle '%s' in inputFacts", actAng)
		}
		if math.Abs(actMeasure-expMeasure) > 1e-2 {
			return fmt.Errorf("input angle mismatch for angle '%s': expected %v, got %v", actAng, expMeasure, actMeasure)
		}
	}

	return nil
}

func findSegmentLength(lengths map[string]float64, segment string) (float64, bool) {
	if lengths == nil {
		return 0, false
	}
	if v, ok := lengths[segment]; ok {
		return v, true
	}
	// Check reversed segment name e.g. "BA" for "AB"
	if len(segment) == 2 {
		rev := string([]byte{segment[1], segment[0]})
		if v, ok := lengths[rev]; ok {
			return v, true
		}
	}
	return 0, false
}

func findAngleMeasure(angles map[string]float64, angle string) (float64, bool) {
	if angles == nil {
		return 0, false
	}
	if v, ok := angles[angle]; ok {
		return v, true
	}
	// Check reversed angle name e.g. "CBA" for "ABC"
	if len(angle) == 3 {
		rev := string([]byte{angle[2], angle[1], angle[0]})
		if v, ok := angles[rev]; ok {
			return v, true
		}
	}
	return 0, false
}

func checkNumericInvariants(res *GeometryResult) error {
	// Check input lengths
	for k, v := range res.InputFacts.Lengths {
		if math.IsNaN(v) || math.IsInf(v, 0) || v <= 0 {
			return fmt.Errorf("invalid input length for '%s': must be finite and positive (> 0), got %v", k, v)
		}
	}

	// Check input angles
	for k, v := range res.InputFacts.Angles {
		if math.IsNaN(v) || math.IsInf(v, 0) || v <= 0 || v >= 180 {
			return fmt.Errorf("invalid input angle for '%s': must be within (0, 180) degrees, got %v", k, v)
		}
	}

	// Check derived lengths
	for k, v := range res.DerivedFacts.Lengths {
		if math.IsNaN(v) || math.IsInf(v, 0) || v <= 0 {
			return fmt.Errorf("invalid derived length for '%s': must be finite and positive (> 0), got %v", k, v)
		}
	}

	// Check derived angles
	for k, v := range res.DerivedFacts.Angles {
		if math.IsNaN(v) || math.IsInf(v, 0) || v <= 0 || v >= 180 {
			return fmt.Errorf("invalid derived angle for '%s': must be within (0, 180) degrees, got %v", k, v)
		}
	}

	// Check metrics
	for k, v := range res.DerivedFacts.Metrics {
		if math.IsNaN(v) || math.IsInf(v, 0) || v <= 0 {
			return fmt.Errorf("invalid metric '%s': must be finite and positive (> 0), got %v", k, v)
		}
	}

	return nil
}

func checkAuxiliaryConstructions(res *GeometryResult) error {
	// 1. Collect known vertices from input facts
	knownVertices := make(map[string]bool)
	for _, fig := range res.InputFacts.Figures {
		for _, v := range fig.Vertices {
			knownVertices[strings.TrimSpace(v)] = true
		}
	}
	for seg := range res.InputFacts.Lengths {
		for _, ch := range seg {
			knownVertices[string(ch)] = true
		}
	}

	// 2. Collect known segments from input figures and lengths
	knownSegments := make(map[string]bool)
	for _, fig := range res.InputFacts.Figures {
		verts := fig.Vertices
		n := len(verts)
		if n >= 2 {
			for j := 0; j < n; j++ {
				v1 := strings.TrimSpace(verts[j])
				v2 := strings.TrimSpace(verts[(j+1)%n])
				addSegmentSymmetric(knownSegments, v1, v2)
			}
		}
	}
	for seg := range res.InputFacts.Lengths {
		trimmedSeg := strings.TrimSpace(seg)
		if len(trimmedSeg) == 2 {
			v1 := string(trimmedSeg[0])
			v2 := string(trimmedSeg[1])
			addSegmentSymmetric(knownSegments, v1, v2)
		}
	}

	// 3. Process each auxiliary construction in strict sequential order
	for i, c := range res.DerivedFacts.AuxiliaryConstructions {
		if strings.TrimSpace(c.Type) == "" {
			return fmt.Errorf("auxiliary construction %d type cannot be empty", i+1)
		}
		if strings.TrimSpace(c.Label) == "" {
			return fmt.Errorf("auxiliary construction %d label cannot be empty", i+1)
		}

		// Verify fromVertex exists in known vertices
		from := strings.TrimSpace(c.FromVertex)
		if from != "" {
			if len(knownVertices) > 0 && !knownVertices[from] {
				return fmt.Errorf("auxiliary construction %d references unknown fromVertex '%s'", i+1, from)
			}
		}

		// Verify toSegment exists in known segments (not just that its endpoint vertices exist)
		toSeg := strings.TrimSpace(c.ToSegment)
		if toSeg != "" {
			if len(toSeg) == 2 {
				v1 := string(toSeg[0])
				v2 := string(toSeg[1])
				if !knownVertices[v1] || !knownVertices[v2] {
					return fmt.Errorf("auxiliary construction %d references invalid segment vertices in toSegment '%s'", i+1, toSeg)
				}
				if len(knownSegments) > 0 && !isSegmentKnown(knownSegments, v1, v2) {
					return fmt.Errorf("auxiliary construction %d references nonexistent segment '%s'", i+1, toSeg)
				}
			}
		}

		// Verify footPoint
		foot := strings.TrimSpace(c.FootPoint)
		if foot != "" {
			// Foot point must not duplicate an existing input vertex
			if isInputVertex(res.InputFacts, foot) {
				return fmt.Errorf("auxiliary construction %d footPoint '%s' cannot duplicate existing input vertex", i+1, foot)
			}
			knownVertices[foot] = true

			// Establish newly created segments with the footPoint
			if from != "" {
				addSegmentSymmetric(knownSegments, from, foot)
			}
			if toSeg != "" && len(toSeg) == 2 {
				v1 := string(toSeg[0])
				v2 := string(toSeg[1])
				addSegmentSymmetric(knownSegments, v1, foot)
				addSegmentSymmetric(knownSegments, v2, foot)
			}
		}

		// Register label as segment if label is 2 vertices
		label := strings.TrimSpace(c.Label)
		if len(label) == 2 {
			addSegmentSymmetric(knownSegments, string(label[0]), string(label[1]))
		}
	}

	return nil
}

func addSegmentSymmetric(segments map[string]bool, v1, v2 string) {
	if v1 == "" || v2 == "" || v1 == v2 {
		return
	}
	segments[v1+v2] = true
	segments[v2+v1] = true
}

func isSegmentKnown(segments map[string]bool, v1, v2 string) bool {
	return segments[v1+v2] || segments[v2+v1]
}

func isInputVertex(facts InputFacts, vertex string) bool {
	for _, fig := range facts.Figures {
		for _, v := range fig.Vertices {
			if strings.EqualFold(strings.TrimSpace(v), vertex) {
				return true
			}
		}
	}
	return false
}

func containsString(arr []string, target string) bool {
	for _, s := range arr {
		if strings.EqualFold(s, target) {
			return true
		}
	}
	return false
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

// ValidateTargetCompatibility validates that the generated structured target matches the user's explicit goal.
func ValidateTargetCompatibility(explicitTarget string, genTarget Target) error {
	trimmedExplicit := strings.TrimSpace(explicitTarget)
	if trimmedExplicit == "" {
		return nil
	}

	upperExplicit := strings.ToUpper(trimmedExplicit)

	// Combine all generated target descriptions and variables
	var combinedGenVars []string
	for _, v := range genTarget.Variables {
		combinedGenVars = append(combinedGenVars, strings.ToUpper(strings.TrimSpace(v)))
	}
	combinedGenDescs := strings.ToUpper(strings.Join(genTarget.Descriptions, " "))
	combinedGenAll := combinedGenDescs + " " + strings.Join(combinedGenVars, " ")

	// 1. Check for specific geometric entity tokens (e.g. BH, AC, ABC, H)
	tokens := entityRegex.FindAllString(upperExplicit, -1)
	var geometricEntities []string
	for _, tok := range tokens {
		if !stopWords[tok] && len(tok) >= 2 {
			geometricEntities = append(geometricEntities, tok)
		}
	}

	// Check if any geometric entities are present in explicitTarget
	if len(geometricEntities) > 0 {
		matchedEntity := false
		for _, entity := range geometricEntities {
			// Check direct match in variables or descriptions
			if containsEntity(combinedGenVars, entity) || containsEntityWord(combinedGenDescs, entity) {
				matchedEntity = true
				break
			}
			// Check reversed segment or angle (e.g. HB for BH, or CBA for ABC)
			if len(entity) == 2 {
				rev := string([]byte{entity[1], entity[0]})
				if containsEntity(combinedGenVars, rev) || containsEntityWord(combinedGenDescs, rev) {
					matchedEntity = true
					break
				}
			} else if len(entity) == 3 {
				rev := string([]byte{entity[2], entity[1], entity[0]})
				if containsEntity(combinedGenVars, rev) || containsEntityWord(combinedGenDescs, rev) {
					matchedEntity = true
					break
				}
			}
		}

		if !matchedEntity {
			return fmt.Errorf("generated target %v (variables: %v) does not solve explicit target '%s': missing referenced geometric element (%s)",
				genTarget.Descriptions, genTarget.Variables, explicitTarget, strings.Join(geometricEntities, ", "))
		}
	}

	// 2. Metric-specific checks
	explicitWantsArea := containsAnyKeyword(upperExplicit, "AREA", "ПЛОЩ", "FLÄCH", "AIRE", "面積", "면적")
	explicitWantsPerimeter := containsAnyKeyword(upperExplicit, "PERIMETER", "ПЕРИМЕТР", "UMFANG", "PÉRIMÈTRE", "周")

	genHasArea := containsAnyKeyword(combinedGenAll, "AREA", "ПЛОЩ", "FLÄCH", "AIRE") || containsEntity(combinedGenVars, "S", "AREA")
	genHasPerimeter := containsAnyKeyword(combinedGenAll, "PERIMETER", "ПЕРИМЕТР", "UMFANG", "PÉRIMÈTRE") || containsEntity(combinedGenVars, "P", "PERIMETER")

	if explicitWantsArea && !genHasArea {
		return fmt.Errorf("generated target %v (variables: %v) does not solve explicit target '%s': expected area calculation",
			genTarget.Descriptions, genTarget.Variables, explicitTarget)
	}

	if explicitWantsPerimeter && !genHasPerimeter {
		return fmt.Errorf("generated target %v (variables: %v) does not solve explicit target '%s': expected perimeter calculation",
			genTarget.Descriptions, genTarget.Variables, explicitTarget)
	}

	return nil
}

func containsEntity(vars []string, entities ...string) bool {
	for _, v := range vars {
		for _, e := range entities {
			if strings.EqualFold(v, e) || strings.Contains(strings.ToUpper(v), strings.ToUpper(e)) {
				return true
			}
		}
	}
	return false
}

func containsEntityWord(text, entity string) bool {
	pattern := fmt.Sprintf(`\b%s\b`, regexp.QuoteMeta(entity))
	matched, _ := regexp.MatchString(pattern, text)
	return matched
}

func containsAnyKeyword(text string, keywords ...string) bool {
	upper := strings.ToUpper(text)
	for _, kw := range keywords {
		if strings.Contains(upper, strings.ToUpper(kw)) {
			return true
		}
	}
	return false
}
