# Rogue: round 70753 (November 2 development notes)

## Final summary (landed on master)

**What changed.** The notes' rogue rework isn't in client 1.60.1.70753 or on Wowhead: the client tree is the old 53
nodes. So every line below is a notes-only value under standards 37, apart from Poised and Death Mark, which come
from the client. The tree is built from the notes in `tools/forever_talents/notes/rogue.json`. Columns and arrows
come from foreverchanges.pro's talent calculator, the only source that places the new talents. Sim changes:

- Poised: 10 Energy off the first ability; Composure adds 5 a rank.
- Shank: the notes-only spell, ID 1399248. Flash of Steel adds Energy and cuts its cooldown; Boot Knife gives it 2
  combo points.
- Venom: now a passive on Shank.
- Poisons: the attack power coefficients are doubled.
- Mutilate: 90% weapon damage.
- Flawless Execution 15.
- Aggression: 5/10/15% on Sinister Strike and Eviscerate.
- Ghostly Strike: opens Ambush and Garrote without Stealth.
- Hemorrhage: raises Garrote too.
- Re-ranked talents: Lethality, Relentless Strikes, Seal Fate, Dual Wield Specialization and Opportunity.
- Blood Splatter.
- Efficient Strikes: covers Hemorrhage too.
- Coup de Grace: raises Death Mark too.
- Death Mark: from the client (12% stored a combo point, Rupture and Garrote twice; cap 156 + 245 + 4% AP a point).
- Removed talents: Murder, Quietus, Cutthroat, Thousand Cuts and Improved Distract.

A test per rule is in `sim/rogue/dps_rogue/devnotes_70753_test.go`.

**DPS effect.** The mechanics alone, on the old builds, are worth +7 to +48 at 180 s (table under "Phase A"). With
the new tree, Shank in every rotation and the talents and rotations searched again, at 120 / 180 / 300 s (20k
iterations; seed 1, seed 1000003 within 0.6). Old is master 5eb381750, after A2:

| Build | Old | New | Race | Talents |
|---|---|---|---|---|
| Combat Swords (class default now) | 808.7 / 780.4 / 759.5 | 995.0 / 963.5 / 937.0 | Undead | `00530320201-3200330112013221551` |
| Seal Fate Swords | 757.4 / 745.5 / 735.0 | 975.9 / 957.0 / 941.0 | Undead | `0053032024105103-320033011201312` |
| Combat Daggers | 814.1 / 789.6 / 769.6 | 972.8 / 946.7 / 923.3 | Undead | `005203002-3203230002012221551-005` |
| Seal Fate Daggers | 765.1 / 751.9 / 742.6 | 936.9 / 920.7 / 906.4 | Undead | `0053032024105103-321303010201312` |
| Subtlety (now Death Mark + Boot Knife) | 746.8 / 731.8 / 722.3 | 896.5 / 881.7 / 868.3 | Night Elf (was Undead) | `00530301-320003-115001331320121151` |
| Mutilate | 754.6 / 742.7 / 734.4 | 881.0 / 866.1 / 855.3 | Undead | `00530320251451031-320303-005` |

The six "(vs Humanoid)" builds gave exactly their base builds' DPS once Murder was gone, and are dropped (user ruling 2026-11-03).

**Searches.** The rotopt talent search moves one point at a time and stalled on the new row gates (it never left
the seed). So I used a stochastic climb of 1 to 5 point moves: 150 valid neighbours a round at 3000 iterations, the
best 5 checked at 10000, kept for 2 SE. It alternated with the rotopt knob search over 120 / 180 / 300 s combined
until neither moved. Seeds were hand-built valid archetypes per build. Notable choices:

