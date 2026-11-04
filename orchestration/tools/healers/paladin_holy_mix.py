"""The Holy paladin's healer stat mix (standards 5a style) from Wowhead's Holy paladin BiS lists, priced at
standards 1b, the way shaman_restoration_mix.py prices the Restoration shaman's.

    python3 -I orchestration/tools/healers/paladin_holy_mix.py FOREVER_GUIDE.html

Run from the repo root (it reads assets/database/db.json). The WoW Forever Holy paladin guide
(forever/guide/classes/paladin/holy/bis-gear) lists no item yet: every slot's table reads "This section
will be filled in once more items are discovered at max level", and its gear planner link holds Wowhead's
obfuscated beta item IDs, which no item in our database has. So the mix is the Classic pre-raid list's
alone (classic/guide/paladin-healing-pre-raid-best-in-slot-bis-gear-wow-classic), each slot's first Best
pick, listed by hand below; the Forever guide's stat advice ("heavy healing power and intellect items ...
a good amount of spirit") agrees with it. The script checks the Forever page still has no picks.
"""
import json, os, sys
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
from bis_parse import picks
db = json.load(open('assets/database/db.json'))
ITEMS = {i['id']: i for i in db['items']}
# shaman_restoration_mix.py's weights and columns (standards 1b, healing on top of spell power 0.45).
W = {'Sta': 1.0, 'Int': 1.0, 'Spi': 1.0, 'SP': 0.86, 'SchoolSP': 0.70, 'Heal': 0.45, 'Hit': 1.0, 'Crit': 1.0, 'MP5': 2.4, 'Str': 1.0, 'Agi': 1.0}
COLS = ['Sta', 'Str', 'Agi', 'Int', 'Spi', 'SP', 'Heal', 'Hit', 'Crit', 'MP5']

def pts(it):
    s = list(it.get('stats') or []) + [0] * 44
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
    for slot, iid, rank in chosen:
        it = ITEMS[iid]
        p = pts(it)
        for c, v in p.items():
            tot[c] += v
        print(f'  {slot:<10} {rank:<5} {iid:>7} {it["name"][:30]:<30} ilvl {it.get("ilvl")}  ' + ' '.join(f'{c} {v:.0f}' for c, v in p.items() if v))
    t = sum(tot.values())
    sh = {c: 100 * v / t for c, v in tot.items()}
    print('  shares: ' + ', '.join(f'{c} {v:.1f}' for c, v in sh.items() if v), f'(total {t:.0f})')
    return sh

if __name__ == '__main__':
    fv = [p for p in picks(sys.argv[1]) if p[2]]
    print('Forever guide picks:', fv or 'none')
    # Each slot's first Best of the Classic pre-raid list: main hand and held off-hand (paladins have no
    # staff), two rings, no trinkets or libram.
    cl = [('Head', 18490), ('Shoulders', 14548), ('Back', 18510), ('Chest', 13346), ('Wrists', 18459),
          ('Hands', 18527), ('Waist', 14553), ('Legs', 18386), ('Feet', 18507), ('Neck', 18723),
          ('Rings', 16058), ('Rings', 13178), ('1H Weapons', 11923), ('Offhands', 18523)]
    shares([(s, i, 'Best') for s, i in cl], 'Wowhead WoW Classic, Holy paladin pre-raid BiS')
