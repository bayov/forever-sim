#!/bin/bash
# usage: tools/audit/build_cli.sh <rev>   ->  prints the path of a wowsimcli built from <rev>
# Builds once per commit into $AUDIT_CACHE/cli/<sha>/ (a detached worktree there, protos generated,
# --tags=with_db). Later calls for the same commit return the cached binary. The worktree stays, so
# scripts can run the sim with that tree as cwd: its path is the binary's directory + /tree.
set -euo pipefail
: "${AUDIT_CACHE:=/tmp/audit-cache}"
REPO="${AUDIT_REPO:-$(git rev-parse --show-toplevel)}"
sha=$(git -C "$REPO" rev-parse --short=12 "$1^{commit}")
out="$AUDIT_CACHE/cli/$sha"
if [ -x "$out/wowsimcli" ]; then echo "$out/wowsimcli"; exit 0; fi
mkdir -p "$out"
[ -d "$out/tree" ] || git -C "$REPO" worktree add -q --detach "$out/tree" "$sha"
(
  cd "$out/tree"
  export PATH="$PATH:$HOME/go/bin:/root/go/bin"
  protoc -I=./proto --go_out=./sim/core ./proto/*.proto
  go build --tags=with_db -o "$out/wowsimcli" ./cmd/wowsimcli/cli_main.go
) >"$out/build.log" 2>&1 || { echo "build_cli.sh: build of $sha failed, see $out/build.log" >&2; exit 1; }
echo "$out/wowsimcli"
