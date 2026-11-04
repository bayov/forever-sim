#!/usr/bin/env python3
"""race_gear.py [ids...] [--write] [--to-cap] [--iters N]: each race's gear where its hit differs (standards 29, gear bullet).

Every set carries hit up to its standards 4 cap, counted at the preset's own race. A race whose hit differs from the
preset race's lands off the cap: the Dwarf's Artisan's Precision (+1% hit with spells and attacks,
sim/core/racials_forever.go) puts a Dwarf 1% over it, and on a preset that is a Dwarf (its set a step lower for that
1%) every other race comes back 1% under. For every build (run_build.CLASS_RACES gives its races; duty and encounter
variants take their base build's files) and every race it runs at other than the preset's own, this works out:

  hit      the race's hit from the gear, talents, set bonus, buffs and consumes (dashstats' final sheet at that race),
           plus a priest's school talent the sheet doesn't hold (Shadow Focus 1% a rank, Holy Precision 6% a rank)
  cap      casters (spell power above attack power): the 16% spell cap; melee and hunters: the 8% yellow cap

Where the build's gear set has hit-trade steps (standards 4, orchestration/tools/hit_steps.py: real synthetic sets
`<set>_hit-1`, `_hit-2`, ... with 10, 20, ... less gear hit rating, 1% each, re-spent at the budget prices, and `_hit+1`,
... with more; the set is the meta's `gearSet`, found in the merged tree's ui/*/gear_sets/ as the one file whose items
are the request's), the trade is a whole step, in either direction (UPDATED 2026-11-04, phase 5):

  target   the cap, or the preset race's own hit where the preset carries more (dual wielders may carry hit past the
           yellow cap by sim, standards 4): the race's step then moves only by the race's difference. --to-cap aims
           at the cap itself, which also re-steps a preset that sits over it.
  step     the preset's own step (the step file the request's equipment is, 0 for the base set) minus
           round(hit - target) (floor(x + 0.5)), clamped to the steps there are
  swap     that step's gear set, when it is another step than the preset's

Otherwise (no steps, or the set can't be found), and for the Dwarf only, the trade is bonus stats:

  excess   min(1%, hit - cap, the gear's own hit)
  swap     the gear's hit lowered by the excess, and its 10 budget points a percent (standards 1b) given to the set's
           other non-Stamina, non-Spirit stats in the set's own proportions (orchestration/gear_targets.md, the
           hit-adjusted rule): shares at the 1b prices over the armor and weapons (trinkets and enchants left out, an
           item's attack power and ranged attack power counted once), each stat's points turned back into its amount.

A swap that takes hit away (a race over the cap: a lower step, or the bonus stats) is simmed at that race with the
preset's gear and with the swap (180 s, seed 1, 20,000 iterations, both with the same seed, so the fights pair up: the
error is the standard error of the per-fight difference) and kept where it gains at least 2 standard errors: dual
wielders' extra hit still lowers their white misses, so it can lose. A higher step (a race under the cap) is kept by
the rule, without a sim: the race's spells and yellow attacks miss more without it.

With --write it writes req/races/<id>.<Race>.json (race_override.py) for each kept swap: `gearSet` and `equipment` (the
step set's items) for a step, `bonusStats` otherwise. Where the file already holds the race's rotation, the gear is
merged into it (`gearSet`, `equipment`, `gearNote`, `gearHow`, `gearGain`) and the rotation is left as it is; its pair
was simmed on the preset's gear, so check it with rotopt. It removes the gear of an earlier run that no longer applies
(the whole file when it holds nothing else).

The gear's hit and crit are one pool under Forever (sim/core/ruleset.go unifyEquipHitAndCrit adds the melee and spell
values, bonus stats included), so the bonus stats swap writes each once: MeleeHit and MeleeCrit for melee and hunters, SpellHit and
SpellCrit for casters. Either works for both kinds of attack.

Writes $DASH_DIR/race_gear.json (every build and race's numbers, keyed `<id>.<Race>`) and prints a table. Run it with
the binaries built from the tree it reads (dashstats panics on a synthetic item it doesn't know):
  DASH_TREE=/home/user/forever-sim DASH_BIN=<bin> python3 tools/dps_report/race_gear.py [ids] [--write]
"""
import glob, json, math, os, re, subprocess, sys, tempfile
from concurrent.futures import ThreadPoolExecutor

sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
from dashpaths import D, BIN, TREE  # noqa: E402
from run_build import CLASS_RACES, run  # noqa: E402
import race_override  # noqa: E402

DWARF = 'Dwarf'  # the only race the bonus stats fallback serves (its +1% hit racial)
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


STEP = re.compile(r'_hit([+-]\d+)$')


def _ids(eq):
    return [it.get('id', 0) for it in (eq or {}).get('items', [])]


def gear_steps(meta, pl):
    """The hit-trade steps of the build's gear set, or None when it has none or the set isn't found (no file, or more
    than one, in ui/*/gear_sets/ named the meta's `gearSet` with the request's items):
    {'dir', 'base', 'steps': {step: gear set name}, 'own': the step the request wears}."""
    gs = meta.get('gearSet')
    if not gs:
        return None
    mine = _ids(pl.get('equipment'))
    dirs = [os.path.dirname(f) for f in sorted(glob.glob(os.path.join(TREE, 'ui', '*', 'gear_sets', gs + '.gear.json')))
            if _ids(json.load(open(f))) == mine]
    if len(dirs) != 1:
        return None
    base = STEP.sub('', gs)
    steps = {0: base} if os.path.exists(os.path.join(dirs[0], base + '.gear.json')) else {}
    for f in glob.glob(os.path.join(dirs[0], base + '_hit*.gear.json')):
        m = STEP.search(os.path.basename(f)[:-len('.gear.json')])
        if m and os.path.basename(f) == base + m.group(0) + '.gear.json':
            steps[int(m.group(1))] = base + m.group(0)
    m = STEP.search(gs)
    own = int(m.group(1)) if m else 0
    if len(steps) < 2 or steps.get(own) != gs:
        return None
    return {'dir': dirs[0], 'base': base, 'steps': steps, 'own': own}


def step_equipment(st, step):
    return json.load(open(os.path.join(st['dir'], st['steps'][step] + '.gear.json')))


def price(k):
    return PRICE['AttackPower'] if k == 'AttackPowerBoth' else PRICE[k]


def at_race(req, race):
    """The request at this race with the preset's own rotation and gear."""
    req = json.loads(json.dumps(req))
    req['raid']['parties'][0]['players'][0]['race'] = 'Race' + race
    return req


def race_hit(meta, req, race):
    """(hit at this race, caster, the school talent counted): dashstats' final sheet plus a priest's school talent."""
    final = dashstats(at_race(req, race))['final']
    caster = final.get('SpellPower', 0) > max(final.get('AttackPower', 0), final.get('RangedAttackPower', 0))
    extra, why = school_hit(meta, req['raid']['parties'][0]['players'][0])
    return ((final.get('SpellHit', 0) + extra) if caster else final.get('MeleeHit', 0)), caster, why


_preset_hit = {}


def plan(bid, race, to_cap=False):
    meta = json.load(open(os.path.join(D, 'req', bid + '.meta.json')))
    req = json.load(open(os.path.join(D, 'req', bid + '.json')))
    pl = req['raid']['parties'][0]['players'][0]
    names = stat_names()
    hit, caster, why = race_hit(meta, req, race)
    cap = SPELL_CAP if caster else MELEE_CAP
    if bid not in _preset_hit:
        _preset_hit[bid] = race_hit(meta, req, meta['race'])[0]
    # the hit the race's step aims at: the cap, or the preset race's own hit where the preset carries more (a dual
    # wielder past the yellow cap by sim, standards 4), so only the race's difference moves it; --to-cap: the cap
    target = cap if to_cap else max(cap, _preset_hit[bid])
    amt, gear_hit_rating = set_stats(pl, names)
    gear_hit = gear_hit_rating / HIT_RATING
    excess = round(max(0.0, min(1.0, hit - cap, gear_hit)), 4)
    p = {'build': bid, 'race': race, 'class': meta['class'], 'spec': meta.get('spec'), 'role': 'caster' if caster else 'melee/ranged',
         'hit': round(hit, 2), 'hitFrom': why, 'cap': cap, 'presetHit': round(_preset_hit[bid], 2), 'target': round(target, 2),
         'gearHit': gear_hit, 'excess': excess}
    st = gear_steps(meta, pl)
    if st:
        lo, hi = min(st['steps']), max(st['steps'])
        step = min(max(st['own'] - math.floor(hit - target + 0.5), lo), hi)
        p.update({'gearSet': meta['gearSet'], 'ownStep': st['own'], 'step': step, 'steps': sorted(st['steps']),
                  'excess': st['own'] - step})
        if step != st['own']:
            p['stepSet'] = st['steps'][step]
            p['equipment'] = step_equipment(st, step)
        elif hit - target >= 0.5:
            p['skip'] = 'the set has no lower hit step'
        elif hit - target <= -0.5:
            p['skip'] = 'the set has no higher hit step'
        else:
            p['skip'] = 'on the cap (within half a step)'
        return p, req, meta
    if race != DWARF:
        p['excess'] = 0
        return p, req, meta
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


