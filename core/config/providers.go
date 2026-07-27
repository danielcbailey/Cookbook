package config

import (
	"log/slog"

	"github.com/danielcbailey/Cookbook/core/models"
	"github.com/danielcbailey/Cookbook/pkg/cache"
	"github.com/danielcbailey/Cookbook/pkg/database"
	"github.com/danielcbailey/Cookbook/pkg/objectstore"
	"github.com/openai/openai-go/v3"
)

type Providers interface {
	Config() *Config
	WithConfig(config *Config) Providers

	OpenAI() *openai.Client
	WithOpenAIClient(client *openai.Client) Providers

	Cache() cache.Cache
	WithCache(c cache.Cache) Providers

	DB() database.Database
	WithDB(db database.Database) Providers

	ObjectStore() objectstore.ObjectStore
	WithObjectStore(store objectstore.ObjectStore) Providers

	User() *models.User
	WithUser(user *models.User) Providers

	Log() *slog.Logger
	WithLog(logger *slog.Logger) Providers
}

type providers struct {
	config       *Config
	openAIClient *openai.Client
	cache        cache.Cache
	db           database.Database
	objectStore  objectstore.ObjectStore
	user         *models.User
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

func (p *providers) DB() database.Database {
	return p.db
}

func (p *providers) WithDB(db database.Database) Providers {
	newProviders := p.copy()
	newProviders.db = db
	return newProviders
}

func (p *providers) ObjectStore() objectstore.ObjectStore {
	return p.objectStore
}

func (p *providers) WithObjectStore(store objectstore.ObjectStore) Providers {
	newProviders := p.copy()
	newProviders.objectStore = store
	return newProviders
}

func (p *providers) User() *models.User {
	return p.user
}

func (p *providers) WithUser(user *models.User) Providers {
	newProviders := p.copy()
	newProviders.user = user
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
	provs := *p
	c := provs
	return &c
}
