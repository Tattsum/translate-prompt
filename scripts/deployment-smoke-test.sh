#!/usr/bin/env bash
# translate-prompt 本番スモークテスト（自動チェック + 手動確認リスト）
#
# 使い方:
#   ./scripts/deployment-smoke-test.sh
#   SPA_URL=https://translate.tattsum.com API_URL=https://prompt-api.tattsum.com ./scripts/deployment-smoke-test.sh
#
# bash で実行すること（sh だと色付き出力が崩れる場合あり）。
#
# Access 設定前: Workers API 経由で health JSON を期待。
# Access 設定後: SPA / API は未認証 302。バックエンド疎通は Fly 直結で確認。
# Access 設定後の GraphQL / CORS はブラウザ（ログイン後）で MANUAL 確認。

set -euo pipefail

SPA_URL="${SPA_URL:-https://translate.tattsum.com}"
API_URL="${API_URL:-https://prompt-api.tattsum.com}"
PAGES_URL="${PAGES_URL:-https://translate-prompt-2el.pages.dev}"
FLY_URL="${FLY_URL:-https://translate-prompt-api.fly.dev}"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

pass() { printf "${GREEN}PASS${NC} %s\n" "$*"; }
fail() { printf "${RED}FAIL${NC} %s\n" "$*"; exit 1; }
warn() { printf "${YELLOW}WARN${NC} %s\n" "$*"; }
info() { printf "     %s\n" "$*"; }

http_status() {
  curl -sS -o /dev/null -w '%{http_code}' --max-time 15 "$1"
}

is_access_redirect() {
  local status=$1
  local body=$2
  [[ "$status" == "302" || "$status" == "303" ]] && return 0
  echo "$body" | grep -qi 'cloudflareaccess.com\|302 Found' && return 0
  return 1
}

echo "=== translate-prompt deployment smoke test ==="
echo "SPA_URL=${SPA_URL}"
echo "API_URL=${API_URL}"
echo "PAGES_URL=${PAGES_URL}"
echo ""

echo "--- 1. API (Workers 経由) ---"
api_tmp=$(mktemp)
api_status=$(curl -sS -o "$api_tmp" -w '%{http_code}' --max-time 15 -X POST "${API_URL}/query" \
  -H 'Content-Type: application/json' \
  -d '{"query":"{ health { status } }"}')
api_body=$(cat "$api_tmp")
rm -f "$api_tmp"

if echo "$api_body" | grep -q '"status":"ok"'; then
  pass "Workers API health: $(echo "$api_body" | tr -d '\n')"
elif is_access_redirect "$api_status" "$api_body"; then
  pass "Workers API HTTP ${api_status} — Cloudflare Access リダイレクト（設定済み）"
  info "未認証 curl では JSON を取得できない。バックエンドは §4 Fly 直結で確認"
else
  fail "Workers API unexpected (HTTP ${api_status}): ${api_body}"
fi

echo ""
echo "--- 2. Pages 直接 (SPA ビルド) ---"
pages_status=$(http_status "${PAGES_URL}/")
if [[ "$pages_status" == "200" ]]; then
  pass "Pages root HTTP ${pages_status}"
else
  fail "Pages root HTTP ${pages_status} (expected 200)"
fi

pages_settings_status=$(http_status "${PAGES_URL}/settings")
if [[ "$pages_settings_status" == "200" ]]; then
  pass "Pages SPA routing /settings HTTP ${pages_settings_status}"
else
  fail "Pages /settings HTTP ${pages_settings_status} (expected 200)"
fi

echo ""
echo "--- 3. SPA Worker プロキシ (translate.tattsum.com) ---"
spa_status=$(http_status "${SPA_URL}/")
case "$spa_status" in
  200)
    warn "SPA URL HTTP ${spa_status} — Access 未設定の可能性（設定後は 302 が期待値）"
    ;;
  302|303)
    pass "SPA URL HTTP ${spa_status} — Cloudflare Access リダイレクト（設定済み）"
    ;;
  522|523)
    fail "SPA URL HTTP ${spa_status} — DNS / Worker 未設定。docs/deployment-dns-setup.md を参照"
    ;;
  *)
    warn "SPA URL HTTP ${spa_status} — 要確認"
    ;;
