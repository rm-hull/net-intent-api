package config

import (
	_ "embed"
	"log/slog"
)

//go:embed prompt.md
var Prompt string

type Config struct {
	Gemini struct {
		APIKey string
		Model  string
		Prompt string
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
