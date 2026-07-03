# Web 公開デプロイ — セッション引き継ぎ（2026-07-03）

別セッションで作業を再開するための **現状スナップショット**。  
設計の正は引き続き [deployment.md](./deployment.md) の「合意サマリ」だが、**実装・インフラの実態**は本ドキュメントを正とする。

---

## 1. プロジェクト目的（再掲）

**translate-prompt** を **招待制 β** として Web 公開する。

| 決定事項 | 結論 |
|---------|------|
| 公開スコープ | 招待制 β（Cloudflare Access） |
| Web 機能 | Analyze / Optimize / Estimate のみ |
| Investigate | Web 無効（`INVESTIGATE_ENABLED=false`）。CLI のみ維持 |
| オリジン | Fly.io（Go モノリス） |
| フロント | Cloudflare Pages（SPA ビルド・デプロイ） |
| エッジ API | Cloudflare Workers（薄いプロキシ） |
| 認証 | Cloudflare Access（ユーザ DB なし） |
| CI/CD | GitHub Actions（test → Fly / Workers / Pages） |
| ドメインゾーン | `tattsum.com`（Cloudflare DNS 管理済み） |

### 設計からの逸脱（意図的な簡略化）

| 項目 | 当初設計（deployment.md） | 実装済みの実態 |
|------|--------------------------|----------------|
| オリジン到達 | Cloudflare Tunnel（`cloudflared`） | **Fly のパブリック URL 直結**（`https://translate-prompt-api.fly.dev`） |
| SPA 公開 URL | 当初 `https://translate.tattsum.com` → Pages カスタムドメイン banned | **`https://translate.tattsum.com` で動作確認済み**（Workers プロキシ経由） |
| API 公開 URL | 当初 `api.prompt.tattsum.com` → 変更後 `prompt-api.tattsum.com` | **`https://prompt-api.tattsum.com` で動作確認済み** |

Tunnel は個人 β では見送り。Workers の `ORIGIN_URL` が Fly のパブリック URL を向いている。

---

## 2. 実装済みコード（WP-1〜4）

ブランチ: **`feat/web-deployment`**（PR #1 は **MERGED**）  
PR: https://github.com/Tattsum/translate-prompt/pull/1

### 2.1 バックエンド（WP-1）

| 項目 | 内容 |
|------|------|
| 環境変数 | `LISTEN_HOST`, `ALLOWED_ORIGINS`, `INVESTIGATE_ENABLED` |
| 設定 | `backend/infrastructure/config/server.go` |
| Investigate ガード | GraphQL + Connect-RPC |
| 本番ビルド | `make build-server-api`（`-tags noembed`、SPA 埋め込みなし） |
| コンテナ | `Dockerfile`（distroless、`LISTEN_HOST=0.0.0.0` デフォルト） |
| Fly 定義 | `fly.toml`（`nrt`, 512MB, `min_machines_running = 1`） |

`fly.toml` の現行 CORS:

```toml
ALLOWED_ORIGINS = "https://translate.tattsum.com"
```

※ SPA が実際に `translate-prompt.pages.dev` で公開される場合は、ここを更新する必要あり。

### 2.2 フロントエンド（WP-2）

| 項目 | 内容 |
|------|------|
| API ベース URL | `VITE_API_BASE_URL`（`frontend/src/api/client.ts`） |
| Workspace Path 非表示 | `VITE_ENABLE_WORKSPACE_PATH` |
| 本番 env | `frontend/.env.production` |

```env
VITE_API_BASE_URL=https://prompt-api.tattsum.com
VITE_ENABLE_WORKSPACE_PATH=false
```

### 2.3 Workers API プロキシ（WP-3）

| 項目 | 内容 |
|------|------|
| 設定 | `wrangler.toml` |
| 実装 | `workers/src/index.ts` |
| Worker 名 | `translate-prompt-api-proxy` |
| ルート | `prompt-api.tattsum.com/*` |
| 転送対象 | `POST /query`（GraphQL）、`/translate_prompt.v1.TranslatePromptService/*`（Connect-RPC）、`OPTIONS` |
| CORS | Go オリジン側（`ALLOWED_ORIGINS`）に委譲。Workers はそのまま返す |
| Access ヘッダ除去 | `cf-access-jwt-assertion` 等をオリジンに転送しない |

