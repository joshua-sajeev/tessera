package main

import (
	"context"
	"log"

	"github.com/joshua-sajeev/tessera/internal/adapters/minio"
	"github.com/joshua-sajeev/tessera/internal/adapters/postgres"
	"github.com/joshua-sajeev/tessera/internal/config"
)

func main() {
	ctx := context.Background()

	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	pool, err := postgres.NewPool(ctx, cfg.Database)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	assetRepo := postgres.NewAssetRepository(pool)
	processingRepo := postgres.NewProcessingRepository(pool)

	storage, err := minio.New(ctx, cfg.Storage)
	if err != nil {
		log.Fatal(err)
	}

	_ = assetRepo
	_ = processingRepo
	_ = storage

	log.Println("Tessera API dependencies initialized successfully")
}
