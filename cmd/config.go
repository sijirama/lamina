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

var setCmd = &cobra.Command{
	Use:   "set [key] [value]",
	Short: "Set a configuration value",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		key := args[0]
		value := args[1]
		err := config.SetConfigValue(key, value)
		if err != nil {
			fmt.Println("❌ Error setting config:", err)
			return
		}
		fmt.Printf("✅ Config set: %s = %s\n", key, value)
	},
}

var getCmd = &cobra.Command{
	Use:   "get [key]",
	Short: "Get a configuration value",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		key := args[0]
		value := config.Get(key)
		if value == "" {
			fmt.Printf("❌ Config key '%s' not found or empty\n", key)
			return
		}
		fmt.Printf("%s = %s\n", key, value)
	},
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all configuration values",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("📋 Current configuration:")

		// Check common config keys
		keys := []string{"PROVIDER", "OPENAI_KEY", "GEMINI_KEY"}
		found := false

		for _, key := range keys {
			value := config.Get(key)
			if value != "" {
				found = true
				// Hide sensitive keys partially
				if key == "OPENAI_KEY" || key == "GEMINI_KEY" {
					if len(value) > 8 {
						maskedValue := value[:4] + "..." + value[len(value)-4:]
						fmt.Printf("  %s = %s\n", key, maskedValue)
					} else {
						fmt.Printf("  %s = %s\n", key, "***")
					}
				} else {
					fmt.Printf("  %s = %s\n", key, value)
				}
			}
		}

		if !found {
			fmt.Println("  No configuration values set")
		}
	},
}

var pathCmd = &cobra.Command{
	Use:   "path",
	Short: "Show configuration file path",
	Run: func(cmd *cobra.Command, args []string) {
		path := config.GetConfigPath()
		fmt.Printf("📁 Config file: %s\n", path)
	},
}