esac

spa_settings_status=$(http_status "${SPA_URL}/settings")
case "$spa_settings_status" in
  302|303)
    info "SPA /settings HTTP ${spa_settings_status} (Access 保護下)"
    ;;
  *)
    info "SPA /settings HTTP ${spa_settings_status}"
    ;;
esac

echo ""
echo "--- 4. Fly 直結 (オリジン疎通) ---"
fly_tmp=$(mktemp)
fly_status=$(curl -sS -o "$fly_tmp" -w '%{http_code}' --max-time 15 -X POST "${FLY_URL}/query" \
  -H 'Content-Type: application/json' \
  -d '{"query":"{ health { status } }"}' 2>/dev/null || echo "000")
fly_body=$(cat "$fly_tmp" 2>/dev/null || true)
rm -f "$fly_tmp"

if echo "$fly_body" | grep -q '"status":"ok"'; then
  pass "Fly direct health ok (HTTP ${fly_status})"
else
  warn "Fly direct HTTP ${fly_status}: ${fly_body:-unreachable}"
fi

echo ""
echo "--- 5. CORS (許可オリジン) ---"
cors_hdr=$(mktemp)
cors_status=$(curl -sS -D "$cors_hdr" -o /dev/null -w '%{http_code}' --max-time 15 -X OPTIONS "${API_URL}/query" \
  -H "Origin: https://evil.example" \
  -H "Access-Control-Request-Method: POST" 2>/dev/null || echo "000")
cors_headers=$(cat "$cors_hdr" 2>/dev/null || true)
rm -f "$cors_hdr"

if is_access_redirect "$cors_status" "$cors_headers"; then
  info "API OPTIONS HTTP ${cors_status} — Access 保護下のため CORS はブラウザ（ログイン後）で確認"
elif echo "$cors_headers" | grep -qi 'access-control-allow-origin: \*'; then
  warn "CORS allows * — ALLOWED_ORIGINS が未適用の可能性"
elif echo "$cors_headers" | grep -qi "access-control-allow-origin: ${SPA_URL}"; then
  pass "CORS allows ${SPA_URL}"
elif echo "$cors_headers" | grep -qi 'access-control-allow-origin'; then
  info "CORS headers present (verify origin manually):"
  echo "$cors_headers" | grep -i access-control || true
else
  info "No CORS allow-origin for evil.example (expected if restricted)"
fi

echo ""
echo "--- 6. GraphQL Playground (本番無効) ---"
playground_status=$(http_status "${API_URL}/playground")
case "$playground_status" in
  404)
    pass "Playground HTTP 404 (disabled in production)"
    ;;
  302|303)
    info "Playground HTTP ${playground_status} — Access 保護下（ログイン後 404 か要確認）"
    ;;
  *)
    warn "Playground HTTP ${playground_status} (expected 404 or Access 302)"
    ;;
esac

echo ""
echo "=== MANUAL: Access ログイン後にブラウザで確認 ==="
cat <<EOF
[ ] 未認証で ${SPA_URL} → Access ログイン画面
[ ] 認証後 SPA 表示（Input ページ）
[ ] Estimate: トークン数が返る
[ ] Analyze: 質問または ready が返る
[ ] Optimize: 最適化結果が返る
[ ] Investigate（Web）: 403 / エラー（INVESTIGATE_ENABLED=false）
[ ] CLI Investigate: ローカルで従来どおり動作
[ ] Workspace Path: Settings に入力欄がない（本番ビルド）

Access 設定手順: docs/deployment-access-setup.md
DNS 設定手順:   docs/deployment-dns-setup.md
EOF

echo ""
echo "=== Done ==="
