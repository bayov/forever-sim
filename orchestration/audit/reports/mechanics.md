# Mechanics pass 42 (2026-10-27, master 06a898751)

Range audited: d3df9247f..06a898751. Four sim commits; TASKS.md gained an item for testing Amplify Curse on Bane of Agony (7166da482). No `forever-update/*` merge.

## Commits

| Commit | Area | Check | Verdict |
|---|---|---|---|
| 500e3a1 | Racials (shared) | Every value against dev notes post 7 (tier 1): Human 1% crit, Spirit x1.10; Dwarf Earth Fury gone, +1% melee/spell hit; Tauren 5% health only; Orc Savage gone, Command x1.05 on non-guardian hunter/warlock pets (same code as Classic racials.go, pets exist before applyRaceEffects, no later overwrite of pet DamageDealtMultiplier); Skyborne 2% haste, Energized +50% for 15 sec. critDamageBonus removed with no users left. | Matches; resolves mechanics-20261027-1 |
| d6fcdd3 | Warlock Spellstone | Shadow only under Forever, as client 70630 1237162/4/5 Misc 32 | Matches; resolves mechanics-20261025-1 |
| db6e43f | Mage Ignite | Arcane Instability and Arcane Power kept off Ignite's DamageMultiplier under Forever, per dev notes 2026-09-24 "no longer double dips" | Consistent with tier 1 |
| a52cdcf | Mage Missile Barrage | User ruling | Not reviewed (ruling) |

Shared run vs published: Human/crit builds fall 0.7-1.7% as the racial pass predicts; arcane mage -8.5% is the Missile Barrage ruling.

## Get current

- Client: newest 1.60 build is still 1.60.1.70630 (probe 300).
- Dev notes thread: 7 posts, newest post 7 (2026-10-26 20:35, edited 21:09), already read. Blue tracker: nothing newer on Forever mechanics (end-of-beta, RP-PvP, podcast posts).

## Open

- mechanics-20261027-2 (Amplify Curse x1.5 still on Bane of Agony, curses.go:74): still open; TASKS.md now gates presets on it.
- Recheck the post 7 racials when a client build ships rows for Artisan's Precision and Command.
