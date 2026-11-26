# Mechanics pass 108 (2026-11-26)

- Master: ab3fff700 (last audited bfbee78fc). Client 1.60.1.71139 (unchanged).
- Result: findings (1 low).

## Repo since bfbee78fc

Four sim commits, all warrior leveling work: 143bf4f72 (ranks by learn level, leveling hooks), 8f5b59325 (Charge, Intercept, Victory Rush, Retaliation, Concussion Blow, Gore Drinker, Cannibalize, Rage decay, stances by level), 1223e764e (big pulls: facing, camps, area reach), 81a2e51fc (Recovery, shield spikes, leveling heals).

| Change | Level 60 effect | Verdict |
|---|---|---|
| aoe.go AoETargetCount / TargetsNear, consumes.go explosives, Dragon Breath Chili, Dragonfire, Caustic Smog | Outside leveling returns ActiveTargetUnits as before | Gated, no change |
| apl_values_encounter.go execute phase -> InExecuteRange | Outside leveling = sim.executePhase <= N, as before | No change |
| environment.go NextTargetUnit, spell_result.go, outcomeEnemyMeleeWhite from behind | Gated on leveling | No change |
| spell_outcome.go: stunned target can't block / dodge / parry | Ungated; Era rule (tier 3) | Right rule; Might of Shahram stuns StunImmune bosses: mechanics-20261126-4 |
| debuffs.go SunderArmorAuraOfRank (90 x rank) | Rank 5 = 450, same as SunderArmorAura | No change |
| Warrior ranks (knownRank) for BT, MS, Execute, HS, Cleave, Rend, Revenge, Shield Slam, Slam, Sunder, TC, Overpower, Pummel, Hamstring, Battle Shout, Demo Shout | Rank max at 60; values equal the old constants | No change |
| Retaliation registered at 60 outside leveling; Victory Rush, Intercept, Concussion Blow, Gore Drinker, Cannibalize, Recovery heal | Leveling-gated or unused by presets | No change |
| MaxRange on melee specials, Thunder Clap 8 yd | Only checked in a leveling session (spell.go:589) | No change |

## Measured (paired seeds, 10k iterations, run_reqs.py, bfbee78fc vs ab3fff700)

| Preset | Old | New | Delta |
|---|---|---|---|
| paladin__level_60 | 930.27 | 931.52 | +1.25 |
| paladin__level_60-duty-crusader | 920.27 | 921.22 | +0.95 |

Only Blackblade of Shahram's stun reaches ret in this range.

## Blizzard and daily update

No forum post past 2026-11-25T23:25:08Z (Hotfixes Nov 25, 2403040/16); dev notes thread 2360696 unchanged (8 posts, last 2026-11-04). No TASKS.md or update_ledger change, no forever-update merge. Client build 71139 still newest per the gate.
