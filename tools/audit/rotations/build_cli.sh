#!/bin/bash
# usage: AUDIT_TREE=<checkout of the commit to sim> AUDIT_WORK=<scratch dir> tools/audit/rotations/build_cli.sh [out_name]
# Builds the sim CLI (with the embedded item db) from $AUDIT_TREE into $AUDIT_WORK/<out_name, default cli>.
set -euo pipefail
: "${AUDIT_TREE:?set AUDIT_TREE}"; : "${AUDIT_WORK:?set AUDIT_WORK}"
out="$AUDIT_WORK/${1:-cli}"; mkdir -p "$AUDIT_WORK"
export PATH="$PATH:$HOME/go/bin"
cd "$AUDIT_TREE"
protoc -I=./proto --go_out=./sim/core ./proto/*.proto   # generated Go protos are not committed
go build --tags=with_db -o "$out" ./cmd/wowsimcli/cli_main.go
echo "$out"
