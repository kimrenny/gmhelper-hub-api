package handler

import (
	"context"
	"encoding/json"

	"gmhelper.solution-hub/internal/model"
	"gmhelper.solution-hub/internal/planner"
	pb "gmhelper.solution-hub/proto"
)

type SolveHandler struct {
	pb.UnimplementedSolutionHubServer
}

func NewSolveHandler() *SolveHandler {
	return &SolveHandler{}
}

func (h *SolveHandler) SolveProblem(ctx context.Context, req *pb.SolveProblemRequest) (*pb.SolveProblemResponse, error) {
	problemType := req.ProblemType
	if problemType == "" {
		var taskPayload map[string]interface{}
		_ = json.Unmarshal([]byte(req.Payload), &taskPayload)
		problemType, _ = taskPayload["problemType"].(string)
	}

	task := model.Task{
		TaskID:      req.TaskId,
		Payload:     req.Payload,
		ProblemType: problemType,
		UserID:      req.UserId,
	}

	strategy := planner.ChooseStrategy(task.ProblemType)

	result := processTask(task, strategy)

	return &pb.SolveProblemResponse{
		TaskId:  req.TaskId,
		Status:  "SUCCESS",
		Result:  result,
		Success: true,
	}, nil
}

func processTask(task model.Task, strategy string) string {
	return "processed:" + strategy + ":" + task.Payload
}
