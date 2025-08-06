package config

var configPath string

// valid configuration keys
var stringConfigKeys = []string{
	"provider",
	"openai_key",
	"gemini_key",
	"database_path",
}

var stringSliceConfigKeys = []string{
	"ignore_patterns",
	"watch_paths",
}

var totalConfigKeys = append(stringConfigKeys, stringSliceConfigKeys...)

// basic YAML structure with defaults
var defaultConfig = `
# Lamina Configuration
provider: gemini
database_path: ~/.lamina/lamina.db
watch_paths:
  - ~/Documents
ignore_patterns:
  - .git
  - node_modules
  - "*.log"
`
