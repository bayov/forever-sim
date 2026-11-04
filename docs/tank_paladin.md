# Protection paladin tank

The Protection paladin (`ProtectionPaladin`, `sim/paladin/protection`, `ui/protection_paladin`) as a tank
under Forever: the rules audit of every tank ability and talent, what was fixed and built, the tests, the
rotation, the synthetic gear, the three presets and their numbers. The tank workers' brief
(`orchestration/brief_tanks.md`) set the work; `docs/tank_sim.md` has the scoring (threat against survival,
healing needed, the encounters) and the first pass of this audit, which this builds on.

**Sources, best first** (`docs/data_sources.md`): Forever's newest client, 1.60.1.70753 (2026-11-02, which
writes the launch realms' October 30 hotfix into the tables; `tools/sources/wago.py`), its talent trees
(`tools/forever_talents/client/paladin.json`), wowsims/forever (github.com/wowsims/forever, its paladin code
and the in-game measurements its commits cite), then Classic. "Client" below means 70753's SpellEffect,
SpellMisc, SpellLevels, SpellPower, SpellCooldowns and SpellDuration rows.

## Realignment after the 70753 update (2026-11-04)

The tank work paused on 2026-11-03 for the update to client 1.60.1.70753 and Blizzard's November 2 development
notes (`orchestration/round_70753/`, standards 37 to 39). This audit had already been read against 70753's rows;
the restart checked what came after it and what master's paladin round changed in the same code.

- **Client 1.60.1.70755** (2026-11-03): every row of the tank's spells is 70753's (`wago.py` SpellEffect,
  SpellMisc, SpellLevels, SpellPower, SpellCooldowns, SpellDuration, SpellAuraOptions, SpellCategories, Spell,
  SpellName for Righteous Fury, every Seal of Fury rank, its hit and Judgement, the taunt, the four Holy Shield
  ranks, Holy Strike, Templar's Bulwark, Divine Shield, Divine Protection, Forbearance, Redoubt, Reckoning, Iron
  Creed, Shieldsworn, Shield Specialization, Sacred Duty, Steadfast Libram's aura, Judgement, Swift Judgement and
  Lay on Hands: no difference). wowsims/forever's 041cd46 (2026-11-03) adds Holy Shield rank 4 from 70755 with the
  values this sim has had since 2026-11-03 (221 a block, 8%, 240 mana, level 60).
- **Server hotfixes since**: push 113483 is rogue only; push 113542, the paladin rows of the same second
  (`round_70753/hotfix_113483_paladin.md`), is Vengeance's 5 stacks and Two-Handed Weapon Specialization's Holy
  school, neither of which a shield tank reaches. No row for a tank spell.
- **The notes' Protection lines**, against the sim (tiers: C client, N notes):

