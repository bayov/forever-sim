"""Print client spell rows by ID. Usage: python3 -I sp.py DIR ID..."""
import csv, sys, os
d = sys.argv[1]; ids = set(sys.argv[2:])
def rows(t, key='SpellID'):
    with open(os.path.join(d, t + '.csv'), encoding='utf-8') as f:
        for r in csv.DictReader(f):
            if r.get(key) in ids: yield r
names = {r['ID']: r['Name_lang'] for r in rows('SpellName', 'ID')}
for i in sys.argv[2:]:
    print('==', i, names.get(i))
    for t in ('SpellAuraOptions', 'SpellClassOptions'):
        for r in rows(t):
            if r['SpellID'] == i: print(' ', t, {k: v for k, v in r.items() if v not in ('0', '')})
    for r in rows('SpellEffect'):
        if r['SpellID'] == i: print('  SpellEffect', {k: v for k, v in r.items() if v not in ('0', '', '0.0')})
