# App-2 実装ロードマップ（LLM + AST）

別セッション・サブエージェントでの実装引き継ぎ用ドキュメント。  
`/dotfiles-plan-first` + `/grill-me` による **2026-07-03 合意内容を正** とする。

Phase 1（ルールベース最適化 + 招待制 β デプロイ）が完了したあとに着手する **アプリケーション機能拡張** の設計書。  
**Deploy-2（一般公開・Turnstile・D1 等）は本ドキュメントのスコープ外** とし、後続フェーズとして [deployment.md](./deployment.md) を参照する。

---

## 1. 目的

translate-prompt のコア価値である **深堀り（Intake）→ 最適化（Optimize）** ループに、LLM による意味理解を投資する。

| 項目 | Phase 1（現状） | App-2（本フェーズ） |
|------|----------------|-------------------|
| 最適化 | ルールベースのみ | ルール優先 + **LLM Refiner（残差）** + **AST 圧縮** |
| Intake Analyze | キーワード・ヒューリスティック | ヒューリスティック常時 + **LLM 補完** |
| LLM API | なし | Gemini（デフォルト）+ Anthropic（`claude` Profile のみ） |
| デプロイ | 招待制 β（Cloudflare Access） | **変更なし**（個人 + 招待者利用） |

### 1.1 設計原則（継承 + 追加）

1. **ルールで整形できることは LLM に投げない** — `automatable: true` は決定的 Stage のまま。
2. **LLM は残差処理** — `automatable: false` ルール、またはヒューリスティック後の補完のみ。
3. **プロンプト文字列は domain に入れない** — 「プロンプト = SQL」たとえに従い、組み立ては infrastructure（`PromptBuilder`）。
4. **TargetProfile ≠ LLM Provider** — 出力形式（codex / claude 等）と、最適化に使う LLM API は別軸。
5. **Port & Adapter** — `domain/llm.Completer` は薄く保ち、SDK は infrastructure に閉じる。
6. **失敗時は劣化運転** — Refiner 失敗時は元 Section を保持。Intake はヒューリスティック結果を常に返せる。

### 1.2 スコープ外（App-2 ではやらない）

| 項目 | 理由 | 参照 |
|------|------|------|
| Turnstile / Rate Limit / 一般公開 | Deploy-2 で実施 | [deployment.md](./deployment.md) |
| D1 履歴 | Deploy-2 以降 | 同上 |
| Cloudflare Tunnel | 現行 Fly 直結を維持 | [deployment-session-handoff.md](./deployment-session-handoff.md) |
| MCP サーバー `optimize_prompt` | App-3 | 本ドキュメント §15 |
| Investigate Web / ZIP アップロード | App-3 候補 | [deployment.md](./deployment.md) |
| `openai-go`（OpenAI SDK） | Gemini に統一（合意） | 本ドキュメント §9 |
| Web ユーザ BYOK（Settings に API キー入力） | 秘密情報をブラウザに載せない | — |

---

## 2. 用語定義

| 用語 | 意味 | 例 |
|------|------|-----|
| **TargetProfile** | 最適化**出力**の形式・公式準拠先 | `claude`, `codex`, `cursor` |
| **LLM Provider** | Refiner / Intake が呼ぶ**生成 API** | Gemini, Anthropic |
| **CompletionIntent** | LLM 呼び出しの**意図**（対象・制約・rule_id）。プロンプト文ではない | domain VO |
| **PromptBuilder** | Intent + Profile + Provider から**方言別メッセージ**を組み立てる | infrastructure |
| **Finding** | Analyze で検出された曖昧性の構造化表現 | domain VO |
| **RefinementIntent** | Section 単位の書き換え意図 | domain VO |
| **残差** | 決定的ルール適用後に残った未処理コンテンツ | LLM Refiner の入力 |

### 2.1 TargetProfile と LLM Provider のルーティング（合意）

