# Mechanics pass 106 (2026-11-25, master 85691655a)

Range: c3bbe33ea..85691655a. Client 1.60.1.71091 still newest (gate). Blizzard: dev notes
t/2360696 unchanged (8 posts, newest 2026-11-04); hotfix thread t/2403040 unchanged (#15,
2026-11-24); blue tracker has a new Forever post, "Balance Druid Adjustments - November 25"
(t/2429007, Kaivax, 2026-11-25 18:32Z, "now live"). Daily update: one TASKS.md item
(85691655a, Blackhand's Breadth hotfix from foreverchanges, Nov 24 push).

## Commits checked

| Commit | Change | Check | Verdict |
|---|---|---|---|
| 2f42f4a44 | Tank batch: rage-gain threat, damage-taken Rage armor rows, flat reduction cap, bear threat | log-sourced (EV-0032..0050, 0154); level 63 row unchanged | settled (ruling) |
| efeb06b95 | Resist constant 8L+19, 75% partial cap, binary spells, Flurry vs Windfury Totem (shaman) | log-sourced (EV-0091, 0092, 0142, 0010); shaman totem swings no longer spend Flurry | settled; resolves mechanics-20261118-1 |
| 60fde111f | Hunter hawk, pet white crit, Explosive Trap, pet autocast | log-sourced (EV-0073, 0127, 0134, 0068, 0187); class rules | settled (ruling) |
| 295206ddf | Item PPMs, set procs, 49 on-hit weapons | Flame Wrath 16559 bp 154 matches; Ironfoe client 10% / 5 s vs sim 3% from logs (server side) | settled (logs) |
| b663145ad | Weapon-proc chains, JoC double roll, Revelation, judgements refresh, damage shields | log-sourced (EV-0145, 0155, 0028, 0117, 0066 ...) | settled (ruling) |
| 22d2cf952 | Explosives keep swing timer, stun explosives miss level 63, demon crit, extra attacks roll Windfury Totem, Shield Slam MH special | log-sourced, user ruling 2026-11-25 | settled (ruling) |

## Findings

| Id | Severity | Title |
|---|---|---|
| mechanics-20261125-1 | high (estimate) | Nov 25 Balance druid changes live; sim has none (Moonkin +10%, Eclipse Wrath crit, Moonfury removed, Improved Starfire 2 points, Starfire cast time) |
| mechanics-20261125-2 | low (estimate) | Blackhand's Breadth mark still Shadow / can miss; Nov 24 hotfix made it physical, unresistable |

Resolved: mechanics-20261118-1 (efeb06b95, shaman half).

## Open items

- Hotfix rows for Blackhand's Breadth and the Balance changes aren't cached (`hotfixes.py rows`
  returned nothing); a fetch from wago.tools/hotfixes would give exact values.
- No sims run: all commits are log-sourced rulings; DPS sizes above are estimates.
