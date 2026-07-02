# API / CLI 仕様

## GraphQL API（gqlgen）

ベース URL: `http://127.0.0.1:8080/query`

開発時 Playground: `http://127.0.0.1:8080/playground`（`ENV=dev`）

### Query

```graphql
query {
  health { status }
  estimate(text: "...", tokenizer: "cl100k_base") { tokens }
}
```

### Mutation

```graphql
mutation {
  analyze(input: {
    prompt: "..."
    config: { targetProfile: CODEX, maxTokens: 8000, tokenizer: "cl100k_base", deepDive: true }
  }) {
    status
    questions { id text ruleId }
    prompt
  }

  investigate(input: {
    workspacePath: "/path/to/repo"
    targetProfile: CURSOR
  }) {
    files { path sectionType contentPreview }
    suggestedCommands
  }

  optimize(input: {
    prompt: "..."
    config: { targetProfile: CODEX, maxTokens: 8000, tokenizer: "cl100k_base" }
    answers: { goal: "テストが通ること" }
  }) {
    optimizedPrompt
    report { inputTokens outputTokens reductionPercent appliedRules { id sourceUrl } }
    artifacts { cursorMdcSuggestions { filename content } }
  }
}
```

スキーマ: `backend/graph/schema.graphqls`

---

## Connect-RPC API（connect-go）

サービス: `translate_prompt.v1.TranslatePromptService`

Protobuf: `backend/proto/translate_prompt/v1/service.proto`

| RPC | 説明 |
|-----|------|
| `Health` | ヘルスチェック |
| `Analyze` | 曖昧性分析 |
| `Investigate` | ワークスペース調査 |
| `Optimize` | 最適化実行 |
| `Estimate` | トークン推定 |

フロントエンドは `@connectrpc/connect-web` の生成クライアント（`frontend/src/gen/`）を使用。

---

## CLI

バイナリ: `translate-prompt`（`backend/cmd/translate-prompt`）

```bash
cat prompt.md | ./bin/translate-prompt --max-tokens 8000 --target-profile codex
./bin/translate-prompt -i input.md -o output.md --max-tokens 12000 --report json
```

## Web サーバ

バイナリ: `server`（`backend/cmd/server`）

```bash
make build-server && ENV=dev ./bin/server --port 8080
```

## コード生成

```bash
make codegen   # gqlgen + buf generate
```
