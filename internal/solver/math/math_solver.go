package math

import (
	"context"
	"errors"
	"fmt"

	"gmhelper.solution-hub/internal/gemini"
	"gmhelper.solution-hub/internal/solver"
)

type MathSolver struct {
	geminiClient gemini.Client
}

func NewMathSolver(client gemini.Client) *MathSolver {
	return &MathSolver{
		geminiClient: client,
	}
}

func (s *MathSolver) Solve(ctx context.Context, task solver.Task) (solver.Result, error) {
	if s.geminiClient == nil {
		return solver.Result{}, errors.New("gemini client is not configured")
	}

	if task.Payload == "" {
		return solver.Result{}, errors.New("empty math task payload")
	}

	prompt := fmt.Sprintf(
		"You are the GMHelper Mathematical Problem Solver. [ProblemType: math]\n"+
			"Solve the following mathematical problem step-by-step and return structured output:\n\n"+
			"Problem Payload:\n%s\n",
		task.Payload,
	)

	output, err := s.geminiClient.Generate(ctx, prompt)
	if err != nil {
		return solver.Result{
			TaskID:      task.TaskID,
			ProblemType: task.ProblemType,
			Success:     false,
		}, fmt.Errorf("math solver gemini generation failed: %w", err)
	}

	return solver.Result{
		TaskID:      task.TaskID,
		ProblemType: task.ProblemType,
		RawOutput:   output,
		Success:     true,
	}, nil
}
