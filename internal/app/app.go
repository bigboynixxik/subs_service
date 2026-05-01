package app

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"subs_service/internal/migrations"
	"subs_service/internal/repository"
	"subs_service/internal/service"
	"subs_service/internal/transport"
	"subs_service/internal/transport/rest"
	"subs_service/pkg/closer"
	"subs_service/pkg/config"
	"subs_service/pkg/logger"
	"subs_service/pkg/migrator"
	"subs_service/pkg/postgres"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/stdlib"
)

type App struct {
	HTTPPort   string
	logs       *slog.Logger
	httpServer *http.Server
	closer     *closer.Closer
}

func New(ctx context.Context) (*App, error) {
	cfg, err := config.LoadConfig(".env")
	if err != nil {
		return nil, fmt.Errorf("app.New load config: %w", err)
	}

	logger.Setup(cfg.AppEnv)
	logs := logger.With("service", "subscribe-service")
	ctx = logger.WithContext(ctx, logs)

	logs.Info("initializing layers", "env", cfg.AppEnv, "port", cfg.HTTPPort)

	pool, err := postgres.NewPool(ctx, cfg.PGDSN)

	if err != nil {
		return nil, fmt.Errorf("app.New pool: %w", err)
	}
	sqlDB := stdlib.OpenDBFromPool(pool)
	defer sqlDB.Close()

	m, err := migrator.EmbedMigrations(sqlDB, migrations.FS, ".")
	if err != nil {
		return nil, fmt.Errorf("app.New migrations: %w", err)
	}
	if err := m.Up(); err != nil {
		return nil, fmt.Errorf("app.New failed to initialize migrations: %w", err)
	}
	logs.Info("migrations applied successfully")

	repo := repository.NewSubscriptionRepo(pool)
	svc := service.NewSubscriptionService(repo)
	sh := rest.NewSubsHandler(svc)

	mux := http.NewServeMux()

	mux.HandleFunc("POST /subscriptions", transport.LoggingMiddleware(logs, sh.CreateSubscription))
	mux.HandleFunc("GET /subscriptions", transport.LoggingMiddleware(logs, sh.ListSubscriptions))
	mux.HandleFunc("GET /subscriptions/{id}", transport.LoggingMiddleware(logs, sh.GetSubscription))
	mux.HandleFunc("PUT /subscriptions/{id}", transport.LoggingMiddleware(logs, sh.UpdateSubscription))
	mux.HandleFunc("DELETE /subscriptions/{id}", transport.LoggingMiddleware(logs, sh.DeleteSubscription))
	mux.HandleFunc("GET /subscriptions/cost", transport.LoggingMiddleware(logs, sh.CalculateTotalCost))

	httpServer := &http.Server{
		Addr:    ":" + cfg.HTTPPort,
		Handler: mux,
	}

	cl := closer.New()

	cl.Add(func(ctx context.Context) error {
		slog.Info("closing database connection pool")
		pool.Close()
		return nil
	})

	cl.Add(func(ctx context.Context) error {
		slog.Info("closing http server")
		return httpServer.Shutdown(ctx)
	})

	return &App{
		HTTPPort:   cfg.HTTPPort,
		logs:       logs,
		httpServer: httpServer,
		closer:     cl,
	}, nil
}

func (a *App) Run() {
	errCh := make(chan error)

	go func() {
		a.logs.Info("starting http server")
		if err := a.httpServer.ListenAndServe(); err != nil {
			errCh <- err
		}
	}()

	a.logs.Info("App.Run starting server",
		"port", a.HTTPPort)

	quit := make(chan os.Signal, 1)

	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-errCh:
		a.logs.Info("app.run server error",
			slog.String("error", err.Error()))
	case sig := <-quit:
		a.logs.Info("App.Run stopping server",
			slog.String("signal", sig.String()))
	}

	a.logs.Info("shutting down server")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := a.closer.Close(shutdownCtx); err != nil {
		a.logs.Error("app.run server shutdown error",
			slog.String("error", err.Error()))
	}
}
