# Web 公開デプロイ設計（Cloudflare 中心）

別セッション・サブエージェントでの実装引き継ぎ用ドキュメント。  
`/dotfiles-plan-first` による合意内容を正とする。

## 合意サマリ

| 決定事項 | 結論 |
|---------|------|
| 公開スコープ | **招待制 β**（Cloudflare Access） |
| Web 機能 | **Analyze / Optimize / Estimate** のみ |
| Investigate | **Web 無効**（`INVESTIGATE_ENABLED=false`）。CLI のみ維持 |
| オリジン | **Fly.io**（Go モノリス維持） |
| フロント | **Cloudflare Pages**（SPA 分離） |
| エッジ API | **Cloudflare Workers**（薄いプロキシ） |
| オリジン到達 | **Cloudflare Tunnel**（`cloudflared`、パブリック IP 非公開） |
| 認証 | **Cloudflare Access**（ユーザ DB なし） |
| DB | **当面なし**（ステートレス） |
| CI/CD | **GitHub Actions**（Fly + Pages + Workers 自動デプロイ） |
| ドメイン | **tattsum.com**（Cloudflare DNS 管理済み） |
| SPA URL | `https://translate.tattsum.com` |
| API URL | `https://prompt-api.tattsum.com` |

### 見送り（Phase 2 以降）

- Hono / D1 / Clerk / Better Auth
- Workers への Go WASM 移植
- Investigate の ZIP アップロード対応
- 一般公開 SaaS 化（ユーザー登録・課金）

---

## アーキテクチャ

```
[ブラウザ]
    │ HTTPS
    ▼
[Cloudflare Access]  ← Google / GitHub ログイン（招待制）
    ├─ translate.tattsum.com ────── Cloudflare Pages（React SPA）
    └─ prompt-api.tattsum.com ── Cloudflare Workers（API プロキシ）
            │
            │ Cloudflare Tunnel（cloudflared、Fly 内で常駐）
            ▼
        Fly.io（Go monolith）
            ├─ POST /query              （GraphQL: analyze, estimate）
            ├─ /translate_prompt.v1.*   （Connect-RPC: health, optimize）
            └─ SPA 配信はしない（Pages に分離）
```

### 責務分離

| レイヤ | 責務 | 置かないもの |
|--------|------|-------------|
| Pages | 静的 SPA 配信 | API ロジック |
| Workers | リバースプロキシ、CORS、（任意）レート制限 | ドメインコア |
| Tunnel | オリジン非公開化 | — |
| Fly.io | Go DDD コア（最適化パイプライン） | ユーザ認証 DB |
| Access | β 招待制の前段認証 | アプリ内セッション |

### 想定コスト

| サービス | 目安 |
|---------|------|
| Cloudflare Pages / Access（50 ユーザー） | $0 |
| Cloudflare Workers | $0〜$5/月（無料枠超過時 Paid） |
| Fly.io | ~$5/月（常時起動マシン） |
| **合計** | **~$5〜10/月** |

---

## 現状ギャップ（実装前に把握）

| 項目 | 現状 | 本番要件 |
|------|------|---------|
| バインド | `127.0.0.1` のみ（`backend/cmd/server/main.go` L59） | `0.0.0.0`（Fly 内） |
| CORS | `Access-Control-Allow-Origin: *` | `https://translate.tattsum.com` のみ |
| SPA 配信 | `go:embed`（`backend/cmd/server/spa.go`） | Pages に分離。サーバは API のみ |
| API ベース URL | 相対パス `/query`, `/`（`frontend/src/api/client.ts`） | `https://prompt-api.tattsum.com` |
| Investigate | Web/CLI 両方有効 | Web 無効（環境変数ガード） |
| Workspace Path UI | Settings に入力欄あり | Web 本番では非表示 |
| デプロイ設定 | なし | Dockerfile, fly.toml, wrangler.toml, CI |

---

## DNS 設計（tattsum.com）

Cloudflare ダッシュボードで以下を設定する（実装セッションの最終手順）。

| レコード | タイプ | 向き先 |
|---------|--------|--------|
| `translate` | CNAME | Cloudflare Pages プロジェクトの `*.pages.dev` |
| `prompt-api` | CNAME / Worker | Workers カスタムドメイン（自動設定） |

**注意**: Fly のパブリック URL は DNS に載せない。Tunnel 経由のみ。

---

## 追加・変更するファイル（実装仕様）

### 新規作成

```
Dockerfile                          # Go サーバ用マルチステージビルド
fly.toml                            # Fly.io アプリ定義
infra/cloudflared/fly.toml          # （任意）Tunnel 用別アプリ
wrangler.toml                       # Workers プロキシ定義
workers/src/index.ts                # API プロキシ（Tunnel オリジンへ転送）
.github/workflows/deploy.yml        # CI/CD
frontend/.env.production            # VITE_API_BASE_URL（git 管理可、秘密なし）
```

