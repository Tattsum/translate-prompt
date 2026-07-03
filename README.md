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
| [docs/llm-setup.md](./docs/llm-setup.md) | LLM API キー取得・環境変数設定 |
| [docs/deployment.md](./docs/deployment.md) | Web 公開デプロイ設計（Cloudflare） |
| [docs/deployment-access-setup.md](./docs/deployment-access-setup.md) | Cloudflare Access 設定手順 |

## Web β（本番）

| コンポーネント | URL |
|---------------|-----|
| SPA | https://translate.tattsum.com |
| API | https://prompt-api.tattsum.com |

```bash
# 本番スモークテスト（自動チェック）
./scripts/deployment-smoke-test.sh
```

招待制 β のため Cloudflare Access 設定が必要です。手順は [docs/deployment-access-setup.md](./docs/deployment-access-setup.md) を参照してください。

## 依存関係の自動更新（Renovate）

[`renovate.json`](./renovate.json) で Go / npm（frontend・workers）/ GitHub Actions / Docker / ツールチェーン（mise・Makefile・buf）を管理します。

- **GitHub Actions**: [`.github/workflows/renovate.yml`](./.github/workflows/renovate.yml) が毎週月曜 10:00 JST に実行され、更新 PR を作成します
- **Dependency Dashboard**: Renovate が Issue を 1 件作成し、保留中の更新を一覧できます

[Mend Renovate GitHub App](https://github.com/apps/renovate) を併用する場合は、Actions ワークフローと二重実行にならないようどちらか一方にしてください。

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
