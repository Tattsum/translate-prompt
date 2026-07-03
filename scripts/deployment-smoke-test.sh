#!/usr/bin/env bash
# translate-prompt 本番スモークテスト（自動チェック + 手動確認リスト）
#
# 使い方:
#   ./scripts/deployment-smoke-test.sh
#   SPA_URL=https://translate.tattsum.com API_URL=https://prompt-api.tattsum.com ./scripts/deployment-smoke-test.sh
#
# Access 設定前でも API / Pages の疎通確認は可能。
# Access 設定後はブラウザでの手動確認（下記 MANUAL セクション）を実施すること。

set -euo pipefail

SPA_URL="${SPA_URL:-https://translate.tattsum.com}"
API_URL="${API_URL:-https://prompt-api.tattsum.com}"
PAGES_URL="${PAGES_URL:-https://translate-prompt-2el.pages.dev}"
FLY_URL="${FLY_URL:-https://translate-prompt-api.fly.dev}"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

pass() { echo -e "${GREEN}PASS${NC} $*"; }
fail() { echo -e "${RED}FAIL${NC} $*"; exit 1; }
warn() { echo -e "${YELLOW}WARN${NC} $*"; }
info() { echo -e "     $*"; }

http_status() {
  curl -sS -o /dev/null -w '%{http_code}' --max-time 15 "$1"
}

http_body() {
  curl -sS --max-time 15 "$@"
}

echo "=== translate-prompt deployment smoke test ==="
echo "SPA_URL=${SPA_URL}"
echo "API_URL=${API_URL}"
echo "PAGES_URL=${PAGES_URL}"
echo ""

echo "--- 1. API Health (Workers 経由) ---"
health=$(http_body -X POST "${API_URL}/query" \
  -H 'Content-Type: application/json' \
  -d '{"query":"{ health { status } }"}')
if echo "$health" | grep -q '"status":"ok"'; then
  pass "Workers API health: ${health}"
else
  fail "Workers API health unexpected: ${health}"
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
info "SPA /settings HTTP ${spa_settings_status}"

echo ""
echo "--- 4. Fly 直結 (参考) ---"
fly_health=$(http_body -X POST "${FLY_URL}/query" \
  -H 'Content-Type: application/json' \
  -d '{"query":"{ health { status } }"}' 2>/dev/null || echo '{"error":"unreachable"}')
if echo "$fly_health" | grep -q '"status":"ok"'; then
  pass "Fly direct health ok"
else
  warn "Fly direct: ${fly_health}"
fi

echo ""
echo "--- 5. CORS (許可オリジン) ---"
cors_headers=$(curl -sS -D - -o /dev/null --max-time 15 -X OPTIONS "${API_URL}/query" \
  -H "Origin: https://evil.example" \
  -H "Access-Control-Request-Method: POST" 2>/dev/null | tr -d '\r')
if echo "$cors_headers" | grep -qi 'access-control-allow-origin: \*'; then
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
if [[ "$playground_status" == "404" ]]; then
  pass "Playground HTTP 404 (disabled in production)"
else
  warn "Playground HTTP ${playground_status} (expected 404)"
fi

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
