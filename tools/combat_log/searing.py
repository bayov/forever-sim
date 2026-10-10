"""Searing Totem bolt timing with totem to mob distance. Usage: python3 -I searing.py LOG [HH:MM:SS start]"""
import sys, math, bisect, statistics as st
import os; sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
import cl

SPEED = 19.0  # yards a second, SpellMisc Speed of 6351

def base(e):
    if e.ev.startswith('SWING_'):
        return 9
    if e.ev.startswith(('SPELL_', 'RANGE_')):
        return 12
    return None

evs = cl.load(sys.argv[1])
evs.sort(key=lambda e: e.t)
since = sys.argv[2] if len(sys.argv) > 2 else '00:00:00'

# Position track of every unit, from the lines where it is the info unit.
track = {}
for e in evs:
    b = base(e)
    if b is None or len(e.f) <= b + 15:
        continue
    try:
        x, y = float(e.f[b + 14]), float(e.f[b + 15])
    except ValueError:
        continue
    track.setdefault(e.f[b], []).append((e.t, x, y))

def pos(guid, t):
    tr = track.get(guid, [])
    i = bisect.bisect_right([p[0] for p in tr], t + 0.01) - 1
    return tr[i] if i >= 0 else None

def moved(guid, t1, t2):
    """Largest distance from the position at t1 over the unit's lines between t1 and t2."""
    p0 = pos(guid, t1)
    if not p0:
        return float('nan')
    return max([0.0] + [math.hypot(x - p0[1], y - p0[2]) for (t, x, y) in track.get(guid, []) if t1 < t <= t2])

summons = [e for e in evs if e.ev == 'SPELL_SUMMON' and e.src.startswith(cl.ME) and 'Searing' in e.dst and e.ts >= since]
rows = []
for sm in summons:
    g = sm.dstGUID
    te = [e for e in evs if e.srcGUID == g]
    starts = [e.t for e in te if e.ev == 'SPELL_CAST_START']
    succ = [e for e in te if e.ev == 'SPELL_CAST_SUCCESS']
    lands = [e for e in te if e.ev in ('SPELL_DAMAGE', 'SPELL_MISSED')]
    print(f'\nTotem at {sm.ts}: {len(succ)} bolts, first cast start {starts[0]-sm.t:.2f} s, last success {succ[-1].t-sm.t:.2f} s' if succ else f'\nTotem at {sm.ts}: no bolts')
    for i, s in enumerate(succ):
        tx, ty = float(s.f[26]), float(s.f[27])
        mp = pos(s.dstGUID, s.t)
        dist = math.hypot(mp[1] - tx, mp[2] - ty) if mp else float('nan')
        hit = next((h for h in lands if h.t >= s.t - 0.001 and h.ev == 'SPELL_DAMAGE' and h.dstGUID == s.dstGUID), None)
        flight = hit.t - s.t if hit and hit.t - s.t < 2 else float('nan')
        nxt = succ[i + 1] if i + 1 < len(succ) else None
        nstart = next((x for x in starts if x > s.t), None)
        gap = nxt.t - s.t if nxt else float('nan')
        wait = nstart - s.t if nstart and nxt else float('nan')
        mv = moved(s.dstGUID, s.t, nxt.t) if nxt else float('nan')
        same = nxt is not None and nxt.dstGUID == s.dstGUID
        rows.append(dict(dist=dist, flight=flight, gap=gap, wait=wait, moved=mv, same=same))
        print(f'  {s.t-sm.t:6.2f}  {s.dst[:20]:20} dist {dist:5.1f}  flight {flight:5.2f} (dist/19 {dist/SPEED:4.2f})'
              f'  gap {gap:5.2f}  wait {wait:5.2f}  mob moved {mv:4.1f}{"" if same or not nxt else "  NEW MOB"}')

# Totals over clean gaps: same mob, the mob moved under 0.5 yards, gap under 3 sec.
clean = [r for r in rows if r['same'] and r['moved'] < 0.5 and r['gap'] < 3.0]
def summary(name, rs):
    if not rs: return
    stalls = [r for r in rs if r['wait'] > 0.45]
    normal = [r for r in rs if r['wait'] <= 0.45]
    print(f'{name}: {len(rs)} gaps, mean {st.mean(r["gap"] for r in rs):.3f}, normal gap median {st.median(r["gap"] for r in normal):.3f}'
          f' (wait {st.median(r["wait"] for r in normal):.3f}), stalls {len(stalls)} (wait median {st.median(r["wait"] for r in stalls) if stalls else float("nan"):.3f})')
print()
summary('all', clean)
summary('totem under 6 yards', [r for r in clean if r['dist'] < 6])
summary('totem 6 to 12 yards', [r for r in clean if 6 <= r['dist'] < 12])
summary('totem 12 yards or more', [r for r in clean if r['dist'] >= 12])
fl = [(r['dist'], r['flight']) for r in rows if r['flight'] == r['flight'] and r['dist'] == r['dist'] and r['moved'] < 0.5]
print('flight / (dist/19), median', round(st.median(f / (dd / SPEED) for dd, f in fl if dd > 3), 3), 'over', sum(dd > 3 for dd, f in fl))
