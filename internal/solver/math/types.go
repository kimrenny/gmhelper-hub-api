package math

// MathStep represents an individual reasoning step in a mathematical solution.
type MathStep struct {
	StepNumber   int    `json:"stepNumber"`
	Title        string `json:"title"`
	Explanation  string `json:"explanation"`
	LatexFormula string `json:"latexFormula"`
}

// MathResult represents the canonical structured solution for a mathematical problem.
type MathResult struct {
	ProblemType    string     `json:"problemType"`
	Status         string     `json:"status"`
	Problem        string     `json:"problem"`
	LatexProblem   string     `json:"latexProblem"`
	Steps          []MathStep `json:"steps"`
	FinalAnswer    string     `json:"finalAnswer"`
	LatexAnswer    string     `json:"latexAnswer"`
	CompositeLatex string     `json:"compositeLatex"`
}