### 変更

| ファイル | 変更内容 |
|---------|---------|
| `backend/cmd/server/main.go` | `LISTEN_HOST` 環境変数（default `127.0.0.1`、本番 `0.0.0.0`）。CORS を `ALLOWED_ORIGINS` で制御。`INVESTIGATE_ENABLED` で Investigate 無効化 |
| `backend/graph/schema.resolvers.go` | Investigate リゾルバにガード |
| `backend/presentation/connectrpc/service.go` | Investigate RPC にガード |
| `frontend/src/api/client.ts` | `VITE_API_BASE_URL` 対応 |
| `frontend/vite.config.ts` | `envDir` / `define` 確認。本番ビルド用 |
| `frontend/src/pages/Settings.tsx` | `VITE_ENABLE_WORKSPACE_PATH` で Workspace Path 非表示 |
| `Makefile` | `build-server-api`（SPA 埋め込みなし）ターゲット追加 |
| `docs/README.md` | 本ドキュメントへのリンク |
| `docs/architecture.md` | 本番構成セクション追記 |

### ローカル開発は維持

- `make serve` / `make dev` は従来どおり localhost で動作すること
- 本番用変更は環境変数で切り替え、デフォルトはローカル向けのまま

---

## 実装詳細

### 1. Go サーバ（Fly.io オリジン）

#### 環境変数

| 変数 | ローカル | 本番（Fly） |
|------|---------|------------|
| `LISTEN_HOST` | `127.0.0.1` | `0.0.0.0` |
| `PORT` | `8080` | `8080` |
| `ENV` | `dev` | `production` |
| `INVESTIGATE_ENABLED` | `true`（省略可） | `false` |
| `ALLOWED_ORIGINS` | `*`（省略可） | `https://translate.tattsum.com` |

#### Dockerfile（方針）

```dockerfile
# ビルド: golang:1.26-alpine
# - SPA は埋め込まない（API のみビルド）
# - go build -o /server ./backend/cmd/server
# 実行: distroless または alpine
# EXPOSE 8080
# CMD ["/server", "--port", "8080"]
```

`make build-server` は `web-build` + embed 前提のため、本番用は **`build-server-api`**（新規）で embed なしビルドに分離する。

#### fly.toml（方針）

```toml
app = "translate-prompt-api"
primary_region = "nrt"  # 東京リージョン

[env]
  LISTEN_HOST = "0.0.0.0"
  ENV = "production"
  INVESTIGATE_ENABLED = "false"
  ALLOWED_ORIGINS = "https://translate.tattsum.com"

[http_service]
  internal_port = 8080
  force_https = true
  auto_stop_machines = false
  auto_start_machines = true
  min_machines_running = 1

# パブリック ingress は閉じ、Tunnel のみ許可する構成を推奨
# （Fly の services を internal のみにするか、cloudflared サイドカーで運用）
```

**Tunnel 構成（推奨）**: Fly アプリ内で `cloudflared` をサイドカーまたは同一マシンで起動し、Tunnel の `service` を `http://localhost:8080` に向ける。Fly のパブリック HTTP を無効化する。

### 2. Cloudflare Workers（API プロキシ）

#### wrangler.toml（方針）

```toml
name = "translate-prompt-api-proxy"
main = "workers/src/index.ts"
compatibility_date = "2025-01-01"

routes = [
  { pattern = "prompt-api.tattsum.com/*", zone_name = "tattsum.com" }
]

[vars]
# Tunnel のホスト名は Cloudflare Zero Trust で発行される URL
# ORIGIN_URL = "https://<tunnel-id>.cfargotunnel.com"
```

#### Workers プロキシ（方針）

転送対象パス:

- `POST /query`（GraphQL）
- `/translate_prompt.v1.TranslatePromptService/*`（Connect-RPC）
- `OPTIONS`（CORS プリフライト）

実装方針:

```typescript
// workers/src/index.ts
// - リクエストを ORIGIN_URL + path + query にそのまま転送
// - Access 通過後の JWT は転送しない（オリジンは Access 外、Tunnel で保護）
// - CORS ヘッダは Workers または Go 側のどちらか一方で付与（二重付与しない）
```

**シークレット**: `ORIGIN_URL` は `wrangler secret put ORIGIN_URL` で設定（Tunnel URL）。

### 3. Cloudflare Pages（SPA）

#### ビルド設定

| 項目 | 値 |
|------|-----|
| ルートディレクトリ | `frontend` |
| ビルドコマンド | `pnpm install && pnpm run build` |
| 出力ディレクトリ | `dist` |
| Node バージョン | LTS（`mise.toml` に合わせる） |

#### 環境変数（Pages）