| TargetProfile | LLM Provider（Refiner / Intake） | 備考 |
|---------------|----------------------------------|------|
| `claude` | **Anthropic** | Claude 向け出力整形とは別に、API は Anthropic |
| `codex` | **Gemini** | 出力は Codex 4 要素。LLM は Gemini |
| `openai` | **Gemini** | 同上 |
| `devin` | **Gemini** | 同上 |
| `cursor` | **Gemini** | 同上 |
| common ルール | **Gemini** | Profile 横断 |

`PromptBuilder` は **TargetProfile 向けの指示文** と **Provider 向けのメッセージ形式** を組み合わせる（SQL の方言とクエリプランの関係）。

---

## 3. アーキテクチャ概要

### 3.1 レイヤー構成（追加分）

```
backend/
├── domain/
│   ├── llm/              # 新規: Completer Port, Intent/Outcome/Budget VO
│   ├── refine/           # 新規: RefinementIntent, RefinementOutcome
│   ├── intake/           # 拡張: Finding, AnalysisReport, MergeFindings
│   ├── prompt/           # 拡張: SectionTypeExamples または Metadata 整備
│   ├── budget/           # 拡張: LLM / Refinement 設定
│   └── optimizer/        # 既存: Stage, Pipeline（変更なし）
├── application/
│   ├── intake/           # 拡張: Heuristic + LLM オーケストレーション
│   └── optimize/         # 拡張: AST Stage + LLMRefiner Stage 挿入
└── infrastructure/
    ├── llm/              # 新規: gemini, anthropic, noop, router, prompts
    ├── ast/                # 新規: goldmark パース, Section マッパー
    ├── stages/             # 拡張: ASTParse, ASTCompress, LLMRefiner
    └── bestpractice/       # 拡張: false ルール読み込み, RulesForRefinement
```

### 3.2 データフロー（App-2 完了時）

```
[React SPA / CLI]
        │
        ▼
[presentation: GraphQL / Connect-RPC / CLI]
        │
        ├─ analyze ──► application/intake
        │                  ├─ HeuristicAnalyzer → []Finding
        │                  ├─ (LLM) Completer via CompletionIntent
        │                  ├─ MergeFindings
        │                  └─ Findings → Questions → AnalyzeResult
        │
        └─ optimize ─► application/optimize
                           ├─ Format Pipeline（既存）
                           └─ Compress Pipeline（拡張）
                                  ├─ NormalizeWhitespace
                                  ├─ ASTParse（goldmark）
                                  ├─ ASTCompress（決定的）
                                  ├─ Deduplicate / Boilerplate / …
                                  ├─ LLMRefiner（automatable: false）
                                  ├─ BudgetAllocate
                                  ├─ TruncateByPriority
                                  └─ AssembleWithProfile
```

### 3.3 「プロンプト = SQL」対応表

| 層 | SQL 世界 | translate-prompt |
|----|---------|------------------|
| 意図・制約 | クエリの意味・WHERE 条件 | `CompletionIntent`, `RefinementIntent`, `Finding` |
| プラン | クエリプランナー | `application/*` UseCase（いつ LLM を呼ぶか） |
| 方言別 SQL 文 | `SELECT …` の具体文 | `PromptBuilder` → Provider 別 `[]Message` |
| 実行 | DB エンジン | `infrastructure/llm` の `Completer` 実装 |
| 結果 | 行セット | `CompletionOutcome`, `RefinementOutcome` |

**禁止**: `domain` パッケージに OpenAI / Anthropic / Gemini の SDK 型や、完成済み system prompt 文字列を置くこと。

---

## 4. ドメインモデル

### 4.1 `domain/llm`

#### `Completer`（Port）

```go
// Completer executes a completion from domain intent. Implementations live in infrastructure.
type Completer interface {
    Complete(ctx context.Context, intent CompletionIntent, budget CompletionBudget) (CompletionOutcome, error)
}
```

