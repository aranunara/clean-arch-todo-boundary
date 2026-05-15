package memory

import (
	"context"
	"sort"
	"sync"

	"clean-arch-todo-boundary/internal/domain"
)

// TodoRepository は usecase.TodoRepository を in-memory map で具体化する。
// 実アプリではこの層が PostgreSQL の INSERT/SELECT/UPDATE などに置き換わる。
type TodoRepository struct {
	mu    sync.RWMutex
	todos map[string]*domain.Todo
}

func NewTodoRepository() *TodoRepository {
	return &TodoRepository{todos: make(map[string]*domain.Todo)}
}

func (r *TodoRepository) Create(_ context.Context, todo *domain.Todo) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.todos[todo.ID] = cloneTodo(todo)
	return nil
}

func (r *TodoRepository) FindByID(_ context.Context, id string) (*domain.Todo, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	todo, ok := r.todos[id]
	if !ok {
		return nil, domain.ErrTodoNotFound
	}
	return cloneTodo(todo), nil
}

func (r *TodoRepository) Update(_ context.Context, todo *domain.Todo) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.todos[todo.ID]; !ok {
		return domain.ErrTodoNotFound
	}

	r.todos[todo.ID] = cloneTodo(todo)
	return nil
}

func (r *TodoRepository) List(_ context.Context) ([]*domain.Todo, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	ids := make([]string, 0, len(r.todos))
	for id := range r.todos {
		ids = append(ids, id)
	}
	sort.Strings(ids)

	todos := make([]*domain.Todo, 0, len(ids))
	for _, id := range ids {
		todos = append(todos, cloneTodo(r.todos[id]))
	}
	return todos, nil
}

func cloneTodo(todo *domain.Todo) *domain.Todo {
	copied := *todo
	return &copied
}
