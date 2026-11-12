# The Restoration shaman's heals: audit, tests and presets (H3, 2026-11-03; round 70753, 2026-11-04)

The healer workers' brief (`orchestration/brief_healers.md`) steps 1, 3 and a re-check of 5 for the
Restoration shaman, the H2 healer (`docs/healing_sim.md`, H2). Branch `healers/shaman`.

**Sources, in order** (`docs/data_sources.md`): Forever's newest client, 1.60.1.70753 (2026-11-02, the build
that writes the October 30 launch-realm hotfix into the tables) on wago.tools (`tools/sources/wago.py`:
SpellEffect, SpellMisc, SpellLevels, SpellPower, SpellCooldowns, SpellCastTimes, SpellDuration,
SpellAuraOptions, SpellClassOptions, SkillLineAbility, TraitDefinition, TraitDefinitionEffectPoints,
TraitNodeEntry and CurvePoint); the tooltips in the same client (Spell.Description); the update ledger
(`orchestration/update_ledger.md`, which has no shaman healing row from 70009 to 70753; the October 30 hotfix
touched the paladin, druid, mage, priest, warlock and hunter, not the shaman tree); then wowsims/sod as a
second reading of a Classic-era mechanic. wowsims/forever has no shaman heals.

Then **Blizzard's development notes, post 9** (2026-11-02,
`orchestration/audit/ref/client_update/sources/devnotes_post9_2026-11-02.txt`), whose one Restoration line is
Riptide's. Client 70753 doesn't carry it, 70755 has the same tables (`orchestration/round_70753/client_70755.md`),
and no server hotfix touches the shaman (`orchestration/round_70753/shaman.md`), so the sim models the notes'
values under Forever (standards 37; under "Notes-only values" below and in the round file).

"Client" below is 70753. `shaman_audit.md` audited the shaman before the heals existed; its section 8 entries
for the healing talents are done again here against `sim/shaman/healing.go` and `talents.go`.

## What changed

1. **Riptide is the development notes' (round 70753)**: no instant heal, ticks only; a 12 sec cooldown (6 in
   the client); 40% less mana, 147 / 180 / 231 for ranks 1 / 2 / 3 (the client's 245 / 300 / 385 × 0.6, the
   notes' "about 40%", rounded to whole mana as SpellPower's ManaCost is). The ticks, their 0.1 coefficient and
   the Chain Heal bonus are the client's. Its Expected Healing (item 3) is now the whole heal over time.
