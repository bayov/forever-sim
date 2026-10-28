#!/bin/bash
# pipeline.sh: every sim step of the DPS report for a set of builds, in one command.
#
#   pipeline.sh hydrate [ids...]        empty $DASH_DIR's outputs, load every build's raw pack (raw/<id>.tgz), rebuild
#                                       the datasets and check their sim numbers against the published report; a
#                                       build outside [ids] that differs stops it (orchestration/dash_raw_store.md)
#   pipeline.sh plan  <ids...>          count every step's sims and check each script can enumerate its work (seconds)
#   pipeline.sh run   <ids...>          add the ids' duty and encounter variants, clean them, plan, then the sim
#                                       steps; prints a timestamp per step
#   pipeline.sh resume <ids...>         run without the clean, to finish a run that stopped (every step skips done work)
#   pipeline.sh clean <ids...>          drop these builds from every raw output, so every step re-sims them
#   pipeline.sh publish [ids...]        build_data.py, the sample fights of [ids] and their variants (events.py),
#                                       every dataset into ui/dps_report/data, and provenance and raw packs for them
#   pipeline.sh shard <k> <n> <ids...>  the ids of shard k of n (0-based), balanced by estimated cost
#   pipeline.sh pack  <tgz> <ids...>    these builds' raw outputs, to merge on another machine (a shard's hand-back)
#   pipeline.sh unpack <tgz>            merge a pack into $DASH_DIR (replaces those builds' files and result lines)
#
# Re-sim a few builds: hydrate, run <ids>, publish <ids>. Add a preset: write req/<id>.*, then the same. Remove one:
# delete req/<id>.*, hydrate, publish. A text or page-data change: hydrate, publish. Everything stale (a shared sim
# input changed): run, publish, with no hydrate. Then commit ui/dps_report/data, provenance.json and raw/ together.
#
# Ids default to every request in $DASH_DIR/req. Builds run longest first (rogues, then hunters ...), so the last
# minutes of each step are not one slow build alone. Every sim splits its iterations into DASH_SPLITS parts (default
# 4, patch 0005) run in parallel, so the raw outputs don't depend on the machine; P=4 processes keep a 4-core machine
# busy to the end.
set -euo pipefail
T="$(cd "$(dirname "$0")" && pwd)"
R="$(cd "$T/../.." && pwd)"
export DASH_DIR="${DASH_DIR:-$T}" DASH_SPLITS="${DASH_SPLITS:-4}"
cd "$DASH_DIR"
P=${P:-4}
cmd=${1:-}; shift || true

ids() {  # the given ids or every request, most expensive first
	python3 - "$@" <<'EOF'
import glob, json, os, sys
ids = sys.argv[1:] or sorted(os.path.basename(p)[:-10] for p in glob.glob('req/*.meta.json'))
# core-minutes per build for every step, from the 2026-10-16 runs (a duty preset gets run_build.py only)
W = {'rogue': 6.6, 'shaman': 6.3, 'warrior': 5.6, 'hunter': 4.9, 'mage': 3.8, 'paladin': 3.4, 'warlock': 3.3, 'druid': 2.5, 'priest': 1.7}
cost = lambda i: W.get(i.split('__')[0], 4) * (0.35 if json.load(open(f'req/{i}.meta.json')).get('duty') else 1)
print(' '.join(sorted(ids, key=lambda i: (-cost(i), i))))
EOF
}

case "$cmd" in
hydrate)
	if [ "$PWD" != "$T" ]; then  # a separate $DASH_DIR reads the inputs from this checkout
		for f in req forever_tt_cache fortt_ids.json talent_ids.json asm_apls bin; do [ -e "$f" ] || ln -s "$T/$f" "$f"; done
	fi
	rm -rf out names weights mana rv asm abl data
	python3 $T/rawstore.py unpack
	python3 $T/build_data.py >build_data.log || { tail -20 build_data.log; exit 1; }
	python3 $T/rawstore.py check "$@"
	;;
clean)
	python3 $T/rawstore.py clean "$@"
	;;
