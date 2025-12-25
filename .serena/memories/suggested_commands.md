# 開発コマンド一覧

## 開発環境
```bash
make docker-up          # PostgreSQL (5433) + Redis (6379) 起動
make migrate-up         # マイグレーション適用
make run                # localhost:8080 でサーバー起動
```

## テスト
```bash
make test               # 全テスト（-race -cover付き）
make test-coverage      # カバレッジレポート生成（coverage.html）
make test-integration   # 統合テスト（Docker必要）
go test -v -run TestName ./path/to/package  # 単一テスト実行
```

## コード品質
```bash
make lint               # golangci-lint 実行
make tidy               # go mod tidy
make install-tools      # golangci-lint と migrate CLI インストール
```

## ビルド
```bash
make build              # bin/api にビルド
make clean              # ビルド成果物削除
```

## Docker
```bash
make docker-up          # コンテナ起動
make docker-down        # コンテナ停止
make docker-logs        # ログ表示
```

## マイグレーション
```bash
make migrate-up         # マイグレーション適用
make migrate-down       # 1つロールバック
make migrate-create     # 新規マイグレーション作成（名前入力）
make migrate-status     # 現在のバージョン確認
```

## 監視スタック
```bash
make monitoring-up      # Prometheus + Grafana 起動
make monitoring-down    # 監視スタック停止
make monitoring-logs    # 監視ログ表示
```

## システムコマンド（Darwin/macOS）
```bash
git status              # Git状態確認
git diff                # 差分確認
ls -la                  # ディレクトリ内容
find . -name "*.go"     # ファイル検索
grep -r "pattern" .     # パターン検索
```
