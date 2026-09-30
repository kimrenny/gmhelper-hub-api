package main

import (
	"context"
	"log"

	"gmhelper.solution-hub/internal/config"
	"gmhelper.solution-hub/internal/gemini"
	"gmhelper.solution-hub/internal/handler"
	"gmhelper.solution-hub/internal/planner"
	"gmhelper.solution-hub/internal/server"
	geoSolver "gmhelper.solution-hub/internal/solver/geometry"
	mathSolver "gmhelper.solution-hub/internal/solver/math"
)

func main() {
	cfg := config.Load()

	if err := cfg.Validate(); err != nil {
		log.Fatalf("[FATAL] Configuration validation error: %v", err)
	}

	ctx := context.Background()
	geminiClient, err := gemini.NewGenAIClient(ctx, cfg)
	if err != nil {
		log.Fatalf("[FATAL] Failed to initialize Gemini client: %v", err)
	}

	mathS := mathSolver.NewMathSolver(geminiClient)
	geoS := geoSolver.NewGeometrySolver(geminiClient)
	plnr := planner.NewDefaultPlanner(mathS, geoS)
	solveHandler := handler.NewSolveHandler(plnr)

	srv := server.NewServer(cfg.GRPCPort, solveHandler)

	if err := srv.Run(); err != nil {
		log.Fatalf("[FATAL] gRPC server error: %v", err)
	}
}
