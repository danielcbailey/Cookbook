package config

import (
	"log/slog"

	"github.com/danielcbailey/Cookbook/pkg/cache"
	"github.com/openai/openai-go/v3"
)

type Providers interface {
	Config() *Config
	WithConfig(config *Config) Providers

	OpenAI() *openai.Client
	WithOpenAIClient(client *openai.Client) Providers

	Cache() cache.Cache
	WithCache(c cache.Cache) Providers

	Log() *slog.Logger
	WithLog(logger *slog.Logger) Providers
}

type providers struct {
	config       *Config
	openAIClient *openai.Client
	cache        cache.Cache
	log          *slog.Logger
}

func NewProviders() Providers {
	return &providers{}
}

func (p *providers) Config() *Config {
	return p.config
}

func (p *providers) WithConfig(config *Config) Providers {
	newProviders := p.copy()
	newProviders.config = config
	return newProviders
}

func (p *providers) OpenAI() *openai.Client {
	return p.openAIClient
}

func (p *providers) WithOpenAIClient(client *openai.Client) Providers {
	newProviders := p.copy()
	newProviders.openAIClient = client
	return newProviders
}

func (p *providers) Cache() cache.Cache {
	return p.cache
}

func (p *providers) WithCache(c cache.Cache) Providers {
	newProviders := p.copy()
	newProviders.cache = c
	return newProviders
}

func (p *providers) Log() *slog.Logger {
	return p.log
}

func (p *providers) WithLog(logger *slog.Logger) Providers {
	newProviders := p.copy()
	newProviders.log = logger
	return newProviders
}

func (p *providers) copy() *providers {
	return &providers{
		config:       p.config,
		openAIClient: p.openAIClient,
		cache:        p.cache,
		log:          p.log,
	}
}
