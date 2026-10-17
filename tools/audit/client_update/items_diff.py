"""usage: AUDIT_TREE=<checkout> python3 -I items_diff.py <oldbuild> <newbuild>
ItemSparse between two builds: added items (name, ilvl, required level, quality) and changed items
with the changed fields. Flags items worn by a DPS-report request (tools/dps_report/req/*.json in
AUDIT_TREE) and items at ilvl >= 55 or required level >= 50, the ones a level 60 preset could use."""
import csv, glob, json, os, sys
csv.field_size_limit(10**9)
sys.path.insert(0, os.path.join(os.path.dirname(os.path.abspath(__file__)), '../../sources'))
import wago  # noqa: E402
old, new = sys.argv[1], sys.argv[2]
SKIP = {'VerifiedBuild', 'SellPrice', 'BuyPrice'}

def load(b):
    return wago.by_id('ItemSparse', b)

used = {}
tree = os.environ.get('AUDIT_TREE')
if tree:
    for f in glob.glob(os.path.join(tree, 'tools/dps_report/req/*.json')):
        if f.endswith('.meta.json'):
            continue
        p = json.load(open(f))['raid']['parties'][0]['players'][0]
        for it in p.get('equipment', {}).get('items', []):
            if it and it.get('id'):
                used.setdefault(str(it['id']), set()).add(os.path.basename(f)[:-5])

a, b = load(old), load(new)
def tag(k, r):
    t = []
    if k in used: t.append(f'USED BY {len(used[k])} req: ' + ', '.join(sorted(used[k])[:4]))
    if int(r['ItemLevel'] or 0) >= 55 or int(r['RequiredLevel'] or 0) >= 50: t.append('level-60 range')
    return ('  <-- ' + '; '.join(t)) if t else ''
added = sorted(set(b) - set(a), key=int)
removed = sorted(set(a) - set(b), key=int)
print(f'added {len(added)} removed {len(removed)}')
for k in added:
    r = b[k]; print(f'  + {k} {r["Display_lang"]} ilvl {r["ItemLevel"]} req {r["RequiredLevel"]} q{r["OverallQualityID"]}{tag(k, r)}')
for k in removed:
    print(f'  - {k} {a[k]["Display_lang"]}{tag(k, a[k])}')
chg = 0
for k in sorted(set(a) & set(b), key=int):
    d = {x: (a[k].get(x), b[k].get(x)) for x in set(a[k]) | set(b[k]) if x not in SKIP and a[k].get(x) != b[k].get(x)}
    if d:
        chg += 1; print(f'  ~ {k} {b[k]["Display_lang"]} {d}{tag(k, b[k])}')
print(f'changed {chg}')
