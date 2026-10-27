#!/usr/bin/env python3
"""run_races.py: race scenario sims for the dashboard's Races tab.

For every preset and every race its class can be (180 s, 10k iterations, seed 1, the preset's
settings otherwise), on top of the race sweep run_build.py already did:
  off            the race's base stats only: every racial left out (DASH_RACIALS_OFF=all)
  off:<part>     one racial part left out, for races with more than one part that can count
  beast / elemental   the same fight against that creature type (presets fight the generic boss)
  humanoid       priests only: Chastise (Dwarf priests) works on Humanoids only
No racial depends on the weapon since client 70425 (Human Ruthless works with any weapon), so there is no weapon
swap variant any more. The 2026-10-26 racial pass replaced Dwarf Earth Fury with Artisan's Precision (precision, 1% hit)
and Orc Savage with Command (command, 5% pet damage, Hunters and Warlocks only).
Results append to rv/results.jsonl, one line per sim, so a rerun only does what is missing.
Needs wowsimcli_r: wowsimcli built from the local dashdata tree with the DASH_RACIALS_OFF switch.
"""
import json, os, subprocess, sys, tempfile, copy, math, threading
from concurrent.futures import ThreadPoolExecutor

sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
from dashpaths import D, BIN, TREE  # base dir, binaries, merged tree (env DASH_DIR / DASH_BIN / DASH_TREE)
CLI = os.path.join(BIN, 'wowsimcli_r')
if not os.path.exists(CLI):  # any wowsimcli built with the DASH_RACIALS_OFF patch will do
    CLI = os.path.join(BIN, 'wowsimcli')
RES = os.path.join(D, 'rv', 'results.jsonl')
os.makedirs(os.path.dirname(RES), exist_ok=True)
ITERS = 10000
sys.path.insert(0, D)
from run_build import CLASS_RACES  # noqa: E402


# Racial parts that can change DPS, by race.
PARTS = {
    'Human': ['spirit', 'ruthless'], 'Dwarf': ['precision', 'chastise'], 'NightElf': ['elune', 'shadowmeld', 'starshards'],
    'Gnome': ['mind', 'eureka'], 'Orc': ['command', 'bloodfury'], 'Undead': ['touch', 'darksacrifice'],
    'Tauren': ['endurance'], 'Troll': ['berserking'], 'SkyborneHighOrder': ['windblessed', 'leyline'], 'SkyborneWindshaper': ['windblessed'],
}
PRIEST_ONLY = {'chastise', 'starshards', 'darksacrifice'}
PET_ONLY = {'command'}  # Orc Command: Hunter and Warlock pets only
TARGETS = {'beast': 'MobTypeBeast', 'elemental': 'MobTypeElemental'}
CLASS_TARGETS = {'Priest': {'humanoid': 'MobTypeHumanoid'}}  # a racial that cares about Humanoids (Chastise)
# The races with a racial that depends on the target's creature type. Every other race fights a beast, elemental or
# humanoid boss exactly as it fights the generic one (the last full run: 378 of 476 target sims bit-identical to the
# race sweep), so those rows copy the sweep's result instead of simming. RACES_ALL_TARGETS=1 sims them all again.
TARGET_RACES = {'beast': {'Dwarf', 'Troll'}, 'elemental': {'SkyborneHighOrder', 'SkyborneWindshaper'}, 'humanoid': {'Dwarf'}}
ALL_TARGETS = os.environ.get('RACES_ALL_TARGETS') == '1'


def player(req):
    return req['raid']['parties'][0]['players'][0]


def parts_for(cls, race, req):
    return [p for p in PARTS.get(race, []) if (p not in PRIEST_ONLY or cls == 'Priest')
            and (p not in PET_ONLY or cls in ('Hunter', 'Warlock'))]


