# Cloudflare Access セットアップ（招待制 β）

手動設定。Zero Trust ダッシュボードで実施する。  
対象アカウント: **Kurohari35@gmail.com's Account**（`tattsum.com` ゾーン）  
Team domain: **`jolly-glitter-2f2b.cloudflareaccess.com`**

## 現状（2026-07-03 設定完了）

| 項目 | 状態 |
|------|------|
| DNS / Workers | 完了 |
| API 疎通 | 完了（Estimate / Analyze / Optimize 動作確認済み） |
| IdP（Google / GitHub） | 完了 |
| Access Application（SPA + API） | **完了** |
| 未認証アクセス | **302** → Access ログイン画面 |

未認証で `https://translate.tattsum.com` にアクセスすると、  
`https://jolly-glitter-2f2b.cloudflareaccess.com/cdn-cgi/access/login/...` へリダイレクトされる。

## 完了チェックリスト

- [x] IdP **Google** と **GitHub** の両方が追加済み
- [x] Application `translate-prompt-spa`（`translate.tattsum.com`）作成済み
- [x] Application `translate-prompt-api`（`prompt-api.tattsum.com`）作成済み
- [x] 両アプリに `invite-only` ポリシー適用済み
- [x] 未認証で SPA → Access ログイン画面（HTTP 302）
- [x] ログイン後 Input ページ表示
- [x] Analyze → Optimize がログイン後も動作（2026-07-03 ブラウザ確認済み）
- [x] `./scripts/deployment-smoke-test.sh` で SPA / API が 302（2026-07-03 確認済み）

## 前提

- DNS が Cloudflare プロキシ（オレンジ雲）ON であること
- SPA: 本番は **Tunnel 不要**（Workers + 公開 DNS）
- SPA: `translate.tattsum.com` → Worker `translate-prompt-web`（[deployment-dns-setup.md](./deployment-dns-setup.md)）
- API: `prompt-api.tattsum.com` → Worker `translate-prompt-api-proxy`

## UI の見分け（重要）

| 画面 | 使う？ |
|------|--------|
| **セルフホストとプライベート → パブリック DNS** | ✅ 本番用 |
| **プライベート宛先**（トンネル割り当てあり） | ❌ Tunnel 連携用 |
| **SaaS アプリケーション** タブ | ❌ 外部 SaaS への SSO 用（今回は不要） |

「アプリケーション名」を **サブドメイン欄に入れない**。サブドメインは URL のホスト名部分のみ（例: `translate`）。

---

## 1. Identity Provider（初回のみ）

方針: **Google と GitHub の両方**を有効にする。招待制の判定は IdP 種別ではなく **Policy の Emails 条件**で行う。

