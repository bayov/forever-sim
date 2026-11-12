#!/bin/bash
# Sets up a comparison of our sim with wowsims/forever (.claude/skills/compare-wowsims-forever):
#   - a checkout of wowsims/forever at $WSF (default ~/wowsims-forever), cloned or moved to its latest master;
#   - our local tools in it (theirs/), kept out of its git by .git/info/exclude: never pushed;
#   - their builds and APLs as JSON in $WSF_WORK/theirs_builds (default ~/wsf_compare);
#   - the binaries in $WSF_WORK/bin: theirs_cmpsim, theirs_stats, ours_cli, ours_stats, and ours_aq_cli and
#     ours_aq_stats (ours with the hunter's AQ book ranks, the only ranks theirs registers).
# Rerun it after either sim changes. It prints both commits: say them when reporting.
set -euo pipefail
HERE=$(cd "$(dirname "$0")" && pwd)
ROOT=$(cd "$HERE/../.." && pwd)
WSF=${WSF:-$HOME/wowsims-forever}
WORK=${WSF_WORK:-$HOME/wsf_compare}
BIN=$WORK/bin
mkdir -p "$BIN" "$WORK/theirs_builds"
export PATH=$PATH:/root/tools/protoc/bin:$HOME/go/bin
command -v protoc >/dev/null || { echo "protoc is missing (tools/dps_report/README.md says where it comes from)"; exit 1; }
command -v protoc-gen-go >/dev/null || go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.36.6

# --- theirs ---
if [ ! -d "$WSF/.git" ]; then
  git clone -q https://github.com/wowsims/forever "$WSF"
else
  git -C "$WSF" fetch -q origin master
  git -C "$WSF" checkout -q --detach origin/master
fi
EX=$WSF/.git/info/exclude
grep -q '^# forever-sim comparison tools' "$EX" 2>/dev/null || printf '# forever-sim comparison tools, never pushed\nbin/\ncmd/cmpsim/\ncmd/cmpstats/\n' >> "$EX"
for t in cmpsim cmpstats; do
  mkdir -p "$WSF/cmd/$t"
  cp "$HERE/theirs/$t/main.go.txt" "$WSF/cmd/$t/main.go"
done
(
  cd "$WSF"
  protoc -I=./proto --go_opt=Mgoogle/protobuf/descriptor.proto=google.golang.org/protobuf/types/descriptorpb \
    --go_out=./sim/core ./proto/*.proto
  go build -tags=with_db -o "$BIN/theirs_cmpsim" ./cmd/cmpsim
  go build -tags=with_db -o "$BIN/theirs_stats" ./cmd/cmpstats
  go build -tags=with_db -o "$BIN/theirs_cli" ./cmd/wowsimcli  # with_db: SIML parsing needs its name table
)
# Their builds (SIML or JSON) and their JSON APLs, as JSON: --apl theirs:<name>.
for f in "$WSF"/ui/specs/*/*/builds/*.siml "$WSF"/ui/specs/*/*/builds/*.build.json "$WSF"/ui/specs/*/*/apls/*.apl.json; do
  [ -e "$f" ] || continue
  rel=${f#"$WSF"/ui/specs/}; class=${rel%%/*}; rest=${rel#*/}; spec=${rest%%/*}
  base=$(basename "$f"); base=${base%.siml}; base=${base%.build.json}; base=${base%.apl.json}
  case $f in
    */apls/*) cp "$f" "$WORK/theirs_builds/apl_${class}_${spec}_$base.json" ;;
    *.siml) "$BIN/theirs_cli" siml parse -json "$f" > "$WORK/theirs_builds/${class}_${spec}_$base.json" 2>/dev/null \
              || echo "  could not parse $rel" ;;
    *) cp "$f" "$WORK/theirs_builds/${class}_${spec}_$base.json" ;;
  esac
done

# --- ours ---
(
  cd "$ROOT"
  [ -e sim/core/proto/common.pb.go ] || protoc -I=./proto --go_out=./sim/core ./proto/*.proto  # a fresh clone has none
  go build -tags=with_db -o "$BIN/ours_cli" ./cmd/wowsimcli
  go build -tags=with_db -o "$BIN/ours_stats" ./tools/wsf_compare/ours_stats
)
# Ours with the hunter's AQ ranks (Aspect of the Hawk VII, Serpent Sting IX), in a throwaway worktree. Only the
# hunter's: the global switch would also raise Battle Shout and Blessing of Might, which theirs keeps at their
# trainer ranks.
AQ=$WORK/aq_tree
git -C "$ROOT" worktree remove --force "$AQ" 2>/dev/null || true
git -C "$ROOT" worktree add -q --detach "$AQ" HEAD
trap 'git -C "$ROOT" worktree remove --force "$AQ" 2>/dev/null || true' EXIT
(
  cd "$AQ"
  grep -q 'if hunter.Env.IsForever() && !hunter.Env.IncludeAQRanks() {' sim/hunter/aspects.go \
    && grep -q 'if !hunter.Env.IncludeAQRanks() {' sim/hunter/serpent_sting.go \
    || { echo "the hunter's AQ rank checks moved: update setup.sh"; exit 1; }
  sed -i 's/if hunter.Env.IsForever() && !hunter.Env.IncludeAQRanks() {/if false {/' sim/hunter/aspects.go
  sed -i 's/if !hunter.Env.IncludeAQRanks() {/if false {/' sim/hunter/serpent_sting.go
  # The stats tool from this checkout, which may be newer than HEAD.
  mkdir -p tools/wsf_compare && rm -rf tools/wsf_compare/ours_stats && cp -r "$HERE/ours_stats" tools/wsf_compare/
  [ -e sim/core/proto/common.pb.go ] || protoc -I=./proto --go_out=./sim/core ./proto/*.proto
  go build -tags=with_db -o "$BIN/ours_aq_cli" ./cmd/wowsimcli
  go build -tags=with_db -o "$BIN/ours_aq_stats" ./tools/wsf_compare/ours_stats
)

echo "theirs: wowsims/forever $(git -C "$WSF" log --oneline -1)"
echo "ours:   forever-sim $(git -C "$ROOT" log --oneline -1)$(git -C "$ROOT" diff --quiet HEAD -- sim proto || echo ' (with uncommitted sim changes)')"
echo "binaries in $BIN, $(ls "$WORK/theirs_builds" | wc -l) of their builds and APLs in $WORK/theirs_builds"
