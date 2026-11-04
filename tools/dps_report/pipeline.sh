#!/bin/bash
# pipeline.sh: every sim step of the DPS report for a set of builds, in one command.
#
#   pipeline.sh hydrate [ids...]        empty $DASH_DIR's outputs, load every build's raw pack (raw/<id>.tgz), rebuild
#                                       the datasets and check their sim numbers against the published report; a
#                                       build outside [ids] that differs stops it (orchestration/dash_raw_store.md)
#   pipeline.sh plan  <ids...>          count every step's sims and check each script can enumerate its work (seconds)
#   pipeline.sh run   <ids...>          add the ids' duty and encounter variants, clean them, plan, then the sim
#                                       steps; prints a timestamp per step, and at the end where the time went (timing.py)
#   pipeline.sh resume <ids...>         run without the clean, to finish a run that stopped (every step skips done work)
#   pipeline.sh clean <ids...>          drop these builds from every raw output, so every step re-sims them
#   pipeline.sh probe [ids...]          which stale builds really moved (probe.py check): each one's sims at a tenth of
#                                       their iterations against the probe in its raw pack; prints what to re-sim, and
#                                       publish records the ones it cleared (default ids: what stale.mjs lists)
#   pipeline.sh publish [ids...]        build_data.py, the sample fights of [ids] and their variants (events.py),
#                                       every dataset into ui/dps_report/data, and provenance and raw packs for them;
#                                       after a probe, it checks the audit first and records the cleared builds
#   pipeline.sh shard <k> <n> <ids...>  the ids of shard k of n (0-based), balanced by estimated cost
#   pipeline.sh pack  <tgz> <ids...>    these builds' raw outputs, to merge on another machine (a shard's hand-back)
#   pipeline.sh unpack <tgz>            merge a pack into $DASH_DIR (replaces those builds' files and result lines)
#   pipeline.sh tankheal [tank|heal]    the Tanking and Healing modes' sims alone (tankheal.py run: tankbench and
#                                       healbench, every preset, seven bosses, races; then publish)
#
# The Tanking and Healing modes (tankheal.py) ride along: run with no ids (or TANKHEAL=1) runs their sims after the
# DPS steps, hydrate unpacks their raw packs (raw/bench/<mode>.tgz), publish builds their datasets (tank_*, heal_*)
# before the copy and, for a mode simmed since, records it (stale.mjs --record-bench) and packs it.
#
# After a code or data change (a game update), stale.mjs says what may have moved: hydrate, probe, run <what probe
# names>, publish <the same ids>. Re-sim a few builds: hydrate, run <ids>, publish <ids>. Add a preset: write req/<id>.*, then the same. Remove one:
# delete req/<id>.*, hydrate, publish. A text or page-data change: hydrate, publish. Everything stale (a shared sim
# input changed): run, publish, with no hydrate. Then commit ui/dps_report/data, provenance.json, raw/ and changes/
# together.
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
# core-minutes per build for every step, from the 2026-10-16 runs (a duty preset or pet variant gets run_build.py only)
W = {'rogue': 6.6, 'shaman': 6.3, 'warrior': 5.6, 'hunter': 4.9, 'mage': 3.8, 'paladin': 3.4, 'warlock': 3.3, 'druid': 2.5, 'priest': 1.7}
cost = lambda i: W.get(i.split('__')[0], 4) * (0.35 if any(json.load(open(f'req/{i}.meta.json')).get(k) for k in ('duty', 'pet')) else 1)
print(' '.join(sorted(ids, key=lambda i: (-cost(i), i))))
EOF
}

case "$cmd" in
hydrate)
	if [ "$PWD" != "$T" ]; then  # a separate $DASH_DIR reads the inputs from this checkout
		for f in req forever_tt_cache fortt_ids.json talent_ids.json asm_apls bin; do [ -e "$f" ] || ln -s "$T/$f" "$f"; done
	fi
	rm -rf out names weights mana rv asm abl probe probe_check data tank heal
	python3 $T/rawstore.py unpack
	python3 $T/tankheal.py unpack
	python3 $T/build_data.py >build_data.log || { tail -20 build_data.log; exit 1; }
	python3 $T/tankheal.py build >tankheal.log || { tail -20 tankheal.log; exit 1; }
	python3 $T/rawstore.py check "$@"
	;;
