package domain

import (
	"errors"
	"testing"
)

func TestNewTodo_NormalizesTitle(t *testing.T) {
	todo, err := NewTodo("1", "  read architecture note  ")
	if err != nil {
		t.Fatalf("NewTodo returned error: %v", err)
	}

	if todo.Title != "read architecture note" {
		t.Fatalf("title = %q, want %q", todo.Title, "read architecture note")
	}
}

func TestNewTodo_RejectsEmptyTitle(t *testing.T) {
	_, err := NewTodo("1", "   ")
	if !errors.Is(err, ErrInvalidTodoTitle) {
		t.Fatalf("error = %v, want %v", err, ErrInvalidTodoTitle)
	}
}

func TestRename_CompletedTodo(t *testing.T) {
	todo, err := NewTodo("1", "write note")
	if err != nil {
		t.Fatalf("NewTodo returned error: %v", err)
	}

	todo.Complete()

	if err := todo.Rename("rewrite note"); !errors.Is(err, ErrTodoCompleted) {
		t.Fatalf("Rename error = %v, want %v", err, ErrTodoCompleted)
	}
}
