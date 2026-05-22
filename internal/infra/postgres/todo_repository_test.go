//go:build pg_integration

package postgres

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"testing"
	"time"

	"clean-arch-todo-boundary/internal/domain"
	"clean-arch-todo-boundary/internal/errs"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
)

const (
	postgresImage = "postgres:18-alpine"
	testDBName    = "todo_test"
	testDBUser    = "todo"
	testDBPass    = "todo"
)

func TestTodoRepository_Postgres(t *testing.T) {
	ctx := context.Background()
	repo, pool := newPostgresRepositoryFixture(t, ctx)

	t.Run("finds seed by id", func(t *testing.T) {
		resetTodos(t, ctx, pool)
		seedTodos(t, ctx, pool,
			todoSeed{id: "1", title: "first", completed: false},
			todoSeed{id: "2", title: "second", completed: true},
		)

		found, err := repo.FindByID(ctx, "2")
		if err != nil {
			t.Fatalf("FindByID returned error: %v", err)
		}
		if found.ID != "2" || found.Title != "second" || !found.Completed {
			t.Fatalf("todo = %+v, want completed second todo", found)
		}
	})

	t.Run("find returns not found", func(t *testing.T) {
		resetTodos(t, ctx, pool)

		_, err := repo.FindByID(ctx, "missing")
		if !errs.Is(err, domain.ErrTodoNotFound) {
			t.Fatalf("error = %v, want %v", err, domain.ErrTodoNotFound)
		}
	})

	t.Run("creates and finds todo", func(t *testing.T) {
		resetTodos(t, ctx, pool)
		todo := mustNewTodo(t, "created", "write postgres test")

		if err := repo.Create(ctx, todo); err != nil {
			t.Fatalf("Create returned error: %v", err)
		}

		found, err := repo.FindByID(ctx, "created")
		if err != nil {
			t.Fatalf("FindByID returned error: %v", err)
		}
		if found.Title != "write postgres test" || found.Completed {
			t.Fatalf("todo = %+v, want created incomplete todo", found)
		}
	})

	t.Run("create rejects duplicate id", func(t *testing.T) {
		resetTodos(t, ctx, pool)
		seedTodos(t, ctx, pool, todoSeed{id: "1", title: "first"})

		err := repo.Create(ctx, mustNewTodo(t, "1", "duplicate"))
		if err == nil {
			t.Fatal("Create returned nil, want duplicate key error")
		}
	})

	t.Run("updates todo", func(t *testing.T) {
		resetTodos(t, ctx, pool)
		seedTodos(t, ctx, pool, todoSeed{id: "1", title: "first"})
		renamed := mustNewTodo(t, "1", "renamed")
		renamed.Complete()

		if err := repo.Update(ctx, renamed); err != nil {
			t.Fatalf("Update returned error: %v", err)
		}

		found, err := repo.FindByID(ctx, "1")
		if err != nil {
			t.Fatalf("FindByID returned error: %v", err)
		}
		if found.Title != "renamed" || !found.Completed {
			t.Fatalf("todo = %+v, want renamed completed todo", found)
		}
	})

	t.Run("update returns not found", func(t *testing.T) {
		resetTodos(t, ctx, pool)

		err := repo.Update(ctx, mustNewTodo(t, "missing", "missing todo"))
		if !errs.Is(err, domain.ErrTodoNotFound) {
			t.Fatalf("error = %v, want %v", err, domain.ErrTodoNotFound)
		}
	})

	t.Run("deletes todo", func(t *testing.T) {
		resetTodos(t, ctx, pool)
		seedTodos(t, ctx, pool, todoSeed{id: "1", title: "first"})

		if err := repo.Delete(ctx, "1"); err != nil {
			t.Fatalf("Delete returned error: %v", err)
		}

		_, err := repo.FindByID(ctx, "1")
		if !errs.Is(err, domain.ErrTodoNotFound) {
			t.Fatalf("error = %v, want %v", err, domain.ErrTodoNotFound)
		}
	})

	t.Run("delete returns not found", func(t *testing.T) {
		resetTodos(t, ctx, pool)

		err := repo.Delete(ctx, "missing")
		if !errs.Is(err, domain.ErrTodoNotFound) {
			t.Fatalf("error = %v, want %v", err, domain.ErrTodoNotFound)
		}
	})

	t.Run("lists todos ordered by id", func(t *testing.T) {
		resetTodos(t, ctx, pool)
		seedTodos(t, ctx, pool,
			todoSeed{id: "2", title: "second"},
			todoSeed{id: "note-3", title: "manual note"},
			todoSeed{id: "1", title: "first"},
		)

		todos, err := repo.List(ctx)
		if err != nil {
			t.Fatalf("List returned error: %v", err)
		}
		gotIDs := make([]string, 0, len(todos))
		for _, todo := range todos {
			gotIDs = append(gotIDs, todo.ID)
		}
		wantIDs := []string{"1", "2", "note-3"}
		if strings.Join(gotIDs, ",") != strings.Join(wantIDs, ",") {
			t.Fatalf("ids = %v, want %v", gotIDs, wantIDs)
		}
	})

	t.Run("max numeric id ignores non numeric ids", func(t *testing.T) {
		resetTodos(t, ctx, pool)
		seedTodos(t, ctx, pool,
			todoSeed{id: "2", title: "second"},
			todoSeed{id: "9223372036854775808", title: "larger than int64 max"},
			todoSeed{id: "note-3", title: "manual note"},
		)

		maxID, err := repo.MaxNumericID(ctx)
		if err != nil {
			t.Fatalf("MaxNumericID returned error: %v", err)
		}
		if maxID != 1<<63 {
			t.Fatalf("maxID = %d, want %d", maxID, uint64(1)<<63)
		}
	})

	t.Run("max numeric id returns zero for empty table", func(t *testing.T) {
		resetTodos(t, ctx, pool)

		maxID, err := repo.MaxNumericID(ctx)
		if err != nil {
			t.Fatalf("MaxNumericID returned error: %v", err)
		}
		if maxID != 0 {
			t.Fatalf("maxID = %d, want 0", maxID)
		}
	})
}

