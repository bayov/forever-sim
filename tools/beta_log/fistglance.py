"""Our white swings by weapon (raw under 150 is the fist) and mob level, with glancing and misses. Usage: python3 -I fistglance.py LOG FROM [TO]"""
import sys, os, collections
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
import cl
evs = cl.load(sys.argv[1]); evs.sort(key=lambda e: e.t)
lo = sys.argv[2]; hi = sys.argv[3] if len(sys.argv) > 3 else '99'
lvl = {}
for e in evs:
    if e.srcGUID.startswith('Creature') and e.ev in ('SWING_DAMAGE', 'SPELL_DAMAGE', 'SPELL_PERIODIC_DAMAGE'):
        try: lvl[e.srcGUID] = int(e.f[-11] if e.ev == 'SWING_DAMAGE' else e.f[-12])
        except ValueError: pass
# Weapon by time: the first fist swing (raw under 150) marks the switch.
me = [e for e in evs if e.src.startswith(cl.ME) and lo <= e.ts < hi and e.ev in ('SWING_DAMAGE', 'SWING_MISSED')]
first = next((e.t for e in me if e.ev == 'SWING_DAMAGE' and cl.dmg(e)['raw'] < 150), None)
c = collections.defaultdict(collections.Counter)
for e in me:
    w = 'fist' if first is not None and e.t >= first else 'mace'
    L = lvl.get(e.dstGUID)
    if e.ev == 'SWING_DAMAGE':
        d = cl.dmg(e)
        c[(w, L)]['landed'] += 1
        if d.get('glancing'): c[(w, L)]['glance'] += 1
        if d.get('crit'): c[(w, L)]['crit'] += 1
    else: c[(w, L)][cl.miss(e)] += 1
print('first fist swing', [e.ts for e in me if first is not None and e.t == first][:1])
for k in sorted(c, key=str): print(k, dict(c[k]))
