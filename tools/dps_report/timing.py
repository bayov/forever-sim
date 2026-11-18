#!/usr/bin/env python3
"""timing.py [log]: where a pipeline.sh run spent its time, from the log timing_cli.py wrote (default: the newest
$DASH_DIR/timing/*.jsonl). pipeline.sh run prints it at the end.

  steps      each step's wall time, the CPU time of its sims, and how many cores those kept busy on average (the CPU
             time of the step scripts themselves isn't counted: a step whose busy figure is well under P is waiting on
             its slowest builds or on its own Python)
  sim kinds  the costliest kinds of sim: a step's sims grouped by binary, iterations, fight length and target count
  phases     hydrate, publish and the binary build (tankheal.py bins), which are no sims: their wall time, from
             timing/phases.jsonl (the latest of each) and the run's own log
  classes    core-minutes per build of each class, every step, a duty or pet variant counted as 0.35 of a base build:
             the same unit as the cost table W in pipeline.sh, which orders and shards builds
"""
import collections, glob, json, os, sys, time

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
        elif 'phase' not in r:
            sims.append(r)
    return start, marks, sims


def owner(step, names):
    """The mark a sim's step counts under: tankheal.py's steps (tank, heal, tank_races, heal_races) run inside the
    tankheal mark."""
    if step in names or 'tankheal' not in names:
        return step
    return 'tankheal' if step.split('_')[0] in ('tank', 'heal') else step


def phases(logs):
    """The latest wall time of each phase outside the sims: timing/phases.jsonl and the run's log."""
    got = {}
    for path in [os.path.join(D, 'timing', 'phases.jsonl')] + list(logs):
        if os.path.exists(path):
            for line in open(path):
                r = json.loads(line)
                if 'phase' in r and r['t'] >= got.get(r['phase'], {}).get('t', 0):
                    got[r['phase']] = r
    return got


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
    names = {m['mark'] for m in marks}
    for s in sims:
        cpu[owner(s['step'], names)] += s['cpu']
        n[owner(s['step'], names)] += 1
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
        if m == 'tankheal':  # the bench steps inside it
            sub = collections.defaultdict(lambda: [0, 0.0])
            for s in sims:
                if owner(s['step'], names) == m and s['step'] != m:
                    sub[s['step']][0] += 1
                    sub[s['step']][1] += s['cpu']
            for k, (cnt, c2) in sorted(sub.items()):
                print(f'  {k:<10} {"":>8} {"":>6} {c2 / 60:>9.1f} {c2 / max(total_cpu, 1e-9):>6.0%} {cnt:>6}')
    ph = phases(logs[-1:])
    if ph:
        print('\nPhases outside the sim steps (the latest of each):')
        for k, r in sorted(ph.items(), key=lambda kv: kv[1]['t']):
            print(f'  {k:<10} {r["wall"] / 60:>6.1f} min wall   {time.strftime("%Y-%m-%d %H:%M", time.gmtime(r["t"]))} UTC')

    kinds = collections.defaultdict(lambda: [0, 0.0])
    for s in sims:
        k = (s['step'], s['bin'], s.get('iters', 0), s.get('dur', 0), s.get('targets', 0), s.get('replay', False),
             s.get('role', ''), s.get('boss', ''), s.get('reqs', 0))
        kinds[k][0] += 1
        kinds[k][1] += s['cpu']
    print('\nCostliest kinds of sim (a step\'s sims by binary, iterations, fight length, targets):')
    for (step, b, it, dur, tg, rp, role, boss, nreq), (cnt, c) in sorted(kinds.items(), key=lambda kv: -kv[1][1])[:15]:
        if b.startswith('dashstats'):
            what = 'dashstats'
        elif b in ('tankbench', 'healbench'):
            what = f'{b} {role}{" " + boss if boss else ""} {it} it x {nreq} req' + (f' x {dur} s' if dur else '')
        else:
            what = f'{b} {it} it x {dur} s' + (f', {tg} targets' if tg > 1 else '') + (', replay' if rp else '')
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
