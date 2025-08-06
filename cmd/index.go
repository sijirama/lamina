package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"lamina/pkg/indexer"
)

// IndexCmd defines the "index" command.
var indexCmd = &cobra.Command{
	Use:   "index",
	Short: "Index a directory",
	Long:  "Index a directory and add it to watch paths",
	Run: func(cmd *cobra.Command, args []string) {
		path := args[0]
		if path == "" {
			fmt.Println("❌ Path is not valid")
			return
		}

		indexer, err := indexer.NewIndexer()
		if err != nil {
			fmt.Println("❌ Error setting config:", err)
			return
		}

		ctx := cmd.Context()
		if err := indexer.IndexPath(ctx, path); err != nil {
			fmt.Println("❌ Error setting config:", err)
			return
		}

		fmt.Printf("✅ Indexed and watching: %s\n", path)
	},
}
