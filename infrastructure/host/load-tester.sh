#!/usr/bin/env bash
set -euo pipefail

# Help
if [ $# -eq 0 ]; then
    echo "Usage: $0 <URL> [DURATION] [MAX_WORKERS]"
    exit 1
fi

URL="$1"
DURATION="${2:-20}"
MAX_WORKERS="${3:-500}"

start=$(date +%s)

# ---- COUNTERS (shared) ----
tmpdir=$(mktemp -d)
total_f="$tmpdir/total"
success_f="$tmpdir/success"
failed_f="$tmpdir/failed"
printf "0" >"$total_f"
printf "0" >"$success_f"
printf "0" >"$failed_f"

pids=()

cleanup() {
    kill "${pids[@]}" 2>/dev/null || true
    elapsed=$(($(date +%s) - start))
    total=$(<"$total_f")
    success=$(<"$success_f")
    failed=$(<"$failed_f")
    echo -e "\n\n=== Results ==="
    echo "Duration: ${elapsed}s | Total: $total | Success: $success | Failed: $failed | RPS: $((total / elapsed))"
    rm -rf "$tmpdir"
    exit 0
}
trap cleanup INT TERM EXIT

inc() {
    local f="$1"
    local lock="${f}.lockdir"

    while ! mkdir "$lock" 2>/dev/null; do
        sleep 0.001
    done

    n=$(<"$f")
    printf "%d" "$((n + 1))" >"$f"

    rmdir "$lock"
}


worker() {
    while true; do
        if curl -sf --max-time 10 --no-keepalive "$URL" >/dev/null 2>&1; then
            inc "$success_f"
        else
            inc "$failed_f"
        fi
        inc "$total_f"
        sleep 0.$((RANDOM % 100))
    done
}

echo ""
echo "Target: $URL | Duration: ${DURATION}s | Max Workers: $MAX_WORKERS"

for ((i=1; i<=MAX_WORKERS; i++)); do
    worker &
    pids+=($!)
    sleep 0.02
done

while [ $(($(date +%s) - start)) -lt "$DURATION" ]; do
    total=$(<"$total_f")
    success=$(<"$success_f")
    failed=$(<"$failed_f")
    printf "\r\n[%3ds] Workers: %3d | Total: %6d | Success: %6d | Failed: %4d | RPS: %4d" \
        "$((DURATION - $(date +%s) + start))" "${#pids[@]}" \
        "$total" "$success" "$failed" "$((total / ($(date +%s) - start + 1)))"
    sleep 1
done
