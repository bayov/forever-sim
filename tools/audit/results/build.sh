#!/bin/bash
# Usage: build.sh
# Regenerates the Go protos in AUDIT_TREE and builds wowsimcli and mksettings (both --tags=with_db) into AUDIT_BIN.
# Needs protoc and protoc-gen-go (see orchestration/audit/handoff/results.md, "Setup").
# Env: AUDIT_TREE (default: this repo), AUDIT_WORK (default ${TMPDIR:-/tmp}/audit-results), AUDIT_BIN (default $AUDIT_WORK/bin).
set -eu
D=$(cd "$(dirname "$0")" && pwd)
AUDIT_TREE=${AUDIT_TREE:-$(cd "$D/../../.." && pwd)}
AUDIT_WORK=${AUDIT_WORK:-${TMPDIR:-/tmp}/audit-results}
AUDIT_BIN=${AUDIT_BIN:-$AUDIT_WORK/bin}
export PATH=$PATH:$HOME/go/bin
mkdir -p "$AUDIT_BIN"
cd "$AUDIT_TREE"
protoc -I=./proto --go_out=./sim/core ./proto/*.proto
go build --tags=with_db -o "$AUDIT_BIN/wowsimcli" ./cmd/wowsimcli
go build --tags=with_db -o "$AUDIT_BIN/mksettings" ./tools/rotopt/mksettings
echo "built $AUDIT_BIN/wowsimcli and $AUDIT_BIN/mksettings at $(git rev-parse --short HEAD)"
