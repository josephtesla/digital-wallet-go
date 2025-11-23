// Package db handles database operations including auto-migrations and seeding.
//
// GORM AutoMigrate automatically creates/updates database tables based on Go struct tags.
// No manual SQL files are needed - the database schema is entirely defined by the models
// in internal/models/ and automatically applied when the application starts.
//
// To add a new table:
// 1. Create a new model struct in internal/models/
// 2. Add it to the AutoMigrate() call in this package
// 3. Restart the application
package db

import (
	"fmt"

	"github.com/josephtesla/digital-wallet-go/internal/models"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

/*
Migrate runs all database migrations using GORM's AutoMigrate.
This creates/updates all tables from the Go model structs.
*/
func Migrate(db *gorm.DB, logger *zap.Logger) error {
	logger.Info("Running database migrations...")

	// AutoMigrate all models
	if err := db.AutoMigrate(
		&models.User{},
		&models.Wallet{},
		&models.Payment{},
		&models.LedgerTransaction{},
		&models.LedgerEntry{},
		&models.BankAccount{},
		&models.Idempotency{},
		&models.SystemAccount{},
	); err != nil {
		logger.Error("Failed to run migrations", zap.Error(err))
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	logger.Info("Database migrations completed successfully")
	return nil
}

// Seed populates the database with initial data (e.g., system accounts).
func Seed(db *gorm.DB, logger *zap.Logger) error {
	logger.Info("Seeding database with initial data...")

	// Create default system accounts if they don't exist
	systemAccounts := []models.SystemAccount{
		{
			Name:    "Settlement Account",
			Type:    "settlement",
			Balance: 0,
		},
		{
			Name:    "Revenue Account",
			Type:    "revenue",
			Balance: 0,
		},
		{
			Name:    "Pending Withdrawal Account",
			Type:    "pending_withdrawal",
			Balance: 0,
		},
	}

	for _, account := range systemAccounts {
		// Check if account already exists
		var existing models.SystemAccount
		if err := db.Where("type = ?", account.Type).First(&existing).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				// Create new account
				if err := db.Create(&account).Error; err != nil {
					logger.Error("Failed to seed system account", zap.String("type", account.Type), zap.Error(err))
					return fmt.Errorf("failed to seed system account: %w", err)
				}
				logger.Info("System account created", zap.String("type", account.Type))
			} else {
				logger.Error("Error checking for existing system account", zap.Error(err))
				return fmt.Errorf("error checking for existing system account: %w", err)
			}
		}
	}

	logger.Info("Database seeding completed successfully")
	return nil
}
