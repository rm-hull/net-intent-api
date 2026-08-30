package config

import (
	_ "embed"
	"log/slog"
)

//go:embed system_prompt.md
var SystemPrompt string

type GeminiOpts struct {
	APIKey       string  `json:"api_key" log:"redact" `
	Model        string  `json:"model"`
	SystemPrompt string  `json:"-"`
}

type UrlScanOpts struct {
	APIKey string `json:"api_key" log:"redact"`
}

type Config struct {
	Gemini         GeminiOpts   `json:"gemini"`
	UrlScan        UrlScanOpts  `json:"url_scan"`
	DevMode        bool         `json:"dev_mode"`
	HttpPort       int          `json:"http_port"`
	TrustedProxies []string     `json:"trusted_proxies"`
	Logger         *slog.Logger `json:"-"`
	LogLevel       string       `json:"log_level"`
}
