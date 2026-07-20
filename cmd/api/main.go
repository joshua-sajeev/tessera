package main

import (
	"context"
	"fmt"
	"log"

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
	fmt.Print(pool)
}
