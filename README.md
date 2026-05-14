# Clean Architecture TODO Boundary

TODO アプリを題材に、Domain / Usecase / Infrastructure / Handler の境界を記録するための小さな Go サンプルです。

## 依存の向き

```text
Handler
  -> Usecase
      -> Domain の Repository interface
          <- Infrastructure の Repository 実装
```

Usecase は保存先が PostgreSQL なのか、メモリなのかを知りません。
Infrastructure が Domain の interface を満たすことで、外側の詳細を内側へ差し込みます。

## 層ごとの責務

| 層 | 責務 | ファイル |
|---|---|---|
| Domain | TODO という業務概念のルール・状態・判断 | `internal/domain/todo.go` |
| Usecase | TODO 作成・変更・完了など、アプリケーション操作の段取り | `internal/usecase/todo_usecase.go` |
| Infrastructure | Repository interface をメモリ保存 / PostgreSQL 保存で具体化 | `internal/infra/memory/todo_repository.go`, `internal/infra/postgres/todo_repository.go` |
| Handler | HTTP リクエスト/レスポンスへの変換 | `internal/handler/httpapi/todo_handler.go` |

## 個人的な理解

```text
Domain = 業務概念を外部事情から切り離して扱う
Usecase = Domain をオーケストレートして操作を完了させる
Infra = Domain / Usecase が必要とする外部依存を具体化する
Handler = HTTP との入出力境界を担当する
```

自分の中では、Domain にあるのは「TODO として正しい状態」です。

- title は空にできない
- title は 100 文字以内
- 完了済み TODO は rename できない
- TODO を complete すると `Completed` が true になる

Usecase にあるのは「操作を成立させる手順」だと考えています。

- ID を発行する
- Domain の `NewTodo` / `Rename` / `Complete` を呼ぶ
- Repository interface 経由で保存する
- 出力 DTO に変換する

Repository は独立した 1 層というより、Domain と Infrastructure の境界として捉えています。

```text
Usecase -> Domain の TodoRepository interface <- Infra の memory.TodoRepository 実装
```

## 実行

Docker で PostgreSQL、golang-migrate、API をまとめて起動する場合:

```bash
make up-all
```

バックグラウンドで起動する場合:

```bash
make up-d
```

`make up-all` は PostgreSQL の healthcheck を待ち、`migrations/` の SQL を `migrate/migrate` で適用してから API をバックグラウンドで起動します。
起動後に、割り当てられた API / DB のポートも表示されます。

起動した API をすぐ試す場合:

```bash
make try
```

CLI から操作する場合:

```bash
go run ./cmd/todoctl url
go run ./cmd/todoctl list
go run ./cmd/todoctl create "層の違いをメモする"
go run ./cmd/todoctl rename 1 "Domain と Usecase の違いをメモする"
go run ./cmd/todoctl complete 1
```

`todoctl` は `TODO_API_URL` が指定されていなければ、`docker compose port api 8080` から API のポートを自動で見つけます。

API と PostgreSQL をまとめて停止する場合:

```bash
make down-all
```

ホスト側のポートは、未指定なら空いているポートが自動で割り当てられます。
割り当てられたポートを確認する場合:

```bash
make ports
```

ホスト側のポートを固定したい場合:

```bash
API_PORT=18080 POSTGRES_PORT=15432 make up-all
```

migration だけ実行したい場合:

```bash
make migrate-up
make migrate-down
```

Docker を使わず、メモリ保存で起動する場合:

```bash
go test ./...
go run ./cmd/server
```

別ターミナルから:

```bash
curl -s -X POST http://127.0.0.1:8080/todos \
  -H 'content-type: application/json' \
  -d '{"title":"層の違いをメモする"}'

curl -s http://127.0.0.1:8080/todos

curl -s -X PATCH http://127.0.0.1:8080/todos/1 \
  -H 'content-type: application/json' \
  -d '{"title":"Domain と Usecase の違いをメモする"}'

curl -s -X POST http://127.0.0.1:8080/todos/1/complete
```
