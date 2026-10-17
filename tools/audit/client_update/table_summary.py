"""usage: AUDIT_WORK=<scratch> python3 -I table_summary.py <oldbuild> <newbuild>
Per table in tables.txt: identical, or rows added / removed / changed by ID (VerifiedBuild ignored).
The first look at a new build: a table marked identical needs no further reading."""
import csv, os, sys
csv.field_size_limit(10**9)
D = os.path.join(os.environ['AUDIT_WORK'], 'db2')
old, new = sys.argv[1], sys.argv[2]
tables = open(os.path.join(os.path.dirname(os.path.abspath(__file__)), 'tables.txt')).read().split()

def load(t, b):
    p = os.path.join(D, f'{t}.{b}.csv')
    if not os.path.exists(p):
        return None
    with open(p, newline='', encoding='utf-8') as f:
        r = csv.DictReader(f)
        return {row['ID']: {k: v for k, v in row.items() if k != 'VerifiedBuild'} for row in r}

for t in tables:
    a, b = load(t, old), load(t, new)
    if a is None or b is None:
        print(f'{t:28} missing ({old}: {a is not None}, {new}: {b is not None})')
        continue
    add = len(set(b) - set(a)); rem = len(set(a) - set(b))
    chg = sum(1 for k in set(a) & set(b) if a[k] != b[k])
    print(f'{t:28} ' + ('identical' if not (add or rem or chg) else f'added {add} removed {rem} changed {chg}'))
