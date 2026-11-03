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
  shared cooldown move the prot suite's DPS by under 1%, and the new tank items join both suites' item runs.

## Rotation

`tools/rotopt/paladin_tank.go` (`-template paladin_tank`): the seal (Fury or Righteousness) kept up and judged
on cooldown with Swift Judgement, Holy Shield on cooldown and before the pull, Holy Strike on cooldown,
Consecration above a mana floor, Templar's Bulwark on a timer, mana floors for each. Nothing reads the tank's
health. The paladin keeps none of the raid's debuffs itself (`tankscore.OwnDebuffs` unchanged): Judgement of
Fury leaves no debuff, and the raid's Judgements are other paladins'.

PRESETS_PLACEHOLDER

## Calls for the user

CALLS_PLACEHOLDER
