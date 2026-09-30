package geometry

import (
	"context"
	"errors"
	"fmt"

	"gmhelper.solution-hub/internal/gemini"
	"gmhelper.solution-hub/internal/solver"
)

type GeometrySolver struct {
	geminiClient gemini.Client
}

func NewGeometrySolver(client gemini.Client) *GeometrySolver {
	return &GeometrySolver{
		geminiClient: client,
	}
}

func (s *GeometrySolver) Solve(ctx context.Context, task solver.Task) (solver.Result, error) {
	if s.geminiClient == nil {
		return solver.Result{}, errors.New("gemini client is not configured")
	}

	if task.Payload == "" {
		return solver.Result{}, errors.New("empty geometry task payload")
	}

	prompt := fmt.Sprintf(
		"You are the GMHelper Geometric Problem Solver. [ProblemType: geometry]\n"+
			"Solve the following geometry problem step-by-step using given figure definitions, lines, and angles, and return structured output:\n\n"+
			"Problem Payload:\n%s\n",
		task.Payload,
	)

	output, err := s.geminiClient.Generate(ctx, prompt)
	if err != nil {
		return solver.Result{
			TaskID:      task.TaskID,
			ProblemType: task.ProblemType,
			Success:     false,
		}, fmt.Errorf("geometry solver gemini generation failed: %w", err)
	}

	return solver.Result{
		TaskID:      task.TaskID,
		ProblemType: task.ProblemType,
		RawOutput:   output,
		Success:     true,
	}, nil
}
