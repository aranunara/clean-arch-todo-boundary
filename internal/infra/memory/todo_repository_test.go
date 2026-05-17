package memory

import (
	"context"
	"errors"
	"testing"

	"clean-arch-todo-boundary/internal/domain"
)

func TestCreateAndFindByID(t *testing.T) {
	ctx := context.Background()
	repo := NewTodoRepository()
	todo := mustNewTodo(t, "1", "write note")

	if err := repo.Create(ctx, todo); err != nil {
		t.Fatalf("Create returned error: %v", err)
	}

	todo.Title = "mutated after create"

	found, err := repo.FindByID(ctx, "1")
	if err != nil {
		t.Fatalf("FindByID returned error: %v", err)
	}
	if found.Title != "write note" {
		t.Fatalf("Title = %q, want %q", found.Title, "write note")
	}

	found.Title = "mutated after find"
	foundAgain, err := repo.FindByID(ctx, "1")
	if err != nil {
		t.Fatalf("FindByID returned error: %v", err)
	}
	if foundAgain.Title != "write note" {
		t.Fatalf("Title = %q, want %q", foundAgain.Title, "write note")
	}
}

func TestFindByID_NotFound(t *testing.T) {
	_, err := NewTodoRepository().FindByID(context.Background(), "missing")
	if !errors.Is(err, domain.ErrTodoNotFound) {
		t.Fatalf("error = %v, want %v", err, domain.ErrTodoNotFound)
	}
}

func TestUpdate(t *testing.T) {
	ctx := context.Background()
	repo := NewTodoRepository()
	todo := mustNewTodo(t, "1", "write note")
	if err := repo.Create(ctx, todo); err != nil {
		t.Fatalf("Create returned error: %v", err)
	}

	renamed := mustNewTodo(t, "1", "rewrite note")
	renamed.Complete()
	if err := repo.Update(ctx, renamed); err != nil {
		t.Fatalf("Update returned error: %v", err)
	}

	renamed.Title = "mutated after update"

	found, err := repo.FindByID(ctx, "1")
	if err != nil {
		t.Fatalf("FindByID returned error: %v", err)
	}
	if found.Title != "rewrite note" {
		t.Fatalf("Title = %q, want %q", found.Title, "rewrite note")
	}
	if !found.Completed {
		t.Fatal("Completed = false, want true")
	}
}

func TestUpdate_NotFound(t *testing.T) {
	todo := mustNewTodo(t, "missing", "write note")

	err := NewTodoRepository().Update(context.Background(), todo)
	if !errors.Is(err, domain.ErrTodoNotFound) {
		t.Fatalf("error = %v, want %v", err, domain.ErrTodoNotFound)
	}
}

func TestDelete(t *testing.T) {
	ctx := context.Background()
	repo := NewTodoRepository()
	if err := repo.Create(ctx, mustNewTodo(t, "1", "write note")); err != nil {
		t.Fatalf("Create returned error: %v", err)
	}

	if err := repo.Delete(ctx, "1"); err != nil {
		t.Fatalf("Delete returned error: %v", err)
	}

	_, err := repo.FindByID(ctx, "1")
	if !errors.Is(err, domain.ErrTodoNotFound) {
		t.Fatalf("error = %v, want %v", err, domain.ErrTodoNotFound)
	}
}

func TestDelete_NotFound(t *testing.T) {
	err := NewTodoRepository().Delete(context.Background(), "missing")
	if !errors.Is(err, domain.ErrTodoNotFound) {
		t.Fatalf("error = %v, want %v", err, domain.ErrTodoNotFound)
	}
}

func TestList(t *testing.T) {
	ctx := context.Background()
	repo := NewTodoRepository()
	for _, todo := range []*domain.Todo{
		mustNewTodo(t, "2", "second"),
		mustNewTodo(t, "1", "first"),
	} {
		if err := repo.Create(ctx, todo); err != nil {
			t.Fatalf("Create returned error: %v", err)
		}
	}

	todos, err := repo.List(ctx)
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}
	if len(todos) != 2 {
		t.Fatalf("len(todos) = %d, want 2", len(todos))
	}
	if todos[0].ID != "1" || todos[1].ID != "2" {
		t.Fatalf("ids = [%s, %s], want [1, 2]", todos[0].ID, todos[1].ID)
	}

	todos[0].Title = "mutated after list"
	found, err := repo.FindByID(ctx, "1")
	if err != nil {
		t.Fatalf("FindByID returned error: %v", err)
	}
	if found.Title != "first" {
		t.Fatalf("Title = %q, want %q", found.Title, "first")
	}
}

func TestMaxNumericID(t *testing.T) {
	ctx := context.Background()
	repo := NewTodoRepository()
	for _, todo := range []*domain.Todo{
		mustNewTodo(t, "2", "second"),
		mustNewTodo(t, "9223372036854775808", "larger than int64 max"),
		mustNewTodo(t, "note-3", "manual note"),
	} {
		if err := repo.Create(ctx, todo); err != nil {
			t.Fatalf("Create returned error: %v", err)
		}
	}

	maxID, err := repo.MaxNumericID(ctx)
	if err != nil {
		t.Fatalf("MaxNumericID returned error: %v", err)
	}
	if maxID != 1<<63 {
		t.Fatalf("maxID = %d, want %d", maxID, uint64(1)<<63)
	}
}

func TestMaxNumericID_Empty(t *testing.T) {
	maxID, err := NewTodoRepository().MaxNumericID(context.Background())
	if err != nil {
		t.Fatalf("MaxNumericID returned error: %v", err)
	}
	if maxID != 0 {
		t.Fatalf("maxID = %d, want 0", maxID)
	}
}

func mustNewTodo(t *testing.T, id, title string) *domain.Todo {
	t.Helper()

	todo, err := domain.NewTodo(id, title)
	if err != nil {
		t.Fatalf("NewTodo returned error: %v", err)
	}
	return todo
}
