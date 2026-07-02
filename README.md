# translate-prompt

エージェント向けプロンプトを、Claude / OpenAI / Devin / Cursor の公式ベストプラクティスに沿って整形し、トークン予算内に最適化するツール。

## クイックスタート

```bash
# ツールチェーン（Go 1.26.4 / Node LTS / pnpm / golangci-lint 2.12）
mise install
make install-tools   # golangci-lint v2

# フォーマット（go fix x2 + gofumpt）
make fmt

# テスト・リント・ビルド
make test
make lint
make build

# Web UI（API + SPA）
make serve          # http://127.0.0.1:8080

# 開発（Vite + API プロキシ）
make dev            # http://127.0.0.1:5173

# CLI
cat prompt.md | ./bin/translate-prompt --max-tokens 8000 --target-profile codex
```

## ドキュメント

実装は [docs/README.md](./docs/README.md) から開始してください。

| ドキュメント | 内容 |
|-------------|------|
| [docs/architecture.md](./docs/architecture.md) | アーキテクチャ・パイプライン |
| [docs/profiles.md](./docs/profiles.md) | TargetProfile 仕様 |
| [docs/best-practices/](./docs/best-practices/) | ルール定義 YAML |
| [docs/intake.md](./docs/intake.md) | 深堀りフロー |
| [docs/api.md](./docs/api.md) | REST API / CLI |
| [docs/implementation-roadmap.md](./docs/implementation-roadmap.md) | 実装チェックリスト |

## 設計原則

- **ルールで整形できることは LLM に投げない**（Phase 1 はルールベースのみ）
- FormatPipeline（品質）→ CompressPipeline（トークン削減）の 2 段実行
- Web UI（React）+ CLI 並行、コアは Go + DDD

## モジュール

```
github.com/Tattsum/translate-prompt          # ルート go.mod
github.com/Tattsum/translate-prompt/backend/...  # Go パッケージ
github.com/Tattsum/translate-prompt/frontend       # SPA embed
```