- Seal Fate builds: Venom in, full Dual Wield Specialization.
- Combat Daggers: Puncturing Wounds and Opportunity; Cold Blood and Vigor out.
- Subtlety: Death Mark, Boot Knife, Blood Splatter and Rupture at any combo point. It drops Ghostly Strike
  (the 7 s opening Ambush doesn't pay for the talent).
- Rotations: a Shank line in every template (Shank at 75 energy or less, which held in every search), a Death Mark
  line, and the Ghostly Strike Ambush and Garrote lines.

**Exploratory search (user request, 2026-11-03).** Six more archetypes went through the same talent climb from a
hand-built valid seed and then the knob search over 120 / 180 / 300 s combined. None beats the build it would
replace or join, so nothing changes:

| Candidate | Seed idea | Best (combined; talent climb at 180 s) | Against |
|---|---|---|---|
| Venom Combat Swords | 21 Assassination for Venom, Combat without Adrenaline Rush | 945.6 | Combat Swords 965.0 |
| Boot Knife Swords | 30 Subtlety for Boot Knife, 21 Combat | 770.8 | Combat Swords 965.0 |
| Death Mark Swords | 31 Subtlety, Hemorrhage builder on the sword set | 812.8 (the climb dropped Death Mark) | Combat Swords 965.0 |
| Ghostly Subtlety | Subtlety with Ghostly Strike and its Ambush window | 881.9 (the climb dropped Ghostly Strike again: the same build as Subtlety but a filler point) | Subtlety 880.5 |
| Mutilate, Subtlety side | Opportunity, Premeditation and Serrated Blades instead of Combat's Precision and Puncturing Wounds | 841.5 | Mutilate 867.4 |
| Seal Fate Daggers, Opportunity | Opportunity in Subtlety instead of Combat depth | 852.8 | Seal Fate Daggers 919.3 |

What the searches keep showing: Shank with Flash of Steel and Adrenaline Rush carries Combat; Venom pays only once
the Seal Fate tree is deep anyway; Boot Knife and Death Mark pay only together and on daggers with Hemorrhage;
Ghostly Strike's opening is worth less than the points around it.

**Races** (180 s, 20k, seed 1; each race's own rotation):

| Build | Ranking |
|---|---|
| Combat Swords | Undead 963.5, Orc 963.5 (tie, stays Undead), Gnome 959.9, Night Elf 958.5, Human 957.5 |
| Seal Fate Swords | Undead 957.0, Gnome (Eureka! override) 957.3 / 956.4 on the two seeds (not ahead on both), Orc 956.2, Night Elf 956.0 |
| Combat Daggers | Undead 946.7, Gnome 946.5, Orc 945.5 |
| Seal Fate Daggers | Undead 920.7, Gnome 918.7, Orc 916.7 |
| Mutilate | Undead 866.1, Gnome (override) 862.7, Orc 861.9 |
| Subtlety | Night Elf 881.7, Undead 880.2 (+1.4 / +1.9 over the Undead on both seeds, so it switches), Orc 876.4 |

- Best overall: the Undead, and the Night Elf on Subtlety.
- Best Horde: the Undead.
- Best Alliance: the Gnome, and the Night Elf on Subtlety.
- The Gnome overrides for Combat Swords and Subtlety are dropped (Eureka! on cooldown now loses there). Mutilate
  and Seal Fate Swords keep theirs, re-made.
- Windshaper and High Order sim identically on every build. Their racials don't reach a rogue's DPS here; check
  this is A2's intent.

**Audit.**
- Results: all 12 requests verified at 20k, seed 1, and the meta is updated. The requests come from mksettings with
  the same arguments; nothing but talents and rotation differs (checked by diff).
- Rotations: timelines and breakdowns of Combat Swords and Subtlety.
  - Combat Swords: Shank every 8 s, Slice and Dice 97%, Adrenaline Rush and Cold Blood placed as intended.
  - Subtlety: Death Mark goes off 9.4 times a fight (148 DPS); Vanish and Shadowmeld feed Ambush.
  - Every line states its intent (standards 31); each change was checked at 120 / 180 / 300 s.
- Gear: no gear change.
- Consumables: unchanged; standards 18 kept (no Expose Armor).
- Mechanics and client_update: the ledger below. Finding client_update-20261103-1 (Poised) is fixed for the rogue.
- Tests: `go test` on sim/... passes except `TestRaidForeverReportHashesStable`, which I rewrite after the last
  merge (the rogue results moved on purpose).

**Open questions for the user.**
1. **The six "(vs Humanoid)" rogue builds are dropped** (user ruling 2026-11-03): they only differed by Murder, which
   the notes removed. Their requests, presets and report hashes are gone; the report data still lists them until
   the orchestrator regenerates it.
2. **Builds to keep.**
   - Subtlety is now a Death Mark / Boot Knife build, so a new "Death Mark Subtlety" build isn't needed; the
     Subtlety build is it.
   - Shank / Boot Knife isn't its own build: Boot Knife needs 25 Subtlety points, and the searches only took it
     with Death Mark.
   - Recommendation: keep the six base builds.
3. **Assassination default.** Seal Fate Swords (957.0) now leads Mutilate (866.1) by 10%. The Assassination family
   default stays Mutilate in the meta and UI. Switch it?
4. **Notes-only assumptions** (notes.md Rogue NTV-6): Shank's damage model, Mutilate's flat bonus at 90%, the Ghostly
   Strike opening, and the unspecified Hemorrhage / Ghostly Strike dagger nerf (the sim keeps 145% and 180%). Plus
   Death Mark's cap applying before the crit roll, and ranks 2 and 3 using rank 1's cap.

**Shared changes.** None in sim/core or sim/common. I added the rogue file `tools/forever_talents/notes/rogue.json`
(A1's hook) and regenerated `proto/rogue.proto` and `ui/core/talents/trees/rogue.json` from it.


Worker branch `round-70753-rogue`. Phase A ledger and notes follow.

## Sources

- Blizzard's notes: `orchestration/audit/ref/client_update/sources/devnotes_post9_2026-11-02.txt`, "- Rogue" (tier 4).
- Client 1.60.1.70753 (wago.tools, fetched 2026-11-03; tier 2). **It has none of the rogue rework.** Shank, Blood
  Splatter, Composure, Flash of Steel, Boot Knife and Quick Getaway don't exist, and every changed rogue row still has
  its old value: Flawless Execution 1310711 bp −10, Aggression 18427 2% a rank with Backstab in its mask, Lethality
  14128, Seal Fate 14186, Relentless Strikes 14179 (1 rank), Dual Wield Specialization 13715, Opportunity 14057,
  Mutilate 1310705/6 at 75%, Hemorrhage 16511 masking Rupture only, Venom 1310703 still a finisher. What 70753 does add
  for the rogue is Poised (1328825/1328826) and Death Mark's four ranks (1316709, 1325756-8, since 70425).
- Wowhead Forever (dataEnv 16 tooltip API, 2026-11-03; tier 3): the old values too (Flawless Execution 10, Lethality
  6%, Aggression 2% with Backstab, Venom a finisher).
- foreverchanges.pro (tier 5): its November 3 patch notes page restates Blizzard's notes; its "Build 1.60.1.70753"
  entry lists only Poised for the rogue; its November 3 hotfix (launch realm cache) has only hunter rows.
- No server hotfix touches the rogue yet.

So for the rework the notes are the only source. The higher tiers don't disagree with them, they predate them: the
client and Wowhead carry the beta's values, and the notes describe what ships on November 4 (as with the hunter
lines, which a launch-realm hotfix confirmed on November 3). The ledger takes the notes where the client simply
hasn't caught up, and says so per row. Where a value is missing from the notes too, the row says what was assumed.
When a client build or hotfix brings these spells, every "notes only" row needs checking against it.

**The rogue talent tree is not reworked in 70753 either**: TraitTree 1111 still has its 53 nodes, with Murder,
Quietus, Cutthroat, Thousand Cuts and Improved Distract and none of the new talents. So A1's client port can't give
the rogue its new tree; it has to be built from the notes (rows and rank counts as the notes give them).

## Notes-only values (standards 37)

All gated under Forever, citing the notes and "client 1.60.1.70753 unchanged". Replace each when a client build or
hotfix brings it: Shank (whole spell, ID 1399248 is the sim's), Venom as a Shank passive (30% / 20% / 6 s), Deadly
Poison 0.009 and Instant Poison 0.01 AP, Mutilate 90%, Flawless Execution 15, Aggression 5/10/15% (no Backstab),
Ghostly Strike's 7 s opening, Hemorrhage on Garrote. Phase B adds Lethality, Blood Splatter, Relentless Strikes, Seal
Fate, Coup de Grace on Death Mark, Efficient Strikes, Composure, Dual Wield Specialization, Flash of Steel, Opportunity,
Boot Knife and the removals.

## Ledger

Status: landed / missing / wrong / partial / n/a. "Phase B" rows need A1's new tree (new proto fields or rank
counts) and are done after "sources landed".

| Note line | Tier and evidence | Sim status (before) | Fix |
|---|---|---|---|
| Slice and Dice no longer stands a sitting target up | notes | n/a | none (no sitting boss) |
| **Poised** (level 18): out of combat at max Energy, next Energy ability −10 | client 70753: 1328825 passive, 1328826 aura 107 misc 14 bp −10, ProcCharges 1, DurationIndex 21, class mask 255727518 / 2099456 / 67108864 (every Energy ability the sim has but Death Mark) | missing (finding client_update-20261103-1) | landed: `sim/rogue/poised.go`. Up at the prepull (Energy starts full); back at the pull if a prepull ability spent it and the Energy is full again at 0. Spent by the first covered cast. Shank (not in the client) is taken to be covered. Composure's +5/+10 hooks in at `poisedReduction` (Phase B). Tests `TestForeverPoisedFirstAbility`, `TestForeverPoisedAtThePull` |
| **Shank** (24, rank 2 at 44): 5 Energy, 10 s CD, off-hand 50% weapon damage, +15 / +25 Energy, can't miss | notes only (no client spell) | missing | landed: `sim/rogue/shank.go`, ID 1399248 is ours. Assumed: 50% of the off hand's *normalized* weapon damage after the usual off-hand halving (and Dual Wield Specialization), dodge/parry/crit possible, Energy only when it lands (80% refund otherwise), no combo point. Rotation: `shankEnergy` knob in `tools/rotopt/rogue_common.go`. Test `TestForeverShank` |
| Deadly Poison AP coefficient 0.009 (was 0.0045) | notes (client has no coefficient) | wrong (0.0045) | landed: `sim/rogue/poisons.go` |
| Instant Poison AP coefficient 0.01 (was 0.005) | notes | wrong (0.005) | landed; `TestForeverInstantPoisonAttackPower` updated |
| Ruthlessness doesn't stand the target up | notes | n/a | none |
| Improved Expose Armor to row 2 | notes | n/a (standards 18: Sunder, never Expose Armor) | tree only (A1) |
| Murder removed | notes | landed (Phase B) | the tree from the notes (`tools/forever_talents/notes/rogue.json`) and the sim on its fields; tests in `devnotes_70753_test.go` |
| Lethality 2 points, 10/20% crit damage (was 5 at 4% a rank) | notes; client 14128 still bp 6 / 5 ranks | landed (Phase B) | the tree from the notes (`tools/forever_talents/notes/rogue.json`) and the sim on its fields; tests in `devnotes_70753_test.go` |
| Vigor to row 3 | notes | landed (value unchanged) | tree only |
| Blood Splatter (row 3, 1 pt): Rupture +6 s, Garrote +9 s | notes only | landed (Phase B) | Rupture 3 more ticks, Garrote 3 more ticks (`TestForeverBloodSplatter`) |
| Relentless Strikes row 5, 5 points 4-20% a combo point for 25 Energy | notes; client 14179 still 1 rank, 20%/CP | landed (Phase B) | the tree from the notes (`tools/forever_talents/notes/rogue.json`) and the sim on its fields; tests in `devnotes_70753_test.go` |
| **Venom** row 5, now a passive: Shank → +30% poison damage, +20% poison chance for 6 s | notes; client 1310703 still the finisher (30%, 10%) | wrong (finisher) | landed: `sim/rogue/venom.go`, the finisher is gone, Shank activates the 6 s aura. Rotation knobs `venomCp`/`venomRefresh` removed. Test `TestForeverVenomFromShank` |
| Seal Fate 3 points 33/66/100% | notes; client 14186 unchanged | landed (Phase B) | the tree from the notes (`tools/forever_talents/notes/rogue.json`) and the sim on its fields; tests in `devnotes_70753_test.go` |
| Mutilate to row 7; 90% weapon damage plus a larger flat amount (was 75% + lower) | notes; client still 75% | wrong (75%) | landed: `sim/rogue/mutilate.go`. The client's flat (23/33/48/67) is added to the weapon damage *before* the percent, so 90% takes it up with it (67 × 0.9 = 60.3 at 60, was 50.25), which is what the notes' "an additional amount that was also increased" reads as. Test `TestForeverMutilateNinetyPercent` |
| Improved Eviscerate → Coup de Grace: Eviscerate 7/13/20%, Death Mark damage and storage +7/13/20% | notes | landed for Eviscerate; Death Mark Phase B | Death Mark part with Death Mark |
| Improved Sinister Strike → Efficient Strikes: −3/−5 Sinister Strike and Hemorrhage | notes | landed (Phase B) | Efficient Strikes on Hemorrhage too | the tree from the notes (`tools/forever_talents/notes/rogue.json`) and the sim on its fields; tests in `devnotes_70753_test.go` |
| Improved Sprint to row 2 | notes | n/a | tree only |
| Composure (row 3, 2 pts): Poised +5/+10 | notes only | landed (Phase B) | the tree from the notes (`tools/forever_talents/notes/rogue.json`) and the sim on its fields; tests in `devnotes_70753_test.go` |
| Flawless Execution −15 Eviscerate Energy (was −10) | notes; client 1310711 still −10 | wrong | landed: `sim/rogue/eviscerate.go`; `TestForeverFlawlessExecutionAndAggression` |
| Weapon Expertise, Aggression to row 4 | notes | n/a | tree only |
| Aggression 5/10/15% Sinister Strike and Eviscerate (was 2/4/6% with Backstab) | notes; client 18427 still 2%, Backstab in mask | wrong | landed: `aggressionMultiplier` (talents.go), off Backstab. Same test |
| Dual Wield Specialization row 5, 2 points 25/50% (was 5 at 5%) | notes | landed (Phase B) | the tree from the notes (`tools/forever_talents/notes/rogue.json`) and the sim on its fields; tests in `devnotes_70753_test.go` |
| Flash of Steel (row 6, 5 pts): Shank +3-15 Energy, −0.4-2.0 s CD | notes only | landed (Phase B) | the tree from the notes (`tools/forever_talents/notes/rogue.json`) and the sim on its fields; tests in `devnotes_70753_test.go` |
| Hemorrhage and Ghostly Strike "much more damage than intended" with a dagger, fixed | notes, no numbers; client still Hemorrhage 145% dagger / 100%, Ghostly Strike 180% / 125% | landed as far as can be told | The sim already uses the tooltip's dagger values (1.45, 1.8), which were the intent; the in-game bug dealt more than these. Kept, open in notes.md until a log or client build gives numbers |
| Setup row 1, 15/30/45% | notes | n/a (procs on the rogue's dodges; the boss is on the tank) | none |
| Opportunity 5 points 4-20% Backstab, Garrote, Ambush, Mutilate (was 2 at 5/10%) | notes; client 14057 still 2 ranks | landed (Phase B) | the tree from the notes (`tools/forever_talents/notes/rogue.json`) and the sim on its fields; tests in `devnotes_70753_test.go` |
| Dirty Deeds, Camouflage, Premeditation, Serrated Blades, Elusiveness, Ghostly Strike, Heightened Senses moves | notes | n/a | tree only |
| Camouflage 3 points | notes | n/a (stealth speed) | none |
| **Ghostly Strike**: next Ambush or Garrote within 7 s needs no Stealth | notes only | missing | landed: `GhostlyStrikeOpenerAura` (ghostly_strike.go), read by Ambush and Garrote, spent by either; ID 14278 tag 1 is ours. Rotation: `ghostlyAmbush` line. Test `TestForeverGhostlyStrikeOpener` |
| Improved Distract removed; Quick Getaway | notes | n/a | none |
| **Hemorrhage** raises Rupture and Garrote 15% (was Rupture only) | notes; client 16511 mask still Rupture only | missing for Garrote | landed: `hemorrhageMultiplier` on both. Test `TestForeverHemorrhageGarrote` |
| Quietus, Cutthroat, Thousand Cuts removed | notes | landed (Phase B) | the tree from the notes (`tools/forever_talents/notes/rogue.json`) and the sim on its fields; tests in `devnotes_70753_test.go` |
| Boot Knife (row 6, 5 pts): Shank 20-100% to add 2 combo points | notes only | landed (Phase B) | the tree from the notes (`tools/forever_talents/notes/rogue.json`) and the sim on its fields; tests in `devnotes_70753_test.go` |
| **Death Mark** (row 7, 31 pt, Curse finisher, 25 Energy, 20 yd, 10 s CD): stores a share of the rogue's damage per combo point (Garrote and Rupture twice), deals it after 6 s, capped | client 70753 (since 70425): 1316709 / 1325756-8 (levels 40/48/56/60), SpellPower 25 Energy + combo points, SpellCooldowns 10 s, GCD category 133 at 1 s, aura 3 period 6000 + aura 4 proc on damage taken (ProcTypeMask 0x800AAAA8: melee, ranged, spells, periodic), physical. SpellDescriptionVariables 1021: `$ratio=12` (12% a combo point), `$maxAP=0.04`; rank 4 cap 156 + 245 a point + 4% AP a point, "before critical strike" | landed (Phase B) | `sim/rogue/death_mark.go` from the client values (`TestForeverDeathMark`) |
| Racials (Dwarf Artisan's Precision, Human Ruthless 1%, Orc Savage removed and Blood Fury 2 min, Tauren Endurance, Troll Berserking 15 s, Undead Touch of the Grave 15 s ICD) and race base stats | notes, client | A2 | shared; the rogue race sweep in Phase B picks them up |
| "Various talents no longer have a penalty ... low rank" | notes | n/a for rogues (A2) | none |
| Throw's 0.5 s windup | notes | n/a (no rotation throws) | A2 |
| Roots from procs get diminishing returns | notes | n/a (PvP) | none |
| Threat and healing rows | — | n/a | tank/healer sessions |

## Phase A DPS effect (mechanics only)

The 12 report requests as committed (their talents and rotations unchanged, so no Shank yet), 180 s, 3000 iterations,
seed 1, master 3462697 vs the branch: Poised, the poison coefficients, Mutilate 90%, Flawless Execution 15, Aggression
5/10/15% (off Backstab) and Hemorrhage on Garrote.

| Build | master | Phase A | Δ |
|---|---|---|---|
| combat-swords | 782.1 | 816.7 | +34.6 |
| combat-swords-humanoid | 808.6 | 844.2 | +35.6 |
| combat-daggers | 791.5 | 798.9 | +7.4 |
| combat-daggers-humanoid | 815.2 | 823.2 | +8.0 |
| mutilate | 743.7 | 790.8 | +47.1 |
| mutilate-humanoid | 769.0 | 817.0 | +48.0 |
| seal-fate-daggers | 752.5 | 770.4 | +17.9 |
| seal-fate-daggers-humanoid | 778.6 | 796.6 | +18.0 |
| seal-fate-swords | 747.7 | 766.3 | +18.6 |
| seal-fate-swords-humanoid | 773.6 | 792.8 | +19.2 |
| subtlety | 733.2 | 740.8 | +7.6 |
| subtlety-humanoid | 758.7 | 766.7 | +8.0 |

Standard errors about 0.6. The rogue goldens move with this; they are regenerated in Phase B with the new trees.

## Builds: a recommendation for the user (to settle after Phase B's numbers)

Shank, Boot Knife, Flash of Steel and Venom on Shank make an off-hand Shank a core part of every spec, and Death Mark
is a new 31-point Subtlety capstone. Phase B sims a Death Mark Subtlety build and a Boot Knife Shank build, and this
section will say whether either should replace one of the current twelve (no build is removed without a ruling).
