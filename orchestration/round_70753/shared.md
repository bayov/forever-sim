# Round 70753: shared sim/core (A2)

The shared lines of the November 2 development notes (post 9): the Races section, the base stat fix, Throw, the
low-rank proc penalty. Source tiers per standards 37.

## Ledger

| Note line | Source that settles it | Sim |
| --- | --- | --- |
| Dwarf: Earth Fury removed, Artisan's Precision +1% spell and attack hit | client 70753 1259719 (auras 54 / 55, 1%) | already in (since the post 7 pass) |
| Human: Ruthless 1% crit, Human Spirit +10% Spirit | client 70753 20597 aura 290 1, 20598 aura 137 10 | already in |
| Orc: Savage removed, Command +5% hunter and warlock pets, Blood Fury 2 min | client 70753: no Savage in skill line 125, Command 20576 (hunters) / 20575 (warlocks), 20572 cooldown 120000 | already in |
| Skyborne: Wind Blessed 2% haste | client 70753 1259710 auras 342 / 65 at 2 | already in |
| Skyborne: Read Leyline 50% regen for 15 sec, Attuned 100% at a Ley Line | client 70753 1270842 bp 50, DurationIndex 8 (15 sec); Attuned 1259691 | already in; no Ley Line in the sim's fights |
| Read Leyline cast time (not in the notes) | client 70753 SpellMisc CastingTimeIndex 3 (0.5 sec; 70707 had 2 sec) | 60f857c94, mechanics-20261103-1 |
| Read Leyline can't be used under Innervate, Innervate replaces it | notes only (no SpellAuraRestrictions row) | 60f857c94 |
| Tauren: Endurance only +5% health | client 70753 20550 aura 133 5, nothing else | already in |
| Troll: Berserking 15 sec | client 70753 20554 DurationIndex 8 | already in |
| Undead: Touch of the Grave 15 sec lockout | client 70753 SpellAuraOptions ProcCategoryRecovery 15000 (1260189, 1260201) | already in |
| All races have Classic's base stats (Tauren warrior +5 Strength over Human) | notes only (the client has no base stat table) | this round's base stats commit: RaceOffsets and the 1.12 rows below 60 |
| Throw uses a 0.5 sec windup | n/a: the sim has no Throw (2764); only hunters auto shoot, and their windup is RangedSwingWindup 0.5 sec already | none |
| Low-rank proc penalty removed (Blizzard keeps it for ranks learned more than 12 levels below) | notes | none: the sim never modelled the penalty (no class or core code scales a proc chance by rank). Blizzard's remaining penalty is not modelled either; only a preset casting a Blizzard rank learned before level 48 would need it |
| Root diminishing returns on procs | n/a (PvP and adds, no DPS) | none |

## Notes-only values (standards 37)

- Race base attributes: Classic's offsets (forever-wiki/racials.md, Base attributes). Open: the Skyborne's corrected
  values ("This fix was also applied to High Order and Windshaper Skyborne base stats"); the sim keeps their
  2026-09-22 sheet offset (-1 Str, +1 Agi, -1 Sta, +1 Int) until a naked Skyborne sheet shows the new one.
- Read Leyline's Innervate rule.
- Sparklance Wand (280031) deals Arcane damage: Blizzard's hotfixes of 2026-11-10 ("now does Arcane damage (was
  Nature damage)"); its ItemSparse row is hidden at 1.60.1.70888, so `foreverWandSchools` in
  `tools/database/gen_db/main.go` sets the school. Kept until a log of its wand shots or its ItemSparse row checks it
  (user ruling 2026-11-12, TASKS.md). No build uses it.
