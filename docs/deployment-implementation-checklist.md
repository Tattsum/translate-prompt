# デプロイ実装チェックリスト

[deployment.md](./deployment.md) の実装進捗管理用。  
**現状の詳細は [deployment-session-handoff.md](./deployment-session-handoff.md) を参照。**

## 前提

- 合意済み: 招待制 β / Fly.io / Pages / Workers / Access
- 目標ドメイン: `translate.tattsum.com`（SPA）, `prompt-api.tattsum.com`（API）
- DNS: `tattsum.com` は Cloudflare 管理済み（アカウント: Kurohari35@gmail.com's Account）
- PR #1 MERGED（WP-1〜4 コード + CI）
- SPA 公開: Pages カスタムドメイン banned → **Workers プロキシ**（方式 A）

---

## WP-1: バックエンド本番対応

- [x] `LISTEN_HOST` 環境変数（default `127.0.0.1`）
- [x] `ALLOWED_ORIGINS` 環境変数（CORS 制限）
- [x] `INVESTIGATE_ENABLED` 環境変数 + GraphQL/Connect ガード
- [x] `Makefile` に `build-server-api`（embed なし）追加
- [x] `Dockerfile` 作成
- [x] `fly.toml` 作成
- [x] Investigate 無効化のテスト追加
- [x] `make serve` / `make test` がローカルで通る

## WP-2: フロントエンド本番対応

- [x] `VITE_API_BASE_URL` 対応（`client.ts`）
- [x] `VITE_ENABLE_WORKSPACE_PATH` で Workspace Path 非表示
- [x] `frontend/.env.production` 作成
- [x] Vitest 更新
- [x] `make dev` がローカルで通る

## WP-3: Workers プロキシ（API）

- [x] `wrangler.toml` 作成
- [x] `workers/src/index.ts` 実装
- [x] `/query` 転送
- [x] `/translate_prompt.v1.TranslatePromptService/*` 転送
- [x] CORS 整合確認（Go 側委譲）
- [ ] `wrangler dev` でローカル検証（任意）

## WP-3b: Workers プロキシ（SPA）

- [x] `wrangler.web.toml` 作成
- [x] `workers/src/web.ts` 実装（Pages リバースプロキシ）
- [x] CI `deploy-workers` に SPA Worker デプロイ追加

## WP-4: CI/CD

- [x] `.github/workflows/deploy.yml` 作成
- [x] test ジョブ（`make test`, `make lint`）
- [x] Fly デプロイジョブ（`ALLOWED_ORIGINS` secrets 同期付き）
- [x] Workers デプロイジョブ（API + SPA）
- [x] Pages デプロイ（wrangler pages）
- [x] GitHub Secrets 設定（`FLY_API_TOKEN`, `CLOUDFLARE_*`）

## WP-5: インフラ手動セットアップ

- [x] Fly アプリ作成・デプロイ（`translate-prompt-api`）
- [ ] ~~Cloudflare Tunnel 作成~~（見送り。Fly パブリック URL 直結）
- [x] Workers `ORIGIN_URL` シークレット設定（`https://translate-prompt-api.fly.dev`）
- [x] Pages プロジェクト作成（`translate-prompt`）
- [x] DNS: `prompt-api` → Workers（動作確認済み）
- [x] DNS: `translate` → A `192.0.2.1` + Worker Route（[手順](./deployment-dns-setup.md)）
- [x] Workers: SPA プロキシ `translate-prompt-web` デプロイ・動作確認済み
- [x] Access: SPA URL アプリ + ポリシー（[手順](./deployment-access-setup.md)）
- [x] Access: `prompt-api.tattsum.com` アプリ + ポリシー

### ドメイン問題（記録）

- [x] `prompt.tattsum.com` Pages カスタムドメイン → **banned**（Workers 方式に切替）
- [x] `translate.tattsum.com` Pages カスタムドメイン → **banned**（Workers 方式に切替）
- [x] `api.prompt.tattsum.com` → TLS 失敗。`prompt-api.tattsum.com` に変更済み
- [x] `translate.tattsum.com` CNAME のみ → **522**（A + Worker に切替済み、200 確認）

## WP-6: スモークテスト

- [x] 自動スクリプト `scripts/deployment-smoke-test.sh`
- [x] Health（`prompt-api.tattsum.com` 経由）成功
- [x] Pages（`translate-prompt-2el.pages.dev`）200
- [x] SPA Worker 経由 `translate.tattsum.com` 302（Access）
- [x] Access ログイン → SPA 表示
- [x] Estimate / Analyze / Optimize 成功（Access ログイン後ブラウザ確認済み）
- [x] Investigate Web 無効確認（`investigate disabled`）
- [ ] CLI Investigate ローカル動作確認
- [x] Playground 本番無効確認（404）
- [x] CORS（許可オリジンのみ）確認

## WP-7: ドキュメント更新

- [x] `docs/deployment-session-handoff.md` 作成（セッション引き継ぎ）
- [x] `docs/deployment.md` を実態に合わせて更新（Tunnel 見送り、Workers SPA、blog パターン）
- [x] `docs/deployment-access-setup.md` 作成
- [x] `docs/deployment-dns-setup.md` 作成
- [x] `docs/architecture.md` に本番構成追記
- [x] `README.md` にデプロイ概要リンク
- [x] 本チェックリストを進行状態に更新

---

## サブエージェント起動例

```
@docs/deployment-session-handoff.md を最優先で読み、
DNS / Access の手動設定とスモークテストを完了してください。
```

| WP | 状態 | 備考 |
|----|------|------|
| WP-1 | 完了 | |
| WP-2 | 完了 | |
| WP-3 | 完了（API） | |
| WP-3b | 完了（コード） | CI デプロイ + DNS 手動 |
| WP-4 | 完了 | |
| WP-5 | 完了 | Access 設定済み（2026-07-03） |
| WP-6 | 完了 | スモークテスト + ブラウザ E2E 確認済み |
| WP-7 | 完了 | |
