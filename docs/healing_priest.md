# Healing priest

The Holy and Discipline priest of the healing sim (`HealingPriest`, `sim/priest/healing`, H3 in
`docs/healing_sim.md`), built on the Restoration shaman's model (H2). Its heals, shields and healing
talents are the priest's (`sim/priest/healing.go`, `healing_talents.go`, Penance's heal in `penance.go`),
its shields go through the absorb pipeline in `sim/core/absorb.go` (below), and its tests are
`sim/raid_healing_priest_test.go`.

Client: 1.60.1.70753 (the November 2 build, which carries the October 30 launch-realm hotfix), with
Blizzard's development notes post 9 (2026-11-02) and the round's rulings (`orchestration/round_70753/`,
standards 37 to 39). Downranking is sim/core's (`sim/core/downrank.go`, every ranked spell under Forever).

## The rules audit

### Heals: every rank against the client

`orchestration/tools/healers/priest_audit.py` reads every rank of every heal in `sim/priest/healing.go` and
compares it with the client's tables (wago.tools' cache of 1.60.1.70753): the heal effect's base points, growth
a level and variance (SpellEffect: the heal is effect 10, Renew's and the shield's an aura effect 6, Prayer of
Mending's the dummy effect 3 that carries its heal), the coefficient (EffectBonusCoefficient), the learn and
max levels (SpellLevels), the cast time (SpellMisc -> SpellCastTimes) and the mana cost (SpellPower).
**All 424 fields are the client's** (0 differ). A rank heals its base at its learn level plus its growth a level
up to its max level, spread evenly by the variance (`healRank.At`), as the shaman's heals do. The client's
tables, which the sim's are:

#### Lesser Heal

| rank | spell | base | per level | variance | coefficient | learned | max level | cast ms | mana |
|---|---|---|---|---|---|---|---|---|---|
| 1 | 2050 Lesser Heal | 51 | 0.3 | 0.1961 | 0.429 | 1 | 3 | 1500 | 30 |
| 2 | 2052 Lesser Heal | 94 | 0.7 | 0.1795 | 0.571 | 4 | 9 | 2000 | 45 |
| 3 | 2053 Lesser Heal | 163 | 1.2 | 0.1507 | 0.714 | 10 | 15 | 2500 | 75 |

#### Heal

| rank | spell | base | per level | variance | coefficient | learned | max level | cast ms | mana |
|---|---|---|---|---|---|---|---|---|---|
| 1 | 2054 Heal | 270 | 2 | 0.1447 | 0.857 | 16 | 21 | 3000 | 155 |
| 2 | 2055 Heal | 375 | 2.9 | 0.1348 | 0.857 | 22 | 27 | 3000 | 205 |
| 3 | 6063 Heal | 523 | 4 | 0.1258 | 0.857 | 28 | 33 | 3000 | 255 |
| 4 | 6064 Heal | 651 | 4.5 | 0.1214 | 0.857 | 34 | 39 | 3000 | 305 |

#### Greater Heal

| rank | spell | base | per level | variance | coefficient | learned | max level | cast ms | mana |
|---|---|---|---|---|---|---|---|---|---|
| 1 | 2060 Greater Heal | 844 | 5.1 | 0.1192 | 0.857 | 40 | 45 | 3000 | 370 |
| 2 | 10963 Greater Heal | 1099 | 5.8 | 0.1148 | 0.857 | 46 | 51 | 3000 | 455 |
| 3 | 10964 Greater Heal | 1403 | 6.6 | 0.1129 | 0.857 | 52 | 57 | 3000 | 545 |
| 4 | 10965 Greater Heal | 1782 | 7.5 | 0.1094 | 0.857 | 58 | 63 | 3000 | 655 |
| 5 | 25314 Greater Heal | 1960 | 8.1 | 0.1096 | 0.857 | 60 | 65 | 3000 | 710 |

#### Flash Heal

| rank | spell | base | per level | variance | coefficient | learned | max level | cast ms | mana |
|---|---|---|---|---|---|---|---|---|---|
| 1 | 2061 Flash Heal | 194 | 1.9 | 0.2047 | 0.429 | 20 | 25 | 1500 | 125 |
| 2 | 9472 Flash Heal | 249 | 2.2 | 0.1958 | 0.429 | 26 | 31 | 1500 | 155 |
| 3 | 9473 Flash Heal | 312 | 2.5 | 0.1833 | 0.429 | 32 | 37 | 1500 | 185 |
| 4 | 9474 Flash Heal | 383 | 2.8 | 0.1777 | 0.429 | 38 | 43 | 1500 | 215 |
| 5 | 10915 Flash Heal | 507 | 3.3 | 0.1728 | 0.429 | 44 | 49 | 1500 | 265 |
| 6 | 10916 Flash Heal | 644 | 3.7 | 0.1705 | 0.429 | 50 | 55 | 1500 | 315 |
| 7 | 10917 Flash Heal | 825 | 4.2 | 0.165 | 0.429 | 56 | 61 | 1500 | 380 |

#### Prayer of Healing

| rank | spell | base | per level | variance | coefficient | learned | max level | cast ms | mana |
|---|---|---|---|---|---|---|---|---|---|
| 1 | 596 Prayer of Healing | 178 | 0.8 | 0.06431 | 0.286 | 30 | 39 | 3000 | 410 |
| 2 | 996 Prayer of Healing | 265 | 1 | 0.06114 | 0.286 | 40 | 49 | 3000 | 560 |
| 3 | 10960 Prayer of Healing | 401 | 1.3 | 0.05621 | 0.286 | 50 | 59 | 3000 | 770 |
| 4 | 10961 Prayer of Healing | 583 | 1.5 | 0.05389 | 0.286 | 60 | 69 | 3000 | 1030 |
| 5 | 25316 Prayer of Healing | 649 | 1.6 | 0.05421 | 0.286 | 60 | 69 | 3000 | 1070 |

#### Holy Nova (its heal spells, 23455 to 27805)

| rank | spell | base | per level | variance | coefficient | learned | max level | cast ms | mana |
|---|---|---|---|---|---|---|---|---|---|
| 1 | 15237 Holy Nova | 51 | 0.4 | 0.1429 | 0.107 | 20 | 26 | 0 | 185 |
| 2 | 15430 Holy Nova | 82 | 0.5 | 0.1304 | 0.107 | 28 | 34 | 0 | 290 |
| 3 | 15431 Holy Nova | 116 | 0.6 | 0.1385 | 0.107 | 36 | 42 | 0 | 400 |
| 4 | 27799 Holy Nova | 159 | 0.7 | 0.1494 | 0.107 | 44 | 50 | 0 | 520 |
| 5 | 27800 Holy Nova | 238 | 0.8 | 0.1423 | 0.107 | 52 | 58 | 0 | 635 |
| 6 | 27801 Holy Nova | 311 | 0.9 | 0.1472 | 0.107 | 60 | 66 | 0 | 750 |

#### Renew (a tick)

| rank | spell | base | per level | variance | coefficient | learned | max level | cast ms | mana |
|---|---|---|---|---|---|---|---|---|---|
| 1 | 139 Renew | 9 | 0 | 0 | 0.2 | 8 | 13 | 0 | 30 |
| 2 | 6074 Renew | 15 | 0 | 0 | 0.2 | 14 | 19 | 0 | 65 |
| 3 | 6075 Renew | 25 | 0 | 0 | 0.2 | 20 | 25 | 0 | 105 |
| 4 | 6076 Renew | 32 | 0 | 0 | 0.2 | 26 | 31 | 0 | 140 |
| 5 | 6077 Renew | 41 | 0 | 0 | 0.2 | 32 | 37 | 0 | 170 |
| 6 | 6078 Renew | 54 | 0 | 0 | 0.2 | 38 | 43 | 0 | 205 |
| 7 | 10927 Renew | 74 | 0 | 0 | 0.2 | 44 | 49 | 0 | 250 |
| 8 | 10928 Renew | 102 | 0 | 0 | 0.2 | 50 | 55 | 0 | 305 |
| 9 | 10929 Renew | 134 | 0 | 0 | 0.2 | 56 | 61 | 0 | 365 |
| 10 | 25315 Renew | 166 | 0 | 0 | 0.2 | 60 | 65 | 0 | 410 |

#### Power Word: Shield (absorb)

| rank | spell | base | per level | variance | coefficient | learned | max level | cast ms | mana |
|---|---|---|---|---|---|---|---|---|---|
| 1 | 17 Power Word: Shield | 44 | 0.8 | 0 | 0.1 | 6 | 11 | 0 | 45 |
| 2 | 592 Power Word: Shield | 86 | 1.2 | 0 | 0.1 | 12 | 17 | 0 | 80 |
| 3 | 600 Power Word: Shield | 154 | 1.6 | 0 | 0.1 | 18 | 23 | 0 | 130 |
| 4 | 3747 Power Word: Shield | 226 | 2 | 0 | 0.1 | 24 | 29 | 0 | 175 |
| 5 | 6065 Power Word: Shield | 291 | 2.3 | 0 | 0.1 | 30 | 35 | 0 | 210 |
| 6 | 6066 Power Word: Shield | 368 | 2.6 | 0 | 0.1 | 36 | 41 | 0 | 250 |
| 7 | 10898 Power Word: Shield | 470 | 3 | 0 | 0.1 | 42 | 47 | 0 | 300 |
| 8 | 10899 Power Word: Shield | 591 | 3.4 | 0 | 0.1 | 48 | 53 | 0 | 355 |
| 9 | 10900 Power Word: Shield | 749 | 3.9 | 0 | 0.1 | 54 | 59 | 0 | 425 |
| 10 | 10901 Power Word: Shield | 928 | 4.3 | 0 | 0.1 | 60 | 65 | 0 | 500 |

#### Prayer of Mending (each jump's heal)

| rank | spell | base | per level | variance | coefficient | learned | max level | cast ms | mana |
|---|---|---|---|---|---|---|---|---|---|
| 1 | 401859 Prayer of Mending | 172 | 0 | 0 | 0.429 | 40 | 40 | 0 | 210 |
| 2 | 1240826 Prayer of Mending | 298 | 0 | 0 | 0.429 | 50 | 50 | 0 | 305 |
| 3 | 1240827 Prayer of Mending | 413 | 0 | 0 | 0.429 | 60 | 60 | 0 | 390 |

Penance's healing bolts (`penance.go`; 402289, 1240723, 1240724, 1316991): 126 + 0.87/level (30 to 39), 200 +
0.98 (40 to 49), 268 + 1.27 (50 to 59), 425 (60), no variance, coefficient 0.19 a bolt, 3 bolts, the cast's
mana and 12 s cooldown (client). The sim had only the damage bolts; the channel now fires the heal bolts at a
friendly target.

Other heal rules (client 70753 unless said):

| Rule | Client / source | Sim |
|---|---|---|
| Renew: 5 ticks, 3 s apart (15 s) | SpellMisc DurationIndex, EffectAuraPeriod 3000 | `RenewTicks`, `RenewTickLength`; a new rank replaces another rank's Renew of ours |
| Power Word: Shield: 30 s, 4 s cooldown (Soul Warding takes it off), Weakened Soul 15 s (6788) | client | as client |
| Power Word: Shield absorb: the rank's base + 10% of healing power (coefficient 0.1, downranked), Improved PW:S 7/14/20% | client | as client |
| Power Word: Shield replaced only by an equal or higher rank | notes ("can now correctly be replaced by a Power Word: Shield of equal or greater rank"); no client field | `ExtraCastCondition`: a lower rank can't be cast over our shield still up. Test `TestHealingPriestShieldReplaceRank` |
| Prayer of Mending: 5 charges, 30 s, 10 s cooldown, jumps on damage taken to a random raid member without it | client (401859 auras; heal 401863) | `registerPrayerOfMendingSpell` |
| Prayer of Healing: its target's party | deep dive 2026-09-30, ImplicitTarget 132/31 | `partyOf(target)` |
| Holy Nova: heals the priest's party (the damage part is not simulated) | client 23455 etc. | `partyOf(priest)` |
| Binding Heal removed | notes ("Binding Heal removed."); client 70753 still has its rows in SkillLineAbility | gone: no rank registered. Test in `TestHealingPriestRanks` |
| Downranking: a rank keeps 1 - 5% a level past 17 of its coefficient | sim/core/downrank.go (user ruling 2026-11-04: every ranked spell) | each heal sets `Rank`/`RequiredLevel`; Prayer of Mending's heal and Penance's heal bolts their rank's `DownrankLevel`. Test in `TestHealingPriestRanks` |
| Power Infusion: 20% more healing as well as damage | client 10060 auras 79 and 136 | `core.PowerInfusionAura` also multiplies `HealingDealtMultiplier` (sim/core/buffs.go) |

### Healing talents

Values are the client's TraitDefinitionEffectPoints curves and the spells each touches its SpellEffect class
masks. Forever's percent spell modifiers multiply each other (as the cost ones do, `reduceCost`).

| Talent | Client 70753 | Sim before | Sim now | Test |
|---|---|---|---|---|
| Spiritual Healing (14898) | 3/7/10% healing, every heal but PW:S (effect 0 the direct heals and Penance, effect 1 Renew) | missing | as client | TalentValues |
| Improved Renew | 5% a rank Renew | missing | as client | TalentValues |
| Improved Healing (14912) | 5% a rank off Lesser Heal, Heal, Greater Heal, Penance, Prayer of Mending | missing | as client | TalentValues |
| Holy Specialization (14889) | 1% crit a rank: Smite, Holy Fire, Penance and every heal (Renew's ticks too), not PW:S | Smite, Holy Fire, Penance only | as client | TalentValues |
| Divine Fury | 0.1 s a rank off Heal and Greater Heal (and Smite, Holy Fire) | damage spells only | as client | TalentValues |
| Spiritual Guidance (14901) | 5% of Spirit a rank as spell damage and as healing (aura 175) | damage only | both | TalentValues |
| Twin Disciplines (1225132) | 1% a rank: Holy Nova, Prayer of Mending, Penance's heal bolts (direct mask), Renew (periodic mask); not PW:S | damage only | as client | TalentValues |
| Mental Agility | 3/7/10% off the instants in its mask: of the heals Renew, PW:S, Holy Nova, Prayer of Mending | damage spells only | as client | TalentValues |
| Twilight Focus (14913) | 23/47/70% pushback protection on the heals with a cast | damage spells only | as client | (pushback only) |
| Improved Power Word: Shield | 7/14/20% absorb | missing | as client | TalentValues |
| Soul Warding | PW:S no cooldown, 15% cost | missing | as client | TalentValues |
| Meditation | 17/33/50% Spirit regen while casting | in | unchanged | TalentValues |
| Inner Focus | next spell free, 25% crit if not periodic | in | heals too | InnerFocusAndPowerInfusion |
| Power Infusion | 20% damage and healing, 15 s | damage only | both (client 10060) | InnerFocusAndPowerInfusion |
| Renewed Hope (425280) | 2% crit a rank for Flash Heal, Lesser Heal, Heal, Greater Heal and Penance on a target with Weakened Soul, and each cuts its Weakened Soul by 1 s a rank | missing | as client | RenewedHope |
| Divine Aegis (431622, 431624) | a crit heal (not a tick) shields its target for 5% a rank of the heal, 12 s, up to 20 a level | missing | as client; a new crit adds to the shield, to the limit (Calls for the user) | DivineAegis |
| Inspiration (14892, 14893) | a crit heal (not a tick) gives 8/17/25% armor for 15 s | missing | as client, one buff a player | Inspiration |
| Litany of Light (1317006) | a heal whose name isn't the last heal's gives back 5/10% of its base cost | missing | as client; the fight's first heal counts | LitanyOfLight |
| Holy Binding (1328569, 1328570) | in Binding Heal's slot since the notes: a Flash Heal on another heals the priest for 50% of it (dummy aura, proc type 16384, a helpful spell done; the tooltip names Flash Heal), no threat | n/a | as client | HolyBinding |
| Blessed Recovery (27811, 27813) | a melee or ranged crit taken, or a hit over 30% of max health, heals 8/17/25% of it over 6 s (3 ticks); a new one carries the old one's rest | missing | as client | BlessedRecovery |
| Spirit of Redemption | | not simulated | not simulated (a dead priest heals; the sim's healers don't die as a rule) | |
| Holy Reach, Spell Warding, Searing Light, Silent Resolve, Martyrdom | range, resistance, damage, threat, pushback on crit | | not healing | |

## The absorb pipeline (sim/core/absorb.go)

Power Word: Shield and Divine Aegis are absorbs: `Unit.NewAbsorbArray(aura)` gives the caster a shield on
every friendly unit (its ranks share one; a new one replaces the old), `Absorb.Apply` puts one up for an amount,
and `Absorb.Add` tops one up to a limit (Divine Aegis).

**What it does to a hit, exactly.** Damage an opponent deals to a unit with shields up goes through them, oldest
first, as the **last step of the damage calculation**: `ApplyPostOutcomeDamageModifiers` (sim/core/spell_result.go)
ends with `result.Target.absorbDamage(sim, spell, result)`, after armor, resists, damage-taken modifiers, the
outcome and the older absorbs (Ice Barrier and the like). Each shield takes `min(its rest, result.Damage)` off
`result.Damage`. So everything after it sees only the damage the shield let through:

- **damageTaken**: the tank score's replay records `result.Damage` when the hit lands (`OnSpellHitTaken` /
  `OnPeriodicDamageTaken` -> `RemoveHealth` -> `tankScore.damageTaken`), which is already the post-absorb amount.
  An absorbed hit is recorded as a smaller hit (0 when the shield takes it whole). `sim/core/tank_score.go` is
  unchanged.
- **GainHealth**: never called. An absorb heals nothing; health simply doesn't go down.
- **healTaken**: untouched. Nothing is recorded as healing taken, so the replay counts the shield once (as
  less damage), never also as healing.
- DTPS and the death check read the same reduced damage.

**What it books** (the brief's default: an absorb is effective healing while it absorbs, its unused part
overhealing): what a shield absorbs is booked on its spell's metrics against the target as `TotalHealing`
(effective). What it never absorbs (it ran out of time, a new shield replaced it, or the target died) is booked
then as `TotalHealing` and `TotalOverhealing`. A shield still up when the fight ends books nothing for its rest.
Shields are not booked as Shielding, so HPS counts them once and EHPS counts what they absorbed.

Tests: `TestHealingPriestPowerWordShield` (the shields' effective healing equals the DTPS they took off,
to 1e-3), `TestHealingPriestShieldUnused` (unused shields are all overheal, no EHPS),
`TestHealingPriestDivineAegis` (its share and limit), and `TestHealingPriestShieldTankScore`: a shielded tank
with tank scoring and no health tracking, 300 iterations; the replay's deaths equal its chance of death, shielded
or not (24.0% shielded against 65.3% without), so the score reads the shields only through damageTaken.

## The presets

Three builds, all on the healing priest synthetic set (990501 to 990514, the healing priest row of
`orchestration/gear_targets.md`), the Restoration shaman's consumables, raid buffs and debuffs, and level 60.
The set's wand (990514, slot 16) carries no stats since 2026-11-04 (healer_fairness-20261104-3): it had 6
Intellect, 10 spell power and 15 healing, 21.4 weighted points (the wand budget of standards 1b), on top of the
902 the four healers' armor spends, while the shaman, paladin and druid fill slot 16 with a stat-less relic.
That put the priest 2.4% over the others; the stats were worth 0.003 carry.

| Build | Talents | Rotation | Knobs (`priest_tuning/presets.sh`) |
|---|---|---|---|
| Holy tank healer (the default) | Holy `00500503113-03505103030121531` | `apls/holy_tank_healer.apl.json` | `fhfirst=1 pws=0 fh=7,4 emergency=0.5 gate=0.7` |
| Holy raid healer | Holy | `apls/holy_raid_healer.apl.json` | `pws=0.6 order=fh,poh,renew,gh renewshare=0.6 fh=7,4` |
| Discipline tank healer | Discipline `005005031330011531-0350510303` | `apls/discipline_tank_healer.apl.json` | `fhfirst=1 fh=7,4 emergency=0.5 gate=0.7 pwsgate=0.5 pwsneed=1.0` |

The rotations are written by `orchestration/tools/healers/priest_tuning/rots.py` (run `presets.sh` from the repo
root; it also writes `tools/healbench/req/priest_holy.json`, `priest_holy_raid.json` and
`priest_discipline.json`, which are the shaman's request with the priest in its slot). Every rotation casts
Flash Heal 7 at whoever in its assignment drops under the emergency line and binds at most 4 ranks of a heal,
each cast once what's missing covers its Expected Healing (or the `share=` part of it, the heal gate). The
tank healers cast Inner Focus right before the biggest heal; the Discipline tank healer also casts Power
Infusion on the priest when a big heal is due. The Holy build has no Power Infusion (it's the Discipline tree's
31-point talent and Holy spends 18 there), so the Holy rotations carry no Power Infusion line
(healer_rotations-20261104-1, healer_fairness-20261104-3; Power Infusion builds under "Talents").

- **Holy tank**: Prayer of Mending and Renew on the tank whenever they're off, Flash Heal 7 and 4 first while
  mana is over 70%, then Inner Focus, Greater Heal 5, 4, 2 and Flash Heal 7, 4.
- **Holy raid**: Prayer of Mending at the most hurt whenever it's ready, a shield on whoever drops under 60%,
  then Flash Heal 7, 4 at the most hurt. That's all it casts: until 2026-11-04 the rotation went on with
  Prayer of Healing 5, Renew, Inner Focus and Greater Heal 5, 2, but Flash Heal 7 and 4 take every hole those
  would, so they never fired (0 casts in logged iterations on five bosses, Inner Focus once), and moving any of
  them ahead of Flash Heal loses (Prayer of Healing first -0.094 mean carry, Renew first -0.182, Inner Focus
  first -0.003; healer_rotations-20261104-1). They're dropped; `rots.py`'s `order=` knob still writes them.
- **Discipline tank**: Power Word: Shield whenever Weakened Soul is off the tank (under 50% mana only once
  what's missing covers a Greater Heal 5), Penance once what's missing covers its 3 bolts, Renew, then the Holy
  tank healer's Flash Heal and Greater Heal.

### Tuning (healbench, damage carried, 300 iterations)

The score is the damage multiplier the healer carries with no death (`-carry`). Onyxia and Broodlord at their
files' lengths, and Onyxia stretched to 180 and 300 s for the long-fight threshold.

Holy tank (`ht0` to `ht3`: the gate on Flash Heal first, and a shield on the tank or not):

| Variant | Onyxia | Onyxia 180 s | Onyxia 300 s | Broodlord | Ragnaros |
|---|---|---|---|---|---|
| shield, gate 70% | 1.244 | 1.097 | 0.904 | 0.866 | 0.578 |
| shield, gate 85% | 1.250 | 1.112 | 0.922 | 0.855 | 0.562 |
| shield, gate 55% | 1.244 | 1.077 | 0.887 | 0.866 | 0.578 |
| **no shield, gate 70% (shipped)** | 1.234 | **1.138** | **0.961** | **0.891** | **0.594** |

The shield costs more mana on the tank than it saves: it's 0.01 ahead on Onyxia's 35 s and behind everywhere
longer.

Discipline tank (the shield's gate and need, Flash Heal first or not):

| Variant | Onyxia | Onyxia 180 s | Onyxia 300 s | Broodlord |
|---|---|---|---|---|
| shield always, Flash Heal first | 1.219 | 0.876 | 0.696 | 0.876 |
| **under 50% mana, shield once a Greater Heal 5 is missing (shipped)** | 1.219 | **0.898** | **0.743** | 0.876 |
| under 70% mana, once half a Greater Heal 5 is missing | 1.219 | 0.898 | 0.722 | 0.876 |
| no Flash Heal first, shield only at full mana or a Greater Heal missing | 1.147 | | 0.818 | 0.833 |
| the same with Flash Heal first | 1.198 | | 0.776 | 0.836 |

Shielding less trades the short fights for the long ones; the shipped one loses nothing short and keeps 0.74
at 300 s. Every Discipline variant OOMs by 140 to 170 s at carried damage.

Holy raid (raid role: 2 healers and 7 Frost mages):

| Variant | Onyxia | Broodlord | Onyxia 300 s |
|---|---|---|---|
| Flash Heal 7, 5, 3 first, shield under 60%, Renew at 60% | 0.716 | 0.834 | 0.561 |
| the same with no shield | 0.664 | 0.777 | 0.564 |
| Prayer of Healing first | 0.650 | 0.815 | 0.555 |
| **Flash Heal 7, 4 first (shipped)** | 0.714 | **0.883** | **0.594** |
| emergency line at 50% | 0.703 | 0.831 | 0.566 |
| the shaman raid healer | 0.693 | 0.869 | 0.558 |

### Talents

Holy (18 Discipline for Inner Focus and Meditation, 33 Holy for Spiritual Guidance and Healing, Holy Binding,
Prayer of Mending and Litany of Light) carries the most on the tank. On the Discipline tank healer's
rotation, against the Discipline build (Onyxia / Broodlord / Onyxia 300 s, 300 iterations):

| Talents | Onyxia | Broodlord | Onyxia 300 s |
|---|---|---|---|
| Discipline `005005031330011531-0350510303` | 1.147 | 0.833 | 0.818 |
| Penance with 30 in Holy `005005031330001-0350510303010153` | 1.137 | 0.807 | 0.863 |
| Discipline 17, Holy 12, no Power Infusion `00500503133001153-035051030301` | 1.161 | 0.833 | 0.812 |

(The shield there only at full mana or a Greater Heal missing, no Flash Heal first.) The hybrids move the
numbers by a few hundredths either way; none reaches the Holy tank healer (1.234, 0.891, 0.961), so Holy is the
default and Discipline the second preset. The strings need no remap for 70753 (`orchestration/talent_remap_70753.md`:
the priest's tree is the one the sim took as 70707+hf).

### Races

On the Holy tank healer (300 iterations): Human 1.232 / 0.859, Dwarf 1.227 / 0.859, Night Elf 1.238 / 0.859,
Gnome 1.227 / 0.849, Undead 1.228 / 0.859, Troll 1.225 / 0.870 (Onyxia / Broodlord). All within 0.02, so the
presets take Human (Alliance) and Undead (Horde).

### Downranking in sim/core

The heals now take their coefficients from the engine (`sim/core/downrank.go`, each heal's `Rank` and
`RequiredLevel`, Prayer of Mending's and Penance's heals a `DownrankLevel`); the priest's own
`downrankCoefficient` is gone. Measured on the tank presets of the time (healbench, 500 iterations, damage
carried), before and after the merge:

| | Onyxia | Broodlord | Magmadar | Ragnaros |
|---|---|---|---|---|
| Holy tank, before | 906 HPS, 1.242 | 931, 0.866 | 713, 2.075 | 729, 0.578 |
| Holy tank, after | 906 HPS, 1.244 | 931, 0.866 | 713, 2.075 | 730, 0.578 |
| Discipline tank, before | 594 HPS, 1.203 | 600, 0.876 | 479, 1.938 | 362, 0.563 |
| Discipline tank, after | 594 HPS, 1.219 | 600, 0.876 | 479, 1.938 | 362, 0.563 |

No change: the formula is the one the priest had (1 - 5% a level past 17 of the rank's learn level, from
level 60) and the bound ranks are high enough to keep most of their coefficient. The 0.016 on Discipline's
Onyxia is the merge's other changes (Human Spirit), not the coefficients.

### The 7 boss files, 1,000 iterations, against the shaman

`tools/healbench`, damage carried with no death, 1,000 iterations a sim. Tank role: the priest's two tank
healers and the Restoration shaman's (`shaman_restoration.json`). Raid role: the Holy raid healer and the
shaman's (`shaman_restoration_raid.json`). "Damage carried" is the score; the other columns are at the file's own
damage (×1.00), and an OOM time in brackets is a projection past the fight's end.

| Boss | Role | Healer | Damage carried | EHPS | Overheal | EH per mana | OOM at |
|---|---|---|---|---|---|---|---|
| Onyxia | tank | Holy tank (priest) | 1.230 | 493 | 46.0% | 4.76 | (308 s) |
| Onyxia | tank | Discipline tank (priest) | 1.195 | 519 | 12.4% | 3.98 | (215 s) |
| Onyxia | tank | Restoration tank (shaman) | 1.017 | 509 | 12.2% | 5.76 | (421 s) |
| Onyxia | raid | Holy raid (priest) | 0.709 | 355 | 33.5% | 3.53 | (953 s) |
| Onyxia | raid | Restoration raid (shaman) | 0.692 | 350 | 38.8% | 4.67 | (1317 s) |
| Broodlord | tank | Holy tank (priest) | 0.891 | 538 | 44.4% | 4.62 | (423 s) |
| Broodlord | tank | Discipline tank (priest) | 0.852 | 515 | 12.9% | 3.75 | (332 s) |
| Broodlord | tank | Restoration tank (shaman) | 0.618 | 390 | 13.1% | 5.59 | (1479 s) |
| Broodlord | raid | Holy raid (priest) | 0.879 | 445 | 35.8% | 3.07 | (712 s) |
| Broodlord | raid | Restoration raid (shaman) | 0.835 | 455 | 41.0% | 4.20 | (1319 s) |
| Chromaggus | tank | Holy tank (priest) | 1.006 | 679 | 41.3% | 5.48 | (215 s) |
| Chromaggus | tank | Discipline tank (priest) | 0.824 | 598 | 10.4% | 4.42 | 106 s |
| Chromaggus | tank | Restoration tank (shaman) | 0.764 | 487 | 10.6% | 5.97 | (1200 s) |
| Chromaggus | raid | Holy raid (priest) | 0.911 | 187 | 44.3% | 2.16 | (2298 s) |
| Chromaggus | raid | Restoration raid (shaman) | 0.969 | 193 | 55.6% | 3.33 | (3193 s) |
| Golemagg | tank | Holy tank (priest) | 1.915 | 427 | 49.6% | 4.55 | (628 s) |
| Golemagg | tank | Discipline tank (priest) | 1.826 | 434 | 15.4% | 3.68 | (405 s) |
| Golemagg | tank | Restoration tank (shaman) | 1.441 | 446 | 12.7% | 5.71 | (747 s) |
| Golemagg | raid | Holy raid (priest) | 2.064 | 225 | 53.7% | 2.47 | (1720 s) |
| Golemagg | raid | Restoration raid (shaman) | 2.083 | 242 | 60.3% | 3.14 | (2832 s) |
| Magmadar | tank | Holy tank (priest) | 2.089 | 399 | 47.2% | 4.39 | (511 s) |
| Magmadar | tank | Discipline tank (priest) | 1.941 | 401 | 16.3% | 3.59 | (354 s) |
| Magmadar | tank | Restoration tank (shaman) | 1.536 | 404 | 12.9% | 5.65 | (809 s) |
| Magmadar | raid | Holy raid (priest) | 1.693 | 394 | 43.1% | 3.34 | (1048 s) |
| Magmadar | raid | Restoration raid (shaman) | 1.605 | 408 | 45.6% | 4.23 | (1826 s) |
| Ragnaros | tank | Holy tank (priest) | 0.582 | 493 | 35.1% | 5.80 | (655 s) |
| Ragnaros | tank | Discipline tank (priest) | 0.562 | 355 | 3.1% | 4.65 | (614 s) |
| Ragnaros | tank | Restoration tank (shaman) | 0.469 | 183 | 8.0% | 6.01 | (2495 s) |
| Ragnaros | raid | Holy raid (priest) | 1.255 | 353 | 36.4% | 2.92 | (1251 s) |
| Ragnaros | raid | Restoration raid (shaman) | 1.203 | 364 | 52.1% | 3.82 | (2132 s) |
| Vaelastrasz | tank | Holy tank (priest) | 0.903 | 1040 | 37.7% | 5.87 | (171 s) |
| Vaelastrasz | tank | Discipline tank (priest) | 0.865 | 917 | 6.0% | 4.72 | (134 s) |
| Vaelastrasz | tank | Restoration tank (shaman) | 0.648 | 402 | 8.3% | 5.65 | (2323 s) |
| Vaelastrasz | raid | Holy raid (priest) | 1.139 | 921 | 32.6% | 6.05 | (638 s) |
| Vaelastrasz | raid | Restoration raid (shaman) | 1.199 | 876 | 24.5% | 5.84 | (532 s) |

On the tank, the Holy tank healer carries more than the shaman's on all 7 files (by 0.11 on Ragnaros to 0.65 on
Magmadar), and the Discipline one more than the shaman's and less than Holy on all 7; Discipline OOMs on
Chromaggus (106 s). On the raid, the Holy raid healer is ahead of the shaman's on Onyxia, Broodlord, Magmadar
and Ragnaros and behind on Chromaggus, Golemagg (0.02) and Vaelastrasz: Chain Heal's jumps beat a Flash Heal at
the most hurt where the whole raid takes damage at once.

## Calls for the user

Each has a default the sim uses now. Say the word and it changes.

1. **Divine Aegis: add to the shield or replace it.** A new crit tops the shield up to the limit (20 a level),
   as the tooltip's "up to" reads. Replacing it would make it a little weaker on a tank hit by bursts.
2. **Litany of Light: the fight's first heal counts, and "another heal" is by name.** Two ranks of Greater Heal
   are the same heal and give nothing back; Greater Heal after Flash Heal gives 5/10% back.
3. **Prayer of Healing is weak under Forever** (its ranks are the client's: 70753 kept them) and the raid
   rotation casts it after Flash Heal. Default: as the client.
4. **The gear is a mix.** The Forever priest guides list no items yet, so the synthetic set averages Wowhead's
   Classic healing priest pre-raid list with the Forever Restoration shaman list (gear_targets.md). Default:
   that mix, until a Forever priest list exists.
5. **Shadowfiend and Lightwell aren't in Forever's priest** (client 70753 has neither talent nor rank), so the
   sim has neither. Spirit of Redemption isn't simulated (a dead priest heals).
6. **A shield's rest at the fight's end is booked as nothing**, neither healing nor overheal, since the fight
   cut it off rather than the shield going to waste. Default: unbooked.
7. **Prayer of Mending jumps on a hit that does damage after shields.** A hit a Power Word: Shield absorbs
   whole doesn't trigger it, as in game for an absorbed hit ("Absorbed" isn't damage taken).
8. **Renewed Hope without Binding Heal.** The notes removed Binding Heal; Renewed Hope keeps its client spells
   (Flash Heal, Lesser Heal, Heal, Greater Heal, Penance on a Weakened Soul target).
9. **Holy Nova heals the party but deals no damage**: its damage on enemies within 10 yards is left out, as a
   healer stands back, and everyone in the party counts as in reach. None of the three rotations casts it.
10. **Spirit is 10% of the set's stat budget (standards 5), not the Classic priest list's 13%.** A priest
    heals off Spirit more than a shaman (Meditation, Spiritual Guidance), so the cap costs it some regen and
    healing. Default: 10%, as every set.
11. **Discipline is weaker than Holy at tank healing** on all 7 boss files (The presets, above): Penance and the
    shield don't make up for Holy's Prayer of Mending, Spiritual Healing and Holy Binding. The default build is
    Holy, Discipline is a preset.
12. **Notes-only values.** Two of the healing rules come from Blizzard's notes, not the client (standards 37):
    a Power Word: Shield can be replaced by an equal or higher rank (`ExtraCastCondition`, test
    `TestHealingPriestShieldReplaceRank`), and Binding Heal is removed (client 70753 still has its rows). Both
    are listed under "Notes-only values" in `orchestration/round_70753/priest.md`.
13. **The priest's tank healers carry a lot more than the shaman's** (0.11 to 0.65 more on all 7 files), much
    more than its raid healer does. The priest's numbers are its client values; the gap may say as much about
    the shaman's tank preset (Healing Wave at 12% overheal, a lot of mana left) as about the priest. Default: as
    benched, for the shaman session to look at.
14. **Spec stays unlaunched** (the brief): `ui/healing_priest` builds, but isn't in the spec list.
