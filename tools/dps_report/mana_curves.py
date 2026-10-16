#!/usr/bin/env python3
"""mana_curves.py [ids] : how each mana-using preset's mana holds up over a long fight.

Every build whose class uses mana (all but warriors and rogues. Feral cat counts because it powershifts
with mana) is simmed with its own request ($DASH_DIR/req/<id>.json: consumes, mana potions, buffs and
rotation as run_build.py runs them), but with a fixed 480s fight and no duration variation.

  - FIGHTS (default 100) one-fight runs with seeds 1001.., each with a debug log. We read the player's
    mana from the log's "Spent / Gained N mana (old --> new)" lines and sample it every 5s.
  - One 5,000-fight run for the sim's own mana metrics: time spent out of mana (secondsOomAvg), time to
    OOM (Tto, the first cast that failed for mana, extrapolated from the spend rate when none failed)
    and mana left at the end.

Many rotations gate spells on mana %, so they never fail a cast and Tto stays high. The mana curve is the
honest signal for those.

Writes $DASH_DIR/mana/<id>.json per build and skips builds that already have one (--force redoes them).
`python3 mana_curves.py --data` only turns mana/*.json into data/mana.json and data/mana_summary.json.
build_data.py does the same as part of its run. PAR (default 8) is the number of sims run side by side.
"""
import json, math, os, re, sys
from concurrent.futures import ThreadPoolExecutor

sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
from run_build import run, D, LINE

DUR = 480
STEP = 5
LOW = 15  # "runs low" means mana below this % of max. Most rotations start holding spells around 15 to 20%.
FIGHTS = int(os.environ.get('FIGHTS', 100))
AGG_ITERS = 5000
PAR = int(os.environ.get('PAR', 8))
NO_MANA = ('Warrior', 'Rogue')
MANA = re.compile(r'^(Spent|Gained) ([\d.]+) mana from .* \(([\d.]+) --> ([\d.]+)\)')


def mana_builds():
    ids = []
    for f in sorted(os.listdir(os.path.join(D, 'req'))):
        if f.endswith('.meta.json'):
            meta = json.load(open(os.path.join(D, 'req', f)))
            if meta['class'] not in NO_MANA:
                ids.append(meta['id'])
    return ids


def long_fight(req):
    req = json.loads(json.dumps(req))
    req['encounter']['duration'] = DUR
    req['encounter']['durationVariation'] = 0
    return req


def fight(req, seed):
    """One logged fight: the player's mana after each change, plus the mana it started with."""
    r = run(req, 1, seed, debug=True)
    pts, start, spent, gained = [], None, 0.0, 0.0
    for line in r['logs'].split('\n'):
        if ' mana from ' not in line:
            continue
        m = LINE.match(line)
        if not m or m.group(2).strip() != '(#1)':
            continue
        mm = MANA.match(m.group(3))
        if not mm:
            continue
        if start is None:
            start = float(mm.group(3))
        if mm.group(1) == 'Spent':
            spent += float(mm.group(2))
        else:
            gained += float(mm.group(4)) - float(mm.group(3))
        pts.append((float(m.group(1)), float(mm.group(4))))
    return {'start': start, 'pts': pts, 'spent': spent, 'gained': gained}


