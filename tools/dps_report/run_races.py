#!/usr/bin/env python3
"""run_races.py: race scenario sims for the dashboard's Races tab.

For every preset and every race its class can be (180 s, 10k iterations, seed 1, the preset's
settings otherwise), on top of the race sweep run_build.py already did:
  off            the race's base stats only: every racial left out (DASH_RACIALS_OFF=all)
  off:<part>     one racial part left out, for races with more than one part that can count
  beast / elemental   the same fight against that creature type (presets fight the generic boss)
  humanoid       priests only: Chastise (Dwarf priests) works on Humanoids only
  weapon         the preset's weapon(s) as the race's racial weapon type (same stats and weapon DPS)
  weaponNoRacial the same weapon swap with that weapon racial left out
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

DB = {i['id']: i for i in json.load(open(os.path.join(TREE, 'assets/database/db.json')))['items']}

AXE, MACE, SWORD, DAGGER, STAFF = 1, 4, 9, 2, 8
WT_NAME = {AXE: 'axe', MACE: 'mace', SWORD: 'sword', DAGGER: 'dagger', STAFF: 'staff', 3: 'fist', 6: 'polearm'}
RACE_WEAPON = {'Human': (SWORD, 'sword'), 'Dwarf': (MACE, 'mace'), 'Orc': (AXE, 'axe')}
# Weapon types each class can use (Classic proficiencies), for the types the racials name.
LEGAL = {'Warrior': {AXE, MACE, SWORD}, 'Paladin': {AXE, MACE, SWORD}, 'Hunter': {AXE, SWORD},
         'Rogue': {SWORD, MACE}, 'Shaman': {AXE, MACE}, 'Mage': {SWORD}, 'Warlock': {SWORD},
         'Priest': {MACE}, 'Druid': {MACE}}
# Real Era Phase 1 weapons whose speed and damage spread a swapped physical weapon takes,
# by (type, hand): 2H, main hand, off hand. Damage is scaled to the preset weapon's DPS.
SPEED_FROM = {(AXE, '2h'): 17104, (MACE, '2h'): 17073, (SWORD, '2h'): 17076,
              (AXE, 'mh'): 17068, (MACE, 'mh'): 11684, (SWORD, 'mh'): 17075,
              (AXE, 'oh'): 12798, (MACE, 'oh'): 18683, (SWORD, 'oh'): 18832}
CASTER = {'Mage', 'Warlock', 'Priest'}
# Racial parts that can change DPS, by race.
PARTS = {
    'Human': ['spirit', 'sword'], 'Dwarf': ['mace', 'chastise'], 'NightElf': ['elune', 'starshards'],
    'Gnome': ['mind', 'eureka'], 'Orc': ['axe', 'bloodfury'], 'Undead': ['touch', 'darksacrifice'],
    'Tauren': ['endurance'], 'Troll': ['berserking'], 'SkyborneHighOrder': ['windblessed'], 'SkyborneWindshaper': ['windblessed'],
}
PRIEST_ONLY = {'chastise', 'starshards', 'darksacrifice'}
WEAPON_PART = {'sword': SWORD, 'mace': MACE, 'axe': AXE}
TARGETS = {'beast': 'MobTypeBeast', 'elemental': 'MobTypeElemental'}
CLASS_TARGETS = {'Priest': {'humanoid': 'MobTypeHumanoid'}}  # a racial that cares about Humanoids (Chastise)
NEW_ID = 9_990_000


def player(req):
    return req['raid']['parties'][0]['players'][0]


def weapons(req):
    eq = player(req)['equipment']['items']
    return [DB.get((eq[i] or {}).get('id')) if len(eq) > i and eq[i] else None for i in (14, 15)]


def parts_for(cls, race, req):
    held = {w['weaponType'] for w in weapons(req) if w}
    out = []
    for p in PARTS.get(race, []):
        if p in PRIEST_ONLY and cls != 'Priest':
            continue
        if p in WEAPON_PART and WEAPON_PART[p] not in held:
            continue
        out.append(p)
    return out


def swapped(req, cls, wtype):
    """The request with its weapon(s) changed to wtype: same stats, same weapon DPS."""
    req = copy.deepcopy(req)
    pl = player(req)
    eq = pl['equipment']['items']
    mh, oh = weapons(req)
    items, nid = [], NEW_ID + wtype * 10
    if cls in CASTER or (mh and mh['weaponType'] == STAFF and cls in ('Shaman', 'Druid')):
        # a caster: the staff's stats on a one-hand weapon of that type, off hand left empty
        it = {k: v for k, v in mh.items() if k in SIMITEM}
        it.update(id=nid, name=f"{mh['name']} as a {WT_NAME[wtype]}", weaponType=wtype, handType=1)
        items.append(it)
        eq[14] = dict(eq[14], id=nid)
        return req, items
    slots = [14, 15]
    if mh and mh['weaponType'] == DAGGER and cls == 'Rogue':
        slots = [15]  # dagger builds keep the main hand dagger (Backstab, Mutilate, Ambush)
    for k, slot in enumerate(slots):
        w = DB.get((eq[slot] or {}).get('id')) if len(eq) > slot and eq[slot] else None
        if not w or not w.get('weaponDamageMin'):
            continue
        hand = '2h' if w['handType'] == 4 else ('mh' if slot == 14 else 'oh')
        ref = DB[SPEED_FROM[(wtype, hand)]]
        dps = (w['weaponDamageMin'] + w['weaponDamageMax']) / 2 / w['weaponSpeed']
        ravg = (ref['weaponDamageMin'] + ref['weaponDamageMax']) / 2
        scale = dps * ref['weaponSpeed'] / ravg
        it = {k2: v for k2, v in w.items() if k2 in SIMITEM}
        it.update(id=nid + k + 1, name=f"{w['name']} as a {WT_NAME[wtype]}", weaponType=wtype,
                  weaponSpeed=ref['weaponSpeed'], weaponDamageMin=round(ref['weaponDamageMin'] * scale, 1),
                  weaponDamageMax=round(ref['weaponDamageMax'] * scale, 1))
        items.append(it)
        eq[slot] = dict(eq[slot], id=nid + k + 1)
    return req, items


SIMITEM = {'id', 'name', 'type', 'armorType', 'weaponType', 'handType', 'rangedWeaponType', 'stats', 'weaponDamageMin', 'weaponDamageMax',
           'weaponSpeed', 'bonusPhysicalDamage', 'setName', 'setId', 'weaponSkills', 'hitRating', 'critRating', 'classAllowlist', 'quality'}


def sim(req, race, off=None, mob=None, items=None):
    req = copy.deepcopy(req)
    pl = player(req)
    pl['race'] = 'Race' + race
    if items:
        pl['database'] = {'items': items}
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


def tasks(bid):
    req = json.load(open(os.path.join(D, 'req', bid + '.json')))
    meta = json.load(open(os.path.join(D, 'req', bid + '.meta.json')))
    cls = meta['class']
    held = {w['weaponType'] for w in weapons(req) if w}
    for race in CLASS_RACES[cls]:
        yield (bid, race, 'off'), lambda race=race: sim(req, race, off='all')
        ps = parts_for(cls, race, req)
        if len(ps) > 1:
            for p in ps:
                yield (bid, race, 'off:' + p), lambda race=race, p=p: sim(req, race, off=p)
        for t, mob in {**TARGETS, **CLASS_TARGETS.get(cls, {})}.items():
            yield (bid, race, t), lambda race=race, mob=mob: sim(req, race, mob=mob)
        if race in RACE_WEAPON:
            wt, part = RACE_WEAPON[race]
            if wt in LEGAL[cls] and wt not in held:
                r2, items = swapped(req, cls, wt)
                yield (bid, race, 'weapon'), lambda race=race, r2=r2, items=items: sim(r2, race, items=items)
                yield (bid, race, 'weaponNoRacial'), lambda race=race, r2=r2, items=items, part=part: sim(r2, race, off=part, items=items)


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
