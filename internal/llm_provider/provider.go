package llmprovider

import (
	"context"

	"github.com/rm-hull/net-intent-api/internal/config"
)

type Provider interface {
	Call(ctx context.Context, systemPrompt, userPrompt string) (string, error)
	Model() string
}

func NewProvider(ctx context.Context, cfg *config.Config) (Provider, error) {
	return NewGoogleProvider(ctx, cfg)
}
