# LLM API キー取得・環境設定

App-2 で追加された Intake LLM 補完・Optimize LLM Refiner を有効化するための手順。  
設計の詳細は [phase2-app-roadmap.md](./phase2-app-roadmap.md) を正とする。

## 必要なキー（TargetProfile 別）

| TargetProfile | 呼び出し先 API | 必要な環境変数 |
|---------------|----------------|----------------|
| `codex`, `openai`, `devin`, `cursor` | Google Gemini | `GOOGLE_API_KEY` または `GEMINI_API_KEY` |
| `claude` | Anthropic | `ANTHROPIC_API_KEY` |

ルーティング実装: `backend/infrastructure/llm/router.go`

どちらのプロファイルも使う場合は **両方のキー** を設定する。

---

## 1. Google Gemini API キー（`GOOGLE_API_KEY` / `GEMINI_API_KEY`）

### 取得手順

1. Google アカウントで [Google AI Studio](https://aistudio.google.com/apikey) を開く
2. **Create API key**（または既存プロジェクトのキー作成）を実行
3. 表示されたキーをコピー（再表示できない場合があるため、安全な場所に保存）

公式ドキュメント: [Gemini API quickstart](https://ai.google.dev/gemini-api/docs/quickstart)

### 注意

- 無料枠・課金上限は Google Cloud / AI Studio のプロジェクト設定で確認する
- キーはリポジトリにコミットしない（`.gitignore` 済みの `.env` や Fly Secrets を使う）
- 本ツールは `GOOGLE_API_KEY` と `GEMINI_API_KEY` の **どちらでも** 読み込む（`backend/infrastructure/llm/config.go`）

---

## 2. Anthropic API キー（`ANTHROPIC_API_KEY`）

### 取得手順

1. [Anthropic Console](https://console.anthropic.com/) にサインアップ / ログイン
2. **Settings → API keys**（直接: [API keys](https://console.anthropic.com/settings/keys)）を開く
3. **Create Key** でキーを発行しコピー

公式ドキュメント: [Anthropic API — Getting started](https://docs.anthropic.com/en/api/getting-started)

### 注意

- `claude` TargetProfile の Intake / Refiner 経路でのみ Anthropic が選ばれる
- 利用上限・請求は Console の Billing で管理する

---

## 3. 環境変数一覧

| 変数 | 必須 | デフォルト | 説明 |
|------|------|-----------|------|
| `LLM_ENABLED` | — | `false` | `true` / `1` でサーバー全体の LLM を有効化 |
| `GOOGLE_API_KEY` | Gemini 利用時 | — | Gemini API キー（推奨名） |
| `GEMINI_API_KEY` | Gemini 利用時 | — | 上記の別名（どちらか一方で可） |
| `ANTHROPIC_API_KEY` | `claude` + LLM 時 | — | Anthropic API キー |
| `LLM_DEFAULT_MAX_CALLS` | — | `3` | 1 リクエストあたりの LLM 呼び出し上限 |
| `LLM_GEMINI_MODEL` | — | `gemini-2.5-flash` | Gemini モデル名 |
| `LLM_ANTHROPIC_MODEL` | — | `claude-sonnet-4-20250514` | Anthropic モデル名 |

読み込み実装: `backend/infrastructure/llm/config.go`  
サーバー起動時のマスタースイッチ: `backend/infrastructure/config/server.go`（`LLM_ENABLED`）

---

## 4. ローカル開発

### 4.1 サーバー（`make serve`）

リポジトリ直下に `.env` を置き、シェルで export するか `direnv` 等を使う。

```bash
# .env.example をコピーして編集
cp .env.example .env

export LLM_ENABLED=true
export GOOGLE_API_KEY="your-gemini-key"
# export ANTHROPIC_API_KEY="your-anthropic-key"  # claude 利用時

make serve
```

### 4.2 CLI

```bash
export LLM_ENABLED=true
export GOOGLE_API_KEY="your-gemini-key"

# Intake の LLM 補完には --deep-dive も必要
cat prompt.md | ./bin/translate-prompt \
  --max-tokens 8000 \
  --target-profile codex \
  --deep-dive \
  --llm
```

`--llm` は `LLM_ENABLED` 環境変数と同等（`backend/presentation/cli/cli.go`）。

### 4.3 Intake で LLM が動く条件

| 経路 | 条件 |
|------|------|
| Web / GraphQL / Connect-RPC | サーバーで `LLM_ENABLED=true` **かつ** リクエストの `deepDive: true` |
| CLI | `--deep-dive` **かつ** `--llm` または `LLM_ENABLED=true` |

Optimize の LLM Refiner は `LLM_ENABLED=true`（と対象ルール・セクションのマッチ）で動く。

---

## 5. 本番（Fly.io Secrets）

[deployment.md](./deployment.md) の API サーバに Secrets を設定する。

```bash
fly secrets set \
  LLM_ENABLED=true \
  GOOGLE_API_KEY="your-gemini-key" \
  ANTHROPIC_API_KEY="your-anthropic-key"
```

任意の上書き:

```bash
fly secrets set LLM_DEFAULT_MAX_CALLS=3 LLM_GEMINI_MODEL=gemini-2.5-flash
```

設定後、アプリは再起動される。動作確認は [deployment-smoke-test スクリプト](../scripts/deployment-smoke-test.sh) に Analyze/Optimize を足すか、手動で GraphQL / UI から確認する。

---

## 6. 動作確認の目安

| 確認 | 期待 |
|------|------|
| `LLM_ENABLED=false` | Phase 1 同等（ヒューリスティックのみ、API 課金なし） |
| Analyze + `deepDive` + キーあり | レスポンス `findings` に `source: llm` が混ざることがある |
| Optimize（cursor 等） | レポートの `applied_rules` に `method: llm` が付くことがある |
| キー未設定 | 劣化運転（エラーにせずヒューリスティック / ルールのみ） |

CI では LLM API を叩かない（noop / 統合テストのみ）。

---

## 7. セキュリティ

- API キーを Git・Issue・ログに載せない
- Web UI から BYOK（ブラウザにキー入力）は **スコープ外**（[phase2-app-roadmap.md](./phase2-app-roadmap.md) §1.2）
- プロンプト本文は外部 LLM に送信される前提で利用する

---

## 関連ドキュメント

- [phase2-app-roadmap.md](./phase2-app-roadmap.md) — App-2 設計・ルーティング合意
- [intake.md](./intake.md) — Analyze / `deep_dive` / `findings`
- [deployment.md](./deployment.md) — Fly / Cloudflare 本番構成
