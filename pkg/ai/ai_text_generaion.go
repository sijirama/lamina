package ai

import (
	"context"
	"errors"
	"fmt"
	"lamina/pkg/config"
	"strings"

	"google.golang.org/genai"
)

// GenerateStructuredQuery generates structured query parsing using your AI provider
func GenerateStructuredQuery(ctx context.Context, query string) (string, error) {
	provider := config.GetProvider()

	switch strings.ToLower(provider) {
	case "openai":
		return "", errors.New("OpenAI provider not yet supported")

	case "gemini":
		return geminiGenerateStructuredOutput(ctx, query)

	default:
		return "", fmt.Errorf("invalid provider %s is not supported", provider)
	}
}

func geminiGenerateStructuredOutput(ctx context.Context, query string) (string, error) {
	geminiKey := config.GetGeminiKey()
	if geminiKey == "" {
		return "", errors.New("Gemini API key not found")
	}

	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  geminiKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return "", fmt.Errorf("failed to create Gemini client: %w", err)
	}

	prompt := fmt.Sprintf(`Parse this natural language search query into structured search parameters.

	Query: "%s"

	Extract search parameters focusing on:
	1. The main semantic content to search for
	2. File type filters (pdf, docx, txt, go, py, js, etc.)
	3. Time-based filters (last week, past month, yesterday, etc.)
	4. Size-based filters (larger than X, smaller than Y)
	5. Path-based filters (specific folders or filename patterns)

	Time parsing rules:
	- "last week" = 7 days ago
	- "past month" = 30 days ago  
	- "yesterday" = 1 day ago
	- "this year" = start of current year
	- Use ISO 8601 format for dates

	File type mapping:
	- "pdfs" = ["pdf"]
	- "documents" = ["pdf", "docx", "txt", "md"]
	- "code files" = ["go", "py", "js", "ts"]
	- "images" = ["jpg", "png", "gif"]

	Only include fields that are explicitly mentioned or can be reasonably inferred.`, query)

	config := &genai.GenerateContentConfig{
		ResponseMIMEType: "application/json",
		ResponseSchema: &genai.Schema{
			Type: genai.TypeObject,
			Properties: map[string]*genai.Schema{
				"semantic_query": {
					Type:        genai.TypeString,
					Description: "The main content/topic to search for",
				},
				"file_types": {
					Type:        genai.TypeArray,
					Items:       &genai.Schema{Type: genai.TypeString},
					Description: "File extensions to filter by",
				},
				"modified_after": {
					Type:        genai.TypeString,
					Description: "ISO 8601 date for files modified after this date",
				},
				"modified_before": {
					Type:        genai.TypeString,
					Description: "ISO 8601 date for files modified before this date",
				},
				"size_min": {
					Type:        genai.TypeInteger,
					Description: "Minimum file size in bytes",
				},
				"size_max": {
					Type:        genai.TypeInteger,
					Description: "Maximum file size in bytes",
				},
				"path_contains": {
					Type:        genai.TypeArray,
					Items:       &genai.Schema{Type: genai.TypeString},
					Description: "Path components that should be present",
				},
				"limit": {
					Type:        genai.TypeInteger,
					Description: "Number of results to return (default 10)",
				},
			},
			PropertyOrdering: []string{
				"semantic_query", "file_types", "modified_after",
				"modified_before", "size_min", "size_max",
				"path_contains", "limit",
			},
		},
	}

	result, err := client.Models.GenerateContent(
		ctx,
		"gemini-2.0-flash-exp",
		genai.Text(prompt),
		config,
	)
	if err != nil {
		return "", fmt.Errorf("failed to generate structured query: %w", err)
	}

	return result.Text(), nil
}
