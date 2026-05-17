package usecase

import (
	"context"
	"errors"
	"sort"
	"testing"

	"clean-arch-todo-boundary/internal/domain"
)

func TestSequentialTodoIDGenerator(t *testing.T) {
	generator := NewSequentialTodoIDGenerator(41)

	if got := generator.NextID(); got != "42" {
		t.Fatalf("first ID = %q, want %q", got, "42")
	}
	if got := generator.NextID(); got != "43" {
		t.Fatalf("second ID = %q, want %q", got, "43")
	}
}

func TestCreateTodo(t *testing.T) {
	t.Run("creates todo with generated id", func(t *testing.T) {
		repo := newFakeTodoRepository()
		uc := NewTodoUseCase(repo, fixedIDGenerator{id: "todo-1"})

		output, err := uc.CreateTodo(context.Background(), CreateTodoInput{
			Title: "  record architecture boundary  ",
		})
		if err != nil {
			t.Fatalf("CreateTodo returned error: %v", err)
		}

		if output.ID != "todo-1" {
			t.Fatalf("ID = %q, want %q", output.ID, "todo-1")
		}
		if output.Title != "record architecture boundary" {
			t.Fatalf("Title = %q, want %q", output.Title, "record architecture boundary")
		}
		if output.Completed {
			t.Fatal("Completed = true, want false")
		}
		if repo.createCalls != 1 {
			t.Fatalf("createCalls = %d, want 1", repo.createCalls)
		}
	})

	t.Run("returns domain validation error without saving", func(t *testing.T) {
		repo := newFakeTodoRepository()
		uc := NewTodoUseCase(repo, fixedIDGenerator{id: "todo-1"})

		_, err := uc.CreateTodo(context.Background(), CreateTodoInput{Title: " "})
		if !errors.Is(err, domain.ErrInvalidTodoTitle) {
			t.Fatalf("error = %v, want %v", err, domain.ErrInvalidTodoTitle)
		}
		if repo.createCalls != 0 {
			t.Fatalf("createCalls = %d, want 0", repo.createCalls)
		}
	})

	t.Run("propagates repository error", func(t *testing.T) {
		wantErr := errors.New("create failed")
		repo := newFakeTodoRepository()
		repo.createErr = wantErr
		uc := NewTodoUseCase(repo, fixedIDGenerator{id: "todo-1"})

		_, err := uc.CreateTodo(context.Background(), CreateTodoInput{Title: "write note"})
		if !errors.Is(err, wantErr) {
			t.Fatalf("error = %v, want %v", err, wantErr)
		}
	})
}

func TestRenameTodo(t *testing.T) {
	t.Run("renames existing todo", func(t *testing.T) {
		repo := newFakeTodoRepository()
		repo.mustStore(t, "1", "write note")
		uc := NewTodoUseCase(repo, fixedIDGenerator{id: "unused"})

		output, err := uc.RenameTodo(context.Background(), RenameTodoInput{
			ID:    "1",
			Title: "  rewrite note  ",
		})
		if err != nil {
			t.Fatalf("RenameTodo returned error: %v", err)
		}

		if output.Title != "rewrite note" {
			t.Fatalf("Title = %q, want %q", output.Title, "rewrite note")
		}
		if repo.updateCalls != 1 {
			t.Fatalf("updateCalls = %d, want 1", repo.updateCalls)
		}
	})

	t.Run("returns not found from repository", func(t *testing.T) {
		repo := newFakeTodoRepository()
		uc := NewTodoUseCase(repo, fixedIDGenerator{id: "unused"})

		_, err := uc.RenameTodo(context.Background(), RenameTodoInput{
			ID:    "missing",
			Title: "rewrite note",
		})
		if !errors.Is(err, domain.ErrTodoNotFound) {
			t.Fatalf("error = %v, want %v", err, domain.ErrTodoNotFound)
		}
		if repo.updateCalls != 0 {
			t.Fatalf("updateCalls = %d, want 0", repo.updateCalls)
		}
	})

	t.Run("does not update when domain rejects rename", func(t *testing.T) {
		repo := newFakeTodoRepository()
		todo := repo.mustStore(t, "1", "write note")
		todo.Complete()
		uc := NewTodoUseCase(repo, fixedIDGenerator{id: "unused"})

		_, err := uc.RenameTodo(context.Background(), RenameTodoInput{
			ID:    "1",
			Title: "rewrite note",
		})
		if !errors.Is(err, domain.ErrTodoCompleted) {
			t.Fatalf("error = %v, want %v", err, domain.ErrTodoCompleted)
		}
		if repo.updateCalls != 0 {
			t.Fatalf("updateCalls = %d, want 0", repo.updateCalls)
		}
	})
}

func TestCompleteTodo(t *testing.T) {
	t.Run("completes existing todo", func(t *testing.T) {
		repo := newFakeTodoRepository()
		repo.mustStore(t, "1", "write note")
		uc := NewTodoUseCase(repo, fixedIDGenerator{id: "unused"})

		output, err := uc.CompleteTodo(context.Background(), "1")
		if err != nil {
			t.Fatalf("CompleteTodo returned error: %v", err)
		}

		if !output.Completed {
			t.Fatal("Completed = false, want true")
		}
		if repo.updateCalls != 1 {
			t.Fatalf("updateCalls = %d, want 1", repo.updateCalls)
		}
	})

	t.Run("returns not found", func(t *testing.T) {
		repo := newFakeTodoRepository()
		uc := NewTodoUseCase(repo, fixedIDGenerator{id: "unused"})

		_, err := uc.CompleteTodo(context.Background(), "missing")
		if !errors.Is(err, domain.ErrTodoNotFound) {
			t.Fatalf("error = %v, want %v", err, domain.ErrTodoNotFound)
		}
		if repo.updateCalls != 0 {
			t.Fatalf("updateCalls = %d, want 0", repo.updateCalls)
		}
	})
}