func TestTodoRepository_PostgresDDLBoundaries(t *testing.T) {
	ctx := context.Background()
	repo, pool := newPostgresRepositoryFixture(t, ctx)

	t.Run("rejects empty title", func(t *testing.T) {
		resetTodos(t, ctx, pool)

		_, err := pool.Exec(ctx, `INSERT INTO todos (id, title) VALUES ($1, $2)`, "empty", "")
		if err == nil {
			t.Fatal("insert returned nil, want check constraint error")
		}
	})

	t.Run("rejects blank title", func(t *testing.T) {
		resetTodos(t, ctx, pool)

		_, err := pool.Exec(ctx, `INSERT INTO todos (id, title) VALUES ($1, $2)`, "blank", "   ")
		if err == nil {
			t.Fatal("insert returned nil, want check constraint error")
		}
	})

	t.Run("rejects null title", func(t *testing.T) {
		resetTodos(t, ctx, pool)

		_, err := pool.Exec(ctx, `INSERT INTO todos (id, title) VALUES ($1, NULL)`, "null-title")
		if err == nil {
			t.Fatal("insert returned nil, want not null constraint error")
		}
	})

	t.Run("sets database defaults", func(t *testing.T) {
		resetTodos(t, ctx, pool)

		_, err := pool.Exec(ctx, `INSERT INTO todos (id, title) VALUES ($1, $2)`, "defaults", "uses defaults")
		if err != nil {
			t.Fatalf("insert returned error: %v", err)
		}

		var completed bool
		var createdAt time.Time
		var updatedAt time.Time
		err = pool.QueryRow(ctx, `
			SELECT completed, created_at, updated_at
			FROM todos
			WHERE id = $1
		`, "defaults").Scan(&completed, &createdAt, &updatedAt)
		if err != nil {
			t.Fatalf("query defaults returned error: %v", err)
		}
		if completed {
			t.Fatal("completed = true, want false")
		}
		if createdAt.IsZero() || updatedAt.IsZero() {
			t.Fatalf("created_at = %v, updated_at = %v; want non-zero timestamps", createdAt, updatedAt)
		}
	})

	t.Run("accepts domain title length boundary", func(t *testing.T) {
		resetTodos(t, ctx, pool)
		title := strings.Repeat("あ", 100)

		if err := repo.Create(ctx, mustNewTodo(t, "max-title", title)); err != nil {
			t.Fatalf("Create returned error: %v", err)
		}

		found, err := repo.FindByID(ctx, "max-title")
		if err != nil {
			t.Fatalf("FindByID returned error: %v", err)
		}
		if found.Title != title {
			t.Fatalf("title length = %d, want 100 runes", len([]rune(found.Title)))
		}
	})
}

