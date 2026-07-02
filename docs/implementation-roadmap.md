# Phase 1 実装ロードマップ

別セッションでの実装チェックリスト。完了したら `[x]` に更新する。

## 0. スキャフォールド

- [x] `git init` + `go mod init github.com/Tattsum/translate-prompt`
- [x] `Makefile`（help, test, lint, build, web-build, serve, dev）
- [x] `mise.toml`（Go 1.26, Node LTS, pnpm）
- [x] `cmd/translate-prompt/main.go`
- [x] `cmd/server/main.go`
- [x] ルート `README.md`（docs へのリンク）

## 1. ドメイン層

- [x] `domain/prompt` — Prompt, Section, SectionType
- [x] `domain/budget` — TokenBudget, OptimizeConfig
- [x] `domain/bestpractice` — TargetProfile, Rule 読み込み
- [x] `domain/optimizer` — Stage interface, Pipeline
- [x] `domain/intake` — Question, Ambiguity, InvestigationResult

## 2. インフラ層

- [x] `infrastructure/tokenizer` — tiktoken-go
- [x] `infrastructure/stages` — Normalize, Dedup, Boilerplate, Budget, Truncate
- [x] `infrastructure/bestpractice` — FormatPipeline stages
- [x] `infrastructure/workspace` — BoundedFSReader
- [x] `docs/best-practices/*.yaml` を読み込むローダー

## 3. アプリケーション層

- [x] `application/intake` — Analyze, Investigate, Merge
- [x] `application/optimize` — Format → Compress オーケストレーション

## 4. Presentation

- [x] `presentation/cli` — flag, stdin/file
- [x] `presentation/web` — REST handlers
- [x] `cmd/server/spa.go` — SPA 配信

## 5. フロントエンド

- [x] `frontend/` — Vite + React + TypeScript + pnpm
- [x] `frontend/embed.go` — go:embed dist
- [x] ページ: Input, Settings, Intake, Result
- [x] Cursor 時 Result タブ: Task / .mdc 案
- [x] `vite.config.ts` — /api proxy
- [x] Vitest — API クライアント, 主要コンポーネント

## 6. テスト

- [x] Stage 単体テスト（table-driven, t.Parallel）
- [x] intake テスト
- [x] `testdata/*.md` ゴールデンファイル
- [x] Profile 別 before/after fixture

## 7. 成功基準

- [x] 代表プロンプトで 20%+ トークン削減
- [x] 5 Profile すべて公式準拠構造で出力
- [x] OptimizeReport に rule ID + source_url
- [x] `make test` / `make lint` / `make serve` 通過

## ファイル一覧（新規作成）

```
go.mod
Makefile
mise.toml
README.md
cmd/translate-prompt/main.go
cmd/server/main.go
cmd/server/spa.go
domain/prompt/prompt.go
domain/budget/budget.go
domain/bestpractice/profile.go
domain/optimizer/pipeline.go
domain/optimizer/stage.go
domain/intake/intake.go
application/optimize/usecase.go
application/intake/usecase.go
infrastructure/tokenizer/tiktoken.go
infrastructure/stages/*.go
infrastructure/bestpractice/*.go
infrastructure/workspace/reader.go
presentation/cli/cli.go
presentation/web/handler.go
frontend/embed.go
frontend/package.json
frontend/vite.config.ts
frontend/src/pages/*.tsx
frontend/src/api/*.ts
```

## Phase 2 / 3（スコープ外）

- Phase 2: LLM Refiner（automatable: false ルール）、AST 圧縮
- Phase 3: MCP サーバー `optimize_prompt`
