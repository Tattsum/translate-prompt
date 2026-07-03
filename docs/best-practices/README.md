# ベストプラクティスルール定義

`docs/best-practices/*.yaml` は実装の**単一の参照源（SSOT）**です。
`infrastructure/bestpractice` はこの YAML を読み込んで Stage を適用します。

App-2（LLM Refiner / AST）のルール拡張は [phase2-app-roadmap.md](../phase2-app-roadmap.md) を正とする。

## YAML スキーマ（基本）

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

## App-2 向け拡張（`automatable: false`）

LLM Refiner Stage 用。`automatable: true` ルールで処理した**残差**に適用する。

```yaml
  - id: common-example-summarize
    description: examples Section の意味保持要約
    automatable: false
    pipeline: compress
    stage: LLMRefiner
    action: summarize_preserve_meaning
    condition:
      section_tag: examples
      remaining_tokens_over_budget: true
    constraints:
      must_not_increase_tokens: true
    intake_on_failure: false
    llm:
      max_output_tokens: 1024
```

| キー | 型 | 説明 |
|------|-----|------|
| `stage: LLMRefiner` | string | Compress パイプライン内の LLM Stage |
| `condition.section_tag` | string | `Section.Metadata["xml_tag"]` または `SectionType` とのマッチ |
| `condition.section_type` | string | `Section.Type`（例: `Task`, `Rules`）とのマッチ |
| `condition.remaining_tokens_over_budget` | bool | 予算逼迫時のみ Refiner 実行 |
| `constraints.must_not_increase_tokens` | bool | 出力トークン増なら棄却 |
| `llm.max_output_tokens` | int | ルール単位の LLM 出力上限 |

初回合意ルール 3 個の定義: [phase2-app-roadmap.md §6.3](../phase2-app-roadmap.md#63-初回-automatable-false-ルール合意-3-個)

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
| `true` | 決定的 Stage で実装（Phase 1 完了分 + 今後の pattern ベース） |
| `false` | LLM Refiner Stage（App-2）。`RulesForRefinement()` が読み込む |

**原則**: 既存 `true` ルールは pattern マッチ分を維持し、語義変換が必要な残差のみ `false` ルールに委譲する。

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
