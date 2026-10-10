"""Run the beta_handoff.md recheck list over a time window. Usage: python3 -I recheck.py LOG HH:MM:SS [HH:MM:SS]"""
import sys, os, collections
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
import cl

evs = cl.load(sys.argv[1]); evs.sort(key=lambda e: e.t)
# File order breaks ties: a proc logged before a swing in the same ms didn't come from that swing.
pos = {id(e): i for i, e in enumerate(evs)}
lo = sys.argv[2]; hi = sys.argv[3] if len(sys.argv) > 3 else '99'
win = [e for e in evs if lo <= e.ts < hi]
me = [e for e in win if e.src.startswith(cl.ME)]
print(f'window {lo} to {win[-1].ts if win else "-"}, {len(win)} lines')

# Mob levels from the mob's own swings (info unit = source, level at row[-11]).
lvl = {}
for e in evs:
    if e.ev == 'SWING_DAMAGE' and e.srcGUID.startswith('Creature'):
        try: lvl[e.srcGUID] = int(e.f[-11])
        except ValueError: pass

swings = [e for e in me if e.ev in ('SWING_DAMAGE', 'SWING_MISSED')]
landed = [e for e in me if e.ev == 'SWING_DAMAGE']
ss = [e for e in me if e.spell == 'Stormstrike' and e.ev in ('SPELL_DAMAGE', 'SPELL_MISSED')]
print(f'swings {len(swings)} (landed {len(landed)}), Stormstrike {len(ss)}')

# White raw range and Stormstrike raw vs the weapon models. Rage of the Storm (3.3, 2H) at sheet
# AP 389 up to the respec and 392 after it. Bloody Brass Knuckles (1.6, 24-46) at 397 AP from
# the 21:30 export, so a raw under 150 is the fist.
wr = sorted(cl.dmg(e)['raw'] for e in landed)
big = [r for r in wr if r >= 150]; fist = [r for r in wr if r < 150]
if big: print(f'white raw, Rage of the Storm {big[0]} to {big[-1]} (model at 389 AP: 191.7 to 238.7, at 392: 192.4 to 239.4)')
if fist: print(f'white raw, Bloody Brass Knuckles {fist[0]} to {fist[-1]} (model at 397 AP: 69.4 to 91.4)')
ssr = [cl.dmg(e)['raw'] for e in ss if e.ev == 'SPELL_DAMAGE']
bad = [r for r in ssr if not 220 <= r <= 272]
print(f'Stormstrike raw {sorted(ssr)}  outside 220.0-271.7: {bad}')

# Lightning Shield orbs.
orb = [e for e in me if e.spell == 'Lightning Shield' and e.ev in ('SPELL_DAMAGE', 'SPELL_MISSED')]
print(f'orbs {len(orb)}: misses {sum(e.ev == "SPELL_MISSED" for e in orb)}, crits {sum(e.ev == "SPELL_DAMAGE" and cl.dmg(e)["crit"] for e in orb)}')

# Flametongue Weapon (10444) and Totem (16368) procs, each to the closest preceding trigger within 200 ms.
def procs(spellid):
    return [e for e in me if e.spellId == spellid and e.ev in ('SPELL_DAMAGE', 'SPELL_MISSED')]
for name, sid, trig in (('Flametongue Weapon', 10444, swings + ss), ('Flametongue Totem', 16368, swings + ss)):
    ps = procs(sid)
    if not ps: continue
    trig = sorted(trig, key=lambda e: pos[id(e)])
    got = collections.Counter()
    for p in ps:
        prev = [t for t in trig if 0 <= p.t - t.t <= 0.2 and pos[id(t)] < pos[id(p)]]
        if prev: got[id(prev[-1])] += 1
        else: print(f'  {name} proc at {p.ts} with no trigger')
    kinds = collections.Counter()
    for t in trig:
        kind = ('swing ' if t.ev.startswith('SWING') else 'SS ') + (('landed' if t.ev.endswith('DAMAGE') else cl.miss(t)))
        kinds[(kind, got[id(t)] > 0)] += 1
    print(f'{name}: {len(ps)} procs', dict(sorted(kinds.items())))

