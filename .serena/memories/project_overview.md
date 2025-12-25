# プロジェクト概要

## 目的
高並行性イベントチケット予約システム。3層防御（分散ロック、楽観的ロック、冪等性キー）でダブルブッキングをゼロに保証。

## 技術スタック
- **言語**: Go 1.25.5
- **Webフレームワーク**: Echo v4
- **データベース**: PostgreSQL（sqlxを使用）
- **キャッシュ/ロック**: Redis（go-redis v9）
- **監視**: Prometheus + Grafana
- **マイグレーション**: golang-migrate
- **リンター**: golangci-lint
- **APIドキュメント**: swaggo/swag
- **ロギング**: uber/zap
- **テスト**: testify/mock, testify/assert

## アーキテクチャ（クリーンアーキテクチャ）
依存の流れ: `api/` → `application/` → `domain/` ← `infrastructure/`

```
cmd/api/main.go              # エントリーポイント（DI構成）
internal/
  domain/                    # 純粋なビジネスロジック（外部依存なし）
    event/                   # イベントエンティティ、リポジトリIF、エラー
    seat/                    # 座席状態: available → reserved → confirmed
    reservation/             # 15分有効期限、冪等性キー
    transaction/             # Tx と Manager インターフェース
  application/               # トランザクション境界を持つサービス層
  infrastructure/
    postgres/                # sqlxベースのリポジトリ実装
    redis/                   # 分散ロック + 座席キャッシュ（Luaスクリプト）
  api/handler/               # EchoハンドラとSwaggerアノテーション
  worker/                    # ExpiredReservationCleaner（自動キャンセル）
db/migrations/               # golang-migrate SQLファイル
```

## 3層防御の実装場所
1. **分散ロック (Redis)**: `internal/infrastructure/redis/distributed_lock.go`
2. **楽観的ロック (PostgreSQL)**: `internal/infrastructure/postgres/seat_repository.go`
3. **冪等性チェック**: `internal/application/reservation_service.go`

## デプロイ環境
- **本番**: Railway（`railway.toml` で設定）
- **CI/CD**: GitHub Actions（`.github/workflows/`）
- Railway は GitHub の main ブランチ push を検知して自動デプロイ

## 重要な設計原則
- ドメイン層は外部依存ゼロ
- トランザクション管理はアプリケーション層のみ
- リポジトリは `*sqlx.Tx` を受け取る
