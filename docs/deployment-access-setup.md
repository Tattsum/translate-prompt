# Cloudflare Access セットアップ（招待制 β）

手動設定。Zero Trust ダッシュボードで実施する。  
対象アカウント: **Kurohari35@gmail.com's Account**（`tattsum.com` ゾーン）

## 現状（2026-07-03）

| 項目 | 状態 |
|------|------|
| DNS / Workers | 完了（`translate.tattsum.com` → 200） |
| API 疎通 | 完了（Estimate / Analyze / Optimize 動作確認済み） |
| **Cloudflare Access** | **未設定**（現状は認証なしでアクセス可能） |

Access 設定後は、未認証アクセスで `302` リダイレクト（ログイン画面）が返るようになる。

## 完了チェックリスト

設定が終わったら以下を確認する。

- [ ] IdP（Google / GitHub 等）が追加済み
- [ ] Application `translate-prompt-spa`（`translate.tattsum.com`）作成済み
- [ ] Application `translate-prompt-api`（`prompt-api.tattsum.com`）作成済み
- [ ] 両アプリに `invite-only` ポリシー適用済み
- [ ] シークレットウィンドウで SPA → Access ログイン画面
- [ ] ログイン後 Input ページ表示
- [ ] Analyze → Optimize が動作
- [ ] `./scripts/deployment-smoke-test.sh` で SPA が `302`（Access リダイレクト）を返す

## 前提

- DNS が Cloudflare プロキシ（オレンジ雲）ON であること
- SPA: `translate.tattsum.com` → Worker `translate-prompt-web`（[deployment-dns-setup.md](./deployment-dns-setup.md)）
- API: `prompt-api.tattsum.com` → Worker `translate-prompt-api-proxy`（設定済み）

## 1. Identity Provider（初回のみ）

1. [Cloudflare Zero Trust](https://one.dash.cloudflare.com/) を開く
2. 右上のアカウントが **Kurohari35@gmail.com's Account** であることを確認
3. **Settings** → **Authentication**
4. 利用する IdP を追加（推奨: **Google** または **GitHub**）
5. テストログインが成功することを確認

> ローカルの `wrangler whoami` は別アカウントの場合がある。Access 設定は Zero Trust ダッシュボードで行う。

## 2. Application 1: SPA

1. **Access** → **Applications** → **Add an application**
2. **Self-hosted** を選択
3. 設定:

| 項目 | 値 |
|------|-----|
| Application name | `translate-prompt-spa` |
| Session Duration | `24 hours`（β 向け。任意で変更可） |
| Application domain | `translate.tattsum.com` |
| Accept all available identity providers | オン（または使用 IdP のみ） |

4. **Next** → Policy を追加:

| 項目 | 値 |
|------|-----|
| Policy name | `invite-only` |
| Action | **Allow** |
| Include | **Emails** で自分のメールを追加（β 初期は個別招待が安全） |

> β 初期は `Emails ending in @...` より **Emails** で個別列挙を推奨。意図しない公開を防げる。

5. **Save application**

**確認:** シークレットウィンドウで `https://translate.tattsum.com` を開き、Access ログイン画面が表示されること。

## 3. Application 2: API

1. 同様に **Add an application** → **Self-hosted**
2. 設定:

| 項目 | 値 |
|------|-----|
| Application name | `translate-prompt-api` |
| Session Duration | `24 hours` |
| Application domain | `prompt-api.tattsum.com` |

3. Policy は SPA と**同じ** `invite-only` 条件を適用（Include のメール一覧を揃える）
4. **Save application**

**確認:** ログイン後、ブラウザ DevTools の Network で `prompt-api.tattsum.com/query` が 200 を返すこと。

## 4. 動作確認

```bash
./scripts/deployment-smoke-test.sh
```

ブラウザ:

1. シークレットウィンドウで `https://translate.tattsum.com` を開く → Access ログイン画面
2. ログイン後 Input ページが表示される
3. Analyze → Optimize が動作する

API はブラウザから `fetch` するため、SPA と同じ Access セッションで `prompt-api.tattsum.com` も保護される。  
Workers API プロキシは Access JWT をオリジン（Fly）に転送**しない**（`workers/src/index.ts`）。

## 5. 招待者の追加

β 招待時は Policy の **Include** にメールアドレスを追加するか、  
メールドメインルール（`@company.com`）で許可する。

## トラブルシューティング

| 症状 | 確認 |
|------|------|
| 522 / 523 | DNS・Worker 未設定。[deployment-dns-setup.md](./deployment-dns-setup.md) |
| Access 画面が出ない | DNS プロキシ OFF、または Application ドメインの不一致 |
| API が CORS エラー | Fly `ALLOWED_ORIGINS=https://translate.tattsum.com`（`fly.toml` + secrets） |
| ログイン後も 403 | Policy の Include 条件に自分のメールが含まれるか |

## 関連

- [deployment.md](./deployment.md) — 合意サマリ・アーキテクチャ
- [deployment-dns-setup.md](./deployment-dns-setup.md) — DNS / Workers Route
- [deployment-session-handoff.md](./deployment-session-handoff.md) — 現状スナップショット
