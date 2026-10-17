"""usage: python3 -I rank.py <build> <effectIndex> <spellID>...
Damage of one effect of each spell rank at level 60: base points plus per-level growth up to the
rank's max level, the variance range, spell coefficient, mana cost and period. What the sim's
rank tables (sim/<class>/*.go) should match."""
import csv, os, sys
sys.path.insert(0, os.path.join(os.path.dirname(os.path.abspath(__file__)), '../../sources'))
import wago  # noqa: E402
b, idx, ids = sys.argv[1], int(sys.argv[2]), sys.argv[3:]
def rows(t): return wago.rows(t, b)
eff = {(r['SpellID'], int(r['EffectIndex'])): r for r in rows('SpellEffect') if r['SpellID'] in ids}
lv = {r['SpellID']: r for r in rows('SpellLevels') if r['SpellID'] in ids}
pw = {r['SpellID']: r for r in rows('SpellPower') if r['SpellID'] in ids}
for i in ids:
    e = eff.get((i, idx)); l = lv.get(i, {})
    if not e: print(i, 'no effect', idx); continue
    bp, ppl, var = float(e['EffectBasePointsF']), float(e['EffectRealPointsPerLevel']), float(e['Variance'])
    base, mx = int(l.get('BaseLevel') or 0), int(l.get('MaxLevel') or 0)
    avg = bp + ppl * max(0, (min(60, mx) if mx else 60) - base)
    print(i, 'spell level', l.get('SpellLevel'), 'max', mx, 'bp', bp, 'ppl', round(ppl, 3), 'var', round(var, 4),
          'at 60', round(avg, 2), 'range', round(avg * (1 - var / 2), 2), round(avg * (1 + var / 2), 2),
          'coef', e['EffectBonusCoefficient'], 'cost', pw.get(i, {}).get('ManaCost'), 'period', e['EffectAuraPeriod'])
