"""The healer stat mix (standards 5a style) from Wowhead's Restoration druid BiS lists, priced at standards 1b."""
import json, os, sys, re
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
from bis_parse import picks
db = json.load(open('assets/database/db.json'))
ITEMS = {i['id']: i for i in db['items']}
SUFFIX = {s['id']: s for s in db['randomSuffixes']}
W = {'Sta': 1.0, 'Int': 1.0, 'Spi': 1.0, 'SP': 0.86, 'SchoolSP': 0.70, 'Heal': 0.45, 'Hit': 1.0, 'Crit': 1.0, 'MP5': 2.4, 'Str': 1.0, 'Agi': 1.0}
COLS = ['Sta', 'Str', 'Agi', 'Int', 'Spi', 'SP', 'Heal', 'Hit', 'Crit', 'MP5']

def stats_of(iid, suffix_name=''):
    it = ITEMS[iid]
    s = list(it.get('stats') or []) + [0] * 44
    if suffix_name:
        opts = [SUFFIX[o] for o in it.get('randomSuffixOptions') or [] if o in SUFFIX and suffix_name.lower() in SUFFIX[o]['name'].lower()]
        if not opts:
            raise SystemExit(f'no {suffix_name} suffix on {iid}')
        best = max(opts, key=lambda o: sum(o['stats']))
        s = [a + b for a, b in zip(s, list(best['stats']) + [0] * 44)]
    return it, s

def pts(it, s):
    p = dict.fromkeys(COLS, 0.0)
    p['Str'], p['Agi'], p['Sta'], p['Int'], p['Spi'] = s[0], s[1], s[2], s[3], s[4]
    p['SP'] = (s[5] + s[42]) * W['SP'] + sum(s[6:12]) * W['SchoolSP']
    p['Heal'] = s[41] * W['Heal']
    p['Hit'] = it.get('hitRating') or 10 * s[13]
    p['Crit'] = it.get('critRating') or 14 * s[14]
    p['MP5'] = s[12] * W['MP5']
    return p

def shares(chosen, label):
    tot = dict.fromkeys(COLS, 0.0)
    print(f'## {label}')
    for slot, iid, suf, *rank in chosen:
        it, s = stats_of(iid, suf)
        p = pts(it, s)
        for c, v in p.items():
            tot[c] += v
        print(f'  {slot:<10} {(rank or [""])[0][:5]:<5} {iid:>7} {it["name"][:30]:<30} {suf:<12} ilvl {it.get("ilvl")}  ' + ' '.join(f'{c} {v:.0f}' for c, v in p.items() if v))
    t = sum(tot.values())
    sh = {c: 100 * v / t for c, v in tot.items()}
    print('  shares: ' + ', '.join(f'{c} {v:.1f}' for c, v in sh.items() if v), f'(total {t:.0f})')
    return sh

# The Forever guide: each slot's first BiS row, the two-hander path (no druid race has a staff or mace
# racial that matters; the synthetic sets use a staff), both BiS rings, trinkets and the relic left out.
fv = picks(sys.argv[1])
chosen, skipped = [], []
def usable(iid, suf):
    it = ITEMS.get(iid)
    if not it:
        return 'not in the item database'
    if it.get('armorType', 0) in (4,):  # plate
        return 'plate'
    if suf and not any(suf.lower() in SUFFIX[o]['name'].lower() for o in it.get('randomSuffixOptions') or [] if o in SUFFIX):
        return f'no "of {suf}" roll'
    return None
by_slot = {}
for slot, rank, items in fv:
    if slot in ('Trinkets', 'Idols', '1H Weapons', 'Offhands'):
        continue
    by_slot.setdefault(slot, []).append((rank, items[0]))
for slot, rows in by_slot.items():
    want = 2 if slot == 'Rings' else 1
    got = 0
    for rank, (iid, text) in rows:
        suf = 'Healing' if 'Healing' in text else ''
        why = usable(iid, suf)
        if why:
            skipped.append(f'{slot} {rank} {iid} ({why})')
            continue
        chosen.append((slot, iid, suf, rank))
        got += 1
        if got == want:
            break
print('skipped:', skipped)
a = shares(chosen, 'Wowhead WoW Forever, Restoration druid Phase 1 BiS')
json.dump(chosen, open(os.devnull, 'w'))

# The Classic pre-raid guide: each slot's first Best, the wrists' first pick that isn't PvP (the slot has
# no Best), the two-hander (Redemption), two rings, no trinkets or idol.
cl = [('Head', 10504, 'Healing'), ('Shoulders', 15061, ''), ('Back', 18510, ''), ('Chest', 13346, ''), ('Wrists', 18525, ''),
      ('Hands', 12554, ''), ('Waist', 14553, ''), ('Legs', 18386, ''), ('Feet', 22247, ''), ('Neck', 18723, ''),
      ('Rings', 13178, ''), ('Rings', 16058, ''), ('2H Weapons', 22406, '')]
b = shares([(s, i, x, 'Best') for s, i, x in cl], 'Wowhead WoW Classic, Restoration druid pre-raid BiS')
avg = {c: (a[c] + b[c]) / 2 for c in COLS}
print('## average of the two: ' + ', '.join(f'{c} {v:.1f}' for c, v in avg.items() if v))
json.dump({'forever': a, 'classic_pre_raid': b, 'average': avg, 'forever_picks': chosen, 'classic_picks': cl},
          open(os.devnull, 'w'), indent=1)