#### `CompletionIntent`（VO）

LLM に渡す**意図**。プロンプトテンプレートの完成文ではない。

| フィールド | 型 | 説明 |
|-----------|-----|------|
| `Purpose` | `CompletionPurpose` | `intake_analyze` \| `section_refine` |
| `TargetProfile` | `budget.TargetProfile` | 出力形式の参照 |
| `RuleRef` | `string` | YAML `rule.id` |
| `SectionRef` | `SectionRef` | 対象 Section（index / id / type） |
| `InputContent` | `string` | 対象テキスト（**データ**でありプロンプトではない） |
| `Constraints` | `CompletionConstraints` | 下表 |
| `Context` | `CompletionContext` | Intake 時のヒューリスティック結果等 |

**`CompletionConstraints`**

| フィールド | 説明 |
|-----------|------|
| `MustNotIncreaseTokens` | `true` のとき出力トークン > 入力なら棄却 |
| `PreserveStructure` | XML タグ・見出し階層を維持 |
| `MaxOutputTokens` | 1 回の呼び出し上限 |

**`CompletionContext`**（Intake 用）

| フィールド | 説明 |
|-----------|------|
| `HeuristicFindings` | `[]intake.Finding` — LLM への入力コンテキスト |
| `PromptSections` | 分析対象 Section の要約（全文 or 先頭 N トークン） |

#### `CompletionOutcome`（VO）

| フィールド | 型 | 説明 |
|-----------|-----|------|
| `Content` | `string` | 生成テキスト（Refiner 時は Section 差し替え用） |
| `Findings` | `[]intake.Finding` | Intake 時のみ。LLM が追加した Finding |
| `Usage` | `CompletionUsage` | 入力/出力トークン数、モデル名 |
| `Provider` | `string` | `gemini` \| `anthropic`（監査用） |

#### `CompletionBudget`（VO）

1 リクエスト（1 回の Analyze または 1 回の Optimize）内の LLM 予算。

| フィールド | デフォルト | 説明 |
|-----------|-----------|------|
| `MaxCalls` | `3` | LLM API 呼び出し回数上限 |
| `MaxOutputTokens` | `min(maxTokens*0.3, 4000)` | 合計出力トークン上限 |
| `TimeoutPerCall` | `30s` | 1 呼び出しあたり |

**既存 `budget.Config.MaxTokens`** は**最終プロンプト予算**（Truncate 管轄）であり、LLM 予算とは別層。

#### ドメインエラー

| エラー | 意味 | 扱い |
|--------|------|------|
| `ErrBudgetExceeded` | 呼び出し回数 or トークン超過 | 以降の LLM 呼び出しをスキップ |
| `ErrRefusal` | モデル拒否 | Refiner: 元 Section 保持 / Intake: ヒューリスティックのみ |
| `ErrProviderUnavailable` | キー未設定・API 障害 | `LLM_ENABLED=false` 相当の劣化運転 |

---

### 4.2 `domain/intake`（拡張）

Phase 1 の `Ambiguity` / `Question` は維持しつつ、LLM 統合用の型を追加する。

#### `Finding`（VO）— 新規

| フィールド | 型 | 説明 |
|-----------|-----|------|
| `ID` | `string` | 安定 ID（マージ・dedupe 用） |
| `Category` | `string` | `goal_unclear`, `scope_missing`, `contradiction`, … |
| `Severity` | `int` | 1–5。高いほど質問優先 |
| `SectionRef` | `llm.SectionRef` | 関連 Section |
| `RuleID` | `string` | 紐づく YAML ルール |
| `Summary` | `string` | 検出内容の要約 |
| `Source` | `FindingSource` | `heuristic` \| `llm` |

#### `AnalysisReport`（Entity）— 新規

| フィールド | 説明 |
|-----------|------|
| `Findings` | `[]Finding` |
| `Status` | `needs_input` \| `ready` |

