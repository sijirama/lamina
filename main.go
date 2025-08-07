package main

import (
	"context"
	"fmt"
	"lamina/cmd"
	"lamina/pkg/database"
	"lamina/pkg/indexer"
	"os"
)

func init() {
	must("Initialize Database", database.NewStorage())
}

func must(action string, err error) {
	if err != nil {
		panic("-> Failed to " + action + ": " + err.Error())
	}
}

func main() {
	// Check if running as daemon
	if len(os.Args) > 1 && os.Args[1] == "daemon" {
		runDaemon()
		return
	}

	// Otherwise, run CLI commands
	if err := cmd.Execute(); err != nil {
		fmt.Println("Lamina Error:", err)
		os.Exit(1)
	}
}

func runDaemon() {
	ctx := context.Background()
	idx, err := indexer.NewIndexer()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Daemon Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("🚀 Lamina daemon starting...")
	idx.Start(ctx)

	// Keep daemon running
	select {}
}
