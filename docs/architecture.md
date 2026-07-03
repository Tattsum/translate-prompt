# アーキテクチャ

## リポジトリ構成

```
translate-prompt/
├── backend/              # Go（DDD コア）
│   ├── cmd/              # CLI / server エントリポイント
│   ├── domain/
│   ├── application/
│   ├── infrastructure/
│   ├── presentation/     # connectrpc, cli, mapper
│   ├── graph/            # gqlgen GraphQL
│   ├── gen/              # protobuf + connect-go 生成物
│   └── proto/            # API 契約（SSOT）
├── frontend/             # React SPA（go:embed で配信）
└── docs/                 # 設計・ルール定義
```

## レイヤー構成（DDD）

`backend/` 内:

| パッケージ | 責務 |
|-----------|------|
| `domain/prompt` | `Prompt`, `Section`, `SectionType` |
| `domain/budget` | `TokenBudget`, `OptimizeConfig` |
| `domain/optimizer` | `Stage`, `Pipeline` |
| `domain/intake` | `Ambiguity`, `Question` |
| `domain/bestpractice` | `TargetProfile`, `Rule` |
| `application/optimize` | 最適化ユースケース |
| `application/intake` | 深堀りオーケストレーション |
| `graph` | gqlgen GraphQL リゾルバ |
| `presentation/connectrpc` | Connect-RPC ハンドラ |
| `presentation/cli` | CLI |

## API（front / server 間）

| プロトコル | エンドポイント | 用途 |
|-----------|---------------|------|
| **GraphQL** (gqlgen) | `POST /query` | analyze, investigate, estimate（urql） |
| **Connect-RPC** (connect-go) | `/translate_prompt.v1.TranslatePromptService/*` | health, optimize（型安全 RPC） |
| Playground (dev) | `GET /playground` | GraphQL 探索（`ENV=dev`） |

フロントエンドは **GraphQL + Connect-Web** を併用する（github-analytics の gqlgen パターン + connect-go 型安全 RPC）。

## データフロー

```
[React SPA]
   ├─ urql ──────────► POST /query (GraphQL)
   └─ connect-web ───► Connect-RPC
              │
              ▼
[graph resolvers / connectrpc.Service]
              │
              ▼
[application/* UseCase]
              │
              ▼
[domain + infrastructure pipelines]
```

FormatPipeline → CompressPipeline の順序は変更なし。

## Web 配信

`frontend/dist` → `go:embed` → `backend/cmd/server/spa.go`

```bash
make serve   # Go API :8080
make dev     # Vite :5173（/query, Connect をプロキシ）
```

## セキュリティ

- ローカル開発: `127.0.0.1` のみバインド
- ワークスペース調査の境界制限は従来どおり

## 本番構成（Web 公開）

招待制 β のデプロイ設計は [deployment.md](./deployment.md) を正とする。  
現状スナップショット: [deployment-session-handoff.md](./deployment-session-handoff.md)

| 環境 | SPA | API | 認証 |
|------|-----|-----|------|
| ローカル | `make serve`（go:embed）または `make dev`（Vite） | `:8080` | なし |
| 本番 β | Workers → Pages（`translate.tattsum.com`） | Workers → Fly.io（`prompt-api.tattsum.com`） | Cloudflare Access（設定手順: [deployment-access-setup.md](./deployment-access-setup.md)） |

### 本番トラフィックフロー

```
[ブラウザ]
    │
    ├─ https://translate.tattsum.com
    │       ↓ Cloudflare Access（招待制）
    │       ↓ Worker「translate-prompt-web」
    │       ↓ リバースプロキシ
    │       translate-prompt.pages.dev（Pages 静的 SPA）
    │
    └─ https://prompt-api.tattsum.com
            ↓ Cloudflare Access（招待制）
            ↓ Worker「translate-prompt-api-proxy」
            ↓ リバースプロキシ（Access JWT はオリジンに転送しない）
            https://translate-prompt-api.fly.dev（Go API）
```

### 本番の制約

- **Investigate**（サーバー FS 読み取り）は Web から無効（`INVESTIGATE_ENABLED=false`）。CLI のみ
- **Workspace Path** は本番ビルドで非表示（`VITE_ENABLE_WORKSPACE_PATH=false`）
- **CORS** は `ALLOWED_ORIGINS=https://translate.tattsum.com` のみ許可
- **GraphQL Playground** は `ENV=production` で無効（404）
- **Tunnel** は未導入。Fly はパブリック URL 直結（個人 β）

### 関連ドキュメント

| ドキュメント | 内容 |
|-------------|------|
| [deployment.md](./deployment.md) | 合意サマリ・WP 定義 |
| [deployment-dns-setup.md](./deployment-dns-setup.md) | DNS / Workers Route |
| [deployment-access-setup.md](./deployment-access-setup.md) | Cloudflare Access 手動設定 |
| [deployment-implementation-checklist.md](./deployment-implementation-checklist.md) | 実装進捗 |
