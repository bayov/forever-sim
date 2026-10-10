"""Our damage on one mob name, by spell: hits, partial resists, full resists. Usage: python3 -I serpent.py LOG NAME"""
import sys, os, collections
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
import cl
evs = cl.load(sys.argv[1]); name = sys.argv[2]
c = collections.defaultdict(lambda: [0, 0, 0, []])
for e in evs:
    if not e.src.startswith(cl.ME) or e.dst != name: continue
    if e.ev in ('SPELL_DAMAGE', 'SPELL_PERIODIC_DAMAGE', 'SPELL_MISSED', 'SWING_DAMAGE', 'SWING_MISSED'):
        k = (e.ev.split('_')[0], e.f[10] if e.ev.startswith('SPELL') else 'melee')
        if e.ev.endswith('MISSED'):
            if cl.miss(e) == 'RESIST': c[k][2] += 1
            continue
        c[k][0] += 1
        # SPELL_DAMAGE tail: amount, raw, overkill, school, resisted, blocked, absorbed
        res = int(e.f[-7]) if e.ev.startswith('SPELL') else 0
        if res > 0: c[k][1] += 1; c[k][3].append((e.ts, "lvl", e.f[-12], "raw", e.f[-10], "resisted", e.f[-7]))
for k, v in sorted(c.items()): print(k, 'hits', v[0], 'partial', v[1], 'full resist', v[2], v[3])
