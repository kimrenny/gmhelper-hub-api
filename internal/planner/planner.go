package planner

import (
	"errors"
	"strings"

	"gmhelper.solution-hub/internal/solver"
)

var ErrUnsupportedProblemType = errors.New("unsupported problem type")

type Planner interface {
	GetSolver(problemType string) (solver.Solver, error)
}

type DefaultPlanner struct {
	mathSolver     solver.Solver
	geometrySolver solver.Solver
}

func NewDefaultPlanner(mathSolver solver.Solver, geometrySolver solver.Solver) *DefaultPlanner {
	return &DefaultPlanner{
		mathSolver:     mathSolver,
		geometrySolver: geometrySolver,
	}
}

func (p *DefaultPlanner) GetSolver(problemType string) (solver.Solver, error) {
	pt := strings.ToLower(strings.TrimSpace(problemType))

	switch {
	case strings.Contains(pt, "math"):
		if p.mathSolver == nil {
			return nil, errors.New("math solver is not configured")
		}
		return p.mathSolver, nil
	case strings.Contains(pt, "geo"):
		if p.geometrySolver == nil {
			return nil, errors.New("geometry solver is not configured")
		}
		return p.geometrySolver, nil
	default:
		return nil, ErrUnsupportedProblemType
	}
}

// ChooseStrategy returns the strategy identifier for logging/diagnostics.
func ChooseStrategy(problemType string) string {
	p := strings.ToLower(problemType)

	switch {
	case strings.Contains(p, "math"):
		return "math_engine"
	case strings.Contains(p, "geo"):
		return "geometry_engine"
	case strings.Contains(p, "logic"):
		return "logic_engine"
	case strings.Contains(p, "text"):
		return "text_engine"
	default:
		return "generic_engine"
	}
}
