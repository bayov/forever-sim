# Standards items 2, 6, 7 (+ old 5a heuristics): item levels, 2-piece sets, enchant names/bans, empty head/legs/neck enchants, trinkets present.
# Usage: AUDIT_WORK=<scratch with enchants.json/items.json from gen_db.py> python3 std.py DIR
# IGNORE: the 'cost .. vs ..', 'armor12', 'caster weapons total' and '5a: ..' lines (pre-1b budget model, superseded by lin2.py).
# Known mislabels: '2H Agility (check source)' and 'Gloves - Superior Agility' on Feral/hunter weapons are effect-id collisions, settled.
import json,glob,os,sys
sys.path.insert(0,os.path.dirname(__file__))
from audit import parse,slot,cost,W
D=sys.argv[1]
items={}
for f in glob.glob(D+'/*.go'):
    for it in parse(f): items[it['id']]=it
SL=['head','neck','shoulder','back','chest','wrist','hands','waist','legs','feet','fing1','fing2','trk1','trk2','mh','oh','rng']
IL={'head':66,'shoulder':66,'chest':66,'hands':66,'legs':66,'feet':66,'neck':71,'back':71,'wrist':71,'waist':71,'fing1':71,'fing2':71}
BAD_ENCH={2544:'Arcanum of Focus (Dire Maul)',2543:'Arcanum of Rapidity (Dire Maul)',2545:'Arcanum of Protection (Dire Maul)',2504:'Weapon - Spell Power (MC formula)',2564:'Gloves - Superior Agility (Cenarion, phase 5?)',2646:'2H Agility (check source)',1255074:'Mystic Thick kit (Rugged +6 exists)',1255073:'Forceful Thick kit (Rugged exists)'}
W_=os.environ.get('AUDIT_WORK','/tmp/audit-gear')
EN=json.load(open(os.path.join(W_,'enchants.json')))
RI=json.load(open(os.path.join(W_,'items.json')))
V=lambda il:1.2*il-36.4
for g in sorted(glob.glob(D+'/*.gear.json')):
    j=json.load(open(g))['items']; ench=[]; msgs=[]; sets={}; armor=0; hit=crit=spi=tot=0; lin={}
    caster_w=0
    for i,e in enumerate(j):
        s=SL[i] if i<len(SL) else str(i); iid=e.get('id')
        if e.get('enchant') in BAD_ENCH: msgs.append(f"{s}: enchant {BAD_ENCH[e['enchant']]}")
        if s in ('trk1','trk2') and iid: msgs.append(f'TRINKET {iid} present')
        ench.append(f"{s}={EN.get(str(e.get('enchant')),{}).get('name','-').replace('Enchant ','') if e.get('enchant') else '-'}")
        if not iid:
            if s=='rng': msgs.append('ranged/relic slot EMPTY')
            continue
        if s in ('neck','head','legs') and not e.get('enchant'): msgs.append(f'{s} unenchanted')
        it=items.get(iid)
        if not it:
            if s=='rng': msgs.append(f'rng is real item {iid} {RI.get(str(iid),{}).get("name")} {RI.get(str(iid),{}).get("stats")} (check it does something; il71 x0.32 = {V(71)*.32:.1f})')
            continue
        c=cost(it); il=int(it['Ilvl'])
        if it['set']: sets[it['set']]=sets.get(it['set'],0)+1
        if s in IL:
            armor+=c
            if il!=IL[s]: msgs.append(f"{s} il{il} != {IL[s]}")
            b=V(IL[s])*slot(it)
            if abs(c/b-1)>.02: msgs.append(f"{s} cost {c:.1f} vs {b:.1f} ({100*(c/b-1):+.1f}%)")
        elif s in ('mh','oh'):
            ht=it['HandType']; caster=it['stats'].get('SpellPower',0)>0 and it['stats'].get('Strength',0)+it['stats'].get('Agility',0)+it['stats'].get('AttackPower',0)==0
            if caster: caster_w+=c
            elif ht=='TwoHand':
                if abs(c/54.8-1)>.02: msgs.append(f"{s} 2H cost {c:.1f} vs 54.8 (il{il})")
            else:
                b=V(71)*.42
                if abs(c/b-1)>.02: msgs.append(f"{s} phys 1H cost {c:.1f} vs {b:.1f} (il{il})")
        elif s=='rng':
            b=V(71)*.32
            if abs(c/b-1)>.03: msgs.append(f"rng cost {c:.1f} vs {b:.1f} (il{il})")
        hit+=it['HitRating']; crit+=it['CritRating']; spi+=it['stats'].get('Spirit',0)
        for k,v in it['stats'].items():
            if k!='Armor' and not (k=='RangedAttackPower' and it['stats'].get('AttackPower')==v): tot+=v*W[k]; lin[k]=lin.get(k,0)+v*W[k]
        tot+=it['HitRating']+it['CritRating']; lin['Hit']=lin.get('Hit',0)+it['HitRating']; lin['Crit']=lin.get('Crit',0)+it['CritRating']
    if caster_w and abs(caster_w-73)>2: msgs.append(f"caster weapons total {caster_w:.1f} vs ~73")
    for k,n in sets.items():
        if n!=2: msgs.append(f"set '{k}' has {n} pieces (want 2)")
    if not sets: msgs.append('no set pieces')
    print(f"{os.path.basename(g)[:62]:62} armor12 {armor:5.1f} hit {hit/10:.1f}% crit {crit/14:.1f}% spi {100*spi/tot if tot else 0:.0f}% cw {caster_w:.1f}")
    if tot:
        nonsta=tot-lin.get('Stamina',0); top=max((v,k) for k,v in lin.items() if k!='Stamina')
        others=[k for k,v in lin.items() if k not in ('Stamina','Hit',top[1]) and v/nonsta>=0.03]
        print(f"     5a: Sta {100*lin.get('Stamina',0)/tot:.0f}% of set; top non-Sta stat {top[1]} {100*top[0]/nonsta:.0f}% of non-Sta; other useful stats (>=3%): {','.join(others)}")
        if top[0]/nonsta>0.45: msgs.append(f"5a: {top[1]} is {100*top[0]/nonsta:.1f}% of non-Stamina budget (>45%)")
        sh=lin.get('Stamina',0)/tot
        if not (0.165<=sh<=0.195): msgs.append(f"5a: Stamina {100*sh:.1f}% of set (want 18% +/-1.5)")
        if len(others)<2: msgs.append('5a: fewer than two other stats beyond hit')
    for m in msgs: print('     -',m)
    print('     ench:',' '.join(ench))
