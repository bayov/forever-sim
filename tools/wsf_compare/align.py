#!/usr/bin/env python3
"""Finds the bonus stats that make theirs' character sheet at the pull match ours, so the DPS gap left is the
mechanics and not the stats. Prints the ALIGN value for run.py (and the residual on stderr).

    align.py <report id | req.json> <out dir> [convert.py args...]
    ALIGN=$(align.py hunter__survival out/sv) run.py hunter__survival out/sv

It matches Strength and Agility first (theirs gets them before Blessing of Kings, so it divides by 1.1), then
attack power and ranged attack power, and for casters Intellect, Spirit, spell power and MP5 (STATS; shamans take both). The gaps it
closes are differences too: say what they were (race bases, buff ranks, how a rating converts) when reporting.
"""
import json, os, re, subprocess, sys

HERE = os.path.dirname(os.path.abspath(__file__))
BIN = os.path.join(os.environ.get('WSF_WORK', os.path.expanduser('~/wsf_compare')), 'bin')
sys.path.insert(0, HERE)
from run import request_path

name, out, rest = sys.argv[1], sys.argv[2], sys.argv[3:]
src = request_path(name)
CONV = [sys.executable, os.path.join(HERE, 'convert.py'), src, out] + rest
klass = json.load(open(src))['raid']['parties'][0]['players'][0]['class']

def their_index(stat):
    proto = open(os.path.join(os.environ.get('WSF', os.path.expanduser('~/wowsims-forever')), 'proto/common.proto')).read()
    body = re.search(r'enum Stat \{(.*?)\}', proto, re.S).group(1)
    return {k: int(v) for k, v in re.findall(r'(\w+)\s*=\s*(\d+);', body)}['Stat' + stat]

def sheet(cmd):
    r = subprocess.run(cmd, capture_output=True, text=True)
    m = re.search(r'^PLAYER (\{.*\})', r.stdout, re.M)
    if not m: sys.exit('no sheet from %s: %s' % (cmd[0], (r.stdout + r.stderr)[-400:]))
    s = json.loads(m.group(1).replace(',}', '}'))
    if 'SpellPower' in s: s['SpellPower'] += s.get('SpellDamage', 0)  # ours: add its spell damage only stat
    return s

# Our sheet names a stat as FlatString does; theirs the same, mostly. (ours name, theirs name, first pass?)
PHYSICAL = [('Strength', 'Strength', True), ('Agility', 'Agility', True), ('AttackPower', 'AttackPower', False),
            ('RangedAttackPower', 'RangedAttackPower', False)]
# Theirs' SpellDamage is both our SpellPower and our SpellDamage (spell damage only), which sheet() adds up.
CASTER = [('Intellect', 'Intellect', True), ('Spirit', 'Spirit', True), ('SpellPower', 'SpellDamage', False), ('MP5', 'MP5', False)]
STATS = CASTER if klass in ('ClassMage', 'ClassWarlock', 'ClassPriest') else PHYSICAL if klass == 'ClassHunter' else PHYSICAL[:3]
if klass in ('ClassShaman', 'ClassPaladin'):  # both: Elemental by its spell power, Enhancement and Retribution by
    STATS = PHYSICAL[:3] + CASTER  # their attack power too (Retribution's Consecration and Exorcism by spell power)
if klass == 'ClassDruid' and '"balanceDruid"' in open(src).read():
    STATS = CASTER

subprocess.run(CONV, check=True, capture_output=True)
ours_bin = 'ours_aq_stats' if klass == 'ClassHunter' and os.environ.get('OURS_AQ', '1') == '1' else 'ours_stats'
ours = sheet([os.path.join(BIN, ours_bin), f'{out}/req_ours.json'])
idx = {t: their_index(t) for _, t, _ in STATS}
align = {t: 0.0 for _, t, _ in STATS}
for it in range(5):
    env = dict(os.environ, ALIGN=','.join(f'{idx[k]}={v}' for k, v in align.items()))
    subprocess.run(CONV, check=True, capture_output=True, env=env)
    th = sheet([os.path.join(BIN, 'theirs_stats'), f'{out}/req_theirs.json', f'{out}/items.json'])
    if json.load(open(f'{out}/req_theirs.json'))['raid']['buffs'].get('battleShout'):
        th['AttackPower'] = th.get('AttackPower', 0) + 115  # their Battle Shout VI lands just after the pull's sheet
    diff = {t: ours.get(o, 0) - th.get(t, 0) for o, t, _ in STATS}
    if all(abs(v) < 0.5 for v in diff.values()): break
    for o, t, first in STATS:
        if first: align[t] += diff[t] / 1.1  # under Blessing of Kings
        elif it > 0: align[t] += diff[t]  # once the primaries match
print(','.join(f'{idx[k]}={round(v, 2)}' for k, v in align.items()))
print('residual', {k: round(v, 1) for k, v in diff.items()}, 'added', {k: round(v, 1) for k, v in align.items()}, file=sys.stderr)
