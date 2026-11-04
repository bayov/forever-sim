#!/usr/bin/env python3
"""timing.py [log]: where a pipeline.sh run spent its time, from the log timing_cli.py wrote (default: the newest
$DASH_DIR/timing/*.jsonl). pipeline.sh run prints it at the end.

  steps      each step's wall time, the CPU time of its sims, and how many cores those kept busy on average (the CPU
             time of the step scripts themselves isn't counted: a step whose busy figure is well under P is waiting on
             its slowest builds or on its own Python)
  sim kinds  the costliest kinds of sim: a step's sims grouped by binary, iterations, fight length and target count
  classes    core-minutes per build of each class, every step, a duty or pet variant counted as 0.35 of a base build:
             the same unit as the cost table W in pipeline.sh, which orders and shards builds
"""
import collections, glob, json, os, sys

sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
from dashpaths import D  # noqa: E402

VARIANT = 0.35  # a duty or pet variant's share of a base build's cost, as in pipeline.sh


def load(path):
    start, marks, sims = {}, [], []
    for line in open(path):
        r = json.loads(line)
        if 'start' in r:
            start = r
        elif 'mark' in r:
            marks.append(r)
        else:
            sims.append(r)
    return start, marks, sims


def main():
    logs = sys.argv[1:] or sorted(glob.glob(os.path.join(D, 'timing', '*.jsonl')))[-1:]
    if not logs:
        sys.exit('no timing log: pipeline.sh run writes one to $DASH_DIR/timing/')
    start, marks, sims = load(logs[-1])
    cores = start.get('cores') or os.cpu_count()
    walls = {a['mark']: b['t'] - a['t'] for a, b in zip(marks, marks[1:])}
    total_wall = marks[-1]['t'] - marks[0]['t'] if len(marks) > 1 else 0
    cpu = collections.Counter()
    n = collections.Counter()
    for s in sims:
        cpu[s['step']] += s['cpu']
        n[s['step']] += 1
    total_cpu = sum(cpu.values())
    ids, base = start.get('ids', []), set(start.get('base', []))
    print(f'{logs[-1]}: {len(ids)} builds ({len(base)} base), P={start.get("P")} on {cores} cores; '
          f'{total_wall / 60:.1f} min wall, {total_cpu / 60:.1f} core-min in {len(sims)} sims'
          + (f', {sum(1 for s in sims if s.get("rc"))} failed' if any(s.get('rc') for s in sims) else ''))

    print(f'\n{"step":<12} {"wall min":>8} {"share":>6} {"core-min":>9} {"share":>6} {"sims":>6} {"busy cores":>10}')
    for m in [m['mark'] for m in marks[:-1]]:
        w, c = walls.get(m, 0), cpu.get(m, 0)
        print(f'{m:<12} {w / 60:>8.1f} {w / max(total_wall, 1e-9):>6.0%} {c / 60:>9.1f} {c / max(total_cpu, 1e-9):>6.0%} '
              f'{n.get(m, 0):>6} {c / max(w, 1e-9):>6.1f}/{cores}')

    kinds = collections.defaultdict(lambda: [0, 0.0])
    for s in sims:
        k = (s['step'], s['bin'], s.get('iters', 0), s.get('dur', 0), s.get('targets', 0), s.get('replay', False))
        kinds[k][0] += 1
        kinds[k][1] += s['cpu']
    print('\nCostliest kinds of sim (a step\'s sims by binary, iterations, fight length, targets):')
    for (step, b, it, dur, tg, rp), (cnt, c) in sorted(kinds.items(), key=lambda kv: -kv[1][1])[:15]:
        what = 'dashstats' if b.startswith('dashstats') else f'{b} {it} it x {dur} s' + (f', {tg} targets' if tg > 1 else '') + (', replay' if rp else '')
        print(f'  {c / max(total_cpu, 1e-9):>4.0%}  {c / 60:>7.1f} core-min  {cnt:>5} sims  {c / cnt:>6.2f} s each  {step}: {what}')

    if ids:
        units = collections.Counter()
        for i in ids:
            units[i.split('__')[0]] += 1 if i in base else VARIANT
        by_cls = collections.Counter()
        for s in sims:
            if s.get('cls'):
                by_cls[s['cls'].removeprefix('Class').lower()] += s['cpu']
        print('\nCore-minutes per base build, every step (pipeline.sh W; a duty or pet variant counts 0.35):')
        for cls in sorted(units, key=lambda c: -by_cls[c] / units[c]):
            print(f'  {cls:<8} {by_cls[cls] / 60 / units[cls]:>6.1f}   ({by_cls[cls] / 60:.1f} core-min over {units[cls]:g} build units)')


if __name__ == '__main__':
    main()