2. **Riptide has three ranks** (was one). The talent teaches rank 1 (408521) at 40, and the Restoration
   trainer line has rank 2 (1239242) at 50 and rank 3 (1239243) at 60 (SkillLineAbility 53198 and 53199). The
   sim had rank 1 only: at 60 it healed 484 to 535 and 89 a tick for 245 mana, where the client's rank 3 heals
   799 to 883 and 161 a tick for 385. Ranks share one cooldown, and one rank of ours on a target replaces
   another's heal over time. Ranks 1 and 2 take Forever's downranking (now sim/core's, item 6). The presets,
   the APLs and the bench's requests cast rank 3 (`1239243`).
3. **Riptide has an Expected Healing** (the whole heal over time; on the client's Riptide, the instant heal's
   average). The raid healer's line "Riptide at the
   most hurt when it fits" compared Health Missing with `spellExpectedHealing(Riptide)`, which Riptide didn't
   have: the APL dropped that value with a warning, so the line cast Riptide whenever the most hurt had none.
4. **Ancestral Fortitude raises all armor** (healer_mechanics-20261104-3, 2026-11-04). 16177 is aura 101
   (MOD_RESISTANCE_PCT) on armor, the priest's Inspiration (14893) aura, which the sim applies to all armor;
   the "from items" talents (Toughness 12299, 20143) are aura 142. Only 16177's tooltip says "from items",
   the talent's (16176) "Increases your target's armor value". So the sim follows the aura
   (`PseudoStats.ArmorMultiplier`, as Inspiration); from 11-03 to 11-04 it raised armor from items only. The
   tank healer's preset takes Ancestral Healing.
5. **Two Restoration relics** are in (`sim/shaman/items.go`): Tidal Totem (272431, Pix Xizzix, Healing
   Wave 5% cheaper, 1291076) and Totem of Urgency (279249, crafted, Lesser Healing Wave +4% crit, 1306448).
   Both are Phase 1 and allowed by standards 8a.
6. **Downranking is sim/core's** (master 8f304b240 and b57f5cc0a, round 70753): every ranked heal (and until 2026-11-08 every damage rank) takes
   `core.DownrankFactor` from its `Rank` and `RequiredLevel` (or `DownrankLevel`), so the shaman's own
   `downrankCoefficient` is gone and Riptide's ranks take the engine's, with the same 5% a level past learn
   level + 17. What it moved here: nothing at 60. Healing Stream Totem is in the penalty now (rank 5 is
   learned at 60, so it keeps its coefficient), and the bench casts top ranks only; master's own shaman bench
   is H2's to the third decimal (Onyxia tank 1.017, below).
7. **The talents are re-picked on the new tree** (round 70753, master f0bb9d4 moved Improved Ghost Wolf to row
   4, where 8 Enhancement points can't reach it). See "Presets" below.

The tests (below) cover each change: the old code fails each of them.

## The audit

### The heals

| Spell | What | Client (70753) | Sim | Source |
|---|---|---|---|---|
| Healing Wave 1-10 | levels, base, growth, variance | 331 1/39/0.7/0.256 … 25357 60/1615/5.5/0.133 (MaxLevel learn + 5) | same, every rank (`HealingWaveRanks`) | SpellLevels, SpellEffect |
| | cost | 25, 45, 80, 155, 200, 265, 340, 440, 560, 620 | same | SpellPower |
| | cast time | 1.5, 2, 2.5, then 3 sec | same | SpellMisc → SpellCastTimes |
| | coefficient | 0.429, 0.571, 0.714, then 0.857 | same, times Forever's downranking | SpellEffect BonusCoefficient, foreverchanges /downranking |
| | GCD | 1.5 sec (category 133) | 1.5 sec | SpellCooldowns |
| Lesser Healing Wave 1-6 | levels, base, growth, variance, cost | 8004 20/153/1.7/0.138/105 … 10468 60/820/4.2/0.109/380 | same | |
| | cast, coefficient | 1.5 sec, 0.429 | same | |
| Chain Heal 1-3 | levels, base, growth, variance, cost | 1064 40/251/2.5/0.140/260, 10622 46/335/2.8/0.138/315, 10623 54/490/3.3/0.132/405 | same | |
| | cast, coefficient | 2.5 sec, 0.714 | same | |
| | targets, falloff | 3 (EffectChainTargets), 50% a jump (EffectChainAmplitude 0.5; 65% only with SoD's Coherence 415096, not in Forever) | 3, 50%, to the most hurt, within the target's party when it's in ours | SpellEffect, tooltip |
| Riptide | ranks | 408521 at 40 (talent), 1239242 at 50, 1239243 at 60 (trainer) | **was 1 rank, now 3** | SkillLineAbility, SpellLevels |
| | heal, growth, variance | 480 + 3.7 to 48; 612 + 4.2 to 58; 841 + 4.9 to 68; 10% | **none under Forever** (the notes); the client's table is kept outside it | SpellEffect eff 0; notes post 9 |
| | ticks | every 3 sec for 15 sec: 89, 115, 161 (+3.7 a level past 60) | 5 ticks of the same | SpellEffect eff 1 (aura 8), SpellDuration |
| | coefficients | 0.214 heal, 0.1 a tick | 0.1 a tick, times sim/core's downranking (rank 1 keeps 85% at 60) | |
| | cost | 245, 300, 385 | **147, 180, 231** (the notes' 40% less) | SpellPower; notes post 9 |
| | cooldown, GCD | 6 sec category 1228 (all ranks), 1.5 sec GCD | **12 sec** (the notes), one timer; 1.5 sec GCD | SpellCooldowns, SpellCategories; notes post 9 |
| | Chain Heal bonus | 25%, a dummy effect (aura 4): "the effectiveness of your Chain Heal casts directly on that target" | 25% on the whole chain when its first target has our Riptide (a call below) | SpellEffect eff 2, tooltip |
| Healing Stream Totem 1-5 | levels, cost, heal | 20-60, 40-80 mana, 5/6/7/9/11 every 2 sec, 0.022 coefficient, 5 min | same (`water_totems.go`) | 5394-10463, 5672-10461 |
| Mana Spring Totem 1-4 | levels, cost, mana | 26/36/46/56, 40/60/80/100 mana, 4/6/8/10 every 2 sec | same | 5675-10497, 5677-10494 |
| Mana Tide Totem 1-4 | levels, cost, mana | 1 (talent)/38/48/58, 10/20/30/60 mana, 88/133/197/290 every 3 sec, 13 sec, 5 min category cooldown | same | 16190, 17354, 17359, 1325256 and their auras |
| Water Shield | globes, mana, cooldowns | 3 globes, 2% of max mana each, 3.5 sec between, 10 min, free, 1.5 sec GCD; proc flags 0x362a8 (hits taken, our heals' and damage spells' crits) | same (`water_shield.go`) | 408510, 408511 |
| Nature's Swiftness | | -100% cast time on the next Nature spell under 10 sec (mask: Lightning Bolt, Chain Lightning, Healing Wave, Lesser Healing Wave, Chain Heal…), 3 min from when it's used up | Nature spells with a cast time, same cooldown rule | 16188 |

No shaman heal changed from 70338 (`shaman_audit.md`) to 70753 but Riptide's ranks, which the earlier audit
never read: it read 408521 alone.

### The talents

| Talent | Client (70753) | Sim | Source | Test |
|---|---|---|---|---|
| Improved Healing Wave (16182) | -0.1 sec a point on Healing Wave (aura 107 cast time, mask 0x40) | same, flat, before haste | curve 85307 | `TestOrcShamanHealTalentPoints` |
| Tidal Focus (16179) | -1% cost a point on Healing Wave, Lesser Healing Wave, Chain Heal and Riptide (0x1c0 + word 2 0x10); +1% melee and spell hit | same; a cost factor on every shaman heal | curves 85321, 126021, 126020 | `TestOrcShamanHealTalentPoints` (Riptide's cost too) |
| Mindfulness (1223033) | 17/33/50% of regen while casting | same | curve 85316 | `mindfulness_test.go` |
| Natural Grace (29187) | -5% threat a point on every magic school | same | curve 85314 | `natural_grace_test.go` |
| Totemic Focus (16173) | -5% a point on every totem's cost | same | curve 85315 | `mana_costs_test.go` |
| Improved Reincarnation (16184) | +2% health a point (the sim's only part) | same | | `base_stats_test.go` |
| Ancestral Healing (16176) | after one of our heals crits, its target's armor +8/17/25% for 15 sec (16177, aura 101, all armor as Inspiration); proc flags 0x15550 (no periodic) | **all armor again (healer_mechanics-20261104-3; items only 11-03 to 11-04)** | curve 85318 | `TestOrcShamanAncestralHealing`, `TestOrcShamanHealTalentPoints`, `TestHealingFightShamanAncestralFortitude` |
| Healing Focus (16181) | 23/47/70% chance a hit doesn't push a heal back | same | curve 85317 | `TestOrcShamanHealingFocusPushback` |
| Water Shield (talent) | teaches 408510 | same | | `mana_returns_test.go` |
| Tidal Mastery (16194) | +1% crit a point on Healing Wave, Lesser Healing Wave, Chain Heal and Riptide (0x5c0 + word 2 0x10; 0x400 is Lightning Shield, whose orbs never crit) | same | curve 85308 | `TestOrcShamanHealTalentPoints` |
| Restorative Totems (16187) | +5% a point on Mana Spring, +10% a point on Healing Stream | same | curves 85320, 85319 | `periodic_test.go`, `mana_returns_test.go` |
| Mana Tide Totem (talent) | teaches 16190 | same | | `mana_returns_test.go` |
| Healing Way (29206) | **a flat** +8/17/25% on Healing Wave (aura 108, misc 0, mask 0x40). Classic's stacking buff on the target (29203) isn't in the client at all, nor are 29202 and 29205 | flat, multiplying with Purification | curve 85322 | `TestOrcShamanHealTalentPoints`, `TestOrcShamanHealAmounts` |
| Nature's Swiftness (16188) | as above | same | | `TestOrcShamanNaturesSwiftnessHeal`, `natures_swiftness_test.go` |
| Purification (16178) | +2% a point on Healing Wave, Lesser Healing Wave, Chain Heal and Riptide's heal (eff 0) and Riptide's ticks (eff 1, misc 22). Not Healing Stream (0x2000) | same | curves 85323, 87379 | `TestOrcShamanHealTalentPoints`, `TestOrcShamanRiptide` |
| Riptide (talent) | rank 1, then ranks 2 and 3 from the trainer | **was rank 1 only** | | `TestOrcShamanRiptideRanks` |
| Ancestral Knowledge (17485, Enhancement) | +2% Intellect a point | same | curve 85266 | `stat_multipliers_test.go` |
| Thundering Strikes (16255, Enhancement) | +1% crit a point on every spell and attack | same, the heals too | talent ranks | `TestOrcShamanThunderingStrikes`, `TestOrcShamanThunderingStrikesHeals` |
| Mental Quickness (30812, Enhancement) | 15% of Intellect a point as spell damage and healing | same (spell power, which heals count) | curves 85302, 85303 | `mental_quickness_test.go` |

### Not changed, and why

- **Notes-only values** (standards 37; client 1.60.1.70753 and 70755 unchanged, no server hotfix): Riptide's
  missing instant heal, its 12 sec cooldown and its 147 / 180 / 231 mana. All under `Env.IsForever()`
  (`riptideNotes`); listed in `orchestration/round_70753/shaman.md` as well. The next hotfix check or the
  launch logs settle them.
- **Healing Wave below level 20.** Classic's servers cut the coefficient of a spell learned below 20; no
  client field holds it and the sim applies it to no spell (Lightning Bolt neither). Leveling only.
- **Chain Heal's Expected Healing** counts the first target only, as HealComm did. Bound ranks pick by it.
- **A HoT's snapshot took the damage multipliers** (`Dot.SnapshotHeal` used `AttackerDamageMultiplier`) where
  a direct heal takes `CasterHealingMultiplier`. Fixed on master by the priest's session (it now takes
  `CasterHealingMultiplier`); no shaman number moved, since nothing a Restoration shaman has moved one but not
  the other.
- **Forever's Tier 1 Restoration set** (The Spiritcaller, 2109; bonus spells 1300960, 1301059, 1301093,
  1301713): +10 Spirit, friendly totems' radius, +26 healing (9 of it damage), Riptide 1 sec off its cooldown
  (aura 107, mask word 2 0x10). The synthetic set carries it on the head and chest, so only the 2 piece is live
  (standards 6; healer_mechanics-20261104-6, `ItemSetRestorationSyntheticPhase1`,
  `TestRestorationSyntheticSet`). What the 4 and 5 pieces would add is under "Presets".
- **Totem of Life and Totem of Flowing Water** (Naxxramas, standards 8a) aren't modelled.

## Fixes after the healer audit (2026-11-04)

Branch `healers/fix-shaman`. Master already had the orchestrator's engine fixes (raid damage share, school crit on
direct heals, health clamp, the tank bench's idle raid fillers), which moved every healer's bench; the tables
under "Presets (round 70753)" below are from before them.

- **healer_mechanics-20261104-3**, fixed: Ancestral Fortitude raises all armor (item 4 under "What changed").
- **healer_mechanics-20261104-6**, fixed: the synthetic head and chest carry the Tier 1 Restoration set, 2 piece
  live. What the rest would add, 1,000 iterations, carry, 7-boss mean: 4 piece +0.003 tank, +0.000 raid; 5 piece
  (also Riptide 1 sec) the same +0.003 tank (the tank APL recasts Riptide only when its HoT is off, so the
  cooldown never binds) and -0.007 raid (noise). Neither is in the preset (standards 6).
- **healer_mechanics-20261104-10**, fixed: the untested list has tests (under "Tests"); every value matched client
  70753. The client's Riptide branch runs only outside Forever and has none (standards 38).
- **healer_fairness-20261104-2**, fixed: the gates re-tuned, 1,000 iterations, carry, mean of the 7 boss files.
  - Tank: Healing Wave 10/8/6 on the tank when it misses at least x of the heal's Expected Healing, with the
    Lesser Healing Wave 6/5 lines either gated the same way at 1.0x or only on mana (at least 50%):

    | Healing Wave gate | Lesser Healing Wave gated | Lesser Healing Wave on mana only |
    |---|---|---|
    | 1.0x (before) | 1.115 | 1.129 |
    | 0.9x | 1.107 | 1.129 |
    | 0.8x | 1.096 | 1.130 |
    | 0.7x | 1.099 | 1.130 |
    | **0.6x** | 1.097 | **1.134** |
    | 0.5x / 0.4x / 0.3x | | 1.134 / 1.133 / 1.134 |

    Kept: 0.6x, and Lesser Healing Wave on mana only (+0.019). The win is almost all the Lesser Healing Wave
    gate's; with it off, the Healing Wave gate is flat from 0.3x to 0.6x.
  - Raid: the Healing Wave gate at the most hurt changes nothing (0.6x and 0.8x equal). Chain Heal 3/2/1's gate:
    1.0x 1.380, 0.9x 1.392, 0.8x 1.395, **0.7x 1.396**, 0.6x 1.387. Kept 0.7x (+0.016).
  - In `ui/restoration_shaman/apls/{tank,raid}_healer.apl.json` and the bench requests that copy them.

**Final, 1,000 iterations, carry, default length** (this branch):

| Role | Magmadar | Golemagg | Ragnaros | Onyxia | Broodlord | Vaelastrasz | Chromaggus | Mean |
|---|---|---|---|---|---|---|---|---|
| Tank | 1.852 | 1.777 | 0.525 | 1.235 | 0.852 | 0.818 | 0.881 | **1.134** |
| Raid | 1.809 | 2.662 | 1.235 | 0.919 | 0.902 | 1.152 | 1.090 | **1.396** |

## Presets (round 70753)

Consumables (2026-11-12): Greater Cleric's Elixir for the Greater Arcane Elixir, Runn Tum Tuber Surprise, the Wizard Oil (the Brilliant Mana Oil is a Zandalar formula), Flask of Supreme Power, in both roles; the rest as every healer (`docs/healing_sim.md`, Consumables).

The bench (`tools/healbench`), one healer, each row at its own length; 500 iterations to screen and 1,000 for
the finalists and the tables below. "Master" is origin/master 3d1cb9701 before this branch, which is H2's to the
third decimal.

**The bench drop after the 11-03 merge was the bench, not the sim.** The shared `scratchpad/healbench` binary
was rebuilt at 09:32 from another healer's branch (same size and time as its builds), which has none of
Riptide's ranks 2 and 3, so the requests' Riptide (1239243) never cast. Built from this branch into a folder of
its own, nothing dropped, and master's own build gives H2's numbers (Onyxia tank 1.017, Golemagg raid 2.083,
Vaelastrasz raid 1.199, Vaelastrasz tank 0.648).

**Talents.** Improved Ghost Wolf left the tree's reach, so the 5 Enhancement points go to Ancestral Knowledge 5
and the 3 left to Thundering Strikes 3 (1% heal crit a point; Guardian Totems and Improved Lightning Shield do
nothing for a healer). Restoration is H2's.

| Talents (1,000 iterations, carry) | Tank healer mean (9 rows) | Raid healer mean (8 rows) |
|---|---|---|
| `-005102`, Ancestral Healing / Healing Focus (H2's) | 0.955 | 1.126 |
| **`-035`** (Thundering Strikes 3), tank with Ancestral Healing, raid with Healing Focus | **0.959** | **1.123** |
| `-053` (Thundering Strikes 5) | 0.959 | 1.122 |
| tank on Healing Focus / raid on Ancestral Healing | 0.918 | 1.094 |

The raid healer's choice is within noise (0.003); `-035` is kept for both so the two presets differ only in
Restoration.

**The tank healer** gains two lines, the brief's gap check (bigger ranks, Riptide's 12 sec, Lesser Healing
Wave, Nature's Swiftness):

- Lesser Healing Wave 6 and 5 at the tank, bound, while mana is at least 50%, right after Riptide: more healing
  a second than Healing Wave for a little more a mana. (Rank 5 had rank 6's condition and never fired; it's dropped,
  healer_rotations-20261104-3: tank mean carry 1.257 with or without it, seed 101, 1,000 iterations; rank 5 in the
  35-50% or 40-50% mana band 1.254 / 1.253.) Floors: 0% 1.035 (runs dry at 300 sec: 0.879), 50% 1.065,
  60% 1.059, 70% 1.051, 80% 1.029; rank 6 alone at 70% 1.038 (the 0.959 base had none).
- Nature's Swiftness when the tank is under 50% health, before Riptide (the bound Healing Wave 10 follows it):
  +0.011 alone, +0.005 with the Lesser Healing Wave lines (50% and 40%: within 0.004; 35%: -0.010).
- Kept: Healing Wave ranks bound by what the tank misses (casting 10 at 75% of its heal: -0.008), Mana Tide at
  its floor (20% and 50%: within 0.001), Water Shield whenever its globes are spent (below 60% mana only:
  +0.003, but it loses both long Onyxia rows).

**The raid healer** is H2's APL on the notes' Riptide. Mana Tide's and Water Shield's mana floors never bind on
the bench (Mana Tide 50%, Water Shield below 40% or 80%: the same numbers to the third decimal; the fights end
first), which is why the 11-03 threshold runs were all equal. Riptide on the tank instead of the most hurt
(-0.007) and Lesser Healing Wave 6 at the most hurt above 70% mana (-0.013) lose.

**Nature's Swiftness on the raid healer (healer_rotations-20261112-2, 2026-11-12)**: the raid APL had no Nature's
Swiftness line, so `autocastOtherCooldowns` raised it at the pull and on every cooldown and the next Chain Heal
took it. It now has its own two lines after the autocast, as the tank healer and both druids: Nature's Swiftness
when the lowest in the assignment is under 65% health, then an instant Healing Wave 10 at them. Raid role, own
lengths, 1,000 iterations, mean carry of the 7 files, seed 101: autocast 1.392, 45% 1.402, 50% 1.411, 55% 1.403,
60% 1.395, **65% 1.412**, 70% 1.406, 75% 1.390; seed 202: autocast 1.389, 50% 1.404, 55% 1.400, **65% 1.411**
(+0.020 / +0.022). NS costs no mana, so the 180 / 300 s side is unchanged.

**Final, 1,000 iterations, carry** (this branch with the priest's and paladin's work merged; the priest and
paladin columns are the coordinator's table from their branches):

| Boss | Tank: master | **Tank: now** | Holy priest | Paladin | Raid: master | **Raid: now** |
|---|---|---|---|---|---|---|
| Magmadar | 1.536 | **1.798** | 2.089 | 1.645 | 1.605 | **1.540** |
| Golemagg | 1.441 | **1.716** | 1.915 | 1.625 | 2.083 | **1.913** |
| Ragnaros | 0.469 | **0.521** | 0.582 | 0.521 | 1.203 | **1.171** |
| Onyxia | 1.017 | **1.141** | 1.230 | 1.081 | 0.692 | **0.678** |
| Broodlord | 0.618 | **0.837** | 0.891 | 0.750 | 0.835 | **0.817** |
| Vaelastrasz | 0.648 | **0.775** | 0.903 | 0.723 | 1.199 | **1.110** |
| Chromaggus | 0.764 | **0.884** | 1.006 | 0.825 | 0.969 | **0.906** |
| Onyxia 180 sec | 0.965 | **1.010** | | | | |
| Onyxia 300 sec | 0.956 | **0.969** | | | | |
| Raid profile | | | | | 0.880 | **0.847** |
| Mean | 0.935 | **1.072** | | | 1.183 | **1.123** |

**What the notes' Riptide costs**, against the same presets on the client's Riptide (11-03 runs): about -0.07 on
the raid healer (1.196 to 1.123: no instant heal on the most hurt, half as often) and about -0.035 on the tank
healer with H2's APL (0.990 to 0.955). The new tank lines more than make it up there; nothing does on the raid
healer, so its drop from master is the notes' and stays.

**The gap to the Holy priest on the tank is the class's.** The shaman now passes the paladin on every boss but
Ragnaros (equal) and stays 0.06 to 0.29 below the priest. With Riptide a 12 sec HoT, its single target tools are
Healing Wave and Lesser Healing Wave (2.5 and 1.5 sec, no reduction past Improved Healing Wave's) and one Nature's
Swiftness; the priest has Flash Heal, Greater Heal, Renew, Power Word: Shield and Prayer of Mending on the tank
as well. No rank, mana gate or cooldown use on the bench closes it: the best lines above are within 0.006 of
each other. On the re-bench after the 2026-11-04 fixes and the tank bench's change (tanks face the boss;
healing_sim.md, H3) the gap is 0.020 on the mean (1.242 against 1.262), the shaman ahead on Broodlord and
Chromaggus; the paladin (1.176) is behind it on every file.

## Tests

Every heal and healing talent has a test whose numbers would fail with the mechanic broken (the new ones were
run against the old code, or with the mechanic taken out, and fail):

- `sim/shaman/enhancement/healing_test.go`: the rank tables (Riptide's three), each heal's amount with healing
  power, Purification and Healing Way, cast times, costs, Nature's Swiftness, Ancestral Healing on armor from
  items, Riptide rank 3's missing instant heal, its ticks, cost and Chain Heal bonus (`TestOrcShamanRiptide`),
  the ranks a level knows, their costs, the shared 12 sec cooldown, one Riptide a target and rank 1's
  downranking (`TestOrcShamanRiptideRanks`), and Riptide's Expected Healing in a rotation
  (`TestOrcShamanRiptideExpectedHealing`).
- `sim/shaman/enhancement/healing_talents_test.go`: every point of every healing talent against its curve
  (Tidal Focus on the notes' Riptide mana), Healing Focus's pushback chance over 400 casts, Thundering Strikes'
  crit on the heals.
- `sim/shaman/enhancement/healing_relics_test.go`: Tidal Totem and Totem of Urgency.
- `sim/raid_healing_shaman_test.go`: on a healing fight, Ancestral Fortitude lands on the healed tank and raises
  all of its armor.
- `sim/shaman/enhancement/healing_untested_test.go` (healer_mechanics-20261104-10, against client 70753's
  tables): `TestOrcShamanHealRankTable` (every rank of Healing Wave, Lesser Healing Wave and Chain Heal: base,
  growth, variance, mana, cast time, coefficient; Healing Wave 2-3 at 2 / 2.5 sec and .571 / .714),
  `TestOrcShamanChainHealAmount` (Chain Heal 3 and its .714), `TestOrcShamanRiptideTickCrits` (ticks crit,
  Attributes_8 0x200, and Tidal Mastery reaches them), `TestOrcShamanRiptideTickCritsProcNothing` (no Ancestral
  Fortitude or Water Shield globe from a tick crit), `TestOrcShamanNaturesSwiftnessKept` (Riptide and a totem
  leave it up), `TestOrcShamanHealingStreamNoTidalTalents`, `TestOrcShamanAncestralFortitudeTwoShamans` (one
  buff, refreshed, not stacked). The client's Riptide branch runs only outside Forever, so it has no test
  (standards 38).
- `sim/shaman/enhancement/healing_set_test.go`: `TestRestorationSyntheticSet`, the Tier 1 set on two pieces and
  its 4 and 5 piece.
- Already there: `sim/raid_healing_test.go` (Chain Heal's jumps and falloff, bound ranks, the mana tools),
  `periodic_test.go` (Healing Stream), `mana_returns_test.go` (Mana Spring, Mana Tide, Water Shield),
  `mana_costs_test.go` (Totemic Focus), `mindfulness_test.go`, `natural_grace_test.go`,
  `natures_swiftness_test.go`.

## Calls for the user

1. **The notes' Riptide** (standards 37). Default: modelled under Forever (no instant heal, 12 sec, 147 / 180 /
   231 mana). The client's instead: about +0.07 raid healer mean carry and +0.035 tank healer (H2's tank APL).
2. **Riptide's Chain Heal bonus on the whole chain.** The client's effect is a dummy, and the tooltip's "your
   Chain Heal casts directly on that target" says which casts, not which jumps. Default: the whole chain of a
   Chain Heal cast at a target with our Riptide. The first
   target only: the same tank numbers, the raid healer within 0.02 a row (e.g. Magmadar 1.621 against 1.639,
   the raid profile 0.909 against 0.900, on the client's Riptide).
3. **Tidal Totem on the tank healer.** Default: not in the preset's gear. With it: Golemagg +0.036, the other
   rows within 0.005 (on the client's Riptide).
4. **The raid healer's Enhancement points.** Default: `-035` (Thundering Strikes 3), as the tank healer. H2's
   `-005102` (still legal, and what `sim/compsearch/presets.go` and `ui/raid_sim/data/table.json` give a raid
   member): +0.003 mean, within noise.
5. **Ancestral Fortitude on all armor or armor from items** (healer_mechanics-20261104-3). Default: all armor,
   as its aura 101 (Inspiration's) says; the buff's tooltip says "from items", and how the server applies 101
   against 142 can't be read from the client. Under 1% of the tank's score either way.
6. **The tank healer's Lesser Healing Wave on mana alone** (healer_fairness-20261104-2). Default: Lesser Healing
   Wave 6/5 at the tank whenever mana is at least 50%, without a health check (+0.019 tank mean); its overheal is
   cut by the bench's stop-cast rule. A player would cast it on a tank about to be hit, as the bench does.