1. [Cloudflare Zero Trust](https://one.dash.cloudflare.com/) を開く
2. アカウントが **Kurohari35@gmail.com's Account** であることを確認
3. **Integrations** → **Identity providers**（または概要の「ID プロバイダーを接続する）→ **Add new**

### 共通 Callback URL

Google / GitHub 両方で同じ URL を使う:

```txt
https://jolly-glitter-2f2b.cloudflareaccess.com/cdn-cgi/access/callback
```

Google の Authorized JavaScript origins:

```txt
https://jolly-glitter-2f2b.cloudflareaccess.com
```

### 1a. Google

1. **Google** を選択
2. [Google Cloud Console](https://console.cloud.google.com/apis/credentials) で OAuth クライアント（Web application）を作成
3. redirect URI / origins は上記を使用
4. Client ID / Secret を Zero Trust に貼り付け → **Save**

### 1b. GitHub

1. **Add new** → **GitHub**
2. [GitHub OAuth Apps](https://github.com/settings/developers) で **New OAuth App**:

| 項目 | 値 |
|------|-----|
| Application name | `translate-prompt-access` |
| Homepage URL | `https://jolly-glitter-2f2b.cloudflareaccess.com` |
| Authorization callback URL | 上記 Callback URL |

1. Client ID / Secret を Zero Trust → **Save** → **Finish setup**

### 1c. Test ボタンについて

IdP 横の **Test** リンクで `Unable to find the requested identity provider!` が出ることがある（Cloudflare 既知事象）。  
**Test は無視**し、Application 作成後に本番 URL で確認する。

### 1d. 両 IdP 共通の注意

| 項目 | 内容 |
|------|------|
| Policy のキー | **Emails**（GitHub username ではなくメール） |
| GitHub のメール | primary email が verified。private email ON 時は `*@users.noreply.github.com` を Policy に追加 |
| 招待時 | 利用者の「Cloudflare に渡るメール」を Policy に追加 |

---

## 2. Application 1: SPA

1. **Access コントロール** → **Applications** → **Add an application**
2. **セルフホストとプライベート** → **パブリック DNS** → **続行**
3. **新規セルフホストアプリケーションを作成** で入力:

### アプリケーションの詳細

| UI 項目 | 値 |
|---------|-----|
| サブドメイン | `translate` |
| ドメイン | `tattsum.com`（ドロップダウン） |
| パス | 空（サイト全体） |

→ 保護対象: `https://translate.tattsum.com`

### Access ポリシー

**新しいポリシーを作成**:

| 項目 | 値 |
|------|-----|
| Policy name | `invite-only` |
| Action | **Allow** |
| Include | **Emails** — 招待メールを列挙（β 初期は個別が安全） |

### 認証

| UI 項目 | 値 |
|---------|-----|
| 利用可能なすべての ID プロバイダーを受け入れる | **ON** |
| インスタント認証を適用 | OFF |
| Cloudflare One Client で認証 | OFF |

### 詳細

| 項目 | 値 |
|------|-----|
| 名前 | `translate-prompt-spa` |
| セッション期間 | `24 hours` |

1. **保存**

---

## 3. Application 2: API

手順は SPA と同じ（**パブリック DNS**）。

| UI 項目 | 値 |
|---------|-----|
| サブドメイン | `prompt-api` |
| ドメイン | `tattsum.com` |
| パス | 空 |
| 名前 | `translate-prompt-api` |
| セッション期間 | `24 hours` |
| 認証 | すべての IdP を受け入れる **ON** |
| Policy | SPA と**同じ** `invite-only`（メール一覧を揃える） |

→ 保護対象: `https://prompt-api.tattsum.com`

---

## 4. 動作確認

```bash
./scripts/deployment-smoke-test.sh
```

未認証 curl では SPA / API は **302**（Access）。オリジン疎通はスクリプト内の Fly 直結で確認。  
GraphQL / CORS / Analyze は **ログイン後ブラウザ**で MANUAL 確認。

手動確認:

```bash
curl -sI https://translate.tattsum.com/ | head -3
# HTTP/2 302
# location: https://jolly-glitter-2f2b.cloudflareaccess.com/cdn-cgi/access/login/...
```

ブラウザ:

1. シークレットウィンドウで `https://translate.tattsum.com` → Access ログイン（Google / GitHub）
2. ログイン後 Input ページ表示
3. Analyze → Optimize が動作

API はブラウザ `fetch` のため SPA と同じ Access セッションで保護される。  
Workers API プロキシは Access JWT を Fly に転送**しない**（`workers/src/index.ts`）。

### API で Application 作成（任意）

IdP 設定済みならスクリプトでも可:

```bash
export CLOUDFLARE_API_TOKEN=...
export CLOUDFLARE_ACCOUNT_ID=...
INVITE_EMAILS='you@example.com' ./scripts/setup-access-applications.sh
```

---

## 5. 招待者の追加

β 招待時は両 Application の Policy **Include** にメールを追加する（SPA / API で一覧を揃える）。  
または `Emails ending in @company.com` でドメイン許可。

Zero Trust → **Access コントロール** → 対象 Application → Policy を編集。

---

## トラブルシューティング

| 症状 | 確認 |
|------|------|
| 522 / 523 | [deployment-dns-setup.md](./deployment-dns-setup.md) |
| Access 画面が出ない（200 のまま） | Application 未作成、またはサブドメイン/ドメインの typo |
| トンネル設定を求められる | **プライベート宛先** を選んでいる → **パブリック DNS** に変更 |
| ログイン後 403 | Policy の Emails に自分のメールが含まれるか |
| GitHub だけ 403 | GitHub が渡すメール（noreply 含む）を Policy に追加 |
| IdP Test でエラー | Test は無視。本番 URL で確認 |
| **`Unable to find the requested identity provider!`** | Test / Finish setup 時の既知事象。IdP が一覧にあれば Application 作成へ進む |

## 関連

- [deployment.md](./deployment.md) — 合意サマリ・アーキテクチャ
- [deployment-dns-setup.md](./deployment-dns-setup.md) — DNS / Workers Route
- [deployment-session-handoff.md](./deployment-session-handoff.md) — 現状スナップショット
