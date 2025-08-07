package cmd

import (
	"context"
	"fmt"
	"lamina/pkg/database"

	"github.com/spf13/cobra"
)

// cmd/search.go
var searchCmd = &cobra.Command{
	Use:   "search [query]",
	Short: "Search files by content",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		query := args[0]

		// Search files
		files, err := database.SearchFiles(ctx, query, 1)
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
		}
	},
}
