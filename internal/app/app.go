package app

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"ms-telecontrol/internal/config"
	httpdelivery "ms-telecontrol/internal/delivery/http"
	"ms-telecontrol/internal/infrastructure/persistence/memory"
	"ms-telecontrol/internal/usecase"
)

type App struct {
	server          *http.Server
	shutdownTimeout time.Duration
	closers         []io.Closer
}

func New(cfg config.Config) (*App, error) {
	commandRepository := memory.NewTelecontrolCommandRepository()
	commandPublisher, commandHooks, closers, err := buildTelecontrolIntegrations(cfg)
	if err != nil {
		return nil, fmt.Errorf("build telecontrol integrations: %w", err)
	}

	telecontrolUsecase := usecase.NewTelecontrolService(commandRepository, commandPublisher, commandHooks...)
	router := httpdelivery.NewRouter(telecontrolUsecase)

	server := &http.Server{
		Addr:              ":" + cfg.HTTPPort,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	return &App{
		server:          server,
		shutdownTimeout: cfg.ShutdownTimeout,
		closers:         closers,
	}, nil
}

func (a *App) Run() error {
	return a.server.ListenAndServe()
}

func (a *App) Shutdown(ctx context.Context) error {
	shutdownCtx, cancel := context.WithTimeout(ctx, a.shutdownTimeout)
	defer cancel()

	errs := make([]error, 0, len(a.closers)+1)

	if err := a.server.Shutdown(shutdownCtx); err != nil {
		errs = append(errs, err)
	}

	for _, closer := range a.closers {
		if err := closer.Close(); err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}
