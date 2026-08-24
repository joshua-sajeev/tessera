package main

import (
	"context"
	"log"
	"net/http"

	httpAdapter "github.com/joshua-sajeev/tessera/internal/adapters/http"
	"github.com/joshua-sajeev/tessera/internal/adapters/http/handler"
	"github.com/joshua-sajeev/tessera/internal/adapters/minio"
	"github.com/joshua-sajeev/tessera/internal/adapters/postgres"
	"github.com/joshua-sajeev/tessera/internal/application/auth"
	"github.com/joshua-sajeev/tessera/internal/application/userapp"
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

	userRepo := postgres.NewUserRepository(pool)
	assetRepo := postgres.NewAssetRepository(pool)
	processingRepo := postgres.NewProcessingRepository(pool)

	storage, err := minio.New(ctx, cfg.Storage)
	if err != nil {
		log.Fatal(err)
	}

	_ = assetRepo
	_ = processingRepo
	_ = storage

	authenticator := auth.NewAuthenticator(userRepo, cfg.APIKey.Prefix, cfg.APIKey.Version)

	userService := userapp.NewUserService(userRepo, cfg.APIKey.Prefix, cfg.APIKey.Version)
	userHandler := handler.NewUserHandler(userService)

	router := httpAdapter.NewRouter(userHandler, authenticator)

	log.Printf("Starting HTTP server on port %s", cfg.Server.Port)
	srv := &http.Server{
		Addr:    ":" + cfg.Server.Port,
		Handler: router,
	}

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("listen and serve: %v", err)
	}
}
