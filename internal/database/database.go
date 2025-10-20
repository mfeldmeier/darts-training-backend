package database

import (
	"fmt"
	"log"

	"darts-training-app/internal/models"
	"darts-training-app/internal/utils"

	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Database struct {
	DB *gorm.DB
}

func NewDatabase(databaseURL string) (*Database, error) {
	db, err := gorm.Open(postgres.Open(databaseURL), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Auto-migrate the schema
	if err := db.AutoMigrate(
		&models.Team{},
		&models.Player{},
		&models.GameMode{},
		&models.TrainingSession{},
		&models.TrainingPlayer{},
		&models.TrainingGame{},
	); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	// Enable UUID extension for PostgreSQL
	if err := db.Exec("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\"").Error; err != nil {
		log.Printf("Warning: Could not enable UUID extension: %v", err)
	}

	return &Database{DB: db}, nil
}

func (d *Database) Close() error {
	sqlDB, err := d.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

func (d *Database) SeedDefaultData() error {
	// Check if game modes already exist
	var count int64
	d.DB.Model(&models.GameMode{}).Count(&count)
	if count > 0 {
		return nil // Data already seeded
	}

	// Create default game modes
	rules501 := `{"startingScore": 501, "finishType": "double", "legs": 3}`
	rulesCricket := `{"numbers": [20, 19, 18, 17, 16, 15, 25], "type": "standard"}`
	rulesClock := `{"sequence": true, "numbers": 20, "bull": true}`

	gameModes := []models.GameMode{
		{
			Name:        "501 Double Out",
			Description: utils.StringPtr("Classic 501 game, must finish on a double"),
			Rules:       rules501,
		},
		{
			Name:        "Cricket",
			Description: utils.StringPtr("Standard cricket game with numbers 20-15 and bull"),
			Rules:       rulesCricket,
		},
		{
			Name:        "Around the Clock",
			Description: utils.StringPtr("Hit numbers 1-20 in sequence, then bull"),
			Rules:       rulesClock,
	},
	}

	for _, gameMode := range gameModes {
		if err := d.DB.Create(&gameMode).Error; err != nil {
			return fmt.Errorf("failed to create game mode %s: %w", gameMode.Name, err)
		}
	}

	log.Println("Default game modes seeded successfully")
	return nil
}

// Helper functions for UUID handling
func StringToUUID(s string) (uuid.UUID, error) {
	return uuid.Parse(s)
}

func UUIDToString(id uuid.UUID) string {
	return id.String()
}