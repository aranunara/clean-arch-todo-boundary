package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"clean-arch-todo-boundary/internal/di"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	container, err := di.NewContainer(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer container.Close()

	server := &http.Server{
		Addr:    container.HTTPAddr(),
		Handler: container.HTTPHandler(),
	}

	go func() {
		log.Printf("listening on http://localhost%s", container.HTTPAddr())
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatal(err)
	}
}
