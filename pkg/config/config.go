package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

func init() {
	home, err := os.UserHomeDir()
	if err != nil {
		panic(fmt.Errorf("could not find home directory: %w", err))
	}

	// Set config file to $HOME/.lamina/config (YAML format)
	configDir := filepath.Join(home, ".lamina")
	configPath = filepath.Join(configDir, "config")
	viper.SetConfigFile(configPath)
	viper.SetConfigType("yaml")
	viper.AutomaticEnv()

	// Create .lamina directory if it doesn't exist
	if err := os.MkdirAll(configDir, 0755); err != nil {
		panic(fmt.Errorf("could not create config directory %s: %w", configDir, err))
	}

	// Create config file if it doesn't exist with basic YAML structure
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		f, err := os.Create(configPath)
		if err != nil {
			panic(fmt.Errorf("could not create config file %s: %w", configPath, err))
		}
		defer f.Close()

		if _, err := f.WriteString(defaultConfig); err != nil {
			panic(fmt.Errorf("could not write to config file %s: %w", configPath, err))
		}

		fmt.Printf("✅ Created config at %s\n", configPath)
	}

	if err := viper.ReadInConfig(); err != nil {
		fmt.Printf("⚠️ Could not read config at %s: %v\n", configPath, err)
	}

	// Set default values in case config file is missing some keys
	viper.SetDefault("provider", "gemini")
	viper.SetDefault("database_path", filepath.Join(configDir, "lamina.db"))
	viper.SetDefault("watch_paths", []string{filepath.Join(home, "Documents")})
	viper.SetDefault("ignore_patterns", []string{".git", "node_modules", "*.log"})

}

func isValidKey(key string) bool {
	key = strings.ToLower(key)
	for _, k := range totalConfigKeys {
		if k == key {
			return true
		}
	}
	return false
}

// isSliceKey checks if a key is in stringSliceConfigKeys.
func isSliceKey(key string) bool {
	key = strings.ToLower(key)
	for _, k := range stringSliceConfigKeys {
		if k == key {
			return true
		}
	}
	return false
}

// Get retrieves a configuration value by key.
func Get(key string) string {
	var result string
	result = viper.GetString(key)
	if len(result) == 0 {
		result = strings.Join(viper.GetStringSlice(key), ", ")
	}
	return result
}

func GetStringSlice(key string) []string {
	return viper.GetStringSlice(key)
}

// GetOpenAIKey returns the OpenAI API key.
func GetOpenAIKey() string {
	return Get("OPENAI_KEY")
}

// GetGeminiKey returns the Gemini API key.
func GetGeminiKey() string {
	return Get("GEMINI_KEY")
}

// GetProvider returns the LLM provider.
func GetProvider() string {
	return Get("PROVIDER")
}

// GetWatchPaths returns the list of paths to index.
func GetWatchPaths() []string {
	paths := viper.GetStringSlice("watch_paths")
	// Resolve ~ to absolute paths
	for i, path := range paths {
		if strings.HasPrefix(path, "~") {
			home, _ := os.UserHomeDir()
			paths[i] = filepath.Join(home, strings.TrimPrefix(path, "~"))
		}
		paths[i] = filepath.Clean(paths[i])
	}
	return paths
}

// GetIgnorePatterns returns the list of ignore patterns.
func GetIgnorePatterns() []string {
	return viper.GetStringSlice("ignore_patterns")
}

func GetIgnorePaths() []string {
	return viper.GetStringSlice("ignore_paths")
}

// GetDatabasePath returns the database path.
func GetDatabasePath() string {
	path := viper.GetString("database_path")
	if strings.HasPrefix(path, "~") {
		home, _ := os.UserHomeDir()
		return filepath.Join(home, strings.TrimPrefix(path, "~"))
	}
	return filepath.Clean(path)
}

// SetConfigValue sets a configuration value and saves it to the config file.
func SetConfigValue(key, value string) error {
	viper.Set(strings.ToUpper(key), value)
	viper.SetConfigType("yaml")
	return viper.WriteConfigAs(configPath)
}

// SetConfigSliceValue sets a configuration slice value and saves it to the config file.
func SetConfigSliceValue(key string, values []string) error {
	viper.Set(strings.ToUpper(key), values)
	viper.SetConfigType("yaml")
	return viper.WriteConfigAs(configPath)
}

// GetConfigPath returns the path to the config file.
func GetConfigPath() string {
	return configPath
}
