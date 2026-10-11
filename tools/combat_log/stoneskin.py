"""How much our Stoneskin Totem takes off mob swings on us (audit 7.29). Usage: python3 -I stoneskin.py LOG [FROM] [TO]

From the audit session's script (2026-10-11). Stoneskin is up from our summon until our next
earth totem, the totem's death or 5 min, whichever comes first. Totem buffs don't show as auras
in the log, and the window doesn't know when we stand out of the totem's range, so some hits
inside it fit nothing.

We take the mobs' normal swings on us (no crit, crushing, block, absorb, glance or overkill)
and find our armor from the ones outside a Stoneskin window: the armor that fits the most of
them with Classic's armor formula on the mob's level. Then we count how many hits inside a
window fit each reduction, taken off the raw before armor (Forever, the sim since 91217b13) or
off the amount after armor (the old sim). Rank 3 takes 11 off at level 30.
"""
import sys, os
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
import cl

evs = cl.load(sys.argv[1]); evs.sort(key=lambda e: e.t)
lo = sys.argv[2] if len(sys.argv) > 2 else '00'
hi = sys.argv[3] if len(sys.argv) > 3 else '99'
EARTH = ('Stoneskin Totem', 'Tremor Totem', 'Strength of Earth Totem', 'Earthbind Totem', 'Stoneclaw Totem')

# Our earth totems in order, each with when it ended.
earth = []
gone = {}
for e in evs:
    if e.ev == 'SPELL_SUMMON' and e.src.startswith(cl.ME) and e.spell and e.spell.startswith(EARTH):
        earth.append([e.t, e.spell, e.dstGUID, e.t + 300])
    if e.ev in ('UNIT_DIED', 'UNIT_DESTROYED') and e.dstGUID:
        gone.setdefault(e.dstGUID, e.t)
for i, w in enumerate(earth):
    if i + 1 < len(earth):
        w[3] = min(w[3], earth[i + 1][0])
    if w[2] in gone:
        w[3] = min(w[3], gone[w[2]])
ss = [(w[0], w[3]) for w in earth if w[1] == 'Stoneskin Totem']
print('Stoneskin windows (sec)', [round(b - a) for a, b in ss])

def stoneskin(t):
    return any(a <= t <= b for a, b in ss)

hits = []
for e in evs:
    if e.ev != 'SWING_DAMAGE' or not e.dst.startswith(cl.ME) or not e.srcGUID.startswith('Creature') or not lo <= e.ts < hi:
        continue
    d = cl.dmg(e)
    if d['crit'] or d['crushing'] or d['blocked'] or d['absorbed'] or d['glancing']:
        continue
    if d['overkill'] and d['overkill'] > 0:
        continue
    L = cl.level_of(e)
    if L is None or not d['raw']:
        continue
    hits.append((L, d['raw'], d['amount'], stoneskin(e.t)))
print(len(hits), 'normal hits on us,', sum(h[3] for h in hits), 'with Stoneskin up')

def dr(A, L):
    return A / (A + 400 + 85 * L)

def taken(h, A, red, before):
    L, raw = h[0], h[1]
    return (raw - red) * (1 - dr(A, L)) if before else raw * (1 - dr(A, L)) - red

def fits(h, A, red, before):
    return abs(taken(h, A, red, before) - h[2]) < 1

plain = [h for h in hits if not h[3]]
best = sorted(((sum(fits(h, A, 0, True) for h in plain), A) for A in range(300, 1800)), reverse=True)[:5]
print('armor that fits the most hits without Stoneskin (fits, armor):', best, 'of', len(plain))
A = best[0][1]
up = [h for h in hits if h[3]]
for red in (0, 4, 7, 11, 13, 30):
    print(f'take {red} off: fits before armor {sum(fits(h, A, red, True) for h in up)}, '
          f'after armor {sum(fits(h, A, red, False) for h in up)}, of {len(up)}')
print('total taken with Stoneskin up', sum(h[2] for h in up), 'model with nothing off',
      round(sum(taken(h, A, 0, True) for h in up)), '11 before armor', round(sum(taken(h, A, 11, True) for h in up)),
      '11 after armor', round(sum(taken(h, A, 11, False) for h in up)))
