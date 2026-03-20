package config

// This is the main configuration block.
// Sections are the only items that should go here,
// with each section having it's own struct.
type Configuration struct {
	Atlassian AtlassianConfig `ini:"atlassian"`
}

type AtlassianConfig struct {
	AtlassianBaseUrl string `ini:"atlassian_base_url"`
	AtlassianApiKey  string `ini:"atlassian_api_key"`
	UserEmail        string `ini:"atlassian_user_email"`
}
