# Client updates: pass 8 (2026-11-03, master f5c99ad018)

Forced pass for round 70753 phase 3. Build 1.60.1.70753 was processed by pass 7; this pass checks the round's port of
Blizzard's development notes post 9 (2026-11-02) and the 2026-11-03 server hotfix, and the standards 37 notes-only values.

## Sources

| Source | Result |
|---|---|
| wago.tools (`wago.py latest --probe 300`) | 1.60.1.70753, no newer build |
| Dev notes thread 2360696 (fetched 17:52Z) | posts /1-/9; /9 (2026-11-03T03:20Z) identical to `ref/client_update/sources/devnotes_post9_2026-11-02.txt` (whitespace only); /1 re-dated, no text change of note |
| foreverchanges.pro/patch-notes | newest entry is the 2026-11-03 hunter hotfix, same 8 spell lines as `sources/foreverchanges_2026-11-03_hotfix_and_70753.txt` |

## 2026-11-03 hotfix (server data) against the sim

All 8 lines landed: Aspect of the Hawk 20-110 (+120 r7), Trueshot Aura 20/35/50/75/100, Barrage 5/10/15%, Sniper Shot
320/450/590, Aimed Shot 40-350, Arcane Shot 25-280, Serpent Sting 15-675, Lightning Reflexes 3%. Ledger rows HUN-HF-03..10.

## Notes-only values (standards 37), per class round file

| Class | Values checked against the notes' text | Text match | Gated |
|---|---|---|---|
| Druid | Shred 200, Claw 140, Ravage 435, Rip/Rake/Bite x1.25 (Bite Energy part unscaled), Starfire x1.4, Rend and Tear 3% | yes | yes |
| Hunter | Arcane Shot AP 0.15, Serpent Sting 5% a tick | yes | no (Forever-only files) |
| Mage | Clearcasting kept on Missile Barrage AM, Ignite on Flamestrike burn | yes | yes |
| Paladin | Vengeance 5 stacks, 2H Spec on Holy 2/4/6% | yes | 2H yes, Vengeance stacks no |
| Priest | none (SW:D 30% is in the client) | - | - |
| Rogue | poisons 0.009/0.01, Mutilate 90%, Flawless 15, Aggression 5/10/15 (no Backstab), Ghostly Strike 7 s, Hemorrhage on Garrote, Lethality 10/20, Relentless 4%/pt/CP 25 Energy, Seal Fate 33/66/100, Coup de Grace 7/13/20, Efficient Strikes 3/5 (SS, Hemo), Composure +5/+10, DW Spec 25/50, Opportunity 4%/pt (Backstab, Garrote, Ambush, Mutilate), Shank (5 Energy, 10 s, 50% OH, 15/25 Energy, no miss), Flash of Steel, Boot Knife, Venom 30%/20%/6 s | yes | poisons and Poised only |
| Shaman | LB 0.78 x Era (r10 334-372), CL 1.45 x LB / 2.25 x mana (248/304/360/495 vs LB 110/135/160/220) / 2.5 s, Era shocks, Overload 10/20/30, MW 4 PPM, Stormstrike 150% 6 s, Mental Dexterity 50/100/150, Flurry 6-30% | yes | Stormstrike only |
| Warlock | Wrack 36 + 2.5/level (86 at 60), 0.286; Shadow and Flame 4%; pets 2 HP per Stamina | yes | Forever-only talents |
| Warrior | Slam 15 s, Imp. Slam 2.5/5, 2H 2/4/6, ITM 3x5, Bloodthrill 2x10, Flurry 3x10 no DW req, Reckless Swings, Improved Whirlwind | yes | yes |
| Shared | race base stats (Skyborne open), Read Leyline | yes | - |

Every listed value matches the notes. The ungated ones are finding client_update-20261103-2 (low, user ruling): no
Forever number moves, but standards 37 and rogue.md say "gated", and rogue Relentless Strikes and Seal Fate were
Classic's values before the round.

## Talent trees from the notes (rogue, warrior, shaman)

`tools/forever_talents/notes/<class>.json` checked against the notes' rows, ranks and prerequisites (every moved row,
new talent, rank change, removal and prerequisite named in post 9), and `ui/core/talents/trees/<class>.json` checked
against the JSON: no cell collisions, every prerequisite resolves, ranks = maxPoints, removed talents gone. All agree.

## Findings

- New: client_update-20261103-2 (low, user ruling): notes-only values not gated (rogue, shaman, hunter, paladin).
- Resolved: client_update-20261103-1 Poised (druid 35408c0af, rogue 31ff01f80).
- Still open, no new evidence in post 9: client_update-20261030-2 (Missile Barrage tick dummy), -20261030-4 (Fengus).

## Ledger

`orchestration/update_ledger.md`, build 70753 section: ROG-U-01 and DRU-U-15 landed; new subsection for post 9 (per-class
ledgers live in `orchestration/round_70753/<class>.md`) and the 11-03 hotfix rows HUN-HF-03..10 (supersede HUN-HF-01/02).
Summary: Hunter 93 rows / 52 landed, Rogue and Druid one more landed each; total 765 / 380 landed / 31 missing.
`foreverClientBuild` stays 1.60.1.70753.
