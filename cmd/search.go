package cmd

import (
	"context"
	"fmt"
	"lamina/pkg/database"
	"strings"

	"github.com/spf13/cobra"
)

var searchCmd = &cobra.Command{
	Use:   "search [query]",
	Short: "Search files by content",
	Long: `Search files using natural language queries that can include both content and metadata filters.

Examples:
  lamina search "all files about machine learning"
  lamina search "PDFs about geospatial indexing modified last week"  
  lamina search "Go code files dealing with databases from this month"
  lamina search "documents containing 'API documentation' larger than 1MB"`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()

		query := strings.Join(args, " ")

		fmt.Printf("🔍 Parsing query: %s\n", query)

		// Parse the natural language query
		params, err := database.ParseQuery(ctx, query)
		if err != nil {
			fmt.Printf("❌ Query parsing error: %v\n", err)
			return
		}

		fmt.Println(params)

		// Show what we parsed (optional debug info)
		if verbose, _ := cmd.Flags().GetBool("verbose"); verbose {
			printParsedParams(params)
		}

		// Search files
		files, err := database.AdvancedSearchFiles(ctx, params)
		if err != nil {
			fmt.Printf("❌ Search error: %v\n", err)
			return
		}

		if len(files) == 0 {
			fmt.Println("No files found matching your query")
			return
		}

		fmt.Printf("Found %d files:\n\n", len(files))
		for i, file := range files {
			fmt.Printf("%d. %s\n", i+1, file.Path)
			fmt.Printf("   📅 Modified: %s\n", file.ModTime.Format("2006-01-02 15:04"))
			fmt.Printf("   📊 Size: %s\n", formatFileSize(file.Size))
			if verbose, _ := cmd.Flags().GetBool("verbose"); verbose {
				// Show content preview
				preview := strings.ReplaceAll(file.Content, "\n", " ")
				if len(preview) > 200 {
					preview = preview[:200] + "..."
				}
				fmt.Printf("   📝 Preview: %s\n", preview)
			}
			fmt.Println()
		}
	},
}

// ==============================================

func printParsedParams(params *database.SearchParams) {
	fmt.Println("🧠 Parsed search parameters:")
	if params.SemanticQuery != "" {
		fmt.Printf("   📝 Content: %s\n", params.SemanticQuery)
	}
	if len(params.FileTypes) > 0 {
		fmt.Printf("   📁 File types: %v\n", params.FileTypes)
	}
	if params.ModifiedAfter != nil {
		fmt.Printf("   📅 Modified after: %s\n", params.ModifiedAfter.Format("2006-01-02"))
	}
	if params.ModifiedBefore != nil {
		fmt.Printf("   📅 Modified before: %s\n", params.ModifiedBefore.Format("2006-01-02"))
	}
	if len(params.PathContains) > 0 {
		fmt.Printf("   🗂️  Path contains: %v\n", params.PathContains)
	}
	if params.SizeMin != nil || params.SizeMax != nil {
		fmt.Printf("   📊 Size range: %s\n", formatSizeRange(params.SizeMin, params.SizeMax))
	}
	fmt.Printf("   🔢 Limit: %d\n", params.Limit)
	fmt.Println()
}

func formatFileSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

func formatSizeRange(min, max *int64) string {
	if min != nil && max != nil {
		return fmt.Sprintf("%s - %s", formatFileSize(*min), formatFileSize(*max))
	} else if min != nil {
		return fmt.Sprintf("> %s", formatFileSize(*min))
	} else if max != nil {
		return fmt.Sprintf("< %s", formatFileSize(*max))
	}
	return "any size"
}