# Searing Totem bolts: damage, crit, miss.
tot = {e.dstGUID for e in win if e.ev == 'SPELL_SUMMON' and e.src.startswith(cl.ME) and 'Searing' in e.dst}
bolts = [e for e in win if e.srcGUID in tot and e.ev in ('SPELL_DAMAGE', 'SPELL_MISSED')]
br = [cl.dmg(e) for e in bolts if e.ev == 'SPELL_DAMAGE']
print(f'Searing bolts {len(bolts)}: hits {len(br)}, misses {len(bolts)-len(br)}, crits {sum(d["crit"] for d in br)}, '
      f'normal raw {min((d["raw"] for d in br if not d["crit"]), default=None)} to {max((d["raw"] for d in br if not d["crit"]), default=None)}')

# Glancing by mob level, over all our white swings.
gl = collections.defaultdict(lambda: [0, 0])
for e in swings:
    L = lvl.get(e.dstGUID)
    gl[L][1] += 1
    if e.ev == 'SWING_DAMAGE' and cl.dmg(e)['glancing']: gl[L][0] += 1
print('glancing by level', dict(gl))

# Partial resists on our spells that can partially resist, by mob level.
pr = collections.defaultdict(lambda: [0, 0])
for e in [x for x in win if (x.src.startswith(cl.ME) or x.srcGUID in tot) and x.ev in ('SPELL_DAMAGE', 'SPELL_PERIODIC_DAMAGE')]:
    if e.spell in ('Stormstrike',): continue
    d = cl.dmg(e)
    if d['school'] == 1: continue
    L = lvl.get(e.dstGUID)
    pr[(e.spell, L)][1] += 1
    if d['resisted']: pr[(e.spell, L)][0] += 1
print('partial resists (resisted, hits) by spell and level', dict(pr))

# Fire Nova: damage spell id and normal raw range (rank 2 at 91 spell power: 122.41 to 136.53).
fn = [e for e in me if e.spell == 'Fire Nova' and e.ev == 'SPELL_DAMAGE']
if fn:
    nr = [cl.dmg(e)['raw'] for e in fn if not cl.dmg(e)['crit']]
    print(f'Fire Nova hits {len(fn)}: spell ids {sorted({e.spellId for e in fn})}, normal raw {min(nr, default=None)} to {max(nr, default=None)}')

# Stormstrike: the Improved Stormstrike regen buff (1238931) comes on every cast, landed or
# not. The Stormstrike mark (17364 aura) comes only on a landed hit (blocks too) that doesn't kill.
res = collections.Counter()
for s in ss:
    kind = 'landed' if s.ev == 'SPELL_DAMAGE' else cl.miss(s)
    if kind == 'landed' and cl.dmg(s)['overkill'] >= 0: kind += ' kill'
    near = [e for e in me if -0.2 <= e.t - s.t <= 0.05 and e.spellId in (1238931, 17364) and e.ev in ('SPELL_AURA_APPLIED', 'SPELL_AURA_REFRESH')]
    res[(kind, 'buff' if any(e.spellId == 1238931 for e in near) else 'NO buff', 'mark' if any(e.spellId == 17364 for e in near) else 'no mark')] += 1
print('Stormstrike buff and mark:', dict(res))

# Healing Stream: crits of our own totems' heals (the suffix ends amount, raw, overheal, absorbed, crit).
hst = {e.dstGUID for e in evs if e.ev == 'SPELL_SUMMON' and e.src.startswith(cl.ME) and 'Healing Stream' in e.dst}
hs = [e for e in win if e.srcGUID in hst and e.ev == 'SPELL_PERIODIC_HEAL']
if hs: print(f'Healing Stream heals {len(hs)}: crits {sum(e.f[-1] == "1" for e in hs)}')
