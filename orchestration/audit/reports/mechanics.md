# Mechanics pass 82 (2026-11-14, master 95289e879)

One new finding (mechanics-20261114-1, medium) and new tier 1 evidence on mechanics-20261112-1, both from the
November 13 hotfix notes.

## Get current

- Client: 1.60.1.70938 still the newest (the gate reported no newer build).
- Blizzard: dev notes t/2360696 unchanged (post 10, 2026-11-04). Hotfix thread t/2403040 has post 10
  (2026-11-14 01:54 UTC) and post 1 edited 01:54 with the "November 13, 2026" section. Mechanics lines:
  - Windfury Totem activated by all main hand melee abilities, not off hand strikes -> update on mechanics-20261112-1.
  - Judgement of Command counts as a melee attack and activates on-hit effects -> mechanics-20261114-1.
  - Gloves of Holy Might only activated by harmful damaging spells and abilities (item not in the sim; the daily
    update's needs-your-call list covers it).
  - Creature stun DR, Wild Growth coefficient 6.6%, Frost Nova breaking, Divine Spirit cost, Siphon Life silence,
    Berserker Rage vs Sap: no DPS-sim effect (Wild Growth is the healer auditors').
  - Item stat fixes (Darkspear signets, Bad Hat, Channeler's Ring, Blessed Band of Light, Taran Icebreaker 2H):
    item data, for the gear / update-sources pipeline.
- Daily update: no TASKS.md or update_ledger.md commit since 1bf5feb4b, so the Nov 13 hotfixes are not in the repo yet.

## Commits 1bf5feb4b..95289e879 (sim code)

| Commit | What | Verdict |
|---|---|---|
| 87c030418 | Bolt's Charged Eyeball, Everlook Pathcarver, Chained Essence of Eranikus under Forever (`sim/common/item_effects.go`) | All values match client 70938: proc 15% / 3 s ICD / ProcTypeMask 332116 / 60 Nature; 140 every 2 s for 20 s, 600 s CD, 20 s category 1141; 607 Physical, 300 s CD, 15 s on 1141, DefenseType 1 on 1295270 and 1315996. Eyeball's 50/50 self/target split is a stated assumption (notes.md), not a finding. Verdict row added. |

Other commits (BiS lists, phase 0 content, db row for 15141, rotopt ranking) are not mechanics.

## Open

- Server hotfix SpellMisc rows after push 113605 still unread (carried from pass 81).
- mechanics-20261114-1 and the -20261112-1 update are unmeasured estimates; a worker should measure with paired seeds.