シークレット `ORIGIN_URL`（CI で設定）:

```
https://translate-prompt-api.fly.dev
```

### 2.3b Workers SPA プロキシ（WP-3b）

| 項目 | 内容 |
|------|------|
| 設定 | `wrangler.web.toml` |
| 実装 | `workers/src/web.ts` |
| Worker 名 | `translate-prompt-web` |
| ルート | `translate.tattsum.com/*` |
| 転送先 | `https://translate-prompt.pages.dev`（`PAGES_HOST`） |

### 2.4 CI/CD（WP-4）

ファイル: `.github/workflows/deploy.yml`

| ジョブ | 内容 |
|--------|------|
| `test` | `make install-tools` → `make test` → `make lint` |
| `deploy-fly` | `flyctl secrets set ALLOWED_ORIGINS=...` → `flyctl deploy` |
| `deploy-workers` | API + SPA プロキシ Worker デプロイ（`wrangler.toml` + `wrangler.web.toml`） |
| `deploy-pages` | プロジェクト自動作成 → `wrangler pages deploy ../frontend/dist` |

トリガー: `push` to `main` / `master`、`pull_request`、`workflow_dispatch`

GitHub Secrets（設定済み）:

| Secret | 用途 |
|--------|------|
| `FLY_API_TOKEN` | Fly.io デプロイ |
| `CLOUDFLARE_API_TOKEN` | Workers / Pages デプロイ |
| `CLOUDFLARE_ACCOUNT_ID` | Cloudflare アカウント識別 |

最新成功 CI run（参照用）: `28581696601`（test, deploy-fly, deploy-workers, deploy-pages すべて成功）

### 2.5 ローカル開発の維持

`make serve` / `make dev` はデフォルト値のまま従来どおり動作する設計。**変更時も維持すること。**

---

## 3. インフラ現状（2026-07-03 時点）

### 3.1 動作確認済み

| コンポーネント | URL | 状態 | 確認方法 |
|----------------|-----|------|---------|
| Fly API（直） | `https://translate-prompt-api.fly.dev/query` | OK | GraphQL health |
| Workers API | `https://prompt-api.tattsum.com/query` | OK | `{"data":{"health":{"status":"ok"}}}` |
| Pages（デフォルト） | `https://translate-prompt-2el.pages.dev` | 200 | HTTP GET |
| **SPA Worker** | `https://translate.tattsum.com` | **302**（Access ログイン） | 未認証リダイレクト |
| **Cloudflare Access** | SPA + API | **設定完了** | IdP: Google + GitHub、`invite-only` Policy |
| Estimate / Analyze | `prompt-api.tattsum.com` | OK | GraphQL 直接確認 |
| Investigate（Web） | `prompt-api.tattsum.com` | OK | `investigate disabled` |
| CORS | `prompt-api.tattsum.com` | OK | `translate.tattsum.com` のみ許可 |
| Playground | `prompt-api.tattsum.com/playground` | OK | 404 |

### 3.2 未達・問題あり

（2026-07-03 時点: Cloudflare Access 設定完了。残タスクは E2E ブラウザ確認のみ）

| コンポーネント | 状態 | 備考 |
|----------------|------|------|
| E2E（ログイン後 Analyze/Optimize） | 随時確認 | Access 設定済み |

### 3.3 Cloudflare アカウント

| 項目 | 値 |
|------|-----|
| 本番デプロイ先アカウント | **`Kurohari35@gmail.com's Account`**（`tattsum.com` ゾーンあり） |
| ローカル `wrangler whoami` | **別アカウント**（`tatsuma.kano@gmail.com`） |

ローカルから `wrangler deploy` するとゾーン未検出エラーになる。**CI（正しい API トークン）経由ではデプロイ成功。**

### 3.4 Cloudflare リソース

