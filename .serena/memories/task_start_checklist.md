# タスク開始時チェックリスト

CI/CD やインフラ関連のタスクを始める前に必ず確認すること。

## 1. デプロイ環境の確認
以下のファイルを確認し、デプロイ先を把握する：
- `railway.toml` → Railway
- `fly.toml` → Fly.io
- `vercel.json` → Vercel
- `netlify.toml` → Netlify
- `render.yaml` → Render
- `app.yaml` → Google App Engine
- `Procfile` → Heroku
- `kubernetes/`, `k8s/` → Kubernetes
- `terraform/` → Terraform (AWS/GCP/Azure等)

## 2. 既存のCI/CDの確認
- `.github/workflows/*.yml`
- `.gitlab-ci.yml`
- `.circleci/config.yml`
- `Jenkinsfile`
- `azure-pipelines.yml`

## 3. 提案前の確認事項
- 既存のデプロイフローと矛盾しないか？
- 冗長な処理を追加していないか？
- ユーザーの環境・制約を理解しているか？

## 教訓
「一般的なベストプラクティス」を押し付けず、まずプロジェクト固有の構成を理解する。
