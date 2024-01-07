package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/averek7/order-api/application"
)

func main() {
	app := application.New()
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel() //Should be called before app.Start(ctx) to avoid goroutine leak

	err := app.Start(ctx)
	if err != nil {
		fmt.Printf("Failed to start application: %v\n", err)
	}

	cancel()
}
