package database

import (
	"fmt"

	sqlite_vec "github.com/asg017/sqlite-vec-go-bindings/cgo"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"lamina/pkg/config"
)

var Store *gorm.DB

func NewStorage() error {
	sqlite_vec.Auto() // Enable sqlite-vec functions

	dbPath := config.GetDatabasePath()

	var err error

	Store, err = gorm.Open(sqlite.Dialector{
		DriverName: "sqlite3",
		DSN:        dbPath,
	}, &gorm.Config{})
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	if err := Store.Exec(`
        CREATE VIRTUAL TABLE IF NOT EXISTS vec_embeddings USING vec0(
            file_id INTEGER PRIMARY KEY,
            embedding FLOAT[3072]
        )
	`).Error; err != nil {
		return fmt.Errorf("failed to create vector table: %w", err)
	}

	// Run migrations
	if err := Store.AutoMigrate(&File{}); err != nil {
		return fmt.Errorf("failed to migrate database: %w", err)
	}

	// Verify sqlite-vec extension
	var vecVersion string
	if err := Store.Raw("SELECT vec_version()").Scan(&vecVersion).Error; err != nil {
		return fmt.Errorf("failed to load sqlite-vec: %w", err)
	}

	fmt.Printf("✅ sqlite-vec version: %s\n", vecVersion)

	return nil
}
