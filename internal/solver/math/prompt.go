package math

import (
	"encoding/json"
	"fmt"
	"strings"
)

// ExtractProblem extracts the mathematical expression or problem text from the task payload.
// It handles JSON payloads (e.g. {"data": "2x + 5 = 15"}) as well as plain string payloads.
func ExtractProblem(payload string) string {
	trimmed := strings.TrimSpace(payload)
	if trimmed == "" {
		return ""
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(trimmed), &parsed); err == nil {
		if data, ok := parsed["data"].(string); ok && strings.TrimSpace(data) != "" {
			return strings.TrimSpace(data)
		}
		if prob, ok := parsed["problem"].(string); ok && strings.TrimSpace(prob) != "" {
			return strings.TrimSpace(prob)
		}
		if expr, ok := parsed["expression"].(string); ok && strings.TrimSpace(expr) != "" {
			return strings.TrimSpace(expr)
		}
	}

	return trimmed
}

// BuildPrompt constructs the structured-output prompt for Gemini to solve a mathematical problem.
func BuildPrompt(problem string) string {
	return fmt.Sprintf(`You are the expert Mathematical Problem Solver for GMHelper.

Solve the following mathematical problem step by step and output the solution in the exact JSON format specified below.

Problem:
%s

Requirements:
1. Solve the mathematical problem accurately, completely, and rigorously rather than merely restating it.
2. Provide discrete, pedagogical step-by-step reasoning in the "steps" array.
3. Every step in "steps" must have:
   - "stepNumber": integer starting at 1 and strictly incrementing by 1 (1, 2, 3, ...).
   - "title": concise plain-text summary of the algebraic or analytical transformation (e.g. "Subtract 5 from both sides").
   - "explanation": clear, thorough plain-text natural-language explanation of why and how this step is performed.
   - "latexFormula": valid KaTeX-compatible LaTeX formula for this transformation.
4. "problem": Plain text or unicode representation of the input problem.
5. "latexProblem": KaTeX-formatted representation of the input problem.
6. "finalAnswer": Clear plain-text final answer (e.g. "x = 5").
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
}`, problem)
}
