package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"clean-arch-todo-boundary/internal/domain"
	"clean-arch-todo-boundary/internal/usecase"
)

func TestCreateTodo(t *testing.T) {
	t.Run("passes request body to usecase", func(t *testing.T) {
		fake := &fakeTodoUseCase{}
		response := serveRequest(fake, http.MethodPost, "/todos", `{"title":"write test"}`)

		if response.Code != http.StatusCreated {
			t.Fatalf("status = %d, want %d", response.Code, http.StatusCreated)
		}
		if fake.createInput.Title != "write test" {
			t.Fatalf("title = %q, want %q", fake.createInput.Title, "write test")
		}

		var body usecase.TodoOutput
		decodeResponse(t, response, &body)
		if body.ID != "created-id" || body.Title != "write test" {
			t.Fatalf("body = %+v, want created todo", body)
		}
	})

	t.Run("returns bad request for invalid json", func(t *testing.T) {
		fake := &fakeTodoUseCase{}
		response := serveRequest(fake, http.MethodPost, "/todos", `{`)

		assertErrorResponse(t, response, http.StatusBadRequest, "invalid json")
		if fake.createCalls != 0 {
			t.Fatalf("createCalls = %d, want 0", fake.createCalls)
		}
	})
}

func TestRenameTodo(t *testing.T) {
	t.Run("passes path id and body title to usecase", func(t *testing.T) {
		fake := &fakeTodoUseCase{}
		response := serveRequest(fake, http.MethodPatch, "/todos/123", `{"title":"renamed"}`)

		if response.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
		}
		if fake.renameInput.ID != "123" {
			t.Fatalf("ID = %q, want %q", fake.renameInput.ID, "123")
		}
		if fake.renameInput.Title != "renamed" {
			t.Fatalf("Title = %q, want %q", fake.renameInput.Title, "renamed")
		}
	})

	t.Run("returns bad request for invalid json", func(t *testing.T) {
		fake := &fakeTodoUseCase{}
		response := serveRequest(fake, http.MethodPatch, "/todos/123", `{`)

		assertErrorResponse(t, response, http.StatusBadRequest, "invalid json")
		if fake.renameCalls != 0 {
			t.Fatalf("renameCalls = %d, want 0", fake.renameCalls)
		}
	})
}

func TestCompleteTodo(t *testing.T) {
	fake := &fakeTodoUseCase{}
	response := serveRequest(fake, http.MethodPost, "/todos/123/complete", "")

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}
	if fake.completeID != "123" {
		t.Fatalf("completeID = %q, want %q", fake.completeID, "123")
	}

	var body usecase.TodoOutput
	decodeResponse(t, response, &body)
	if !body.Completed {
		t.Fatal("Completed = false, want true")
	}
}

func TestDeleteTodo(t *testing.T) {
	t.Run("passes path id to usecase", func(t *testing.T) {
		fake := &fakeTodoUseCase{}
		response := serveRequest(fake, http.MethodDelete, "/todos/123", "")

		if response.Code != http.StatusNoContent {
			t.Fatalf("status = %d, want %d", response.Code, http.StatusNoContent)
		}
		if fake.deleteID != "123" {
			t.Fatalf("deleteID = %q, want %q", fake.deleteID, "123")
		}
		if fake.deleteCalls != 1 {
			t.Fatalf("deleteCalls = %d, want 1", fake.deleteCalls)
		}
		if response.Body.Len() != 0 {
			t.Fatalf("body = %q, want empty body", response.Body.String())
		}
	})

	t.Run("returns not found", func(t *testing.T) {
		fake := &fakeTodoUseCase{deleteErr: domain.ErrTodoNotFound}
		response := serveRequest(fake, http.MethodDelete, "/todos/missing", "")

		assertErrorResponse(t, response, http.StatusNotFound, "todo not found")
	})
}

func TestListTodos(t *testing.T) {
	fake := &fakeTodoUseCase{
		listOutput: []usecase.TodoOutput{
			{ID: "1", Title: "first"},
			{ID: "2", Title: "second", Completed: true},
		},
	}
	response := serveRequest(fake, http.MethodGet, "/todos", "")

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}

	var body []usecase.TodoOutput
	decodeResponse(t, response, &body)
	if len(body) != 2 {
		t.Fatalf("len(body) = %d, want 2", len(body))
	}
	if body[1].ID != "2" || !body[1].Completed {
		t.Fatalf("body[1] = %+v, want completed second todo", body[1])
	}
}

