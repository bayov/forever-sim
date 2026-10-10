"""The Elemental respec run (beta_handoff.md item 12). Usage: python3 -I respec.py LOG FROM [TO]"""
import sys, os, collections
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
import cl

evs = cl.load(sys.argv[1]); evs.sort(key=lambda e: e.t)
lo = sys.argv[2]; hi = sys.argv[3] if len(sys.argv) > 3 else '99'
win = [e for e in evs if lo <= e.ts < hi]
me = [e for e in win if e.src.startswith(cl.ME)]
lvl = {}
for e in evs:
    if not e.srcGUID.startswith('Creature'): continue
    try:
        if e.ev == 'SWING_DAMAGE': lvl[e.srcGUID] = int(e.f[-11])
        elif e.ev in ('SPELL_DAMAGE', 'SPELL_PERIODIC_DAMAGE'): lvl[e.srcGUID] = int(e.f[-12])
    except (ValueError, IndexError): pass
print(f'window {lo} to {win[-1].ts}')

def show(name, evs_):
    hits = [e for e in evs_ if e.ev in ('SPELL_DAMAGE', 'SPELL_PERIODIC_DAMAGE')]
    miss = [e for e in evs_ if e.ev in ('SPELL_MISSED', 'SPELL_PERIODIC_MISSED')]
    d = [cl.dmg(e) for e in hits]
    print(f'{name}: {len(hits)} hits, {len(miss)} missed {dict(collections.Counter(cl.miss(e) for e in miss))}, crits {sum(x["crit"] for x in d)}, resisted {sum(bool(x["resisted"]) for x in d)}')
    print('   normal raw', sorted(x['raw'] for x in d if not x['crit']))
    print('   crit raw/amount', sorted((x['raw'], x['amount']) for x in d if x['crit']))
    print('   ids', sorted({e.spellId for e in evs_}), 'levels', dict(collections.Counter(lvl.get(e.dstGUID) for e in hits)))

show('Fire Nova', [e for e in me if e.spell == 'Fire Nova' and e.ev != 'SPELL_CAST_SUCCESS' and e.ev.startswith('SPELL_') and ('DAMAGE' in e.ev or 'MISSED' in e.ev)])
show('Flametongue Totem', [e for e in me if e.spellId == 16368])
show('Flame Shock direct', [e for e in me if e.spell == 'Flame Shock' and e.ev in ('SPELL_DAMAGE', 'SPELL_MISSED')])
show('Flame Shock ticks', [e for e in me if e.spell == 'Flame Shock' and e.ev in ('SPELL_PERIODIC_DAMAGE', 'SPELL_PERIODIC_MISSED')])
for sp in ('Earth Shock', 'Frost Shock', 'Lightning Bolt', 'Lightning Shield', 'Flametongue Weapon'):
    x = [e for e in me if e.spell == sp and ('DAMAGE' in e.ev or 'MISSED' in e.ev)]
    if x: show(sp, x)
tot = {e.dstGUID for e in evs if e.ev == 'SPELL_SUMMON' and e.src.startswith(cl.ME) and 'Searing' in e.dst}
x = [e for e in win if e.srcGUID in tot and ('DAMAGE' in e.ev or 'MISSED' in e.ev)]
if x: show('Searing Totem', x)

# Auras on us, and what came with each of our spell crits.
aur = collections.Counter((e.ev, e.spellId, e.spell) for e in win if e.dst.startswith(cl.ME) and e.ev.startswith('SPELL_AURA') and e.spell not in ('Thrill of Adventure',))
print('auras on us', dict(aur))
for e in me:
    if e.ev in ('SPELL_DAMAGE', 'SPELL_PERIODIC_DAMAGE') and cl.dmg(e)['crit'] and cl.dmg(e)['school'] != 1:
        near = [f'{a.ev} {a.spell}' for a in win if a.dst.startswith(cl.ME) and a.ev.startswith('SPELL_AURA') and -0.05 <= a.t - e.t <= 0.3 and a.spell not in ('Thrill of Adventure',)]
        print('  crit', e.ts, e.spell, e.ev, near)
