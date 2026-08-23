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
		Prompt:   config.Prompt,
		HttpPort: 8080,
	}

	var rootCommand = &cobra.Command{
		Use:   "net-intent-api",
		Short: "Automated intent scoring API",
		RunE: func(cmd *cobra.Command, args []string) error {
			if !cmd.Flags().Changed("api-key") {
				cliConfig.APIKey = os.Getenv("GEMINI_API_KEY")
			}
			if !cmd.Flags().Changed("model") {
				cliConfig.Model = os.Getenv("GEMINI_MODEL")
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

	rootCommand.Flags().StringVar(&cliConfig.APIKey, "api-key", "", "LLM API key")
	rootCommand.Flags().StringVar(&cliConfig.Model, "model", "", "LLM model")
	rootCommand.Flags().BoolVar(&cliConfig.DevMode, "dev-mode", false, "enable dev mode")
	rootCommand.Flags().IntVar(&cliConfig.HttpPort, "http-port", cliConfig.HttpPort, "HTTP server port")
	rootCommand.Flags().StringSliceVar(&cliConfig.TrustedProxies, "trusted-proxy", nil, "trusted proxy address; may be repeated")

	if err := rootCommand.Execute(); err != nil {
		os.Exit(-1)
	}
}
