"""Spell crits by spell group and mob level. Usage: python3 -I crit.py LOG [FROM] [TO]"""
import sys, os, collections
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
import cl

evs = cl.load(sys.argv[1]); evs.sort(key=lambda e: e.t)
lo = sys.argv[2] if len(sys.argv) > 2 else '00'
hi = sys.argv[3] if len(sys.argv) > 3 else '99'

lvl = {}
for e in evs:
    if e.srcGUID.startswith('Creature') and e.ev in ('SWING_DAMAGE', 'SPELL_DAMAGE', 'RANGE_DAMAGE'):
        try:
            v = int(e.f[-11] if e.ev == 'SWING_DAMAGE' else e.f[-12])
            if 1 <= v <= 63: lvl[e.srcGUID] = v
        except ValueError: pass

totems = {e.dstGUID: e.dst for e in evs if e.ev == 'SPELL_SUMMON' and e.src.startswith(cl.ME)}
cnt = collections.defaultdict(lambda: [0, 0])
for e in evs:
    if not (lo <= e.ts < hi) or e.ev != 'SPELL_DAMAGE': continue
    if not (e.src.startswith(cl.ME) or e.srcGUID in totems): continue
    d = cl.dmg(e)
    if d['school'] == 1: continue  # Windfury Weapon attacks, Stormstrike
    if e.spellId == 10444: g = 'FT Weapon'
    elif e.spellId == 16368: g = 'FT Totem'
    elif e.spell == 'Lightning Shield': g = 'LS orb'
    elif e.srcGUID in totems and 'Searing' in totems[e.srcGUID]: g = 'other'
    elif e.spell in ('Earth Shock', 'Flame Shock', 'Frost Shock', 'Fire Nova', 'Lightning Bolt'): g = 'other'
    else: g = 'unknown ' + str(e.spell)
    L = lvl.get(e.dstGUID)
    cnt[(g, L)][1] += 1
    cnt[(g, L)][0] += d['crit']

groups = sorted({g for g, _ in cnt})
for g in groups:
    rows = sorted(((L, c) for (gg, L), c in cnt.items() if gg == g), key=lambda x: (x[0] is None, x[0] or 0))
    tc = sum(c[0] for _, c in rows); tn = sum(c[1] for _, c in rows)
    print(f'{g}: {tc}/{tn} ({100*tc/max(tn,1):.1f}%)  ' + '  '.join(f'L{L}: {c[0]}/{c[1]}' for L, c in rows))
    le30 = [c for L, c in rows if L is not None and L <= 30]
    a = sum(c[0] for c in le30); b = sum(c[1] for c in le30)
    print(f'   level 30 or lower: {a}/{b} ({100*a/max(b,1):.1f}%)')
