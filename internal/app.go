package internal

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/Depado/ginprom"
	"github.com/cockroachdb/errors"
	"github.com/gin-contrib/cors"
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
const HTTP_PORT = 8080

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
	// logger.Info("Configuration on startup", "config", app.Config) // TODO: populate from cmd line

	r := gin.Default()
	// if err := r.SetTrustedProxies([]string{"192.168.1.2"}); err != nil { // TODO: from config
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

	api := r.Group("/v1/net-intent")
	api.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{http.MethodGet, http.MethodPost, http.MethodOptions},
		AllowHeaders:     []string{"Authorization", "Content-Type", "X-API-Key"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))
	api.GET("/ping", func(c *gin.Context) { // TODO: temporary - walking skeleton
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	addr := fmt.Sprintf(":%d", HTTP_PORT)
	logger.Info("Starting HTTP server", "addr", addr)
	if err := r.Run(addr); err != nil {
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
