#!/usr/bin/env python3
# Usage: python3 -I db2_rowdiff.py <db dir> <old buildno> <new buildno> [--show N]
# Row-diffs every B_<old>_<Table>.csv against B_<new>_<Table>.csv by ID: counts of added, removed and
# changed rows per table, then the changed rows field by field (spell names from the new SpellName),
# and the names of added spells. Run with python3 -I: the CSVs are downloaded, untrusted data.
import csv, os, sys
d, old, new = sys.argv[1], sys.argv[2], sys.argv[3]
show = int(sys.argv[sys.argv.index('--show') + 1]) if '--show' in sys.argv else 40
def load(p): return {r['ID']: r for r in csv.DictReader(open(p, newline='', encoding='utf-8'))}
nmf = os.path.join(d, f'B_{new}_SpellName.csv')
nm = {k: r['Name_lang'] for k, r in load(nmf).items()} if os.path.exists(nmf) else {}
def spell(r): return next((r[k] for k in ('SpellID', 'Spell') if k in r), None)
for f in sorted(os.listdir(d)):
    if not (f.startswith(f'B_{old}_') and f.endswith('.csv')): continue
    t = f[len(f'B_{old}_'):-4]; g = os.path.join(d, f'B_{new}_{t}.csv')
    if not os.path.exists(g): print(f'{t}: missing in {new}'); continue
    A, B = load(os.path.join(d, f)), load(g)
    add, rem = sorted(set(B) - set(A)), sorted(set(A) - set(B))
    ch = sorted(i for i in set(A) & set(B) if any(A[i][k] != B[i].get(k) for k in A[i]))
    print(f'{t}: +{len(add)} -{len(rem)} ~{len(ch)}')
    for i in ch[:show]:
        s = spell(B[i]) or (i if t == 'SpellName' else None)
        print(f'   ~{i} {nm.get(s, "") if s else ""}', {k: (A[i][k], B[i].get(k)) for k in A[i] if A[i][k] != B[i].get(k)})
    if t in ('SpellEffect', 'SpellName', 'SkillLineAbility'):
        names = sorted({nm.get(spell(B[i]) or i, '?') for i in add})
        if names: print('   added spells:', ', '.join(names[:show]))
