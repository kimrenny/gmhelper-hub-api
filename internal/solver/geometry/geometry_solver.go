package geometry

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"gmhelper.solution-hub/internal/gemini"
	"gmhelper.solution-hub/internal/solver"
)

// GeometrySolver solves geometric problems using Google Gemini and strictly validates the output.
type GeometrySolver struct {
	geminiClient gemini.Client
}

// NewGeometrySolver creates a new instance of GeometrySolver.
func NewGeometrySolver(client gemini.Client) *GeometrySolver {
	return &GeometrySolver{
		geminiClient: client,
	}
}

// Solve executes the geometry solving pipeline: fact extraction, prompt construction,
// Gemini inference, schema validation, parity checking, and canonical JSON serialization.
func (s *GeometrySolver) Solve(ctx context.Context, task solver.Task) (solver.Result, error) {
	if s.geminiClient == nil {
		return solver.Result{
			TaskID:      task.TaskID,
			ProblemType: task.ProblemType,
			Success:     false,
		}, errors.New("gemini client is not configured")
	}

	trimmedPayload := strings.TrimSpace(task.Payload)
	if trimmedPayload == "" {
		return solver.Result{
			TaskID:      task.TaskID,
			ProblemType: task.ProblemType,
			Success:     false,
		}, errors.New("empty geometry task payload")
	}

	if err := ctx.Err(); err != nil {
		return solver.Result{
			TaskID:      task.TaskID,
			ProblemType: task.ProblemType,
			Success:     false,
		}, fmt.Errorf("geometry solver context canceled or timed out: %w", err)
	}

	inputFacts, summary, err := ExtractInputFacts(task.Payload)
	if err != nil {
		return solver.Result{
			TaskID:      task.TaskID,
			ProblemType: task.ProblemType,
			Success:     false,
		}, fmt.Errorf("failed to extract input facts from payload: %w", err)
	}

	prompt := BuildPrompt(inputFacts, summary)

	rawOutput, err := s.geminiClient.Generate(ctx, prompt)
	if err != nil {
		return solver.Result{
			TaskID:      task.TaskID,
			ProblemType: task.ProblemType,
			Success:     false,
		}, fmt.Errorf("geometry solver gemini generation failed: %w", err)
	}

	geomResult, err := ValidateGeometryResult(rawOutput, &inputFacts)
	if err != nil {
		return solver.Result{
			TaskID:      task.TaskID,
			ProblemType: task.ProblemType,
			Success:     false,
		}, fmt.Errorf("geometry solver output validation failed: %w", err)
	}

	// Preserve input task metadata in canonical result
	if inputFacts.ExplicitTarget != "" && geomResult.InputFacts.ExplicitTarget == "" {
		geomResult.InputFacts.ExplicitTarget = inputFacts.ExplicitTarget
	}
	if len(inputFacts.AdditionalConditions) > 0 && len(geomResult.InputFacts.AdditionalConditions) == 0 {
		geomResult.InputFacts.AdditionalConditions = inputFacts.AdditionalConditions
	}

	normalizedJSON, err := json.Marshal(geomResult)
	if err != nil {
		return solver.Result{
			TaskID:      task.TaskID,
			ProblemType: task.ProblemType,
			Success:     false,
		}, fmt.Errorf("failed to serialize validated geometry result: %w", err)
	}

	return solver.Result{
		TaskID:      task.TaskID,
		ProblemType: task.ProblemType,
		RawOutput:   string(normalizedJSON),
		Success:     true,
	}, nil
}
