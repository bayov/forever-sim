# Usage: python3 regen.py > drift_report.txt
# Drift check: rebuilds every tools/dps_report/req/<id>.json in AUDIT_TREE with mksettings from its meta (spec mapped from the
# id, gearSet, apl, talents, race, 180 s), then prints the field-by-field differences against the committed request, ignoring
# simOptions and duration. Compare the output with orchestration/audit/ref/results/drift_report.txt.
# Env: AUDIT_TREE (default: this repo), AUDIT_WORK (default ${TMPDIR:-/tmp}/audit-results; regenerated JSON goes to
#      $AUDIT_WORK/drift/), AUDIT_BIN (dir holding mksettings; default $AUDIT_WORK/bin, built by build.sh).
import json,glob,os,re,subprocess,sys
D=os.path.dirname(os.path.abspath(__file__))
TREE=os.environ.get('AUDIT_TREE') or os.path.abspath(os.path.join(D,'..','..','..'))
WORK=os.environ.get('AUDIT_WORK') or os.path.join(os.environ.get('TMPDIR','/tmp'),'audit-results')
BIN=os.environ.get('AUDIT_BIN') or os.path.join(WORK,'bin')
os.makedirs(os.path.join(WORK,'drift'),exist_ok=True)
os.chdir(TREE)
def spec_for(m):
    i=m['id']
    c,rest=i.split('__')
    if c=='druid': return 'feral' if 'feral' in rest else 'balance'
    if c=='hunter':
        base={'beast-mastery':'hunter_bm','lone-wolf':'hunter_lw','marksmanship':'hunter_mm','survival':'hunter_sv'}[rest.split('-duty')[0]]
        return base
    if c=='mage': return 'mage_'+rest
    if c=='paladin': return 'ret'
    if c=='priest': return 'shadow'
    if c=='rogue':
        if rest.startswith('mutilate'): return 'rogue_assassination'
        if rest.startswith('subtlety'): return 'rogue_subtlety'
        return 'rogue_combat'
    if c=='shaman': return 'ele' if 'elemental' in rest else 'enh'
    if c=='warlock': return rest.split('-duty')[0]
    if c=='warrior': return 'arms' if rest.startswith('arms') else 'fury'
def flat(x,p=''):
    out={}
    if isinstance(x,dict):
        for k,v in x.items(): out.update(flat(v,p+'.'+k))
    elif isinstance(x,list):
        for j,v in enumerate(x): out.update(flat(v,p+'[%d]'%j))
    else: out[p]=x
    return out
IGN=re.compile(r'(simOptions|duration|durationVariation)')
for f in sorted(glob.glob('tools/dps_report/req/*.meta.json')):
    m=json.load(open(f)); i=m['id']; sp=spec_for(m)
    out=os.path.join(WORK,'drift',i+'.regen.json')
    r=subprocess.run([os.path.join(BIN,'mksettings'),'-spec',sp,'-gear',m['gearSet'],'-apl',m['apl'],'-talents',m['talents'],'-race',m['race'],'-duration','180','-out',out],capture_output=True,text=True)
    if r.returncode: print(f'## {i} spec={sp} MKSETTINGS FAIL: {r.stderr.strip()[:200]}'); continue
    a=flat(json.load(open(f.replace('.meta.json','.json')))); b=flat(json.load(open(out)))
    diffs=[]
    for k in sorted(set(a)|set(b)):
        if IGN.search(k): continue
        va,vb=a.get(k,'<absent>'),b.get(k,'<absent>')
        if va!=vb: diffs.append((k,va,vb))
    # collapse APL diffs
    apl=[d for d in diffs if '.rotation' in d[0] or 'priorityList' in d[0] or 'prepullActions' in d[0]]
    other=[d for d in diffs if d not in apl]
    print(f'## {i} spec={sp} diffs={len(diffs)} (apl/prepull {len(apl)})')
    for k,va,vb in other[:25]: print(f'   {k}: committed={str(va)[:60]} regen={str(vb)[:60]}')
    if apl: print(f'   [apl/prepull] e.g. {apl[0][0]}: {str(apl[0][1])[:50]} -> {str(apl[0][2])[:50]}')
