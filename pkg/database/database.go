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

// NewStorage initializes the database with GORM and sqlite-vec.
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

// SaveFile saves or updates a file's metadata.
func (s *Storage) SaveFile(file *File) error {
	return s.db.Save(file).Error
}

// SaveEmbedding saves or updates a file's embedding.
func (s *Storage) SaveEmbedding(embedding *Embedding) error {
	return s.db.Save(embedding).Error
}

// GetFileByPath retrieves a file by its path.
func (s *Storage) GetFileByPath(path string) (*File, error) {
	var file File
	if err := s.db.Where("path = ?", path).First(&file).Error; err != nil {
		return nil, err
	}
	return &file, nil
}

// DeleteFileByPath deletes a file and its embedding by path.
func (s *Storage) DeleteFileByPath(path string) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		var file File
		if err := tx.Where("path = ?", path).First(&file).Error; err != nil {
			return err
		}
		if err := tx.Where("file_id = ?", file.ID).Delete(&Embedding{}).Error; err != nil {
			return err
		}
		return tx.Delete(&file).Error
	})
}
