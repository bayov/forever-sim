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
4. **Ancestral Fortitude raises armor from items only** (16177: "Increases your armor value from items",
   aura 101 on armor). The sim raised all of the target's armor (Agility's and buffs' too). It's the
   paladin's Toughness way now (`ApplyDynamicEquipScaling`). The tank healer's preset takes Ancestral Healing.
5. **Two Restoration relics** are in (`sim/shaman/items.go`): Tidal Totem (272431, Pix Xizzix, Healing
   Wave 5% cheaper, 1291076) and Totem of Urgency (279249, crafted, Lesser Healing Wave +4% crit, 1306448).
   Both are Phase 1 and allowed by standards 8a.
6. **Downranking is sim/core's** (master 8f304b240 and b57f5cc0a, round 70753): every ranked spell takes
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
| Ancestral Healing (16176) | after one of our heals crits, its target's armor from items +8/17/25% for 15 sec (16177, aura 101); proc flags 0x15550 (no periodic) | **was all armor, now armor from items** | curve 85318 | `TestOrcShamanAncestralHealing`, `TestOrcShamanHealTalentPoints`, `TestHealingFightShamanAncestralFortitude` |
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
- **A HoT's snapshot takes the damage multipliers** (`Dot.SnapshotHeal` in `sim/core/spell_result.go` uses
  `AttackerDamageMultiplier`: `DamageDealtMultiplier` and the school's damage multiplier, not
  `HealingDealtMultiplier`), where a direct heal takes `CasterHealingMultiplier`. Nothing a Restoration
  shaman wears or gets moves one but not the other (Natural Alignment Crystal moves both by 20%), so no shaman
  number moves. It's shared core, and Rejuvenation, Regrowth and Renew go through it: for the orchestrator.
- **The Spiritcaller** (2109, the Restoration tier set: 5 pieces take 1 sec off Riptide's cooldown) has no
  stats in the client yet, and the synthetic set doesn't use it.
- **Totem of Life and Totem of Flowing Water** (Naxxramas, standards 8a) aren't modelled.

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
  its armor from items only.
- Already there: `sim/raid_healing_test.go` (Chain Heal's jumps and falloff, bound ranks, the mana tools),
  `periodic_test.go` (Healing Stream), `mana_returns_test.go` (Mana Spring, Mana Tide, Water Shield),
  `mana_costs_test.go` (Totemic Focus), `mindfulness_test.go`, `natural_grace_test.go`,
  `natures_swiftness_test.go`.
