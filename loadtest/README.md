# 負荷テスト

[k6](https://k6.io/) を使用した負荷テストシナリオです。

## セットアップ

### k6 のインストール

```bash
# macOS
brew install k6

# Ubuntu/Debian
sudo apt-get install k6

# Docker
docker pull grafana/k6
```

## テストシナリオ

### 1. スモークテスト (`smoke.js`)

基本的な動作確認用。サーバーが正常に応答することを確認します。

```bash
k6 run loadtest/smoke.js
```

**チェック項目:**
- ヘルスチェック `/api/v1/health`
- メトリクス `/metrics`
- イベント一覧 `/api/v1/events`

### 2. 予約シナリオ (`reservation.js`)

本格的な負荷テスト。2つのシナリオを含みます。

```bash
k6 run loadtest/reservation.js
```

**シナリオ構成:**

| シナリオ | 内容 | 目的 |
|---------|------|------|
| `normal_flow` | 通常のユーザーフロー | 段階的負荷でのパフォーマンス測定 |
| `concurrent_reservation` | 50人が同じ座席を同時予約 | 分散ロックの動作確認 |

**タイムライン:**
```
0s ─────────────── 2m ─────────────── 3m
       normal_flow (10 VU)
                              concurrent_reservation (50 VU)
```

### 3. ストレステスト (`stress-simple.js`)

200 並行ユーザーまでの高負荷テスト。混合ワークロード（読み取り80%、書き込み20%）。

```bash
k6 run loadtest/stress-simple.js
```

**負荷パターン:**
```
0s ──── 10s ──── 30s ──── 60s ──── 90s ──── 120s
   50VU     100VU     100VU    200VU     200VU → 0
```

### 4. 競合テスト (`concurrent-100.js`)

100人が同時に同じ座席を予約するテスト。

```bash
CONCURRENT_USERS=100 k6 run loadtest/concurrent-100.js
```

## カスタムメトリクス

| メトリクス | 説明 |
|-----------|------|
| `reservation_success` | 予約成功数 |
| `reservation_conflict` | 競合による失敗数 |
| `reservation_error` | その他エラー数 |
| `reservation_duration_ms` | 予約処理時間 |

## 閾値

```javascript
thresholds: {
  http_req_duration: ['p(95)<500'],  // 95%が500ms以内
  http_req_failed: ['rate<0.1'],     // エラー率10%未満
  reservation_success: ['count>0'],  // 最低1件成功
}
```

## 実行例

### ローカル環境

```bash
# サーバー起動
docker compose up -d
PORT=8081 go run ./cmd/api &

# スモークテスト
k6 run loadtest/smoke.js

# 本番シナリオ
k6 run loadtest/reservation.js
```

### 別ホストへのテスト

```bash
k6 run -e BASE_URL=https://api.example.com loadtest/reservation.js
```

## 期待される結果

### ストレステスト（200 VU）

```
█ THRESHOLDS 
  http_req_duration ✓ 'p(95)<1000' p(95)=40.23ms
  http_req_duration ✓ 'p(99)<2000' p(99)=148.2ms
  http_req_failed   ✓ 'rate<0.1' rate=0.00%

█ TOTAL RESULTS 
  http_reqs: 171380 (1426.9 req/sec)
  vus_max: 200
```

| 指標 | 結果 |
|------|------|
| スループット | 1,426 req/sec |
| p95 | 40.23 ms |
| p99 | 148.2 ms |
| エラー率 | 0.00% |

### 同時予約テスト

50人が同時に同じ座席を予約した場合:

```
✅ reservation_success: 1  (1人のみ成功)
⚠️ reservation_conflict: 49 (49人は競合で失敗)
❌ reservation_error: 0    (エラーなし)
```

これは**正常な動作**です。分散ロックと楽観的ロックにより、二重予約を防止しています。

## トラブルシューティング

### 接続エラー

```
ERRO[0000] request failed error="Get http://localhost:8081/...": dial tcp: connection refused"
```

→ サーバーが起動していることを確認してください。

### 全リクエストがエラー

→ Docker Compose で PostgreSQL と Redis が起動しているか確認:
```bash
docker compose ps
```