#### `MergeFindings`（ドメインサービス）

```
MergeFindings(heuristic []Finding, llm []Finding) []Finding
```

- 同一 `Category` + 同一 `SectionRef` は dedupe（**Severity が高い方を採用**）。
- `Source=llm` の質問文案は `Summary` に反映し、後段で `Question` 生成。

#### `Question` 生成

**LLM は `Question` を直接返さない。** `Finding` → `Question` 変換は domain または application の決定的ロジック。

```go
func QuestionsFromFindings(findings []Finding) []Question
```

---

### 4.3 `domain/refine`（新規）

#### `RefinementIntent`（VO）

| フィールド | 説明 |
|-----------|------|
| `SectionRef` | 対象 Section |
| `RuleRef` | `automatable: false` ルール ID |
| `InputContent` | 原文 |
| `Constraints` | `CompletionConstraints` と同型または埋め込み |

#### `RefinementOutcome`（VO）

| フィールド | 説明 |
|-----------|------|
| `Content` | 書き換え後テキスト |
| `Applied` | `true` = 採用 / `false` = 棄却（トークン増等） |
| `RejectReason` | `token_increase`, `parse_error`, `budget_exceeded`, … |

---

### 4.4 `domain/prompt`（拡張）

`common-example-summarize` は `examples` Section を対象とする。現行 `ParseSections` では XML タグ `examples` は `SectionTypeCode` にマップされる（[parse.go](../backend/infrastructure/stages/parse.go)）。

**App-2 での対応（合意）:**

1. `Section.Metadata["xml_tag"]` に元タグ名（`examples` 等）を格納する。
2. ルール `condition.section_tag: examples` でマッチさせる。
3. （任意・後続）`SectionTypeExamples` を追加し、予算配分を独立させる。

AST 圧縮は `Section` 単位で動作し、goldmark は **Section.Content 内部** をパースする。

---

### 4.5 `domain/budget.Config`（拡張）

| フィールド | 型 | デフォルト | 説明 |
|-----------|-----|-----------|------|
| `LLMEnabled` | `bool` | `false` | LLM 呼び出しのマスタースイッチ |
| `LLMMaxCalls` | `int` | `3` | 1 リクエストあたり |
| `LLMModelGemini` | `string` | `gemini-2.5-flash` | 上書き可 |
| `LLMModelAnthropic` | `string` | `claude-sonnet-5` 等 | 上書き可 |

環境変数からの読み込みは **infrastructure/config** の責務。`domain/budget` は値の型のみ。

---

## 5. Intake LLM 設計

### 5.1 フロー（合意: ヒューリスティック常時 + LLM 補完）

```
1. HeuristicAnalyzer.Analyze(prompt, config) → []Finding   // 常に実行
2. if !config.LLMEnabled → goto 4
3. Completer.Complete(CompletionIntent{
       Purpose: intake_analyze,
       Context: { HeuristicFindings, PromptSections },
   }) → CompletionOutcome.Findings
4. merged := MergeFindings(heuristic, llmFindings)
5. questions := QuestionsFromFindings(merged)
6. status := ready if len(questions)==0 else needs_input
```

### 5.2 スイッチ

| スイッチ | 効果 |
|---------|------|
| `LLM_ENABLED=false` | ステップ 3 スキップ。Phase 1 同等の Analyze |
| `deep_dive=true`（既存） | ヒューリスティックの網羅性を維持したうえで LLM 補完を実行（`LLM_ENABLED` も必要） |
| `deep_dive=false` | ヒューリスティックのみ（LLM 呼ばない） |

### 5.3 GraphQL / API への影響

- `analyze` レスポンスに `findings` フィールド追加（任意・デバッグ用）。
- `questions` は従来どおり。生成元が `Finding` 経由になる。
- `OptimizeReport` とは独立。

---

## 6. Refiner（LLM Stage）設計

### 6.1 パイプライン挿入位置

