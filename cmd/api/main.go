package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os/signal"
	"syscall"

	"ms-telecontrol/internal/app"
	"ms-telecontrol/internal/config"
)

func main() {
	cfg := config.Load()

	application, err := app.New(cfg)
	if err != nil {
		log.Fatalf("build app: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	errCh := make(chan error, 1)

	go func() {
		errCh <- application.Run()
	}()

	select {
	case <-ctx.Done():
		if err := application.Shutdown(context.Background()); err != nil {
			log.Fatalf("shutdown server: %v", err)
		}
	case err := <-errCh:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("run server: %v", err)
		}
	}
}