func TestListTodos(t *testing.T) {
	t.Run("returns todo outputs", func(t *testing.T) {
		repo := newFakeTodoRepository()
		repo.mustStore(t, "1", "first")
		completed := repo.mustStore(t, "2", "second")
		completed.Complete()
		uc := NewTodoUseCase(repo, fixedIDGenerator{id: "unused"})

		outputs, err := uc.ListTodos(context.Background())
		if err != nil {
			t.Fatalf("ListTodos returned error: %v", err)
		}

		if len(outputs) != 2 {
			t.Fatalf("len(outputs) = %d, want 2", len(outputs))
		}
		if outputs[0].ID != "1" || outputs[0].Title != "first" || outputs[0].Completed {
			t.Fatalf("outputs[0] = %+v, want first incomplete todo", outputs[0])
		}
		if outputs[1].ID != "2" || outputs[1].Title != "second" || !outputs[1].Completed {
			t.Fatalf("outputs[1] = %+v, want second completed todo", outputs[1])
		}
	})

	t.Run("propagates repository error", func(t *testing.T) {
		wantErr := errors.New("list failed")
		repo := newFakeTodoRepository()
		repo.listErr = wantErr
		uc := NewTodoUseCase(repo, fixedIDGenerator{id: "unused"})

		_, err := uc.ListTodos(context.Background())
		if !errors.Is(err, wantErr) {
			t.Fatalf("error = %v, want %v", err, wantErr)
		}
	})
}

func TestDeleteTodo(t *testing.T) {
	t.Run("deletes existing todo", func(t *testing.T) {
		repo := newFakeTodoRepository()
		repo.mustStore(t, "1", "write note")
		uc := NewTodoUseCase(repo, fixedIDGenerator{id: "unused"})

		if err := uc.DeleteTodo(context.Background(), "1"); err != nil {
			t.Fatalf("DeleteTodo returned error: %v", err)
		}
		if repo.deleteCalls != 1 {
			t.Fatalf("deleteCalls = %d, want 1", repo.deleteCalls)
		}
		if _, ok := repo.todos["1"]; ok {
			t.Fatal("todo still exists after delete")
		}
	})

	t.Run("returns repository error", func(t *testing.T) {
		wantErr := errors.New("delete failed")
		repo := newFakeTodoRepository()
		repo.deleteErr = wantErr
		uc := NewTodoUseCase(repo, fixedIDGenerator{id: "unused"})

		err := uc.DeleteTodo(context.Background(), "1")
		if !errors.Is(err, wantErr) {
			t.Fatalf("error = %v, want %v", err, wantErr)
		}
		if repo.deleteCalls != 1 {
			t.Fatalf("deleteCalls = %d, want 1", repo.deleteCalls)
		}
	})
}

type fixedIDGenerator struct {
	id string
}

func (g fixedIDGenerator) NextID() string {
	return g.id
}

type fakeTodoRepository struct {
	todos       map[string]*domain.Todo
	createErr   error
	updateErr   error
	deleteErr   error
	listErr     error
	createCalls int
	updateCalls int
	deleteCalls int
}

func newFakeTodoRepository() *fakeTodoRepository {
	return &fakeTodoRepository{todos: make(map[string]*domain.Todo)}
}

func (r *fakeTodoRepository) Create(_ context.Context, todo *domain.Todo) error {
	r.createCalls++
	if r.createErr != nil {
		return r.createErr
	}
	r.todos[todo.ID] = cloneTodo(todo)
	return nil
}

func (r *fakeTodoRepository) FindByID(_ context.Context, id string) (*domain.Todo, error) {
	todo, ok := r.todos[id]
	if !ok {
		return nil, domain.ErrTodoNotFound
	}
	return cloneTodo(todo), nil
}

func (r *fakeTodoRepository) Update(_ context.Context, todo *domain.Todo) error {
	r.updateCalls++
	if r.updateErr != nil {
		return r.updateErr
	}
	if _, ok := r.todos[todo.ID]; !ok {
		return domain.ErrTodoNotFound
	}
	r.todos[todo.ID] = cloneTodo(todo)
	return nil
}

func (r *fakeTodoRepository) Delete(_ context.Context, id string) error {
	r.deleteCalls++
	if r.deleteErr != nil {
		return r.deleteErr
	}
	if _, ok := r.todos[id]; !ok {
		return domain.ErrTodoNotFound
	}
	delete(r.todos, id)
	return nil
}

func (r *fakeTodoRepository) List(context.Context) ([]*domain.Todo, error) {
	if r.listErr != nil {
		return nil, r.listErr
	}

	ids := make([]string, 0, len(r.todos))
	for id := range r.todos {
		ids = append(ids, id)
	}
	sort.Strings(ids)

	todos := make([]*domain.Todo, 0, len(r.todos))
	for _, id := range ids {
		todos = append(todos, cloneTodo(r.todos[id]))
	}
	return todos, nil
}

func (r *fakeTodoRepository) MaxNumericID(context.Context) (uint64, error) {
	return 0, nil
}

func (r *fakeTodoRepository) mustStore(t *testing.T, id, title string) *domain.Todo {
	t.Helper()

	todo, err := domain.NewTodo(id, title)
	if err != nil {
		t.Fatalf("NewTodo returned error: %v", err)
	}
	r.todos[id] = todo
	return todo
}

func cloneTodo(todo *domain.Todo) *domain.Todo {
	copied := *todo
	return &copied
}
