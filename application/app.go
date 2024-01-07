package application

import (
	"context"
	"fmt"
	"net/http"

	"github.com/redis/go-redis/v9"
)

type App struct {
	router http.Handler
	rdb    *redis.Client
}

func New() *App {
	app := &App{
		router: loadRoutes(),
		rdb:    redis.NewClient(&redis.Options{}),
	}
	return app
}

func (app *App) Start(ctx context.Context) error {
	server := &http.Server{
		Addr:    ":3000",
		Handler: app.router,
	}

	err := app.rdb.Ping(ctx).Err()

	if err != nil {
		return fmt.Errorf("Failed to listen to redis: %w", err)
	}

	fmt.Println("Starting Server...")
	ch := make(chan error, 1) // channel to receive error from server

	go func() {
		err = server.ListenAndServe()
		if err != nil {
			ch <- fmt.Errorf("Failed to listen to server: %w", err)
		}
		close(ch)
	}()

	return nil
}
