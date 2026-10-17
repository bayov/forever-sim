# Usage: python3 tal.py <talents json> <talent string>
# Prints the talents a talent string picks.
import json,sys
t=json.load(open(sys.argv[1]))
s=sys.argv[2].split('-')
for ti,tree in enumerate(t):
    tal=tree['talents']
    st=s[ti] if ti<len(s) else ''
    for i,c in enumerate(st):
        if c!='0': print(tree['name'],tal[i]['fieldName'],c,'/',tal[i]['maxPoints'],tal[i].get('spellIds',[])[:1])
