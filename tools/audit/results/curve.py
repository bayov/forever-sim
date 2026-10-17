# Usage: python3 -I curve.py <build> <spellId> [<spellId> ...]   (build like 1.60.1.70450 or 70450)
# Prints the talent-rank curve points (TraitDefinitionEffectPoints -> CurvePoint) for each spell, from the Forever client
# tables TraitDefinition, TraitDefinitionEffectPoints and CurvePoint at that build (tools/sources/wago.py downloads them).
import os, sys
sys.path.insert(0, os.path.join(os.path.dirname(os.path.abspath(__file__)), '../../sources'))
import wago  # noqa: E402
B = sys.argv.pop(1)
td = {}
for r in wago.rows('TraitDefinition', B):
    td.setdefault(r['SpellID'], []).append(r['ID'])
ep = {}
for r in wago.rows('TraitDefinitionEffectPoints', B):
    ep.setdefault(r['TraitDefinitionID'], []).append(r)
cp = {}
for r in wago.rows('CurvePoint', B):
    cp.setdefault(r['CurveID'], []).append((int(r['OrderIndex']), float(r['Pos_0']), float(r['Pos_1'])))
for s in sys.argv[1:]:
    if s not in td:
        print(s, 'no trait')
        continue
    for t in td[s]:
        for e in ep.get(t, []):
            pts = sorted(cp.get(e['CurveID'], []))
            print(s, 'trait', t, 'eff', e['EffectIndex'], 'op', e['OperationType'], ' '.join('%g:%g' % (p[1], p[2]) for p in pts))
