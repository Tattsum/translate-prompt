# 深堀り（Intake）フロー

plan-first / grill-me 相当。いきなり圧縮せず、不足情報を補完してから最適化する。

## 状態遷移

```
Analyze → AskQuestions（曖昧）→ MergeAnswers → Analyze
Analyze → Investigate（要コンテキスト）→ MergeContext → Analyze
Analyze → Optimize（十分）→ 完了
```

## AmbiguityAnalyzer（Phase 1: ヒューリスティック）

| 検出パターン | 質問例 |
|-------------|--------|
| 目的不明 | 成功条件・完了の定義は？ |
| スコープ未指定 | 対象ファイル・モジュールの範囲は？ |
| 矛盾する要求 | どちらを優先しますか？ |
| 受け入れ基準なし | 完了条件・検証コマンドは？ |
| トークン予算未指定 | max-tokens・TargetProfile は？ |
| Devin: スコープ過大 | 1 セッション 1 PR に分割しますか？ |
| Devin: Scope out 未指定 | 触ってはいけない領域は？ |
| Cursor: alwaysApply 過多 | どの規約を常時適用に残すか？ |
| Cursor: Rules vs Skills | 手順は Rule と Skill のどちら？ |

## WorkspaceInvestigator

### 読み取り対象

- `README.md`, `CONTEXT.md`, `go.mod`, `package.json`
- `.cursor/rules/`, `.cursor/skills/`, `AGENTS.md`
- ディレクトリ構成（深さ 2）

### 制限

- ファイル数 20、合計 100KB
- `.env` 等の秘密ファイルはスキップ

### マージ先

調査結果は `Code` / `Rules` Section としてプロンプトに統合。

## IntakeUseCase 戻り値

| Status | 意味 |
|--------|------|
| `StatusNeedsInput` | 質問リストを返し、回答待ち |
| `StatusReady` | OptimizeUseCase へ進む |

## CLI / Web

- CLI: `--deep-dive`, `--workspace <path>`
- Web: Settings の深堀り ON/OFF、`/intake` ページ

---

## App-2: LLM 統合（合意）

設計の正: [phase2-app-roadmap.md](./phase2-app-roadmap.md)

### フロー

```
HeuristicAnalyzer → []Finding（常時）
    ↓ LLM_ENABLED && deep_dive
Completer（intake_analyze）→ 追加 []Finding
    ↓
MergeFindings → QuestionsFromFindings → AnalyzeResult
```

### Phase 1 との差分

| 項目 | Phase 1 | App-2 |
|------|---------|-------|
| 検出 | `Ambiguity` / キーワード | `Finding`（`Source`: heuristic \| llm） |
| LLM | なし | ヒューリスティック結果をコンテキストに補完 |
| 質問生成 | ヒューリスティック直結 | `Finding` 経由（LLM は Question を直接返さない） |

### スイッチ

| 条件 | 動作 |
|------|------|
| `LLM_ENABLED=false`（デフォルト） | Phase 1 同等 |
| `deep_dive=false` | ヒューリスティックのみ |
| `deep_dive=true` + `LLM_ENABLED=true` | ヒューリスティック + LLM 補完 |

### GraphQL（デバッグ）

`analyze` レスポンスの `findings` は `deep_dive=true` 時にマージ済み `Finding` 一覧を返す（質問生成の根拠確認用）。`questions` は従来どおり UI 向け。

```graphql
mutation {
  analyze(input: { prompt: "...", config: { deepDive: true, ... } }) {
    status
    questions { id text }
    findings { id category severity source summary }
  }
}
```
