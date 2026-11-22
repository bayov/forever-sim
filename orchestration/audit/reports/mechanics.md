# Mechanics pass 98 (2026-11-22, master 7a6d1beb9)

Range 8b5137c6d..7a6d1beb9: 14 sim commits, 18 sim files. No new client build (1.60.1.71037 still newest).
No new Blizzard post. No new finding; mechanics-20261118-1 is half fixed (warrior), still open for shaman.

## Get current

| Source | Seen | Newer? |
|---|---|---|
| Client | 1.60.1.71037 (gate probe) | no |
| Hotfix thread t/2403040 | 14 posts, last 2026-11-21 00:18:29Z, post 1 edited 00:19:18Z | no |
| Dev notes t/2360696 | last 2026-11-04 19:35Z | no |
| Blue tracker | newest 2026-11-21 00:18Z (t/2403040/14) | no |
| Daily update | 287787c2d, 7876cfdd3, 7a6d1beb9 (TASKS): server pushes 114138-114326 on 71037 (region 93 rows) named | see below |

Hotfix pushes the daily update read (114204 Stormshroud, 114283 JoW/JoL/JoJ level-1 copies, 114313 Execute,
114314 Nature's Bounty energy, 114318 Spiritcaller): Execute is now in the sim (b7edf19bb); Feralheart's 4pc energy
part and the Spiritcaller re-sim are already TASKS.md items; the rest change no sim value.

## Commits triaged

| Commit | Change | Check | Verdict |
|---|---|---|---|
| b7edf19bb | WF Totem swing spends no warrior Flurry, procs no Bloodthrill; Execute rolls the totem twice (core RollWindfuryTotemOnCast) | 71037 Execute rows Attributes_3 0x70400 (Always Hit); push 114313 adds Suppress Weapon Procs; double roll from logs (EV-0155), user ruling 2026-11-22, knob warr_exec_wf2. Roll split out of createExtraAttackAura keeps the 20% / 100 ms cooldown | ok; resolves the warrior half of mechanics-20261118-1 |
| aa394d3be | Avoided Execute refunds nothing under Forever | consistent with an Always Hit cast row; log-sourced | ok |
| 990cade4e | Deep Wounds: critting hand's weapon, no attacker modifiers, no bleed from magic-table crits | log-sourced (EV-0139) | ok |
| bd7029410 | Judgement of Wisdom pays only on landed swings (Forever) | log-sourced (EV-0018); Classic path unchanged | ok |
| 840029bd8 | Maelstrom Weapon: no stack from the WF Totem swing | logs after push 114060 (EV-0009) | ok |
| 56e5f070d | Flame Shock GCD hasted under Forever | magic spell, same rule as other spells | ok |
| 4192d6e49 | Ice Lance 0.11 SP under Forever | user ruling 2026-11-22 (logs, EV-0171) | ok, settled |
| 051a21cd3 | Hunter pet keeps summon-time AP | log-sourced (EV-0072); class auditor's area | ok |
| 3960dd1f6 | Wrack: half points on the dot's own grid (core Dot.SetTickPeriod) | only Wrack calls SetTickPeriod (git grep) | ok |
| 303023018, 07c1ac42b | Linken's Boomerang, Cloak of Fire uses | 15712 bp 150, 180 s, cat 1141 10 s; 18364 aura 23 5 s x 15 s -> 18374 bp 25, 120 s | ok, tier 1 |
| 7876cfdd3 | 8 Dungeon Set 1 reveal sets | ItemSetSpell 71037 rows listed in verdicts.md | ok, tier 1 |
| b21a1d281, 5657e2538 | notes only | - | - |

## Open
- mechanics-20261118-1: shaman Flurry still spends a stack on the WF Totem swing (sim/shaman/talents.go:491).
- Other open findings unchanged (20261120-2, 20261118-2, 20261113-2, 20261104-1).
