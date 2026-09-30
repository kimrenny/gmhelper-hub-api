package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"gmhelper.solution-hub/internal/planner"
	"gmhelper.solution-hub/internal/solver"
	pb "gmhelper.solution-hub/proto"
)

type SolveHandler struct {
	pb.UnimplementedSolutionHubServer
	planner planner.Planner
}

func NewSolveHandler(p planner.Planner) *SolveHandler {
	return &SolveHandler{
		planner: p,
	}
}

func (h *SolveHandler) SolveProblem(ctx context.Context, req *pb.SolveProblemRequest) (*pb.SolveProblemResponse, error) {
	if req == nil {
		return &pb.SolveProblemResponse{
			Status:  "INVALID_ARGUMENT",
			Result:  "request cannot be nil",
			Success: false,
		}, nil
	}

	taskID := strings.TrimSpace(req.TaskId)
	if taskID == "" {
		return &pb.SolveProblemResponse{
			Status:  "INVALID_ARGUMENT",
			Result:  "taskId cannot be empty",
			Success: false,
		}, nil
	}

	payload := strings.TrimSpace(req.Payload)
	if payload == "" {
		return &pb.SolveProblemResponse{
			TaskId:  taskID,
			Status:  "INVALID_ARGUMENT",
			Result:  "payload cannot be empty",
			Success: false,
		}, nil
	}

	problemType := strings.TrimSpace(req.ProblemType)
	if problemType == "" {
		// Fallback: inspect payload if problemType was serialized inside JSON
		var taskPayload map[string]interface{}
		if err := json.Unmarshal([]byte(payload), &taskPayload); err == nil {
			if pt, ok := taskPayload["problemType"].(string); ok {
				problemType = pt
			}
		}
	}
	if problemType == "" {
		problemType = "math" // default if unspecified
	}

	task := solver.Task{
		TaskID:      taskID,
		ProblemType: problemType,
		Payload:     payload,
		UserID:      strings.TrimSpace(req.UserId),
	}

	startTime := time.Now()
	log.Printf("[SolveHandler] Processing task %s (type: %s, user: %s)", task.TaskID, task.ProblemType, task.UserID)

	s, err := h.planner.GetSolver(task.ProblemType)
	if err != nil {
		log.Printf("[SolveHandler] Unsupported problem type for task %s: %s", task.TaskID, task.ProblemType)
		return &pb.SolveProblemResponse{
			TaskId:  task.TaskID,
			Status:  "UNSUPPORTED_PROBLEM_TYPE",
			Result:  fmt.Sprintf("unsupported problem type '%s'", task.ProblemType),
			Success: false,
		}, nil
	}

	result, err := s.Solve(ctx, task)
	duration := time.Since(startTime)

	if err != nil {
		log.Printf("[SolveHandler] Solver execution failed for task %s after %v: %v", task.TaskID, duration, err)
		return &pb.SolveProblemResponse{
			TaskId:  task.TaskID,
			Status:  "SOLVER_ERROR",
			Result:  err.Error(),
			Success: false,
		}, nil
	}

	log.Printf("[SolveHandler] Task %s completed successfully in %v", task.TaskID, duration)

	return &pb.SolveProblemResponse{
		TaskId:  task.TaskID,
		Status:  "SUCCESS",
		Result:  result.RawOutput,
		Success: true,
	}, nil
}