func newPostgresRepositoryFixture(t *testing.T, ctx context.Context) (*TodoRepository, *pgxpool.Pool) {
	t.Helper()

	startCtx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()

	container, err := tcpostgres.Run(startCtx,
		postgresImage,
		tcpostgres.WithDatabase(testDBName),
		tcpostgres.WithUsername(testDBUser),
		tcpostgres.WithPassword(testDBPass),
		tcpostgres.BasicWaitStrategies(),
	)
	testcontainers.CleanupContainer(t, container)
	if err != nil {
		t.Fatalf("start postgres container: %v", err)
	}

	databaseURL, err := container.ConnectionString(startCtx, "sslmode=disable")
	if err != nil {
		t.Fatalf("build postgres connection string: %v", err)
	}

	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		t.Fatalf("open pgx pool: %v", err)
	}
	t.Cleanup(pool.Close)
	if err := pool.Ping(ctx); err != nil {
		t.Fatalf("ping postgres: %v", err)
	}

	applyMigrations(t, ctx, pool)

	repo, err := NewTodoRepository(ctx, databaseURL)
	if err != nil {
		t.Fatalf("new todo repository: %v", err)
	}
	t.Cleanup(repo.Close)

	return repo, pool
}

func applyMigrations(t *testing.T, ctx context.Context, pool *pgxpool.Pool) {
	t.Helper()

	_, testFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("locate test file")
	}
	migrationPattern := filepath.Join(filepath.Dir(testFile), "..", "..", "..", "migrations", "*.up.sql")
	paths, err := filepath.Glob(migrationPattern)
	if err != nil {
		t.Fatalf("glob migrations: %v", err)
	}
	if len(paths) == 0 {
		t.Fatalf("no migrations found: %s", migrationPattern)
	}
	sort.Strings(paths)

	for _, path := range paths {
		ddl, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read migration %s: %v", path, err)
		}
		if _, err := pool.Exec(ctx, string(ddl)); err != nil {
			t.Fatalf("apply migration %s: %v", path, err)
		}
	}
}

type todoSeed struct {
	id        string
	title     string
	completed bool
}

func resetTodos(t *testing.T, ctx context.Context, pool *pgxpool.Pool) {
	t.Helper()

	if _, err := pool.Exec(ctx, `TRUNCATE TABLE todos`); err != nil {
		t.Fatalf("truncate todos: %v", err)
	}
}

func seedTodos(t *testing.T, ctx context.Context, pool *pgxpool.Pool, todos ...todoSeed) {
	t.Helper()

	for _, todo := range todos {
		if _, err := pool.Exec(ctx, `
			INSERT INTO todos (id, title, completed)
			VALUES ($1, $2, $3)
		`, todo.id, todo.title, todo.completed); err != nil {
			t.Fatalf("seed todo %q: %v", todo.id, err)
		}
	}
}

func mustNewTodo(t *testing.T, id, title string) *domain.Todo {
	t.Helper()

	todo, err := domain.NewTodo(id, title)
	if err != nil {
		t.Fatalf("NewTodo returned error: %v", err)
	}
	return todo
}
