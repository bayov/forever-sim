#!/usr/bin/env python3
"""race_gear.py [ids...] [--write] [--iters N]: the Dwarf's gear (standards 29, gear bullet).

Dwarf Artisan's Precision adds 1% hit with spells and attacks (sim/core/racials_forever.go). Every set carries hit up to
its standards 4 cap, so on a Dwarf that 1% can go over it. For every build a Dwarf can be (run_build.CLASS_RACES; duty
and encounter variants take their base build's file), this works out:

  hit      the Dwarf's hit from the gear, talents, set bonus, buffs and consumes (dashstats' final sheet at Dwarf), plus a
           priest's school talent the sheet doesn't hold (Shadow Focus 1% a rank, Holy Precision 6% a rank)
  cap      casters (spell power above attack power): the 16% spell cap; melee and hunters: the 8% yellow cap
  excess   min(1%, hit - cap, the gear's own hit)
  swap     the gear's hit lowered by the excess, and its 10 budget points a percent (standards 1b) given to the set's
           other non-Stamina, non-Spirit stats in the set's own proportions (orchestration/gear_targets.md, the
           hit-adjusted rule): shares at the 1b prices over the armor and weapons (trinkets and enchants left out, an
           item's attack power and ranged attack power counted once), each stat's points turned back into its amount.

Then it sims Dwarf with the preset's gear and with the swap (180 s, seed 1, 20,000 iterations, both with the same
seed, so the fights pair up: the error is the standard error of the per-fight difference). With --write it writes
req/races/<id>.Dwarf.json (race_override.py) where the swap gains at least 2 standard errors, and removes a file of an
earlier run that no longer does. Dual wielders' extra hit still lowers their white misses, so a swap can lose.

The gear's hit and crit are one pool under Forever (sim/core/ruleset.go unifyEquipHitAndCrit adds the melee and spell
values, bonus stats included), so the swap writes each once: MeleeHit and MeleeCrit for melee and hunters, SpellHit and
SpellCrit for casters. Either works for both kinds of attack.

Writes $DASH_DIR/race_gear.json (every build's numbers) and prints a table.
"""
import json, math, os, subprocess, sys, tempfile
from concurrent.futures import ThreadPoolExecutor

sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
from dashpaths import D, BIN, TREE  # noqa: E402
from run_build import CLASS_RACES, run  # noqa: E402
import race_override  # noqa: E402

RACE = 'Dwarf'
# standards 1b prices, per point of the item's stat (hit and crit per rating point: 10 a percent of hit, 14 of crit)
PRICE = {'Strength': 1, 'Agility': 1, 'Intellect': 1, 'SpellPower': 0.86, 'AttackPower': 0.5, 'RangedAttackPower': 0.4,
         'MP5': 2.4, 'CritRating': 1.0, 'ArcanePower': 0.7, 'FirePower': 0.7, 'FrostPower': 0.7, 'NaturePower': 0.7,
         'ShadowPower': 0.7, 'HolyPower': 0.7}
HIT_RATING, CRIT_RATING = 10.0, 14.0
SPELL_CAP, MELEE_CAP = 16.0, 8.0
TRINKET = 12  # proto.ItemType


def stat_names():
    i = race_override.stat_index  # loads the enum
    i('Strength')
    return {v: k[4:] for k, v in race_override._STATS.items()}


def dashstats(req):
    with tempfile.NamedTemporaryFile('w', suffix='.json', delete=False) as t:
        json.dump(req, t)
    try:
        r = subprocess.run([os.path.join(BIN, 'dashstats'), t.name], capture_output=True, text=True, cwd=TREE, check=True)
    finally:
        os.unlink(t.name)
    return json.loads(r.stdout)


def talent_ranks(cls, talents):
    trees = json.load(open(os.path.join(TREE, 'ui/core/talents/trees', cls.lower() + '.json')))
    out = {}
    for tree, s in zip(trees, talents.split('-')):
        for t, ch in zip(tree['talents'], s):
            out[t['fieldName']] = int(ch)
    return out


def school_hit(meta, pl):
    """Spell hit a priest's talent gives its main school, which the stat sheet doesn't hold (sim/priest/talents.go)."""
    if meta['class'] != 'Priest':
        return 0.0, ''
    t = talent_ranks('Priest', pl['talentsString'])
    if meta.get('spec') == 'Smite' and t.get('holyPrecision'):
        return 6.0 * t['holyPrecision'], f"Holy Precision {t['holyPrecision']}/3"
    if t.get('shadowFocus'):
        return 1.0 * t['shadowFocus'], f"Shadow Focus {t['shadowFocus']}/5"
    if t.get('holyPrecision'):
        return 6.0 * t['holyPrecision'], f"Holy Precision {t['holyPrecision']}/3"
    return 0.0, ''


