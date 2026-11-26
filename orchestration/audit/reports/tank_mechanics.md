# Tank mechanics pass 35 (2026-11-26, master bfbee78fc)

Quiet: no new finding. Due for client 1.60.1.71091 -> 1.60.1.71139.

## What changed

- Code, 85691655a..bfbee78fc (58 commits): tools/dps_report, ui/dps_report, DPS spec APLs/presets, tools/rotopt,
  orchestration docs. No file under `sim/`, no tank UI (`ui/tank_warrior`, `ui/protection_paladin`,
  `ui/feral_tank_druid`, `ui/warden_shaman`), no `tools/tankbench`, no `docs/tank_*`. Update ledger unchanged.
- Blizzard: hotfix thread t/2403040 has post 16 (Nov 25, 2026-11-25T23:25Z). Dev notes t/2360696 and class tuning
  t/2404092 have nothing new.

## Client 71091 -> 71139 (`wago.py diff --set mechanics`, tier 1)

| Change | Tank reach | Verdict |
|---|---|---|
| Sayge's Dark Fortune 692746-700260 and 1331584-98: EffectAura 80/101/79/22 -> 4 or 137, misc values moved | none: the sim applies 23735-23769 and world buffs don't work in raids (ForeverRaidInstance) | no tank effect; shared buffs are the mechanics pass's |
| World buffs + Battle Shout 334022: Attributes_6 0x800 | none (flag only; Nov 23 "world buffs fall off when changing areas") | ok |
| Fengus' Ferocity 1331556-9: 230 -> 115 | none (Nov 23 "twice as much Spell Power") | not a tank buff |
| Rogue poisons / Vanish: Attributes_5 0x2 | none (poisons BoP hotfix) | - |
| Reinforce 1214709 and curve 97574 sign flip | durability only | - |
| Weight of Duty 1335948 7 -> 8, Explosive Runeshard, Holy Fire duration | no tank preset | - |
| Clam foods recategorized, Nutritious Food item effects | none | - |
| SpellLevels, SpellPower, SpellCooldowns, SpellShapeshift(Form), Trait tables, ChrRaces | no change | - |

## Nov 25 hotfixes against the tanks (tier 1 notes)

| Note | Sim | Verdict |
|---|---|---|
| Paladin: Judgements failed to refresh from attacks below 20 s | `refreshesJudgement` (sim/core/debuffs.go:784) refreshes JotC/JoW on any landed white swing at any duration | matches the fixed game |
| Paladin: Exorcism couldn't crit | no protection_paladin APL casts Exorcism | no tank effect |
| Libram of Law on Judgement of Command | no tank preset | - |
| Warrior: Weaponmaster in BGs | Arms, PvP | - |

## Open findings

Unchanged (no code change could fix them): 20261124-4 (Stoneform autocast in scoring runs), 20261124-3 (racial
active tests), 20261117-1 (UW skips Sunder), 20261108-1 (Warden shock threat test), 20261107-1 (stale "client
disagrees" comments).

## Not done

No tankbench run: nothing in the sim or the tank presets changed, so the pass 34 numbers stand.
