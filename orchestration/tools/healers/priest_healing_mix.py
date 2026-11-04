"""The healing priest's stat mix (standards 5a style) from Wowhead's priest healing BiS list, priced at standards 1b.

Run from the repo root (it reads assets/database/db.json). Wowhead's two WoW Forever priest guides (Holy and
Discipline, forever/guide/classes/priest/{holy,discipline}/bis-gear) have every slot table empty ("This section
will be filled in once more items are discovered at max level") and a gear planner of datamined event items
the guide calls a placeholder, so the Classic pre-raid healing priest guide is the one list. Its picks are by
hand below: each slot's first Best, the two-hander path (Redemption), the first two Best rings, no trinkets or
wand.
"""
import json

db = json.load(open('assets/database/db.json'))
ITEMS = {i['id']: i for i in db['items']}
W = {'Sta': 1.0, 'Int': 1.0, 'Spi': 1.0, 'SP': 0.86, 'Heal': 0.45, 'Crit': 1.0, 'MP5': 2.4}
COLS = list(W)


def pts(iid):
    it = ITEMS[iid]
    s = list(it.get('stats') or []) + [0] * 44
    p = dict.fromkeys(COLS, 0.0)
    p['Sta'], p['Int'], p['Spi'] = s[2], s[3], s[4]
    p['SP'] = (s[5] + s[42]) * W['SP'] + sum(s[6:12]) * 0.70
    p['Heal'] = s[41] * W['Heal']
    p['Crit'] = it.get('critRating') or 14 * s[14]
    p['MP5'] = s[12] * W['MP5']
    return it, p


PICKS = [('Head', 13102), ('Shoulders', 22234), ('Back', 18510), ('Chest', 14154), ('Wrists', 18497), ('Hands', 20717),
         ('Waist', 18327), ('Legs', 18386), ('Feet', 22084), ('Neck', 18723), ('Rings', 13178), ('Rings', 22334),
         ('2H Weapons', 22406)]

tot = dict.fromkeys(COLS, 0.0)
print('## Wowhead WoW Classic, healing priest pre-raid BiS')
for slot, iid in PICKS:
    it, p = pts(iid)
    for c, v in p.items():
        tot[c] += v
    print(f'  {slot:<10} {iid:>6} {it["name"][:30]:<30} ilvl {it.get("ilvl")}  ' + ' '.join(f'{c} {v:.0f}' for c, v in p.items() if v))
t = sum(tot.values())
print('  shares: ' + ', '.join(f'{c} {100 * v / t:.1f}' for c, v in tot.items() if v), f'(total {t:.0f})')
