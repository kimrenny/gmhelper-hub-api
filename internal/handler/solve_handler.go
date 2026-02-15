package handler

import (
	"context"

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

func (h *SolveHandler) Solve(ctx context.Context, req *pb.SolveRequest) (*pb.SolveResponse, error) {
	task := model.Task{
		TaskID:      req.TaskId,
		ProblemType: req.ProblemType,
		Payload:     req.Payload,
	}

	strategy := planner.ChooseStrategy(task.ProblemType)

	result := processTask(task, strategy)

	return &pb.SolveResponse{
		TaskId:   task.TaskID,
		Strategy: strategy,
		Result:   result,
		Success:  true,
	}, nil
}

func processTask(task model.Task, strategy string) string {
	return "processed:" + strategy + ":" + task.Payload
}
