package planner

import "strings"

func ChooseStrategy(problemType string) string {
	p := strings.ToLower(problemType)

	switch {
	case strings.Contains(p, "math"):
		return "math_engine"
	case strings.Contains(p, "geometry"):
		return "geometry_engine"
	case strings.Contains(p, "logic"):
		return "logic_engine"
	case strings.Contains(p, "text"):
		return "text_engine"
	default:
		return "generic_engine"
	}
}
