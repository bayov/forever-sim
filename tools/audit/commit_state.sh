#!/bin/bash
# usage: tools/audit/commit_state.sh "<commit message>" [extra paths...]
# Commits the audit loop's machine-written files (orchestration/audit/state, findings, log.md, ref,
# reports, plus any extra paths, such as the docs an auditor "owns" in the registry) and pushes them
# to master as HEAD:master from whatever local branch the run is on (README: a throwaway audit-run branch),
# rebasing over anyone else's push (4 tries, never force). Does nothing when they didn't change.
set -uo pipefail
cd "$(git rev-parse --show-toplevel)"
P="orchestration/audit"
msg="$1"; shift
for p in "$P/state" "$P/findings" "$P/log.md" "$P/ref" "$P/reports" "$@"; do
  [ -e "$p" ] && git add -A "$p"
done
git diff --cached --quiet && { echo "commit_state: nothing to commit"; exit 0; }
git commit -q -m "$msg" || exit 1
for i in 1 2 3 4; do
  git pull -q --rebase origin master && git push -q origin HEAD:master && { echo "commit_state: pushed $(git rev-parse --short HEAD)"; exit 0; }
  sleep $((2 ** i))
done
echo "commit_state: push failed after 4 tries" >&2; exit 1