def set_stats(pl, names):
    """The armor and weapons' stats (no trinkets, no enchants), by budget category: amounts and hit rating."""
    db = {i['id']: i for i in json.load(open(os.path.join(TREE, 'assets/database/db.json')))['items']}
    amt, hit = {}, 0.0
    for it in pl['equipment']['items']:
        if not it.get('id'):
            continue
        x = db[it['id']]
        if x.get('type') == TRINKET:
            continue
        hit += x.get('hitRating', 0)
        s = {names[k]: v for k, v in enumerate(x.get('stats', [])) if v}
        ap, rap = s.pop('AttackPower', 0), s.pop('RangedAttackPower', 0)
        both = min(ap, rap)  # one item's attack power and ranged attack power count once, as attack power
        for k, v in (('AttackPowerBoth', both), ('AttackPower', ap - both), ('RangedAttackPower', rap - both),
                     ('CritRating', x.get('critRating', 0))):
            if v:
                amt[k] = amt.get(k, 0) + v
        for k, v in s.items():
            if k in PRICE:
                amt[k] = amt.get(k, 0) + v
    return amt, hit


def price(k):
    return PRICE['AttackPower'] if k == 'AttackPowerBoth' else PRICE[k]


def at_race(req):
    """The request at Dwarf with the preset's own rotation and gear."""
    req = json.loads(json.dumps(req))
    req['raid']['parties'][0]['players'][0]['race'] = 'Race' + RACE
    return req


def plan(bid):
    meta = json.load(open(os.path.join(D, 'req', bid + '.meta.json')))
    req = json.load(open(os.path.join(D, 'req', bid + '.json')))
    pl = req['raid']['parties'][0]['players'][0]
    names = stat_names()
    st = dashstats(at_race(req))
    final = st['final']
    caster = final.get('SpellPower', 0) > max(final.get('AttackPower', 0), final.get('RangedAttackPower', 0))
    extra, why = school_hit(meta, pl)
    hit = (final.get('SpellHit', 0) + extra) if caster else final.get('MeleeHit', 0)
    cap = SPELL_CAP if caster else MELEE_CAP
    amt, gear_hit_rating = set_stats(pl, names)
    gear_hit = gear_hit_rating / HIT_RATING
    excess = round(max(0.0, min(1.0, hit - cap, gear_hit)), 4)
    p = {'build': bid, 'class': meta['class'], 'spec': meta.get('spec'), 'role': 'caster' if caster else 'melee/ranged',
         'hit': round(hit, 2), 'hitFrom': why, 'cap': cap, 'gearHit': gear_hit, 'excess': excess}
    if excess <= 0:
        return p, req, meta
    pts = excess * HIT_RATING
    total = sum(price(k) * v for k, v in amt.items())
    add = {k: pts * price(k) * v / total / price(k) for k, v in amt.items()}  # amount = points x share / price
    bonus = {('SpellHit' if caster else 'MeleeHit'): -excess}
    for k, v in add.items():
        if k == 'CritRating':
            bonus['SpellCrit' if caster else 'MeleeCrit'] = v / CRIT_RATING
        elif k == 'AttackPowerBoth':
            bonus['AttackPower'] = bonus.get('AttackPower', 0) + v
            bonus['RangedAttackPower'] = bonus.get('RangedAttackPower', 0) + v
        else:
            bonus[k] = bonus.get(k, 0) + v
    p['bonusStats'] = {k: round(v, 3) for k, v in bonus.items()}
    p['shares'] = {k: round(100 * price(k) * v / total, 1) for k, v in sorted(amt.items(), key=lambda kv: -price(kv[0]) * kv[1])}
    return p, req, meta


LABEL = {'SpellPower': 'spell power', 'AttackPower': 'attack power', 'RangedAttackPower': 'ranged attack power', 'Strength': 'Strength',
         'Agility': 'Agility', 'Intellect': 'Intellect', 'MP5': 'MP5', 'MeleeCrit': '% crit', 'SpellCrit': '% crit'}


def describe(p):
    gains = [(k, v) for k, v in p['bonusStats'].items() if v > 0 and not (k == 'RangedAttackPower' and p['bonusStats'].get('AttackPower') == v)]
    gains.sort(key=lambda kv: -kv[1] * (14 if kv[0].endswith('Crit') else 1))
    parts = [(f'{v:.2f}{LABEL[k]}' if k.endswith('Crit') else f'{v:.1f} {LABEL[k]}') for k, v in gains
             if v >= (0.05 if k.endswith('Crit') else 0.5)][:4]  # the main ones: bonusStats has them all
    pct = f"{p['excess']:g}%"
    return f"{pct} hit over the cap re-spent as {', '.join(parts[:-1]) + ' and ' + parts[-1] if len(parts) > 1 else parts[0]}"


