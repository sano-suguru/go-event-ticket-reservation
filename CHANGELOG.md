# Changelog

## [0.2.0](https://github.com/sano-suguru/go-event-ticket-reservation/compare/v0.1.0...v0.2.0) (2025-12-26)


### Features

* 10万座席ベンチマークテストとバルクINSERT最適化 ([e8adb80](https://github.com/sano-suguru/go-event-ticket-reservation/commit/e8adb80fe093b403e2e6ffce8b9b89931c6b45e1))
* 200VUストレステスト追加とドキュメント更新 ([51b77c2](https://github.com/sano-suguru/go-event-ticket-reservation/commit/51b77c241919aaebd1ac3dc18c621d44d34c24f9))
* Prometheus + Grafana 監視スタックを追加 ([2f77faa](https://github.com/sano-suguru/go-event-ticket-reservation/commit/2f77faac5f8685e04032cf536b6a5e12ec81f6be))
* Railway デプロイ用の Prometheus/Grafana サービスディレクトリを追加 ([345593b](https://github.com/sano-suguru/go-event-ticket-reservation/commit/345593b081702141f5d164d28b25e3123bec9cb0))
* Railwayデプロイ対応を追加 ([c3fc6d9](https://github.com/sano-suguru/go-event-ticket-reservation/commit/c3fc6d951a03e8a0327e3cefc120dfbda1428618))
* Redis分散ロックによる高トラフィック対策を実装 ([c7e5f18](https://github.com/sano-suguru/go-event-ticket-reservation/commit/c7e5f18981d6947ce57dff03fb8e3e68b25518fa))
* Redis座席キャッシュとCI/CDパイプラインを追加 ([87f7a20](https://github.com/sano-suguru/go-event-ticket-reservation/commit/87f7a20fd1e7c6fefcc5a6ff2d9a8521b5ab8325))
* イベントCRUD機能を実装 ([4da0cc5](https://github.com/sano-suguru/go-event-ticket-reservation/commit/4da0cc57b14ddee88d3421dc7185ba926782fb24))
* プロジェクト基盤セットアップ ([41471c5](https://github.com/sano-suguru/go-event-ticket-reservation/commit/41471c5c4c08274ba18195b045923d4b82f11f19))
* メトリクス・OpenAPI・負荷テストを追加 (Phase 3完了) ([d499696](https://github.com/sano-suguru/go-event-ticket-reservation/commit/d499696bf419ee13760e1e24257f3bfdf16ed9bf))
* 座席・予約機能を実装 ([ae97e36](https://github.com/sano-suguru/go-event-ticket-reservation/commit/ae97e36362ecddb8cda1802f906e46d8fb5e855e))
* 構造化ロギングと期限切れ予約クリーンアップを実装 ([d3783eb](https://github.com/sano-suguru/go-event-ticket-reservation/commit/d3783ebb994372cd0a6aea7b83f2f249e9e26a11))
* 水平スケーリング構成と100人同時競合テストを追加 ([565c006](https://github.com/sano-suguru/go-event-ticket-reservation/commit/565c006148fb2e1aff9e5f450216d17fbdbf792c))
* 起動時にマイグレーションを自動実行 ([9b202e5](https://github.com/sano-suguru/go-event-ticket-reservation/commit/9b202e53c470081bfc8f89f34e7ea54cfdad09d8))


### Bug Fixes

* .gitignore の api パターンを修正し不足ファイルを追加 ([85b8037](https://github.com/sano-suguru/go-event-ticket-reservation/commit/85b803752a49eae74f5e154115c59680615152d8))
* CIテストの並列実行を無効化してテスト間干渉を防止 ([cc87a47](https://github.com/sano-suguru/go-event-ticket-reservation/commit/cc87a4747e8a939fdd15b83bb867765b4a0363aa))
* CI互換性のためGoバージョンを1.23に変更 ([307b43e](https://github.com/sano-suguru/go-event-ticket-reservation/commit/307b43eb7a9b8b6327493a7e26f528d1f62d325a))
* E2Eテストにバリデータを追加 ([a985230](https://github.com/sano-suguru/go-event-ticket-reservation/commit/a985230a83e2724b12e843170fc789da3a815c4e))
* Go バージョンを 1.24 に統一 ([36da40f](https://github.com/sano-suguru/go-event-ticket-reservation/commit/36da40fdf46a072faa848b82b61ab3638afc3e4f))
* golang-migrate を v4.19.1 に更新 ([b3ed962](https://github.com/sano-suguru/go-event-ticket-reservation/commit/b3ed962c68c0d089bb9b702a523b435e36ee00b8))
* golang-migrate を v4.19.1 に更新 ([849ac23](https://github.com/sano-suguru/go-event-ticket-reservation/commit/849ac238395e7333cfd17b234787cce7c63eac97))
* golangci-lint の非推奨設定を修正 ([256536d](https://github.com/sano-suguru/go-event-ticket-reservation/commit/256536d2f52f5ecce1119a5c8d1e1ce2b5ed8033))
* golangci-lint-action を v9 にアップグレード ([7c5be3f](https://github.com/sano-suguru/go-event-ticket-reservation/commit/7c5be3f0b4f58a4d78823edcf9ead5e9cafaa436))
* golangci-lint-action を v9 にアップグレード ([3a3b378](https://github.com/sano-suguru/go-event-ticket-reservation/commit/3a3b3782f4cdaba9f6872d43fb77abcbd9d7ff73))
* Prometheus Dockerfile を busybox 互換に修正（マルチステージビルド） ([39c7b83](https://github.com/sano-suguru/go-event-ticket-reservation/commit/39c7b830b634d579874c1e02f13500f2cc426dde))
* Swagger のホスト設定を削除 ([88b2ac3](https://github.com/sano-suguru/go-event-ticket-reservation/commit/88b2ac3a32841f2fc8879b19dbc8def6b1b2caf9))
* Swagger ヘルスチェックのルーターパスを修正 ([ba86fd9](https://github.com/sano-suguru/go-event-ticket-reservation/commit/ba86fd967afd6d272ae47f24e681163a89152998))
* マイグレーションエラー変数のシャドウイングを修正 ([70a1011](https://github.com/sano-suguru/go-event-ticket-reservation/commit/70a1011b08040733120fe175cbc0a853fe78eb9a))
* ルートレベルにヘルスチェックを追加 ([1797338](https://github.com/sano-suguru/go-event-ticket-reservation/commit/1797338518dc433f85bb688ed56819f4c0be9656))
* 座席一括作成時にDBが生成したIDを返すように修正 ([6bdee85](https://github.com/sano-suguru/go-event-ticket-reservation/commit/6bdee85a1f2c9ebe1454faaa54ed2f1588401c34))


### Code Refactoring

* /metrics エンドポイントの認証を削除し意図的に公開 ([5a7fb7c](https://github.com/sano-suguru/go-event-ticket-reservation/commit/5a7fb7cb074a754f42086322de823fa45b5268ed))
* DIP違反を修正しテストカバレッジを60%+に向上 ([c01bc2d](https://github.com/sano-suguru/go-event-ticket-reservation/commit/c01bc2d2c2ac1b672f8bfde9545ee9acd4892371))
* DIP違反を修正しテストカバレッジを60%+に向上 ([5a06f90](https://github.com/sano-suguru/go-event-ticket-reservation/commit/5a06f90b00457b897561c5d3f2d7fb79881c2d03))
* Echo ベストプラクティスに準拠 ([fd3042e](https://github.com/sano-suguru/go-event-ticket-reservation/commit/fd3042e60836f3122c5b92a85e459ae0ff9e68bc))
* Echo ベストプラクティスに準拠 ([ed75e55](https://github.com/sano-suguru/go-event-ticket-reservation/commit/ed75e553b1c8a4a590f416756178662aaf06d938))
