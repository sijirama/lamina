package indexer

import (
	"context"
	"crypto/sha256"
	"fmt"
	sqlite_vec "github.com/asg017/sqlite-vec-go-bindings/cgo"
	"lamina/pkg/ai"
	"lamina/pkg/database"
	"os"
)

func (i *Indexer) indexFile(ctx context.Context, filePath string) error {
	shouldReindex, err := i.shouldReindex(filePath)
	fmt.Println("Should reindex ", filePath, " is ", shouldReindex)
	if err != nil {
		return err
	}

	if !shouldReindex {
		fmt.Printf("⏭️  Skipping unchanged file: %s\n", filePath)
		return nil
	}

	fmt.Printf("📄 Indexing: %s\n", filePath)

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

	// Save to DB (upsert)
	file := database.File{
		Path:        filePath,
		ContentHash: contentHash,
		Size:        info.Size(),
		ModTime:     info.ModTime(),
		Content:     string(content),
	}

	if err := database.Store.Save(&file).Error; err != nil {
		return err
	}

	fmt.Println("Saved file for: ", file.ContentHash)

	// Save embedding
	return database.Store.Exec(`
	    INSERT OR REPLACE INTO vec_embeddings(file_id, embedding) 
	    VALUES (?, ?)
	`, file.ID, vectorBlob).Error
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
