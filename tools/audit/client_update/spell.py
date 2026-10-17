"""usage: python3 -I spell.py <build> <spellID or exact name>...
Everything the client says about the spells: each effect (type, aura, base points, variance,
per-level, spell and AP coefficients, misc, trigger, period, class masks), then the non-zero
fields of SpellMisc, SpellCooldowns, SpellPower, SpellAuraOptions, SpellLevels, SpellCategories,
SpellClassOptions, SpellProcsPerMinute, SpellEquippedItems, the cast time and the tooltip text.
Run it on the old and the new build and compare."""
import csv, os, sys
csv.field_size_limit(10**9)
sys.path.insert(0, os.path.join(os.path.dirname(os.path.abspath(__file__)), '../../sources'))
import wago  # noqa: E402
b, args = sys.argv[1], sys.argv[2:]
def rows(t):
    try:
        return wago.rows(t, b)
    except OSError:
        return []
names = {r['ID']: r['Name_lang'] for r in rows('SpellName')}
ids = {a for a in args if a.isdigit()} | {i for i, n in names.items() if n in args}
ct = {r['ID']: r['Base'] for r in rows('SpellCastTimes')}
NOISE = ('0', '', '0.0')
for i in sorted(ids, key=int):
    print(f'=== {i} {names.get(i)}')
    for r in rows('SpellEffect'):
        if r['SpellID'] == i and r['Effect'] != '0':
            print(f"  [{r['EffectIndex']}] eff {r['Effect']} aura {r['EffectAura']} bp {r['EffectBasePointsF']} var {r['Variance']}"
                  f" ppl {r['EffectRealPointsPerLevel']} coef {r['EffectBonusCoefficient']} ap {r['BonusCoefficientFromAP']}"
                  f" misc {r['EffectMiscValue_0']} trig {r['EffectTriggerSpell']} period {r['EffectAuraPeriod']}"
                  f" target {r['ImplicitTarget_0']} mask {r['EffectSpellClassMask_0']},{r['EffectSpellClassMask_1']},{r['EffectSpellClassMask_2']},{r['EffectSpellClassMask_3']}")
    for t in ('SpellMisc', 'SpellCooldowns', 'SpellPower', 'SpellAuraOptions', 'SpellLevels', 'SpellCategories',
              'SpellClassOptions', 'SpellProcsPerMinute', 'SpellEquippedItems', 'SpellInterrupts'):
        for r in rows(t):
            if r.get('SpellID') == i:
                d = {k: v for k, v in r.items() if v not in NOISE and k not in ('ID', 'SpellID', 'VerifiedBuild')}
                if t == 'SpellMisc' and r.get('CastingTimeIndex') in ct: d['castMs'] = ct[r['CastingTimeIndex']]
                print(f'  {t}: {d}')
    for r in rows('Spell'):
        if r['ID'] == i: print('  text:', r['Description_lang'][:400].replace('\n', ' '))
