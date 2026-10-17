#!/bin/bash
# Usage: AUDIT_TREE=<sim checkout> AUDIT_WORK=<scratch dir> tools/audit/rank_tuning/build.sh
# Applies the rank-tuning env-var patch to AUDIT_TREE (a throwaway checkout of origin/master, never
# committed), regenerates the Go protos and builds $AUDIT_WORK/cli_master (wowsimcli with the item db).
set -euo pipefail
: "${AUDIT_TREE:?set AUDIT_TREE to a scratch checkout of origin/master}"; : "${AUDIT_WORK:?set AUDIT_WORK}"
HERE="$(cd "$(dirname "$0")" && pwd)"
mkdir -p "$AUDIT_WORK"
cd "$AUDIT_TREE"
if git apply --check "$HERE/rt_sensitivity.patch" 2>/dev/null; then git apply "$HERE/rt_sensitivity.patch"
elif git apply --reverse --check "$HERE/rt_sensitivity.patch" 2>/dev/null; then echo "patch already applied"
else echo "patch does not apply: re-add the RT_* hooks by hand (see handoff gotchas) and refresh the patch"; exit 1; fi
export PATH="$PATH:$(go env GOPATH)/bin"
command -v protoc-gen-go >/dev/null || go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
protoc -I=./proto --go_out=./sim/core ./proto/*.proto
go build --tags=with_db -o "$AUDIT_WORK/cli_master" ./cmd/wowsimcli
echo "built $AUDIT_WORK/cli_master at $(git rev-parse --short HEAD)"
