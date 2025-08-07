package indexer

import (
	"context"
	"crypto/sha256"
	"fmt"
	sqlite_vec "github.com/asg017/sqlite-vec-go-bindings/cgo"
	"gorm.io/gorm/clause"
	"lamina/pkg/ai"
	"lamina/pkg/database"
	"os"
)

func (i *Indexer) indexFile(ctx context.Context, filePath string) error {
	shouldReindex, err := i.shouldReindex(filePath)
	if err != nil {
		return err
	}

	if !shouldReindex {
		fmt.Printf("⏭️  Skipping unchanged file: %s\n", filePath)
		return nil
	}

	// Read file
	content, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	info, _ := os.Stat(filePath)
	contentHash := fmt.Sprintf("%x", sha256.Sum256(content))

	// Generate embeddings
	embeddingFloats, err := ai.GenerateEmbedding(ctx, string(content))
	if err != nil {
		return err
	}

	vectorBlob, err := sqlite_vec.SerializeFloat32(embeddingFloats)
	if err != nil {
		return err
	}

	// Save to DB
	file := database.File{
		Path:        filePath,
		ContentHash: contentHash,
		Size:        info.Size(),
		ModTime:     info.ModTime(),
		Content:     string(content),
	}

	// Use ON CONFLICT DO UPDATE for proper upsert
	if err := database.Store.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "path"}},
		DoUpdates: clause.AssignmentColumns([]string{"content_hash", "size", "mod_time", "content", "updated_at"}),
	}).Create(&file).Error; err != nil {
		return err

	}

	// Save embedding
	// Try to update first
	result := database.Store.Exec(`
	    UPDATE vec_embeddings SET embedding = ? WHERE file_id = ?
	`, vectorBlob, file.ID)

	// If no rows were affected, insert new record
	if result.RowsAffected == 0 {
		err = database.Store.Exec(`
			INSERT INTO vec_embeddings(file_id, embedding) 
			VALUES (?, ?)
		    `, file.ID, vectorBlob).Error
		if err != nil {
			return err
		}
	}

	fmt.Printf("✅ Successfully indexed: %s\n", file.ContentHash)
	return nil
}

func (i *Indexer) shouldReindex(filePath string) (bool, error) {
	// Get file info
	info, err := os.Stat(filePath)
	if err != nil {
		return false, err
	}

	// Read and hash content
	content, err := os.ReadFile(filePath)
	if err != nil {
		return false, err
	}

	currentHash := fmt.Sprintf("%x", sha256.Sum256(content))

	// Check if file exists in DB
	var existingFile database.File
	err = database.Store.Where("path = ?", filePath).First(&existingFile).Error

	if err != nil {
		// File not in DB, needs indexing
		return true, nil
	}

	// Compare hash and mod time
	return existingFile.ContentHash != currentHash ||
		existingFile.ModTime.Before(info.ModTime()), nil
}
