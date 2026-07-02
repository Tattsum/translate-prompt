# デプロイ実装チェックリスト

[deployment.md](./deployment.md) の実装進捗管理用。別セッションで `[x]` に更新する。

## 前提

- 合意済み: 招待制 β / Fly.io / Pages / Workers / Tunnel / Access
- ドメイン: `prompt.tattsum.com`, `api.prompt.tattsum.com`
- DNS: `tattsum.com` は Cloudflare 管理済み

---

## WP-1: バックエンド本番対応

- [ ] `LISTEN_HOST` 環境変数（default `127.0.0.1`）
- [ ] `ALLOWED_ORIGINS` 環境変数（CORS 制限）
- [ ] `INVESTIGATE_ENABLED` 環境変数 + GraphQL/Connect ガード
- [ ] `Makefile` に `build-server-api`（embed なし）追加
- [ ] `Dockerfile` 作成
- [ ] `fly.toml` 作成
- [ ] Investigate 無効化のテスト追加
- [ ] `make serve` / `make test` がローカルで通る

## WP-2: フロントエンド本番対応

- [ ] `VITE_API_BASE_URL` 対応（`client.ts`）
- [ ] `VITE_ENABLE_WORKSPACE_PATH` で Workspace Path 非表示
- [ ] `frontend/.env.production` または Pages 環境変数ドキュメント化
- [ ] Vitest 更新
- [ ] `make dev` がローカルで通る

## WP-3: Workers プロキシ

- [ ] `wrangler.toml` 作成
- [ ] `workers/src/index.ts` 実装
- [ ] `/query` 転送
- [ ] `/translate_prompt.v1.TranslatePromptService/*` 転送
- [ ] CORS 整合確認
- [ ] `wrangler dev` でローカル検証

## WP-4: CI/CD

- [ ] `.github/workflows/deploy.yml` 作成
- [ ] test ジョブ（`make test`, `make lint`）
- [ ] Fly デプロイジョブ
- [ ] Workers デプロイジョブ
- [ ] Pages デプロイ（Git 連携 or wrangler pages）

## WP-5: インフラ手動セットアップ

- [ ] Fly アプリ作成・初回デプロイ
- [ ] Cloudflare Tunnel 作成・Fly に cloudflared 配置
- [ ] Workers `ORIGIN_URL` シークレット設定
- [ ] Pages プロジェクト作成・Git 連携
- [ ] DNS: `prompt` CNAME → Pages
- [ ] DNS: `api.prompt` → Workers
- [ ] Access: `prompt.tattsum.com` アプリ + ポリシー
- [ ] Access: `api.prompt.tattsum.com` アプリ + ポリシー
- [ ] GitHub Secrets 設定（`FLY_API_TOKEN`, `CLOUDFLARE_*`）

## WP-6: スモークテスト

- [ ] Access ログイン → SPA 表示
- [ ] Health / Estimate / Analyze / Optimize 成功
- [ ] Investigate Web 無効確認
- [ ] CLI Investigate ローカル動作確認
- [ ] Playground 本番無効確認

## WP-7: ドキュメント更新

- [ ] `docs/architecture.md` に本番構成追記
- [ ] `README.md` にデプロイ概要リンク
- [ ] 本チェックリストを完了状態に更新

---

## サブエージェント起動例

各 WP は独立して依頼可能。実装セッションの最初に以下を添付する:

```
@docs/deployment.md と docs/deployment-implementation-checklist.md を読み、
WP-N を実装してください。合意内容は deployment.md の「合意サマリ」が正です。
ローカル開発（make serve / make dev）は壊さないこと。
```

| WP | 推奨 subagent_type | 並列可否 |
|----|-------------------|---------|
| WP-1 | generalPurpose | WP-2 と並列可 |
| WP-2 | generalPurpose | WP-1 と並列可 |
| WP-3 | generalPurpose | WP-1 完了後 |
| WP-4 | generalPurpose | WP-1〜3 のファイル構成確定後 |
| WP-5 | shell + 手動 | WP-1〜4 完了後 |
| WP-6 | generalPurpose | WP-5 完了後 |
| WP-7 | generalPurpose | 最後 |
