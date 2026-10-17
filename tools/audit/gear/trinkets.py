# On-use trinket check (standards 1b): every core.NewSimpleStatOffensiveTrinketEffect in sim/*/forever_synthetic_trinkets.go,
# averaged over a 180 s fight (uses at 0 s and every cooldown): amount x weight x seconds active / 180.
# Target 27 for spell power, 37 for AP / Str / Agi / RAP (AP and RAP on one effect count once). Flags beyond +/-3%.
# Usage: AUDIT_TREE=<checkout> python3 trinkets.py [git ref, default: working tree]
import re,os,sys,subprocess,glob
T=os.environ.get('AUDIT_TREE','.'); ref=sys.argv[1] if len(sys.argv)>1 else None
W={'SpellPower':.86,'AttackPower':.5,'RangedAttackPower':.4,'Strength':1,'Agility':1}
def files():
    if ref:
        out=subprocess.run(['git','-C',T,'ls-tree','-r','--name-only',ref,'sim'],capture_output=True,text=True).stdout.split()
        for p in out:
            if p.endswith('forever_synthetic_trinkets.go'): yield p,subprocess.run(['git','-C',T,'show',f'{ref}:{p}'],capture_output=True,text=True).stdout
    else:
        for p in glob.glob(os.path.join(T,'sim/*/forever_synthetic_trinkets.go')): yield p,open(p).read()
for p,src in files():
    for m in re.finditer(r'NewSimpleStatOffensiveTrinketEffect\((\w+),\s*stats\.Stats\{([^}]*)\},\s*time\.(Second|Minute)\*(\d+),\s*time\.(Second|Minute)\*(\d+)\)',src):
        st={k:float(v) for k,v in re.findall(r'stats\.(\w+):\s*([\d.]+)',m.group(2))}
        dur=int(m.group(4))*(60 if m.group(3)=='Minute' else 1); cd=int(m.group(6))*(60 if m.group(5)=='Minute' else 1)
        active=sum(min(dur,180-t) for t in range(0,180,cd))
        if 'AttackPower' in st and 'RangedAttackPower' in st: st.pop('RangedAttackPower')
        val=sum(v*W.get(k,0) for k,v in st.items())*active/180
        tgt=27 if 'SpellPower' in st else 37
        print(f"{p.split('/')[1]:8} {m.group(1):40} {st} {dur}s/{cd}s avg {val:.1f} vs {tgt}{'' if abs(val/tgt-1)<=.03 else '  !'}")