publish)
	if [ $# -gt 0 ]; then set -- $(python3 $T/rawstore.py variants "$@"); fi  # what run <ids> re-simmed
	python3 $T/rawstore.py prune  # removed presets: their outputs, packs and sample fights
	rm -rf data
	python3 $T/build_data.py >build_data.log || { tail -20 build_data.log; exit 1; }
	# the sample fights of the re-simmed builds; the event lookups always cover every build's fights
	if [ $# -gt 0 ]; then
		python3 $T/events.py --data "$DASH_DIR/data" --out "$R/ui/dps_report/data/events" "$@"
	else
		python3 $T/events.py --data "$DASH_DIR/data" --out "$R/ui/dps_report/data/events" --lookups-only
	fi
	python3 $T/rawstore.py copy
	python3 $T/changes.py build  # the Changes tab: the live update against the report's new data
	if [ $# -gt 0 ]; then (cd "$R" && node tools/dps_report/stale.mjs --record "$@" --iterations 20000 ${SIM_COMMIT:+--commit $SIM_COMMIT}); fi
	# only the re-simmed builds' packs: the others are already the outputs they were made from
	if [ $# -gt 0 ]; then python3 $T/rawstore.py pack "$@"; fi
	echo "next: give the live update public lines for what changed (changes.py check), check the report (make dps-report-artifact), commit ui/dps_report, tools/dps_report/raw and tools/dps_report/changes together, push, then deploy"
	;;
plan | run | resume)
	if [ $# -gt 0 ]; then set -- $(python3 $T/rawstore.py variants "$@"); fi
	ALL=$(ids "$@")
	if [ "$cmd" = run ]; then python3 $T/rawstore.py clean $ALL; fi
	BASE=$(for i in $ALL; do python3 -c "import json,sys; sys.exit(0 if json.load(open('req/$i.meta.json')).get('duty') else 1)" || echo $i; done)
	MANA=$(for i in $ALL; do case $i in warrior__* | rogue__*) ;; *) echo $i ;; esac; done)
	python3 - $BASE <<EOF
import os, sys
sys.path.insert(0, '$T')
import run_races, run_assumptions, weights, json
base = sys.argv[1:]
w = sum(len(weights.stat_list(json.load(open(f'req/{b}.meta.json')))[1]) + 2 for b in base)
r = sum(1 for b in base if not json.load(open(f'req/{b}.meta.json')).get('encounter') for _ in run_races.tasks(b))
a = sum(1 for b in base for _ in run_assumptions.tasks(b))
import glob
W = {'rogue': 6.6, 'shaman': 6.3, 'warrior': 5.6, 'hunter': 4.9, 'mage': 3.8, 'paladin': 3.4, 'warlock': 3.3, 'druid': 2.5, 'priest': 1.7}
cost = lambda i: W.get(i.split('__')[0], 4) * (0.35 if json.load(open(f'req/{i}.meta.json')).get('duty') else 1)
every = [os.path.basename(p)[:-10] for p in glob.glob('req/*.meta.json')]
# the whole 66-build set: about 380 core-minutes with these scripts (2026-10-16 timings, scaled; the weights step
# saves about 30% with paired errors, since melee hit and crit steps still need 40k)
eta = 380 * sum(cost(i) for i in '$ALL'.split()) / sum(cost(i) for i in every) / os.cpu_count()
print(f'plan: $(echo $ALL | wc -w) builds; weights {w} sims, races {r}, assumption combos {a}, mana $(echo $MANA | wc -w) builds; '
      f'about {eta:.0f} min on {os.cpu_count()} cores')
EOF
	[ "$cmd" = plan ] && exit 0
	# resume skips what a stopped run finished: run_build.py and mana_curves.py --force redo a build whatever is there,
	# and ablate.py redoes every build it is given; the other steps skip done work themselves
	RB=$ALL MANA_FORCE=--force ABL=$BASE
	if [ "$cmd" = resume ]; then
		RB=$(for i in $ALL; do [ -e out/$i.json ] || echo $i; done)
		MANA_FORCE=
		ABL=$(python3 -c "import json,os,sys; d=json.load(open('abl/ablate.json'))['builds'] if os.path.exists('abl/ablate.json') else {}; print(' '.join(b for b in sys.argv[1:] if b not in d))" $BASE)
	fi
	echo "== run_build $(date -u +%H:%M:%S)"
	[ -n "$RB" ] && echo $RB | tr ' ' '\n' | xargs -P$P -n1 python3 $T/run_build.py
	echo "== weights $(date -u +%H:%M:%S)"
	echo $BASE | tr ' ' '\n' | xargs -P$P -n1 python3 $T/weights.py
	echo "== races $(date -u +%H:%M:%S)"
	python3 $T/run_races.py $BASE | tail -3
	echo "== mana $(date -u +%H:%M:%S)"
	[ -n "$MANA" ] && PAR=$P python3 $T/mana_curves.py $MANA_FORCE $MANA | tail -3
	echo "== ablate $(date -u +%H:%M:%S)"
	[ -n "$ABL" ] && ABL_MERGE=1 PAR=$P python3 $T/ablate.py $ABL | tail -1
	echo "== assumptions $(date -u +%H:%M:%S)"
	PAR=$P python3 $T/run_assumptions.py $BASE | tail -2
	echo "== done $(date -u +%H:%M:%S)"
	echo "next: pipeline.sh publish $*"
	;;
shard)
	k=$1 n=$2; shift 2
	python3 - $k $n $(ids "$@") <<'EOF'
import json, sys
k, n, ids = int(sys.argv[1]), int(sys.argv[2]), sys.argv[3:]
W = {'rogue': 6.6, 'shaman': 6.3, 'warrior': 5.6, 'hunter': 4.9, 'mage': 3.8, 'paladin': 3.4, 'warlock': 3.3, 'druid': 2.5, 'priest': 1.7}
load, out = [0.0] * n, [[] for _ in range(n)]
for i in ids:  # already most expensive first: give each build to the least loaded shard
    c = W.get(i.split('__')[0], 4) * (0.35 if json.load(open(f'req/{i}.meta.json')).get('duty') else 1)
    j = load.index(min(load)); load[j] += c; out[j].append(i)
print(' '.join(out[k]))
EOF
	;;
pack)
	tgz=$1; shift
	tmp=$(mktemp -d)
	python3 - "$tmp" "$@" <<'EOF'
