package internal

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/Depado/ginprom"
	"github.com/cockroachdb/errors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rm-hull/godx"
	"github.com/rm-hull/net-intent-api/internal/logging"
	sloggin "github.com/samber/slog-gin"

	healthcheck "github.com/tavsec/gin-healthcheck"
	hc_config "github.com/tavsec/gin-healthcheck/config"
)

const DEV_MODE = false // FIXME: replace with env vars

func Server() error {

	if !DEV_MODE {
		gin.SetMode(gin.ReleaseMode)
	}
	logLevel := logging.ParseLogLevel("INFO")

	logger := slog.New(slog.NewJSONHandler(
		os.Stderr,
		&slog.HandlerOptions{
			Level:       logLevel,
			AddSource:   true,
			ReplaceAttr: logging.ReplaceAttr}))

	if err := godotenv.Load(); err != nil {
		logger.Warn("No .env file found")
	}
	godx.Diagnostics(logger)
	// logger.Info("Configuration on startup", "config", app.Config)

	r := gin.Default()
	// if err := r.SetTrustedProxies([]string{"192.168.1.2"}); err != nil {
	// 	return errors.Wrap(err, "failed to trust proxies")
	// }

	prometheus := ginprom.New(
		ginprom.Path("/metrics"),
		ginprom.Ignore("/healthz", "/metrics"),
	)
	r.Use(
		gin.Recovery(),
		sloggin.NewWithConfig(logger, *newStructuredLoggingConfig()),
		prometheus.Instrument(),
	)
	if err := healthcheck.New(r, hc_config.DefaultConfig(), nil); err != nil {
		return errors.Wrap(err, "failed to initialize healthcheck")
	}

	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	if err := r.Run(); err != nil {
		return errors.Wrap(err, "failed to run server")
	}

	return nil
}

func newStructuredLoggingConfig() *sloggin.Config {
	config := sloggin.DefaultConfig()
	config.WithUserAgent = true
	config.WithClientIP = true
	config.Filters = append(config.Filters, sloggin.IgnorePath("/healthz", "/metrics", "/dns-query"))
	return &config
}
