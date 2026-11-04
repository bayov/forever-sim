"""mkvar.py <base.json> <out.json> arg... : one variant request.
  key=Value  sets a consumes field (key=- removes it, true/false for booleans);
  apl=<itemId>@<index>  inserts {"castSpell":{"spellId":{"itemId":N}}} into the APL priority list at index (end = last);
  prof1=<Profession> / prof2=<Profession> sets a profession."""
import json, sys
base, out = sys.argv[1:3]
d = json.load(open(base)); pl = d['raid']['parties'][0]['players'][0]; c = pl['consumes']
ins = []
for a in sys.argv[3:]:
    k, v = a.split('=', 1)
    if k == 'apl':
        it, pos = v.split('@'); ins.append((int(it), pos))
    elif k in ('prof1', 'prof2'): pl['profession' + k[-1]] = v
    elif v == '-': c.pop(k, None)
    elif v in ('true', 'false'): c[k] = v == 'true'
    else: c[k] = v
pri = pl['rotation']['priorityList']
for it, pos in reversed(ins):
    i = len(pri) if pos == 'end' else int(pos)
    pri.insert(i, {"action": {"castSpell": {"spellId": {"itemId": it}}}, "notes": "Explosive (round 4 search)."})
json.dump(d, open(out, 'w'), indent=1)