| 種別 | 名前 / 設定 |
|------|------------|
| Fly アプリ | `translate-prompt-api`（リージョン `nrt`） |
| Workers | `translate-prompt-api-proxy`（API）、`translate-prompt-web`（SPA・新規） |
| Workers ルート | `prompt-api.tattsum.com/*`、`translate.tattsum.com/*`（CI デプロイ後） |
| Workers シークレット | `ORIGIN_URL=https://translate-prompt-api.fly.dev` |
| Pages プロジェクト | `translate-prompt` |
| Pages Production branch | `master` |

### 3.5 DNS（tattsum.com ゾーン、ユーザー提供スクショベース）

| 名前 | タイプ | 内容 | プロキシ | 備考 |
|------|--------|------|---------|------|
| `prompt-api` | A | `192.0.2.1` | ON | Workers ルート経由で API 動作 |
| `translate` | CNAME | `translate-prompt-2el.pages.dev` | ON | **522** → **A `192.0.2.1` に置換予定**（[deployment-dns-setup.md](./deployment-dns-setup.md)） |
| `www` | CNAME | `tattsum.com` | — | www → apex リダイレクト |
| `tattsum.com`（apex） | **Worker** | `blog` | — | blog サイト（後述） |

---

## 4. 遭遇した問題と解決策

| # | 問題 | 原因 | 対応 | 状態 |
|---|------|------|------|------|
| 1 | Fly ヘルスチェック失敗 / 522 | `LISTEN_HOST` 未設定で `127.0.0.1` バインド | Dockerfile + `fly.toml` で `0.0.0.0` | 解決 |
| 2 | Fly CI `context deadline exceeded` | Depot ビルダータイムアウト | `--local-only --depot=false` | 解決 |
| 3 | Workers ゾーン未検出（ローカル） | ローカル wrangler が別アカウント | CI + 正しい Secrets でデプロイ | 解決（ローカルは未対応） |
| 4 | Pages `Project not found` | プロジェクトが別アカウントにのみ存在 | CI で `wrangler pages project create` を追加 | 解決 |
| 5 | `api.prompt.tattsum.com` TLS handshake failure | 入れ子サブドメインは Universal SSL 非対象 | `prompt-api.tattsum.com` に変更 | 解決 |
| 6 | Total TLS | 有料プランが必要 | 見送り | — |
| 7 | `prompt.tattsum.com` Pages 追加 | **banned domain** エラー | カスタムドメイン不可 | 未解決 |
| 8 | `translate.tattsum.com` Pages 追加 | **banned domain** エラー | カスタムドメイン不可 | 未解決 |
| 9 | `translate.tattsum.com` 522 | DNS CNAME のみ。Pages カスタムドメイン UI で登録できない | Workers プロキシ + DNS A レコードに切替 | **解決** |
| 10 | Cloudflare Access | 未設定 | [deployment-access-setup.md](./deployment-access-setup.md) で Google/GitHub IdP + Application 2 つ | **解決**（2026-07-03） |

### banned domain について

Cloudflare Pages の「カスタムドメイン追加」UI で `*.tattsum.com` が拒否される。  
エラーメッセージは **banned domain** 系。

任意の問い合わせ先: `abusereply@cloudflare.com`（`tattsum.com` の Pages カスタムドメイン解禁依頼）

---

## 5. blog（tattsum.com）の構成 — 参考パターン

ユーザーが「blog の構成を参考にできるか」と質問。調査結果を以下に記録。

### 5.1 blog の実態

```
tattsum.com（DNS: タイプ Worker → "blog"）
    ↓
Cloudflare Worker「blog」
    ↓
OpenNext + Next.js（Workers 上で動作）
```

`curl -I https://tattsum.com/` で確認できたレスポンスヘッダー:

```
HTTP/2 200
content-type: text/html; charset=utf-8
x-powered-by: Next.js
x-opennext: 1
server: cloudflare
```

`www.tattsum.com` は `301` → `https://tattsum.com/`

### 5.2 blog と translate-prompt の比較

| 項目 | blog（tattsum.com） | translate-prompt（現状） |
|------|---------------------|--------------------------|
| 配信基盤 | **Cloudflare Workers** | **Cloudflare Pages** + カスタムドメイン（失敗） |
| DNS | タイプ **Worker** → `blog` | タイプ **CNAME** → `*.pages.dev` |
| カスタムドメイン UI | **Pages のカスタムドメイン機能を使っていない** | Pages カスタムドメインで **banned** |
| 結果 | 200 OK | 522 / banned |

