# Railway 本番環境用監視スタック

Railway に Prometheus + Grafana をデプロイして、本番環境のメトリクスを可視化します。

## アーキテクチャ

```
Railway Project
├── ticket-api (Go アプリ)
│   └── /metrics エンドポイント（Basic認証付き）
├── prometheus
│   └── ticket-api をスクレイプ
└── grafana
    └── Prometheus からデータ取得
```

## デプロイ手順

### 1. Prometheus をデプロイ

```bash
# Railway プロジェクトにサービスを追加
railway service create prometheus

# monitoring/ ディレクトリからビルド（コンテキストが重要）
cd monitoring
railway up --service prometheus --dockerfile railway/Dockerfile.prometheus
```

環境変数を設定:

```bash
railway variables --service prometheus \
  --set "METRICS_USER=grafana" \
  --set "METRICS_PASSWORD=<your-password>" \
  --set "TICKET_API_HOST=go-event-ticket-reservation-production.up.railway.app"
```

### 2. Grafana をデプロイ

```bash
railway service create grafana

# monitoring/ ディレクトリからビルド
cd monitoring
railway up --service grafana --dockerfile railway/Dockerfile.grafana
```

環境変数を設定:

```bash
railway variables --service grafana \
  --set "GF_SECURITY_ADMIN_PASSWORD=<secure-password>" \
  --set "PROMETHEUS_URL=http://prometheus.railway.internal:9090"
```

### 3. Grafana にアクセス

Railway のダッシュボードから Grafana サービスの URL を確認してアクセス:

- **User**: `admin`
- **Password**: 設定した `GF_SECURITY_ADMIN_PASSWORD`

## 注意事項

- Railway の無料プランでは複数サービスに制限があります
- Prometheus のストレージは Railway の永続化ストレージを使用してください
- 本番環境では `GF_SECURITY_ADMIN_PASSWORD` を強力なパスワードに変更してください

## 代替案: シンプルな構成

コストを抑えたい場合は、ローカルの監視スタック（`make monitoring-up`）を使い、必要に応じて本番メトリクスをローカルで確認することもできます。
