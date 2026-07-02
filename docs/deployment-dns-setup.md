# DNS / Workers Route セットアップ（SPA）

`translate.tattsum.com` を **Pages カスタムドメイン UI ではなく** Workers 経由で公開する。  
`prompt-api.tattsum.com` と同型（Workers Route + プロキシ ON）。

## 背景

Cloudflare Pages のカスタムドメイン追加で `*.tattsum.com` が **banned** となる。  
`blog`（`tattsum.com` apex）と同様、Workers で配信する。

## アーキテクチャ

```
translate.tattsum.com（DNS: A 192.0.2.1, プロキシ ON）
    ↓ Workers Route（wrangler.web.toml で CI デプロイ時に登録）
Worker「translate-prompt-web」
    ↓ リバースプロキシ（Host: translate-prompt.pages.dev）
https://translate-prompt.pages.dev（CI deploy-pages で更新）
```

## 手順

### 1. SPA プロキシ Worker をデプロイ

main ブランチ push で CI の `deploy-workers` ジョブが `wrangler.web.toml` をデプロイする。  
手動の場合:

```bash
cd workers
pnpm install
pnpm exec wrangler deploy --config ../wrangler.web.toml
# CLOUDFLARE_API_TOKEN / CLOUDFLARE_ACCOUNT_ID が Kurohari35 アカウント向けであること
```

デプロイ後、Zero Trust / Workers で `translate-prompt-web` が存在し、  
ルート `translate.tattsum.com/*` が付与されていることを確認。

### 2. DNS レコードを整理

Cloudflare ダッシュボード → **tattsum.com** → **DNS**

**削除または置き換え**（旧設定・522 の原因）:

| 名前 | 旧タイプ | 旧内容 | 問題 |
|------|---------|--------|------|
| `translate` | CNAME | `translate-prompt-2el.pages.dev` 等 | Pages カスタムドメイン未登録で 522 |

**新規設定**（`prompt-api` と同型）:

| 名前 | タイプ | 内容 | プロキシ |
|------|--------|------|---------|
| `translate` | **A** | `192.0.2.1` | **ON**（オレンジ雲） |

> `192.0.2.1` は Cloudflare Workers 向けのプレースホルダー A レコード（RFC 5737 TEST-NET-1）。  
> 実トラフィックは Workers Route で処理される。

**代替（blog apex と同型）**: DNS タイプ **Worker** → `translate-prompt-web`  
サブドメインでも利用可能な場合はこちらでも可。本リポジトリは `prompt-api` に合わせ **A + Route** を推奨。

### 3. Pages カスタムドメイン UI は使わない

Workers & Pages → `translate-prompt` → Custom domains に  
`translate.tattsum.com` を**追加しない**（banned エラーになる）。

### 4. 疎通確認

```bash
# Worker デプロイ + DNS 反映後
curl -sI https://translate.tattsum.com/ | head -5
# 期待: HTTP/2 200（Access 未設定時）または 302（Access 設定後）

./scripts/deployment-smoke-test.sh
```

## 参考: 既存レコード

| 名前 | タイプ | 用途 |
|------|--------|------|
| `prompt-api` | A `192.0.2.1` | API Worker `translate-prompt-api-proxy` |
| `translate` | A `192.0.2.1` | SPA Worker `translate-prompt-web` |
| `tattsum.com` | Worker `blog` | ブログ（OpenNext） |

## 関連ファイル

| ファイル | 役割 |
|---------|------|
| `wrangler.web.toml` | Worker 名・ルート・`PAGES_HOST` |
| `workers/src/web.ts` | Pages へのリバースプロキシ |
| `.github/workflows/deploy.yml` | `deploy-workers` 内で SPA Worker デプロイ |
