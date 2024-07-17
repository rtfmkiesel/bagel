package main

import (
	"bagel/internal/database"
	"bagel/internal/logger"
	"bagel/internal/router"
	"bagel/internal/semgrep"
)

func main() {
	db, err := database.Init()
	if err != nil {
		logger.Fatal(err)
	}

	if err := semgrep.StartWorkers(db); err != nil {
		logger.Fatal(err)
	}

	router.Start(db)

	if semgrep.WaitForShutdown() {
		if err := router.Stop(); err != nil {
			logger.Fatal(err)
		}

		semgrep.StopWorkers()

		if err := database.Close(db); err != nil {
			logger.Fatal(err)
		}
	}
}
