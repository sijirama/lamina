package ai

import (
	"context"
	"errors"
	"fmt"
	"google.golang.org/genai"
	"lamina/pkg/config"
	"strings"
)

func GenerateEmbedding(ctx context.Context, content string) (embedding []float32, err error) {
	provider := config.GetProvider()

	switch strings.ToLower(provider) {
	case "openai":
		return nil, errors.New("OpenAi provider not yet supported")
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
			genai.NewContentFromText(content, genai.RoleUser),
		}

		result, err := client.Models.EmbedContent(ctx,
			"gemini-embedding-exp-03-07",
			contents,
			&genai.EmbedContentConfig{
				TaskType: "RETRIEVAL_DOCUMENT", // For indexing documents
			},
		)
		if err != nil {
			return nil, fmt.Errorf("failed to generate embedding: %w", err)
		}

		if len(result.Embeddings) == 0 {
			return nil, errors.New("no embeddings returned from API")
		}

		return result.Embeddings[0].Values, nil
	default:
		return nil, errors.New(fmt.Sprintf("Invalid embedding model provider %s is not a valid provider", provider))

	}
}
