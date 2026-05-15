package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"clean-arch-todo-boundary/internal/usecase"
)

func TestCreateTodo_UsesUseCaseInterface(t *testing.T) {
	fake := &fakeTodoUseCase{}
	handler := NewTodoHandler(fake).Routes()

	request := httptest.NewRequest(http.MethodPost, "/todos", strings.NewReader(`{"title":"write test"}`))
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusCreated)
	}
	if fake.createInput.Title != "write test" {
		t.Fatalf("title = %q, want %q", fake.createInput.Title, "write test")
	}

	var body usecase.TodoOutput
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.ID != "fake-id" {
		t.Fatalf("id = %q, want %q", body.ID, "fake-id")
	}
}

type fakeTodoUseCase struct {
	createInput usecase.CreateTodoInput
}

func (f *fakeTodoUseCase) CreateTodo(_ context.Context, input usecase.CreateTodoInput) (*usecase.TodoOutput, error) {
	f.createInput = input
	return &usecase.TodoOutput{
		ID:    "fake-id",
		Title: input.Title,
	}, nil
}

func (f *fakeTodoUseCase) RenameTodo(_ context.Context, input usecase.RenameTodoInput) (*usecase.TodoOutput, error) {
	return &usecase.TodoOutput{
		ID:    input.ID,
		Title: input.Title,
	}, nil
}

func (f *fakeTodoUseCase) CompleteTodo(_ context.Context, id string) (*usecase.TodoOutput, error) {
	return &usecase.TodoOutput{
		ID:        id,
		Completed: true,
	}, nil
}

func (f *fakeTodoUseCase) ListTodos(context.Context) ([]usecase.TodoOutput, error) {
	return nil, nil
}
