#!/bin/bash
# pipeline.sh: every sim step of the DPS report for a set of builds, in one command.
#
#   pipeline.sh plan  <ids...>          count every step's sims and check each script can enumerate its work (seconds)
#   pipeline.sh run   <ids...>          plan, then the sim steps; prints a timestamp per step
#   pipeline.sh shard <k> <n> <ids...>  the ids of shard k of n (0-based), balanced by estimated cost
#   pipeline.sh pack  <tgz> <ids...>    these builds' raw outputs, to merge on another machine
#   pipeline.sh unpack <tgz>            merge a pack into $DASH_DIR (replaces those builds' files and result lines)
#
# Ids default to every request in $DASH_DIR/req. Builds run longest first (rogues, then hunters ...), so the last
# minutes of each step are not one slow build alone. Every sim splits its iterations over all cores itself
# (sim/core/sim_concurrent.go), so P=4 processes keep a 4-core machine busy to the end.
set -euo pipefail
T="$(cd "$(dirname "$0")" && pwd)"
cd "${DASH_DIR:-$T}"
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
plan | run)
	ALL=$(ids "$@")
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
	echo "== run_build $(date -u +%H:%M:%S)"
	echo $ALL | tr ' ' '\n' | xargs -P$P -n1 python3 $T/run_build.py
	echo "== weights $(date -u +%H:%M:%S)"
	echo $BASE | tr ' ' '\n' | xargs -P$P -n1 python3 $T/weights.py
	echo "== races $(date -u +%H:%M:%S)"
	python3 $T/run_races.py $BASE | tail -3
	echo "== mana $(date -u +%H:%M:%S)"
	[ -n "$MANA" ] && PAR=$P python3 $T/mana_curves.py --force $MANA | tail -3
	echo "== ablate $(date -u +%H:%M:%S)"
	ABL_MERGE=1 PAR=$P python3 $T/ablate.py $BASE | tail -1
	echo "== assumptions $(date -u +%H:%M:%S)"
	PAR=$P python3 $T/run_assumptions.py $BASE | tail -2
	echo "== done $(date -u +%H:%M:%S)"
	echo "next: python3 $T/build_data.py, then python3 $T/events.py --data \"${DASH_DIR:-$T}/data\" <these ids> (the events view), then merge.mjs"
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
