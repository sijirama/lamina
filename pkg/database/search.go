package database

import (
	"context"
	"fmt"
	"lamina/pkg/ai"
	"strings"

	sqlite_vec "github.com/asg017/sqlite-vec-go-bindings/cgo"
)

// AdvancedSearchFiles performs search with parsed parameters
func AdvancedSearchFiles(ctx context.Context, params *SearchParams) ([]File, error) {
	var files []File

	// Start with base query
	query := Store.Model(&File{})

	// Apply metadata filters first
	if len(params.FileTypes) > 0 {
		// Convert file types to LIKE conditions for file extensions
		var conditions []string
		var args []interface{}
		for _, ext := range params.FileTypes {
			conditions = append(conditions, "path LIKE ?")
			args = append(args, "%."+ext)
		}
		query = query.Where(strings.Join(conditions, " OR "), args...)
	}

	if params.ModifiedAfter != nil {
		query = query.Where("mod_time >= ?", *params.ModifiedAfter)
	}

	if params.ModifiedBefore != nil {
		query = query.Where("mod_time <= ?", *params.ModifiedBefore)
	}

	if params.SizeMin != nil {
		query = query.Where("size >= ?", *params.SizeMin)
	}

	if params.SizeMax != nil {
		query = query.Where("size <= ?", *params.SizeMax)
	}

	if len(params.PathContains) > 0 {
		for _, pathPart := range params.PathContains {
			query = query.Where("path LIKE ?", "%"+pathPart+"%")
		}
	}

	// If we have semantic query, do vector search first then filter
	if params.SemanticQuery != "" {
		// Generate embedding for semantic query
		queryEmbedding, err := ai.GenerateQueryEmbedding(ctx, params.SemanticQuery)
		if err != nil {
			return nil, err
		}

		// Serialize query embedding
		queryBlob, err := sqlite_vec.SerializeFloat32(queryEmbedding)
		if err != nil {
			return nil, err
		}

		// Get candidate file IDs from vector search (larger limit for filtering)
		var fileIDs []uint
		vectorLimit := params.Limit * 10 // Get more candidates to filter
		err = Store.Raw(`
			SELECT file_id FROM vec_embeddings 
			WHERE embedding MATCH ? 
			ORDER BY distance 
			LIMIT ?
		`, queryBlob, vectorLimit).Scan(&fileIDs).Error
		if err != nil {
			return nil, err
		}

		if len(fileIDs) > 0 {
			query = query.Where("id IN ?", fileIDs)
			// Order by the vector search relevance (maintain original order)
			orderCases := make([]string, len(fileIDs))
			for i, id := range fileIDs {
				orderCases[i] = fmt.Sprintf("WHEN %d THEN %d", id, i)
			}
			query = query.Order(fmt.Sprintf("CASE id %s END", strings.Join(orderCases, " ")))
		} else {
			// No vector matches, return empty result
			return []File{}, nil
		}
	} else {
		// No semantic search, just order by modification time
		query = query.Order("mod_time DESC")
	}

	err := query.Limit(params.Limit).Find(&files).Error
	return files, err
}

/*

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





*/
