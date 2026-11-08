"""Glancing by mob level over our white swings. Usage: python3 -I glance.py LOG [FROM] [TO]

Mob levels come from swing lines the mob swings (the level field is the source's there) and
from spell lines where the mob is the target (the level field is the target's there). On a
mob's own spell the field is our item level, so those lines don't count."""
import sys, os, collections
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
import cl

evs = cl.load(sys.argv[1])
lo = sys.argv[2] if len(sys.argv) > 2 else '00'
hi = sys.argv[3] if len(sys.argv) > 3 else '99'
lvl = {}
for e in evs:
    try:
        if e.ev == 'SWING_DAMAGE' and e.srcGUID.startswith('Creature'): lvl[e.srcGUID] = int(e.f[-11])
        elif e.ev in ('SPELL_DAMAGE', 'SPELL_PERIODIC_DAMAGE') and e.dstGUID.startswith('Creature'): lvl[e.dstGUID] = int(e.f[-12])
    except (ValueError, IndexError):
        pass
gl = collections.defaultdict(lambda: [0, 0])
names = collections.defaultdict(set)
FIST_FROM, skipped = '21:31:52', 0
for e in evs:
    if not (lo <= e.ts < hi and e.me_src and e.ev in ('SWING_DAMAGE', 'SWING_MISSED')):
        continue
    # From 21:31:52 the user swings a fist weapon on a low Unarmed skill, which glances far more.
    if e.ts >= FIST_FROM:
        skipped += 1
        continue
    L = lvl.get(e.dstGUID)
    gl[L][1] += 1
    names[L].add(e.dst)
    if e.ev == 'SWING_DAMAGE' and cl.dmg(e)['glancing']:
        gl[L][0] += 1
for L in sorted(gl, key=lambda x: (x is None, x)):
    g, n = gl[L]
    print(f'{L}: {g} of {n}  {sorted(names[L])}')
if skipped: print(f'left out {skipped} fist swings from {FIST_FROM}')
