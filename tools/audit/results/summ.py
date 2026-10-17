# Usage: python3 summ.py <wowsimcli out.json>
# Per-action breakdown of the first player (and its pets): damage share, casts, hits, crits, ticks, misses, average per hit.
import json,sys
d=json.load(open(sys.argv[1]))
p=d['raidMetrics']['parties'][0]['players'][0]
dur=None
dps=p['dps']['avg']; print("DPS %.1f  (stdev %.1f)"%(dps,p['dps'].get('stdev',0)))
def acts(u,label):
    rows=[]
    for a in u.get('actions',[]):
        i=a['id']; key=i.get('spellId') or i.get('itemId') or i.get('otherId'); tag=i.get('tag',0)
        dmg=sum(t.get('damage',0) for t in a.get('targets',[])); casts=sum(t.get('casts',0) for t in a.get('targets',[]))
        hits=sum(t.get('hits',0) for t in a.get('targets',[])); crits=sum(t.get('crits',0) for t in a.get('targets',[]))
        ticks=sum(t.get('ticks',0) for t in a.get('targets',[])); ctk=sum(t.get('critTicks',0) for t in a.get('targets',[]))
        miss=sum(t.get('misses',0)+t.get('dodges',0)+t.get('parries',0) for t in a.get('targets',[]))
        rows.append((dmg,label,key,tag,casts,hits,crits,ticks,ctk,miss))
    return rows
rows=acts(p,'P')
for pet in p.get('pets',[]): rows+=acts(pet,'pet:'+pet['name'])
tot=sum(r[0] for r in rows)
it=None
for r in sorted(rows,key=lambda r:-r[0]):
    if r[0]<=0 and r[4]==0: continue
    print("%-14s %-8s t%-2s dmg%%=%5.1f casts=%7.1f hits=%7.1f crit=%6.1f ticks=%6.1f ctick=%5.1f miss=%5.1f avg=%s"%(r[1][:14],r[2],r[3],100*r[0]/tot if tot else 0,r[4],r[5],r[6],r[7],r[8],r[9], round(r[0]/(r[5]+r[6]+r[7]),1) if (r[5]+r[6]+r[7]) else '-'))
