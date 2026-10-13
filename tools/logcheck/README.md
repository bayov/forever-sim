# logcheck

Compares a WoW Forever combat log with the sim.

We give it a log window and the `/bayov export` of the character who played it. It builds that
character in the sim and replays our casts at the times the log has them, against a mob of the
same level, a few hundred times. Then it prints each ability side by side, log and sim:

- how often it hit, crit, glanced, was blocked, missed, was dodged or parried, with the log's
  95% interval and the chance of the log's count under the sim's rate (exact binomial)
- its raw damage, which is the hit before the mob's armor, resists, crit and damage taken
  modifiers (the log's raw field and the sim's `RawDamage`), as a range and an average. The sim's
  hits are truncated to whole points first, because the log's raw field is truncated (below)
- partial resists on spells, by the share resisted
- for what we didn't cast (procs), how many came per 100 landed white swings

A row with less than a 1 in 1000 chance under the sim gets `!!` and goes in the list at the
end. So each new log rechecks everything we've settled so far.

```sh
go run --tags=with_db ./tools/logcheck -from 21:32:25 -to 22:16:00
```

Run it from the repo root. A run takes about a second.

## Inputs

- **The log.** With no `-log`, it runs `tools/combat_log/archive.sh` first and reads the newest
  log in `~/forever-logs/_classic_beta_/`. `-from` and `-to` take `HH:MM:SS` on the log's first
  day (a time before the log starts is the next day) or `YYYY-MM-DD HH:MM:SS`.
- **The export.** With no `-export`, it takes the newest one in `~/forever-logs/exports/` taken
  before the window ends. When the user pastes an export, save it there as
  `export-<date>_<time>.json` with the date and time from its `when.date`. The header prints which
  one it took and how far from the window.
- **Client tables.** It downloads SpellName and SpellEffect for the export's client build from
  wago.tools into `~/forever-logs/wago/<build>/`, or uses the newest build there when wago.tools
  doesn't have it.

## What the sim gets

- The export's race, level, gear, enchants, talents and professions. Gear goes in by item type,
  enchant IDs go through the client's SpellEffect table when the sim's database keys an enchant
  by its spell. An enchant the sim's database lacks (Forever's leg and boot enchants 8482 and
  8485) goes in as bonus stats from the client's SpellItemEnchantment row, when all it does is
  add stats. Anything it can't place is listed as a note.
- The export's weapon skills, taken off the sim's full skill (`-full-skill` turns that off). The
  sim's glancing and parry chances don't follow skill, so an under-skilled weapon still shows
  high glances and parries.
- We stand in front of the mob, where it can parry and block, as when questing (`-behind` for
  the other way).
- Our shaman imbue, from the last imbue we cast before the window ends (the export can't see it
  under Forever).
- Mob hits on us from the log, as raid damage hits per minute, so Lightning Shield goes off.
- No buffs, no consumes, and no rotation beyond the replayed casts. The mana is refilled before
  each cast and the GCD freed, because neither is what we check here.

The replay runs once per mob level. Damaging casts are replayed only when they went at a mob of
that level (or, with no target like Fire Nova, during a fight with one). Buffs and totems are
replayed in every level's run.

## Reading the report

- **Sheet check.** The export's sheet against the sim's stats for the same character. The sheet
  has the buffs that were up when the user exported, so a gap can be a buff. Melee crit takes
  0.04% off per missing point of main hand skill, like the client's sheet. Any gap prints a
  warning and goes first in the flagged list, because every damage row depends on these stats.
  Check it before reading a damage flag: on 2026-10-13 a 6 spell power gap (two enchants the
  sim lacked) made Flame Shock and Frost Shock look wrong when they weren't.
- **Rounding.** The log's raw field is truncated to whole points, so we truncate the sim's hits
  before comparing raw. With the sheet's spell power the sim gives Flame Shock ticks 21.7,
  Flametongue Totem's attack 8.8 and Flame Shock's direct hit 64.3, and the 10-10 log has raw
  21, 8 and 64 on every hit. Rounding would give 22 and 9. `-rounding round` compares the other
  way. The damage the mob takes (the amount) is rounded differently: up with a chance equal to
  the fraction, so on average it's the unrounded hit. In the same window Flametongue Totem's
  attack did 9 on 148 of 188 hits (79%, for 8.8), the Flame Shock ticks 22 on 29 of 41 (71%, for
  21.7) and the direct hit 65 on 2 of 7 (for 64.3). So the sim's damage needs no rounding.
- **Our logged stats changed.** Each of our casts carries max health, attack power, spell power
  and armor. A change means a gear swap or a stat buff, and the sim uses the export's gear for
  the whole window. Pick a window without swaps for damage checks.
- **Spell IDs.** The sim sometimes keeps one ID where the client logs another (Windfury
  Weapon's attacks are 8232 in the sim and 439440 in the log). The report pairs a sim-only spell
  with a log-only spell of the same client name and says so. `nameAliases` in `report.go` holds
  the pairs whose names differ too (Flametongue Totem Proc and Flametongue Attack).
- **Only in the sim** lists what the sim did that the log never had.
- `-min` (default 5) hides abilities with fewer attempts in the log.

## How it hooks into the sim

`core.EnvHooks` (in `sim/core/environment.go`) runs on each new sim just before it's finalized.
logcheck uses it to put a permanent aura on our character and pets that records every spell
result, and to take the skill gaps off. The sim itself never sets a hook.
