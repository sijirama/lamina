package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"lamina/pkg/config"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Configuration management",
	Long:  `Manage Lamina configuration settings`,
}

var pathCmd = &cobra.Command{
	Use:   "path",
	Short: "Show configuration file path",
	Run: func(cmd *cobra.Command, args []string) {
		path := config.GetConfigPath()
		fmt.Printf("📁 Config file: %s\n", path)
	},
}
