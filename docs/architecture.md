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

| 環境 | SPA | API | 認証 |
|------|-----|-----|------|
| ローカル | `make serve`（go:embed）または `make dev`（Vite） | `:8080` | なし |
| 本番 β | Cloudflare Pages（`translate.tattsum.com`） | Workers → Tunnel → Fly.io（`prompt-api.tattsum.com`） | Cloudflare Access |

本番では Investigate（サーバー FS 読み取り）は Web から無効。CLI のみ。