def paired(req, bonus, iters):
    def one(r):
        res = run(r, iters, 1, save_all=True)
        return res['raidMetrics']['parties'][0]['players'][0]['dps']
    swap = json.loads(json.dumps(req))
    pl = swap['raid']['parties'][0]['players'][0]
    bs = pl.setdefault('bonusStats', {})
    for name, v in bonus.items():
        field, i = race_override.stat_index(name)
        arr = bs.setdefault(field, [])
        arr.extend([0] * (i + 1 - len(arr)))
        arr[i] += v
    with ThreadPoolExecutor(2) as ex:
        a, b = ex.map(one, [req, swap])
    d = [y - x for x, y in zip(a['allValues'], b['allValues'])]
    m = sum(d) / len(d)
    sd = math.sqrt(sum((x - m) ** 2 for x in d) / (len(d) - 1))
    return a['avg'], b['avg'], m, sd / math.sqrt(len(d)), math.hypot(a['stdev'], b['stdev']) / math.sqrt(len(d))


def main(args):
    write = '--write' in args
    iters = int(args[args.index('--iters') + 1]) if '--iters' in args else 20000
    ids = [a for a in args if not a.startswith('--') and not a.isdigit()]
    if not ids:
        ids = sorted(f[:-10] for f in os.listdir(os.path.join(D, 'req')) if f.endswith('.meta.json'))
    os.environ.setdefault('DASH_SPLITS', '4')
    out_p = os.path.join(D, 'race_gear.json')
    results = json.load(open(out_p)) if os.path.exists(out_p) else {}
    for bid in ids:
        meta = json.load(open(os.path.join(D, 'req', bid + '.meta.json')))
        if meta.get('baseBuild') or RACE not in CLASS_RACES[meta['class']]:
            continue  # variants use their base build's file; a class Dwarf can't be has nothing to do
        p, req, meta = plan(bid)
        path = race_override.own_path(bid, RACE)
        if meta['race'] == RACE:
            p['skip'] = 'the preset is a Dwarf: its own set is the preset'
        elif p['excess'] <= 0:
            p['skip'] = 'no hit over the cap' if p['gearHit'] > 0 else 'the set carries no hit'
        else:
            dw = at_race(req)
            dw['encounter']['duration'] = 180
            base, swap, gain, se, se_unpaired = paired(dw, p['bonusStats'], iters)
            p.update({'dpsPreset': round(base, 2), 'dpsSwap': round(swap, 2), 'gain': round(gain, 2), 'se': round(se, 2),
                      'seUnpaired': round(se_unpaired, 2),
                      'keep': gain >= 2 * se, 'note': describe(p)})
        results[bid] = p
        json.dump(results, open(out_p, 'w'), indent=1)
        line = f"{bid:40s} hit {p['hit']:5.2f} cap {p['cap']:4.1f} excess {p['excess']:.2f}"
        if 'gain' in p:
            line += f"  {p['dpsPreset']:8.1f} -> {p['dpsSwap']:8.1f}  gain {p['gain']:+6.2f} ± {p['se']:.2f}  {'KEEP' if p['keep'] else 'drop'}  {p['note']}"
        else:
            line += '  ' + p['skip']
        print(line, flush=True)
        if not write:
            continue
        if p.get('keep'):
            if os.path.exists(path):
                old = json.load(open(path))
                if 'rotation' in old:
                    print(f'  {path} has a rotation: merge the gear by hand and confirm the pair', file=sys.stderr)
                    continue
            os.makedirs(os.path.dirname(path), exist_ok=True)
            shares = ', '.join(f'{k} {v}%' for k, v in p['shares'].items())
            how = (f"race_gear.py: Dwarf hit {p['hit']:g}%{' (with ' + p['hitFrom'] + ')' if p['hitFrom'] else ''} against the "
                   f"{p['cap']:g}% {'spell' if p['role'] == 'caster' else 'yellow'} cap, the set's own hit {p['gearHit']:g}%: "
                   f"{p['excess']:g}% ({p['excess'] * HIT_RATING:g} budget points) re-spent over the set's shares at the 1b prices ({shares}). "
                   f"Hit and crit are one pool for the gear, so each is written once. Dwarf {p['dpsPreset']:.1f} with the preset's "
                   f"gear, {p['dpsSwap']:.1f} with the swap; gain.se is the paired error (the same seeded fights), "
                   f"{p['seUnpaired']:.2f} unpaired.")
            json.dump({'bonusStats': p['bonusStats'], 'note': p['note'],
                       'gain': {'dps': p['gain'], 'se': p['se'], 'iterations': iters, 'seed': 1, 'duration': 180}, 'how': how},
                      open(path, 'w'), indent=1)
            open(path, 'a').write('\n')
            print('  wrote', path)
        elif os.path.exists(path) and 'rotation' not in json.load(open(path)) and 'gain' in p:
            os.remove(path)
            print('  removed', path, '(no longer gains 2 SE)')


if __name__ == '__main__':
    main(sys.argv[1:])
