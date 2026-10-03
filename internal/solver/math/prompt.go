package math

import (
	"encoding/json"
	"fmt"
	"strings"
)

// ExtractProblem extracts the mathematical expression or problem text from the task payload.
func ExtractProblem(payload string) string {
	prob, _ := ExtractProblemAndLanguage(payload)
	return prob
}

// ExtractProblemAndLanguage extracts the problem text and requested language code from the task payload.
func ExtractProblemAndLanguage(payload string) (string, string) {
	trimmed := strings.TrimSpace(payload)
	if trimmed == "" {
		return "", "en"
	}

	problem := trimmed
	lang := "en"

	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(trimmed), &parsed); err == nil {
		lang = extractMathLanguage(parsed)
		if data, ok := parsed["data"].(string); ok && strings.TrimSpace(data) != "" {
			problem = strings.TrimSpace(data)
		} else if prob, ok := parsed["problem"].(string); ok && strings.TrimSpace(prob) != "" {
			problem = strings.TrimSpace(prob)
		} else if expr, ok := parsed["expression"].(string); ok && strings.TrimSpace(expr) != "" {
			problem = strings.TrimSpace(expr)
		}
	}

	if lang == "en" {
		if detectUkrainian(problem) {
			lang = "uk"
		} else if detectRussian(problem) {
			lang = "ru"
		}
	}

	return problem, lang
}

func detectUkrainian(text string) bool {
	lower := strings.ToLower(text)
	for _, r := range text {
		if r == '\u0456' || r == '\u0406' || r == '\u0457' || r == '\u0407' || r == '\u0454' || r == '\u0404' || r == '\u0491' || r == '\u0490' {
			return true
		}
	}
	ukrainianStems := []string{"знайти", "обчислити", "розв'язати", "розвязати", "рівняння", "похідна", "інтеграл", "відповідь"}
	for _, stem := range ukrainianStems {
		if strings.Contains(lower, stem) {
			return true
		}
	}
	return false
}

func detectRussian(text string) bool {
	lower := strings.ToLower(text)
	for _, r := range text {
		if r == '\u044b' || r == '\u042b' || r == '\u044d' || r == '\u042d' || r == '\u044a' || r == '\u042a' || r == '\u0451' || r == '\u0401' {
			return true
		}
	}
	russianStems := []string{"найти", "вычислить", "решить", "уравнение", "производная", "интеграл", "ответ"}
	for _, stem := range russianStems {
		if strings.Contains(lower, stem) {
			return true
		}
	}
	for _, r := range text {
		if (r >= '\u0430' && r <= '\u044f') || (r >= '\u0410' && r <= '\u042f') {
			return true
		}
	}
	return false
}

func extractMathLanguage(parsed map[string]interface{}) string {
	if parsed == nil {
		return "en"
	}
	for _, key := range []string{"language", "lang", "locale"} {
		if val, ok := parsed[key]; ok && val != nil {
			if strVal, ok := val.(string); ok {
				trimmed := strings.ToLower(strings.TrimSpace(strVal))
				if trimmed != "" {
					return normalizeLanguageCode(trimmed)
				}
			}
		}
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

// BuildPrompt constructs the structured-output prompt for Gemini to solve a mathematical problem.
func BuildPrompt(problem string, lang ...string) string {
	targetLang := "en"
	if len(lang) > 0 && lang[0] != "" {
		targetLang = lang[0]
	}
	langName := getLanguageName(targetLang)

	return fmt.Sprintf(`You are the expert Mathematical Problem Solver for GMHelper.

Solve the following mathematical problem step by step and output the solution in the exact JSON format specified below.

Language Requirement (CRITICAL):
- All human-readable natural language text (such as "problem", step "title"s, step "explanation"s, and "finalAnswer") MUST be written in %s.
- Do NOT output explanations in English unless English is the requested language (%s).
- Mathematical formulas and expressions in "latexProblem", "latexFormula", "latexAnswer", and "compositeLatex" must remain valid standard LaTeX notation without word translation.

Problem:
%s

Requirements:
1. Solve the mathematical problem accurately, completely, and rigorously rather than merely restating it.
2. Provide discrete, pedagogical step-by-step reasoning in the "steps" array.
3. Every step in "steps" must have:
   - "stepNumber": integer starting at 1 and strictly incrementing by 1 (1, 2, 3, ...).
   - "title": concise plain-text summary of the algebraic or analytical transformation in %s (e.g. "Subtract 5 from both sides" / "Вычесть 5 из обеих частей").
   - "explanation": clear, thorough plain-text natural-language explanation of why and how this step is performed in %s.
   - "latexFormula": valid KaTeX-compatible LaTeX formula for this transformation.
4. "problem": Plain text representation of the input problem in %s.
5. "latexProblem": KaTeX-formatted representation of the input problem.
6. "finalAnswer": Clear plain-text final answer in %s (e.g. "x = 5").
7. "latexAnswer": KaTeX-formatted final answer (e.g. "x = 5").
8. "compositeLatex": A single multiline LaTeX string joining each step formula with "\\\n" line breaks for frontend KaTeX rendering.
9. "problemType" MUST be "math".
10. "status" MUST be "completed".

Formatting and Security Constraints (CRITICAL):
- Respond ONLY with valid, RFC 8259 compliant JSON.
- All textual fields ("problem", "title", "explanation", "finalAnswer") MUST be STRICT PLAIN TEXT.
- HTML tags, XML, MathML, rich text, and Markdown markup (such as <p>, <div>, <span>, <br>, <b>, <i>, <a>, <strong>, <em>) are STRICTLY FORBIDDEN anywhere in the response.
- Do NOT wrap explanations or any string fields in HTML tags or Markdown formatting.
- Mathematical expressions belong in the dedicated LaTeX fields ("latexProblem", "latexFormula", "latexAnswer", "compositeLatex") using valid KaTeX syntax.
- "compositeLatex" must contain only pure LaTeX/math expressions separated by "\\\n", suitable for direct frontend KaTeX rendering.
- Do NOT wrap output in Markdown code fences (e.g. do NOT use `+"```"+`json or `+"```"+`).
- Do NOT include any commentary, text, or explanations outside the JSON object.
- Do NOT include dangerous TeX control sequences (\def, \let, \write, \input, \catcode).
- Ensure all LaTeX braces '{' and '}' are properly balanced.
- Do NOT include additional arbitrary top-level or nested fields.

JSON Schema:
{
  "problemType": "math",
  "status": "completed",
  "problem": "string",
  "latexProblem": "string",
  "steps": [
    {
      "stepNumber": 1,
      "title": "string",
      "explanation": "string",
      "latexFormula": "string"
    }
  ],
  "finalAnswer": "string",
  "latexAnswer": "string",
  "compositeLatex": "string"
}`, langName, langName, problem, langName, langName, langName, langName)
}
