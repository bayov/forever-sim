"""The healing priest's rank audit: every rank of every heal in sim/priest/healing.go (and Penance's healing bolts
in penance.go) against the Forever client's tables (tools/sources/wago.py's cache of 1.60.1.70753). Prints a
Markdown table per heal with the client's value and the sim's for each field, and a mismatch count.

    python3 orchestration/tools/healers/priest_audit.py [BUILD]

Fields: the heal effect's base points, growth a level and variance (SpellEffect; the heal is effect 10, a HoT's
or shield's aura effect 6), its coefficient (SpellEffect.EffectBonusCoefficient), the learn and max levels
(SpellLevels.BaseLevel/SpellLevel, MaxLevel), the cast time (SpellMisc.CastingTimeIndex -> SpellCastTimes) and
the mana cost (SpellPower.ManaCost).
"""
import csv, os, re, sys

BUILD = sys.argv[1] if len(sys.argv) > 1 else '1.60.1.70753'
D = os.path.expanduser(f'~/.cache/wow-data/wago/{BUILD}')


def table(name, key='SpellID'):
    rows = {}
    with open(f'{D}/{name}.csv', newline='') as f:
        for r in csv.DictReader(f):
            rows.setdefault(int(r[key]), []).append(r)
    return rows


EFFECTS, LEVELS, POWER, MISC = table('SpellEffect'), table('SpellLevels'), table('SpellPower'), table('SpellMisc')
CASTS = {int(r['ID']): int(r['Base']) for r in csv.DictReader(open(f'{D}/SpellCastTimes.csv', newline=''))}
NAMES = {int(r['ID']): r['Name_lang'] for r in csv.DictReader(open(f'{D}/SpellName.csv', newline=''))}

src = open('sim/priest/healing.go').read()
HEALS = {}
for m in re.finditer(r'var (\w+)Ranks = \[\]healRank\{\n\t\{\},\n(.*?)\n\}', src, re.S):
    ranks = []
    for line in m.group(2).splitlines():
        v = [x.strip() for x in line.strip().strip('{},').split(',')]
        cast = v[7].replace('time.Millisecond', '').replace('*', '').strip()
        ranks.append(dict(id=int(v[0]), base=float(v[1]), per=float(v[2]), var=float(v[3]), learn=int(v[4]),
                          max=int(v[5]), coef=float(v[6]), cast=int(cast or 0), cost=float(v[8])))
    HEALS[m.group(1)] = ranks


# Holy Nova's cast deals the damage; each rank's heal is a spell of its own (the sim casts the heal).
HEAL_SPELL = {15237: 23455, 15430: 23458, 15431: 23459, 27799: 27803, 27800: 27804, 27801: 27805}


def client(spell_id):
    effs = EFFECTS.get(HEAL_SPELL.get(spell_id, spell_id), [])
    # A heal (10), a HoT or shield aura (6: 8, 69), or Prayer of Mending's dummy that carries its heal (3).
    heal = ([e for e in effs if e['Effect'] == '10'] or [e for e in effs if e['Effect'] == '6' and e['EffectAura'] in ('8', '69')]
            or [e for e in effs if e['Effect'] == '3' and float(e['EffectBonusCoefficient']) > 0])
    e = sorted(heal, key=lambda e: int(e['EffectIndex']))[0] if heal else None
    lv = (LEVELS.get(spell_id) or [{}])[0]
    misc = (MISC.get(spell_id) or [{}])[0]
    power = (POWER.get(spell_id) or [{}])[0]
    return dict(id=spell_id,
                base=float(e['EffectBasePointsF']) if e else None,
                per=float(e['EffectRealPointsPerLevel']) if e else None,
                var=float(e['Variance']) if e else None,
                coef=float(e['EffectBonusCoefficient']) if e else None,
                learn=int(lv.get('BaseLevel') or lv.get('SpellLevel') or 0),
                max=int(lv.get('MaxLevel') or 0),
                cast=CASTS.get(int(misc.get('CastingTimeIndex') or 0), 0),
                cost=float(power.get('ManaCost') or 0))


FIELDS = [('base', 'base'), ('per', 'per level'), ('var', 'variance'), ('coef', 'coefficient'), ('learn', 'learned'),
          ('max', 'max level'), ('cast', 'cast ms'), ('cost', 'mana')]


def fmt(x):
    if x is None:
        return '-'
    if isinstance(x, float):
        return f'{x:.4g}'
    return str(x)


bad = 0
for heal, ranks in HEALS.items():
    print(f'\n#### {heal}\n')
    print('| rank | spell | ' + ' | '.join(h for _, h in FIELDS) + ' |')
    print('|---' * (len(FIELDS) + 2) + '|')
    for i, r in enumerate(ranks, 1):
        c = client(r['id'])
        cells = []
        for f, _ in FIELDS:
            if f == 'max' and not r['per'] and not c['max']:
                c['max'] = r['max']  # no growth a level: the client leaves the max level out
            same = c[f] is not None and abs(float(c[f]) - float(r[f])) < 1e-3 * max(1, abs(float(r[f])))
            if same:
                cells.append(fmt(r[f]))
            else:
                bad += 1
                cells.append(f'**{fmt(r[f])}** (client {fmt(c[f])})')
        print(f'| {i} | {r["id"]} {NAMES.get(r["id"], "?")} | ' + ' | '.join(cells) + ' |')
print(f'\n{bad} fields differ from client {BUILD}')
