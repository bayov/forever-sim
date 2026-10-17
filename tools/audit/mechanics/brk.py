# Usage: python3 brk.py <wowsimcli output.json>
# Prints DPS and a per-action damage breakdown from a CLI result.
import json,sys
d=json.load(open(sys.argv[1]))
p=d['raidMetrics']['parties'][0]['players'][0]
print('DPS %.1f +- %.1f'%(p['dps']['avg'],p['dps'].get('stdev',0)))
rows={}
for a in p['actions']:
    i=a['id']; k=str(i.get('spellId') or i.get('itemId') or i.get('otherId'))+('/'+str(i['tag']) if i.get('tag') else '')
    dmg=sum(t.get('damage',0) for t in a.get('targets',[]))
    hits=sum(t.get('hits',0)+t.get('crits',0) for t in a.get('targets',[]))
    casts=sum(t.get('casts',0) for t in a.get('targets',[]))
    rows[k]=(dmg,hits,casts)
tot=sum(r[0] for r in rows.values())
cum=0
for k,(dm,h,c) in sorted(rows.items(),key=lambda x:-x[1][0]):
    if dm<=0: continue
    cum+=dm
    print('%-22s %5.1f%% cum %5.1f%% hits %d casts %d'%(k,100*dm/tot,100*cum/tot,h,c))
    if cum/tot>0.97: break
