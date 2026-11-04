# Hotfix push 113539 (warrior: Reckless Swings, Improved Whirlwind, Spearing Strike)

Source: wago.tools/hotfixes, product wowf, search "enus 113539 ", cached at
`/root/.cache/wow-data/wago/hotfixes/wowf/search/enus_113539.json` (fetched 2026-11-04 10:17Z, 848 of 848 rows, none
missing). The push is logged four times, 212 rows each: region 91 build 70707 at 2026-11-03 18:29:17, region 93 build
70707 at 17:34:20, region 93 build 70755 at 20:06:19 and region 91 build 70755 at 2026-11-04 00:14:20. The four copies
are identical (same records, status and data). Note that the 00:14:20 time belongs to the region 91 / 70755 copy; the
region 91 / 70707 copy was logged at 18:29:17, the same second as the rogue push 113483 and the paladin push 113542.

Old values come from the 70707 base tables and the 70755 client tables (`tools/sources/wago.py`; 70755's
SpellPowerDifficulty can't be downloaded, HTTP 400). The two builds agree on every row below. None of the new records
(Reckless Swings, the new curves, node 113636) is in either client. Decoded with
`hotfixes.py rows 1.60.1.70707 <Table> --region 91`.

Status 1 is data, 2 is a deleted record, 3 is an invalid record. 198 of the 212 rows are warrior. The other 14 are four
Spell / SpellName pairs re-sent with no change (19070 Heavy Scorpid Belt, 19075 Heavy Scorpid Leggings, 19100 Heavy
Scorpid Shoulders, 19226 Zergling) and the six invalid PlayerCondition / SpellScript rows at the end (no data, so no
owner); no other class appears.

| Table | Rows (status) |
|---|---|
| CooldownSetSpell | 56 (1) |
| CurvePoint | 28 (1), 13 (2) |
| TraitNodeGroupXTraitNode | 10 (1), 6 (2) |
| SpellCooldowns | 11 |
| Curve | 11 |
| Spell, SpellName | 10 each |
| SpellEffect | 9 |
| SpellMisc | 5 |
| TraitNode, TraitNodeEntry | 5 each |
| TraitDefinitionEffectPoints | 4 (1), 2 (2) |
| TraitEdge | 2 (1), 2 (2) |
| SpellScript | 5 (3) |
| SpellClassOptions, SpellLabel, SpellLevels, SpellPower, SpellXSpellVisual | 2 each |
| SpellPowerDifficulty | 2 (2) |
| SpellAuraOptions, SpellCategories, SkillLineAbility, TraitDefinition, TraitNodeXTraitNodeEntry | 1 each |
| PlayerCondition | 1 (3) |

## SpellEffect

| Spell | Effect row | Field | Old | New |
|---|---|---|---|---|
| 1310222 Spearing Strike | 1339573 eff1 (Effect 31, weapon %) | base points | 40 | **75** |
| 1310222 Spearing Strike | 1339768 eff2 (Effect 3, dummy: the Giants/Dragonkin multiplier) | base points | 2 | **1** |
| 12862 Improved Slam | 1339521 eff2 (aura 107 misc 11 cooldown, mask0 2097152 Slam + Concussion Blow) | base points | -3000 | -5000 (now overridden by curve 128144, below) |
| 1310315 Raging Blows -> Improved Whirlwind | 1339744 eff0 | aura / misc / mask1 / points | 4 dummy / 0 / 0 / 100 | **107 misc 11 (cooldown) on mask1 4 (Whirlwind)**, -2000 (curve 128139: -1000 / -2000) |
| 1310315 | 1345686 eff1 | aura / misc / masks / points | 107 misc 14 (cost) on mask0 4194304 + mask1 4 (Cleave + Whirlwind), -30 | **4 dummy**, no mask, 30 (curve 128138: 15 / 30) |
| 12321 Booming Voice | 1360666 eff1 (aura 108 misc 14, cost %) | EffectSpellClassMask_1 | 0 | 32 (adds Piercing Howl); mask0 983040 (the Shouts) and -25 unchanged |
| 1329554 Reckless Swings (Rage) | 1369166 eff0 (new) | Effect 30 energize, misc 1 (Rage), target 1 (caster) | none | 50 (5 Rage) |
| 1329554 | 1369168 eff1 (new) | Effect 3 dummy | none | 1 (the tooltip's divisor) |
| 1329555 Reckless Swings (cast and buff) | 1369167 eff0 (new) | aura 42 proc trigger spell, target caster | none | triggers 1329554, base 0 (curve 128140 gives 50) |

Spearing Strike's eff0 (1339572, Effect 121 normalized weapon damage, 0 points) is not in the push and stays. Its
tooltip, also not in the push, reads "deals $s2% weapon damage. Deals an additional ${$s2*$s3}% weapon damage against
Giants, Dragonkin, and mounted targets": 40% and 40x2 = 80% more before, **75% and 75x1 = 75% more** now (150% in all
against Giants and Dragonkin, was 120%).

Improved Slam's other effects (eff0 casting time and eff1 global cooldown, -250 / -500 by curves 109365 / 109364, and
the eff3-7 rank overrides) are not in the push. Its masks also cover Concussion Blow 12809 (mask0 2097152 is shared),
before and after.

