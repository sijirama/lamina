package database

import (
	"time"
)

// File represents a file's metadata.
type File struct {
	ID           uint      `gorm:"primaryKey"`
	Path         string    `gorm:"uniqueIndex;not null"`
	Name         string    `gorm:"not null"`
	Size         int64     `gorm:"not null"`
	LastModified time.Time `gorm:"not null"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// Embedding represents a file's vector embedding.
type Embedding struct {
	ID        uint   `gorm:"primaryKey"`
	FileID    uint   `gorm:"uniqueIndex;not null"`
	Vector    []byte `gorm:"type:BLOB;not null"` // sqlite-vec BLOB
	CreatedAt time.Time
	UpdatedAt time.Time
}
