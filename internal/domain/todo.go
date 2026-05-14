package domain

import (
	"context"
	"errors"
	"strings"
	"unicode/utf8"
)

const maxTodoTitleLength = 100

var (
	ErrInvalidTodoID    = errors.New("invalid todo id")
	ErrInvalidTodoTitle = errors.New("invalid todo title")
	ErrTodoCompleted    = errors.New("todo already completed")
	ErrTodoNotFound     = errors.New("todo not found")
)

// Todo は TODO という業務概念を表す。
// HTTP や DB の都合ではなく、TODO としての状態とルールだけを持つ。
type Todo struct {
	ID        string
	Title     string
	Completed bool
}

func NewTodo(id, title string) (*Todo, error) {
	if strings.TrimSpace(id) == "" {
		return nil, ErrInvalidTodoID
	}

	normalizedTitle, err := normalizeTitle(title)
	if err != nil {
		return nil, err
	}

	return &Todo{
		ID:        id,
		Title:     normalizedTitle,
		Completed: false,
	}, nil
}

func (t *Todo) Rename(title string) error {
	if t.Completed {
		return ErrTodoCompleted
	}

	normalizedTitle, err := normalizeTitle(title)
	if err != nil {
		return err
	}

	t.Title = normalizedTitle
	return nil
}

func (t *Todo) Complete() {
	t.Completed = true
}

func normalizeTitle(title string) (string, error) {
	trimmed := strings.TrimSpace(title)
	if trimmed == "" {
		return "", ErrInvalidTodoTitle
	}
	if utf8.RuneCountInString(trimmed) > maxTodoTitleLength {
		return "", ErrInvalidTodoTitle
	}
	return trimmed, nil
}

// TodoRepository は TODO の永続化に必要な契約を表す。
// SQL や保存方式は Infrastructure 側で具体化する。
type TodoRepository interface {
	Create(ctx context.Context, todo *Todo) error
	FindByID(ctx context.Context, id string) (*Todo, error)
	Update(ctx context.Context, todo *Todo) error
	List(ctx context.Context) ([]*Todo, error)
}
