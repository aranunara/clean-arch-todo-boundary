package domain

import (
	"strings"
	"testing"

	"clean-arch-todo-boundary/internal/errs"
)

func TestNewTodo(t *testing.T) {
	tests := []struct {
		name      string
		id        string
		title     string
		wantTitle string
		wantErr   error
	}{
		{
			name:      "normalizes title",
			id:        "1",
			title:     "  read architecture note  ",
			wantTitle: "read architecture note",
		},
		{
			name:      "accepts 100 rune title",
			id:        "1",
			title:     strings.Repeat("あ", 100),
			wantTitle: strings.Repeat("あ", 100),
		},
		{
			name:    "rejects empty id",
			id:      "  ",
			title:   "write note",
			wantErr: ErrInvalidTodoID,
		},
		{
			name:    "rejects empty title",
			id:      "1",
			title:   "   ",
			wantErr: ErrInvalidTodoTitle,
		},
		{
			name:    "rejects title longer than 100 runes",
			id:      "1",
			title:   strings.Repeat("あ", 101),
			wantErr: ErrInvalidTodoTitle,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			todo, err := NewTodo(tt.id, tt.title)
			if !errs.Is(err, tt.wantErr) {
				t.Fatalf("error = %v, want %v", err, tt.wantErr)
			}
			if tt.wantErr != nil {
				return
			}

			if todo.ID != tt.id {
				t.Fatalf("ID = %q, want %q", todo.ID, tt.id)
			}
			if todo.Title != tt.wantTitle {
				t.Fatalf("Title = %q, want %q", todo.Title, tt.wantTitle)
			}
			if todo.Completed {
				t.Fatal("Completed = true, want false")
			}
		})
	}
}

func TestRename(t *testing.T) {
	tests := []struct {
		name      string
		initial   string
		next      string
		completed bool
		wantTitle string
		wantErr   error
	}{
		{
			name:      "renames and normalizes title",
			initial:   "write note",
			next:      "  rewrite note  ",
			wantTitle: "rewrite note",
		},
		{
			name:      "rejects empty title",
			initial:   "write note",
			next:      " ",
			wantTitle: "write note",
			wantErr:   ErrInvalidTodoTitle,
		},
		{
			name:      "rejects completed todo",
			initial:   "write note",
			next:      "rewrite note",
			completed: true,
			wantTitle: "write note",
			wantErr:   ErrTodoCompleted,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			todo, err := NewTodo("1", tt.initial)
			if err != nil {
				t.Fatalf("NewTodo returned error: %v", err)
			}
			if tt.completed {
				todo.Complete()
			}

			err = todo.Rename(tt.next)
			if !errs.Is(err, tt.wantErr) {
				t.Fatalf("error = %v, want %v", err, tt.wantErr)
			}
			if todo.Title != tt.wantTitle {
				t.Fatalf("Title = %q, want %q", todo.Title, tt.wantTitle)
			}
		})
	}
}

func TestComplete(t *testing.T) {
	todo, err := NewTodo("1", "write note")
	if err != nil {
		t.Fatalf("NewTodo returned error: %v", err)
	}

	todo.Complete()

	if !todo.Completed {
		t.Fatal("Completed = false, want true")
	}
}
