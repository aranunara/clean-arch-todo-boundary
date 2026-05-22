package httpapi

import (
	"context"
	"encoding/json"
	"net/http"

	"clean-arch-todo-boundary/internal/errs"
	"clean-arch-todo-boundary/internal/usecase"
)

type TodoUseCase interface {
	CreateTodo(ctx context.Context, input usecase.CreateTodoInput) (*usecase.TodoOutput, error)
	RenameTodo(ctx context.Context, input usecase.RenameTodoInput) (*usecase.TodoOutput, error)
	CompleteTodo(ctx context.Context, id string) (*usecase.TodoOutput, error)
	DeleteTodo(ctx context.Context, id string) error
	ListTodos(ctx context.Context) ([]usecase.TodoOutput, error)
}

type TodoHandler struct {
	todoUseCase TodoUseCase
}

func NewTodoHandler(todoUseCase TodoUseCase) *TodoHandler {
	return &TodoHandler{todoUseCase: todoUseCase}
}

func (h *TodoHandler) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /todos", h.listTodos)
	mux.HandleFunc("POST /todos", h.createTodo)
	mux.HandleFunc("PATCH /todos/{id}", h.renameTodo)
	mux.HandleFunc("POST /todos/{id}/complete", h.completeTodo)
	mux.HandleFunc("DELETE /todos/{id}", h.deleteTodo)
	return mux
}

func (h *TodoHandler) createTodo(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Title string `json:"title"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}

	output, err := h.todoUseCase.CreateTodo(r.Context(), usecase.CreateTodoInput{
		Title: body.Title,
	})
	if err != nil {
		writeUseCaseError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, output)
}

func (h *TodoHandler) renameTodo(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Title string `json:"title"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}

	output, err := h.todoUseCase.RenameTodo(r.Context(), usecase.RenameTodoInput{
		ID:    r.PathValue("id"),
		Title: body.Title,
	})
	if err != nil {
		writeUseCaseError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, output)
}

func (h *TodoHandler) completeTodo(w http.ResponseWriter, r *http.Request) {
	output, err := h.todoUseCase.CompleteTodo(r.Context(), r.PathValue("id"))
	if err != nil {
		writeUseCaseError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, output)
}

func (h *TodoHandler) deleteTodo(w http.ResponseWriter, r *http.Request) {
	if err := h.todoUseCase.DeleteTodo(r.Context(), r.PathValue("id")); err != nil {
		writeUseCaseError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *TodoHandler) listTodos(w http.ResponseWriter, r *http.Request) {
	outputs, err := h.todoUseCase.ListTodos(r.Context())
	if err != nil {
		writeUseCaseError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, outputs)
}

func writeUseCaseError(w http.ResponseWriter, err error) {
	switch {
	case errs.IsInvalidInput(err):
		writeError(w, http.StatusBadRequest, "invalid todo input")
	case errs.IsNotFound(err):
		writeError(w, http.StatusNotFound, "todo not found")
	case errs.IsConflict(err):
		writeError(w, http.StatusConflict, "completed todo cannot be renamed")
	default:
		writeError(w, http.StatusInternalServerError, "internal server error")
	}
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
