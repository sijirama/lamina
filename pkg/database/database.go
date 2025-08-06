package database

import (
	"fmt"

	sqlite_vec "github.com/asg017/sqlite-vec-go-bindings/cgo"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"lamina/pkg/config"
)

type Storage struct {
	db *gorm.DB
}

var Store *Storage

func NewStorage() error {
	sqlite_vec.Auto() // Enable sqlite-vec functions

	dbPath := config.GetDatabasePath()

	db, err := gorm.Open(sqlite.Dialector{
		DriverName: "sqlite3",
		DSN:        dbPath,
	}, &gorm.Config{})
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	// Run migrations
	if err := db.AutoMigrate(&File{}, &Embedding{}); err != nil {
		return fmt.Errorf("failed to migrate database: %w", err)
	}

	// Verify sqlite-vec extension
	var vecVersion string
	if err := db.Raw("SELECT vec_version()").Scan(&vecVersion).Error; err != nil {
		return fmt.Errorf("failed to load sqlite-vec: %w", err)
	}

	Store = &Storage{db: db}

	fmt.Printf("✅ sqlite-vec version: %s\n", vecVersion)

	return nil
}
