package math

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"gmhelper.solution-hub/internal/gemini"
	"gmhelper.solution-hub/internal/solver"
)

// MathSolver solves mathematical problems using Google Gemini and validates the structured response.
type MathSolver struct {
	geminiClient gemini.Client
}

// NewMathSolver creates a new instance of MathSolver.
func NewMathSolver(client gemini.Client) *MathSolver {
	return &MathSolver{
		geminiClient: client,
	}
}

// Solve executes the end-to-end math solving pipeline: problem extraction, prompt generation,
// Gemini inference, unmarshaling, contract validation, and canonical JSON serialization.
func (s *MathSolver) Solve(ctx context.Context, task solver.Task) (solver.Result, error) {
	if s.geminiClient == nil {
		return solver.Result{
			TaskID:      task.TaskID,
			ProblemType: task.ProblemType,
			Success:     false,
		}, errors.New("gemini client is not configured")
	}

	problem := ExtractProblem(task.Payload)
	if strings.TrimSpace(problem) == "" {
		return solver.Result{
			TaskID:      task.TaskID,
			ProblemType: task.ProblemType,
			Success:     false,
		}, errors.New("empty math task payload or problem statement")
	}

	if err := ctx.Err(); err != nil {
		return solver.Result{
			TaskID:      task.TaskID,
			ProblemType: task.ProblemType,
			Success:     false,
		}, fmt.Errorf("math solver context canceled or timed out: %w", err)
	}

	prompt := BuildPrompt(problem)

	rawOutput, err := s.geminiClient.Generate(ctx, prompt)
	if err != nil {
		return solver.Result{
			TaskID:      task.TaskID,
			ProblemType: task.ProblemType,
			Success:     false,
		}, fmt.Errorf("math solver gemini generation failed: %w", err)
	}

	mathResult, err := ValidateMathResult(rawOutput)
	if err != nil {
		return solver.Result{
			TaskID:      task.TaskID,
			ProblemType: task.ProblemType,
			Success:     false,
		}, fmt.Errorf("math solver output validation failed: %w", err)
	}

	normalizedJSON, err := json.Marshal(mathResult)
	if err != nil {
		return solver.Result{
			TaskID:      task.TaskID,
			ProblemType: task.ProblemType,
			Success:     false,
		}, fmt.Errorf("failed to serialize validated math result: %w", err)
	}

	return solver.Result{
		TaskID:      task.TaskID,
		ProblemType: task.ProblemType,
		RawOutput:   string(normalizedJSON),
		Success:     true,
	}, nil
}