def describe_step(p):
    if p['excess'] > 0:
        return f"{p['excess']:g}% hit over the cap traded through the set's hit {p['step']:+d} step"
    return f"{-p['excess']:g}% hit under the cap made up with the set's hit {p['step']:+d} step"


def paired(req, bonus, iters, equipment=None):
    """The preset's gear against the swap: bonus stats on top of it, or another gear set (equipment)."""
    def one(r):
        res = run(r, iters, 1, save_all=True)
        return res['raidMetrics']['parties'][0]['players'][0]['dps']
    swap = json.loads(json.dumps(req))
    pl = swap['raid']['parties'][0]['players'][0]
    if equipment is not None:
        pl['equipment'] = json.loads(json.dumps(equipment))
    bs = pl.setdefault('bonusStats', {}) if bonus else {}
    for name, v in (bonus or {}).items():
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


GEAR_KEYS = ('bonusStats', 'gearSet', 'equipment', 'gearNote', 'gearHow', 'gearGain')


def hit_text(p):
    return (f"{p['race']} hit {p['hit']:g}%{' (with ' + p['hitFrom'] + ')' if p['hitFrom'] else ''} against the "
            f"{p['cap']:g}% {'spell' if p['role'] == 'caster' else 'yellow'} cap")


def gear_record(p, eq, iters):
    """The override keys for a kept swap: (gear keys, note, gain or None, how)."""
    gain = ({'dps': p['gain'], 'se': p['se'], 'iterations': iters, 'seed': 1, 'duration': 180} if 'gain' in p else None)
    pair = (f" {p['race']} {p['dpsPreset']:.1f} with the preset's gear, {p['dpsSwap']:.1f} with the swap; gain.se is the "
            f"paired error (the same seeded fights), {p['seUnpaired']:.2f} unpaired." if gain else '')
    if eq:
        more = 'less' if p['excess'] > 0 else 'more'
        how = (f"race_gear.py: {hit_text(p)}. The preset wears {p['gearSet']} (hit step {p['ownStep']:+d}); {p['race']} "
               f"wears the set's step {p['step']:+d}, {p['stepSet']} ({abs(p['excess']) * HIT_RATING:g} {more} gear hit "
               f"rating, traded at the budget prices by orchestration/tools/hit_steps.py)."
               + (pair if gain else ' Under the cap, the higher step is kept by the rule (standards 29), without a sim.'))
        return {'gearSet': p['stepSet'], 'equipment': eq}, p['note'], gain, how
    shares = ', '.join(f'{k} {v}%' for k, v in p['shares'].items())
    how = (f"race_gear.py: {hit_text(p)}, the set's own hit {p['gearHit']:g}%: {p['excess']:g}% "
           f"({p['excess'] * HIT_RATING:g} budget points) re-spent over the set's shares at the 1b prices ({shares}). "
           f"Hit and crit are one pool for the gear, so each is written once." + pair)
    return {'bonusStats': p['bonusStats']}, p['note'], gain, how


