.PHONY: all build run test lint clean docker-up docker-down migrate-up migrate-down migrate-create monitoring-up monitoring-down monitoring-logs help

# デフォルトタスク
all: lint test build

# ビルド
build:
	@echo "==> アプリケーションをビルドしています..."
	go build -o bin/api ./cmd/api

# 実行
run:
	@echo "==> アプリケーションを起動しています..."
	go run ./cmd/api

# テスト
test:
	@echo "==> テストを実行しています..."
	go test -v -race -cover ./...

# テスト（カバレッジ付き）
test-coverage:
	@echo "==> カバレッジ付きテストを実行しています..."
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# 統合テスト（Docker環境が必要）
test-integration:
	@echo "==> 統合テストを実行しています（Docker環境が必要）..."
	go test -v -race -tags=integration -cover ./...

# Lint
lint:
	@echo "==> Lintを実行しています..."
	@which golangci-lint > /dev/null || (echo "golangci-lint がインストールされていません。\n  brew install golangci-lint  または  go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest" && exit 1)
	golangci-lint run ./...

# 依存関係の整理
tidy:
	@echo "==> 依存関係を整理しています..."
	go mod tidy

# クリーン
clean:
	@echo "==> ビルド成果物を削除しています..."
	rm -rf bin/
	rm -f coverage.out coverage.html

# Docker関連
docker-up:
	@echo "==> Dockerコンテナを起動しています..."
	docker-compose up -d

docker-down:
	@echo "==> Dockerコンテナを停止しています..."
	docker-compose down

docker-logs:
	@echo "==> Dockerログを表示しています..."
	docker-compose logs -f

# マイグレーション関連
MIGRATE_CMD=migrate -path db/migrations -database "postgres://postgres:postgres@localhost:5433/ticket_reservation?sslmode=disable"

migrate-up:
	@echo "==> マイグレーションを適用しています..."
	$(MIGRATE_CMD) up

migrate-down:
	@echo "==> マイグレーションをロールバックしています..."
	$(MIGRATE_CMD) down 1

migrate-create:
	@echo "==> マイグレーションファイルを作成しています..."
	@read -p "マイグレーション名を入力: " name; \
	migrate create -ext sql -dir db/migrations -seq $$name

migrate-status:
	@echo "==> マイグレーション状態を確認しています..."
	$(MIGRATE_CMD) version

# ツールのインストール
install-tools:
	@echo "==> 開発ツールをインストールしています..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# 監視スタック（Prometheus + Grafana）
monitoring-up:
	@echo "==> 監視スタックを起動しています..."
	docker compose -f docker-compose.yml -f docker-compose.monitoring.yml up -d

monitoring-down:
	@echo "==> 監視スタックを停止しています..."
	docker compose -f docker-compose.yml -f docker-compose.monitoring.yml down

monitoring-logs:
	@echo "==> 監視スタックのログを表示しています..."
	docker compose -f docker-compose.yml -f docker-compose.monitoring.yml logs -f prometheus grafana

# ヘルプ
help:
	@echo "利用可能なコマンド:"
	@echo "  make build            - アプリケーションをビルド"
	@echo "  make run              - アプリケーションを起動"
	@echo "  make test             - ユニットテストを実行"
	@echo "  make test-coverage    - カバレッジ付きテストを実行"
	@echo "  make test-integration - 統合テストを実行（Docker環境が必要）"
	@echo "  make lint             - Lintを実行"
	@echo "  make tidy             - 依存関係を整理"
	@echo "  make clean            - ビルド成果物を削除"
	@echo "  make docker-up        - Dockerコンテナを起動"
	@echo "  make docker-down      - Dockerコンテナを停止"
	@echo "  make docker-logs      - Dockerログを表示"
	@echo "  make monitoring-up    - 監視スタック（Prometheus+Grafana）を起動"
	@echo "  make monitoring-down  - 監視スタックを停止"
	@echo "  make monitoring-logs  - 監視スタックのログを表示"
	@echo "  make migrate-up       - マイグレーションを適用"
	@echo "  make migrate-down     - マイグレーションをロールバック"
	@echo "  make migrate-create   - マイグレーションファイルを作成"
	@echo "  make install-tools    - 開発ツールをインストール"
