package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"

	"github.com/danicc097/todo-ddd-example/internal"
	"github.com/danicc097/todo-ddd-example/internal/infrastructure"
	infraHttp "github.com/danicc097/todo-ddd-example/internal/infrastructure/http"
	"github.com/danicc097/todo-ddd-example/internal/infrastructure/logger"
	"github.com/danicc097/todo-ddd-example/internal/infrastructure/outbox"
	sharedHttp "github.com/danicc097/todo-ddd-example/internal/shared/infrastructure/http"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	var envPath string
	flag.StringVar(&envPath, "env", ".env", "Environment Variables filename")
	flag.Parse()

	if _, err := os.Stat(envPath); err == nil {
		if err := godotenv.Load(envPath); err != nil {
			slog.Warn("failed to load env file", slog.String("path", envPath), slog.String("error", err.Error()))
		}
	}

	cfg, err := internal.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	isProd := cfg.Env == internal.AppEnvProd

	shutdownLogger, err := logger.Init(ctx, cfg.LogLevel, isProd, cfg.OTEL.Endpoint)
	if err != nil {
		return fmt.Errorf("logger init failed: %w", err)
	}

	defer func() { _ = shutdownLogger(context.Background()) }()

	container, cleanupInfra, err := infrastructure.NewContainer(ctx, cfg)
	if err != nil {
		return fmt.Errorf("failed to initialize infrastructure: %w", err)
	}
	defer cleanupInfra()

	services, err := infrastructure.NewServices(ctx, cfg, container)
	if err != nil {
		return fmt.Errorf("failed to initialize services: %w", err)
	}

	relay := outbox.NewRelay(container.Pool, container.MultiBroker)
	go relay.Start(ctx)

	closers, err := infrastructure.RegisterSubscribers(container.MQConn, services.ScheduleRepo)
	if err != nil {
		return fmt.Errorf("failed to register subscribers: %w", err)
	}

	for _, c := range closers {
		defer c.Close()
	}

	handler := &infrastructure.CompositeHandler{
		TodoHandler:      services.TodoHandler,
		UserHandler:      services.UserHandler,
		WorkspaceHandler: services.WorkspaceHandler,
		AuthHandler:      services.AuthHandler,
		ScheduleHandler:  services.ScheduleHandler,
	}

	r, err := infraHttp.NewRouter(infraHttp.RouterConfig{
		Env:           cfg.Env,
		Pool:          container.Pool,
		Redis:         container.Redis,
		TokenVerifier: services.TokenProvider.Verifier,
		Handler:       handler,
		WSHandler:     handler.WS,
	})
	if err != nil {
		return fmt.Errorf("failed to initialize router: %w", err)
	}

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           r,
		ReadHeaderTimeout: 1 * time.Second,
	}

	serverErrors := make(chan error, 1)

	go func() {
		slog.InfoContext(ctx, "Application server starting", slog.String("port", cfg.Port))

		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErrors <- fmt.Errorf("server listen error: %w", err)
		}
	}()

	if cfg.Env != internal.AppEnvProd {
		docsURL := fmt.Sprintf("http://127.0.0.1:%s%s", cfg.Port, sharedHttp.RouteDocs)
		slog.InfoContext(ctx, "API server started", slog.String("docs_url", docsURL))
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-serverErrors:
		return err
	case sig := <-quit:
		slog.Info("Shutting down server...", slog.String("signal", sig.String()))
	}

	cancel()

	timeoutCtx, cancelTimeout := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelTimeout()

	if err := srv.Shutdown(timeoutCtx); err != nil {
		return fmt.Errorf("server forced to shutdown: %w", err)
	}

	slog.Info("Server exiting gracefully")

	return nil
}