**Compress パイプライン内、`TruncateByPriority` の直前。**

理由:

- Format 完了後の `Section` 構造を入力にできる。
- 予算超過が見えたあと、Truncate で切り落とす前に意味保持圧縮できる。
- Format 段階での LLM 書き換えはプロファイル再整形リスクが高い。

### 6.2 `LLMRefinerStage`

```go
type LLMRefinerStage struct {
    Completer llm.Completer
    Loader    *bestpractice.Loader
    Counter   optimizer.TokenCounter
}
```

処理:

1. `TargetProfile` から `RulesForRefinement()`（`automatable: false`, `pipeline: compress`, `stage: LLMRefiner`）を取得。
2. 各ルールの `condition` を評価（`section_tag`, `remaining_tokens_over_budget`, …）。
3. 対象 Section ごとに `RefinementIntent` → `Completer.Complete`。
4. `MustNotIncreaseTokens` 違反時は `RefinementOutcome.Applied=false`、元 Section 保持。
5. `AppliedRule` に `method: "llm"`, `model`, `rule_id` を記録。

### 6.3 初回 `automatable: false` ルール（合意: 3 個）

既存 `automatable: true` ルールは **pattern マッチ分のみ** 決定的に実行。語義変換が必要な残差を false ルールに委譲する。

#### 6.3.1 `common-example-summarize`（新規 / common.yaml）

| 項目 | 値 |
|------|-----|
| 説明 | `examples` Section の冗長な例を意味保持で要約 |
| pipeline | `compress` |
| stage | `LLMRefiner` |
| action | `summarize_preserve_meaning` |
| condition | `section_tag: examples` かつ予算逼迫時 |
| constraints | `must_not_increase_tokens: true` |
| intake_on_failure | `false` |

#### 6.3.2 `cursor-actionable-semantic`（新規 / cursor.yaml）

| 項目 | 値 |
|------|-----|
| 説明 | 「きれいに」「best practices」等を具体的制約に言い換え |
| 由来 | 既存 `cursor-actionable`（`automatable: true`）の **残差** |
| condition | `section_tag: rules` かつ曖昧パターン検出済み |
| action | `make_actionable_semantic` |
| intake_on_failure | `false` |

既存 `cursor-actionable` は patterns マッチ時の**軽量置換**のみ継続。

#### 6.3.3 `claude-explicit-semantic`（新規 / claude.yaml）

| 項目 | 値 |
|------|-----|
| 説明 | 依頼口調・曖昧動詞の命令形化（regex 非該当分） |
| 由来 | 既存 `claude-explicit`（`automatable: true`）の **残差** |
| condition | `section_type: task` かつ pattern ルール適用後 |
| action | `rewrite_imperative_semantic` |
| intake_on_failure | `false` |

### 6.4 YAML スキーマ拡張（false ルール）

```yaml
rules:
  - id: common-example-summarize
    description: examples Section の意味保持要約
    source_url: https://platform.claude.com/docs/...
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
      # プロンプトテンプレートは infrastructure/bestpractice が rule_id で解決
```

`domain/bestpractice.Rule` への追加フィールド:

| YAML キー | Go フィールド | 説明 |
|-----------|--------------|------|
| `constraints` | `Constraints map[string]bool` | ドメイン制約 |
| `llm.max_output_tokens` | `LLMMaxOutputTokens int` | ルール単位上限 |
| `condition.section_tag` | `Condition["section_tag"]` | Metadata マッチ |

`TargetProfile` に追加メソッド:

```go
func (tp *TargetProfile) RulesForRefinement() []Rule  // automatable==false && stage==LLMRefiner
```

---

## 7. AST 圧縮設計（goldmark）

App-2 に **フル実装** を含める（合意）。LLM Refiner と併用する。

### 7.1 責務分担

