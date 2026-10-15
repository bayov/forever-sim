#!/usr/bin/env python3
"""run_assumptions.py [ids]: the sims behind the dashboard's assumption scenarios (assumptions_def.py).

Binary: $DASH_BIN/wowsimcli_assume (else wowsimcli), built from the merged tree with patches/0004 (the DASH_ASSUME knobs,
sim/core/assumptions.go). With no DASH_ASSUME it is bit-identical to the plain build (check with --identical <plain wowsimcli>).

Per build an assumption touches, seed 1, the build's request (req/<id>.json) otherwise:
  base      no knob set: 180 s at 20k iterations, 120 and 300 s at 5k
  single    one assumption at one non-default setting, the same three runs
  grid      builds with 3 or fewer assumptions: every combination of two or more non-default settings, 180 s at 20k
  check     builds with more: 12 random combinations, 180 s at 20k, to measure the multiplicative combination's error
A setting with the ice_lance edit runs the preset rotation and one with Ice Lance on every Fingers of Frost charge, and keeps
the better (no setting uses it now: the Frost and Frostfire presets cast Ice Lance themselves). A setting with the
no_ice_lance edit (Ice Lance at the client's coefficient, 0) runs the preset rotation and the build's rotation that never
casts Ice Lance (asm_apls/<id>.no_ice_lance.apl.json, the Level 60 APL before the presets took Ice Lance up), and keeps
the better. A setting with the raid_hits=N edit sets the shaman's or warrior's raid damage hits a minute (raidDamageHitsPerMinute) to N.
Results append to asm/results.jsonl keyed (build, combo, secs, rot), so a rerun only does what is missing.
"""
import copy, itertools, json, math, os, random, subprocess, sys, tempfile, threading
from concurrent.futures import ThreadPoolExecutor

sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
from dashpaths import D, BIN, TREE  # noqa: E402
from assumptions_def import A  # noqa: E402

CLI = os.path.join(BIN, 'wowsimcli_assume')
if not os.path.exists(CLI):  # any wowsimcli built with the DASH_ASSUME patch will do
    CLI = os.path.join(BIN, 'wowsimcli')
OUT = os.path.join(D, 'asm')
RES = os.path.join(OUT, 'results.jsonl')

LIVE = [a for a in A if a.get('status') not in ('pending', 'inert')]
BY_KEY = {a['key']: a for a in A}
STAFF_IMBUE = {'mage__frostfire': 'ImbueGreaterFlame', 'mage__fire': 'ImbueGreaterFlame', 'mage__frost': 'ImbueGreaterFrost', 'mage__arcane': 'ImbuePrecision'}
ICE_LANCE = {'action': {'castSpell': {'spellId': {'spellId': 1240047}}, 'condition': {'auraIsActive': {'auraId': {'spellId': 400670}}}},
             'notes': 'Ice Lance on a Fingers of Frost charge (assumption scenario).'}


def setting(key, sid):
    return next(s for s in BY_KEY[key]['settings'] if s['id'] == sid)


def combo_str(combo):
    return '|'.join(f'{k}={v}' for k, v in sorted(combo))


def request(bid, combo, rot):
    req = json.load(open(os.path.join(D, 'req', bid + '.json')))
    pl = req['raid']['parties'][0]['players'][0]
    env = {}
    for k, sid in combo:
        st = setting(k, sid)
        env.update(st['env'])
        if st['edit'] == 'staff_imbue':
            pl.setdefault('mage', {}).setdefault('options', {})['staffImbue'] = STAFF_IMBUE[bid]
        if st['edit'] == 'no_staff_imbue':  # the presets carry the scroll; this setting takes it off
            pl.get('mage', {}).get('options', {}).pop('staffImbue', None)
        if (st['edit'] or '').startswith('raid_hits='):  # the shaman's raid damage hits a minute
            spec = next(k for k in ('elementalShaman', 'enhancementShaman', 'warrior') if k in pl)
            pl[spec].setdefault('options', {})['raidDamageHitsPerMinute'] = float(st['edit'].split('=')[1])
        if st['edit'] == 'no_ice_lance' and rot == 'noil':
            pl['rotation'] = json.load(open(os.path.join(os.path.dirname(os.path.abspath(__file__)), 'asm_apls', bid + '.no_ice_lance.apl.json')))
        if st['edit'] == 'ice_lance' and rot == 'il':
            pr = pl['rotation']['priorityList']
            pr.insert(len(pr) - 1, copy.deepcopy(ICE_LANCE))  # just ahead of the filler
    return req, env


def run(req, env, iters, secs):
    req = copy.deepcopy(req)
    req['encounter']['duration'] = secs
    req['simOptions'] = {'iterations': iters, 'randomSeed': '1'}
    e = dict(os.environ)
    e.pop('DASH_NAMES', None)
    e.pop('DASH_RACIALS_OFF', None)
    if env:
        e['DASH_ASSUME'] = ','.join(f'{k}={v}' for k, v in env.items())
    else:
        e.pop('DASH_ASSUME', None)
    with tempfile.TemporaryDirectory() as d:
        i, o = os.path.join(d, 'i.json'), os.path.join(d, 'o.json')
        json.dump(req, open(i, 'w'))
        subprocess.run([CLI, 'sim', '--infile', i, '--outfile', o], check=True, capture_output=True, cwd=TREE, env=e)
        r = json.load(open(o))
    if r.get('error'):
        raise RuntimeError(r['error']['message'][:300])
    p = r['raidMetrics']['parties'][0]['players'][0]['dps']
    return round(p['avg'], 3), round(p['stdev'] / math.sqrt(iters), 3)


