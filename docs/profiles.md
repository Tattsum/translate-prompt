# TargetProfile 仕様

`OptimizeConfig.TargetProfile` で出力形式を切り替える。

## claude

参照: https://platform.claude.com/docs/en/build-with-claude/prompt-engineering/claude-prompting-best-practices

```xml
<task>明示的・直接的なタスク指示</task>
<context>背景・調査結果</context>
<rules>規約（重複除去済み）</rules>
<examples>コード例・few-shot</examples>
<constraints>制約・禁止事項</constraints>
```

ルール: `docs/best-practices/claude.yaml` + `common.yaml`

## codex（デフォルト）

参照: https://developers.openai.com/codex/learn/best-practices

```markdown
## Goal
## Context
## Constraints
## Done when
```

ルール: `docs/best-practices/codex.yaml` + `common.yaml`

## openai

参照: https://developers.openai.com/api/docs/guides/prompt-guidance

Codex 4 要素 + 簡潔化（`openai-concise`）+ 検証ループ明示。

ルール: `docs/best-practices/openai.yaml` + `common.yaml`

## devin

参照: https://docs.devin.ai/learn-about-devin/prompting

### 一般: What / How / Result

### PR 向け: Devin Session Brief

Intake が PR 向けと判定した場合:

- Goal (one PR-sized outcome)
- Repo / base branch
- Scope — in / Scope — out
- Acceptance criteria
- Deliverable

devin-task-triage スキルの Session Brief 形式と整合。

ルール: `docs/best-practices/devin.yaml` + `common.yaml`

## cursor

参照: https://cursor.com/docs/rules , https://cursor.com/help/customization/skills

### 出力 2 系統（Result でタブ分離）

1. **チャット用 Task** — outcome-first、簡潔
2. **.mdc 推奨案** — 繰り返し規約の永続化（frontmatter 付き）

| 入力 | 処理 |
|------|------|
| タスク | チャット用に簡潔化 |
| 短い規約 | `.mdc` 案へ移動 |
| 多段手順 | `@skill-name` 参照 |
| コード | `@path` 参照 |

alwaysApply 予算: 200 語 / 50 行以下。

ルール: `docs/best-practices/cursor.yaml` + `common.yaml`
