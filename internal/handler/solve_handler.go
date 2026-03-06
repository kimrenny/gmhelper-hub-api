package handler

import (
	"context"
	"encoding/json"

	"gmhelper.solution-hub/internal/model"
	"gmhelper.solution-hub/internal/planner"
	pb "gmhelper.solution-hub/proto"
)

type SolveHandler struct {
	pb.UnimplementedSolutionHubServiceServer
}

func NewSolveHandler() *SolveHandler {
	return &SolveHandler{}
}

func (h *SolveHandler) SubmitTask(ctx context.Context, req *pb.SubmitTaskRequest) (*pb.SubmitTaskResponse, error) {
	var taskPayload map[string]interface{}
	_ = json.Unmarshal([]byte(req.TaskJson), &taskPayload)

	problemType, _ := taskPayload["problemType"].(string)

	task := model.Task{
		TaskID:      req.TaskId,
		Payload:     req.TaskJson,
		ProblemType: problemType,
		UserID:      req.UserId,
	}

	strategy := planner.ChooseStrategy(task.ProblemType)

	result := processTask(task, strategy)

	return &pb.SubmitTaskResponse{
		Status: result,
	}, nil
}

func processTask(task model.Task, strategy string) string {
	return "processed:" + strategy + ":" + task.Payload
}