### 5.3 参考にできる方針（合意候補）

**Pages のカスタムドメインを諦め、SPA も Workers 経由で公開する**（blog と同型）。

#### 推奨アーキテクチャ（方式 A: Pages プロキシ）

```
translate.tattsum.com（DNS: Worker または Workers Route）
    ↓
Worker「translate-prompt-web」（新規）
    ↓ リバースプロキシ
translate-prompt.pages.dev（CI でデプロイ済みの静的 SPA）
```

```
prompt-api.tattsum.com（既存・動作中）
    ↓
Worker「translate-prompt-api-proxy」（既存）
    ↓
https://translate-prompt-api.fly.dev（Fly.io）
```

#### 実装方式の選択肢

| 方式 | 内容 | 難易度 | 備考 |
|------|------|--------|------|
| **A. Pages プロキシ** | 新 Worker が `translate-prompt.pages.dev` に転送 | **低（推奨）** | Pages CI はそのまま。公開ドメインだけ Worker が肩代わり |
| B. Workers Static Assets | `frontend/dist` を Worker に直接デプロイ | 中 | Pages と二重管理になりやすい |
| C. OpenNext 化 | blog と同じ Next.js on Workers | 高 | Vite/React SPA には過剰 |

#### DNS 設定（blog と同様、Pages UI は使わない）

1. Pages の「カスタムドメイン」UIは**使わない**
2. いずれか:
   - **DNS タイプ Worker** → 新 Worker 名（blog と同型）
   - **Workers Route** + `A` レコード `192.0.2.1`（`prompt-api` と同型）
3. 既存の `translate` CNAME（Pages 向け）は削除または置き換え

#### banned との関係（推定）

| 機能 | tattsum.com サブドメイン |
|------|--------------------------|
| **Pages カスタムドメイン** | banned |
| **Workers ルート / Worker DNS** | `prompt-api.tattsum.com` と `tattsum.com`（blog）で動作実績あり |

### 5.4 β 代替案（Workers 実装を後回しにする場合）

SPA を **`https://translate-prompt.pages.dev`** のまま運用し、以下を合わせる:

- Fly `ALLOWED_ORIGINS` → `https://translate-prompt.pages.dev`
- Cloudflare Access → `translate-prompt.pages.dev` + `prompt-api.tattsum.com`

カスタムドメインなしで最短 β 公開は可能。ただし URL が `*.pages.dev` になる。

---

## 6. 未完了タスク（次セッション）

### 6.1 必須（公開まで）

| # | タスク | 詳細 | 状態 |
|---|--------|------|------|
| 1 | **SPA Workers プロキシ実装** | `wrangler.web.toml` + `workers/src/web.ts` + CI | **完了** |
| 2 | **`ALLOWED_ORIGINS` 更新** | `fly.toml` + CI secrets 同期 | **完了** |
| 3 | **DNS 整理** | [deployment-dns-setup.md](./deployment-dns-setup.md) | **完了** |
| 4 | **Cloudflare Access 設定** | [deployment-access-setup.md](./deployment-access-setup.md) | **完了**（2026-07-03） |
| 5 | **E2E スモークテスト** | `./scripts/deployment-smoke-test.sh` + ブラウザ確認 | **完了**（2026-07-03） |

### 6.2 推奨（品質・運用）

| # | タスク | 詳細 |
|---|--------|------|
| 6 | `docs/deployment.md` 更新 | Tunnel → Fly 直結、ドメイン問題、blog パターンを反映 |
| 7 | `deployment-implementation-checklist.md` 完了更新 | WP-1〜4 を `[x]`、WP-5〜7 を進行中に |
| 8 | `docs/architecture.md` 本番構成追記 | WP-7 |
| 9 | Tunnel 導入（任意） | セキュリティ強化時。現状は Fly パブリック URL で運用中 |
| 10 | Cloudflare 問い合わせ（任意） | Pages カスタムドメイン解禁 |

