package domain

import (
	"strings"
	"unicode/utf8"

	"clean-arch-todo-boundary/internal/errs"
)

const maxTodoTitleLength = 100

var (
	ErrInvalidTodoID    = errs.Mark(errs.New("invalid todo id"), errs.ErrInvalidInput)
	ErrInvalidTodoTitle = errs.Mark(errs.New("invalid todo title"), errs.ErrInvalidInput)
	ErrTodoCompleted    = errs.Mark(errs.New("todo already completed"), errs.ErrConflict)
	ErrTodoNotFound     = errs.Mark(errs.New("todo not found"), errs.ErrNotFound)
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
