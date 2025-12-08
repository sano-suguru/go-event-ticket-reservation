# gettext (envsubst) をインストールするためにマルチステージビルド
FROM alpine:3.19 AS builder
RUN apk add --no-cache gettext

FROM prom/prometheus:v2.47.0

# envsubst バイナリをコピー
COPY --from=builder /usr/bin/envsubst /usr/bin/envsubst

# 設定ファイルをコピー
COPY prometheus.yml /etc/prometheus/prometheus.yml

# 環境変数を使用して設定を動的に置換
CMD ["sh", "-c", "envsubst < /etc/prometheus/prometheus.yml > /tmp/prometheus.yml && prometheus --config.file=/tmp/prometheus.yml --storage.tsdb.path=/prometheus --web.console.libraries=/usr/share/prometheus/console_libraries --web.console.templates=/usr/share/prometheus/consoles"]
