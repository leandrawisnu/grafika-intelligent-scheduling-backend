package main

import (
	"github.com/grafika-scheduling/backend/internal/config"
	"github.com/grafika-scheduling/backend/internal/database"
	"github.com/grafika-scheduling/backend/internal/router"
	"github.com/grafika-scheduling/backend/pkg/mlclient"
)

func main() {
	cfg := config.Load()

	db := database.Connect(cfg.DatabaseURL)
	mlClient := mlclient.NewClient(cfg.MLServiceURL)

	app := router.New(db, mlClient)

	app.Listen(":" + cfg.ServerPort)
}
