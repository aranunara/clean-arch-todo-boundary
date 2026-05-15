# Clean Architecture TODO Boundary

Clean Architecture の境界を TODO アプリで確認するための小さな Go サンプルです。

Domain / Usecase / Infrastructure / Handler を分け、Usecase が保存先の詳細を知らない構成にしています。保存先は in-memory と PostgreSQL の 2 種類を用意しています。

設計の考え方は [ARCHITECTURE.md](./ARCHITECTURE.md) にまとめています。

## Features

- Go 標準の `net/http` による TODO API
- Usecase の Repository interface と Infrastructure 実装の分離
- in-memory Repository
- PostgreSQL Repository
- Docker Compose による PostgreSQL / migration / API 起動
- golang-migrate 互換の SQL migration
- 起動中の API ポートを自動検出する `todoctl` CLI
- `internal/di` による依存関係の組み立て

## Requirements

- Go 1.26+
- Docker
- Docker Compose

## Quick Start

API、PostgreSQL、migration をまとめて起動します。

```bash
make up-all
```

`make up-all` はバックグラウンドで起動し、API と DB に割り当てられたホスト側ポートを表示します。ポートは未指定なら空いているものが自動で使われます。

すぐに動作確認する場合:

```bash
make try
```

停止する場合:

```bash
make down-all
```

データ volume も削除する場合:

```bash
make down-v
```

## CLI

`todoctl` は Docker Compose で起動した API のポートを自動で検出します。

```bash
go run ./cmd/todoctl url
go run ./cmd/todoctl list
go run ./cmd/todoctl create "層の違いをメモする"
go run ./cmd/todoctl rename 1 "Domain と Usecase の違いをメモする"
go run ./cmd/todoctl complete 1
```

API URL を明示したい場合:

```bash
TODO_API_URL=http://127.0.0.1:18080 go run ./cmd/todoctl list
```

Makefile 経由で CLI を呼ぶこともできます。

```bash
make cli args='list'
make cli args='create "README を整える"'
```

## HTTP API

API のポートを確認します。

```bash
make ports
```

例として `http://127.0.0.1:18080` に API が出ている場合:

```bash
curl -s -X POST http://127.0.0.1:18080/todos \
  -H 'content-type: application/json' \
  -d '{"title":"層の違いをメモする"}'

curl -s http://127.0.0.1:18080/todos

curl -s -X PATCH http://127.0.0.1:18080/todos/1 \
  -H 'content-type: application/json' \
  -d '{"title":"Domain と Usecase の違いをメモする"}'

curl -s -X POST http://127.0.0.1:18080/todos/1/complete
```

## Local Run

Docker を使わずに起動すると、保存先は in-memory になります。

```bash
go run ./cmd/server
```

この場合は `http://127.0.0.1:8080` で API が起動します。

## Migration

migration ファイルは [migrations/](./migrations) にあります。

```bash
make migrate-up
make migrate-down
```

新しい migration を作る場合:

```bash
make migrate-create name=add_due_date_to_todos
```

## Make Targets

```text
make up-all        API / PostgreSQL / migration を起動
make down-all      API / PostgreSQL を停止
make down-v        停止して volume も削除
make ports         API / DB の割り当てポートを表示
make logs          Compose サービスのログを表示
make try           CLI で TODO 作成と一覧取得を試す
make cli args=...  todoctl を実行
make test          go test ./...
make lint          golangci-lint run
make ci            test と lint を実行
make run           in-memory 保存で API をローカル起動
```
