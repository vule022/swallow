package config

const (
	DefaultProvider    = "openai"
	DefaultModel       = "gpt-4o"
	DefaultBaseURL     = "https://api.openai.com/v1"
	DefaultMaxTokens   = 4096
	DefaultTemperature = 0.3
	DefaultCopyMode    = "print"

	EnvAPIKey      = "SWALLOW_API_KEY"
	ConfigDirName  = ".swallow"
	ConfigFileName = "config.json"
	DBFileName     = "swallow.db"

	InboxDirName     = "inbox"
	ProcessedDirName = "processed"
	HooksDirName     = "hooks"
)
