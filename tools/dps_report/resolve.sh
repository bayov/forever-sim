#!/bin/bash
# Resolves the routine conflicts while merging class heads into the merged dashboard tree:
# assets/database/db.{json,bin} take ours (they are regenerated with gen_db after all merges),
# *.md files take a union of both sides. Other conflicts are left for a manual fix.
# Tree: $DASH_TREE, default a w-dash checkout next to this repository.
HERE="$(cd "$(dirname "$0")" && pwd)"
cd "${DASH_TREE:-$HERE/../../../w-dash}" || exit 1
TMP=$(mktemp -d)
trap 'rm -rf "$TMP"' EXIT
for f in $(git diff --name-only --diff-filter=U); do
  case $f in
    assets/database/db.json|assets/database/db.bin) git checkout --ours -- "$f"; git add "$f";;
    *.md) git show :2:"$f" > "$TMP/ours"; git show :1:"$f" > "$TMP/base" 2>/dev/null || : > "$TMP/base"; git show :3:"$f" > "$TMP/theirs"
          git merge-file --union -p "$TMP/ours" "$TMP/base" "$TMP/theirs" > "$f"; git add "$f";;
  esac
done
