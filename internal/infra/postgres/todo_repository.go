package postgres

import (
	"context"
	"strconv"

	"clean-arch-todo-boundary/internal/domain"
	"clean-arch-todo-boundary/internal/errs"
	"clean-arch-todo-boundary/internal/usecase"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TodoRepository struct {
	pool *pgxpool.Pool
}

// TodoRepository が UsecaseのTodoRepository を満たしているか
var _ usecase.TodoRepository = (*TodoRepository)(nil)

func NewTodoRepository(ctx context.Context, databaseURL string) (*TodoRepository, error) {
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, err
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, err
	}

	return &TodoRepository{pool: pool}, nil
}

func (r *TodoRepository) Close() {
	r.pool.Close()
}

func (r *TodoRepository) Create(ctx context.Context, todo *domain.Todo) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO todos (id, title, completed)
		VALUES ($1, $2, $3)
	`, todo.ID, todo.Title, todo.Completed)
	return err
}

func (r *TodoRepository) FindByID(ctx context.Context, id string) (*domain.Todo, error) {
	todo, err := scanTodo(r.pool.QueryRow(ctx, `
		SELECT id, title, completed
		FROM todos
		WHERE id = $1
	`, id))
	if errs.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrTodoNotFound
	}
	return todo, err
}

func (r *TodoRepository) Update(ctx context.Context, todo *domain.Todo) error {
	tag, err := r.pool.Exec(ctx, `
		UPDATE todos
		SET title = $2,
			completed = $3,
			updated_at = now()
		WHERE id = $1
	`, todo.ID, todo.Title, todo.Completed)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrTodoNotFound
	}
	return nil
}

func (r *TodoRepository) Delete(ctx context.Context, id string) error {
	tag, err := r.pool.Exec(ctx, `
		DELETE FROM todos
		WHERE id = $1
	`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrTodoNotFound
	}
	return nil
}

func (r *TodoRepository) List(ctx context.Context) ([]*domain.Todo, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, title, completed
		FROM todos
		ORDER BY id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	todos := make([]*domain.Todo, 0)
	for rows.Next() {
		todo, err := scanTodo(rows)
		if err != nil {
			return nil, err
		}
		todos = append(todos, todo)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return todos, nil
}

func (r *TodoRepository) MaxNumericID(ctx context.Context) (uint64, error) {
	var maxID string
	if err := r.pool.QueryRow(ctx, `
		SELECT COALESCE(MAX(id::numeric), 0)::text
		FROM todos
		WHERE id ~ '^[0-9]+$'
	`).Scan(&maxID); err != nil {
		return 0, err
	}
	return strconv.ParseUint(maxID, 10, 64)
}

func scanTodo(row pgx.Row) (*domain.Todo, error) {
	var todo domain.Todo
	if err := row.Scan(&todo.ID, &todo.Title, &todo.Completed); err != nil {
		return nil, err
	}
	return &todo, nil
}
