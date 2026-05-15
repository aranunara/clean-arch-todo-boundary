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

## New と Create の違い

`domain.NewTodo` は、Todo entity を不変条件を満たした状態で生成する Domain logic です。
ここで守っているのは「Todo として存在するなら常に満たすべき条件」です。

- id は空にできない
- title は空にできない
- title は 100 文字以内
- 初期状態では `Completed` は false

一方で、`usecase.CreateTodo` はアプリケーション操作の段取りです。

- ID を発行する
- `domain.NewTodo` で正しい Todo を生成する
- Repository interface 経由で保存する
- 出力 DTO に変換する

つまり、`NewTodo` は「Todo を正しい状態で作る」責務を持ち、`CreateTodo` は「TODO 作成という操作を成立させる」責務を持ちます。

## 条件をどこに置くか

条件を Domain に置くか Usecase に置くかは、その条件がどれくらい普遍的かで判断します。

```text
Domain:
  その entity がいつ・どの操作から扱われても常に守るべき条件

Usecase:
  その操作・フロー・権限・外部都合に限った成立条件
```

たとえば、この TODO では「title は空にできない」「完了済み TODO は rename できない」は Domain の責務です。
どのユースケースから扱っても、Todo として破ってはいけない条件だからです。

一方で、次のような条件は Usecase の責務として扱えます。

- 1 ユーザーが 1 日に作れる TODO は 10 件まで
- 特定の画面から作成するときだけ title の重複を禁止する
- 管理者操作の場合だけ通常とは違う変更を許可する
- 外部 API を呼び出して結果を保存する

Usecase 固有の条件を通過したあとでも、最終的に保存される entity は Domain の不変条件を満たしている必要があります。

```text
Usecase 固有条件を確認
  -> Domain の New / method で不変条件を守る
  -> Repository interface 経由で保存する
```

なので、Domain を通る理由は Usecase と Infrastructure の間に無理に層を挟むためではなく、保存する対象である entity の正しさを一箇所で守るためです。
