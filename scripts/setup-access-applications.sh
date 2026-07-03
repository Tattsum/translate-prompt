#!/usr/bin/env bash
# Cloudflare Access Application / Policy を API で作成する（IdP はダッシュボードで事前設定）
#
# 前提:
#   - Zero Trust で Google / GitHub IdP が追加済み
#   - CLOUDFLARE_API_TOKEN（Access: Edit 権限）
#   - CLOUDFLARE_ACCOUNT_ID
#
# 使い方:
#   export CLOUDFLARE_API_TOKEN=...
#   export CLOUDFLARE_ACCOUNT_ID=...
#   INVITE_EMAILS='kurohari35@gmail.com,other@example.com' ./scripts/setup-access-applications.sh
#
# ドライラン:
#   DRY_RUN=1 INVITE_EMAILS='you@example.com' ./scripts/setup-access-applications.sh

set -euo pipefail

ACCOUNT_ID="${CLOUDFLARE_ACCOUNT_ID:?CLOUDFLARE_ACCOUNT_ID is required}"
TOKEN="${CLOUDFLARE_API_TOKEN:?CLOUDFLARE_API_TOKEN is required}"
INVITE_EMAILS="${INVITE_EMAILS:?INVITE_EMAILS is required (comma-separated)}"
SESSION_DURATION="${SESSION_DURATION:-24h}"
DRY_RUN="${DRY_RUN:-0}"

API="https://api.cloudflare.com/client/v4/accounts/${ACCOUNT_ID}/access/apps"

auth_header=(-H "Authorization: Bearer ${TOKEN}" -H "Content-Type: application/json")

build_emails_include() {
  local json='['
  local first=1
  IFS=',' read -ra emails <<< "$INVITE_EMAILS"
  for raw in "${emails[@]}"; do
    local email
    email="$(echo "$raw" | xargs)"
    [[ -z "$email" ]] && continue
    if [[ "$first" -eq 0 ]]; then json+=','; fi
    json+="{\"email\":{\"email\":\"${email}\"}}"
    first=0
  done
  json+=']'
  echo "$json"
}

create_app() {
  local name="$1"
  local domain="$2"
  local includes
  includes="$(build_emails_include)"

  local payload
  payload=$(cat <<EOF
{
  "name": "${name}",
  "type": "self_hosted",
  "session_duration": "${SESSION_DURATION}",
  "domain": "${domain}",
  "auto_redirect_to_identity": false,
  "allowed_idps": [],
  "policies": [
    {
      "name": "invite-only",
      "decision": "allow",
      "include": ${includes}
    }
  ]
}
EOF
)

  echo "=== ${name} (${domain}) ==="
  if [[ "$DRY_RUN" == "1" ]]; then
    echo "$payload" | jq .
    return 0
  fi

  local resp
  resp=$(curl -sS "${auth_header[@]}" -X POST "$API" -d "$payload")
  if echo "$resp" | jq -e '.success == true' >/dev/null; then
    echo "Created: $(echo "$resp" | jq -r '.result.name') id=$(echo "$resp" | jq -r '.result.id')"
  else
    echo "$resp" | jq .
    echo "Failed to create ${name}" >&2
    return 1
  fi
}

echo "Creating Access applications for translate-prompt..."
echo "Invite emails: ${INVITE_EMAILS}"
echo ""

create_app "translate-prompt-spa" "translate.tattsum.com"
create_app "translate-prompt-api" "prompt-api.tattsum.com"

echo ""
echo "Done. Verify with: ./scripts/deployment-smoke-test.sh"
echo "Expected: SPA returns HTTP 302 (Access redirect)"
