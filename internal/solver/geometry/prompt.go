package geometry

import (
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strings"
)

var (
	textLengthRegex = regexp.MustCompile(`\b([A-Za-z]{2})\s*=\s*([0-9]+(?:\.[0-9]+)?)\b`)
	textAngleRegex  = regexp.MustCompile(`(?i)(?:∠|\\angle|angle|угол|кут)?\s*([A-Za-z]{3})\s*=\s*([0-9]+(?:\.[0-9]+)?)\s*°?`)
	textTargetRegex = regexp.MustCompile(`(?i)(?:find|calculate|compute|determine|знайти|обчислити|найти|вычислить|berechne|finde|trouver|calculer)\s+([A-Za-z]{1,4})\b`)
)

// ExtractInputFacts extracts geometric facts with strict source precedence:
// 1. Explicit textual problem statement
// 2. Explicit additional conditions
// 3. Explicit target
// 4. Structured semantic metadata (lengths, angles from text)
// 5. Visual figure topology / labels (schematic visual context)
// 6. Canvas coordinates / visual dimensions (never authoritative)
func ExtractInputFacts(payload string) (InputFacts, string, error) {
	trimmed := strings.TrimSpace(payload)
	if trimmed == "" {
		return InputFacts{}, "", fmt.Errorf("empty geometry payload")
	}

	var rawMap map[string]interface{}
	if err := json.Unmarshal([]byte(trimmed), &rawMap); err != nil {
		// Non-JSON plain text payload: treat entire text as problem statement
		problemText := trimmed
		explicitTarget := extractTargetFromText(problemText)
		extractedLengths := extractLengthsFromText(problemText)
		extractedAngles := extractAnglesFromText(problemText)
		language := extractLanguage(nil, nil, nil, explicitTarget, []string{problemText}, problemText)

		facts := InputFacts{
			Problem:              problemText,
			Figures:              make([]Figure, 0),
			Lengths:              extractedLengths,
			Angles:               extractedAngles,
			ExplicitTarget:       explicitTarget,
			AdditionalConditions: make([]string, 0),
			Language:             language,
		}
		return facts, formatFactsSummary(facts), nil
	}

	targetMap := rawMap
	var taskMap map[string]interface{}
	if taskVal, ok := rawMap["task"].(map[string]interface{}); ok {
		targetMap = taskVal
		taskMap = taskVal
	}

	var factsObjMap map[string]interface{}
	if obj, ok := rawMap["inputFacts"].(map[string]interface{}); ok {
		factsObjMap = obj
	} else if taskMap != nil {
		if obj, ok := taskMap["inputFacts"].(map[string]interface{}); ok {
			factsObjMap = obj
		}
	}

	problemText := extractProblemText(rawMap, taskMap, factsObjMap)
	explicitTarget := extractExplicitTarget(rawMap, taskMap, factsObjMap)
	additionalConditions := extractAdditionalConditions(rawMap, taskMap, factsObjMap)
	language := extractLanguage(rawMap, taskMap, factsObjMap, explicitTarget, additionalConditions, problemText)

	// Case 1: Payload contains "inputFacts" object
	if factsObj, ok := rawMap["inputFacts"]; ok {
		factsBytes, err := json.Marshal(factsObj)
		if err == nil {
			var facts InputFacts
			if err := json.Unmarshal(factsBytes, &facts); err == nil && (len(facts.Figures) > 0 || len(facts.Lengths) > 0 || explicitTarget != "" || len(additionalConditions) > 0 || problemText != "" || facts.Problem != "") {
				if facts.Lengths == nil {
					facts.Lengths = make(map[string]float64)
				}
				if facts.Angles == nil {
					facts.Angles = make(map[string]float64)
				}
				if facts.Problem == "" {
					facts.Problem = problemText
				}
				if facts.ExplicitTarget == "" {
					facts.ExplicitTarget = explicitTarget
				}
				if len(facts.AdditionalConditions) == 0 {
					facts.AdditionalConditions = additionalConditions
				}
				if facts.Language == "" {
					facts.Language = language
				}

				// Override visual lengths with authoritative textual lengths
				overrideWithTextualFacts(&facts)

				return facts, formatFactsSummary(facts), nil
			}
		}
	}

	// Case 2: Payload directly contains "figures" array
	if _, ok := rawMap["figures"]; ok {
		var facts InputFacts
		if err := json.Unmarshal([]byte(trimmed), &facts); err == nil && (len(facts.Figures) > 0 || len(facts.Lengths) > 0 || explicitTarget != "" || len(additionalConditions) > 0 || problemText != "" || facts.Problem != "") {
			if facts.Lengths == nil {
				facts.Lengths = make(map[string]float64)
			}
			if facts.Angles == nil {
				facts.Angles = make(map[string]float64)
			}
			if facts.Problem == "" {
				facts.Problem = problemText
			}
			if facts.ExplicitTarget == "" {
				facts.ExplicitTarget = explicitTarget
			}
			if len(facts.AdditionalConditions) == 0 {
				facts.AdditionalConditions = additionalConditions
			}
			if facts.Language == "" {
				facts.Language = language
			}

			// Override visual lengths with authoritative textual lengths
			overrideWithTextualFacts(&facts)

			return facts, formatFactsSummary(facts), nil
		}
	}

	// Extract from canvas figure dictionary (optional visual context)
	facts := InputFacts{
		Problem:              problemText,
		Figures:              make([]Figure, 0),
		Lengths:              make(map[string]float64),
		Angles:               make(map[string]float64),
		ExplicitTarget:       explicitTarget,
		AdditionalConditions: additionalConditions,
		Language:             language,
	}

	// Sort keys for deterministic extraction
	keys := make([]string, 0, len(targetMap))
	for k := range targetMap {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	ignoredMetadataKeys := map[string]bool{
		"target":               true,
		"explicittarget":       true,
		"goal":                 true,
		"additionalconditions": true,
		"conditions":           true,
		"task":                 true,
		"given":                true,
		"solution":             true,
		"answer":               true,
		"inputfacts":           true,
		"figures":              true,
		"language":             true,
		"lang":                 true,
		"locale":               true,
		"problem":              true,
		"problemstatement":     true,
		"statement":            true,
		"text":                 true,
		"prompt":               true,
		"data":                 true,
	}

	for _, figID := range keys {
		if ignoredMetadataKeys[strings.ToLower(figID)] {
			continue
		}

		figVal := targetMap[figID]
		figMap, ok := figVal.(map[string]interface{})
		if !ok {
			continue
		}

		typePrefix := strings.ToLower(strings.Split(figID, "_")[0])
		if typePrefix == "pencil" {
			continue
		}

		var vertices []string
		if pointsArr, ok := figMap["points"].([]interface{}); ok {
			for _, pVal := range pointsArr {
				if pMap, ok := pVal.(map[string]interface{}); ok {
					if label, ok := pMap["label"].(string); ok && strings.TrimSpace(label) != "" {
						vertices = append(vertices, strings.TrimSpace(label))
					}
				} else if labelStr, ok := pVal.(string); ok && strings.TrimSpace(labelStr) != "" {
					vertices = append(vertices, strings.TrimSpace(labelStr))
				}
			}
		} else if pointsMap, ok := figMap["points"].(map[string]interface{}); ok {
			for pKey, pVal := range pointsMap {
				if pMap, ok := pVal.(map[string]interface{}); ok {
					if label, ok := pMap["label"].(string); ok && strings.TrimSpace(label) != "" {
						vertices = append(vertices, strings.TrimSpace(label))
					} else if strings.TrimSpace(pKey) != "" {
						vertices = append(vertices, strings.TrimSpace(pKey))
					}
				} else if labelStr, ok := pVal.(string); ok && strings.TrimSpace(labelStr) != "" {
					vertices = append(vertices, strings.TrimSpace(labelStr))
				} else if strings.TrimSpace(pKey) != "" {
					vertices = append(vertices, strings.TrimSpace(pKey))
				}
			}
		}

		// Deduce vertices from lines if vertices array was empty
		if len(vertices) == 0 {
			vSet := make(map[string]bool)
			if linesMap, ok := figMap["lines"].(map[string]interface{}); ok {
				for lineName := range linesMap {
					for _, ch := range lineName {
						vSet[string(ch)] = true
					}
				}
			}
			for v := range vSet {
				vertices = append(vertices, v)
			}
			sort.Strings(vertices)
		}

		figureType := typePrefix
		if figureType == "rectangle" && allSidesEqual(figMap) {
			figureType = "square"
		}

		facts.Figures = append(facts.Figures, Figure{
			ID:       figID,
			Type:     figureType,
			Vertices: vertices,
		})

		// Extract visual lines / lengths from canvas
		if linesMap, ok := figMap["lines"].(map[string]interface{}); ok {
			for lineName, lineVal := range linesMap {
				if num, ok := lineVal.(float64); ok && num > 0 {
					facts.Lengths[strings.ToUpper(strings.TrimSpace(lineName))] = num
				}
			}
		}

		// Extract visual angles from canvas
		if anglesMap, ok := figMap["angles"].(map[string]interface{}); ok {
			for angleName, angleVal := range anglesMap {
				if num, ok := angleVal.(float64); ok && num > 0 {
					facts.Angles[strings.ToUpper(strings.TrimSpace(angleName))] = num
				}
			}
		}
	}

	// Apply source precedence: explicit textual facts override drawing measurements
	overrideWithTextualFacts(&facts)

	if len(facts.Figures) == 0 && len(facts.Lengths) == 0 && facts.ExplicitTarget == "" && len(facts.AdditionalConditions) == 0 && strings.TrimSpace(facts.Problem) == "" {
		return InputFacts{}, "", fmt.Errorf("no geometric figures, conditions, problem statement, or target found in payload")
	}

	return facts, formatFactsSummary(facts), nil
}

func extractProblemText(maps ...map[string]interface{}) string {
	for _, m := range maps {
		if m == nil {
			continue
		}
		for _, key := range []string{"problem", "problemStatement", "statement", "text", "prompt", "given", "data"} {
			if val, ok := m[key]; ok && val != nil {
				if strVal, ok := val.(string); ok {
					trimmed := strings.TrimSpace(strVal)
					if trimmed != "" {
						return trimmed
					}
				}
			}
		}
	}
	return ""
}

func overrideWithTextualFacts(facts *InputFacts) {
	combinedText := facts.Problem + " " + strings.Join(facts.AdditionalConditions, " ")
	if facts.ExplicitTarget != "" {
		combinedText += " " + facts.ExplicitTarget
	}

	textLengths := extractLengthsFromText(combinedText)
	for seg, lVal := range textLengths {
		facts.Lengths[seg] = lVal
	}

	textAngles := extractAnglesFromText(combinedText)
	for ang, aVal := range textAngles {
		facts.Angles[ang] = aVal
	}
}

func extractLengthsFromText(text string) map[string]float64 {
	res := make(map[string]float64)
	matches := textLengthRegex.FindAllStringSubmatch(text, -1)
	for _, m := range matches {
		if len(m) == 3 {
			seg := strings.ToUpper(strings.TrimSpace(m[1]))
			var val float64
			if _, err := fmt.Sscanf(m[2], "%f", &val); err == nil && val > 0 {
				res[seg] = val
			}
		}
	}
	return res
}

func extractAnglesFromText(text string) map[string]float64 {
	res := make(map[string]float64)
	matches := textAngleRegex.FindAllStringSubmatch(text, -1)
	for _, m := range matches {
		if len(m) == 3 {
			ang := strings.ToUpper(strings.TrimSpace(m[1]))
			var val float64
			if _, err := fmt.Sscanf(m[2], "%f", &val); err == nil && val > 0 && val < 180 {
				res[ang] = val
			}
		}
	}
	return res
}

func extractTargetFromText(text string) string {
	match := textTargetRegex.FindStringSubmatch(text)
	if len(match) == 2 {
		return "Find " + strings.ToUpper(strings.TrimSpace(match[1]))
	}
	return ""
}

func extractExplicitTarget(maps ...map[string]interface{}) string {
	for _, m := range maps {
		if m == nil {
			continue
		}
		for _, key := range []string{"target", "explicitTarget", "goal"} {
			if val, ok := m[key]; ok && val != nil {
				if strVal, ok := val.(string); ok {
					trimmed := strings.TrimSpace(strVal)
					if trimmed != "" {
						return trimmed
					}
				}
			}
		}
	}
	return ""
}

func extractAdditionalConditions(maps ...map[string]interface{}) []string {
	var result []string
	seen := make(map[string]bool)

	for _, m := range maps {
		if m == nil {
			continue
		}
		for _, key := range []string{"additionalConditions", "conditions"} {
			if val, ok := m[key]; ok && val != nil {
				if sliceVal, ok := val.([]interface{}); ok {
					for _, item := range sliceVal {
						if strVal, ok := item.(string); ok {
							trimmed := strings.TrimSpace(strVal)
							if trimmed != "" && !seen[trimmed] {
								seen[trimmed] = true
								result = append(result, trimmed)
							}
						}
					}
				} else if strSliceVal, ok := val.([]string); ok {
					for _, item := range strSliceVal {
						trimmed := strings.TrimSpace(item)
						if trimmed != "" && !seen[trimmed] {
							seen[trimmed] = true
							result = append(result, trimmed)
						}
					}
				}
			}
		}
	}
	return result
}

func allSidesEqual(figMap map[string]interface{}) bool {
	linesMap, ok := figMap["lines"].(map[string]interface{})
	if !ok || len(linesMap) < 2 {
		return false
	}
	var firstVal float64
	first := true
	for _, v := range linesMap {
		num, ok := v.(float64)
		if !ok {
			return false
		}
		if first {
			firstVal = num
			first = false
		} else if num != firstVal {
			return false
		}
	}
	return true
}

func formatFactsSummary(facts InputFacts) string {
	var sb strings.Builder

	if strings.TrimSpace(facts.Problem) != "" {
		sb.WriteString(fmt.Sprintf("Problem Statement (AUTHORITATIVE):\n  * %s\n", facts.Problem))
	}

	if len(facts.Figures) > 0 {
		sb.WriteString("Diagram Figures (Optional / Schematic Visual Context):\n")
		for _, fig := range facts.Figures {
			if len(fig.Vertices) > 0 {
				sb.WriteString(fmt.Sprintf("  - Figure: %s (%s) with vertices: %s\n", fig.ID, fig.Type, strings.Join(fig.Vertices, ", ")))
			} else {
				sb.WriteString(fmt.Sprintf("  - Figure: %s (%s)\n", fig.ID, fig.Type))
			}
		}
	} else {
		sb.WriteString("Diagram: None supplied (Solve from textual problem and mathematical conditions alone).\n")
	}

	if len(facts.Lengths) > 0 {
		sb.WriteString("Given side lengths (Authoritative):\n")
		keys := make([]string, 0, len(facts.Lengths))
		for k := range facts.Lengths {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			sb.WriteString(fmt.Sprintf("  * %s = %v\n", k, facts.Lengths[k]))
		}
	}

	if len(facts.Angles) > 0 {
		sb.WriteString("Given angles (Authoritative):\n")
		keys := make([]string, 0, len(facts.Angles))
		for k := range facts.Angles {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			sb.WriteString(fmt.Sprintf("  * ∠%s = %v°\n", k, facts.Angles[k]))
		}
	}

	if len(facts.AdditionalConditions) > 0 {
		sb.WriteString("Additional geometric constraints (Authoritative):\n")
		for _, cond := range facts.AdditionalConditions {
			sb.WriteString(fmt.Sprintf("  * %s\n", cond))
		}
	}

	if facts.ExplicitTarget != "" {
		sb.WriteString(fmt.Sprintf("Explicit problem goal (Authoritative):\n  * %s\n", facts.ExplicitTarget))
	}

	return sb.String()
}

func extractLanguage(rawMap, taskMap, factsObjMap map[string]interface{}, explicitTarget string, additionalConditions []string, extraTexts ...string) string {
	for _, m := range []map[string]interface{}{rawMap, taskMap, factsObjMap} {
		if m == nil {
			continue
		}
		for _, key := range []string{"language", "lang", "locale"} {
			if val, ok := m[key]; ok && val != nil {
				if strVal, ok := val.(string); ok {
					trimmed := strings.ToLower(strings.TrimSpace(strVal))
					if trimmed != "" {
						return normalizeLanguageCode(trimmed)
					}
				}
			}
		}
	}

	// Auto-detect language from explicitTarget, additionalConditions, and extra problem texts
	combinedText := explicitTarget + " " + strings.Join(additionalConditions, " ")
	for _, t := range extraTexts {
		combinedText += " " + t
	}

	if detectUkrainian(combinedText) {
		return "uk"
	}
	if detectRussian(combinedText) {
		return "ru"
	}
	if detectGerman(combinedText) {
		return "de"
	}
	if detectFrench(combinedText) {
		return "fr"
	}

	return "en"
}

func normalizeLanguageCode(code string) string {
	lower := strings.ToLower(strings.TrimSpace(code))
	switch {
	case strings.HasPrefix(lower, "ru"):
		return "ru"
	case strings.HasPrefix(lower, "uk"), strings.HasPrefix(lower, "ua"):
		return "uk"
	case strings.HasPrefix(lower, "de"):
		return "de"
	case strings.HasPrefix(lower, "fr"):
		return "fr"
	case strings.HasPrefix(lower, "ja"):
		return "ja"
	case strings.HasPrefix(lower, "ko"):
		return "ko"
	case strings.HasPrefix(lower, "zh"):
		return "zh"
	default:
		return "en"
	}
}

func getLanguageName(lang string) string {
	switch normalizeLanguageCode(lang) {
	case "ru":
		return "Russian (Русский)"
	case "uk":
		return "Ukrainian (Українська)"
	case "de":
		return "German (Deutsch)"
	case "fr":
		return "French (Français)"
	case "ja":
		return "Japanese (日本語)"
	case "ko":
		return "Korean (한국어)"
	case "zh":
		return "Chinese (中文)"
	default:
		return "English"
	}
}

func detectUkrainian(text string) bool {
	lower := strings.ToLower(text)
	// Specific Ukrainian runes: і (\u0456), ї (\u0457), є (\u0454), ґ (\u0491)
	for _, r := range text {
		if r == '\u0456' || r == '\u0406' || r == '\u0457' || r == '\u0407' || r == '\u0454' || r == '\u0404' || r == '\u0491' || r == '\u0490' {
			return true
		}
	}
	// Common Ukrainian word stems
	ukrainianStems := []string{"знайти", "обчислити", "висота", "висоту", "трикутник", "площа", "довжина", "перпендикуляр"}
	for _, stem := range ukrainianStems {
		if strings.Contains(lower, stem) {
			return true
		}
	}
	return false
}

func detectRussian(text string) bool {
	lower := strings.ToLower(text)
	// Specific Russian runes: ы (\u044b), э (\u044d), ъ (\u044a), ё (\u0451)
	for _, r := range text {
		if r == '\u044b' || r == '\u042b' || r == '\u044d' || r == '\u042d' || r == '\u044a' || r == '\u042a' || r == '\u0451' || r == '\u0401' {
			return true
		}
	}
	// Common Russian word stems
	russianStems := []string{"найти", "вычислить", "высота", "высоту", "треугольник", "площадь", "длина", "перпендикуляр"}
	for _, stem := range russianStems {
		if strings.Contains(lower, stem) {
			return true
		}
	}
	// General Cyrillic check
	for _, r := range text {
		if (r >= '\u0430' && r <= '\u044f') || (r >= '\u0410' && r <= '\u042f') {
			return true
		}
	}
	return false
}

func detectGerman(text string) bool {
	lower := strings.ToLower(text)
	return strings.Contains(lower, "berechne") || strings.Contains(lower, "finde") || strings.Contains(lower, "fläch") || strings.Contains(lower, "dreieck")
}

func detectFrench(text string) bool {
	lower := strings.ToLower(text)
	return strings.Contains(lower, "trouver") || strings.Contains(lower, "calculer") || strings.Contains(lower, "triangle") && strings.Contains(lower, "aire")
}

// BuildPrompt constructs the structured-output prompt for Gemini to solve a geometric problem.
func BuildPrompt(facts InputFacts, summary string) string {
	factsJSON, _ := json.MarshalIndent(facts, "", "  ")

	var metadataSection strings.Builder
	if facts.ExplicitTarget != "" {
		metadataSection.WriteString(fmt.Sprintf("\nExplicit Problem Goal (AUTHORITATIVE - you MUST solve this exact goal):\n%s\n", facts.ExplicitTarget))
	}
	if len(facts.AdditionalConditions) > 0 {
		metadataSection.WriteString("\nAdditional Geometric Constraints:\n")
		for _, cond := range facts.AdditionalConditions {
			metadataSection.WriteString(fmt.Sprintf("- %s\n", cond))
		}
	}

	var requirement1 string
	if facts.ExplicitTarget != "" {
		requirement1 = fmt.Sprintf("Solve the explicitly specified problem goal (\"%s\"). Do NOT replace it with another geometric objective or solve for an unrelated metric unless needed as an intermediate step to find the target.", facts.ExplicitTarget)
	} else {
		requirement1 = "Solve the geometry problem accurately, completely, and rigorously rather than merely restating it."
	}

	langName := getLanguageName(facts.Language)

	return fmt.Sprintf(`You are the expert Geometric Problem Solver for GMHelper.

Solve the following geometry problem step by step and output the solution in the exact JSON format specified below.

Mathematical Authority and Source Precedence (CRITICAL):
- The problem statement, explicit textual conditions, and explicit target are the AUTHORITATIVE source of mathematical facts.
- Any supplied diagram is optional visual context. Treat drawings as schematic unless the text explicitly establishes dimensions.
- Never infer an exact length, angle, ratio, or other mathematical measurement from canvas/pixel coordinates or visual proportions.
- If textual data and drawing metadata conflict, ALWAYS use the textual value (explicit text > drawing measurements).
- If no diagram/figure is supplied, solve the geometry problem from the textual information alone.
- Solve the requested problem goal accurately. Do NOT invent a different target or replace the task with a simpler problem.

Language Requirement (CRITICAL):
- All human-readable natural language text MUST be written in %s.
- This applies to: "problemStatement", "target.descriptions", step "title"s, step "explanation"s, and "finalAnswer".
- Do NOT output explanations in English unless English is the requested language (%s).
- Mathematical entities, variable names (e.g. AB, BH, ∠ABC), and LaTeX expressions MUST remain in standard mathematical notation and NOT translated.

Immutable Input Facts:
%s

Input Facts Summary:
%s%s
Requirements:
1. %s
2. Provide discrete, pedagogical step-by-step reasoning in the "steps" array.
3. Every step in "steps" must have:
   - "stepNumber": integer starting at 1 and strictly incrementing by 1 (1, 2, 3, ...).
   - "title": concise name of the theorem or principle applied in %s (e.g. "Теорема Пифагора" / "Pythagorean Theorem").
   - "explanation": clear, thorough natural-language geometric deduction and justification in %s.
   - "latexFormula": valid KaTeX-compatible LaTeX formula for this transformation.
4. "problemStatement": Comprehensive natural language description of the problem and target in %s.
5. "inputFacts": MUST preserve original immutable given facts. If no diagram was provided, figures array may be empty or represent the problem's conceptual figures.
6. "target": Structured declaration of target descriptions (in %s) and target variable names.
7. "derivedFacts": Strongly-typed container for AI deductions:
   - "auxiliaryConstructions": explicit record of auxiliary points/lines constructed by the proof (e.g. altitude foot point H).
   - "lengths": map of calculated segment lengths discovered during the solution.
   - "angles": map of calculated angle degrees discovered during the solution.
   - "metrics": solved global properties (e.g. "area", "perimeter", "radius").
8. "finalAnswer": Clean plain-text final statement in %s.
9. "latexAnswer": KaTeX-formatted final answer.
10. "problemType" MUST be "geometry".
11. "status" MUST be "completed".

Output Constraints (CRITICAL):
- Respond ONLY with valid, RFC 8259 compliant JSON.
- Do NOT wrap output in Markdown code fences (e.g. do NOT use `+"```"+`json or `+"```"+`).
- Do NOT include any text, HTML, or explanations outside the JSON object.
- Do NOT include dangerous TeX control sequences (\def, \let, \write, \input, \catcode).
- Ensure all LaTeX braces '{' and '}' are properly balanced.
- Do NOT include additional arbitrary top-level or nested fields.

JSON Schema:
{
  "problemType": "geometry",
  "status": "completed",
  "problemStatement": "string",
  "inputFacts": {
    "figures": [
      {
        "id": "string",
        "type": "string",
        "vertices": ["string"]
      }
    ],
    "lengths": {
      "string": 0.0
    },
    "angles": {
      "string": 0.0
    }
  },
  "target": {
    "descriptions": ["string"],
    "variables": ["string"]
  },
  "derivedFacts": {
    "auxiliaryConstructions": [
      {
        "type": "string",
        "label": "string",
        "fromVertex": "string",
        "toSegment": "string",
        "footPoint": "string"
      }
    ],
    "lengths": {
      "string": 0.0
    },
    "angles": {
      "string": 0.0
    },
    "metrics": {
      "string": 0.0
    }
  },
  "steps": [
    {
      "stepNumber": 1,
      "title": "string",
      "explanation": "string",
      "latexFormula": "string"
    }
  ],
  "finalAnswer": "string",
  "latexAnswer": "string"
}`, langName, langName, string(factsJSON), summary, metadataSection.String(), requirement1, langName, langName, langName, langName, langName)
}
