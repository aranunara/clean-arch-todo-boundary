package usecase

import (
	"context"
	"strconv"
	"sync/atomic"

	"clean-arch-todo-boundary/internal/domain"
)

type TodoUseCase struct {
	todoRepo domain.TodoRepository
	nextID   atomic.Uint64
}

func NewTodoUseCase(todoRepo domain.TodoRepository) *TodoUseCase {
	return &TodoUseCase{todoRepo: todoRepo}
}

func NewTodoUseCaseWithInitialID(todoRepo domain.TodoRepository, initialID uint64) *TodoUseCase {
	uc := NewTodoUseCase(todoRepo)
	uc.nextID.Store(initialID)
	return uc
}

type CreateTodoInput struct {
	Title string
}

type RenameTodoInput struct {
	ID    string
	Title string
}

type TodoOutput struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Completed bool   `json:"completed"`
}

func (uc *TodoUseCase) CreateTodo(ctx context.Context, input CreateTodoInput) (*TodoOutput, error) {
	id := strconv.FormatUint(uc.nextID.Add(1), 10)

	todo, err := domain.NewTodo(id, input.Title)
	if err != nil {
		return nil, err
	}

	if err := uc.todoRepo.Create(ctx, todo); err != nil {
		return nil, err
	}

	return toTodoOutput(todo), nil
}

func (uc *TodoUseCase) RenameTodo(ctx context.Context, input RenameTodoInput) (*TodoOutput, error) {
	todo, err := uc.todoRepo.FindByID(ctx, input.ID)
	if err != nil {
		return nil, err
	}

	if err := todo.Rename(input.Title); err != nil {
		return nil, err
	}

	if err := uc.todoRepo.Update(ctx, todo); err != nil {
		return nil, err
	}

	return toTodoOutput(todo), nil
}

func (uc *TodoUseCase) CompleteTodo(ctx context.Context, id string) (*TodoOutput, error) {
	todo, err := uc.todoRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	todo.Complete()

	if err := uc.todoRepo.Update(ctx, todo); err != nil {
		return nil, err
	}

	return toTodoOutput(todo), nil
}

func (uc *TodoUseCase) ListTodos(ctx context.Context) ([]TodoOutput, error) {
	todos, err := uc.todoRepo.List(ctx)
	if err != nil {
		return nil, err
	}

	outputs := make([]TodoOutput, 0, len(todos))
	for _, todo := range todos {
		outputs = append(outputs, *toTodoOutput(todo))
	}
	return outputs, nil
}

func toTodoOutput(todo *domain.Todo) *TodoOutput {
	return &TodoOutput{
		ID:        todo.ID,
		Title:     todo.Title,
		Completed: todo.Completed,
	}
}
