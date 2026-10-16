#!/usr/bin/env python3
"""weights.py <id>...: stat weights for each build ($DASH_DIR/req/<id>.json) into $DASH_DIR/weights/<id>.json.

Each stat is added as a bonus stat (it goes through the same path as gear, so Forever's hit and crit
unification and every stat multiplier apply), simmed at 180 s with the same seed and iteration count as
the baseline, and the DPS change is divided by the amount added. Hit is also measured downward, since a
build at the hit cap gains nothing from more but loses from less.

Every run uses seed 1, so fight i of the bonus run and fight i of the baseline share their random rolls. The
error of a weight is the standard error of those paired per-fight differences, about 80 times smaller than two
independent runs' (measured: Fury +60 AP, paired sd 0.9 DPS against a fight-to-fight sd of 52.6). So ITERS
(default 5,000) is enough for most stats. A stat whose paired error is still larger than the error the old method
had at 40,000 unpaired iterations (sqrt(2) x the fight sd / sqrt(40,000)) is simmed again at MAX_ITERS, so no weight
is ever less precise than before. Each row records its iterations.
"""
import json, math, os, sys
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
from run_build import run, D

ITERS = int(os.environ.get('ITERS', 5000))
MAX_ITERS = int(os.environ.get('MAX_ITERS', 40000))
OLD_ITERS = 40000  # the unpaired method's iterations, whose error a weight must match or beat
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
    return p['dps']['avg'], p['dps']['allValues']


def paired(base_vals, vals):
    """Mean and standard error of the per-fight DPS differences (same seed, so fight i matches fight i)."""
    d = [b - a for a, b in zip(base_vals, vals)]
    m = sum(d) / len(d)
    return m, math.sqrt(sum((x - m) ** 2 for x in d) / (len(d) - 1) / len(d))


def measure(req, stats, iters):
    base, bv = dps(run(req, iters, 1, save_all=True))
    sd = math.sqrt(sum((x - base) ** 2 for x in bv) / (len(bv) - 1))
    old_se = math.sqrt(2) * sd / math.sqrt(OLD_ITERS)  # a step's DPS error with two independent 40k runs
    rows = []
    for st in stats:
        for sign in ((1, -1) if st == 'MeleeHit' else (1,)):
            _, v = dps(run(with_bonus(req, st, sign * STEP[st]), iters, 1, save_all=True))
            d, se = paired(bv, v)
            name = st if sign > 0 else 'MeleeHitDown'
            rows.append({'stat': name, 'label': LABEL[name], 'step': sign * STEP[st], 'dpsPer': sign * d / STEP[st], 'sePer': se / STEP[st], 'oldSe': old_se})
    return base, rows


def main(bid):
    req = json.load(open(os.path.join(D, 'req', bid + '.json')))
    meta = json.load(open(os.path.join(D, 'req', bid + '.meta.json')))
    primary, stats = stat_list(meta)
    iters = ITERS
    base, rows = measure(req, stats, iters)
    for r in rows:
        r['iters'] = iters
    # Only the stats whose paired error is still worse than the old method's are simmed again, straight at MAX_ITERS
    # (melee hit and crit steps change the proc timing, so their fights drift apart and pairing helps less).
    noisy = sorted({'MeleeHit' if r['stat'] == 'MeleeHitDown' else r['stat'] for r in rows if r['sePer'] * abs(r['step']) > r['oldSe']})
    if noisy and iters < MAX_ITERS:
        iters = MAX_ITERS
        base, redo = measure(req, noisy, iters)
        for r in redo:
            r['iters'] = iters
        rows = [next((x for x in redo if x['stat'] == r['stat']), r) for r in rows]
    pv = next(r['dpsPer'] for r in rows if r['stat'] == primary)
    for r in rows:
        r['vsPrimary'] = r['dpsPer'] / pv if pv else None
        r['perRating'] = r['dpsPer'] / RATING[r['stat']] if r['stat'] in RATING else None
        r['perRatingVsPrimary'] = (r['perRating'] / pv) if r['perRating'] is not None and pv else None
    for r in rows:
        r.pop('oldSe', None)
    out = {'id': bid, 'primary': primary, 'primaryLabel': LABEL[primary], 'base': base, 'iters': iters, 'paired': True, 'rows': rows}
    os.makedirs(os.path.join(D, 'weights'), exist_ok=True)
    json.dump(out, open(os.path.join(D, 'weights', bid + '.json'), 'w'))
    print(bid, ' '.join(f"{r['stat']}={r['dpsPer']:.3f}" for r in rows), flush=True)


if __name__ == '__main__':
    for b in sys.argv[1:]:
        # Skip a build only when its weights are newer than its request (a changed request, or a re-run that refreshed
        # the request, sims it again). WEIGHTS_FORCE=1 redoes them all.
        w, r = os.path.join(D, 'weights', b + '.json'), os.path.join(D, 'req', b + '.json')
        if os.environ.get('WEIGHTS_FORCE') == '1' or not os.path.exists(w) or os.path.getmtime(w) < os.path.getmtime(r):
            main(b)
