package geometry

// Figure represents a geometric figure drawn by the user or supplied in the task.
type Figure struct {
	ID       string   `json:"id"`
	Type     string   `json:"type"`
	Vertices []string `json:"vertices"`
}

// InputFacts represents the immutable geometric ground truth provided in the original task.
type InputFacts struct {
	Figures              []Figure           `json:"figures"`
	Lengths              map[string]float64 `json:"lengths"`
	Angles               map[string]float64 `json:"angles"`
	ExplicitTarget       string             `json:"explicitTarget,omitempty"`
	AdditionalConditions []string           `json:"additionalConditions,omitempty"`
}

// Target represents the unknown quantities or properties requested by the problem.
type Target struct {
	Descriptions []string `json:"descriptions"`
	Variables    []string `json:"variables"`
}

// AuxiliaryConstruction records an auxiliary point, segment, or altitude constructed during solving.
type AuxiliaryConstruction struct {
	Type       string `json:"type"`
	Label      string `json:"label"`
	FromVertex string `json:"fromVertex,omitempty"`
	ToSegment  string `json:"toSegment,omitempty"`
	FootPoint  string `json:"footPoint,omitempty"`
}

// DerivedFacts contains the geometric deductions, auxiliary elements, angles, lengths, and global metrics.
type DerivedFacts struct {
	AuxiliaryConstructions []AuxiliaryConstruction `json:"auxiliaryConstructions,omitempty"`
	Lengths                map[string]float64      `json:"lengths,omitempty"`
	Angles                 map[string]float64      `json:"angles,omitempty"`
	Metrics                map[string]float64      `json:"metrics,omitempty"`
}

// GeometryStep represents an individual logical proof or calculation step.
type GeometryStep struct {
	StepNumber   int    `json:"stepNumber"`
	Title        string `json:"title"`
	Explanation  string `json:"explanation"`
	LatexFormula string `json:"latexFormula"`
}

// GeometryResult represents the canonical structured solution for a geometry problem.
type GeometryResult struct {
	ProblemType      string         `json:"problemType"`
	Status           string         `json:"status"`
	ProblemStatement string         `json:"problemStatement"`
	InputFacts       InputFacts     `json:"inputFacts"`
	Target           Target         `json:"target"`
	DerivedFacts     DerivedFacts   `json:"derivedFacts"`
	Steps            []GeometryStep `json:"steps"`
	FinalAnswer      string         `json:"finalAnswer"`
	LatexAnswer      string         `json:"latexAnswer"`
}
