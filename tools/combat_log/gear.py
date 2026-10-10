"""Our logged stats on each of our casts, printed when they change. Usage: python3 -I gear.py LOG [FROM] [TO]

Each SPELL_CAST_SUCCESS line carries the caster's advanced fields: max health, attack power,
spell power, armor, max mana and item level. A change in them shows a gear swap or a stat
buff or debuff, like the user's gear swap at 22:11:30 on 2026-10-10 (spell power 49 to 53).

The spell power field is the gear's spell power without Mental Quickness (it read 49 when the
sheet said 91 with Mental Quickness 2). The attack power field doesn't match the sheet, so it
only shows that something changed.
"""
import sys, os
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
import cl

evs = cl.load(sys.argv[1]); evs.sort(key=lambda e: e.t)
lo = sys.argv[2] if len(sys.argv) > 2 else '00'
hi = sys.argv[3] if len(sys.argv) > 3 else '99'
last = None
for e in evs:
    if e.ev != 'SPELL_CAST_SUCCESS' or not e.src.startswith(cl.ME) or not lo <= e.ts < hi or len(e.f) < 31:
        continue
    f = e.f
    key = (f[15], f[16], f[17], f[18], f[24], f[30])
    if key != last:
        print(e.ts, 'maxHP', f[15], 'AP', f[16], 'SP', f[17], 'armor', f[18], 'maxMana', f[24], 'ilvl', f[30], '|', e.spell)
        last = key