def sample(pts, start, mx):
    """Mana % of max at every STEP, holding the last value."""
    out, i, cur = [], 0, start
    for k in range(DUR // STEP + 1):
        t = k * STEP
        while i < len(pts) and pts[i][0] <= t:
            cur = pts[i][1]
            i += 1
        out.append(min(100.0, 100 * cur / mx))
    return out


def sim_build(bid):
    req = long_fight(json.load(open(os.path.join(D, 'req', bid + '.json'))))
    with ThreadPoolExecutor(PAR) as ex:
        fights = list(ex.map(lambda s: fight(req, s), range(1001, 1001 + FIGHTS)))
    agg = run(req, AGG_ITERS, 1)
    p = agg['raidMetrics']['parties'][0]['players'][0]
    logged = [f for f in fights if f['start'] is not None]
    if not logged:
        return {'id': bid, 'fights': 0, 'note': 'no player mana changes in the log'}
    # Every unit starts a fight at full mana, so the first change's old value is the max.
    mx = max(f['start'] for f in logged)
    series, low = [], []
    for f in logged:
        series.append([round(v, 2) for v in sample(f['pts'], f['start'], mx)])
        low.append(next((round(t, 2) for t, v in f['pts'] if t >= 0 and 100 * v / mx < LOW), None))
    return {'id': bid, 'dur': DUR, 'step': STEP, 'fights': len(logged), 'maxMana': mx, 'series': series, 'lowAt': low,
            'spentPerSec': sum(f['spent'] for f in logged) / len(logged) / DUR,
            'gainedPerSec': sum(f['gained'] for f in logged) / len(logged) / DUR,
            'oomSecs': p.get('secondsOomAvg', 0), 'tto': p['tto']['avg'], 'manaEnd': p['manaEnd']['avg'], 'aggIters': AGG_ITERS}


def q(vals, f):
    """Nearest-rank quantile, as run_build.py takes DPS percentiles."""
    return vals[min(len(vals) - 1, int(f * len(vals)))]


def r1(v):
    return None if v is None or math.isinf(v) else round(v, 1)


def datasets():
    """(mana, mana_summary) rows from $DASH_DIR/mana/*.json."""
    curve, summary = [], []
    folder = os.path.join(D, 'mana')
    if not os.path.isdir(folder):
        return curve, summary
    for f in sorted(os.listdir(folder)):
        b = json.load(open(os.path.join(folder, f)))
        if not b.get('fights'):
            continue
        cols = list(zip(*b['series']))
        for k, col in enumerate(cols):
            s = sorted(col)
            curve.append({'build': b['id'], 't': k * b['step'], 'pct': r1(sum(s) / len(s)), 'p25': r1(q(s, 0.25)), 'p75': r1(q(s, 0.75))})
        low = sorted(math.inf if v is None else v for v in b['lowAt'])
        at = lambda t: r1(sum(cols[t // b['step']]) / len(cols[t // b['step']]))
        summary.append({'build': b['id'], 'fights': b['fights'], 'dur': b['dur'], 'maxMana': round(b['maxMana']),
                        'lowAt': r1(q(low, 0.5)), 'lowAtP25': r1(q(low, 0.25)), 'lowAtP75': r1(q(low, 0.75)),
                        'lowPct': LOW, 'shareLow': r1(100 * sum(1 for v in low if not math.isinf(v)) / len(low)),
                        'pctAt60': at(60), 'pctAt180': at(180), 'pctAt300': at(300), 'pctAt480': at(b['dur']),
                        'minPct': min(r['pct'] for r in curve if r['build'] == b['id']),
                        'spentPerSec': r1(b['spentPerSec']), 'gainedPerSec': r1(b['gainedPerSec']),
                        'oomSecs': r1(b['oomSecs']), 'tto': r1(b['tto']), 'manaEnd': round(b['manaEnd'])})
    return curve, summary


def write_data():
    curve, summary = datasets()
    os.makedirs(os.path.join(D, 'data'), exist_ok=True)
    for name, rows in (('mana', curve), ('mana_summary', summary)):
        p = os.path.join(D, 'data', name + '.json')
        json.dump(rows, open(p, 'w'), separators=(',', ':'))
        print(f'{name}: {len(rows)} rows, {os.path.getsize(p) / 1024:.0f} KB')


def main(args):
    if '--data' in args:
        write_data()
        return
    force = '--force' in args
    ids = [a for a in args if not a.startswith('--')] or mana_builds()
    os.makedirs(os.path.join(D, 'mana'), exist_ok=True)
    for bid in ids:
        path = os.path.join(D, 'mana', bid + '.json')
        if os.path.exists(path) and not force:
            continue
        b = sim_build(bid)
        json.dump(b, open(path, 'w'), separators=(',', ':'))
        if b['fights']:
            low = sorted(math.inf if v is None else v for v in b['lowAt'])
            print(f"{bid}: max {b['maxMana']:.0f}, below {LOW}% at {q(low, 0.5)} (median), oom {b['oomSecs']:.1f} s, tto {b['tto']:.0f}")
        else:
            print(f'{bid}: {b["note"]}')


if __name__ == '__main__':
    main(sys.argv[1:])