import json, os, shutil, sys
tmp, ids = sys.argv[1], set(sys.argv[2:])
for d, ext in (('out', '.json'), ('names', '.tsv'), ('weights', '.json'), ('mana', '.json')):
    for i in ids:
        if os.path.exists(f'{d}/{i}{ext}'):
            os.makedirs(f'{tmp}/{d}', exist_ok=True); shutil.copy(f'{d}/{i}{ext}', f'{tmp}/{d}/')
for f in ('rv/results.jsonl', 'asm/results.jsonl'):
    if os.path.exists(f):
        os.makedirs(os.path.dirname(f'{tmp}/{f}'), exist_ok=True)
        open(f'{tmp}/{f}', 'w').writelines(l for l in open(f) if json.loads(l)['build'] in ids)
if os.path.exists('abl/ablate.json'):
    a = json.load(open('abl/ablate.json')); a['builds'] = {b: x for b, x in a['builds'].items() if b in ids}
    os.makedirs(f'{tmp}/abl'); json.dump(a, open(f'{tmp}/abl/ablate.json', 'w'))
EOF
	tar -czf "$tgz" -C "$tmp" . && rm -rf "$tmp" && echo "packed $# builds into $tgz"
	;;
unpack)
	tmp=$(mktemp -d); tar -xzf "$1" -C "$tmp"
	python3 - "$tmp" <<'EOF'
import glob, json, os, shutil, sys
tmp = sys.argv[1]
ids = {os.path.basename(p)[:-5] for p in glob.glob(f'{tmp}/out/*.json')}
for d in ('out', 'names', 'weights', 'mana'):
    for p in glob.glob(f'{tmp}/{d}/*'):
        os.makedirs(d, exist_ok=True); shutil.copy(p, d)
for f in ('rv/results.jsonl', 'asm/results.jsonl'):
    if os.path.exists(f'{tmp}/{f}'):
        new = open(f'{tmp}/{f}').readlines(); got = {json.loads(l)['build'] for l in new}
        old = [l for l in open(f)] if os.path.exists(f) else []
        os.makedirs(os.path.dirname(f), exist_ok=True)
        open(f, 'w').writelines([l for l in old if json.loads(l)['build'] not in got] + new)
if os.path.exists(f'{tmp}/abl/ablate.json'):
    a = json.load(open(f'{tmp}/abl/ablate.json'))
    if os.path.exists('abl/ablate.json'):
        old = json.load(open('abl/ablate.json')); old['builds'].update(a['builds']); a = old
    os.makedirs('abl', exist_ok=True); json.dump(a, open('abl/ablate.json', 'w'), indent=1)
print(f'merged {len(ids)} builds')
EOF
	rm -rf "$tmp"
	;;
*)
	sed -n '2,13p' "$0"; exit 1 ;;
esac
