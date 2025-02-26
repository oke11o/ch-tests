package main

import (
	"context"
	"fmt"

	"github.com/oke11o/ch-tests/internal/config"
	"github.com/oke11o/ch-tests/internal/migrator"
)

func main() {
	ctx := context.Background()
	cfg, err := config.NewConfig(ctx, "TMPAPP")
	if err != nil {
		panic(err)
	}
	fmt.Printf("%+v\n", cfg)
	mig, err := migrator.NewMigrator(ctx, cfg.CH, cfg.MigrationPath)
	if err != nil {
		panic(err)
	}
	err = mig.Run()
	if err != nil {
		panic(err)
	}
	fmt.Println("Migrations completed successfully")
}
