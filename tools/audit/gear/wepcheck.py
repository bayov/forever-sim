# Standards 22 weapon check: every synthetic weapon used by a set, its type/hand, damage, speed and DPS,
# against the same item in a previous extract (DPS must stay within 0.02, damage is rounded to 0.1, when speed or type changes).
# Usage: python3 wepcheck.py NEWDIR [OLDDIR]
import sys,glob,json,os
sys.path.insert(0,os.path.dirname(__file__)); import audit
new=sys.argv[1]; old=sys.argv[2] if len(sys.argv)>2 else None
used=set()
for f in glob.glob(new+'/*.gear.json'):
    for i in json.load(open(f))['items']:
        if i.get('id'): used.add(i['id'])
def load(d):
    r={}
    for f in glob.glob(d+'/*.go'):
        for it in audit.parse(f): r[it['id']]=it
    return r
N=load(new); O=load(old) if old else {}
dps=lambda it:(it['WeaponDamageMin']+it['WeaponDamageMax'])/2/it['WeaponSpeed']
for iid in sorted(used):
    it=N.get(iid)
    if not it or not it['WeaponSpeed']: continue
    o=O.get(iid); line=f"{iid} {it['name'][:36]:36} {it['WeaponType'] or it['RangedWeaponType']} {it['HandType']} {it['WeaponDamageMin']}-{it['WeaponDamageMax']}@{it['WeaponSpeed']} dps {dps(it):.2f} spread {it['WeaponDamageMax']/it['WeaponDamageMin']:.3f}"
    if o and o['WeaponSpeed']:
        chg=(o['WeaponSpeed'],o['WeaponType'],o['WeaponDamageMin'])!=(it['WeaponSpeed'],it['WeaponType'],it['WeaponDamageMin'])
        line+=f" | was {o['WeaponType'] or o['RangedWeaponType']} @{o['WeaponSpeed']} dps {dps(o):.2f}"+(' CHANGED' if chg else '')+(' DPS-DRIFT!' if abs(dps(o)-dps(it))>0.02 else '')
    elif old: line+=' | NEW'
    print(line)
