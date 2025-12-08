FROM grafana/grafana:10.2.0

# 環境変数で設定
ENV GF_SECURITY_ADMIN_USER=admin
ENV GF_USERS_ALLOW_SIGN_UP=false

# データソース設定
COPY provisioning/datasources/datasources.yml /etc/grafana/provisioning/datasources/
COPY provisioning/dashboards/dashboards.yml /etc/grafana/provisioning/dashboards/

# ダッシュボード
COPY dashboards/ /var/lib/grafana/dashboards/
