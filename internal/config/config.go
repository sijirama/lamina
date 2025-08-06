package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

var configPath string

func init() {
	home, err := os.UserHomeDir()
	if err != nil {
		panic(fmt.Errorf("could not find home directory: %w", err))
	}

	// YAML format with .lamina filename
	configPath = filepath.Join(home, ".lamina")
	viper.SetConfigFile(configPath)
	viper.SetConfigType("yaml")
	viper.AutomaticEnv()

	// Create file if it doesn't exist with basic YAML structure
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		f, err := os.Create(configPath)
		if err != nil {
			panic(fmt.Errorf("could not create config file: %w", err))
		}
		defer f.Close()

		// Write basic YAML structure
		_, err = f.WriteString("# Lamina Configuration\n")
		if err != nil {
			panic(fmt.Errorf("could not write to config file: %w", err))
		}

		fmt.Printf("✅ Created config at %s\n", configPath)
	}

	if err := viper.ReadInConfig(); err != nil {
		fmt.Printf("⚠️ Could not read config at %s: %v\n", configPath, err)
	}
}

func Get(key string) string {
	return viper.GetString(key)
}

func GetOpenAIKey() string {
	return Get("OPENAI_KEY")
}

func GetGeminiKey() string {
	return Get("GEMINI_KEY")
}

func GetProvider() string {
	return Get("PROVIDER")
}

func SetConfigValue(key, value string) error {
	viper.Set(strings.ToUpper(key), value)
	viper.SetConfigType("yaml")
	return viper.WriteConfigAs(configPath)
}

func GetConfigPath() string {
	return configPath
}
