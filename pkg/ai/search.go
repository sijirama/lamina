package ai

import (
	"context"
	"errors"
	"fmt"
	"lamina/pkg/config"
	"strings"

	"google.golang.org/genai"
)

func GenerateQueryEmbedding(ctx context.Context, query string) ([]float32, error) {
	provider := config.GetProvider()

	switch strings.ToLower(provider) {
	case "gemini":
		geminiKey := config.GetGeminiKey()
		if geminiKey == "" {
			return nil, errors.New("Gemini API key not found")
		}

		client, err := genai.NewClient(ctx, &genai.ClientConfig{
			APIKey:  geminiKey,
			Backend: genai.BackendGeminiAPI,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to create Gemini client: %w", err)
		}

		contents := []*genai.Content{
			genai.NewContentFromText(query, genai.RoleUser),
		}

		result, err := client.Models.EmbedContent(ctx,
			"gemini-embedding-exp-03-07",
			contents,
			&genai.EmbedContentConfig{
				TaskType: "RETRIEVAL_QUERY", // For search queries
			},
		)
		if err != nil {
			return nil, fmt.Errorf("failed to generate query embedding: %w", err)
		}

		if len(result.Embeddings) == 0 {
			return nil, errors.New("no embeddings returned from API")
		}

		return result.Embeddings[0].Values, nil

	default:
		return nil, fmt.Errorf("unsupported provider: %s", provider)
	}
}
