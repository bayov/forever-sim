#!/usr/bin/env python3
"""The Tanks rows of orchestration/gear_targets.md (standards 5a, 36), the way gear_targets_v3.py made the DPS rows.

Reads wowsims/sod's level 60 tank gear sets (ui/<spec>/gear_sets, trinkets excluded, AP and RAP counted once,
SoD's weapon feral AP excluded, as the DPS rows) and prices them at standards 1b with standards 36's tank prices:
Defense 1.5 a point, dodge 12, parry 20 and block chance 5 a 1%, block value 0.4, resistance 1.0 a school
(tank_prices.py fits them on Era level 60 items). SoD's expertise and haste aren't priced in 1b and drop out, as
for the DPS rows. Armor is free (1b); SoD's db keeps no bonus armor apart from the base, so none is counted.
Target per spec: Stamina the spec's SoD average share; Spirit 10% for mana users (standards 5: paladin, shaman);
resistances 0 (fight gear, see standards 36); the other stats in the SoD average proportions, re-normalised.
Then the alternatives' effect: each row's tank stat amounts on a whole synthetic set (standards 1b budgets, the
class's archetype f, a 0.72 x L(71) shield) at the standard's prices and at each alternative price set.
Usage: gear_targets_tanks.py SOD_UI_DIR [SOD_DB]
"""
import json, os, statistics, sys

SOD_UI = sys.argv[1]
SOD_DB = sys.argv[2] if len(sys.argv) > 2 else os.path.join(SOD_UI, '..', 'assets', 'database', 'db.json')
DB = {i['id']: i for i in json.load(open(SOD_DB))['items']}
HERE = os.path.dirname(os.path.abspath(__file__))
FIT = json.load(open(os.path.join(HERE, 'tank_prices.json')))

STD = {'Def': 1.5, 'Dodge': 12.0, 'Parry': 20.0, 'Block': 5.0, 'BlockV': 0.4, 'Res': 1.0}
ALT = {
    'standard 36': STD,
    'the fit as fitted (p 1.7)': {'Def': FIT['prices']['def'], 'Dodge': FIT['prices']['dodge'], 'Parry': FIT['prices']['parry'],
                                  'Block': FIT['prices']['block'], 'BlockV': FIT['prices']['blockv'], 'Res': FIT['prices']['res']},
    'p 1.5 fit': {k2: FIT['alternatives']['p 1.5, blues and epics'][k] for k, k2 in
                  (('def', 'Def'), ('dodge', 'Dodge'), ('parry', 'Parry'), ('block', 'Block'), ('blockv', 'BlockV'), ('res', 'Res'))},
    'bis_budget_rules.md (blockv 0.15)': {'Def': 1.56, 'Dodge': 12.4, 'Parry': 20.5, 'Block': 4.5, 'BlockV': 0.15, 'Res': 1.0},
}
BASE = {'Sta': 1.0, 'Str': 1.0, 'Agi': 1.0, 'Int': 1.0, 'Spi': 1.0, 'SP': 0.86, 'SchoolSP': 0.70, 'Hit': 1.0, 'Crit': 1.0, 'MP5': 2.4}
COLS = ['Sta', 'Str', 'Agi', 'Int', 'Spi', 'AP', 'SP', 'Hit', 'Crit', 'MP5', 'Def', 'Dodge', 'Parry', 'Block', 'BlockV', 'Res']
TANKC = ['Def', 'Dodge', 'Parry', 'Block', 'BlockV']


def pts(it, P):
    """Weighted points per column (amount x price)."""
    s = (it.get('stats') or []) + [0] * 44
    p = dict.fromkeys(COLS, 0.0)
    p['Str'], p['Agi'], p['Sta'], p['Int'], p['Spi'] = s[0], s[1], s[2], s[3], s[4]
    p['SP'] = (s[5] + s[42]) * BASE['SP'] + sum(s[6:12]) * BASE['SchoolSP']
    ap, rap = s[17], s[27]
    p['AP'] = max(ap * 0.5, rap * 0.4) if ap and rap else ap * 0.5 + rap * 0.4
    p['Hit'] = it.get('hitRating') or 10 * max(s[18], s[13])
    p['Crit'] = it.get('critRating') or 14 * max(s[19], s[14])
    p['MP5'] = s[12] * BASE['MP5']
    p['Def'] = s[28] * P['Def']
    p['Block'] = s[29] * P['Block']
    p['BlockV'] = s[30] * P['BlockV']
    p['Dodge'] = s[31] * P['Dodge']
    p['Parry'] = s[32] * P['Parry']
    p['Res'] = sum(s[35:40]) * P['Res']
    return p


def shares(path, P):
    tot = dict.fromkeys(COLS, 0.0)
    for slot, e in enumerate(json.load(open(path)).get('items', [])):
        if not e or not e.get('id') or slot in (12, 13) or e['id'] not in DB:
            continue
        for c, v in pts(DB[e['id']], P).items():
            tot[c] += v
    t = sum(tot.values()) or 1
    return {c: 100 * v / t for c, v in tot.items()}


