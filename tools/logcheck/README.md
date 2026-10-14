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
- how often the amount the mob took rounded up, for abilities whose hit has one fixed value
  (below)
- partial resists on spells, by the share resisted
- for what we didn't cast (procs), how many came per 100 landed white swings

A row with less than a 1 in 1000 chance under the sim gets `!!` and goes in the list at the
end. So each new log rechecks everything we've settled so far.

```sh
go run --tags=with_db ./tools/logcheck -from 21:32:25 -to 22:16:00
```

Run it from the repo root. A run takes about a second.

## The audit

```sh
go run --tags=with_db ./tools/logcheck -audit
```

This checks every stretch of the archived logs that we haven't checked yet. The `log-audit` skill
(`.claude/skills/log-audit/SKILL.md`) runs it and triages what it flags.

It splits each log into windows where the sim's character holds still: one of our characters,
with the same gear, talents and level all through. It cuts at each change BayovCore's history
recorded, at each change of the item level on our casts that the history doesn't have, at each
logout or `/reload`, and at breaks of more than 30 minutes between fights. It runs logcheck on
each new window (four at a time) and saves the report in `~/forever-logs/audit/reports/`. The
ledger `~/forever-logs/audit/ledger.jsonl` keeps each window with its flags, so the next audit
skips it. A window with fewer than 60 of our attacks is noted as skipped.

A window waits until the history covers it, because WoW saves the history only at logout and
`/reload`. The addon writes a logout record just before WoW saves, so a window is ready once
the character has a history record from after it ends. A character with no history waits
until the log has been quiet for 30 minutes.

The digest prints each new window's flags, marked `NEW` when no earlier window flagged that
ability and kind, or `known` when `known.json` (next to this README) has them. Then it counts
each ability and kind across the whole ledger. Windows where the character changed in a way
the sim can't follow (a `change` flag, only seen before the history) are left out of the
counts, because their flags may only show that change. `-redo` runs every window again with the
sim as it is now, which is how we see what a fix cleared.

## Inputs

- **The log.** With no `-log`, it runs `tools/combat_log/archive.sh` first and reads the newest
  log in `~/forever-logs/_classic_beta_/`. `-from` and `-to` take `HH:MM:SS` on the log's first
  day (a time before the log starts is the next day) or `YYYY-MM-DD HH:MM:SS`.
- **The export.** With no `-export`, it takes the newest export taken before the window ends,
  of the one character of ours (one we have exports of) that has lines in the window. When we
  played two of them in the window, `-char Name-Realm` picks one.
  It looks at the pasted ones in `~/forever-logs/exports/` and at BayovCore's history in
  `~/forever-logs/history/` (below). When the user pastes an export, save it in `exports/` as
  `export-<date>_<time>.json` with the date and time from its `when.date`. The header prints which
  one it took and how far from the window.
- **The history.** BayovCore's export module (`Modules/Export.lua` in bayovui) also records the
  character as it changes. It takes a full snapshot at login, `/reload` and `/bayov export
  mark`. After that it records each gear swap, enchant, talent change, level up and skill up,
  within about 2 seconds of when it happened, and the sheet after a gear, talent or level change.
  WoW writes it to disk on `/reload` or logout, and the addon adds a logout record just before,
  which says the history is complete up to then. `archive.sh` runs
  `tools/combat_log/history.py`, which adds the new records to
  `~/forever-logs/history/<Name>-<Realm>.jsonl` (the sheets and talent lists are in the
  `.blobs.json` next to it). logcheck rebuilds a whole export for each record, so the header
  names a history line like `Bayov-ClassicBetaPvP.jsonl:5 (history, change)`.
- **Client tables.** It downloads SpellName and SpellEffect for the export's client build from
  wago.tools into `~/forever-logs/wago/<build>/`, or uses the newest build there when wago.tools
  doesn't have it.

## What the sim gets

- The export's race, level, gear, enchants, talents and professions. Gear goes in by item type,
  enchant IDs go through the client's SpellEffect table when the sim's database keys an enchant
  by its spell. An enchant the sim's database lacks (Forever's leg and boot enchants 8482 and
  8485) goes in as bonus stats from the client's SpellItemEnchantment row, when all it does is
  add stats. Anything it can't place is listed as a note.
- The export's weapon skills, as the sim's main hand and off hand skill gaps (`-full-skill`
  turns that off). Misses, dodges, parries, glances and crits follow them. When we swung a
  weapon below full skill in the window, the report prints a warning and puts it in the flagged
  list ahead of the melee rows, because skill rises as we swing and the sim holds it still.
- We stand in front of the mob, where it can parry and block, as when questing (`-behind` for
  the other way).
- Our shaman imbue, from the last imbue we cast before the window ends (the export can't see it
  under Forever). When the cast is in an earlier log file, the imbue's procs in the window say
  which one it is.
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
- **Amounts.** For an ability whose every sim hit has the same value (Flametongue's attack,
  Flame Shock's ticks, Lightning Shield's orbs), the amount row counts how often the log's
  amount was raw + 1 and tests that count against the sim's fraction. The test is exact
  (Poisson binomial over each hit's own chance, a crit has its own fraction), so a handful of
  hits only flags when it's far off. The row also says where the log alone puts the hit, as a
  95% interval: 7 Flame Shock hits put it at 64.11-64.62, while 201 Flametongue attacks put it
  at 8.73-8.84. Hits with another raw (other stats) or an amount no hit with that raw can give
  (a modifier the sim doesn't have, like the Stormstrike mark) are left out and counted. The
  hit doesn't depend on the mob's level, so "Amounts on all mob levels together" runs the same
  check on every level's hits at once. A spell with a damage range can land anywhere between
  two whole points, so its round ups say nothing and it gets no amount row.
- **Changes between exports.** The log has no character level or skill (the level field on
  our own lines is our item level). So we compare the character's exports from the last one
  before the window to the first one after it, and warn about a level up, a skill up of a
  weapon either export holds, a gear swap or a talent change. A change between two exports
  happened somewhere between them, so the warning says which exports bracket it. With no export
  before or after the window, a note says changes at that end don't show. Between two history
  records the warning gives the time of the change instead, and a change recorded after the
  window ends isn't one. Weapon skill ups are summed into one line per skill, because history
  has a record for each point. Without history, take an export right before and right after a
  test to pin it down.
- **Our logged stats changed.** Each of our casts carries max health, attack power, spell power,
  armor and item level. A change means a gear swap, a stat buff or a level up (more health with
  the same item level says a level up or a buff), and the sim uses the export's gear for the
  whole window. Pick a window without swaps for damage checks.
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
