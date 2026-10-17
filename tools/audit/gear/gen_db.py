# Builds the two lookup tables std.py reads from the checkout's assets/database/db.json.
# Usage: AUDIT_TREE=<checkout> AUDIT_WORK=<scratch> [AUDIT_GITREF=origin/master] python3 gen_db.py   -> $AUDIT_WORK/enchants.json, $AUDIT_WORK/items.json
import json,os
T=os.environ.get('AUDIT_TREE','.'); W=os.environ.get('AUDIT_WORK','/tmp/audit-gear'); os.makedirs(W,exist_ok=True)
R=os.environ.get('AUDIT_GITREF')
if R:
    import subprocess; d=json.loads(subprocess.run(['git','-C',T,'show',R+':assets/database/db.json'],capture_output=True,text=True,check=True).stdout)
else: d=json.load(open(os.path.join(T,'assets/database/db.json')))
sd=lambda a:{str(i):v for i,v in enumerate(a or []) if v}
json.dump({str(e['effectId']):{'name':e.get('name'),'type':e.get('type'),'stats':sd(e.get('stats'))} for e in d['enchants']},open(os.path.join(W,'enchants.json'),'w'))
json.dump({str(i['id']):{'name':i.get('name'),'ilvl':i.get('ilvl'),'type':i.get('type'),'stats':sd(i.get('stats')),'phase':i.get('phase')} for i in d['items']},open(os.path.join(W,'items.json'),'w'))
print('wrote',W+'/enchants.json',W+'/items.json')
