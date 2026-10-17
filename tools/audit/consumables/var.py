# usage: python3 -I var.py <base.json> <outprefix> 'name:{json edits}'...   Makes sim request variants of a mksettings JSON.
# edits: {"consumes":{...},"delconsumes":[...],"bonus":{statIdx:val},"ench":{slot:id or 0},"iters":N}
import json,sys
base=json.load(open(sys.argv[1])); pre=sys.argv[2]
for arg in sys.argv[3:]:
    name,ed=arg.split(':',1); ed=json.loads(ed)
    d=json.loads(json.dumps(base)); p=d['raid']['parties'][0]['players'][0]
    d['simOptions']['iterations']=ed.get('iters',20000)
    p['consumes'].update(ed.get('consumes',{}))
    for k in ed.get('delconsumes',[]): p['consumes'].pop(k,None)
    for k,v in ed.get('ench',{}).items():
        it=p['equipment']['items'][int(k)]
        if v: it['enchant']=v
        else: it.pop('enchant',None)
    if 'bonus' in ed:
        st=[0.0]*45
        for k,v in ed['bonus'].items(): st[int(k)]=v
        p['bonusStats']={'stats':st}
    json.dump(d,open(f'{pre}_{name}.json','w'))
    print(f'{pre}_{name}.json')
