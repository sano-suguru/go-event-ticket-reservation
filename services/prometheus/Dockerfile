FROM prom/prometheus:v2.47.0

# 設定ファイルをコピー
COPY prometheus.yml /etc/prometheus/prometheus.yml

# envsubst をインストールして環境変数を置換
RUN apk add --no-cache gettext

# 環境変数を使用して設定を動的に置換
CMD ["sh", "-c", "envsubst < /etc/prometheus/prometheus.yml > /tmp/prometheus.yml && prometheus --config.file=/tmp/prometheus.yml --storage.tsdb.path=/prometheus --web.console.libraries=/usr/share/prometheus/console_libraries --web.console.templates=/usr/share/prometheus/consoles"]
