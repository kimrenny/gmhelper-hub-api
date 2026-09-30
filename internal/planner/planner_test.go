package planner

import (
	"context"
	"testing"

	"gmhelper.solution-hub/internal/gemini"
	"gmhelper.solution-hub/internal/solver"
	geoSolver "gmhelper.solution-hub/internal/solver/geometry"
	mathSolver "gmhelper.solution-hub/internal/solver/math"
)

func TestPlanner_GetSolver_Math(t *testing.T) {
	mockGemini := gemini.NewMockClient("ok", nil)
	m := mathSolver.NewMathSolver(mockGemini)
	g := geoSolver.NewGeometrySolver(mockGemini)

	p := NewDefaultPlanner(m, g)

	s, err := p.GetSolver("math")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s == nil {
		t.Fatal("expected solver instance, got nil")
	}

	task := solver.Task{TaskID: "1", ProblemType: "math", Payload: "1+1"}
	res, err := s.Solve(context.Background(), task)
	if err != nil || res.RawOutput != "ok" {
		t.Errorf("solver execution failed: %v", err)
	}
}

func TestPlanner_GetSolver_Geometry(t *testing.T) {
	mockGemini := gemini.NewMockClient("ok", nil)
	m := mathSolver.NewMathSolver(mockGemini)
	g := geoSolver.NewGeometrySolver(mockGemini)

	p := NewDefaultPlanner(m, g)

	s, err := p.GetSolver("geometry")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s == nil {
		t.Fatal("expected solver instance, got nil")
	}
}

func TestPlanner_GetSolver_Unsupported(t *testing.T) {
	mockGemini := gemini.NewMockClient("ok", nil)
	m := mathSolver.NewMathSolver(mockGemini)
	g := geoSolver.NewGeometrySolver(mockGemini)

	p := NewDefaultPlanner(m, g)

	_, err := p.GetSolver("unknown_type")
	if err == nil {
		t.Fatal("expected error for unsupported problem type, got nil")
	}
	if err != ErrUnsupportedProblemType {
		t.Errorf("expected ErrUnsupportedProblemType, got %v", err)
	}
}

func TestChooseStrategy(t *testing.T) {
	if s := ChooseStrategy("math"); s != "math_engine" {
		t.Errorf("expected math_engine, got %s", s)
	}
	if s := ChooseStrategy("geometry"); s != "geometry_engine" {
		t.Errorf("expected geometry_engine, got %s", s)
	}
	if s := ChooseStrategy("unknown"); s != "generic_engine" {
		t.Errorf("expected generic_engine, got %s", s)
	}
}
