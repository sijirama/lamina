package database

import (
	"context"
	sqlite_vec "github.com/asg017/sqlite-vec-go-bindings/cgo"
	"lamina/pkg/ai"
)

func SearchFiles(ctx context.Context, query string, limit int) ([]File, error) {
	// Generate embedding for the search query
	queryEmbedding, err := ai.GenerateQueryEmbedding(ctx, query)
	if err != nil {
		return nil, err
	}

	// Serialize query embedding
	queryBlob, err := sqlite_vec.SerializeFloat32(queryEmbedding)
	if err != nil {
		return nil, err
	}

	// Search for similar vectors
	var fileIDs []uint
	err = Store.Raw(`
		SELECT file_id FROM vec_embeddings 
		WHERE embedding MATCH ? 
		ORDER BY distance 
		LIMIT ?
	    `, queryBlob, limit).Scan(&fileIDs).Error

	if err != nil {
		return nil, err
	}

	// Get the actual files
	var files []File
	return files, Store.Where("id IN ?", fileIDs).Find(&files).Error
}