| 処理 | 担当 | 決定的 / LLM |
|------|------|-------------|
| リスト項目の重複マージ | `ASTCompress` | 決定的 |
| ネスト見出しの浅化 | `ASTCompress` | 決定的 |
| CodeBlock 内コメント除去 | `ASTCompress` | 決定的（言語ヒューリスティック） |
| Example 段落の意味要約 | `LLMRefiner`（`common-example-summarize`） | LLM |
| 曖昧文言の具体化 | `LLMRefiner` | LLM |

### 7.2 Stage

#### `ASTParseStage`

- 入力: `prompt.Section`（Content が Markdown 想定）
- goldmark で AST 化 → 内部 `domain/ast` ノードツリー（または infrastructure 内一時表現）
- XML タグのみの Section はスキップ（Format 済み構造を壊さない）

#### `ASTCompressStage`

- リスト重複除去、連続空行 collapse、コードコメント strip
- 出力を `Section.Content` に書き戻し
- `AppliedRule` に `method: "ast"` を記録

### 7.3 `domain/ast`（最小）

| 型 | 説明 |
|----|------|
| `Node` | interface（`Paragraph`, `List`, `CodeBlock`, `Heading`） |
| `Document` | `[]Node` |

**goldmark 型は infrastructure に閉じる。** `domain/ast` は独自の軽量ノード。

### 7.4 Compress 内の順序（確定）

```
NormalizeWhitespace
→ ASTParse
→ ASTCompress
→ DeduplicateExact
→ RemoveBoilerplate / …
→ LLMRefiner
→ BudgetAllocate
→ TruncateByPriority
→ EnforceRulesBudget
→ AssembleWithProfile
```

AST で機械的に削ったあと LLM で意味圧縮する順。

---

## 8. infrastructure/llm

### 8.1 パッケージ構成

```
infrastructure/llm/
  completer.go      # Completer を満たすファサード
  router.go         # TargetProfile → gemini | anthropic
  gemini.go         # google.golang.org/genai
  anthropic.go      # anthropic-sdk-go
  noop.go           # テスト用固定応答
  prompts.go        # PromptBuilder: Intent → []ProviderMessage
  config.go         # 環境変数読み込み（API キーはここだけ）
```

### 8.2 依存（合意・セキュア選定済み）

| モジュール | バージョン目安 | 用途 |
|-----------|--------------|------|
| `google.golang.org/genai` | 最新 GA（≥ v1.62 系） | Gemini |
| `github.com/anthropics/anthropic-sdk-go` | **≥ v1.55.1** | Anthropic（CVE 修正済み版） |
| `github.com/yuin/goldmark` | **≥ v1.8.2** | AST（HTML renderer は使わない） |

**不採用:**

| モジュール | 理由 |
|-----------|------|
| `openai-go` | Gemini に統一（合意） |
| `github.com/flexigpt/inference-go` | 実績薄。Port は自前で十分 |
| `github.com/google/generative-ai-go` | レガシー。`genai` に移行 |

### 8.3 環境変数

キー取得・ローカル / Fly 設定の手順: **[llm-setup.md](./llm-setup.md)**

| 変数 | 必須 | 説明 |
|------|------|------|
| `LLM_ENABLED` | — | `true` で LLM 有効。デフォルト `false` |
| `GOOGLE_API_KEY` | Gemini 使用時 | Gemini API キー（推奨名） |
| `GEMINI_API_KEY` | Gemini 使用時 | 上記の別名（どちらか一方で可） |
| `ANTHROPIC_API_KEY` | `claude` Profile + LLM 時 | Anthropic |
| `LLM_DEFAULT_MAX_CALLS` | — | デフォルト `3` |
| `LLM_GEMINI_MODEL` | — | 未設定時 `gemini-2.5-flash` |
| `LLM_ANTHROPIC_MODEL` | — | 未設定時は実装で sane default |

### 8.4 実行場所（合意）

