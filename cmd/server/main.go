package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/jackc/pgx/v5/pgxpool"

	"muzz-explore-service/internal/config"
	"muzz-explore-service/internal/db"
	"muzz-explore-service/internal/server"
	"muzz-explore-service/internal/service"
)

func main() {
	cfg := config.Load()

	// Initialize postgres connection
	pool, err := pgxpool.New(context.Background(), cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	// Create queries
	queries := db.New(pool)

	// Initialize service
	exploreService := service.NewExploreService(queries)

	// Create and start server
	srv := server.NewGRPCServer(exploreService)

	// Handle graceful shutdown
	go func() {
		if err := srv.Start(cfg.GRPCPort); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	// Graceful shutdown
	log.Println("shutting down gRPC server...")
	srv.GracefulStop()
}