### 6.3 方式 A 実装済みファイル

```
workers/src/web.ts              # Pages プロキシ Worker
wrangler.web.toml               # translate-prompt-web 用
.github/workflows/deploy.yml    # web Worker デプロイ + Fly secrets 同期
scripts/deployment-smoke-test.sh
docs/deployment-access-setup.md
docs/deployment-dns-setup.md
fly.toml                        # ALLOWED_ORIGINS=https://translate.tattsum.com
docs/deployment.md              # 実態反映済み
```

---

## 7. 主要ファイルパス

| パス | 役割 |
|------|------|
| `docs/deployment.md` | 設計・合意・WP 定義（要更新） |
| `docs/deployment-implementation-checklist.md` | 進捗チェックリスト |
| `docs/deployment-session-handoff.md` | **本ドキュメント（現状スナップショット）** |
| `fly.toml` | Fly 環境変数・マシン設定 |
| `Dockerfile` | API コンテナビルド |
| `wrangler.toml` | API プロキシ Worker |
| `wrangler.web.toml` | SPA プロキシ Worker |
| `workers/src/index.ts` | API プロキシ実装 |
| `workers/src/web.ts` | SPA プロキシ実装 |
| `scripts/deployment-smoke-test.sh` | 本番スモークテスト |
| `frontend/.env.production` | フロント本番 env |
| `frontend/src/api/client.ts` | API クライアント |
| `.github/workflows/deploy.yml` | CI/CD |
| `backend/infrastructure/config/server.go` | サーバ設定 |
| `backend/cmd/server/main.go` | エントリポイント |

---

## 8. スモークテストチェックリスト（未実施分）

| # | 確認項目 | 期待結果 | 状態 |
|---|---------|---------|------|
| 1 | 未認証で SPA URL にアクセス | Access ログイン画面（302） | **OK** |
| 2 | 認証後 SPA 表示 | Input ページ | **OK** |
| 3 | Health | `prompt-api.tattsum.com` 経由で `ok` | **OK** |
| 4 | Estimate | トークン数が返る | **OK** |
| 5 | Analyze | 質問または ready | **OK** |
| 6 | Optimize | 最適化結果 | **OK**（Access ログイン後） |
| 7 | Investigate（Web） | 403 / エラー | **OK** |
| 8 | CLI Investigate | ローカルで動作 | 未確認 |
| 9 | CORS | 許可オリジンのみ拒否 | **OK** |
| 10 | Playground | 本番 404 | **OK** |
| 11 | SPA Worker + Access | `translate.tattsum.com` 302 → ログイン | **OK** |

---

## 9. セキュリティメモ（現状）

| 項目 | 状態 |
|------|------|
| `INVESTIGATE_ENABLED=false`（Fly） | 設定済み |
| CORS 制限 | 設定済み（`https://translate.tattsum.com`） |
| Access 招待制 | **設定済み**（Google + GitHub、`invite-only` Policy） |
| Fly オリジン非公開（Tunnel） | **未導入**（パブリック URL 直結） |
| GraphQL Playground 本番無効 | 確認済み（404） |

---

## 10. 次セッション用プロンプト（コピペ用）

以下を新しいセッションの最初のメッセージとして貼り付ける。

---

```
translate-prompt 招待制 β は公開完了（2026-07-03）。

## 必読

- docs/deployment-session-handoff.md
- docs/deployment-access-setup.md

## 完了済み

- Cloudflare Access（Google + GitHub IdP、`invite-only` Policy）
- translate.tattsum.com / prompt-api.tattsum.com — 未認証 302
- スモークテスト + ブラウザ E2E（Estimate / Analyze / Optimize）

## 運用タスク

- β 招待: Policy の Emails に追加（deployment-access-setup.md §5）
- CLI Investigate ローカル動作の随時確認
- Tunnel 導入（任意・セキュリティ強化時）
```

---

## 関連リンク

- PR #1: https://github.com/Tattsum/translate-prompt/pull/1（MERGED）
- Pages: https://translate-prompt.pages.dev
- API: https://prompt-api.tattsum.com/query
- blog 参考: https://tattsum.com（Worker `blog`、OpenNext）