| 実行場所 | API キー | 備考 |
|---------|---------|------|
| **CLI** | ローカル環境変数（BYOK） | `translate-prompt --deep-dive` 等 |
| **Web（招待制 β）** | Fly Secrets | `LLM_ENABLED=true` で有効化 |
| **ローカル `make serve`** | ローカル env | Web 相当の動作確認 |

`domain` / `application` はキーを知らない。

---

## 9. テスト戦略

### 9.1 単体

| 対象 | 方針 |
|------|------|
| `MergeFindings` | table-driven、dedupe / severity |
| `QuestionsFromFindings` | 決定的変換 |
| `LLMRefinerStage` | `noop.Completer` + 固定 `CompletionOutcome` |
| `ASTCompressStage` | `testdata/ast/*.md` ゴールデン |
| `PromptBuilder` | スナップショット（Provider 別メッセージ構造） |

### 9.2 統合

| 対象 | 方針 | 実装 |
|------|------|------|
| `Optimize` with LLM | noop または VCR 録画リプレイ | `application/optimize/integration_test.go` |
| `Analyze` with LLM | ヒューリスティック only / LLM 補完の両 fixture | `application/intake/integration_test.go` |
| CI | **LLM API を叩かない**（デフォルト）。手動または nightly で実 API | noop Completer のみ |

### 9.3 成功基準（App-2）

- [x] `LLM_ENABLED=false` で Phase 1 と同等の回帰（既存テスト PASS）
- [x] 3 つの false ルールが noop で `AppliedRule.method=llm` を記録
- [x] Intake: ヒューリスティック only / LLM 補完の両方で `Status` が決定的に検証可能
- [x] AST: 代表 Markdown でリスト重複・コメント除去がゴールデン一致
- [x] `make test` / `make lint` 通過
- [x] 秘密情報がログ・レポートに出力されない

---

## 10. セキュリティ

| 項目 | 対策 |
|------|------|
| API キー | Fly Secrets / ローカル env のみ。コード・ログ・`OptimizeReport` に含めない |
| プロンプト本文 | 外部 LLM API に送信される前提で利用者に明示（README） |
| リクエストサイズ | 既存 Go サーバ + 将来 Workers で上限（Deploy-2） |
| `LLM_ENABLED` デフォルト false | 意図しない API 課金を防止 |
| goldmark XSS | `renderer/html` は使用しない（AST パースのみ） |
| Web Investigate | 引き続き無効（`INVESTIGATE_ENABLED=false`） |

---

## 11. 実装チェックリスト

完了したら `[x]` に更新する。

### 11.0 ドメイン層

- [x] `domain/llm` — `Completer`, `CompletionIntent`, `CompletionOutcome`, `CompletionBudget`, errors
- [x] `domain/refine` — `RefinementIntent`, `RefinementOutcome`
- [x] `domain/intake` — `Finding`, `AnalysisReport`, `MergeFindings`, `QuestionsFromFindings`
- [x] `domain/budget` — `LLMEnabled`, `LLMMaxCalls`, モデル名フィールド
- [x] `domain/bestpractice` — `RulesForRefinement`, Rule 拡張フィールド
- [x] `domain/prompt` — `Section.Metadata` 整備（`xml_tag`）
- [x] `domain/ast` — 最小 Node 型

### 11.1 インフラ層

- [x] `infrastructure/llm` — gemini, anthropic, noop, router, prompts, config
- [x] `infrastructure/ast` — goldmark パース, Section マッパー
- [x] `infrastructure/stages` — `ASTParse`, `ASTCompress`, `LLMRefiner`
- [x] `infrastructure/stages/parse.go` — `Metadata["xml_tag"]` 設定
- [x] `infrastructure/config` — LLM 環境変数
- [x] `docs/best-practices/*.yaml` + embed ルール — false ルール 3 個追加

### 11.2 アプリケーション層

- [x] `application/intake` — Heuristic + LLM + Merge オーケストレーション
- [x] `application/optimize` — Compress パイプライン順序更新

