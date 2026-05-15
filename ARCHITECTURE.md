# Architecture

このリポジトリは、TODO アプリを題材に Domain / Usecase / Infrastructure / Handler の境界を確認するためのサンプルです。

## 依存の向き

```text
Handler
  -> Usecase
      -> Usecase の Repository interface
          <- Infrastructure の Repository 実装
```

Usecase は保存先が PostgreSQL なのか、メモリなのかを知りません。
Infrastructure が Usecase の interface を満たすことで、外側の詳細を内側へ差し込みます。

`internal/di` は本番起動時の依存解決だけを担当します。Handler / Usecase / Infrastructure の間に直接の具象依存を増やさず、どの実装を使うかはここで組み立てます。

## 層ごとの責務

| 層 | 責務 | ファイル |
|---|---|---|
| Domain | TODO という業務概念のルール・状態・判断 | `internal/domain/todo.go` |
| Usecase | TODO 作成・変更・完了など、アプリケーション操作の段取り | `internal/usecase/todo_usecase.go` |
| Infrastructure | Repository interface をメモリ保存 / PostgreSQL 保存で具体化 | `internal/infra/memory/todo_repository.go`, `internal/infra/postgres/todo_repository.go` |
| Handler | HTTP リクエスト/レスポンスへの変換 | `internal/handler/httpapi/todo_handler.go` |
| DI | 本番起動時の依存関係の解決 | `internal/di/container.go` |

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

Repository は独立した 1 層というより、Usecase と Infrastructure の境界として捉えています。

```text
Usecase -> TodoRepository interface <- Infra の Repository 実装
```