clean)
	python3 $T/rawstore.py clean "$@"
	;;
probe)
	python3 $T/probe.py check "$@"
	;;
publish)
	if [ $# -gt 0 ]; then set -- $(python3 $T/rawstore.py variants "$@"); fi  # what run <ids> re-simmed
	python3 $T/probe.py audit "$@"  # after a probe check: its audit builds must come out unchanged (exits 1 if not)
	PASSED=$(python3 $T/probe.py passed "$@")  # the builds it cleared, recorded below without sims
	python3 $T/rawstore.py prune  # removed presets: their outputs, packs and sample fights
	rm -rf data
	python3 $T/build_data.py >build_data.log || { tail -20 build_data.log; exit 1; }
	# the sample fights of the re-simmed builds; the event lookups always cover every build's fights
	if [ $# -gt 0 ]; then
		python3 $T/events.py --data "$DASH_DIR/data" --out "$R/ui/dps_report/data/events" "$@"
	else
		python3 $T/events.py --data "$DASH_DIR/data" --out "$R/ui/dps_report/data/events" --lookups-only
	fi
	python3 $T/tankheal.py build >tankheal.log || { tail -20 tankheal.log; exit 1; }  # the tank_* and heal_* datasets
	python3 $T/rawstore.py copy
	if [ $# -gt 0 ]; then (cd "$R" && node tools/dps_report/stale.mjs --record "$@" --iterations 20000 ${SIM_COMMIT:+--commit $SIM_COMMIT}); fi
	BENCHED=$(python3 $T/tankheal.py fresh)  # the Tanking and Healing modes simmed since the last publish
	if [ -n "$BENCHED" ]; then
		(cd "$R" && node tools/dps_report/stale.mjs --record-bench $BENCHED --iterations 1000 ${SIM_COMMIT:+--commit $SIM_COMMIT})
		python3 $T/tankheal.py pack && python3 $T/tankheal.py fresh --clear
	fi
	if [ -n "$PASSED" ]; then (cd "$R" && node tools/dps_report/stale.mjs --record $PASSED --by probe ${SIM_COMMIT:+--commit $SIM_COMMIT}); fi
	rm -f probe_audit.json
	# only the re-simmed builds' packs: the others are already the outputs they were made from
	if [ $# -gt 0 ]; then python3 $T/rawstore.py pack "$@"; fi
	python3 $T/changes.py build  # the Changes tab: the live update against the report's new data
	echo "next: give the live update public lines for what changed (changes.py check, then changes.py build), check the report (make dps-report-artifact), commit ui/dps_report, tools/dps_report/raw and tools/dps_report/changes together, push, then deploy"
	;;
plan | run | resume)
	if [ $# -gt 0 ]; then set -- $(python3 $T/rawstore.py variants "$@"); fi
	ALL=$(ids "$@")
	if [ "$cmd" = run ]; then python3 $T/rawstore.py clean $ALL; fi
	BASE=$(for i in $ALL; do python3 -c "import json,sys; m = json.load(open('req/$i.meta.json')); sys.exit(0 if m.get('duty') or m.get('pet') else 1)" || echo $i; done)
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
cost = lambda i: W.get(i.split('__')[0], 4) * (0.35 if any(json.load(open(f'req/{i}.meta.json')).get(k) for k in ('duty', 'pet')) else 1)
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
	# what every sim costs: the binaries run through timing_cli.py, which logs each call's CPU and wall time to
	# timing/<start>.jsonl (with a line per step below); timing.py sums it up at the end
	REAL_BIN=$(cd "${DASH_BIN:-$DASH_DIR/bin}" && pwd)
	mkdir -p timing/bin
	for b in wowsimcli wowsimcli_r wowsimcli_assume dashstats; do
		rm -f timing/bin/$b
		if [ -e "$REAL_BIN/$b" ]; then ln -s "$T/timing_cli.py" timing/bin/$b; fi
	done
	export TIMING_REAL_BIN=$REAL_BIN DASH_BIN=$DASH_DIR/timing/bin TIMING_LOG=$DASH_DIR/timing/$(date -u +%Y%m%dT%H%M%SZ).jsonl
	python3 -c 'import json, os, sys; print(json.dumps({"start": sys.argv[1], "P": int(sys.argv[2]), "cores": os.cpu_count(), "ids": sys.argv[3].split(), "base": sys.argv[4].split()}))' \
		"$cmd" "$P" "$ALL" "$BASE" >"$TIMING_LOG"
	mark() {  # a step starts: its time on screen and in the log, and the step the log's sims count under
		echo "== $1 $(date -u +%H:%M:%S)"
		export TIMING_STEP=$1
		echo "{\"mark\": \"$1\", \"t\": $(date +%s.%N)}" >>"$TIMING_LOG"
	}
	mark run_build
	[ -n "$RB" ] && echo $RB | tr ' ' '\n' | xargs -P$P -n1 python3 $T/run_build.py
	mark weights
	echo $BASE | tr ' ' '\n' | xargs -P$P -n1 python3 $T/weights.py
	mark races
	python3 $T/run_races.py $BASE | tail -3
	mark mana
	[ -n "$MANA" ] && PAR=$P python3 $T/mana_curves.py $MANA_FORCE $MANA | tail -3
	mark ablate
	[ -n "$ABL" ] && ABL_MERGE=1 PAR=$P python3 $T/ablate.py $ABL | tail -1
	mark assumptions
	PAR=$P python3 $T/run_assumptions.py $BASE | tail -2
	mark probe  # these sims at a tenth of their iterations, kept in the packs for the next probe check
	PROBE_P=$P python3 $T/probe.py run $ALL
	if [ $# -eq 0 ] || [ -n "${TANKHEAL:-}" ]; then  # a full run (or TANKHEAL=1): the Tanking and Healing modes too
		mark tankheal
		TH=tank\ heal
		if [ "$cmd" = resume ]; then TH=$(for m in tank heal; do [ -e $m/.fresh ] || echo $m; done); fi
		[ -n "$TH" ] && TANKHEAL_BIN=$REAL_BIN python3 $T/tankheal.py run $TH $([ "$cmd" = resume ] && echo --resume)
	fi
	mark done
	python3 $T/timing.py "$TIMING_LOG"
	echo "next: pipeline.sh publish $*"
	;;
tankheal)
	TANKHEAL_BIN=${TANKHEAL_BIN:-${DASH_BIN:-$DASH_DIR/bin}} python3 $T/tankheal.py run "$@"
	echo "next: pipeline.sh publish"
	;;
shard)
	k=$1 n=$2; shift 2
	python3 - $k $n $(ids "$@") <<'EOF'
import json, sys
k, n, ids = int(sys.argv[1]), int(sys.argv[2]), sys.argv[3:]
W = {'rogue': 6.6, 'shaman': 6.3, 'warrior': 5.6, 'hunter': 4.9, 'mage': 3.8, 'paladin': 3.4, 'warlock': 3.3, 'druid': 2.5, 'priest': 1.7}
load, out = [0.0] * n, [[] for _ in range(n)]
for i in ids:  # already most expensive first: give each build to the least loaded shard
    c = W.get(i.split('__')[0], 4) * (0.35 if any(json.load(open(f'req/{i}.meta.json')).get(k) for k in ('duty', 'pet')) else 1)
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
for d, ext in (('out', '.json'), ('names', '.tsv'), ('weights', '.json'), ('mana', '.json'), ('probe', '.json')):
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
for d in ('out', 'names', 'weights', 'mana', 'probe'):
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
	sed -n '2,22p' "$0"; exit 1 ;;
esac
