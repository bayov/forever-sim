#!/usr/bin/env python3
"""weights.py <id>...: stat weights for each build ($DASH_DIR/req/<id>.json) into $DASH_DIR/weights/<id>.json.

Each stat is added as a bonus stat (it goes through the same path as gear, so Forever's hit and crit
unification and every stat multiplier apply), simmed at 180 s with the same seed and iteration count as
the baseline, and the DPS change is divided by the amount added. Hit is also measured downward, since a
build at the hit cap gains nothing from more but loses from less.
"""
import json, math, os, sys
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
from run_build import run, D

ITERS = 40000
STAT_INDEX = {'Strength': 0, 'Agility': 1, 'Intellect': 3, 'Spirit': 4, 'SpellPower': 5, 'MP5': 12, 'AttackPower': 17,
              'MeleeHit': 18, 'MeleeCrit': 19, 'RangedAttackPower': 27}
LABEL = {'Strength': 'Strength', 'Agility': 'Agility', 'Intellect': 'Intellect', 'Spirit': 'Spirit', 'SpellPower': 'Spell power',
         'MP5': 'Mana per 5 s', 'AttackPower': 'Attack power', 'RangedAttackPower': 'Ranged attack power', 'MeleeHit': 'Hit (1%)',
         'MeleeHitDown': 'Hit (losing 1%)', 'MeleeCrit': 'Crit (1%)'}
STEP = {'Strength': 30, 'Agility': 30, 'Intellect': 30, 'Spirit': 30, 'SpellPower': 30, 'MP5': 12, 'AttackPower': 60,
        'RangedAttackPower': 60, 'MeleeHit': 1, 'MeleeCrit': 1}
# Rating points per 1% (sim/core/ruleset.go: HitRatingPerPercent, CritRatingPerPercent)
RATING = {'MeleeHit': 10, 'MeleeHitDown': 10, 'MeleeCrit': 14}


def stat_list(meta):
    c, spec, bid = meta['class'], meta['spec'], meta['id']
    caster = ['SpellPower', 'Intellect', 'Spirit', 'MP5', 'MeleeHit', 'MeleeCrit']
    melee = ['AttackPower', 'Strength', 'Agility', 'MeleeHit', 'MeleeCrit']
    if c in ('Mage', 'Warlock', 'Priest') or (c == 'Druid' and spec == 'Balance') or (c == 'Shaman' and spec == 'Elemental'):
        return 'SpellPower', caster
    if c == 'Hunter' and 'survival' not in bid:
        return 'RangedAttackPower', ['RangedAttackPower', 'Agility', 'Intellect', 'MeleeHit', 'MeleeCrit']
    if c in ('Paladin', 'Shaman'):
        return 'AttackPower', melee + ['SpellPower', 'Intellect']
    return 'AttackPower', melee


def with_bonus(req, stat, amount):
    r = json.loads(json.dumps(req))
    pl = r['raid']['parties'][0]['players'][0]
    arr = [0.0] * 44
    arr[STAT_INDEX[stat]] = amount
    pl['bonusStats'] = {'stats': arr}
    return r


def dps(r):
    p = r['raidMetrics']['parties'][0]['players'][0]
    return p['dps']['avg'], p['dps']['stdev']


def main(bid):
    req = json.load(open(os.path.join(D, 'req', bid + '.json')))
    meta = json.load(open(os.path.join(D, 'req', bid + '.meta.json')))
    primary, stats = stat_list(meta)
    base, sd = dps(run(req, ITERS, 1))
    se_diff = math.sqrt(2) * sd / math.sqrt(ITERS)
    rows = []
    for st in stats:
        d, _ = dps(run(with_bonus(req, st, STEP[st]), ITERS, 1))
        rows.append({'stat': st, 'label': LABEL[st], 'step': STEP[st], 'dpsPer': (d - base) / STEP[st], 'sePer': se_diff / STEP[st]})
        if st == 'MeleeHit':
            d2, _ = dps(run(with_bonus(req, st, -STEP[st]), ITERS, 1))
            rows.append({'stat': 'MeleeHitDown', 'label': LABEL['MeleeHitDown'], 'step': -STEP[st], 'dpsPer': (base - d2) / STEP[st], 'sePer': se_diff / STEP[st]})
    pv = next(r['dpsPer'] for r in rows if r['stat'] == primary)
    for r in rows:
        r['vsPrimary'] = r['dpsPer'] / pv if pv else None
        r['perRating'] = r['dpsPer'] / RATING[r['stat']] if r['stat'] in RATING else None
        r['perRatingVsPrimary'] = (r['perRating'] / pv) if r['perRating'] is not None and pv else None
    out = {'id': bid, 'primary': primary, 'primaryLabel': LABEL[primary], 'base': base, 'iters': ITERS, 'rows': rows}
    os.makedirs(os.path.join(D, 'weights'), exist_ok=True)
    json.dump(out, open(os.path.join(D, 'weights', bid + '.json'), 'w'))
    print(bid, ' '.join(f"{r['stat']}={r['dpsPer']:.3f}" for r in rows), flush=True)


if __name__ == '__main__':
    for b in sys.argv[1:]:
        if not os.path.exists(os.path.join(D, 'weights', b + '.json')):
            main(b)
