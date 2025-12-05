.PHONY: all build run test lint clean docker-up docker-down migrate-up migrate-down migrate-create help

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

# Lint
lint:
	@echo "==> Lintを実行しています..."
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
MIGRATE_CMD=migrate -path db/migrations -database "postgres://postgres:postgres@localhost:5432/ticket_reservation?sslmode=disable"

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

# ヘルプ
help:
	@echo "利用可能なコマンド:"
	@echo "  make build          - アプリケーションをビルド"
	@echo "  make run            - アプリケーションを起動"
	@echo "  make test           - テストを実行"
	@echo "  make test-coverage  - カバレッジ付きテストを実行"
	@echo "  make lint           - Lintを実行"
	@echo "  make tidy           - 依存関係を整理"
	@echo "  make clean          - ビルド成果物を削除"
	@echo "  make docker-up      - Dockerコンテナを起動"
	@echo "  make docker-down    - Dockerコンテナを停止"
	@echo "  make docker-logs    - Dockerログを表示"
	@echo "  make migrate-up     - マイグレーションを適用"
	@echo "  make migrate-down   - マイグレーションをロールバック"
	@echo "  make migrate-create - マイグレーションファイルを作成"
	@echo "  make install-tools  - 開発ツールをインストール"
