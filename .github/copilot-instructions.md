# Go イベントチケット予約システム - AI エージェント向けガイドライン

本プロジェクトは、楽観的ロック、分散ロック、冪等性などの高度なバックエンド技術を実装した、高並行性イベントチケット予約システムです。二重予約ゼロを保証するため、3 層防御を実装しています。

## 最初に必ず読むこと

**調査や実装を開始する前に、必ず以下のドキュメントを読んでください：**

- **`README.md`**: プロジェクトの背景と目的、アーキテクチャ設計、データベース設計、API 設計
- **`docs/PROJECT_PLAN.md`**: 実装フェーズと優先順位、コード例とベストプラクティス

## 技術スタック

- **言語**: Go（最新安定版）
- **フレームワーク**: Echo（Web フレームワーク）
- **データベース**: PostgreSQL（メイン）、Redis（キャッシュ・分散ロック）
- **ORM/クエリ**: `sqlx`を使用（重量級 ORM は使用しない）
- **インフラ**: ローカル開発は Docker Compose

## アーキテクチャ概要

**Clean Architecture**（依存方向: `api/` → `application/` → `domain/` ← `infrastructure/`）

```
cmd/api/main.go          # エントリーポイント、DI構築
internal/
  domain/{event,seat,reservation}/  # エンティティ、リポジトリIF、エラー（外部依存なし）
  application/           # Service層（トランザクション境界、ユースケース）
  infrastructure/postgres/  # sqlxによるDB実装
  infrastructure/redis/     # 分散ロック、キャッシュ
  api/handler/           # Echoハンドラー、DTO変換
db/migrations/           # SQLマイグレーションファイル（golang-migrate）
```

## 必須の実装パターン

### 1. 分散ロック（Redis SetNX）

座席予約の排他制御。`internal/infrastructure/redis/distributed_lock.go` を参照：

```go
// ロック取得（TTL付き、リトライ対応）
lock, err := s.lockManager.AcquireLockWithRetry(ctx, lockKey, 10*time.Second, 3, 100*time.Millisecond)
defer lock.Release(ctx)  // 必ず defer で解放

// Luaスクリプトで所有者確認と削除をアトミックに実行
```

### 2. 楽観的ロック（version + WHERE 句）

`seats`テーブル更新時は必ず status 条件を含める。`seat_repository.go`を参照：

```go
// 更新件数 != 対象件数 なら ErrSeatAlreadyReserved を返す
query := `UPDATE seats SET status = 'reserved', version = version + 1
          WHERE id = ANY($1) AND status = 'available'`
rows, _ := result.RowsAffected()
if int(rows) != len(seatIDs) { return seat.ErrSeatAlreadyReserved }
```

### 3. トランザクション境界

**Application 層のみ**でトランザクション管理。Repository は`*sqlx.Tx`を受け取る：

```go
tx, _ := s.db.BeginTxx(ctx, nil)
defer tx.Rollback()  // 必須

s.reservationRepo.Create(ctx, tx, res)
s.seatRepo.ReserveSeats(ctx, tx, seatIDs, res.ID)

tx.Commit()
```

### 4. 冪等性チェック

予約作成時に`idempotency_key`で重複防止：

```go
existing, err := s.reservationRepo.GetByIdempotencyKey(ctx, input.IdempotencyKey)
if err == nil { return existing, nil }  // 既存予約を返却
```

## ドメインエラー

各ドメインの`errors.go`で定義。ハンドラーで`errors.Is()`で判定：

- `seat.ErrSeatAlreadyReserved`, `seat.ErrOptimisticLockConflict`
- `reservation.ErrReservationNotFound`, `event.ErrEventNotOpen`

## 開発コマンド

```bash
docker compose up -d     # PostgreSQL:5433, Redis:6379
make migrate-up          # マイグレーション適用
make run                 # サーバー起動 (localhost:8080)
make test                # 全テスト実行（-race -cover付き）
make lint                # golangci-lint
```

## テストガイドライン

- すべてのテストで`testify`（assert、require、mock）を使用
- **単体テスト**: Domain/Application 層にフォーカス、リポジトリはモック
- **統合テスト**: 実 DB/Redis（Docker 経由）を使用
- **並行テスト**: `sync.WaitGroup` + `atomic`で競合検証（`reservation_service_test.go`参照）
- **E2E テスト**: `e2e/reservation_flow_test.go` - httptest + 実 DB/Redis
- **テストセットアップ**: 実 DB 接続失敗時は`t.Skipf()`でスキップ
- **テーブル駆動テスト**を推奨

## AI との開発方針：TDD を活用する

本プロジェクトでは **TDD（テスト駆動開発）** を基本とする。AI エージェントは以下のサイクルを守ること：

### TDD サイクル（Red → Green → Refactor）

1. **Red**: テストを 1 つ書いて実行し、失敗することを確認する（実装前なので失敗する）
2. **Green**: テストを通す最小限のコードを実装する
3. **Refactor**: テストが通る状態を維持しながらコードを改善する
4. 1 に戻り、次のテストケースへ

### 使い分けの指針

| 状況                     | アプローチ          | 例                                            |
| ------------------------ | ------------------- | --------------------------------------------- |
| **複雑・不確実性が高い** | TDD（1 テストずつ） | 楽観的ロック、分散ロック、予約フロー          |
| **仕様が明確・シンプル** | テスト一括作成      | バリデーション、日付フォーマット、単純な CRUD |

### 重要なルール

- **テストを書いてから実装する**。実装を先に書かない
- **小さく始める**。最初から完璧な設計を目指さず、テストを通しながら設計を育てる
- **AI の暴走を防ぐガードレール**として、テストを活用する。テストが通らない変更は許可しない
- 複雑なロジック（並行処理制御、トランザクション境界）は必ず TDD で進める

## 開発ワークフロー

- **Makefile**: 一般的なタスクには`make`コマンドを使用（`make run`、`make test`、`make migrate-up`）
- **API 仕様**: ハンドラーに OpenAPI/Swagger アノテーションを記述して API を定義
- **マイグレーション**: スキーマ変更時は必ず新しいマイグレーションファイルを作成（`make migrate-create`）

## コード規約

- I/O 関数は第一引数に`context.Context`
- SQL プレースホルダーは`$1, $2`（PostgreSQL 形式）。動的な SQL 文字列の構築は避ける
- 配列パラメータは`pq.Array()`を使用
- ログは`zap`で構造化（`logger.With(zap.String(...))`）
- エラーはコンテキストを付与してラップ: `fmt.Errorf("予約作成に失敗: %w", err)`
- 設定は環境変数から読み込み（`internal/config`を使用）
