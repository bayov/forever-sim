#!/bin/bash
# Usage: AUDIT_TREE=<repo checkout at the ref to audit> AUDIT_WORK=<scratch dir> tools/audit/fairness/build.sh
# Builds one wowsimcli (with the item database) from AUDIT_TREE into $AUDIT_WORK/wowsimcli.
# Every class lives on master now, so one binary sims every request.
set -e
export PATH=$PATH:$HOME/go/bin
mkdir -p "$AUDIT_WORK"
cd "$AUDIT_TREE"
make sim/core/proto/api.pb.go >/dev/null
go build --tags=with_db -o "$AUDIT_WORK/wowsimcli" ./cmd/wowsimcli
echo "built $AUDIT_WORK/wowsimcli from $(git rev-parse --short HEAD)"
