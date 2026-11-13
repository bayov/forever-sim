# tank_mechanics pass 13 (2026-11-13, master 71562a8b03, client 1.60.1.70888+hf)

Result: quiet. No new findings; three findings verified resolved.

**Getting current:** wago's latest is still 1.60.1.70888 (unchanged since pass 11). No Blizzard notes newer than the
Nov 11 hotfixes, which pass 12 already read.

**Changed since 8306f1d52 (watched paths):** sim/paladin/{judgement,seal_of_fury,soc,sor}.go (f1c07e13a, Judgement is a
ranged attack), sim/druid/healing.go (Nature's Swiftness comment, Restoration only), sim/core/heal_targets.go and
apl_values_resources.go (healer Most Hurt / Lowest Health read seen health), tools/tankbench/req/*.json (preset re-tunes).

## Protection paladin

| Item | Client 70888 | Sim | Verdict | Tier |
|---|---|---|---|---|
| Judgement of Fury 20414 (and lower ranks) | DefenseType 2, notes say ranged | ranged special, ranged table, NoBlockDodgeParry | right | 1 (notes; client row not updated) |
| Judgement of Righteousness 20286 / Command 20966 | DefenseType 2 | same, via rangedJudgement | right | 1 (notes) |
| Judgement threat | - | unchanged, 1.9 x 1.33 a point (TestJudgementOfFuryRanged) | right | as before |
| Precision on the ranged roll | 20189 unscoped A_MOD_HIT_CHANCE | MeleeHit stat feeds PhysicalHitChance for ranged too | right | 1 |
| Weapon skill on ranged table | - | base for level; presets are Undead, no racial lost | right | 3 |
| Melee-keyed procs (JoW/JoL, weapon procs) | hotfix: no longer | ProcMaskRangedSpecial | right | 1 (notes) |
| Windfury Totem | - | judgements not in windfuryTotemAbilityIds | right | logs |

## Feral tank (bear)

| Item | Client | Sim | Verdict | Tier |
|---|---|---|---|---|
| Tauren Endurance 20550 | aura 133, +5% | MultiplyStat(Health, 1.05) | right | 1 |
| Feral Instinct 3 (Max survival) | 10% Swipe a point | swipe.go:53, tested | right | 1 |
| Brutal Impact 1 | Bash only | bear.go:420 | right (no tank effect) | 1 |
| Maul / Swipe / FR APL thresholds | - | rotation choice | tank_rotations' scope | - |

## Warden, warrior

Consumable swaps only (Wizard Oil, Supreme Power, Titans, Greater Mageblood, Healing Stream r5 10463), all verified in
pass 12's verdicts. Windfury rule: Heroic Strike, Sunder, Cleave, Maul roll; Shield Slam doesn't (list checked).

## Engine

- Healer seen-health values: no tankbench req reads health (grep for currentHealth / lowestHealth / mostHurt: none),
  so the tank score's replay is unaffected.

## Tests run (-tags=with_db)

sim/paladin/protection (TestJudgementOfFuryRanged, TestSealOfFury*), sim/paladin/retribution (TestJudgementsAreRanged),
sim/core Windfury, sim/warrior/tank_warrior Windfury: all pass. (Without with_db the retribution test panics on a missing
item; build tag only, not a bug.)

## Findings

- Resolved: 20261112-1 (judgement ranged), 20261112-2 (Windfury rule), 20261111-1 (shaman doc).
- Still open: 20261108-1 (Warden shock threat under Stormstrike untested), 20261107-1 (stale client labels).