def sim(req, race, off=None, mob=None):
    req = copy.deepcopy(req)
    pl = player(req)
    pl['race'] = 'Race' + race
    if mob:
        req['encounter']['targets'][0]['mobType'] = mob
    req['simOptions'] = {'iterations': ITERS, 'randomSeed': '1'}
    env = dict(os.environ)
    env.pop('DASH_NAMES', None)
    if off:
        env['DASH_RACIALS_OFF'] = off
    with tempfile.TemporaryDirectory() as d:
        i, o = os.path.join(d, 'i.json'), os.path.join(d, 'o.json')
        json.dump(req, open(i, 'w'))
        subprocess.run([CLI, 'sim', '--infile', i, '--outfile', o], check=True, capture_output=True, cwd=TREE, env=env)
        r = json.load(open(o))
    if r.get('error'):
        raise RuntimeError(r['error']['message'][:500])
    p = r['raidMetrics']['parties'][0]['players'][0]['dps']
    return round(p['avg'], 3), round(p['stdev'] / math.sqrt(ITERS), 3)


def sweep_result(bid, race):
    """The race sweep's (dps, se) for this race from out/<bid>.json (run_build.py), or None."""
    try:
        o = json.load(open(os.path.join(D, 'out', bid + '.json')))
    except (OSError, ValueError):
        return None
    if os.path.getmtime(os.path.join(D, 'out', bid + '.json')) < os.path.getmtime(os.path.join(D, 'req', bid + '.json')):
        return None  # simmed before the request last changed
    r = next((r for r in o.get('races', []) if r['race'] == race), None)
    return (round(r['dps'], 3), round(r['se'], 3)) if r else None


def tasks(bid):
    req = json.load(open(os.path.join(D, 'req', bid + '.json')))
    meta = json.load(open(os.path.join(D, 'req', bid + '.meta.json')))
    cls = meta['class']
    for race in CLASS_RACES[cls]:
        yield (bid, race, 'off'), lambda race=race: sim(req, race, off='all')
        ps = parts_for(cls, race, req)
        if len(ps) > 1:
            for p in ps:
                yield (bid, race, 'off:' + p), lambda race=race, p=p: sim(req, race, off=p)
        for t, mob in {**TARGETS, **CLASS_TARGETS.get(cls, {})}.items():
            same = None if ALL_TARGETS or race in TARGET_RACES[t] else sweep_result(bid, race)
            if same:
                yield (bid, race, t), lambda same=same: same
            else:
                yield (bid, race, t), lambda race=race, mob=mob: sim(req, race, mob=mob)


def main(ids):
    done = set()
    if os.path.exists(RES):
        for line in open(RES):
            x = json.loads(line)
            done.add((x['build'], x['race'], x['variant']))
    todo = [(k, f) for b in ids for k, f in tasks(b) if k not in done]
    print(len(todo), 'sims to run', flush=True)
    lock = threading.Lock()
    n = [0]

    def go(kf):
        (b, race, v), f = kf
        try:
            dps, se = f()
        except Exception as e:
            print('FAIL', b, race, v, str(e)[:300], flush=True)
            return
        with lock:
            with open(RES, 'a') as fh:
                fh.write(json.dumps({'build': b, 'race': race, 'variant': v, 'dps': dps, 'se': se, 'iters': ITERS}) + '\n')
            n[0] += 1
            if n[0] % 50 == 0:
                print(n[0], 'done', flush=True)

    with ThreadPoolExecutor(4) as ex:
        list(ex.map(go, todo))
    print('finished', n[0], flush=True)


if __name__ == '__main__':
    ids = sys.argv[1:] or sorted(os.path.basename(p)[:-5] for p in os.listdir(os.path.join(D, 'req')) if not p.endswith('.meta.json'))
    ids = [i.replace('.json', '') for i in ids if not i.endswith('.meta')]
    # the encounter presets (a creature-type boss) get only run_build's race sweep
    ids = [i for i in ids if not json.load(open(os.path.join(D, 'req', i + '.meta.json'))).get('encounter')]
    main(ids)
