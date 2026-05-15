package di

import (
	"context"
	"log"
	"net/http"
	"os"

	"clean-arch-todo-boundary/internal/handler/httpapi"
	"clean-arch-todo-boundary/internal/infra/memory"
	"clean-arch-todo-boundary/internal/infra/postgres"
	"clean-arch-todo-boundary/internal/usecase"
)

type Container struct {
	httpAddr    string
	httpHandler http.Handler
	closeFuncs  []func()
}

func NewContainer(ctx context.Context) (*Container, error) {
	todoRepo, initialID, closeRepo, err := buildTodoRepository(ctx)
	if err != nil {
		return nil, err
	}

	idGenerator := usecase.NewSequentialTodoIDGenerator(initialID)
	todoUseCase := usecase.NewTodoUseCase(todoRepo, idGenerator)
	todoHandler := httpapi.NewTodoHandler(todoUseCase)

	return &Container{
		httpAddr:    ":" + envOrDefault("PORT", "8080"),
		httpHandler: todoHandler.Routes(),
		closeFuncs:  []func(){closeRepo},
	}, nil
}

func (c *Container) HTTPAddr() string {
	return c.httpAddr
}

func (c *Container) HTTPHandler() http.Handler {
	return c.httpHandler
}

func (c *Container) Close() {
	for i := len(c.closeFuncs) - 1; i >= 0; i-- {
		c.closeFuncs[i]()
	}
}

func buildTodoRepository(ctx context.Context) (usecase.TodoRepository, uint64, func(), error) {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Println("DATABASE_URL is empty; using in-memory repository")
		return memory.NewTodoRepository(), 0, func() {}, nil
	}

	todoRepo, err := postgres.NewTodoRepository(ctx, databaseURL)
	if err != nil {
		return nil, 0, nil, err
	}

	initialID, err := todoRepo.MaxNumericID(ctx)
	if err != nil {
		todoRepo.Close()
		return nil, 0, nil, err
	}

	log.Println("using postgres repository")
	return todoRepo, initialID, todoRepo.Close, nil
}

func envOrDefault(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
