# Usage: python3 compare.py summary <summary.txt> [<ref summary.txt>]
#        python3 compare.py builds <ui/dps_report/data/builds.json> <summary.txt>
# summary: flags every request whose 20k run is more than TOL (0.15) off its published DPS, or failed, and (with a reference)
#          lists every build whose own DPS moved since the reference run.
# builds:  checks the dashboard's published builds.json against a summary (for a pass where only ui/dps_report changed).
# Summary lines are runreq.sh output: "<id> mine <dps> pub <dps> diff <d>". Exit status 1 if anything is flagged.
import json,sys
TOL=0.15
def load(p):
    r={}
    for line in open(p):
        f=line.split()
        if len(f)>=2 and f[1]=='SIMFAIL': r[f[0]]=None
        elif len(f)>=5 and f[1]=='mine': r[f[0]]=(float(f[2]),None if f[4]=='None' else float(f[4]))
    return r
bad=0
if sys.argv[1]=='summary':
    s=load(sys.argv[2])
    for i,v in sorted(s.items()):
        if v is None: print('SIMFAIL',i); bad=1
        elif v[1] is None: print('NO-PUBLISHED',i,v[0]); bad=1
        elif abs(v[0]-v[1])>TOL: print('OFF %-44s mine %7.1f pub %7.1f diff %+.1f'%(i,v[0],v[1],v[0]-v[1])); bad=1
    print('%d requests, %d within %.2f of published'%(len(s),sum(1 for v in s.values() if v and v[1] is not None and abs(v[0]-v[1])<=TOL),TOL))
    if len(sys.argv)>3:
        ref=load(sys.argv[3])
        for i in sorted(set(s)|set(ref)):
            if i not in ref or i not in s: print('ADDED' if i not in ref else 'REMOVED',i); continue
            a,b=ref[i],s[i]
            if a is None or b is None: continue
            if abs(a[0]-b[0])>TOL: print('MOVED %-44s %7.1f -> %7.1f (%+.1f)'%(i,a[0],b[0],b[0]-a[0]))
elif sys.argv[1]=='builds':
    b={r['id']:r['dps'] for r in json.load(open(sys.argv[2]))}; s=load(sys.argv[3])
    for i in sorted(set(b)|set(s)):
        if i not in s or s[i] is None: print('NOT-IN-SUMMARY',i); bad=1; continue
        if i not in b: print('NOT-IN-BUILDS',i); bad=1; continue
        if abs(b[i]-s[i][0])>TOL: print('OFF %-44s builds %7.1f summary %7.1f'%(i,b[i],s[i][0])); bad=1
    print('%d builds checked'%len(b))
else: raise SystemExit('usage: compare.py summary <summary.txt> [<ref summary.txt>] | compare.py builds <builds.json> <summary.txt>')
sys.exit(bad)
