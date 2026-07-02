# translate-prompt ドキュメント

別セッションでの実装引き継ぎ用ドキュメント集です。設計の合意内容は本ディレクトリを正とします。

## 読む順序（実装者向け）

1. [architecture.md](./architecture.md) — 全体構成・パイプライン・パッケージ
2. [profiles.md](./profiles.md) — TargetProfile 別の出力形式と例
3. [best-practices/README.md](./best-practices/README.md) — ルール YAML のスキーマと参照元
4. [intake.md](./intake.md) — 深堀り（Intake）フロー
5. [api.md](./api.md) — REST API / CLI 仕様
6. [implementation-roadmap.md](./implementation-roadmap.md) — Phase 1 実装チェックリスト

## Web 公開（デプロイ）

7. [deployment-session-handoff.md](./deployment-session-handoff.md) — **現状スナップショット・次セッション引き継ぎ（最優先）**
8. [deployment.md](./deployment.md) — Cloudflare 中心の本番構成・合意内容・実装仕様
9. [deployment-implementation-checklist.md](./deployment-implementation-checklist.md) — デプロイ実装チェックリスト（進捗管理用）
10. [deployment-dns-setup.md](./deployment-dns-setup.md) — DNS / Workers Route（SPA 手動設定）
11. [deployment-access-setup.md](./deployment-access-setup.md) — Cloudflare Access 手動設定

## プロジェクト概要

**translate-prompt** は、エージェント向けプロンプト（タスク指示 + ルール + スキル + コード + 履歴）を、公式ベストプラクティスに沿って整形しつつトークン予算内に最適化する Go ツールです。

### 合意済み要件

| 項目 | 方針 |
|------|------|
| 最適化 | ハイブリッド（ルール前処理 → Phase 2 で LLM） |
| Phase 1 | ルールベースのみ（LLM API なし） |
| UI | React SPA + CLI 並行 |
| ベストプラクティス | Claude / OpenAI / Devin / Cursor 公式ガイド準拠 |
| 原則 | **ルールで整形できることは LLM に投げない** |

### TargetProfile 一覧

| Profile | 用途 |
|---------|------|
| `claude` | Anthropic Claude（XML 構造化） |
| `codex` | OpenAI Codex（4 要素、デフォルト） |
| `openai` | OpenAI GPT API（outcome-first + 検証） |
| `devin` | Devin（What/How/Result または Session Brief） |
| `cursor` | Cursor Agent（Task + @参照 + .mdc 案） |

## モジュールパス

```
github.com/Tattsum/translate-prompt
```

## 関連ファイル

計画上の主要パスは [implementation-roadmap.md](./implementation-roadmap.md) を参照。
