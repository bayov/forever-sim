# Rogue: round 70753 (November 2 development notes)

Worker branch `round-70753-rogue`. Phase A (mechanics that don't need the new trees) is done; Phase B (the new
talent tree, re-optimization, races, audit) waits for A1's tree port. The final summary goes here when Phase B lands.

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
| Murder removed | notes | Phase B | remove `applyMurder` once the proto field goes |
| Lethality 2 points, 10/20% crit damage (was 5 at 4% a rank) | notes; client 14128 still bp 6 / 5 ranks | Phase B | `lethality()` to 0.10 a rank on the new 2-rank field |
| Vigor to row 3 | notes | landed (value unchanged) | tree only |
| Blood Splatter (row 3, 1 pt): Rupture +6 s, Garrote +9 s | notes only | Phase B (new talent) | to do: Rupture 3 more ticks (2 s ticks), Garrote 3 more ticks (3 s ticks) |
| Relentless Strikes row 5, 5 points 4-20% a combo point for 25 Energy | notes; client 14179 still 1 rank, 20%/CP | Phase B (bool → int32) | to do |
| **Venom** row 5, now a passive: Shank → +30% poison damage, +20% poison chance for 6 s | notes; client 1310703 still the finisher (30%, 10%) | wrong (finisher) | landed: `sim/rogue/venom.go`, the finisher is gone, Shank activates the 6 s aura. Rotation knobs `venomCp`/`venomRefresh` removed. Test `TestForeverVenomFromShank` |
| Seal Fate 3 points 33/66/100% | notes; client 14186 unchanged | Phase B (rank count) | to do |
| Mutilate to row 7; 90% weapon damage plus a larger flat amount (was 75% + lower) | notes; client still 75% | wrong (75%) | landed: `sim/rogue/mutilate.go`. The client's flat (23/33/48/67) is added to the weapon damage *before* the percent, so 90% takes it up with it (67 × 0.9 = 60.3 at 60, was 50.25), which is what the notes' "an additional amount that was also increased" reads as. Test `TestForeverMutilateNinetyPercent` |
| Improved Eviscerate → Coup de Grace: Eviscerate 7/13/20%, Death Mark damage and storage +7/13/20% | notes | landed for Eviscerate; Death Mark Phase B | Death Mark part with Death Mark |
| Improved Sinister Strike → Efficient Strikes: −3/−5 Sinister Strike and Hemorrhage | notes | partial (Sinister Strike only) | Phase B (field rename): Hemorrhage cost too |
| Improved Sprint to row 2 | notes | n/a | tree only |
| Composure (row 3, 2 pts): Poised +5/+10 | notes only | Phase B | `poisedReduction` |
| Flawless Execution −15 Eviscerate Energy (was −10) | notes; client 1310711 still −10 | wrong | landed: `sim/rogue/eviscerate.go`; `TestForeverFlawlessExecutionAndAggression` |
| Weapon Expertise, Aggression to row 4 | notes | n/a | tree only |
| Aggression 5/10/15% Sinister Strike and Eviscerate (was 2/4/6% with Backstab) | notes; client 18427 still 2%, Backstab in mask | wrong | landed: `aggressionMultiplier` (talents.go), off Backstab. Same test |
| Dual Wield Specialization row 5, 2 points 25/50% (was 5 at 5%) | notes | Phase B (rank count) | `dwsMultiplier` |
| Flash of Steel (row 6, 5 pts): Shank +3-15 Energy, −0.4-2.0 s CD | notes only | Phase B | Shank's gain and cooldown |
| Hemorrhage and Ghostly Strike "much more damage than intended" with a dagger, fixed | notes, no numbers; client still Hemorrhage 145% dagger / 100%, Ghostly Strike 180% / 125% | landed as far as can be told | The sim already uses the tooltip's dagger values (1.45, 1.8), which were the intent; the in-game bug dealt more than these. Kept, open in notes.md until a log or client build gives numbers |
| Setup row 1, 15/30/45% | notes | n/a (procs on the rogue's dodges; the boss is on the tank) | none |
| Opportunity 5 points 4-20% Backstab, Garrote, Ambush, Mutilate (was 2 at 5/10%) | notes; client 14057 still 2 ranks | Phase B (rank count) | to do (Mutilate already covered) |
| Dirty Deeds, Camouflage, Premeditation, Serrated Blades, Elusiveness, Ghostly Strike, Heightened Senses moves | notes | n/a | tree only |
| Camouflage 3 points | notes | n/a (stealth speed) | none |
| **Ghostly Strike**: next Ambush or Garrote within 7 s needs no Stealth | notes only | missing | landed: `GhostlyStrikeOpenerAura` (ghostly_strike.go), read by Ambush and Garrote, spent by either; ID 14278 tag 1 is ours. Rotation: `ghostlyAmbush` line. Test `TestForeverGhostlyStrikeOpener` |
| Improved Distract removed; Quick Getaway | notes | n/a | none |
| **Hemorrhage** raises Rupture and Garrote 15% (was Rupture only) | notes; client 16511 mask still Rupture only | missing for Garrote | landed: `hemorrhageMultiplier` on both. Test `TestForeverHemorrhageGarrote` |
| Quietus, Cutthroat, Thousand Cuts removed | notes | Phase B | remove code and template lines with the proto fields |
| Boot Knife (row 6, 5 pts): Shank 20-100% to add 2 combo points | notes only | Phase B | to do |
| **Death Mark** (row 7, 31 pt, Curse finisher, 25 Energy, 20 yd, 10 s CD): stores a share of the rogue's damage per combo point (Garrote and Rupture twice), deals it after 6 s, capped | client 70753 (since 70425): 1316709 / 1325756-8 (levels 40/48/56/60), SpellPower 25 Energy + combo points, SpellCooldowns 10 s, GCD category 133 at 1 s, aura 3 period 6000 + aura 4 proc on damage taken (ProcTypeMask 0x800AAAA8: melee, ranged, spells, periodic), physical. SpellDescriptionVariables 1021: `$ratio=12` (12% a combo point), `$maxAP=0.04`; rank 4 cap 156 + 245 a point + 4% AP a point, "before critical strike" | Phase B (talent) | to do: model from the client (above) |
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
