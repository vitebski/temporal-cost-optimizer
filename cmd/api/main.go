package main

import (
	"errors"
	"log"
	"net/http"
	"time"

	"temporal-cost-optimizer/internal/config"
	"temporal-cost-optimizer/internal/httpapi"
	"temporal-cost-optimizer/internal/sampledata"
)

func main() {
	cfg, err := config.LoadFile(".env")
	if err != nil {
		log.Fatalf("failed to load configuration: %v", err)
	}

	samples := sampledata.NewService()

	server := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           httpapi.NewRouter(samples, samples),
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("starting Temporal Cost Copilot backend on %s with generated sample data", cfg.HTTPAddr)
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("server failed: %v", err)
	}
}
