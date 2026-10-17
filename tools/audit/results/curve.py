# Usage: python3 curve.py <spellId> [<spellId> ...]
# Prints the talent-rank curve points (TraitDefinitionEffectPoints -> CurvePoint) for each spell, from the Forever client
# tables TraitDefinition.csv, TraitDefinitionEffectPoints.csv and CurvePoint.csv in AUDIT_DB (default $AUDIT_WORK/db; see
# orchestration/audit/ref/results/MANIFEST.md for how to download them).
import csv, os, sys
W = os.environ.get('AUDIT_WORK') or os.path.join(os.environ.get('TMPDIR', '/tmp'), 'audit-results')
D = (os.environ.get('AUDIT_DB') or os.path.join(W, 'db')) + '/'
td = {}
for r in csv.DictReader(open(D + 'TraitDefinition.csv')):
    td.setdefault(r['SpellID'], []).append(r['ID'])
ep = {}
for r in csv.DictReader(open(D + 'TraitDefinitionEffectPoints.csv')):
    ep.setdefault(r['TraitDefinitionID'], []).append(r)
cp = {}
for r in csv.DictReader(open(D + 'CurvePoint.csv')):
    cp.setdefault(r['CurveID'], []).append((int(r['OrderIndex']), float(r['Pos_0']), float(r['Pos_1'])))
for s in sys.argv[1:]:
    if s not in td:
        print(s, 'no trait')
        continue
    for t in td[s]:
        for e in ep.get(t, []):
            pts = sorted(cp.get(e['CurveID'], []))
            print(s, 'trait', t, 'eff', e['EffectIndex'], 'op', e['OperationType'], ' '.join('%g:%g' % (p[1], p[2]) for p in pts))
