package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"clean-arch-todo-boundary/internal/domain"
	"clean-arch-todo-boundary/internal/handler/httpapi"
	"clean-arch-todo-boundary/internal/infra/memory"
	"clean-arch-todo-boundary/internal/infra/postgres"
	"clean-arch-todo-boundary/internal/usecase"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	todoRepo, initialID, closeRepo := openTodoRepository(ctx)
	defer closeRepo()

	todoUseCase := usecase.NewTodoUseCaseWithInitialID(todoRepo, initialID)
	todoHandler := httpapi.NewTodoHandler(todoUseCase)

	addr := ":" + envOrDefault("PORT", "8080")
	server := &http.Server{
		Addr:    addr,
		Handler: todoHandler.Routes(),
	}

	go func() {
		log.Printf("listening on http://localhost%s", addr)
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

func openTodoRepository(ctx context.Context) (domain.TodoRepository, uint64, func()) {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Println("DATABASE_URL is empty; using in-memory repository")
		return memory.NewTodoRepository(), 0, func() {}
	}

	todoRepo, err := postgres.NewTodoRepository(ctx, databaseURL)
	if err != nil {
		log.Fatalf("open postgres repository: %v", err)
	}

	initialID, err := todoRepo.MaxNumericID(ctx)
	if err != nil {
		todoRepo.Close()
		log.Fatalf("load current todo id: %v", err)
	}

	log.Println("using postgres repository")
	return todoRepo, initialID, todoRepo.Close
}

func envOrDefault(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
