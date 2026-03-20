package config

// This is the main configuration block.
// Sections are the only items that should go here,
// with each section having it's own struct.
type Configuration struct {
	Atlassian AtlassianConfig `ini:"atlassian"`
}

type AtlassianConfig struct {
	AtlassianBaseUrl string `ini:"ATLASSIAN_BASE_URL"`
	AtlassianApiKey  string `ini:"ATLASSIAN_API_KEY"`
	UserEmail        string `ini:"ATLASSIAN_USER_EMAIL"`
}
