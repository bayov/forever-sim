"""Which of our crits give Elemental Devastation (beta_handoff.md item 5). Usage: python3 -I devastation.py LOG [FROM] [TO]

For each of our crits we look for the buff (30165) being applied or refreshed on us within
WINDOW seconds after it. A crit that lands next to another of our crits shares the buff
event, so we count it apart as "shared". The last part lists each buff event and the crits
before it, which shows a buff that came from something we don't count as a trigger.
"""
import sys, os, collections
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
import cl

ED = 30165
WINDOW = 0.2

evs = cl.load(sys.argv[1]); evs.sort(key=lambda e: e.t)
lo = sys.argv[2] if len(sys.argv) > 2 else '00'
hi = sys.argv[3] if len(sys.argv) > 3 else '99'
win = [e for e in evs if lo <= e.ts < hi]
print(f'window {lo} to {win[-1].ts if win else "-"}')

totems = {e.dstGUID: e.dst for e in evs if e.ev == 'SPELL_SUMMON' and e.src.startswith(cl.ME)}

def kind(e):
    if e.ev.startswith('SWING_DAMAGE'):
        return 'white swing'
    if e.srcGUID in totems:
        return totems[e.srcGUID]
    if e.spellId == 10444:
        return 'Flametongue Weapon'
    if e.spellId == 16368:
        return 'Flametongue Totem'
    if e.ev == 'SPELL_PERIODIC_DAMAGE':
        return e.spell + ' tick'
    return e.spell + (' (physical)' if cl.dmg(e)['school'] == 1 else '')

crits = [e for e in win if (e.src.startswith(cl.ME) or e.srcGUID in totems)
         and e.ev in ('SWING_DAMAGE', 'SPELL_DAMAGE', 'SPELL_PERIODIC_DAMAGE') and cl.dmg(e)['crit']]
hits = collections.Counter(kind(e) for e in win if (e.src.startswith(cl.ME) or e.srcGUID in totems)
                           and e.ev in ('SWING_DAMAGE', 'SPELL_DAMAGE', 'SPELL_PERIODIC_DAMAGE'))
buffs = [e for e in win if e.dst.startswith(cl.ME) and e.spellId == ED
         and e.ev in ('SPELL_AURA_APPLIED', 'SPELL_AURA_REFRESH')]

rows = collections.defaultdict(lambda: collections.Counter())
for c in crits:
    got = any(0 <= b.t - c.t <= WINDOW for b in buffs)
    shared = any(o is not c and abs(o.t - c.t) <= WINDOW for o in crits)
    rows[kind(c)]['shared' if shared else ('buff' if got else 'no buff')] += 1

print(f'{"crits by source":24} {"landed":>6} {"crits":>5} {"buff":>5} {"no buff":>7} {"shared":>6}')
for k in sorted(rows, key=lambda k: -sum(rows[k].values())):
    r = rows[k]
    print(f'{k:24} {hits[k]:6} {sum(r.values()):5} {r["buff"]:5} {r["no buff"]:7} {r["shared"]:6}')

print(f'\n{len(buffs)} buff events (applied or refreshed)')
for b in buffs:
    before = [f'{kind(c)} {c.ts}' for c in crits if 0 <= b.t - c.t <= WINDOW]
    print(f'  {b.ts} {b.ev[11:]:8} after: {", ".join(before) or "no crit of ours"}')
