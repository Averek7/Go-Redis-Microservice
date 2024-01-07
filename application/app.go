package application

import (
	"context"
	"fmt"
	"net/http"
	"time"

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

	// The `defer func() { ... }()` statement is deferring the execution of the enclosed function until the
	// surrounding function (`Start()`) returns.
	defer func() {
		if err := app.rdb.Close(); err != nil {
			fmt.Printf("Failed to close redis: %v\n", err)
		}
	}()

	fmt.Println("Starting Server...")
	ch := make(chan error, 1) // channel to receive error from server

	go func() { // go routine to start server
		err = server.ListenAndServe()
		if err != nil {
			ch <- fmt.Errorf("Failed to listen to server: %w", err)
		}
		close(ch)
	}()

	select {
	case err = <-ch:
		return err
	case <-ctx.Done(): //block is triggered when the context passed to the `Start()` method is canceled or times out. It is used to gracefully shutdown the server.
		timeout, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()
		return server.Shutdown(timeout)

	}
}
