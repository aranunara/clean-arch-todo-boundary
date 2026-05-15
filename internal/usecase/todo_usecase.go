package usecase

import (
	"context"
	"strconv"
	"sync/atomic"

	"clean-arch-todo-boundary/internal/domain"
)

// TodoRepository は Usecase 層に置く。
// これらの永続化操作を必要としているのは Todo という Domain model ではなく、
// TODO 作成・変更・完了・一覧取得というアプリケーション操作だから。
type TodoRepository interface {
	Create(ctx context.Context, todo *domain.Todo) error
	FindByID(ctx context.Context, id string) (*domain.Todo, error)
	Update(ctx context.Context, todo *domain.Todo) error
	List(ctx context.Context) ([]*domain.Todo, error)
}

type TodoIDGenerator interface {
	NextID() string
}

type SequentialTodoIDGenerator struct {
	nextID atomic.Uint64
}

func NewSequentialTodoIDGenerator(initialID uint64) *SequentialTodoIDGenerator {
	generator := &SequentialTodoIDGenerator{}
	generator.nextID.Store(initialID)
	return generator
}

func (g *SequentialTodoIDGenerator) NextID() string {
	return strconv.FormatUint(g.nextID.Add(1), 10)
}

type TodoUseCase struct {
	todoRepo    TodoRepository
	idGenerator TodoIDGenerator
}

func NewTodoUseCase(todoRepo TodoRepository, idGenerator TodoIDGenerator) *TodoUseCase {
	return &TodoUseCase{
		todoRepo:    todoRepo,
		idGenerator: idGenerator,
	}
}

func NewTodoUseCaseWithInitialID(todoRepo TodoRepository, initialID uint64) *TodoUseCase {
	return NewTodoUseCase(todoRepo, NewSequentialTodoIDGenerator(initialID))
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
	id := uc.idGenerator.NextID()

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
