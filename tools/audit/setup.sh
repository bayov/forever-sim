#!/bin/bash
# usage: source tools/audit/setup.sh
# Idempotent toolchain setup for an audit pass in a fresh cloud session: protoc, protoc-gen-go,
# and the env vars every audit script reads. Safe to source more than once.
export AUDIT_REPO="${AUDIT_REPO:-$(git rev-parse --show-toplevel)}"
export AUDIT_WORK="${AUDIT_WORK:-/tmp/audit-work}"
export AUDIT_CACHE="${AUDIT_CACHE:-/tmp/audit-cache}"
export PATH="$PATH:$HOME/go/bin:/root/go/bin"
mkdir -p "$AUDIT_WORK" "$AUDIT_CACHE"
if ! command -v protoc >/dev/null; then
  (apt-get install -y -qq protobuf-compiler >/dev/null 2>&1 || (apt-get update -qq >/dev/null 2>&1 && apt-get install -y -qq protobuf-compiler >/dev/null 2>&1)) \
    || echo "setup.sh: protoc install failed" >&2
fi
command -v protoc-gen-go >/dev/null || go install google.golang.org/protobuf/cmd/protoc-gen-go@latest >/dev/null 2>&1 \
  || echo "setup.sh: protoc-gen-go install failed" >&2
git -C "$AUDIT_REPO" config user.name >/dev/null || git -C "$AUDIT_REPO" config user.name "Claude"
git -C "$AUDIT_REPO" config user.email >/dev/null || git -C "$AUDIT_REPO" config user.email "noreply@anthropic.com"
