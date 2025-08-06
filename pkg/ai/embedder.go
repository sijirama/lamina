package ai

import (
	"context"
	"errors"
	"fmt"
	"lamina/pkg/config"
	"strings"

	"github.com/tmc/langchaingo/llms/googleai"
)

func GenerateEmbedding(ctx context.Context, content string) (embedding []float32, err error) {
	provider := config.GetProvider()

	switch strings.ToLower(provider) {
	case "openai":
		return nil, errors.New("OpenAi provider not yet supported")
	case "gemini":
		gemini_key := config.GetGeminiKey()
		llm, err := googleai.New(ctx, googleai.WithAPIKey(gemini_key), googleai.WithDefaultModel("gemini-2.0-flash"))
		if err != nil {
			return nil, errors.New("Failed to initialize Gemini")
		}
		embeddings, err := llm.CreateEmbedding(ctx, []string{content})
		return embeddings[0], err
	default:
		return nil, errors.New(fmt.Sprintf("Invalid embedding model provider %s is not a valid provider", provider))

	}
}
