"""Flametongue Totem procs per landed white swing, with the totem's state. Usage: python3 -I ftt.py LOG FROM [TO]"""
import sys, os, math, collections
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
import cl
evs = cl.load(sys.argv[1]); evs.sort(key=lambda e: e.t)
lo = sys.argv[2]; hi = sys.argv[3] if len(sys.argv) > 3 else '99'
pos = {id(e): i for i, e in enumerate(evs)}
def xy(e):
    try:
        if e.ev == 'SWING_DAMAGE': return float(e.f[-15]), float(e.f[-14])
        if e.ev == 'SPELL_CAST_SUCCESS': return float(e.f[-5]), float(e.f[-4])
    except (ValueError, IndexError): return None
died = collections.defaultdict(list)
for e in evs:
    if e.ev in ('UNIT_DIED', 'PARTY_KILL'): died[e.dstGUID].append(e.t)
ft = [e for e in evs if e.src.startswith(cl.ME) and e.spellId == 16368]
fire = None  # (name, xy, guid)
totem_guid = {}
res = collections.Counter()
for e in evs:
    if not (lo <= e.ts < hi): continue
    if e.src.startswith(cl.ME) and e.ev == 'SPELL_CAST_SUCCESS':
        if e.spell == 'Flametongue Totem': fire = ('FT', xy(e))
        elif e.spell and e.spell.startswith('Searing Totem'): fire = ('Searing', None)
        elif e.spell == 'Totemic Recall': fire = None
    if e.ev == 'SPELL_SUMMON' and e.src.startswith(cl.ME) and 'Flametongue' in e.dst: totem_guid[e.dstGUID] = 1
    if e.ev in ('UNIT_DIED', 'UNIT_DESTROYED') and e.dstGUID in totem_guid and fire and fire[0] == 'FT': fire = None
    if e.src.startswith(cl.ME) and e.ev == 'SWING_DAMAGE':
        d = cl.dmg(e)
        got = any(0 <= p.t - e.t <= 0.2 and pos[id(p)] > pos[id(e)] for p in ft)
        if d['overkill'] >= 0: why = 'kill'
        elif any(0 <= t - e.t <= 0.13 for t in died[e.dstGUID]): why = 'mob died within 130 ms'
        elif fire is None: why = 'no fire totem'
        elif fire[0] != 'FT': why = 'Searing down'
        else:
            dist = math.dist(xy(e), fire[1]) if xy(e) and fire[1] else -1
            why = 'in range' if dist <= 30 else 'over 30 yd'
        res[(why, 'proc' if got else 'no proc')] += 1
        if why == 'in range' and not got: print('  unexplained', e.ts, e.dst, f'{dist:.1f} yd')
for k, v in sorted(res.items()): print(k, v)
