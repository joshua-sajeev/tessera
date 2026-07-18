package main

import (
	"fmt"
	"log"

	"github.com/joshu-sajeev/tessera/internal/config"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(cfg.Database.Host)
}
