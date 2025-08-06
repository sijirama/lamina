package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {

	//subcommands to configCmd
	configCmd.AddCommand(listCmd)
	configCmd.AddCommand(pathCmd)

	rootCmd.AddCommand(configCmd)
}

var rootCmd = &cobra.Command{
	Use:   "lamina",
	Short: "Lamina - semantic file assistant",
	Long:  `Lamina is a local semantic assistant that helps you search and manage your files with natural language.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Welcome to Lamina 🧠 — try `lamina --help`")
	},
}

func Execute() error {
	return rootCmd.Execute()
}
