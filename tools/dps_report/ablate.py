#!/usr/bin/env python3
"""ablate.py: for every preset, sim it with each consume / buff / debuff / profession left out, same seed.
An entry whose removal leaves the DPS bit-identical has no effect in the sim. Writes abl/ablate.json."""
import json, os, re, subprocess, sys, copy, glob
from concurrent.futures import ThreadPoolExecutor
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
from dashpaths import D, BIN, TREE  # base dir, binaries, merged tree (env DASH_DIR / DASH_BIN / DASH_TREE)
SIM = os.path.join(BIN, 'wowsimcli')
ITERS, SEED = int(os.environ.get('ITERS', 300)), os.environ.get('SEED', '11')
os.makedirs(os.path.join(D, 'abl', 'tmp'), exist_ok=True)


def words(k):
    return re.sub(r'(?<=[a-z0-9])(?=[A-Z])', ' ', k[0].upper() + k[1:])


def run(req, tag):
    req = copy.deepcopy(req)
    req['simOptions'] = {'iterations': ITERS, 'randomSeed': SEED}
    i, o = os.path.join(D, 'abl', 'tmp', tag + '.i.json'), os.path.join(D, 'abl', 'tmp', tag + '.o.json')
    json.dump(req, open(i, 'w'))
    subprocess.run([SIM, 'sim', '--infile', i, '--outfile', o], capture_output=True, cwd=TREE)
    try:
        r = json.load(open(o))
    except Exception as e:
        return {'error': 'no output'}
    finally:
        for p in (i, o):
            try: os.remove(p)
            except OSError: pass
    if r.get('error'):
        return {'error': r['error'].get('message', '')[:300]}
    return {'dps': r['raidMetrics']['parties'][0]['players'][0]['dps']['avg']}


def entries(req):
    pl = req['raid']['parties'][0]['players'][0]
    out = []
    for cat, holder in (('Consumes', ('player', 'consumes')), ('Self buffs', ('player', 'buffs')), ('Raid buffs', ('raid', 'buffs')),
                        ('Party buffs', ('party', 'buffs')), ('Debuffs', ('raid', 'debuffs'))):
        obj = {'player': pl, 'raid': req['raid'], 'party': req['raid']['parties'][0]}[holder[0]].get(holder[1]) or {}
        for k, v in obj.items():
            if v in (False, 0, '', None) or v == 'TristateEffectMissing':
                continue
            if cat == 'Consumes':
                txt = words(k) + ': ' + (words(str(v)) if isinstance(v, str) else str(v))
            else:
                txt = words(k) + (' (improved)' if v == 'TristateEffectImproved' else '' if v in (True, 'TristateEffectRegular') else f' ({v})')
            out.append((cat, txt, holder, k))
    for pk in ('profession1', 'profession2'):
        if pl.get(pk):
            out.append(('Professions', words(pl[pk]), ('player', pk), None))
    return out


def drop(req, holder, k):
    r = copy.deepcopy(req)
    pl = r['raid']['parties'][0]['players'][0]
    if k is None:
        pl.pop(holder[1], None)
        return r
    obj = {'player': pl, 'raid': r['raid'], 'party': r['raid']['parties'][0]}[holder[0]][holder[1]]
    obj.pop(k)
    return r


def main():
    ids = sys.argv[1:] or [os.path.basename(p)[:-5] for p in sorted(glob.glob(os.path.join(D, 'req', '*.json'))) if not p.endswith('.meta.json')]
    jobs = []
    reqs = {}
    for bid in ids:
        req = json.load(open(os.path.join(D, 'req', bid + '.json')))
        reqs[bid] = req
        jobs.append((bid, None, None, req))
        for cat, txt, holder, k in entries(req):
            jobs.append((bid, cat, txt, drop(req, holder, k)))
    # pairs that only make sense together: the 3% crit pair and the armor pairs
    PAIRS = [('Raid buffs', 'leaderOfThePack', 'moonkinAura'), ('Debuffs', 'sunderArmor', 'exposeArmor'), ('Debuffs', 'faerieFire', 'curseOfRecklessness'),
             ('Debuffs', 'demoralizingShout', 'demoralizingRoar')]
    for bid, req in reqs.items():
        for cat, a, b in PAIRS:
            obj = req['raid']['buffs' if cat == 'Raid buffs' else 'debuffs']
            if obj.get(a) not in (None, False, 'TristateEffectMissing') and obj.get(b) not in (None, False, 'TristateEffectMissing'):
                r = copy.deepcopy(req)
                o2 = r['raid']['buffs' if cat == 'Raid buffs' else 'debuffs']
                o2.pop(a); o2.pop(b)
                jobs.append((bid, 'pair', a + '+' + b, r))
    print(len(jobs), 'sims', flush=True)
    res = {}

    def work(j):
        bid, cat, txt, req = j
        tag = re.sub(r'\W', '_', f'{bid}_{cat}_{txt}')[:120]
        return j, run(req, tag)

    with ThreadPoolExecutor(int(os.environ.get('PAR', 4))) as ex:
        for (bid, cat, txt, _), r in ex.map(work, jobs):
            res.setdefault(bid, {'base': None, 'drop': [], 'pairs': []})
            if cat is None:
                res[bid]['base'] = r
            elif cat == 'pair':
                res[bid]['pairs'].append({'pair': txt, **r})
            else:
                res[bid]['drop'].append({'category': cat, 'item': txt, **r})
    for bid, x in res.items():
        b = x['base'].get('dps')
        for d in x['drop'] + x['pairs']:
            if 'dps' in d and b is not None:
                d['delta'] = d['dps'] - b
                d['same'] = d['dps'] == b
    out = os.path.join(D, 'abl', 'ablate.json')
    json.dump({'iters': ITERS, 'seed': SEED, 'builds': res}, open(out, 'w'), indent=1)
    print('wrote', out)


main()
