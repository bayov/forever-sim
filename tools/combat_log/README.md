# Combat log helpers

Scripts that read a WoW Forever combat log: the parser, damage and proc counts, glancing,
parry haste, partial resists and gear swaps. They started in the beta verification session
(`beta_handoff.md` has the tests and the recheck list they served), but they work on any log
from the beta or the live game. They moved here from a session scratchpad when `/tmp` was
wiped on 2026-10-10.

## Reading the log

The user turns logging on with `/combatlog` in the game, and the file lands in the client's
`Logs` directory as `WoWCombatLog-MMDDYY_HHMMSS.txt`. For the beta that is
`/mnt/d/Gaming/World of Warcraft/_classic_beta_/Logs/`.

Each client start begins a new file (on 2026-10-10, `WoWCombatLog-101026_150733.txt` up to
22:16 and `WoWCombatLog-101026_221851.txt` after a restart), so take the newest one.

The file grows while the user plays. Copy it to a scratch file first and run every script on
that copy, or two scripts can read different data:

```sh
cp "/mnt/d/Gaming/World of Warcraft/_classic_beta_/Logs/WoWCombatLog-101026_150733.txt" /tmp/log.txt
python3 -I tools/combat_log/recheck.py /tmp/log.txt 21:32:25 [22:16:00]
```

Every script takes the log and an optional window, as `HH:MM:SS` strings compared against
each line's time of day. They import `cl.py` from their own directory, so `python3 -I` works.

Some scripts hold numbers from one character and one day, like the weapon models in
`recheck.py` and the fist weapon cutoff in `glance.py`. The last section lists them.

## The scripts

- `cl.py` parses the log into events. `cl.dmg(e)` reads a damage line's tail (amount, raw,
  overkill, school, resisted, blocked, absorbed, crit, glancing, crushing), and `cl.miss(e)`
  reads a miss type. `cl.ME` is our character's name prefix ("Bayov-").
- `recheck.py` runs most of the recheck list over a window: white and Stormstrike raw against
  the weapon models, Lightning Shield orbs, Flametongue procs per trigger, Searing bolts,
  glancing and partial resists by mob level, Fire Nova's spell id and raw, the Stormstrike
  buff and mark, and Healing Stream crits.
- `ftt.py` sorts every landed white swing by the Flametongue Totem's state (in range, a kill,
  the mob dying within 130 ms, the totem down, past its 5 minutes or over 30 yards) and
  whether it procced.
- `crit.py` counts spell crits by group (Flametongue Totem, Flametongue Weapon, orbs, the
  rest) and mob level.
- `glance.py` counts glancing blows by mob level. It leaves out fist weapon swings from
  21:31:52 on 2026-10-10 (`FIST_FROM`), because the user leveled Unarmed then.
- `fistglance.py` splits white swings by weapon (a raw under 150 is the fist) and mob level,
  with glances, crits and miss types.
- `parry.py` checks the mob's next swing after each parry against Classic's parry haste rule.
- `respec.py` reports the Elemental respec run (item 12): each fire and shock spell's hits,
  crits and raw, the auras on us, and the auras that came with each of our spell crits.
- `searing.py` times Searing Totem's bolts against the totem to mob distance.
- `resists.py` lists partial and full resists of our spells on one mob name.
- `gear.py` prints our logged max health, spell power, armor, max mana and item level from
  our casts whenever they change, which shows gear swaps and stat buffs or debuffs.
- `client_spell.py` prints the client's rows for spell IDs from a directory of wago.tools
  CSVs (SpellName, SpellAuraOptions, SpellClassOptions, SpellEffect). Fetch each table from
  `https://wago.tools/db2/<Table>/csv?build=1.60.1.70338` with a desktop user agent.

## Things that are fixed in the code

- The weapon models in `recheck.py`: Rage of the Storm (3.3) at 389 and 392 attack power,
  and Bloody Brass Knuckles (1.6, 24 to 46) at 397. Update them from a new `/bayov export`.
- The mob's level is the level field on lines the mob is the source of (`f[-11]` on
  SWING_DAMAGE, `f[-12]` on SPELL_DAMAGE). On our own lines that field is our item level.
