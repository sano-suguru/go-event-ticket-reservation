# Context7 MCP 使用ガイド

## 概要
Context7 MCPは、ライブラリの最新ドキュメントとコード例をLLMのプロンプトに直接取得するためのMCPサーバー。

## 利用可能なツール

### 1. `resolve-library-id`
ライブラリ名をContext7互換のIDに解決する。

**パラメータ:**
- `libraryName` (必須): 検索するライブラリ名（例: "React", "Echo", "Redis"）

**戻り値:** マッチするライブラリのリスト（Context7互換ID付き）

### 2. `get-library-docs`
Context7 IDを使用してドキュメントを取得する。

**パラメータ:**
- `context7CompatibleLibraryID` (必須): Context7互換ライブラリID（例: `/vercel/next.js`, `/labstack/echo`）
- `topic` (オプション): 特定のトピックに絞る（例: "routing", "middleware"）
- `tokens` (オプション, デフォルト5000): 取得する最大トークン数（1000未満は1000に自動調整）

## ワークフロー

```
1. resolve-library-id でライブラリ名からIDを解決
   例: libraryName="echo" → /labstack/echo

2. get-library-docs でドキュメントを取得
   例: context7CompatibleLibraryID="/labstack/echo", topic="middleware"
```

## 使用例

### ユーザーが「use context7」を含むプロンプトを送信した場合

```
「Echo v4のミドルウェアの使い方を教えて use context7」
```

1. `resolve-library-id(libraryName="echo")` を呼び出し
2. 結果から適切なIDを選択（例: `/labstack/echo`）
3. `get-library-docs(context7CompatibleLibraryID="/labstack/echo", topic="middleware")` を呼び出し
4. 取得したドキュメントを元に回答

### ユーザーが直接IDを指定した場合

```
「/vercel/next.js のルーティングについて教えて」
```

この場合、`resolve-library-id` をスキップして直接 `get-library-docs` を呼び出せる。

## このプロジェクトで使う可能性のあるライブラリ

| ライブラリ | 用途 |
|-----------|------|
| Echo | Webフレームワーク |
| sqlx | データベースアクセス |
| go-redis | Redisクライアント |
| zap | ロギング |
| testify | テスト |
| prometheus | メトリクス |
| golang-migrate | マイグレーション |

## 注意事項
- `use context7` をプロンプトに含めるか、自動呼び出しルールを設定する
- トークン数を増やすとより多くのコンテキストが得られるが、コスト増加に注意
- 最新のドキュメントを取得するため、古い情報に依存しなくて済む
