package geometry

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

// ExtractInputFacts extracts the immutable geometric facts from various payload formats:
// 1. Direct structured JSON containing "inputFacts" or "figures"
// 2. gmhelper-api wrapper: {"task": { ... canvas dict ... }, "given": "..."}
// 3. Canvas figure dictionary: {"triangle_1": {"points": [...], "lines": {...}, "angles": {...}}}
func ExtractInputFacts(payload string) (InputFacts, string, error) {
	trimmed := strings.TrimSpace(payload)
	if trimmed == "" {
		return InputFacts{}, "", fmt.Errorf("empty geometry payload")
	}

	var rawMap map[string]interface{}
	if err := json.Unmarshal([]byte(trimmed), &rawMap); err != nil {
		return InputFacts{}, "", fmt.Errorf("failed to parse geometry payload as JSON: %w", err)
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

	explicitTarget := extractExplicitTarget(rawMap, taskMap, factsObjMap)
	additionalConditions := extractAdditionalConditions(rawMap, taskMap, factsObjMap)

	// Case 1: Payload contains "inputFacts" object
	if factsObj, ok := rawMap["inputFacts"]; ok {
		factsBytes, err := json.Marshal(factsObj)
		if err == nil {
			var facts InputFacts
			if err := json.Unmarshal(factsBytes, &facts); err == nil && len(facts.Figures) > 0 {
				if facts.Lengths == nil {
					facts.Lengths = make(map[string]float64)
				}
				if facts.Angles == nil {
					facts.Angles = make(map[string]float64)
				}
				if facts.ExplicitTarget == "" {
					facts.ExplicitTarget = explicitTarget
				}
				if len(facts.AdditionalConditions) == 0 {
					facts.AdditionalConditions = additionalConditions
				}
				return facts, formatFactsSummary(facts), nil
			}
		}
	}

	// Case 2: Payload directly contains "figures" array
	if _, ok := rawMap["figures"]; ok {
		var facts InputFacts
		if err := json.Unmarshal([]byte(trimmed), &facts); err == nil && len(facts.Figures) > 0 {
			if facts.Lengths == nil {
				facts.Lengths = make(map[string]float64)
			}
			if facts.Angles == nil {
				facts.Angles = make(map[string]float64)
			}
			if facts.ExplicitTarget == "" {
				facts.ExplicitTarget = explicitTarget
			}
			if len(facts.AdditionalConditions) == 0 {
				facts.AdditionalConditions = additionalConditions
			}
			return facts, formatFactsSummary(facts), nil
		}
	}

	// Extract from canvas figure dictionary
	facts := InputFacts{
		Figures:              make([]Figure, 0),
		Lengths:              make(map[string]float64),
		Angles:               make(map[string]float64),
		ExplicitTarget:       explicitTarget,
		AdditionalConditions: additionalConditions,
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

		// Extract lines / lengths
		if linesMap, ok := figMap["lines"].(map[string]interface{}); ok {
			for lineName, lineVal := range linesMap {
				if num, ok := lineVal.(float64); ok && num > 0 {
					facts.Lengths[lineName] = num
				}
			}
		}

		// Extract angles
		if anglesMap, ok := figMap["angles"].(map[string]interface{}); ok {
			for angleName, angleVal := range anglesMap {
				if num, ok := angleVal.(float64); ok && num > 0 {
					facts.Angles[angleName] = num
				}
			}
		}
	}

	if len(facts.Figures) == 0 && len(facts.Lengths) == 0 {
		return InputFacts{}, "", fmt.Errorf("no geometric figures or elements found in payload")
	}

	return facts, formatFactsSummary(facts), nil
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
	for _, fig := range facts.Figures {
		if len(fig.Vertices) > 0 {
			sb.WriteString(fmt.Sprintf("- Figure: %s (%s) with vertices: %s\n", fig.ID, fig.Type, strings.Join(fig.Vertices, ", ")))
		} else {
			sb.WriteString(fmt.Sprintf("- Figure: %s (%s)\n", fig.ID, fig.Type))
		}
	}

	if len(facts.Lengths) > 0 {
		sb.WriteString("Given side lengths:\n")
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
		sb.WriteString("Given angles:\n")
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
		sb.WriteString("Additional geometric constraints:\n")
		for _, cond := range facts.AdditionalConditions {
			sb.WriteString(fmt.Sprintf("  * %s\n", cond))
		}
	}

	if facts.ExplicitTarget != "" {
		sb.WriteString(fmt.Sprintf("Explicit problem goal:\n  * %s\n", facts.ExplicitTarget))
	}

	return sb.String()
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

	return fmt.Sprintf(`You are the expert Geometric Problem Solver for GMHelper.

Solve the following geometry problem step by step and output the solution in the exact JSON format specified below.

Immutable Input Facts:
%s

Input Facts Summary:
%s%s
Requirements:
1. %s
2. Provide discrete, pedagogical step-by-step reasoning in the "steps" array.
3. Every step in "steps" must have:
   - "stepNumber": integer starting at 1 and strictly incrementing by 1 (1, 2, 3, ...).
   - "title": concise name of the theorem or principle applied (e.g. "Pythagorean Theorem").
   - "explanation": clear, thorough natural-language geometric deduction and justification.
   - "latexFormula": valid KaTeX-compatible LaTeX formula for this transformation.
4. "problemStatement": Comprehensive natural language description of the problem and target.
5. "inputFacts": MUST EXACTLY PRESERVE all original figures, lengths, and angles given in the input. Do NOT alter, hallucinate, or overwrite immutable input facts.
6. "target": Structured declaration of target descriptions and target variable names.
7. "derivedFacts": Strongly-typed container for AI deductions:
   - "auxiliaryConstructions": explicit record of auxiliary points/lines constructed by the proof (e.g. altitude foot point H).
   - "lengths": map of calculated segment lengths discovered during the solution.
   - "angles": map of calculated angle degrees discovered during the solution.
   - "metrics": solved global properties (e.g. "area", "perimeter", "radius").
8. "finalAnswer": Clean plain-text final statement.
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
}`, string(factsJSON), summary, metadataSection.String(), requirement1)
}
