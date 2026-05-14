package usecase

import (
	"context"
	"errors"
	"testing"

	"clean-arch-todo-boundary/internal/domain"
	"clean-arch-todo-boundary/internal/infra/memory"
)

func TestCreateTodo(t *testing.T) {
	uc := NewTodoUseCase(memory.NewTodoRepository())

	output, err := uc.CreateTodo(context.Background(), CreateTodoInput{
		Title: "record architecture boundary",
	})
	if err != nil {
		t.Fatalf("CreateTodo returned error: %v", err)
	}

	if output.ID == "" {
		t.Fatal("ID is empty")
	}
	if output.Completed {
		t.Fatal("Completed = true, want false")
	}
}

func TestRenameTodo_CompletedTodo(t *testing.T) {
	ctx := context.Background()
	uc := NewTodoUseCase(memory.NewTodoRepository())

	created, err := uc.CreateTodo(ctx, CreateTodoInput{Title: "write note"})
	if err != nil {
		t.Fatalf("CreateTodo returned error: %v", err)
	}
	if _, err := uc.CompleteTodo(ctx, created.ID); err != nil {
		t.Fatalf("CompleteTodo returned error: %v", err)
	}

	_, err = uc.RenameTodo(ctx, RenameTodoInput{
		ID:    created.ID,
		Title: "rewrite note",
	})
	if !errors.Is(err, domain.ErrTodoCompleted) {
		t.Fatalf("RenameTodo error = %v, want %v", err, domain.ErrTodoCompleted)
	}
}
