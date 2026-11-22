# Tank mechanics, pass 27

- Master: 7a6d1beb9d (from 8b5137c6d). Client: 1.60.1.71037 (unchanged since pass 25; no ledger entries since pass 26).
- Result: quiet. Open findings unchanged (20261117-1, 20261108-1, 20261107-1).

## What changed on master (8 watched files)

| file / change | source | tank effect |
|---|---|---|
| core debuffs.go: Judgement of Wisdom pays only on landed swings under Forever (bd7029410) | wowsims/forever 9f254e8d87, EV-0018 (Forever log, tier 1 logs) | small: fewer mana returns, so less mana-gain threat (0.5 a mana, unmultiplied). TestProtection TPS -0.1 to -0.5%, TestWardenShaman moves with it (219829b3a); no avoidance, mitigation or health effect |
| warrior deep_wounds.go: critting hand's weapon, no attacker damage bonus, no bleed from Thunder Clap crits (990cade4e) | wowsims/forever EV-0139 (tier 1 logs) | none: no tank warrior preset takes Deep Wounds (`35--251533120331221351`, `05-05050005-25053...`, `05-05050005-25033...`). `SpellFlagIgnoreAttackerModifiers` skips damage multipliers only, not threat (spell_result.go:796) |
| warrior execute.go: no base-cost refund when avoided; Execute rolls Windfury Totem twice (aa394d3be, b7edf19bb) | wowsims/forever 7025f28b98, EV-0155 (tier 1 logs) | none: tank APLs stay in Defensive Stance, cast no Execute |
| warrior talents.go: Bloodthrill and Flurry ignore Windfury Totem's swing | wowsims/forever EV-0056/0136/0138/0140 (tier 1 logs) | none: tank presets take neither talent; Unbridled Wrath (5/5 on Level 60 Protection) unchanged |
| shaman talents.go: Maelstrom Weapon skips Windfury Totem's swing (840029bd8) | wowsims/forever EV-0009 (tier 1 logs) | none: no Warden preset takes Maelstrom Weapon |
| shaman flame_shock.go: hasted GCD under Forever (56e5f070d) | standards 27, wowsims/forever | none: Warden tank APLs cast no Flame Shock |
| paladin item_sets_pve.go: Soulforge and Forever Soulforge (1673) share Crusader's Wrath (7876cfdd3) | client 71037 sets | none in presets; the tank warrior golden gains one AllItems row (Battlegear of Heroism 1778) |
| core attack.go: `SetExtraAttackSourceForTest` | test helper | none |

Read in full: `applyDeepWounds`/`procDeepWounds`, `registerExecuteSpell`, `applyBloodthrill`, the Flurry consumption trigger, `newJudgementOfWisdomAura`, `applyMaelstromWeapon`'s trigger.

## Tests

`go test -tags=with_db ./sim/warrior/tank_warrior/ ./sim/paladin/protection/ ./sim/shaman/warden/ ./sim/druid/feral/...`: all pass. Without `-tags=with_db` they panic on missing items, as expected.

## Findings

No new finding. The JoW change is a logged Forever rule (tier 1 logs) whose only tank effect is mana-gain threat, itself still open in verdicts (unmultiplied, Classic tier 3).
