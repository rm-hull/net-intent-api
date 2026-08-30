package main

import (
	"log/slog"
	"os"

	"github.com/joho/godotenv"
	server "github.com/rm-hull/net-intent-api/internal/cmd"
	"github.com/rm-hull/net-intent-api/internal/config"
	"github.com/rm-hull/net-intent-api/internal/logging"
	"github.com/spf13/cobra"
)

func main() {
	_ = godotenv.Load()
	cliConfig := config.Config{
		HttpPort: 8080,
	}

	cliConfig.Gemini.SystemPrompt = config.SystemPrompt

	var rootCommand = &cobra.Command{
		Use:   "net-intent-api",
		Short: "Automated intent scoring API",
		RunE: func(cmd *cobra.Command, args []string) error {
			if !cmd.Flags().Changed("urlscan-api-key") {
				cliConfig.UrlScan.APIKey = os.Getenv("URLSCAN_API_KEY")
			}
			if !cmd.Flags().Changed("gemini-api-key") {
				cliConfig.Gemini.APIKey = os.Getenv("GEMINI_API_KEY")
			}
			if !cmd.Flags().Changed("gemini-model") {
				cliConfig.Gemini.Model = os.Getenv("GEMINI_MODEL")
			}
			if !cmd.Flags().Changed("log-level") {
				cliConfig.LogLevel = "INFO"
			}

			cliConfig.Logger = slog.New(slog.NewJSONHandler(
				os.Stderr,
				&slog.HandlerOptions{
					Level:       logging.ParseLogLevel(cliConfig.LogLevel),
					AddSource:   true,
					ReplaceAttr: logging.ReplaceAttr}))

			return server.Start(&cliConfig)
		}}

	rootCommand.Flags().StringVar(&cliConfig.UrlScan.APIKey, "urlscan-api-key", "", "UrlScan API key")
	rootCommand.Flags().StringVar(&cliConfig.Gemini.APIKey, "gemini-api-key", "", "Gemini API key")
	rootCommand.Flags().StringVar(&cliConfig.Gemini.Model, "gemini-model", "", "Gemini model")
	rootCommand.Flags().BoolVar(&cliConfig.DevMode, "dev-mode", false, "enable dev mode")
	rootCommand.Flags().IntVar(&cliConfig.HttpPort, "http-port", cliConfig.HttpPort, "HTTP server port")
	rootCommand.Flags().StringSliceVar(&cliConfig.TrustedProxies, "trusted-proxy", nil, "trusted proxy address; may be repeated")

	if err := rootCommand.Execute(); err != nil {
		os.Exit(-1)
	}
}
