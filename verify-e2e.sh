#!/usr/bin/env bash
set -euo pipefail

BASE_URL="http://localhost:8080"
DB_URL="file:./data.db?cache=shared&_fk=1&_busy_timeout=10000"
PASS=0
FAIL=0

ok()   { echo "✅ $1"; ((PASS++)) || true; }
fail() { echo "❌ $1"; ((FAIL++)) || true; }

# ── Start server ─────────────────────────────────────────────────────────────
echo "==> Building server..."
cd fms
CGO_ENABLED=1 go build -o /tmp/fms-e2e ./cmd/server 2>&1
rm -f data.db

echo "==> Starting server..."
/tmp/fms-e2e serve --database-url "$DB_URL" > /tmp/fms-e2e.log 2>&1 &
SERVER_PID=$!

cleanup() {
  kill "$SERVER_PID" 2>/dev/null || true
  rm -f /tmp/fms-e2e data.db /tmp/dummy.zip
}
trap cleanup EXIT

# Wait for server to be ready (up to 10s)
for i in $(seq 1 20); do
  if curl -sf "$BASE_URL/health" >/dev/null 2>&1; then
    echo "==> Server ready"
    break
  fi
  sleep 0.5
done

# ── Phase 2: Library upload ───────────────────────────────────────────────────
echo ""
echo "── Phase 2: Library Upload ──"

MANIFEST='{"name":"test-fw","versionString":"1.0.0","versionNumber":"100","targetComponent":"BIOS"}'
mkdir -p /tmp/fwbundle
echo "$MANIFEST" > /tmp/fwbundle/manifest.json
(cd /tmp/fwbundle && zip /tmp/dummy.zip manifest.json) >/dev/null 2>&1

HTTP_STATUS=$(curl -s -o /tmp/upload_resp.json -w "%{http_code}" -F "file=@/tmp/dummy.zip" "$BASE_URL/library/upload")
if [ "$HTTP_STATUS" = "201" ]; then
  ok "POST /library/upload returned 201"
else
  fail "POST /library/upload returned $HTTP_STATUS (expected 201)"
  cat /tmp/upload_resp.json
fi

FW_COUNT=$(curl -sf "$BASE_URL/firmwareprofiles" | python3 -c "import sys,json; d=json.load(sys.stdin); print(len(d) if isinstance(d,list) else 1)" 2>/dev/null || echo 0)
if [ "$FW_COUNT" -ge 1 ]; then
  ok "FirmwareProfile record created (count=$FW_COUNT)"
else
  fail "FirmwareProfile record NOT found in database"
fi

# ── Phase 3: UpdateJob → UpdateTask splitting ─────────────────────────────────
echo ""
echo "── Phase 3: UpdateJob Splitter ──"

JOB_RESP=$(curl -sf -X POST "$BASE_URL/updatejobs" \
  -H "Content-Type: application/json" \
  -d '{"apiVersion":"example.fabrica.dev/v1","kind":"UpdateJob","metadata":{"name":"e2e-job"},"spec":{"targetNodes":["nodeA","nodeB"],"firmwareRef":"test-fw"}}')

JOB_UID=$(echo "$JOB_RESP" | python3 -c "import sys,json; print(json.load(sys.stdin)['metadata']['uid'])" 2>/dev/null || echo "")
if [ -n "$JOB_UID" ]; then
  ok "UpdateJob created (uid=$JOB_UID)"
else
  fail "UpdateJob creation failed"
  echo "$JOB_RESP"
fi

echo "==> Waiting 5s for reconciler to create UpdateTasks..."
sleep 5

TASK_COUNT=$(curl -sf "$BASE_URL/updatetasks" | python3 -c "import sys,json; d=json.load(sys.stdin); print(len(d) if isinstance(d,list) else 0)" 2>/dev/null || echo 0)
if [ "$TASK_COUNT" -ge 2 ]; then
  ok "UpdateTasks created (count=$TASK_COUNT, expected ≥2)"
else
  fail "Expected ≥2 UpdateTasks, got $TASK_COUNT"
fi

# ── Phase 4: UpdateTask execution (expect Failed – no real BMC) ───────────────
echo ""
echo "── Phase 4: UpdateTask Execution ──"

echo "==> Waiting 12s for UpdateTask reconciler to attempt Redfish calls..."
sleep 12

TASKS_JSON=$(curl -sf "$BASE_URL/updatetasks")
FAILED_COUNT=$(echo "$TASKS_JSON" | python3 -c "
import sys, json
tasks = json.load(sys.stdin)
if not isinstance(tasks, list): tasks = [tasks]
print(sum(1 for t in tasks if t.get('status', {}).get('state') in ('Failed', 'Success')))
" 2>/dev/null || echo 0)

if [ "$FAILED_COUNT" -ge 1 ]; then
  ok "UpdateTask state transitioned (Failed/Success count=$FAILED_COUNT) – execution loop confirmed"
else
  fail "No UpdateTasks reached terminal state after reconciliation"
  echo "$TASKS_JSON" | python3 -m json.tool 2>/dev/null || echo "$TASKS_JSON"
fi

# ── Phase 5: LookupJob execution ──────────────────────────────────────────────
echo ""
echo "── Phase 5: LookupJob Execution ──"

LOOKUP_RESP=$(curl -sf -X POST "$BASE_URL/lookupjobs" \
  -H "Content-Type: application/json" \
  -d '{"apiVersion":"example.fabrica.dev/v1","kind":"LookupJob","metadata":{"name":"e2e-lookup"},"spec":{"targetNode":"nodeA"}}')

LOOKUP_UID=$(echo "$LOOKUP_RESP" | python3 -c "import sys,json; print(json.load(sys.stdin)['metadata']['uid'])" 2>/dev/null || echo "")
if [ -n "$LOOKUP_UID" ]; then
  ok "LookupJob created (uid=$LOOKUP_UID)"
else
  fail "LookupJob creation failed"
  echo "$LOOKUP_RESP"
fi

echo "==> Waiting 12s for LookupJob reconciler..."
sleep 12

LOOKUP_STATE=$(curl -sf "$BASE_URL/lookupjobs/$LOOKUP_UID" | python3 -c "import sys,json; print(json.load(sys.stdin).get('status',{}).get('state',''))" 2>/dev/null || echo "")
if [ "$LOOKUP_STATE" = "Complete" ] || [ "$LOOKUP_STATE" = "Failed" ]; then
  ok "LookupJob reached terminal state: $LOOKUP_STATE"
else
  fail "LookupJob state is '$LOOKUP_STATE' (expected Complete or Failed)"
fi

# ── Summary ───────────────────────────────────────────────────────────────────
echo ""
echo "══════════════════════════════════"
echo "Results: $PASS passed, $FAIL failed"
if [ "$FAIL" -eq 0 ]; then
  echo "✅ All E2E checks passed!"
  exit 0
else
  echo "❌ Some checks failed. Check /tmp/fms-e2e.log for server logs."
  exit 1
fi