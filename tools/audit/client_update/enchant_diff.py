"""usage: python3 -I enchant_diff.py <build> <build> [<build>...]
SpellItemEnchantment between each consecutive pair of the builds given: added, removed and changed
enchants with the changed fields. The sim's Forever enchants live in tools/database/forever_enchants.go."""
import csv, os, sys
csv.field_size_limit(10**9)
sys.path.insert(0, os.path.join(os.path.dirname(os.path.abspath(__file__)), '../../sources'))
import wago  # noqa: E402
B = sys.argv[1:]
def load(b): return wago.by_id('SpellItemEnchantment', b)
prev = load(B[0])
for b in B[1:]:
    cur = load(b); print('==', b)
    for k in sorted(set(prev) | set(cur), key=int):
        o, n = prev.get(k), cur.get(k)
        if o == n: continue
        if not o: print(' ADDED', k, n['Name_lang'], {x: v for x, v in n.items() if v not in ('0', '')}); continue
        if not n: print(' REMOVED', k, o['Name_lang']); continue
        print(' ', k, n['Name_lang'], {x: (o.get(x), n.get(x)) for x in set(n) | set(o) if o.get(x) != n.get(x) and x != 'VerifiedBuild'})
    prev = cur