### 11.3 Presentation

- [x] GraphQL スキーマ — `findings`（任意）
- [x] `OptimizeReport` — LLM / AST 適用メタデータ
- [x] CLI — `--llm` / env 連携

### 11.4 テスト・ドキュメント

- [x] `testdata/ast/` ゴールデン
- [x] noop Completer 契約テスト
- [x] [intake.md](./intake.md) Phase 2 節（本書へのリンク）
- [x] [best-practices/README.md](./best-practices/README.md) スキーマ更新

---

## 12. 実装順序（合意）

```
① domain（llm, refine, intake 拡張, ast, budget）
② infrastructure/llm（noop 先行 → gemini → anthropic → PromptBuilder）
③ application/intake（LLM 統合）+ テスト
④ infrastructure/stages LLMRefiner + false ルール YAML + テスト
⑤ infrastructure/ast + AST Stages + テスト
⑥ presentation / CLI / レポート拡張
⑦ 実 API 手動検証（任意）
```

各ステップで `make test` を維持する。**TDD**: Port と domain サービスから着手。

---

## 13. Phase 3 との境界（App-3 プレビュー）

App-2 完了後の次フェーズ。本ドキュメントでは設計のみ記載し、実装は行わない。

| 項目 | 内容 |
|------|------|
| MCP サーバー | `optimize_prompt` / `analyze_prompt` / `estimate_tokens` |
| 実装方針 | `application/*` をそのまま呼ぶ `presentation/mcp` 薄ラッパー |
| トランスポート | stdio 先行、SSE は後続 |
| Deploy-2 | Turnstile, Rate Limit, 一般公開 |
| SaaS | Hono BFF, D1, 認証, 課金 |

App-2 で固定する型（`CompletionIntent`, `OptimizeReport`）は MCP の入出力契約の土台になる。

---

## 14. 関連ドキュメント

| ドキュメント | 関係 |
|-------------|------|
| [architecture.md](./architecture.md) | DDD レイヤー全体 |
| [implementation-roadmap.md](./implementation-roadmap.md) | Phase 1 完了チェックリスト |
| [intake.md](./intake.md) | Intake フロー（Phase 2 節を追加） |
| [best-practices/README.md](./best-practices/README.md) | ルール YAML SSOT |
| [api.md](./api.md) | GraphQL / Connect-RPC |
| [deployment.md](./deployment.md) | Deploy-2（本フェーズ外） |
| [deployment-session-handoff.md](./deployment-session-handoff.md) | 現行本番スナップショット |

---

## 15. 合意サマリ（一覧）

| # | 決定事項 | 結論 |
|---|---------|------|
| 1 | App-2 スコープ | LLM 基盤 + Intake LLM + Refiner + AST。Deploy は後回し |
| 2 | LLM 投資 | 共有 Port + Intake / Refiner 両方 |
| 3 | LLM Port | 薄い `domain/llm.Completer` |
| 4 | プロンプト | infrastructure `PromptBuilder`（プロンプト = SQL） |
| 5 | ドメインモデル | `llm` + `intake` 拡張 + `refine` + `ast` |
| 6 | Intake | ヒューリスティック常時 + LLM 補完 + MergeFindings |
| 7 | Refiner 位置 | Compress 内・Truncate 直前 |
| 8 | false ルール | `common-example-summarize`, `cursor-actionable-semantic`, `claude-explicit-semantic` |
| 9 | Provider | Gemini デフォルト + Anthropic は `claude` のみ |
| 10 | OpenAI SDK | 不採用 |
| 11 | キー | CLI BYOK + Fly Secrets、`LLM_ENABLED` デフォルト false |
| 12 | AST | goldmark を App-2 に含める |
| 13 | 実装順 | domain → llm infra → intake → refiner → ast → presentation |

**合意日**: 2026-07-03  
**合意方法**: `/dotfiles-plan-first` + `/grill-me`