def write_file(path, p, eq, iters):
    keys, note, gain, how = gear_record(p, eq, iters)
    old = json.load(open(path)) if os.path.exists(path) else {}
    if 'rotation' in old:  # the race's own rotation: keep it and its note, and add the gear beside it
        out = {k: v for k, v in old.items() if k not in GEAR_KEYS}
        out.update(keys)
        out.update({'gearNote': note, 'gearHow': how})
        if gain:
            out['gearGain'] = gain
        print(f'  {path}: gear merged beside the race\'s rotation, whose pair was simmed on the preset\'s gear: '
              'check it with rotopt', file=sys.stderr)
    else:
        out = dict(keys, note=note, **({'gain': gain} if gain else {}), how=how)
    json.dump(out, open(path, 'w'), indent=1)
    open(path, 'a').write('\n')
    print('  wrote', path)


def drop_gear(path, why):
    """Remove an earlier run's gear from a race's file (the file, when it holds nothing else)."""
    if not os.path.exists(path):
        return
    old = json.load(open(path))
    if not any(k in old for k in GEAR_KEYS):
        return
    if 'rotation' in old:
        json.dump({k: v for k, v in old.items() if k not in GEAR_KEYS}, open(path, 'w'), indent=1)
        open(path, 'a').write('\n')
        print(f'  removed the gear from {path} ({why})')
    else:
        os.remove(path)
        print(f'  removed {path} ({why})')


def main(args):
    write = '--write' in args
    iters = int(args[args.index('--iters') + 1]) if '--iters' in args else 20000
    ids = [a for a in args if not a.startswith('--') and not a.isdigit()]
    if not ids:
        ids = sorted(f[:-10] for f in os.listdir(os.path.join(D, 'req')) if f.endswith('.meta.json'))
    os.environ.setdefault('DASH_SPLITS', '4')
    out_p = os.path.join(D, 'race_gear.json')
    results = json.load(open(out_p)) if os.path.exists(out_p) else {}
    results = {k: v for k, v in results.items() if '.' in k}  # drop the Dwarf-only format's entries (keyed by build)
    for bid in ids:
        meta = json.load(open(os.path.join(D, 'req', bid + '.meta.json')))
        if meta.get('baseBuild'):
            continue  # variants use their base build's files
        for race in CLASS_RACES[meta['class']]:
            if race == meta['race']:
                continue  # the preset's own race is the preset
            p, req, meta = plan(bid, race, '--to-cap' in args)
            eq = p.pop('equipment', None)  # a step's gear set: the override file holds it, race_gear.json names it
            path = race_override.own_path(bid, race)
            if p.get('skip'):
                pass
            elif p['excess'] == 0:
                p['skip'] = ('the set carries no hit' if p['gearHit'] <= 0 else 'no hit over the cap') if race == DWARF \
                    else 'no hit-trade steps'
            elif p['excess'] < 0:
                p.update({'keep': True, 'note': describe_step(p)})  # under the cap: the higher step by the rule
            else:
                rq = at_race(req, race)
                rq['encounter']['duration'] = 180
                base, swap, gain, se, se_unpaired = paired(rq, p.get('bonusStats'), iters, eq)
                p.update({'dpsPreset': round(base, 2), 'dpsSwap': round(swap, 2), 'gain': round(gain, 2), 'se': round(se, 2),
                          'seUnpaired': round(se_unpaired, 2),
                          'keep': gain >= 2 * se, 'note': describe_step(p) if eq else describe(p)})
            results[f'{bid}.{race}'] = p
            json.dump(results, open(out_p, 'w'), indent=1)
            line = f"{bid:40s} {race:20s} hit {p['hit']:5.2f} target {p['target']:5.2f} excess {p['excess']:+.2f}"
            if 'step' in p:
                line += f" step {p['ownStep']:+d} -> {p['step']:+d}"
            if 'gain' in p:
                line += f"  {p['dpsPreset']:8.1f} -> {p['dpsSwap']:8.1f}  gain {p['gain']:+6.2f} ± {p['se']:.2f}  {'KEEP' if p['keep'] else 'drop'}  {p['note']}"
            elif p.get('keep'):
                line += f"  KEEP (rule)  {p['note']}"
            else:
                line += '  ' + p['skip']
            print(line, flush=True)
            if not write:
                continue
            if p.get('keep'):
                write_file(path, p, eq, iters)
            else:
                drop_gear(path, 'loses or ties at 2 SE' if 'gain' in p else p['skip'])


if __name__ == '__main__':
    main(sys.argv[1:])