| 変数 | 値 |
|------|-----|
| `VITE_API_BASE_URL` | `https://prompt-api.tattsum.com` |
| `VITE_ENABLE_WORKSPACE_PATH` | `false` |

#### フロントエンド変更（`client.ts`）

```typescript
const apiBase = import.meta.env.VITE_API_BASE_URL ?? ''

const graphqlClient = createUrqlClient({
  url: `${apiBase}/query`,
  exchanges: [fetchExchange],
})

const connectTransport = createConnectTransport({ baseUrl: apiBase || '/' })
```

### 4. Cloudflare Access

Zero Trust ダッシュボードで設定（手動・初回のみ）。

1. **Applications** → Add application → Self-hosted
2. アプリ 1: `translate.tattsum.com`（Pages）
3. アプリ 2: `prompt-api.tattsum.com`（Workers）
4. **Policy**: Allow — Emails ending in `@<your-email-domain>` または特定メール / GitHub OAuth
5. β 招待者はメールまたは IdP グループで追加

Access を Workers / Pages の前段に置くため、DNS は Cloudflare プロキシ（オレンジ雲）必須。

### 5. Investigate 無効化

#### バックエンド

`INVESTIGATE_ENABLED=false` のとき:

- GraphQL `investigate` mutation → `errors.New("investigate disabled")` または HTTP 403
- Connect `Investigate` RPC → `connect.CodeUnimplemented` または `PermissionDenied`

#### フロントエンド

- Settings の Workspace Path 入力を `VITE_ENABLE_WORKSPACE_PATH !== 'true'` で非表示
- `investigate()` 呼び出し箇所があれば UI から除去（現状 Intake ページは analyze 質問のみで investigate 未使用）

#### CLI

変更なし。`translate-prompt --workspace` はローカル利用を継続。

---

## GitHub Actions（deploy.yml 方針）

トリガー: `push` to `main`（または `workflow_dispatch`）

### ジョブ構成

```yaml
jobs:
  test:
    # make test && make lint

  deploy-fly:
    needs: test
    # flyctl deploy --remote-only
    # secrets: FLY_API_TOKEN

  deploy-workers:
    needs: test
    # wrangler deploy
    # secrets: CLOUDFLARE_API_TOKEN, CLOUDFLARE_ACCOUNT_ID

  deploy-pages:
    needs: test
    # Cloudflare Pages Git 連携が別途ある場合はスキップ可
    # または wrangler pages deploy frontend/dist
```

### 必要な GitHub Secrets

| Secret | 用途 |
|--------|------|
| `FLY_API_TOKEN` | Fly.io デプロイ |
| `CLOUDFLARE_API_TOKEN` | Workers / Pages デプロイ |
| `CLOUDFLARE_ACCOUNT_ID` | Cloudflare アカウント識別 |

**リポジトリ Secrets に載せないもの**: Tunnel トークンは Fly の secret または Cloudflare Zero Trust で管理。

---

## サブエージェント向けタスク分割

別セッションでは以下を **独立した作業単位** として並列または順次実行できる。

### WP-1: バックエンド本番対応

**プロンプト例**:

> `docs/deployment.md` の WP-1 に従い、Go サーバの本番対応を実装してください。
> - `LISTEN_HOST`, `ALLOWED_ORIGINS`, `INVESTIGATE_ENABLED` 環境変数
> - `build-server-api` Makefile ターゲット（embed なし）
> - `Dockerfile`, `fly.toml`
> - テスト追加（Investigate 無効時のガード）
> - ローカル `make serve` は壊さないこと

**完了条件**:

- [ ] `INVESTIGATE_ENABLED=false` で Investigate が 403/Unimplemented
- [ ] `LISTEN_HOST=0.0.0.0` で外部バインド可能
- [ ] `docker build` 成功

### WP-2: フロントエンド本番対応

**プロンプト例**:

> `docs/deployment.md` の WP-2 に従い、SPA の API ベース URL 外部化と Settings UI 調整を実装してください。
> - `VITE_API_BASE_URL`, `VITE_ENABLE_WORKSPACE_PATH`
> - `frontend/src/api/client.ts` 更新
> - Vitest 更新
> - ローカル `make dev` は相対パスのまま動作すること

**完了条件**:

- [ ] 本番ビルドで API が `prompt-api.tattsum.com` を向く
- [ ] Workspace Path が本番ビルドで非表示

### WP-3: Workers プロキシ

**プロンプト例**:

> `docs/deployment.md` の WP-3 に従い、Cloudflare Workers API プロキシを実装してください。
> - `wrangler.toml`, `workers/src/index.ts`
> - `/query` と Connect-RPC パスの転送
> - CORS は Go 側と整合させる

**完了条件**:

- [ ] `wrangler dev` でローカル Go にプロキシ可能
- [ ] OPTIONS プリフライト成功

