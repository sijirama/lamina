package database

import (
	"context"
	"encoding/json"
	"fmt"
	"lamina/pkg/ai"
	"strings"
	"time"
)

type SearchParams struct {
	SemanticQuery  string     `json:"semantic_query"`
	FileTypes      []string   `json:"file_types"`
	ModifiedAfter  *time.Time `json:"modified_after"`
	ModifiedBefore *time.Time `json:"modified_before"`
	SizeMin        *int64     `json:"size_min"`
	SizeMax        *int64     `json:"size_max"`
	PathContains   []string   `json:"path_contains"`
	Limit          int        `json:"limit"`
}

func ParseQuery(ctx context.Context, query string) (*SearchParams, error) {
	response, err := ai.GenerateStructuredQuery(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to parse query: %w", err)
	}

	// Response should already be clean JSON from structured output
	jsonStr := strings.TrimSpace(response)

	var params SearchParams
	if err := json.Unmarshal([]byte(jsonStr), &params); err != nil {
		// Fallback: treat entire query as semantic search
		return &SearchParams{
			SemanticQuery: query,
			Limit:         10,
		}, nil
	}

	// Set default limit if not specified
	if params.Limit == 0 {
		params.Limit = 10
	}

	return &params, nil
}
