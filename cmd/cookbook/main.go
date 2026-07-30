package main

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	coreapi "github.com/danielcbailey/Cookbook/core/api"
	"github.com/danielcbailey/Cookbook/core/apicommon"
	"github.com/danielcbailey/Cookbook/core/config"
	cache "github.com/danielcbailey/Cookbook/pkg/cache/local"
	database "github.com/danielcbailey/Cookbook/pkg/database/local"
	objectstore "github.com/danielcbailey/Cookbook/pkg/objectstore/local"
	"github.com/openai/openai-go/v3"
	openaioption "github.com/openai/openai-go/v3/option"
)

// defaultDataDir holds the on-disk state of the local backends. The
// single-machine deployment has no external services to point at, so the
// location is taken from the environment rather than the shared config.
const defaultDataDir = "data"

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		slog.Error("failed to load config", slog.Any("error", err))
		os.Exit(1)
	}

	providers := config.NewProviders().WithConfig(cfg)

	appCtx, cancel := context.WithCancel(context.Background())

	providers, err = initProviders(providers, cfg)
	if err != nil {
		slog.Error("provider setup failed", slog.Any("error", err))
		os.Exit(1)
	}

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		err := runServer(appCtx, providers, cfg.HTTPPort)
		if err != nil {
			slog.Error("http server failure", slog.Any("error", err))
		}
		wg.Done()
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	cancel()
	slog.Info("service termination requested")
	wg.Wait()
}

func initProviders(providers config.Providers, cfg *config.Config) (config.Providers, error) {
	openAIClient := openai.NewClient(openaioption.WithAPIKey(cfg.OpenAIKey))
	providers = providers.WithOpenAIClient(&openAIClient)

	dataDir := os.Getenv("DATA_DIR")
	if dataDir == "" {
		dataDir = defaultDataDir
	}

	dbDir := filepath.Join(dataDir, "db")
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}

	db, err := database.NewLocal(dbDir)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize local database: %w", err)
	}
	providers = providers.WithDB(db)

	providers = providers.WithCache(cache.NewLocal())

	store, err := objectstore.NewLocal(filepath.Join(dataDir, "objects"))
	if err != nil {
		return nil, fmt.Errorf("failed to initialize local object store: %w", err)
	}
	providers = providers.WithObjectStore(store)

	slogOptions := &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}

	providers = providers.WithLog(slog.New(slog.NewTextHandler(os.Stdout, slogOptions)))
	return providers, nil
}

func runServer(ctx context.Context, providers config.Providers, port int) error {
	mux := http.NewServeMux()
	for pattern, handler := range coreapi.Handlers() {
		mux.Handle(pattern, apicommon.WithLoggingProviders(providers, handler))
	}

	server := http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
	}

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = server.Shutdown(shutdownCtx)
	}()

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return err
	}

	slog.Info("Starting HTTP server", slog.String("address", listener.Addr().String()))

	err = server.Serve(listener)
	if err == http.ErrServerClosed {
		return nil
	}

	return err
}
