package main

import (
	"errors"
	"log"
	"net/http"
	"os"
	"time"

	"temporal-cost-optimizer/internal/analyzer"
	"temporal-cost-optimizer/internal/config"
	"temporal-cost-optimizer/internal/httpapi"
	"temporal-cost-optimizer/internal/optimizer"
	"temporal-cost-optimizer/internal/temporalcloud"
)

func main() {
	cfg := config.Load(os.LookupEnv)
	temporalClient := temporalcloud.NewClient(cfg.Temporal)

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
