package config

import (
	_ "embed"
	"log/slog"
)

//go:embed system_prompt.md
var SystemPrompt string

type Config struct {
	Gemini struct {
		APIKey       string
		Model        string
		SystemPrompt string
	}
	UrlScan struct {
		APIKey string
	}
	DevMode        bool
	HttpPort       int
	TrustedProxies []string
	Logger         *slog.Logger
	LogLevel       string
}
