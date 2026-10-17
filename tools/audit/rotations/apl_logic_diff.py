#!/usr/bin/env python3
# usage: apl_logic_diff.py OLD_REV NEW_REV [--repo .]
# Lists APL files under ui/**/apls whose logic changed between two revisions, ignoring "notes"
# (workers often reword notes only). Skips leveling/level20/25/30 rotations. Prints changed lines.
import json,subprocess,argparse,difflib
ap=argparse.ArgumentParser(); ap.add_argument('old'); ap.add_argument('new'); ap.add_argument('--repo',default='.'); a=ap.parse_args()
def sh(*x): return subprocess.run(['git','-C',a.repo,*x],capture_output=True,text=True).stdout
files=[f for f in sh('diff','--name-only',a.old,a.new,'--',':(glob)ui/**/apls/**').split() if not any(k in f for k in ('leveling','level20','level25','level30'))]
def lines(t):
    if not t: return []
    j=json.loads(t); out=[]
    for k in ('prepullActions','priorityList'):
        for e in j.get(k,[]): out.append(k[:3]+' '+json.dumps({kk:v for kk,v in e.items() if kk!='notes'},sort_keys=True))
    return out
for f in files:
    x=lines(sh('show',f'{a.old}:{f}')); y=lines(sh('show',f'{a.new}:{f}'))
    d=[l for l in difflib.unified_diff(x,y,lineterm='',n=0) if not l.startswith(('---','+++','@@'))]
    if d:
        print('=====',f)
        for l in d: print('  ',l[:300])
print(f'# {len(files)} apl files touched')
