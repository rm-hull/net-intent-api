package cmd

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/Depado/ginprom"
	"github.com/cockroachdb/errors"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rm-hull/godx"
	llmprovider "github.com/rm-hull/net-intent-api/internal/clients/llm_provider"
	"github.com/rm-hull/net-intent-api/internal/config"
	"github.com/rm-hull/net-intent-api/internal/handlers"
	"github.com/rm-hull/net-intent-api/internal/logging"
	"github.com/rm-hull/net-intent-api/internal/middlewares"
	"github.com/rm-hull/net-intent-api/internal/service/rdap"
	"github.com/rm-hull/net-intent-api/internal/service/urlscan"
	sloggin "github.com/samber/slog-gin"
	healthcheck "github.com/tavsec/gin-healthcheck"
	hc_config "github.com/tavsec/gin-healthcheck/config"
)

func Start(cfg *config.Config) error {

	if !cfg.DevMode {
		gin.SetMode(gin.ReleaseMode)
	}

	godx.Diagnostics(cfg.Logger)
	cfg.Logger.Info("Parsed configuration", "config", cfg)

	rdapService := rdap.NewService(5 * time.Minute)
	urlScanService := urlscan.NewUrlScanService(cfg.UrlScan.APIKey, 5*time.Minute)
	provider, err := llmprovider.NewProvider(context.Background(), cfg)
	if err != nil {
		return errors.Wrapf(err, "failed to initialize LLM provider")
	}

	r := gin.New()
	if cfg.DevMode {
		cfg.Logger.Warn("pprof endpoints are enabled and exposed. Do not run with this flag in production.")
		pprof.Register(r)
	}
	if err := r.SetTrustedProxies(cfg.TrustedProxies); err != nil {
		return errors.Wrap(err, "failed to trust proxies")
	}

	prometheus := ginprom.New(
		ginprom.Path("/metrics"),
		ginprom.Ignore("/healthz", "/metrics"),
	)
	r.Use(
		gin.Recovery(),
		sloggin.NewWithConfig(cfg.Logger, *logging.NewStructuredLoggingConfig()),
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
	auth := middlewares.RequireAnyAuth(middlewares.ProxyAuth(cfg.DevMode))
	api.GET("/analyze", auth, handlers.Analyze(cfg, provider, urlScanService, rdapService))
	api.GET("/urlscan", auth, handlers.UrlScan(urlScanService))
	api.GET("/rdap", auth, handlers.RdapHandler(rdapService))

	addr := fmt.Sprintf(":%d", cfg.HttpPort)
	cfg.Logger.Info("Starting HTTP server", "addr", addr)
	if err := r.Run(addr); err != nil {
		return errors.Wrap(err, "failed to run server")
	}

	return nil
}
