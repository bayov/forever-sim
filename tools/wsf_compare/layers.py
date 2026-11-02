#!/usr/bin/env python3
"""Builds the two character sheets up one layer at a time (naked without talents, then talents, gear, consumables,
buffs) and prints both at each, so a stat gap shows which layer brings it.

    layers.py <out dir>      reads <out>/req_ours.json, req_theirs.json and items.json (convert.py's)
"""
import copy, json, os, re, subprocess, sys

BIN = os.path.join(os.environ.get('WSF_WORK', os.path.expanduser('~/wsf_compare')), 'bin')
out = sys.argv[1]
O = json.load(open(f'{out}/req_ours.json')); T = json.load(open(f'{out}/req_theirs.json'))
klass = O['raid']['parties'][0]['players'][0]['class']
CLASS_KEY = klass.removeprefix('Class').lower()

def strip(r, ours, level):
    r = copy.deepcopy(r); p = r['raid']['parties'][0]['players'][0]
    if level < 4:
        r['raid']['buffs'] = {}; r['raid']['debuffs'] = {}; p['buffs'] = {}; r['raid']['parties'][0]['buffs'] = {}
    if level < 3:
        if ours: p['consumes'] = {}
        else: p['consumables'] = {}
    if level < 2:
        p['equipment'] = {'items': []}
        if not ours: p.pop('bonusStats', None)
    if level < 1: p['talentsString'] = ''
    return r

def sheet(cmd):
    r = subprocess.run(cmd, capture_output=True, text=True)
    return json.loads(re.search(r'^PLAYER (\{.*\})', r.stdout, re.M).group(1).replace(',}', '}'))

# ours name | theirs name, where they differ
KEYS = ['Strength', 'Agility', 'Stamina', 'Intellect', 'Spirit', 'AttackPower', 'RangedAttackPower',
        'MeleeHit|PhysicalHitPercent', 'MeleeCrit|PhysicalCritPercent', 'SpellPower|SpellDamage', 'NaturePower|NatureDamage',
        'FirePower|FireDamage', 'FrostPower|FrostDamage', 'ShadowPower|ShadowDamage', 'ArcanePower|ArcaneDamage',
        'SpellHit|SpellHitPercent', 'SpellCrit|SpellCritPercent', 'MP5', 'Armor', 'Health', 'Mana']
ours_bin = 'ours_aq_stats' if klass == 'ClassHunter' and os.environ.get('OURS_AQ', '1') == '1' else 'ours_stats'
tmp_o, tmp_t = f'{out}/layer_ours.json', f'{out}/layer_theirs.json'
for level, name in enumerate(['naked, no talents', '+ talents', '+ gear', '+ consumables', '+ buffs']):
    json.dump(strip(O, True, level), open(tmp_o, 'w')); json.dump(strip(T, False, level), open(tmp_t, 'w'))
    a = sheet([os.path.join(BIN, ours_bin), tmp_o])
    a['SpellPower'] = a.get('SpellPower', 0) + a.get('SpellDamage', 0)  # theirs' SpellDamage is both of ours
    b = sheet([os.path.join(BIN, 'theirs_stats'), tmp_t, f'{out}/items.json'])
    cells = []
    for k in KEYS:
        ko, kt = (k.split('|') + [k])[:2]
        va, vb = a.get(ko, 0), b.get(kt, 0)
        if va or vb:
            cells.append(f'{ko[:10]} {va:.1f}/{vb:.1f}' + (' *' if abs(va - vb) > 0.05 else ''))
    print(f'{name:18s}', ' | '.join(cells))
print('ours/theirs; * marks a difference')
