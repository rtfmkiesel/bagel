package database

import (
	"bagel/internal/logger"
	"bagel/internal/semgrep"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

const (
	// Hardcoded database name for now
	dbName = "bagel.db"
)

// Init sets up the database connection
func Init() (db *gorm.DB, err error) {
	db, err = gorm.Open(sqlite.Open(dbName), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	logger.Info("Connected to database %s", dbName)

	if err := db.AutoMigrate(&semgrep.Scan{}); err != nil {
		return nil, err
	}
	logger.Info("Migrated database")

	return db, nil
}

// Close closes the database connection
func Close(db *gorm.DB) (err error) {
	dbBagel, err := db.DB()
	if err != nil {
		return err
	}

	if err := dbBagel.Close(); err != nil {
		return err
	}
	logger.Info("Closed database connection")

	return nil
}