| Note | Tier | Sim |
|---|---|---|
| Righteous Fury: 33% all threat, 90% more on Holy | N, C | right (`righteous_fury.go`, `TestRighteousFuryThreat`) |
| Righteous Fury at level 10, free with The Quick and the Dead | N, C | n/a at level 60 |
| Salvation, Mass Salvation, Revoke removed | N, C | SafeDPS's reference member makes 1 threat a point of damage (the orchestrator's `tank_score.go` and `tankbench -dps-threat`); no bench request carries Salvation |
| Judgement of Fury 3 sec | N, C | right (the taunt, 20232, isn't modelled: a tank on top gains nothing) |
| Reckoning performs its number of attacks by stacks; no stacks while sitting | N | right: the next swing spends every stored attack, up to 4 (`core.StoreExtraMHAttack`); sitting n/a |
| Divine Intellect to Protection row 1; Toughness row 3; Anticipation row 4; Templar's Bulwark row 3; Holy Shield row 5 | N, C | the 10-30 trees the sim already reads (`client/paladin.json`); no talent string moved |
| Sacred Duty row 2: 30 sec / 1 min off Divine Shield, Divine Protection and Templar's Bulwark, 5 / 10 sec off Forbearance, no Stamina | N, C | right (`TestTemplarsBulwark`) |
| Improved Seal of Fury, One-Handed Weapon Specialization removed | N, C | right (not in the tree) |
| Templar's Bulwark is a Magic effect and dispellable | N, **C**: 1311015's SpellCategories DispelType is 1 (Magic) in 70753 and 70755, 0 in 70707 (round 70753's `paladin.md` and foreverchanges.pro list it as notes-only; the client does carry it) | moves nothing: nothing in the sim dispels a tank's buff and no scoring encounter has a boss who dispels; the sim's auras have no dispel type to set |
| Improved Righteous Fury 3 / 6 / 9% | N, C | right |
| Shield Specialization 15 / 30% block value, 50 / 100% for 4% of max mana | N, C | right (its mana makes no threat) |
| Swift Judgement unleashes the seal, its own cooldown | N, C | right (master's `swift_judgement.go`; it judges Seal of Fury too) |
| Stoicism 10 / 20% off stun and charm | N, C | not simmed; worth nothing on the scoring bosses |
| Iron Creed 3 points, 10 / 20 / 30% Holy Strike threat, 2 / 4 / 6% damage taken for 6 sec | N, C | right (any Holy Strike, as wowsims/forever measured) |
| Shieldsworn | N, C | right (the tooltip names Holy Shield and Shield Specialization only, so a Shieldsworn Holy Strike rolls no Reckoning) |
| Two-Handed Weapon Specialization's threat cut | N, C | n/a with a shield |
| Talents lose the low-rank proc penalty; spells keep Forever's downranking penalty | N | master's `core/downrank.go` (b57f5cc0a). **Fixed here**: Seal of Fury's hit and Judgement of Fury, rank-less sub-spells, now carry their seal rank's learn level (`DownrankLevel`), as master's round did for Seal of Righteousness and Seal of Command; the top rank (58) loses nothing at 60. `TestSealOfFury` |

**Master's paladin round against this work.** The automatic merge left `sim/paladin/` consistent: master
changed `talents.go` (Vengeance, Two-Handed Weapon Specialization), `consecration.go`, `soc.go` and `sor.go`
(the downranking penalty), none of which a shield tank's Seal of Fury, Holy Shield, Holy Strike, Templar's Bulwark
or immunities read; nothing is applied twice or lost (every tank test passes). Two things the merge exposed:

- `TestProtection`'s settings runs (blank gear, the UI's `basic_prot` rotation) had dropped by about half since
  2026-11-03, not under 1% as this doc said: the rotation cast Holy Shield only while aura 20134 was up, the old
  Redoubt id that this work corrected to 20128, so it never cast it. Both UI rotations now read 20128 and cast
  Holy Shield rank 4; the suite is back within 0.5% of master's golden, except the unbuffed settings runs on
  blank gear (-2% to -16%: rank 4 costs 240 mana to rank 3's 195).
- `TestRaidForeverReportHashesStable`: the ten report paladin builds (Retribution, Shock Ret, Sub Holy and their
  duties) moved on the tank branch (659cad198) against master's hashes, and nothing here moves them further. Their
  DPS, TPS and every action's casts, damage and threat are identical to master's to the last digit; the result
  only lists 12 more auras at zero uptime (Seal of Fury's ranks, Divine Shield and Protection, Templar's Bulwark,
  Forbearance and so on, which the tank code registers for every paladin), and the hash covers the aura list. The
  paladin lines of `sim/testdata/report_hashes.txt` are rewritten for that reason alone. (The ten warrior lines also
  move on 659cad198, threat only, DPS identical; that is the warrior work's, not this.)
- `TestRetribution`'s two lines for this set (AllItems, 365.2 to 341.3 and 354.6 to 330.1) were stale since the
  70753 merge (bc2bfa009): master's round moved the Retribution baseline from 354.57 to 330.13 (its own
  `RetributionSyntheticSpellPowerOn-UseTrinket` line, equal to this set's trinket line before and after), and the
  merge took master's golden but kept the branch's values for lines master doesn't have. The set's own stats don't
  move them (the same 341.3 / 330.1 with the old or the new rings and main hand).
- The synthetic set's hit was the gear row's 3.2% (5.7% with Precision); the restart brief sizes it to the 8% cap
  counting talents (standards 4), +29 at Balanced (Gear, below).

**What the update moved** (Balanced, the mean of the six scoring encounters, the 2026-11-03 bench request: the
synthetic set before the hit change, Tauren, the old rotation; tankbench, 1,000 iterations): **2085 before, 1133
after**. The paladin's own rules moved it by -8 (2085 to 2077 with the old 0.7 reference, within the noise); the
SafeDPS reference member's 1 threat a point of damage (Salvation gone) is the rest, -944: SafeDPS 1572 to 1101.
Through rotopt's objective (the template's defaults, 2,000 iterations) it is 3348 to 2101.

## Rules audit

### Abilities

| Ability | Client (70753) | Sim before | Source, tier | Verdict |
|---|---|---|---|---|
| Righteous Fury (25780) | +33% all threat, +90% more on Holy (effects 4 and 0, each its own factor), 30 min, 30% of base mana | 1.33 x, Holy 1.9 x on top, on for the fight (option) | client | right |
| Improved Righteous Fury's damage cut (Righteous Fury effect 1) | 0 base, the talent's 3% a rank | 3% a rank while Righteous Fury is on | client | right |
| Seal of Fury (20423 at 58, 1311649 to 20422 below) | 200 mana, 30 sec; each white hit 35 Holy (20418) + 10% of spell power; with a shield an absorb of 50% of the hit; Judgement of Fury (20414): 153 + 3.69 a level to 64, 8.76% spread, 42.9% of spell power, melee table, then a 3 sec taunt (20232) | **missing** | client; wowsims/forever's seal_of_fury.go reads the same | **built** (`seal_of_fury.go`) |
| Seal of Fury's effect 0 / Judgement of Fury's effect 2 | a dummy of 1607 + 42 a level, 9% / 18% of spell power | - | client, no tooltip names it | not read, a call (below) |
| Judgement of Fury's taunt (20232) | taunt, 3 sec | - | client | not modelled: a tank on top gains nothing from a taunt |
| Seal of Righteousness (20293) | the DPS audit's | right | client, wowsims/forever | right; 97 behind Seal of Fury at Balanced (Valor set) |
| Judgement (20271) | 6% of base mana, 10 sec, off the global cooldown, doesn't consume the seal | the same | client | right |
| Holy Shield (20925 / 20927 / 20928) | 30% block, 4 charges, 10 sec, 75 / 110 / 153 a block, 8% of spell power, 1.2 x threat, 100 / 150 / 195 mana at 30 / 40 / 50, one 10 sec category cooldown (931) | the same, each rank its own cooldown | client | right; the ranks now share their cooldown |
| Holy Shield rank 4 (1239570) | 60, 240 mana, 221 a block, the rest as above | **missing** (the top rank was 20928) | client (new in the hotfix) | **added** |
| Holy Strike (10333) | 20 mana, 10 sec (category 2404), 50% weapon + 93 Holy, Holy school, melee table | the DPS audit's | client | right; its threat is its damage x Righteous Fury x Iron Creed (no flat threat in the rows) |
| Consecration (20924) | 565 mana, 8 sec, 8 sec cooldown | the DPS audit's | client | right |
| Templar's Bulwark (1311015, talent) | absorb 100% of max health, 8 sec, 110 mana, 5 min, off the global cooldown, Forbearance | **missing** (the talent did nothing) | client; wowsims/forever's templars_bulwark.go | **built** (`templars_bulwark.go`) |
| Divine Shield (642, 1020) | immune to all, -50% damage dealt, 10 / 12 sec, 75 / 110 mana, 5 min (category 37), Forbearance | missing | client | **built** (`divine_shield.go`), not in a scoring rotation |
| Divine Protection (498, 5573) | immune to all, pacify, 6 / 8 sec, 15 / 35 mana, the same category | missing | client | **built**, not in a scoring rotation |
| Forbearance (25771) | 1 min, Sacred Duty -5 sec a rank | the same | client | right |
| Lay on Hands (10310) | heals max health, drains all mana, 550 mana to the target, 20 min | the same, cast by name only | client | right (a health button, out of scoring) |
| Devotion Aura (10293) | 735 armor | the raid buff's 735 under Forever (both factions) | client | right: the tank's own aura is the raid's |
| Blessing of Kings (20217) | 10% stats, 1 hr | the raid buff | client | right |
| Blessing of Sanctuary (20911-20914) | **no row**: gone from Forever | registered when chosen | client | **fixed**: Classic only |
| Exorcism, Holy Wrath, Hammer of Wrath | the DPS audit's | right | client | n/a on the scoring bosses (no Undead or Demon, execute only) |
| Taunts | Forever has no Righteous Defense or Hand of Reckoning trainer spell; Judgement of Fury is the taunt | - | client | n/a |

### Talents (Protection, and what a tank build reaches elsewhere)

| Talent | Client (70753, the trait ranks) | Sim before | Source, tier | Verdict |
|---|---|---|---|---|
| Divine Intellect | 2% Intellect a rank | the same | client | right |
| Redoubt | 10% of damaging melee hits taken give 6% block a rank (5/5: 30%) for 10 sec or 5 blocks (20128) | the same, aura id 20134 (no such spell) | client | right; id fixed to 20128 |
| Precision | 1% melee and spell hit a rank | the same | client | right |
| Guardian's Favor | Blessing of Protection and Freedom | not simmed | client | worth nothing on a boss |
| Sacred Duty | -30 sec a rank off Divine Shield, Divine Protection and Templar's Bulwark, -5 sec a rank off Forbearance | Forbearance only (the spells didn't exist) | client | **fixed** with the spells |
| Templar's Bulwark | see Abilities | missing | client | **built** |
| Improved Hammer of Justice | Hammer of Justice's cooldown | not simmed | client | worth nothing on a boss |
| Shield Specialization | +15% block value a rank; a block (or a Shieldsworn Holy Strike) restores 4% of max mana, 50% a rank, once every 3 sec | the same, but the mana made threat | client: 1310925 has the no threat attribute (Attributes_1 0x400); wowsims/forever reads it the same | **fixed**: its mana makes no threat (`core/mana.go` honours a source's NoThreat for mana, as for Rage) |
| Toughness | 2% armor from items a rank | the same | client | right |
| Swift Judgement | a free Judgement on its own 1 min cooldown, off the global cooldown | the same | client | right |
| Improved Righteous Fury | 3% less damage taken a rank with Righteous Fury | the same | client | right |
| Anticipation | 4 Defense a rank | the same | client | right |
| Stoicism | stun and charm durations | not simmed | client | worth nothing on the scoring bosses |
| Holy Shield | see Abilities | ranks 1-3 | client | rank 4 **added** |
| Reckoning | 8% a rank after a block (at most once every 1.5 sec), 20% a rank after a crit taken; up to 4 stored extra attacks, spent by the next swing | the same, but the 1.5 sec held crits back too | client ("Cannot occur from Blocks more than once every 1.5 sec"), wowsims/forever | **fixed**: the cooldown is the blocks' |
| Iron Creed | +10% Holy Strike threat a rank; with Righteous Fury a Holy Strike takes 2% a rank off damage taken for 6 sec | the same, but only after a Holy Strike that landed | client; wowsims/forever checked in game that a missed or avoided Holy Strike still gives the cut (talents_protection.go) | **fixed**: any Holy Strike |
| Shieldsworn | Holy Strike 4 sec faster, deals Holy Shield's damage and rolls Shield Specialization's mana without a charge or cooldown; spells blockable for 6 sec (1328551) | the same, but the spell block | client | right; the spell block is moot: the encounters' magic is unblockable either way (wowsims/forever reads it as letting melee of any school be blocked) |
| Benediction (Retribution) | 2% off instant costs a rank | the same | client | right |
| Improved Judgement (Retribution) | -1 sec a rank | the same | client | right |
| Deflection (Retribution) | 1% parry a rank | the same | client | right |
| Eye for an Eye (Retribution) | 5% of a crit taken back, a rank | not simmed | client | worth nothing under Forever's 0.6% boss crit (docs/tank_sim.md) |
| Pursuit of Justice (Retribution) | movement speed | not simmed | client | worth nothing in the score; a call (below) |

### Relics

| Libram | Client | Verdict |
|---|---|---|
| Steadfast Libram (279247, crafted) | +30% block value while Holy Shield is up (1306749, aura 272) | **built**, added to the other block value percents as Shield Specialization's (the same aura); the preset's libram |
| Sentinel's Libram (272434, a dungeon drop) | Swift Judgement -10 sec | **built**; 104 behind Steadfast at Balanced |
| Libram of Law (272435) | 4% on Judgement of Righteousness and Judgement of Fury | Judgement of Fury **added** to it; 98 behind Steadfast |

### Threat, as the sim has it

- Everything the paladin does is threat = damage x the spell's multiplier x 1.33 (Righteous Fury's all threat),
  and Holy spells x 1.9 more. Holy Shield's block damage carries 1.2 on top, Holy Strike Iron Creed's 1.3 at 3/3.
  The client states no flat threat on any paladin spell, so there is none (wowsims/forever has none either).
- Mana gains make 0.5 threat a mana under Forever (the raid's Judgement of Wisdom, mana potions), split over
  the enemies, except a source with the no threat attribute: Shield Specialization's.
- Not modelled: the taunt. Seal of Fury's absorb makes no threat (it heals nothing).

## What was built and fixed

- **Seal of Fury**, all seven ranks, with its absorb and Judgement of Fury; `PaladinSeal.Fury` (proto
  `paladin.proto`, 3) for the primary seal option and the UI. On the Valor set at Balanced it scores 97 over
  Seal of Righteousness.
- **Holy Shield rank 4** and the shared cooldown. The bench's rotation cast rank 3.
- **Templar's Bulwark**, **Divine Shield** and **Divine Protection**, with Sacred Duty's cooldown cut. None
  fires by itself: a rotation casts them by name (on a health condition in the UI, on a timer in scoring).
- **Shield Specialization's mana makes no threat**; **Reckoning's** block cooldown no longer holds back crits;
  **Iron Creed** from any Holy Strike; **Redoubt's** aura id; **Blessing of Sanctuary** Classic only.
- **The tank librams** (Steadfast, Sentinel's) and Libram of Law on Judgement of Fury.
- **What it costs the class to leave the health buttons out of scoring:** Templar's Bulwark on a timer (cast
  from 10 to 20 sec in, then on cooldown) is worth about 280 to 320 score points at Balanced (it's the
  difference between the rotation with and without it, 3,000 iterations, the six scoring encounters), so the
  presets keep it on that timer, which reads no health. Divine Shield, Divine Protection and Lay on Hands stay
  out: an immunity's worth depends on what the boss does to an immune tank (a call, below), and Lay on Hands
  heals in reaction to health.

## Tests

`sim/paladin/protection/tank_test.go`, a level 60 Human paladin with a sword and a shield against a level 63
boss under Forever; each checks the mechanic against the same paladin without it:

- `TestHolyShieldRank4`: cost, level, 30% block, 10 sec, 4 charges, the shared cooldown, 221 + 8% of spell
  power a block, threat 1.2 x 1.9 x 1.33, a block spends a charge and deals the damage.
- `TestSealOfFury`, `TestSealOfFuryNeedsShield`: cost, duration, a white hit's 35 + 10% of spell power at
  Righteous Fury's threat, the absorb of half of it (a boss swing loses exactly that), none with a
  two-hander, none on a miss, Judgement of Fury's range at level 60 and that Judgement judges it.
- `TestTemplarsBulwark`: cost, cooldown, no global cooldown, Forbearance, the absorb holds exactly max
  health, refused under Forbearance, Sacred Duty 2/2's 4 min and 50 sec Forbearance, nothing without the talent.
- `TestDivineShieldAndProtection`: 0 damage taken, half damage dealt and back, the shared cooldown, Divine
  Protection stops the swings and they come back.
- `TestRighteousFuryThreat`: 1.33 all threat, 1.9 on Holy Strike, a white swing's own multiplier untouched,
  Improved Righteous Fury 9% only with Righteous Fury.
- `TestProtectionStatTalents`: Anticipation's 20 Defense and its 0.8% dodge on the tank sheet, Toughness'
  10% of item armor, Redoubt's 30% block, 10 sec, gone after 5 blocks, a 10% proc rate over 20,000 swings,
  Shield Specialization's 30% of the shield's block value, Precision's melee and spell hit, Divine Intellect.
- `TestShieldSpecializationMakesNoThreat`: 4% of max mana and no threat, where 100 other mana makes 66.5.
- `TestIronCreedAndShieldsworn`: 1.3 x Holy Strike threat, the 6% cut after a missed Holy Strike, none
  without Righteous Fury, Shieldsworn's 6 sec cooldown and Holy Shield's damage without a charge.
- `TestReckoning`: a 40% chance a block at 5/5. `TestReckoningCharges` (retribution package): two crits store
  two charges, many blocks inside 1.5 sec one.
- `TestBlessingOfSanctuaryGone`.
- `TestProtection` (the class suite) and `TestRetribution` moved: Iron Creed, Reckoning and Holy Shield's
  shared cooldown move the prot suite's item runs by under 1%, and the new tank items join both suites' item
  runs. (Its settings runs fell by half until the 2026-11-04 fix of the UI rotation's Redoubt id, above.)

## Rotation

`tools/rotopt/paladin_tank.go` (`-template paladin_tank`): the seal (Fury or Righteousness) kept up and judged
on cooldown with Swift Judgement, Holy Shield on cooldown and before the pull, Holy Strike on cooldown,
Consecration above a mana floor, Templar's Bulwark on a timer, mana floors for each. Nothing reads the tank's
health. The paladin keeps none of the raid's debuffs itself (`tankscore.OwnDebuffs` unchanged): Judgement of
Fury leaves no debuff, and the raid's Judgements are other paladins'.

## Presets (step 6)

**Round 2 (2026-11-04, tank-paladin, on the engine of 5668ea5: the tank faces the boss, the logs' swing gap with
parryHaste off, health-blind survival cooldowns, Elixir of the Phalanx and Greater Fortitude on every tank).**
Before the searches: Reckoning takes spell crits (tank_mechanics-20261104-9), the trinket line is first
(tank_rotations-20261104-6), the set's Stamina share is back at its row (tank_gear-20261104-1). rotopt on the six
scoring encounters, 1,000 iterations a candidate and 3,000 to confirm, `-freeze bulwarkAt` at 10 (below):

- **Talents.** The talent search at each weight from its preset's string found only mana tie-breaks that score
  lower at 3,000: Max threat `05004-25300102413001531-50231` 2775.5 against the start's 2782.0, Balanced
  `02-25300102513501531-50232` 2384.5 against 2387.2, Max survival `01-25300102513501531-50233` 1326.9 against
  1336.9. The strings stay.
- **Rotation.** The knob search kept every default at all three weights; the APLs rotopt wrote are the
  committed ones.
- **Races** (3,000 iterations, the preset's build): Undead / Night Elf / Human / Dwarf, Max threat 2782.0 /
  2784.4 / 2681.8 / 2662.5 (Night Elf +2.4, inside the noise of about 3: a tie, Undead kept), Balanced 2387.2 /
  2353.0 / 2227.7 / 2209.0, Max survival 1336.9 / 1223.0 / 1017.6 / 997.3.
- **Templar's Bulwark at 13 sec** (tank_rotations-20261104-3; the knob now steps by 1 sec). Paired, 3,000, 13 s
  minus 10 s: on the five scoring bosses without Broodlord -1.1 ± 1.1 (Max threat), -6.2 ± 3.0 (Balanced),
  -18.5 ± 9.0 (Max survival); with Broodlord +16.5 ± 1.4 / +49.3 ± 3.9 / +147.9 ± 11.7. Broodlord alone is +105
  / +327 / +980 (13 s covers his first Mortal Strike, Knock Away and Blast Wave, timers his file calls probably
  wrong), Vaelastrasz -24 / -66 / -199, Onyxia +9 / +26 / +79. So 10 s stays: a call (below).
- **Max survival is Balanced's build** (tank_fairness-20261104-4): searched at w = 3 on its own, it lands on the
  same talents, rotation and race. The two presets are the same; Max survival keeps its own name and files.
- **What the round moved** (Balanced unless it says otherwise): the trinket first +14 at every weight (the
  audit's paired 3,000); the Stamina rebalance -12 / -27 / -72 (Max threat / Balanced / Max survival; the old
  set emulated with rotopt's `-bonus Stamina=-20,Defense=6,BlockValue=10,Strength=5,Agility=2`, 3,000, same
  seed): less block value at the paladin's ×1.6 costs more than 20 Stamina buys. Reckoning's spell crits are
  under 1% of a point.

The first round, for its history:

rotopt (`-template paladin_tank`) on the six scoring encounters, the settings of
`tools/tankbench/req/paladin_protection_balanced.json` (the synthetic set and its enchants, Steadfast Libram,
Seal of Fury, Devotion Aura, Righteous Fury, Engineering and Blacksmithing), 600 iterations a candidate, 3,000
to confirm, and 20,000 paired iterations where the confirms were inside the noise. Everything below is on the
merged code (the healer merge and `claude/clever-galileo-qim1ee`, 2026-11-04); the pre-merge bests re-evaluated
at 3,000 within 10 of their old scores (Human Balanced 2432.0 to 2422.5, Undead 2544.2 to 2543.7).

| Preset | Weight | Race | Talents | Rotation | Request, APL |
|---|---|---|---|---|---|
| Max threat | 1/3 | Undead | `05005-25300102413001531-5023` | the template's defaults | `paladin_protection_max_threat.json`, `level60_max_threat.apl.json` |
| Balanced (default) | 1 | Undead | `04-25300102513501531-5023` | the template's defaults | `paladin_protection_balanced.json`, `level60_balanced.apl.json` |
| Max survival | 3 | Undead | `04-25300102513501531-5023` | the template's defaults | `paladin_protection_max_survival.json`, `level60_max_survival.apl.json` |

- **Talents.** Balanced's search (from the pre-merge best, Undead) took three rounds to
  `03-15310102513501531-50231` (Divine Strength 5 to 4, Anticipation back, Sanctified Judgement +1, Divine
  Intellect -1, Guardian's Favor +1), each step inside its own noise; at 20,000 paired iterations it is 2538.0
  against the start's 2544.5, so the start stays. Max survival's search went to `03-25300102513501531-50231`:
  1228.7 against the start's 1232.4 at 20,000, so it keeps Balanced's string. Max threat's search from
  `05005-25300102413001531-5023` (Improved Seals 5, no Anticipation, one less Toughness) found nothing in a
  round: 3034.2, against 2981.3 for Balanced's string at that weight; Balanced's string scores 2544.5 at w = 1 to
  its 2535.3, and 1230.9 at w = 3 to its 1039.1.
- **Rotation.** The knob search at each weight kept every default (`seal=1` Fury, `resealWindow=0`,
  `holyShieldMana=0`, `holyShieldFirst=1`, `prepullHolyShield=1`, `holyStrikeMana=0`, `judgementMana=0`,
  `swiftJudgement=1`, `consecrationMana=60`, `consecrationFirst=0`, `bulwarkAt=10`); no alternative beats it by
  more than its noise (the closest, Max threat's `bulwarkAt=20`, +2.8 against a noise of 4.2). The three APLs are
  the same rotation today; each keeps its own file for the next search.
- **Races** (3,000 iterations, the preset's string; Max threat's top two at 20,000):

| Race | Max threat | Balanced | Max survival |
|---|---|---|---|
| **Undead** | **3034.2** | **2543.7** | **1230.9** |
| Night Elf | 3031.1 | 2524.5 | 1170.5 |
| Troll | 2981.6 | 2451.4 | 1028.7 |
| Orc | 2972.8 | 2444.6 | 1010.9 |
| Tauren | 2944.6 | 2447.3 | 1112.3 |
| Human | 2948.6 | 2422.5 | 998.0 |
| Gnome | 2937.8 | 2413.9 | 994.4 |
| Dwarf | 2928.6 | 2404.1 | 979.5 |
| Skyborne (either) | 2951.3 | 2417.3 | - |

  **Why Undead leads, per encounter** (Balanced, 3,000 iterations, Human / Dwarf / Undead): Magmadar 2867 / 2844
  / 3000, Golemagg 2643 / 2617 / 2764, Onyxia 2931 / 2924 / 3093, Broodlord 1817 / 1802 / 1889, Vaelastrasz 2812 /
  2793 / 2935, Chromaggus 1467 / 1446 / 1582. It leads on every encounter by about the same, fear or no fear, so
  it is not Will of the Forsaken (the scoring fights leave Panic and Bellowing Roar out anyway). It is Forever's
  Touch of the Grave (`sim/core/racials_forever.go`): every ability rolls 20% to drain up to 5% of max health, at
  most once every 15 sec, as Shadow damage that heals the tank. Against Human it is +26 to +47 SafeDPS a fight (the
  drain's damage at Righteous Fury's threat) and 3 to 23 less healing needed. The drain is modelled at its cap
  ("up to" isn't published), so this is the racial's upper bound: a call (below). Undead's 10 Shadow resistance
  and its base stats (below Human's but for Spirit) don't explain it. Night Elf is second, a tie at Max threat
  (3.1 apart, paired noise about 1.7).
- **The UI**: `ui/protection_paladin/presets.ts` and `sim.ts` carry the synthetic gear (`GearLevel60Synthetic`),
  the three APLs, the two talent strings and three builds (`makePresetBuild` with the race, the professions and
  `{primarySeal: Fury, righteousFury: true, aura: DevotionAura}`), Balanced first. The page's defaults (its
  launch state) are as they were. `npx tsc --noEmit -p .` is clean for `ui/protection_paladin`.
- `paladin_protection_valor.json` (the old Valor bench request) is removed; nothing read it.
  `TestForeverProtTankScore` reads the new Balanced request and passes as it is.

## Numbers (step 7)

`tools/tankbench -iters 1000 -encounter tools/encounters/<e>.json` for the three requests on all seven
Anniversary encounters (round 2, 2026-11-04; the rows of `orchestration/audit/ref/tank_bench/summary.tsv`).
Healing needed is at 5% deaths, in HPS; the three scores are the objective at w = 1/3, 1 and 3; the mean is over
the six scoring encounters (Ragnaros is reported, not scored).

| Preset | TPS | DPS | SafeDPS | DTPS | Healing needed | Max threat | Balanced | Max survival |
|---|---|---|---|---|---|---|---|---|
| Max threat | 1,345 | 589 | 1,494 | 320 | 238 | **2,785** | 2,380 | 1,164 |
| Balanced | 1,313 | 576 | 1,457 | 301 | 208 | 2,742 | **2,398** | 1,337 |
| Max survival | 1,313 | 576 | 1,457 | 301 | 208 | 2,742 | 2,398 | **1,337** |

Before round 2 (the master bench of 23b2ec2, the same requests before this round's changes): 2,786 / 2,411 /
1,422 at each preset's own weight.

Per encounter (Max survival is Balanced's build, so its rows equal Balanced's and are left out):

| Encounter | Preset | TPS | DPS | SafeDPS | DTPS | Healing needed | Death floor | Max threat | Balanced | Max survival |
|---|---|---|---|---|---|---|---|---|---|---|
| Magmadar | Max threat | 1331 | 584 | 1453 | 163 | 30 | 0.0% | 2872 | 2802 | 2594 |
| Magmadar | Balanced | 1296 | 569 | 1414 | 150 | 14 | 0.0% | 2811 | 2777 | 2676 |
| Golemagg | Max threat | 1304 | 572 | 1420 | 201 | 84 | 0.0% | 2743 | 2549 | 1966 |
| Golemagg | Balanced | 1276 | 560 | 1388 | 186 | 62 | 0.0% | 2705 | 2561 | 2128 |
| Onyxia | Max threat | 1399 | 614 | 1557 | 178 | 34 | 0.0% | 3056 | 2939 | 2586 |
| Onyxia | Balanced | 1366 | 600 | 1524 | 165 | 13 | 0.0% | 3025 | 2979 | 2842 |
| Broodlord | Max threat | 1327 | 581 | 1460 | 321 | 364 | 0.0% | 2555 | 1828 | -356 |
| Broodlord | Balanced | 1294 | 567 | 1422 | 302 | 307 | 0.0% | 2538 | 1924 | 83 |
| Vaelastrasz | Max threat | 1413 | 618 | 1682 | 626 | 435 | 0.0% | 3182 | 2816 | 1720 |
| Vaelastrasz | Balanced | 1380 | 604 | 1638 | 604 | 403 | 0.0% | 3107 | 2769 | 1753 |
| Chromaggus | Max threat | 1294 | 567 | 1390 | 429 | 484 | 0.0% | 2302 | 1344 | -1529 |
| Chromaggus | Balanced | 1267 | 556 | 1358 | 400 | 451 | 0.0% | 2269 | 1376 | -1302 |
| Ragnaros | Max threat | 1172 | 518 | 1270 | 953 | 1609 | 0.0% | 847 | -2540 | -12700 |
| Ragnaros | Balanced | 1142 | 506 | 1235 | 948 | 1626 | 0.0% | 759 | -2665 | -12936 |

Max threat buys 37 SafeDPS (32 TPS) for 30 more healing needed: Improved Seals over Anticipation's 20 Defense.

**Next to the other tanks** (Balanced preset, mean of the six, 1,000 iterations, from their docs on this branch):

| Tank | SafeDPS | DTPS | Healing needed | Max threat | Balanced | Max survival |
|---|---|---|---|---|---|---|
| Protection paladin (round 2) | 1,457 | 301 | 208 | 2,742 | 2,398 | 1,363 |
| Feral tank (`docs/tank_druid.md`) | 1,714 | 570 | 582 | 2,904 | 1,855 | -1,292 |
| Warden shaman (`docs/tank_shaman.md`) | 1,745 | 816 | 1,055 | 2,493 | 501 | -5,475 |
| Protection warrior | no committed numbers yet | | | | | |

## Calls for the user

Each with the default this work uses and what the alternatives measure (Balanced, the mean of the six scoring
encounters, unless it says otherwise).

1. **Seal of Fury's effect 0 and Judgement of Fury's effect 2.** The client gives both a dummy effect of 1607 +
   42 a level with 9% / 18% of spell power, and no tooltip says what it does. **Default: not read** (the seal's
   35 Holy a hit, its absorb and the Judgement's damage are what the tooltips name, and what wowsims/forever
   models). Not measured: no source says whether it is damage, an absorb cap or a server-side script, so there is
   nothing to model. If it is damage it would be the seal's largest part; an in-game log of a Judgement of Fury
   would settle it.
2. **The immunities out of scoring.** Divine Shield, Divine Protection and Lay on Hands are built, tested and in
   the UI, but no preset casts them: an immune or pacified tank keeps no threat and the boss turns to someone
   else, which the tank score doesn't model (it has one tank), and Lay on Hands reacts to health. **Ruled
   out** (user, 2026-11-04: a bubble loses the paladin its aggro; standards 40). No paladin tank rotation casts
   Divine Shield, Divine Protection or Blessing of Protection, and `no_bubble_test.go` fails if one does. Templar's Bulwark stays in on a timer (`bulwarkAt=10`, worth 280 to 320 at Balanced), and it shares
   Forbearance with them, so casting either would cost it.
3. **Pursuit of Justice** (movement speed). **Default: not taken**; the score has no movement, so it is worth 0 in
   every preset. A call only if movement matters on a fight the bench doesn't model.
4. **Hit to the cap, and the Stamina it no longer takes.** The synthetic set's hit is sized to the 8% cap counting
   Precision (standards 4), +29 at Balanced over the row's 3.2%. Round 2 moved what pays for it from Stamina to
   the other stats (tank_gear-20261104-1, gear_targets.md's hit-adjusted rule): Stamina 32.6% against the row's
   32.4. **Default: the row**, which costs 12 / 27 / 72 (Max threat / Balanced / Max survival) against the old
   set: block value at the paladin's ×1.6 is worth more to the score than Stamina.
5. **Templar's Bulwark is Magic** (dispellable). The client carries it since 70753 (SpellCategories DispelType 1
   on 1311015; round 70753's `paladin.md` and foreverchanges.pro list it as notes-only). **Default: as the client
   says, and it moves nothing**: no scoring boss dispels, and the sim's auras have no dispel type to set. The
   round file's note could be corrected to "client and notes".
6. **Undead over Human and Dwarf.** Undead leads at every weight, through Touch of the Grave modelled at its cap
   (5% of max health every 15 sec at the most, "up to" with no published amount; Presets, above). **Default:
   Undead** (+121 over Human at Balanced, +140 over Dwarf, +19 over Night Elf; +86 / +105 / +3 at Max threat;
   +233 / +251 / +60 at Max survival). If the drain is well under its cap the lead shrinks, and Night Elf (at
   Max survival perhaps Tauren's 5% health) may be the pick: Undead's base stats are below Human's but for Spirit.
   Not measured with a smaller drain (it needs a sim change to the shared racial). An in-game log
   of the drain's amount would settle it.
7. **Templar's Bulwark at 13 sec instead of 10** (tank_rotations-20261104-3). 13 s wins on all six scoring bosses
   (+16.5 ± 1.4 / +49.3 ± 3.9 / +147.9 ± 11.7, Max threat / Balanced / Max survival, paired, 3,000) but loses
   on the five without Broodlord (-1.1 ± 1.1 / -6.2 ± 3.0 / -18.5 ± 9.0). **Default: 10 s**, since the gain
   is Broodlord's first-ability timers, which his file calls probably wrong. Take 13 s if those timers are
   confirmed (a log of the pull).
8. **Templar's Bulwark on a timer every pull** (the core worker, tank_fairness-20261104-2 (b)). It absorbs 100%
   of max health for 8 sec once a fight, worth about 38 DTPS on a 53 sec encounter (Magmadar / Golemagg: +38 /
   +31 DTPS with it off). **Default: in**, at 10 s. The alternative: a scored preset keeps no emergency button
   on a timer (the line removed); the presets would lose that much.
9. **Block value's price for the paladin** (tank_fairness-20261104-2 (a)). Standards 36 prices block value at 0.4
   a point for every tank; the paladin gets ×1.6 from it (Shield Specialization and Steadfast), so its set
   carries 72 more points than the warrior's for the same budget (+32 / +30 DTPS on Magmadar / Golemagg). The set
   follows its row (6.7% against 6.8% hit-adjusted). **Default: as standards 36 has it.** The alternative, a
   class price of 0.64, would move about 50 points of block value into the other stats.
10. **Steadfast Libram's +30 block value** (tank_fairness-20261104-2 (c)). It rests on the in-game capture of the
    item (BayovCore ItemsSeen, client 70450, `assets/db_inputs/forever_seen_in_game.json:79`); Wowhead lists no
    stats for it and nothing newer exists. **Default: +30.** A look at its tooltip on the current build would
    confirm it (+14 / +13 DTPS on Magmadar / Golemagg with it off).