func TestUseCaseErrorResponses(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		wantStatus int
		wantError  string
	}{
		{
			name:       "invalid id",
			err:        domain.ErrInvalidTodoID,
			wantStatus: http.StatusBadRequest,
			wantError:  "invalid todo input",
		},
		{
			name:       "invalid title",
			err:        domain.ErrInvalidTodoTitle,
			wantStatus: http.StatusBadRequest,
			wantError:  "invalid todo input",
		},
		{
			name:       "not found",
			err:        domain.ErrTodoNotFound,
			wantStatus: http.StatusNotFound,
			wantError:  "todo not found",
		},
		{
			name:       "completed todo",
			err:        domain.ErrTodoCompleted,
			wantStatus: http.StatusConflict,
			wantError:  "completed todo cannot be renamed",
		},
		{
			name:       "unknown error",
			err:        errors.New("database unavailable"),
			wantStatus: http.StatusInternalServerError,
			wantError:  "internal server error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fake := &fakeTodoUseCase{createErr: tt.err}
			response := serveRequest(fake, http.MethodPost, "/todos", `{"title":"write test"}`)

			assertErrorResponse(t, response, tt.wantStatus, tt.wantError)
		})
	}
}

func TestUnknownRoute(t *testing.T) {
	fake := &fakeTodoUseCase{}
	response := serveRequest(fake, http.MethodGet, "/missing", "")

	if response.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusNotFound)
	}
}

type fakeTodoUseCase struct {
	createInput usecase.CreateTodoInput
	renameInput usecase.RenameTodoInput
	completeID  string
	deleteID    string
	listOutput  []usecase.TodoOutput

	createErr   error
	renameErr   error
	completeErr error
	deleteErr   error
	listErr     error

	createCalls int
	renameCalls int
	deleteCalls int
}

func (f *fakeTodoUseCase) CreateTodo(_ context.Context, input usecase.CreateTodoInput) (*usecase.TodoOutput, error) {
	f.createCalls++
	f.createInput = input
	if f.createErr != nil {
		return nil, f.createErr
	}
	return &usecase.TodoOutput{
		ID:    "created-id",
		Title: input.Title,
	}, nil
}

func (f *fakeTodoUseCase) RenameTodo(_ context.Context, input usecase.RenameTodoInput) (*usecase.TodoOutput, error) {
	f.renameCalls++
	f.renameInput = input
	if f.renameErr != nil {
		return nil, f.renameErr
	}
	return &usecase.TodoOutput{
		ID:    input.ID,
		Title: input.Title,
	}, nil
}

func (f *fakeTodoUseCase) CompleteTodo(_ context.Context, id string) (*usecase.TodoOutput, error) {
	f.completeID = id
	if f.completeErr != nil {
		return nil, f.completeErr
	}
	return &usecase.TodoOutput{
		ID:        id,
		Completed: true,
	}, nil
}

func (f *fakeTodoUseCase) DeleteTodo(_ context.Context, id string) error {
	f.deleteCalls++
	f.deleteID = id
	return f.deleteErr
}

func (f *fakeTodoUseCase) ListTodos(context.Context) ([]usecase.TodoOutput, error) {
	if f.listErr != nil {
		return nil, f.listErr
	}
	return f.listOutput, nil
}

func serveRequest(todoUseCase TodoUseCase, method, path, body string) *httptest.ResponseRecorder {
	handler := NewTodoHandler(todoUseCase).Routes()
	request := httptest.NewRequest(method, path, strings.NewReader(body))
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	return response
}

func decodeResponse(t *testing.T, response *httptest.ResponseRecorder, value any) {
	t.Helper()

	if err := json.NewDecoder(response.Body).Decode(value); err != nil {
		t.Fatalf("decode response: %v", err)
	}
}

func assertErrorResponse(t *testing.T, response *httptest.ResponseRecorder, wantStatus int, wantMessage string) {
	t.Helper()

	if response.Code != wantStatus {
		t.Fatalf("status = %d, want %d", response.Code, wantStatus)
	}

	var body map[string]string
	decodeResponse(t, response, &body)
	if body["error"] != wantMessage {
		t.Fatalf("error = %q, want %q", body["error"], wantMessage)
	}
}
