package handler

import (
	"context"
	"testing"

	pb "gmhelper.solution-hub/proto"
)

func TestSolveHandler_SolveProblem(t *testing.T) {
	h := NewSolveHandler()

	req := &pb.SolveProblemRequest{
		TaskId:      "test-task-1",
		ProblemType: "math",
		Payload:     `{"data":"x+1=2"}`,
		UserId:      "user-123",
	}

	resp, err := h.SolveProblem(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.TaskId != "test-task-1" {
		t.Errorf("expected TaskId 'test-task-1', got '%s'", resp.TaskId)
	}

	if !resp.Success {
		t.Errorf("expected Success true, got false")
	}

	if resp.Status != "SUCCESS" {
		t.Errorf("expected Status 'SUCCESS', got '%s'", resp.Status)
	}

	expectedResult := "processed:math_engine:{\"data\":\"x+1=2\"}"
	if resp.Result != expectedResult {
		t.Errorf("expected Result '%s', got '%s'", expectedResult, resp.Result)
	}
}

func TestSolveHandler_SolveProblem_Geometry(t *testing.T) {
	h := NewSolveHandler()

	req := &pb.SolveProblemRequest{
		TaskId:      "test-task-2",
		ProblemType: "geometry",
		Payload:     `{"rect":{}}`,
		UserId:      "",
	}

	resp, err := h.SolveProblem(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.TaskId != "test-task-2" {
		t.Errorf("expected TaskId 'test-task-2', got '%s'", resp.TaskId)
	}

	expectedResult := "processed:geometry_engine:{\"rect\":{}}"
	if resp.Result != expectedResult {
		t.Errorf("expected Result '%s', got '%s'", expectedResult, resp.Result)
	}
}
