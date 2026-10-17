# Usage: python3 unread_fields.py
# Lists every field set (non-default) in the committed DPS report requests that no non-test Go code under sim/ reads: it maps
# each JSON key to its Go field through the protobuf tags in sim/core/proto/*.pb.go (run build.sh first so they exist), then
# looks for the field name or its getter in sim/ outside proto/. Catches a request carrying a field the sim dropped (the old
# startingTotems kind of drift). Expected output today: only "UNREAD notes" (APL rotation notes, read by the UI only).
# Env: AUDIT_TREE (default: this repo).
import json,glob,re,os
D=os.path.dirname(os.path.abspath(__file__))
os.chdir(os.environ.get('AUDIT_TREE') or os.path.abspath(os.path.join(D,'..','..','..')))
jmap={}
for f in glob.glob('sim/core/proto/*.pb.go'):
    for line in open(f):
        m=re.match(r'\s*(\w+)\s+\S+\s+`protobuf:"([^"]*)"',line)
        if not m: continue
        go,tag=m.groups(); d=dict(kv.split('=',1) for kv in tag.split(',') if '=' in kv)
        jmap.setdefault(d.get('json',d.get('name')),set()).add(go)
if not jmap: raise SystemExit('no sim/core/proto/*.pb.go: run build.sh first')
words=set()
for root,_,files in os.walk('sim'):
    if '/proto' in root: continue
    for fn in files:
        if fn.endswith('.go') and not fn.endswith('_test.go'):
            words.update(re.findall(r'\b[A-Z]\w*\b',open(os.path.join(root,fn)).read()))
keys={}
def walk(x,path):
    if isinstance(x,dict):
        for k,v in x.items():
            if not isinstance(v,(dict,list)) and v not in (False,0,'',None): keys.setdefault(k,set()).add(path)
            walk(v,path+'.'+k)
    elif isinstance(x,list):
        for v in x: walk(v,path+'[]')
for f in glob.glob('tools/dps_report/req/*.json'):
    if f.endswith('.meta.json'): continue
    walk(json.load(open(f)),'')
for k in sorted(keys):
    gos=jmap.get(k)
    if not gos: print('NOT-IN-PROTO', k, sorted(keys[k])[:2]); continue
    if not any(g in words or ('Get'+g) in words for g in gos): print('UNREAD', k, sorted(gos), sorted(keys[k])[:2])