## SpellMisc

| Spell | Field | Old | New |
|---|---|---|---|
| 12292 Sweeping Strikes | Attributes_14 | 0 | 64 |
| 12809 Concussion Blow | Attributes_14 | 0 | 64 |
| 1310315 Improved Whirlwind | all | | re-sent, no change (passive) |
| 1329554 Reckless Swings (Rage) | new | none | Attributes_0 0x50000, Attributes_15 8192, instant, range 1 (self), Physical |
| 1329555 Reckless Swings | new | none | Attributes_0 0x0C040010, Attributes_1 0x18000, Attributes_14 64, Attributes_15 8192, instant, **DurationIndex 31 (8000 ms)**, range 1 (self), Physical, icon 132334 |

Attributes_14 64 is the flag 70753 gave 1374 spells with no DPS reading (`update_ledger.md`); 70755 still lacked it on
these two. Read with TrinityCore's names (not checked for Forever): 1329555's Attributes_0 bits 4, 18, 26, 27 are
ability, don't sheathe, aura is a debuff (can't be clicked off) and usable while sitting; Attributes_1 bits 15 and 16
are immunity flags. 1329554's bits 16 and 18 are not in shapeshift and don't sheathe.

## Spell and SpellName

| Record | Old | New |
|---|---|---|
| SpellName 1310315 | Raging Blows | **Improved Whirlwind** |
| Spell 1310315 Description | "Reduces the Rage cost of your Cleave and Whirlwind abilities by ${$m2/-10}." | "Reduces the cooldown of your Whirlwind ability by ${$m1/-1000} sec and increases the damage it deals to the closest enemy by $s2%." |
| Spell 12862 Improved Slam Description | "Reduces the global cooldown and cast time of your Slam ability by ${$m1/-1000}.2 sec. In addition, Slam no longer interrupts or delays your melee swing and Slam's cooldown is reduced by ${$m3/-1000}.1 sec." | "Reduces the cooldown of your Slam ability by ${$m3/-1000}.1 sec, the cast time and global cooldown by ${$m1/-1000}.2 sec, and Slam no longer interrupts or delays your melee swing." (reworded, same values) |
| Spell 12321 Booming Voice Description | "... reduces their Rage cost by $s2%." | "... reduces the Rage cost by of your Shouts and Howls by $s2%." (sic) |
| SpellName / Spell 1329554 (new) | none | "Reckless Swings": "${$s1/10} Rage generated." |
| SpellName / Spell 1329555 (new) | none | "Reckless Swings": "Your next $n melee attacks generate ${$1329554s1/10/$1329554s2} additional Rage. Lasts $d." Aura: "Melee attacks grant ${$1329554s1/10/$1329554s2} additional Rage." |
| Spell / SpellName 12834 Deep Wounds | | re-sent, no change |
| SpellName 12321, 12862 | | re-sent, no change |

