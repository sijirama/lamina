package database

import (
	"time"
)

// File represents a file's metadata.
type File struct {
	ID          uint   `gorm:"primaryKey"`
	Path        string `gorm:"uniqueIndex;not null"`
	ContentHash string `gorm:"index"`
	Size        int64  `gorm:"not null"`
	ModTime     time.Time
	Content     string `gorm:"type:text"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// Embedding represents a file's vector embedding.
type Embedding struct {
	ID     uint   `gorm:"primaryKey"`
	FileID uint   `gorm:"index"`
	Vector []byte `gorm:"type:blob"`
	File   File   `gorm:"foreignKey:FileID"`
}
