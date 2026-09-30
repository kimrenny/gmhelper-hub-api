package solver

import "context"

type Task struct {
	TaskID      string
	ProblemType string
	Payload     string
	UserID      string
}

type Result struct {
	TaskID      string
	ProblemType string
	RawOutput   string
	Success     bool
}

type Solver interface {
	Solve(ctx context.Context, task Task) (Result, error)
}
