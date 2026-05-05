package main

import (
	"errors"
	"log"
	"net/http"
	"time"

	"temporal-cost-optimizer/internal/analyzer"
	"temporal-cost-optimizer/internal/config"
	"temporal-cost-optimizer/internal/httpapi"
	"temporal-cost-optimizer/internal/optimizer"
	"temporal-cost-optimizer/internal/temporalcloud"
)

func main() {
	cfg, err := config.LoadFile(".env")
	if err != nil {
		log.Fatalf("failed to load configuration: %v", err)
	}

	temporalClient, err := temporalcloud.NewClient(cfg.Temporal)
	if err != nil {
		log.Fatalf("failed to create Temporal Cloud client: %v", err)
	}
	defer temporalClient.Close()

	server := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           httpapi.NewRouter(analyzer.NewService(temporalClient), optimizer.NewService(temporalClient)),
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("starting Temporal Cost Copilot backend on %s", cfg.HTTPAddr)
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("server failed: %v", err)
	}
}