### WP-4: CI/CD

**プロンプト例**:

> `docs/deployment.md` の WP-4 に従い、`.github/workflows/deploy.yml` を作成してください。
> - test → fly / workers / pages の順
> - Secrets 名はドキュメント通り

**完了条件**:

- [ ] PR で test ジョブが走る
- [ ] main push でデプロイ（Secrets 設定後）

### WP-5: インフラ手動セットアップ＆スモークテスト

**プロンプト例**:

> `docs/deployment.md` の WP-5 に従い、Cloudflare / Fly の手動セットアップ手順を実行し、スモークテストしてください。
> - Tunnel 作成・Fly への cloudflared 配置
> - Access ポリシー（prompt / api 両方）
> - DNS レコード
> - エンドツーエンド確認

**完了条件**:

- [ ] Access ログイン後に `translate.tattsum.com` が表示
- [ ] Analyze → Optimize フローが動作
- [ ] Investigate が Web から失敗する（期待通り）

---

## 手動セットアップ手順（WP-5 詳細）

### A. Fly.io

```bash
# 初回のみ
fly auth login
fly apps create translate-prompt-api
fly secrets set INVESTIGATE_ENABLED=false ALLOWED_ORIGINS=https://translate.tattsum.com

# デプロイ（CI 整備後は Actions から）
fly deploy
```

### B. Cloudflare Tunnel

1. Zero Trust → Networks → Tunnels → Create tunnel
2. コネクタ名: `translate-prompt-fly`
3. Fly マシンで `cloudflared tunnel run --token <TOKEN>`
4. Public Hostname: なし（Workers から private network 経由で到達する構成でも可）
5. または Internal URL を Workers の `ORIGIN_URL` に設定

**簡易構成（個人 β）**: Tunnel のサービスを `http://localhost:8080` に向け、Workers の `ORIGIN_URL` に Tunnel の DNS 名（`*.cfargotunnel.com`）を設定。Fly のパブリック URL は使わない。

### C. Cloudflare Workers

```bash
cd workers
pnpm install  # または npm
npx wrangler secret put ORIGIN_URL
npx wrangler deploy
```

### D. Cloudflare Pages

1. Workers & Pages → Create → Connect to Git
2. リポジトリ選択、ビルド設定は上表参照
3. カスタムドメイン `translate.tattsum.com` を追加
4. 環境変数 `VITE_API_BASE_URL`, `VITE_ENABLE_WORKSPACE_PATH` を設定

### E. Cloudflare Access

1. Access → Applications → 2 アプリ追加（prompt / api）
2. Policy: 自分のメール + 招待者
3. Session duration: 24h 程度（β）

---

## スモークテストチェックリスト

| # | 確認項目 | 期待結果 |
|---|---------|---------|
| 1 | 未認証で `translate.tattsum.com` にアクセス | Access ログイン画面 |
| 2 | 認証後 SPA 表示 | Input ページが表示 |
| 3 | Health | `prompt-api.tattsum.com` 経由で Connect Health が `ok` |
| 4 | Estimate | トークン数が返る |
| 5 | Analyze | 質問または ready が返る |
| 6 | Optimize | 最適化結果とレポートが返る |
| 7 | Investigate（Web） | 403 / エラー（無効化確認） |
| 8 | CLI Investigate | ローカルで従来どおり動作 |
| 9 | CORS | `translate.tattsum.com` 以外のオリジンから API が拒否される |
| 10 | Playground | 本番では `/playground` が 404 |

---

## セキュリティチェックリスト

- [ ] Fly オリジンは Tunnel 経由のみ（パブリック HTTP 最小化）
- [ ] `INVESTIGATE_ENABLED=false`（サーバ FS 読み取り防止）
- [ ] CORS は `translate.tattsum.com` のみ
- [ ] Access ポリシーで招待者限定
- [ ] GraphQL Playground は本番無効（`ENV=production`）
- [ ] GitHub Secrets に API トークンのみ（`.env` をコミットしない）
- [ ] リクエストボディサイズ上限の検討（将来: Workers で制限）

---

## 将来フェーズ（参考）

| フェーズ | 追加要素 |
|---------|---------|
| Phase 2（一般公開） | Turnstile、厳格レート制限、D1 で履歴 |
| Phase 3（SaaS） | Hono BFF + D1、Clerk または Better Auth、課金 |

現時点では **Phase 1（招待制 β）のみ** をスコープとする。

---

## 関連ドキュメント

- [architecture.md](./architecture.md) — アプリ内 DDD 構成
- [api.md](./api.md) — API 仕様
- [implementation-roadmap.md](./implementation-roadmap.md) — Phase 1 アプリ実装（完了済み）
- [deployment-implementation-checklist.md](./deployment-implementation-checklist.md) — 実装チェックリスト（進捗管理用）