## SpellCooldowns, SpellCategories, CooldownSetSpell

| Record | Spell | Field | Old | New |
|---|---|---|---|---|
| SpellCooldowns 53310, 54250, 54482, 54498, 94698, 100998-101002 | Slam, every rank: 1464, 8820, 11604, 11605, 1240193, 1310196-1310200 | CategoryRecoveryTime | 18000 | **15000** |
| SpellCooldowns 102442 (new) | 1329555 Reckless Swings | RecoveryTime / StartRecoveryTime | none | **15000 / 1000** |
| SpellCategories 191006 (new) | 1329555 | StartRecoveryCategory | none | **133** (the global cooldown); Category 0, DefenseType 0 |

Slam's StartRecoveryTime (1500) is unchanged.

CooldownSetSpell: 56 rows, all in set 5236 (the warrior's, 70 rows before). Two are new, both for Reckless Swings
1329555: 200558 at OrderIndex 16, Flags 0 (ahead of Charge) and 200557 at OrderIndex 68, Flags 2 (after Sunder Armor,
before Battle Shout). The other 54 rows (199558-200450: Charge, Intercept, Pummel, Shield Bash, Shield Block,
Concussion Blow, Disarm, Intimidating Shout, Taunt, Mocking Blow, Challenging Shout, Victory Rush, Berserker Rage,
Bloodrage, Last Stand, Retaliation, Shield Wall, Recklessness, Victorious, Deep Wounds, Bloodthrill, Sweeping Strikes,
Improved Hamstring, Piercing Howl, Blood Craze, Enrage, Flurry, Death Wish, Improved Shield Bash, Rend, Hamstring,
Thunder Clap, Mortal Strike, Demoralizing Shout, Bloodthirst, Sunder Armor, Battle Shout, Stances) only move their
OrderIndex, up by one (old 16-66 to 17-67) or two (old 67-69 to 69-71), to make room. No other field changes. This is the
cooldown display list; it has no effect on play.

## SpellPower and SpellPowerDifficulty

| Record | Spell | Old | New |
|---|---|---|---|
| SpellPower 315180 (new) | 1329555 Reckless Swings | none | **PowerType -2 (health), PowerCostPct 25** (25% of base health) |
| SpellPower 315179 (new) | 1329554 (the Rage energize) | none | PowerType -2, every cost 0 |
| SpellPowerDifficulty 315179, 315180 | | | deleted (status 2) |

## SpellAuraOptions, SpellClassOptions, SpellLevels, SkillLineAbility, SpellLabel, SpellXSpellVisual

| Record | Spell | New |
|---|---|---|
| SpellAuraOptions 249144 | 1329555 | **ProcChance 100, ProcCharges 4, ProcTypeMask_0 4** (melee auto attack), no PPM, no proc cooldown |
| SpellClassOptions 85341 | 1329555 | SpellClassSet 4 (warrior), SpellClassMask_1 131072 (bit 17, shared with Berserker Rage, Bloodrage, Death Wish, Enrage) |
| SpellClassOptions 85340 | 1329554 | SpellClassSet 4, no mask |
| SpellLevels 154307, 154308 | 1329554, 1329555 | all 0 |
| SkillLineAbility 59905 | 1329555 | SkillLine 256 (Fury) |
| SpellLabel 216816, 216817 | 1329554, 1329555 | label 25 |
| SpellXSpellVisual 540846, 540847 | 1329554, 1329555 | visual 0 / 47 |

ProcTypeMask 4 is the same proc type as Flurry's buff 12966 (3 charges), Unbridled Wrath 12322 and Windfury Totem. The
DB2 has no hit mask (that is server side), so it does not say whether a miss or dodge uses a charge.

## TraitDefinition, TraitDefinitionEffectPoints, Curve, CurvePoint

The curve value at rank N is the total for that rank and replaces the effect's base points.

| Talent (definition, spell) | Curve | Old ranks | New ranks |
|---|---|---|---|
| Flurry (135459, 12319) | 83289 | 5/10/15/20/25 | **10/20/30** (241862-3 changed, 336293-5 deleted, 372098 added; 372099 deleted, never in a client) |
| Two-Handed Weapon Specialization (135479, 12163) | 83308 | 1/2/3 | **2/4/6** |
| Improved Tactical Mastery (135485, 12295) | 83313 | 3/6/9/12/15 | **5/10/15** (241955-7 deleted, 372110 added) |
| Bloodthrill (141974, 1289682) | 109366 | 4/8/12/16/20 | **10/20** (336296-8 deleted) |
| Enrage (135462, 12317) | 83294 | 2/4/6/8/10 | 2/4/6/8/10, points 3-5 re-keyed (241883-5 deleted, 372095-7 added, same values) |
| Improved Whirlwind (135508, 1310315; was Raging Blows) | 128139 (new), TDEP 25229 eff0 | none | **-1000 / -2000 ms** Whirlwind cooldown |
| Improved Whirlwind | 128138 (new), TDEP 25230 eff1 | none | **15 / 30** (closest enemy %) |
| Reckless Swings (145966 new, 1329555) | 128140 (new), TDEP 25231 eff0 | none | 50 / 50 |
| Improved Slam (142604, 12862) | 128144 (new), TDEP 25234 eff2 | -3000 flat (no curve) | **-2500 / -5000 ms** Slam cooldown |
| TDEP 25232, 25233 | | | deleted (never in a client) |

Curves 128142 (-300 / -500) and 128143 (-30 / -50) are new but no definition uses them (probably the deleted TDEP
25232 / 25233). The Curve rows 83289, 83294, 83308, 83313 and 109366 are re-sent unchanged.

## TraitNodeEntry, TraitNode, TraitEdge, groups (tree 1117, Fury)

| Entry | Talent | MaxRanks old | New |
|---|---|---|---|
| 130658 | Flurry | 5 | 3 |
| 130684 | Improved Tactical Mastery | 5 | 3 |
| 130707 | Raging Blows -> Improved Whirlwind (def 135508) | 1 | 2 |
| 137214 | Bloodthrill | 5 | 2 |
| 141298 (new) | Reckless Swings (def 145966) | none | 1, in new node 113636 (TraitNodeXTraitNodeEntry 139984) |

Rows and columns are 1-based, Fury columns X 5020 / 5620 / 6220 / 6820, rows Y 2130 + 600 a row.

| Node | Talent | Old | New |
|---|---|---|---|
| 113636 (new) | Reckless Swings | none | X5620 Y3930: row 4 col 2, Flags 8 |
| 105977 | Improved Whirlwind (was Raging Blows) | row 4 col 2 | X5020 Y5130: row 6 col 1 |
| 105928 | Flurry | row 6 col 2 | X6220: row 6 col 3 |
| 105931 | Enrage | row 4 col 3 | X6820 Y3330: row 3 col 4 |
| 113569 | Gore Drinker | row 6 col 3 | Y4530: row 5 col 3 |

| Edge | Change |
|---|---|
| 136735 Enrage -> Gore Drinker | removed |
| 136736 Death Wish -> Flurry | removed |
| 136777 Death Wish -> Bloodthirst | added |
| 136778 Blood Craze -> Gore Drinker | added |

TraitNodeGroupXTraitNode: 6 deleted (253211, 253220, 268088, 269272, 269287, 269681) and 10 added (269865-269874),
moving nodes 105931, 105977, 113569 and the new 113636 between groups 11657, 11659-11662 and 12826-12828.

Against `tools/forever_talents/data/warrior.json` (0-based there): Enrage (2,3), Gore Drinker (4,2), Death Wish,
Bloodthirst and both edges already agree. The sim still has **Raging Blows** at (3,1), where the client now has
Reckless Swings; it has Reckless Swings at (3,2) (client (3,1)), Flurry at col 1 (client col 2) and Improved Whirlwind
at col 2 (client col 0). Both Fury presets (`30313113002-0505300551120100321`, `30313113032-2505001501122100321`) spend
a point in Raging Blows.

## Invalid rows

PlayerCondition 168228 and SpellScript 836, 4653, 100117, 106154, 110057 are status 3. None of them is in the 70707 or
70755 client.

## What it means for the sim

| Spell | Hotfix rows | Before (client 70707/70755 / sim) | After | Sim fix needed |
|---|---|---|---|---|
| Spearing Strike 1310222 | SpellEffect 1339573 eff1, 1339768 eff2 | 40% normalized, +40x2 = 80% vs Giants, Dragonkin (and mounted) / sim 0.4 and +0.8 | 75% normalized, +75x1 = 75% more (150% in all) | **yes**: `sim/warrior/spearing_strike.go`, `weaponPercent := 0.75`, `+= 0.75` against Giants and Dragonkin (was 0.4 / 0.8); cite push 113539, SpellEffect 1339573 (eff1 75) and 1339768 (eff2 1). Normalization stays (eff0 Effect 121). The Dragonkin preset uses it; re-sim it |
| Raging Blows 1310315 | SpellName, Spell, SpellEffect 1339744 / 1345686, entry 130707, node 105977 | -3 Rage on Cleave and Whirlwind / sim `Talents.RagingBlows` takes 3 off Cleave (`heroic_strike_cleave.go:57`) and Whirlwind (`whirlwind.go:35`) | gone: the spell and its node are Improved Whirlwind | **yes**: drop the Raging Blows discount under Forever (Cleave 20, Whirlwind 25), remove the talent from the Forever tree, and re-spend the Fury presets' Raging Blows point |
| Improved Whirlwind 1310315 | SpellEffect 1339744 / 1345686, curves 128139 / 128138, entry 130707 (2 ranks), node 105977 | not in any client / sim (notes) 10 - points sec, +15% a point on the first target's hits | -1 / -2 s Whirlwind cooldown (spell mod), +15 / 30% to the closest enemy (a dummy: server script) | no for values: the sim matches. Cite 1310315 eff0/eff1, curves 128139/128138 in `whirlwind.go`. Tree position: row 6 col 1 (0-based (5,0)) |
| Reckless Swings 1329555 / 1329554 | see the section below | not in any client / sim placeholder 1399001, 15 s, off GCD, 4 landed white or Heroic Strike swings, flat 5 Rage, 8 s, no health cost | 15 s, **1 s GCD**, 25% of base health, 4 charges on melee auto attacks, 5 Rage each, 8 s | **yes**: id 1399001 -> 1329555 (buff and cast) and 1329554 (the Rage, for the metrics); add `DefaultCast: core.Cast{GCD: time.Second}` (StartRecoveryTime 1000, category 133); comments cite the push. Re-run the rotation search (it was tuned for an off-GCD cast). Tree position: row 4 col 2 (0-based (3,1)) |
| Slam (all ranks) | SpellCooldowns 53310 ... 101002 | 18 s / sim 15 s (notes) | 15 s | no, the sim matches; cite push 113539 in `slam.go` (it still says notes) |
| Improved Slam 12862 | SpellEffect 1339521, curve 128144, TDEP 25234, Spell text | -3 s flat at both ranks / sim 2.5 / 5 s (notes) | -2.5 / -5 s | no, sim matches; cite. Side note: the same mask takes the time off Concussion Blow (not simmed) |
| Flurry 12319 | curve 83289, entry 130658 | 5 ranks of 5% / sim 3 ranks, 10/20/30% (notes) | 3 ranks, 10/20/30% | no, sim matches (`makeFlurryAura`); cite. Tree: col 3 (0-based 2), the sim has col 1 |
| Two-Handed Weapon Specialization 12163 | curve 83308 | 1/2/3% / sim 2/4/6% | 2/4/6% | no, sim matches; cite |
| Improved Tactical Mastery 12295 | curve 83313, entry 130684 | 5 ranks of 3 / sim 3 ranks of 5 | 3 ranks, 5/10/15 | no, sim matches (`stances.go`); cite |
| Bloodthrill 1289682 | curve 109366, entry 137214 | 5 ranks of 4% / sim 2 ranks of 10% | 2 ranks, 10/20% | no, sim matches; cite |
| Enrage 12317 | curve 83294 (re-keyed), node 105931 | row 4 col 3 / sim (2,3) | same values, row 3 col 4 | no |
| Gore Drinker, Death Wish, Bloodthirst | node 113569, edges | Enrage -> Gore Drinker, Death Wish -> Flurry / sim already has the new edges | Blood Craze -> Gore Drinker, Death Wish -> Bloodthirst | no (tree already matches) |
| Booming Voice 12321 | SpellEffect 1360666, Spell text | Shouts only / sim cuts the Shouts' cost (`shouts.go`) | Shouts and Piercing Howl | no: Piercing Howl isn't simmed |
| Sweeping Strikes 12292, Concussion Blow 12809 | SpellMisc Attributes_14 0 -> 64, CooldownSetSpell order | | | no: the flag 1374 spells got in 70753, no DPS reading; nothing else on either changes |
| Deep Wounds 12834 | Spell, SpellName, CooldownSetSpell order | | re-sent unchanged | no |
| Tree 1117 | TraitNode, TraitEdge, groups, TraitNodeEntry | sim tree from the notes | client tree above | **yes**: re-import the tree from `1.60.1.70707+hf` (Raging Blows out, Reckless Swings to (3,1), Improved Whirlwind to (5,0), Flurry to (5,2)), remap the talent strings and presets |

## Reckless Swings

Two new spells, neither in any client (70707, 70753, 70755):

- **1329555 "Reckless Swings"**: the talent spell (TraitDefinition 145966, entry 141298, 1 rank, node 113636, Fury row
  4 col 2), the cast and the buff in one. SkillLine 256 (Fury), class mask1 bit 17.
- **1329554 "Reckless Swings"**: the triggered Rage gain, "${$s1/10} Rage generated."

| Question | Hotfix answer | Rows |
|---|---|---|
| Cooldown | **15 s** | SpellCooldowns 102442 RecoveryTime 15000 |
| GCD | **on the GCD, 1 s** (StartRecoveryTime 1000 in global cooldown category 133; Bloodthirst, Whirlwind, Spearing Strike use the same category with 1500) | SpellCooldowns 102442, SpellCategories 191006 |
| Cast time, range | instant, self | SpellMisc 874779 CastingTimeIndex 1, RangeIndex 1 |
| Cost | **25% of base health** (PowerType -2, PowerCostPct 25); no Rage | SpellPower 315180 |
| Duration | **8 s** | SpellMisc 874779 DurationIndex 31 (8000 ms) |
| Charges | **4**, proc chance 100 | SpellAuraOptions 249144 ProcCharges 4 |
| What consumes a charge | ProcTypeMask_0 **4, melee auto attack**: the white swing, both hands; the same type as Flurry's buff 12966 and Unbridled Wrath. No class mask, no hit mask in the DB2 | SpellAuraOptions 249144, SpellEffect 1369167 (aura 42, no mask) |
| Rage a charge | **5**: 1329554 eff0 Effect 30 energize, misc 1 Rage, 50 (tenths). An energize, so a crit doesn't double it | SpellEffect 1369166; tooltip $s1/10/$s2 with eff1 1369168 = 1 |
| Rank curve | 128140, 50 / 50 on 1329555 eff0 (the proc trigger, base 0). Only one rank exists | TDEP 25231 |

Against the sim's assumptions (`sim/warrior/reckless_swings.go`):

| Assumption | Verdict |
|---|---|
| Spell id 1399001 (placeholder) | **changes**: 1329555 (cast and buff), 1329554 (Rage) |
| 15 s cooldown, 8 s, 4 attacks, 5 Rage | **confirmed** |
| Off the GCD | **changes**: it triggers a 1 s GCD and waits on it |
| No health cost | the cost is real (25% of base health) and confirms the dev notes; for DPS it is still nothing, so not modelling it stays right, but the comment should say the data has it |
| Flat 5 Rage, not doubled by a crit | **confirmed** (an energize of the triggered spell) |
| Landed swings, Heroic Strike (and Cleave) included | **partly**: the proc type is melee auto attack, the same as Flurry's charges and Unbridled Wrath, which the sim already lets Heroic Strike and Cleave trigger (their proc mask includes `ProcMaskMeleeMHAuto`), so including them is consistent. "Landed" is not in the DB2: the sim's Flurry uses a charge on any swing, Reckless Swings only on a landed one. Still an assumption; a combat log settles it |

## Other pushes logged with it

Probed 2026-11-04 with `hotfixes.page('enus <push> ', n)` for pushes 113525-113560 (1.5 s apart), then every page of
the pushes with spell rows. Times are the region 91 / 70707 copy where there is one; the 2026-11-04 00:14:20 time is the
region 91 / 70755 copy, which most pushes of the batch share.

| Push | Rows (all copies) | Logged | What | Warrior rows |
|---|---|---|---|---|
| 113526 | about 40 | 2026-11-04 04:40:22 only | Creature, CreatureDifficulty | none |
| 113527 | 39 | 11-03 18:29:17 (and 70755) | warlock: Wrack, curves | none |
| 113528 | 5 | 11-04 00:14:20 | GlobalStrings | none |
| 113530 | 6 | 11-03 18:29:17 | Clearcasting (mage) SpellAuraOptions, a SpellScript | none |
| 113531 | about 35 | 11-03 10:15 to 11-04 00:14 | items, recipes, TransmogHoliday | none |
| 113534 | 250 (also region 95 / 70753) | 11-03 18:29:17 | hunter: Arcane Shot, Serpent Sting, Aspect of the Hawk, Lacerating Strikes (37 SpellEffect, curves) | none |
| 113535 | 24 | 11-03 18:29:17 | creatures | none |
| 113537 | 116 | 11-03 18:29:17 | Counterspell - Silenced (18469, 1324401, new 1329552 / 1329553), Create Spellstone 17728, Enchant Weapon - Winter's Might 21931 | none |
| 113539 | 848 | 11-03 18:29:17 | this push | all |
| 113542 | 56 | 11-03 18:29:17 | paladin: Vengeance, Two-Handed Weapon Specialization 20111 (`hotfix_113483_paladin.md`) | none |
| 113548 | 54 | 11-03 18:29:17 | Death Pact, Baron Rivendare's Soul Drain | none |
| 113554 | 15 | 10-31 to 11-04 | Locale | none |
| 113557 | 775 (also region 95 / 70753) | 11-03 (10:15 to 20:06), 11-04 00:14:20 | 155 SpellEffect rows of Naxxramas encounter spells (Explode Bug, Plague Wave, Despawn ...) | none |
| 113545, 113546, 113556, 113559, 113560 | | 10-30 to 11-02, or 11-04 00:54 / 01:40 (113559) | other dates: spell text, SpellScript, quest items, GameObjects | none |
| 113525, 113529, 113532, 113533, 113536, 113538, 113540, 113541, 113543, 113544, 113547, 113549-113553, 113555, 113558 | 0 | | nothing logged | |

No push of the batch but 113539 has a warrior row.
