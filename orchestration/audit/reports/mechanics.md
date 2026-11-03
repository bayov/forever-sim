# Mechanics pass 61 (2026-11-03, master f5c99ad01)

Range 9da90fcde..f5c99ad01 (20 non-merge sim commits), forced for round 70753 phase 3 (client 1.60.1.70753,
dev notes post 9). No shared sims were used; nothing here needed new numbers.

## Get current

| Source | Read | Result |
|---|---|---|
| Dev notes thread 2360696 | posts 1-9 via the forum JSON | post 9 (created 2026-11-03T03:20Z) is the newest; the saved copy (ref/client_update/sources/devnotes_post9_2026-11-02.txt) matches the live text line for line |
| Client | 1.60.1.70753 | still newest (gate) |
| Server hotfixes on 70753 | wago.tools search pages | newest spell push 113557, as seen in pass 60 (rows re-logged under region 95 on 11-03 14:14) |
| Daily update | TASKS.md / update_ledger.md | only tank-bench commits; no forever-update merges |

## Shared commits

| Commit | Change | Check | Verdict |
|---|---|---|---|
| 6d9b4d7e6 | Dragonfire and Caustic Smog can't crit | 70753 SpellMisc Attributes_2: 1252554 0x20000000, 1252543 0x30080000 | holds (tier 1) |
| c975a8297 | Trueshot Aura 20/35/50/75/100 | foreverchanges 2026-11-03 hotfix text, post 9 line 172 | holds |
| 60f857c94 | Read Leyline 0.5 sec; not under Innervate, ended by it | CastingTimeIndex 3 (pass 60); Innervate rule is post 9 wording | holds; notes-only part listed in shared.md |
| 7598b8015 | Classic race base attributes | post 9 line 23; 1.12 level 1 offsets | holds (notes only, tier 1 + tier 3 table); Skyborne and Gnome Int open per README |
| cee6b8333 | bear swing Rage as two-hander (RageAsTwoHand) | wowsims/forever combat-log reading, cited in bear.go | plausible, tank only |
| 35408c0af / druid | ForeverDemoralizingRoarAura 30 sec | existing rank table | tank only |

## Notes-only values (standards 37), all nine classes

Each class round file's "Notes-only values" list was checked against post 9's wording and a code spot check for the
Forever gate.

| Class | Values | Notes match | Gate / cite |
|---|---|---|---|
| Druid | Shred 200, Claw 140, Ravage 435, Rip/Rake/Bite x1.25, Starfire x1.40, Rend and Tear 3% | yes | `TernaryFloat64(IsForever)` in each file |
| Hunter | Arcane Shot 0.15, Serpent Sting 5% | yes | Forever-only shot files (stated) |
| Mage | Clearcasting kept under Missile Barrage, Flamestrike ignite | yes | IsForever |
| Paladin | Vengeance 5 stacks, 2H Spec on Holy 2/4/6 | yes | Forever-only talents, cited |
| Priest | none | - | - |
| Rogue | poisons 0.009 / 0.01, Mutilate 90% (flat x0.9 too), Flawless 15, Aggression 5/10/15, Venom, Shank, talents | yes | poisons gated; talent values sit in the Forever-only tree |
| Shaman | LB 0.78 x Era (rank 10 +76%, notes also say "+80%"), CL 1.45x / 2.25x / 2.5 s, Era shocks, LO 10/20/30, MW 2x, SS 150% 6 s, MD 50/100/150, Flurry 6-30, Totemic Call | yes | Forever tables |
| Warlock | Wrack 86 at 60 / 0.286, S&F 4%, pets 2 hp per Sta | yes | Forever talents |
| Warrior | Slam 15 s, Imp Slam 2.5/5, 2H Spec 2/4/6, ITM 3x5, Bloodthrill 2x10, Flurry 3x10, Reckless Swings, Imp Whirlwind | yes | IsForever |

Missing: the tank threat lines (Bear 115%, Defensive Stance 80%, Defiance 7/13/20%): mechanics-20261103-2.

## Talent trees from tools/forever_talents/notes (rogue, warrior, shaman)

Diffed client/<class>.json before (78b617dc2) and after the notes overlay, and the generated ui trees, against post 9's
rows, ranks and prerequisites. Every row move, rank count, rank value, removal and prerequisite named in the notes is
there: rogue (Murder, Improved Distract, Quietus, Cutthroat, Thousand Cuts gone; Blood Splatter, Composure, Coup de
Grace, Efficient Strikes, Flash of Steel, Quick Getaway, Boot Knife, Death Mark added; 20+ row moves), warrior (Enrage
r3, Reckless Swings r4, Gore Drinker r5 under Blood Craze, Flurry 3 points without Death Wish, Improved Whirlwind r6,
Bloodthirst under Death Wish, ITM 3, Bloodthrill 2), shaman (Call of Thunder for Lava Burst, Lightning Mastery under it,
Improved Ghost Wolf r4). Guesses the notes don't settle (columns, Venom and Weapon Expertise losing prerequisites that
would point downward, Lightning Mastery needing Call of Thunder) are labelled as guesses in the notes files. No finding.

## Findings

- New: mechanics-20261103-2 (low): tank threat notes-only values missing.
- Resolved: mechanics-20261031-1 (Efficiency on Lacerate), c975a8297.

## Left open

- Human base Spirit: the sim takes 20 as the Human baseline (21 in 1.12 read as The Human Spirit baked in); no tier 1
  source decides it. 1 Spirit on every non-Human offset.
