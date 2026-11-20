# Mechanics pass 94 (2026-11-20, master 4be96efef)

Range e07e8fdb8..4be96efef (two sim commits, item-use tooling and a trinket priority). Client unchanged at
1.60.1.70959. One new finding, from Blizzard's November 19 hotfix post: mechanics-20261120-1 (high, needs a user
ruling). Open findings mechanics-20261118-2, 20261118-1, 20261113-2 and 20261104-1 stay open (not touched).

## Get current

| Source | Newest | New since pass 93 |
|---|---|---|
| Client build | 1.60.1.70959 (`wago.py latest --probe 300`) | none |
| Dev notes t/2360696 | post 10, 2026-11-04 19:35Z | none, no edits |
| Hotfix post t/2403040 | post 13, 2026-11-20 04:00Z; post 1 edited 04:01Z | November 19 hotfixes (below) |
| Blue posters (Aggrend, Linxy user actions) | Linxy, anniversary post 2026-11-17 | nothing mechanical |
| Daily update / TASKS | no TASKS.md, update_ledger.md or forever-update merges in range | none (TASKS.md:480, from 11-19, already asks for the 2H Spec ruling) |

November 19 hotfixes, by sim relevance:

| Line | Sim | Verdict |
|---|---|---|
| Retribution: Two-Handed Weapon Specialization "now correctly increases all Holy Damage you deal while a two-handed weapon is equipped" | Holy part only on seals, Holy Strike, Judgement of Command, Consecration (`pal_2h_spec_holy`) | **finding mechanics-20261120-1** |
| Shaper's Terrace, Vitrix's Poison Shock less damage | not simmed | none |
| Camping, quests, Skyborne riding discount, Tauren Cultivation, Motes soulbound | no combat effect | none |

## Commits

| Commit | Change | Verdict |
|---|---|---|
| bb5c326b6 | `Player.item_uses_off` / `item_effects_off` (core `character.go`, `major_cooldown.go`): skip an item's effects, or register its use as never castable; `ItemHasUse` for rotopt | Tooling only. Both maps are nil unless the request sets them, so every report preset sims as before. No client value claimed. |
| 2c6005e2b | Adaptive Combat Assistant (272437) cooldown priority Low, so it goes after Chained Essence / Blackhand's Breadth on the shared 15 s | Ordering only, no client value; worn only in the Enhancement Phase 0 BiS list, no report preset. |

## Size of the new finding (estimate)

Damage shares from the shared run (`$AUDIT_WORK/runs/master`, `brk.py`) of the Holy spells the sim keeps out of
the 6%: `paladin__level_60` Hammer of Wrath 5.1% + ranged Judgement of Righteousness 3.4% (about +0.5%);
`paladin__level_60_sub_holy` 5.0% + 3.8% (about +0.5%); `paladin__level_60_shock_ret` Holy Shock 11.8% + Hammer
of Wrath 4.8% (about +1.0%). Not measured: the fix needs a code change.