def keys_for(bid):
    return [a['key'] for a in LIVE if bid in a['builds']]


def nondefault(key):
    a = BY_KEY[key]
    return [s['id'] for s in a['settings'] if s['id'] != a['default']]


def rots(combo):
    edits = {setting(k, s)['edit'] for k, s in combo}
    return ['preset'] + (['il'] if 'ice_lance' in edits else []) + (['noil'] if 'no_ice_lance' in edits else [])


def tasks(bid):
    ks = keys_for(bid)
    if not ks:
        return
    yield (bid, (), 'base')
    for k in ks:
        for s in nondefault(k):
            yield (bid, ((k, s),), 'single')
    if len(ks) <= 3:
        opts = [[None] + [(k, s) for s in nondefault(k)] for k in ks]
        for pick in itertools.product(*opts):
            c = tuple(x for x in pick if x)
            if len(c) >= 2:
                yield (bid, c, 'grid')
    else:
        rnd = random.Random(bid)
        seen = set()
        while len(seen) < 12:
            n = rnd.randint(2, min(4, len(ks)))
            c = tuple(sorted((k, rnd.choice(nondefault(k))) for k in rnd.sample(ks, n)))
            seen.add(c)
        for c in sorted(seen):
            yield (bid, c, 'check')


def main(ids):
    os.makedirs(OUT, exist_ok=True)
    done = set()
    if os.path.exists(RES):
        for line in open(RES):
            x = json.loads(line)
            done.add((x['build'], x['combo'], x['secs'], x['rot']))
    jobs = []
    for bid in ids:
        for b, combo, kind in tasks(bid):
            lens = [(180, 20000)] + ([(120, 5000), (300, 5000)] if kind in ('base', 'single') else [])
            for secs, iters in lens:
                for rot in rots(combo):
                    if (b, combo_str(combo), secs, rot) not in done:
                        jobs.append((b, combo, kind, secs, iters, rot))
    print(len(jobs), 'sims to run', flush=True)
    lock = threading.Lock()
    n = [0]

    def go(j):
        b, combo, kind, secs, iters, rot = j
        req, env = request(b, combo, rot)
        try:
            dps, se = run(req, env, iters, secs)
        except Exception as e:  # noqa: BLE001
            print('FAIL', b, combo_str(combo), secs, rot, str(e)[:200], flush=True)
            return
        with lock:
            with open(RES, 'a') as fh:
                fh.write(json.dumps({'build': b, 'combo': combo_str(combo), 'kind': kind, 'secs': secs, 'iters': iters, 'rot': rot, 'dps': dps, 'se': se}) + '\n')
            n[0] += 1
            if n[0] % 50 == 0:
                print(n[0], 'done', flush=True)

    with ThreadPoolExecutor(int(os.environ.get('PAR', 4))) as ex:
        list(ex.map(go, jobs))
    print('finished', n[0], flush=True)


def identical(ids, iters=1000):
    """Every build with no DASH_ASSUME on both binaries: the outputs' DPS must match exactly."""
    bad = 0
    for bid in ids:
        req = json.load(open(os.path.join(D, 'req', bid + '.json')))
        req['simOptions'] = {'iterations': iters, 'randomSeed': '1'}
        outs = []
        for cli in (os.environ.get('PLAIN_CLI') or os.path.join(BIN, 'wowsimcli'), CLI):
            with tempfile.TemporaryDirectory() as d:
                i, o = os.path.join(d, 'i.json'), os.path.join(d, 'o.json')
                json.dump(req, open(i, 'w'))
                e = {k: v for k, v in os.environ.items() if not k.startswith('DASH_')}
                subprocess.run([cli, 'sim', '--infile', i, '--outfile', o], check=True, capture_output=True, cwd=TREE, env=e)
                r = json.load(open(o))
                pl = r['raidMetrics']['parties'][0]['players'][0]
                acts = sorted(json.dumps([a['id'], [t.get('damage', 0) for t in a.get('targets', [])]], sort_keys=True) for a in pl.get('actions', []))
                outs.append(json.dumps([pl['dps'], r['raidMetrics']['dps'], acts], sort_keys=True))  # the action list's order is not stable
        same = outs[0] == outs[1]
        bad += not same
        print(bid, 'identical' if same else 'DIFFERENT', flush=True)
    print('builds differing:', bad)


if __name__ == '__main__':
    args = sys.argv[1:]
    allids = sorted(os.path.basename(p)[:-5] for p in os.listdir(os.path.join(D, 'req')) if p.endswith('.json') and not p.endswith('.meta.json'))
    if args and args[0] == '--identical':
        identical(args[1:] or allids)
    else:
        main(args or allids)
