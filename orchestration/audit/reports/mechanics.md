# Mechanics pass 80 (2026-11-13, master 3a86d4a6b)

One new finding (low): mechanics-20261113-2, Weakening Disease (Frightskull Shaft) lowers the target's Strength by 50
in the sim, where the Forever client lowers its attack power by 92 (no player DPS effect). Open findings
mechanics-20261113-1, 20261112-1 (Holy Strike and the totem) and 20261104-1 stay open; nothing on master touched them.

## Get current

- Client: newest 1.60 build is still 1.60.1.70938 (`wago.py latest --probe 300`). No new spell data.
- Blizzard: dev notes t/2360696 newest post 10 (2026-11-04), hotfix thread t/2403040 newest post 9 (2026-11-12
  19:10 UTC). Nothing newer than the last pass.
- Daily update / log audit: eac5f72d1 (log audit 2026-11-13) is the source of the commits below.

## Commits cad23ae1a..3a86d4a6b (sim code)

| Commit | What | Verdict |
|---|---|---|
| 041bb0a90 | Weakening Disease 18633 under Forever: 15 every 2 sec, 7 ticks (14 sec) | Matches 70938 (effect #0 aura 3 bp 15 per 2000, duration 14000). Effect #1 is aura 99 bp -92 (attack power), sim still applies -50 Strength: new finding -2 (low) |
| d22fe3e23 | Frightskull Shaft 3.9 PPM under Forever (Classic 0.5 kept) | Client has ProcChance 101 and no PPM row, so the log is the best source: 3.9 PPM at 3.4 speed = 22.1%, log 21.8% (190/873) white, 22.8% Stormstrike. OK |
| c7a99fa49 | Windfury Weapon's main hand attacks (8232/8235/10486/16362) roll Windfury Totem under Forever | Matches t/2403040 post 6 ("granting their attacks a chance to activate both") and the log (47/162). Still excluded as extra attacks? No: the imbue attacks are not `IsExtraAttack`, OH attacks are OHSpecial and don't roll. OK |
| 4c4790e3e | Flametongue Weapon rank 6 cut 5.5 below client points | Class-side, log-backed (about 15 windows, levels 55-60, two weapon speeds). Same method as rank 5. OK |
| 1743980e6 | Shadow Oil / Frost Oil procs keyed as 350026 / 350025 under Forever | 70938 has both (Shadow Bolt bp 52 rpl 0.9, Frostbolt bp 33 rpl 0.7); ID only, damage ranges unchanged. OK |

## Knock-on

- `windfuryTotemRollsOnForever` is only used by the Windfury Totem rollsOn; Holy Strike (open -1) is still not in
  `windfuryTotemAbilityIds`.
- Frightskull Shaft goes through `CreateWeaponProcSpellByRuleset`, so Classic keeps 0.5 PPM and 8 x 15 ticks.

## Verdicts

`ref/mechanics/verdicts.md`: as-of line moved to 3a86d4a6b; Windfury imbue + totem row notes c7a99fa49; class-side line
adds Flametongue rank 6, Frightskull Shaft rate and disease, and the oil spell IDs.
