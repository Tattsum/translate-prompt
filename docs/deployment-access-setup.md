# Cloudflare Access セットアップ（招待制 β）

手動設定。Zero Trust ダッシュボードで実施する。  
対象アカウント: **Kurohari35@gmail.com's Account**（`tattsum.com` ゾーン）

## 前提

- DNS が Cloudflare プロキシ（オレンジ雲）ON であること
- SPA: `translate.tattsum.com` → Worker `translate-prompt-web`（[deployment-dns-setup.md](./deployment-dns-setup.md)）
- API: `prompt-api.tattsum.com` → Worker `translate-prompt-api-proxy`（設定済み）

## 1. Identity Provider（初回のみ）

1. [Cloudflare Zero Trust](https://one.dash.cloudflare.com/) を開く
2. **Settings** → **Authentication**
3. 利用する IdP を追加（例: **Google** または **GitHub**）
4. テストログインが成功することを確認

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
| Include | 例: `Emails ending in @yourdomain.com`、または `Emails` で招待メールを列挙 |

5. **Save application**

## 3. Application 2: API

1. 同様に **Add an application** → **Self-hosted**
2. 設定:

| 項目 | 値 |
|------|-----|
| Application name | `translate-prompt-api` |
| Session Duration | `24 hours` |
| Application domain | `prompt-api.tattsum.com` |

3. Policy は SPA と同じ招待条件（`invite-only`）を適用
4. **Save application**

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
