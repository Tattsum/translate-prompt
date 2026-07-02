# ベストプラクティスルール定義

`docs/best-practices/*.yaml` は実装の**単一の参照源（SSOT）**です。
`infrastructure/bestpractice` はこの YAML を読み込んで Stage を適用します。

## YAML スキーマ

```yaml
profile: claude          # claude | codex | openai | devin | cursor
version: "1.0.0"
last_reviewed: "2026-07-02"
references:              # 公式ドキュメント URL
  - url: https://...
    title: "..."

rules:
  - id: claude-explicit
    description: 曖昧な依頼口調を命令形に変換
    source_url: https://platform.claude.com/docs/...
    source_section: "Be explicit and clear"
    automatable: true
    pipeline: format       # format | compress
    stage: ApplyBestPracticeProfile
    patterns:              # オプション: 検出用正規表現
      - "してほしい"
      - "お願い"
    action: rewrite_imperative  # 実装側で解釈するアクション ID
    intake_on_failure: true     # 自動化失敗時 Intake へ
```

## ファイル構成

| ファイル | 内容 |
|---------|------|
| `common.yaml` | 全 Profile 共通（filler 除去、例の集約） |
| `claude.yaml` | Claude XML 構造化 |
| `codex.yaml` | 4 要素テンプレート |
| `openai.yaml` | outcome-first + 簡潔化 |
| `devin.yaml` | What/How/Result, Session Brief |
| `cursor.yaml` | Rules/Skills 分離, .mdc 生成 |

## automatable フラグ

| 値 | 意味 |
|----|------|
| `true` | Phase 1 でスクリプト実装 |
| `false` | Phase 2 LLM Stage 候補。README に記載のみ |

## 公式参照元一覧

| プロバイダ | URL |
|-----------|-----|
| Claude | https://platform.claude.com/docs/en/build-with-claude/prompt-engineering/claude-prompting-best-practices |
| Codex | https://developers.openai.com/codex/learn/best-practices |
| OpenAI API | https://developers.openai.com/api/docs/guides/prompt-guidance |
| Devin | https://docs.devin.ai/learn-about-devin/prompting |
| Cursor Rules | https://cursor.com/docs/rules |
| Cursor Skills | https://cursor.com/help/customization/skills |

## レビュー手順

1. 公式ドキュメント更新を確認（四半期ごと推奨）
2. `last_reviewed` を更新
3. ルール ID の追加・廃止は `version` を bump
4. ゴールデンファイルテストを更新