def avg(files, P):
    rows = [shares(os.path.join(SOD_UI, f), P) for f in files]
    return {c: statistics.mean(r[c] for r in rows) for c in COLS}, len(rows)


L = lambda il: 1.125 * il + 2.0
F = {'physical': (0.950, 0.959, 0.911), 'hybrid': (1.093, 1.103, 1.048)}


def armor(f):
    return 3 * L(66) * f[0] + 3 * .75 * L(66) * f[1] + .75 * L(71) * f[1] + 5 * .5625 * L(71) * f[2]


SHIELD = 0.72 * L(71)
G = lambda d, fs: [f'{d}/gear_sets/{f}' for f in fs]
# (name, lists, mana user, armor archetype, weapons and shield budget, the weapons in words)
SPECS = [
    ('Protection warrior', G('tank_warrior', ['phase_4_tanky.gear.json', 'phase_4_damage.gear.json']), False, 'physical',
     0.39 * L(71) + SHIELD + 0.24 * L(71), 'one-hander + shield + ranged'),
    ('Protection paladin', G('protection_paladin', ['p4prot.gear.json', 'p5prot.gear.json']), True, 'hybrid',
     0.45 * L(71) + SHIELD, 'one-hander + shield + libram'),
    ('Feral tank', G('feral_tank_druid', ['phase_6.gear.json', 'phase_7_prebis.gear.json', 'phase_7.gear.json']), False, 'physical',
     0.90 * L(76), 'two-hander + idol'),
    ('Warden shaman', G('warden_shaman', ['phase_4_enh_tank.gear.json']), True, 'hybrid',
     0.45 * L(71) + SHIELD, 'one-hander + shield + totem'),
]
ALSO = [('Warden shaman, SoD Elemental tank list (not used)', G('warden_shaman', ['phase_4_ele_tank.gear.json']), True, 'hybrid',
         0.45 * L(71) + SHIELD, '')]


def target(a, mana):
    sta = a['Sta']
    fixed = {'Sta': sta, 'Spi': 10.0 if mana else 0.0, 'Res': 0.0}
    rest = [c for c in COLS if c not in fixed]
    room = 100 - sum(fixed.values())
    base = sum(a[c] for c in rest) or 1
    return dict(fixed, **{c: a[c] * room / base for c in rest})


out = []
print('| Set | SoD lists | ' + ' | '.join(COLS) + ' | top stat / non-Sta |')
for name, files, mana, arch, weap, words in SPECS + ALSO:
    a, n = avg(files, STD)
    t = target(a, mana)
    nonsta = 100 - t['Sta']
    top = max((c for c in COLS if c != 'Sta'), key=lambda c: t[c])
    budget = armor(F[arch]) + weap
    amounts = {}
    for lab, P in ALT.items():
        aa, _ = avg(files, P)
        tt = target(aa, mana)
        amounts[lab] = {c: tt[c] / 100 * budget / P[c] for c in TANKC}
        amounts[lab]['share'] = sum(tt[c] for c in TANKC)
        amounts[lab]['Sta'] = tt['Sta'] / 100 * budget
        amounts[lab]['other'] = (100 - tt['Sta'] - amounts[lab]['share']) / 100 * budget
    ftank = {'tank f, Era tank epics': (1.082, 1.069, 0.931), 'tank f, Era tank epics without resistance': (1.021, 1.024, 0.927)}
    fb = {k: armor(v) + weap for k, v in ftank.items()}
    keep_res = a['Res'] / 100 * budget
    out.append({'spec': name, 'lists': n, 'files': files, 'sod': a, 'target': t, 'arch': arch, 'budget': budget, 'weapons': words,
                'amounts': amounts})
    print(f'| {name} | {n} | ' + ' | '.join(f'{t[c]:.1f}' if t[c] >= 0.05 else '-' for c in COLS) + f' | {top} {100 * t[top] / nonsta:.0f}% |')
    print(f'    SoD own: ' + ', '.join(f'{c} {a[c]:.1f}' for c in COLS if a[c] >= 0.05))
    print(f'    set budget {budget:.1f} ({arch} armor {armor(F[arch]):.1f} + {words} {weap:.1f})')
    for lab, am in amounts.items():
        print(f'      {lab:36}: tank share {am["share"]:.1f}%  ' + ', '.join(f'{c} {am[c]:.1f}' for c in TANKC if am[c] > 0)
              + f'; Sta {am["Sta"]:.0f}, other stats {am["other"]:.1f} points')
    for k, b in fb.items():
        print(f'      budget at {k}: {b:.1f} ({100 * (b / budget - 1):+.1f}%)')
    print(f'      resistances kept at SoD\'s share: {keep_res:.0f} resistance points, taken from the other stats')
json.dump(out, open(os.path.join(HERE, 'gear_targets_tanks.json'), 'w'), indent=1)
