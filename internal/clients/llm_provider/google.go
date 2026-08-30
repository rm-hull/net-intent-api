package llmprovider

import (
	"context"

	"github.com/cockroachdb/errors"
	"github.com/rm-hull/net-intent-api/internal/config"
	"google.golang.org/genai"
)

type GoogleProvider struct {
	client *genai.Client
	model  string
}

func NewGoogleProvider(ctx context.Context, cfg *config.Config) (Provider, error) {
	client, err := genai.NewClient(ctx, &genai.ClientConfig{APIKey: cfg.Gemini.APIKey})
	if err != nil {
		return nil, errors.Wrap(err, "failed to initialize Google client")
	}

	return &GoogleProvider{
		client: client,
		model:  cfg.Gemini.Model,
	}, nil
}

func (provider *GoogleProvider) Call(ctx context.Context, systemPrompt, userPrompt string) (string, error) {
	config := &genai.GenerateContentConfig{}
	if systemPrompt != "" {
		config.SystemInstruction = &genai.Content{
			Parts: []*genai.Part{genai.NewPartFromText(systemPrompt)},
		}
	}

	result, err := provider.client.Models.GenerateContent(
		ctx,
		provider.model,
		genai.Text(userPrompt),
		config,
	)
	if err != nil {
		return "", errors.Wrap(err, "failed to generate content:")
	}

	return result.Text(), nil
}

func (provider *GoogleProvider) Model() string {
	return provider.model
}
