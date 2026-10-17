"""usage: python3 -I table_summary.py <oldbuild> <newbuild>
Per table in wago.py's client_update set: identical, or rows added / removed / changed by ID (VerifiedBuild ignored).
The first look at a new build: a table marked identical needs no further reading."""
import csv, os, sys
csv.field_size_limit(10**9)
sys.path.insert(0, os.path.join(os.path.dirname(os.path.abspath(__file__)), '../../sources'))
import wago  # noqa: E402
old, new = sys.argv[1], sys.argv[2]
tables = wago.SETS['client_update']

def load(t, b):
    rows = wago.by_id(t, b)
    if rows is None:
        return None
    return {k: {c: v for c, v in row.items() if c != 'VerifiedBuild'} for k, row in rows.items()}

for t in tables:
    a, b = load(t, old), load(t, new)
    if a is None or b is None:
        print(f'{t:28} missing ({old}: {a is not None}, {new}: {b is not None})')
        continue
    add = len(set(b) - set(a)); rem = len(set(a) - set(b))
    chg = sum(1 for k in set(a) & set(b) if a[k] != b[k])
    print(f'{t:28} ' + ('identical' if not (add or rem or chg) else f'added {add} removed {rem} changed {chg}'))
