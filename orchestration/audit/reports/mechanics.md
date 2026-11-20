# Mechanics pass 96 (2026-11-20, master e6eaa9470)

Range 813f2b1e9..e6eaa9470: no sim or mksettings commits; one TASKS.md commit (e6eaa9470: Shank proc check for
PR #135, Blizzard's Nov 19 wording for Two-Handed Weapon Specialization, already in mechanics-20261120-1, unplaced
Nov 20 pushes, a failing Onyxia Blood Talisman test). The pass was due to the new client build 1.60.1.71037.
One new finding (low): mechanics-20261120-2.

## Get current

| Source | Seen | Newer? |
|---|---|---|
| Client | 1.60.1.71037 (was 70959) | yes, diffed below |
| Hotfix thread t/2403040 | 11 posts, last 2026-11-20 04:00:55Z (post 1 edited 04:01:18Z) | no |
| Dev notes t/2360696 | last 2026-11-04 19:35Z (post 1 edited 2026-11-03 03:20Z) | no |
| Blue tracker | 2026-11-20 18:00Z weekly news (t/2423714): links the class-design podcast, no mechanics text | nothing mechanical |
| Daily update | e6eaa9470 (TASKS.md only) | nothing new for shared mechanics |

## Client 1.60.1.71037 against 70959

`wago.py diff --by-spell`: 119 spells. Shared and class rows that matter:

| Spells | Change | Sim |
|---|---|---|
| Flurry 12966, Unbridled Wrath 12322, Reckless Swings 1329555 | ProcTypeMask 4 -> 20 (Nov 16 hotfix baked) | UW and RS landed; Flurry tracked as client_update-20261117-2 |
| Shaman Flurry 16257 | unchanged (ProcTypeMask 4, 500 ms) | mechanics-20261118-1 stays open |
| Windfury Totem 8516 / 10608 / 10610 | category 2619, 100 ms category recovery | matches buffs.go:2211 (100 ms) |
| Warchief's Blessing, Might of Stormwind, Rallying Cry (6 ids) | Attributes_3 0x100000 -> 0x100 (players only) | world buffs off in presets; matches Nov 17 note |
| Execute 5308, 20658-20662 | proc-suppress bits on ranks 1-4, 0x40000 on all (dummy cast) | no effect on the sim's Execute |
| Blue Dragon 23688, Heroism 23682 | 3%, 198 (push 114044 baked) | matches |
| Hunter T1 5P 1301252 | mask 135168 (Aimed + Multi-Shot) | matches |
| Rogue T1 5P 1301708 | mask now Eviscerate + Death Mark | **sim Eviscerate only: mechanics-20261120-2** |
| Vampiric Embrace 15286 | 20 -> 10 | healer_mechanics-20261117-16 |
| Genesis 1223081 | mask_3 +128 | healing only, not checked further |

Other rows are NPC, housing, profession, mount or quest spells. SpellItemEnchantment: +8773 Fishing Lure only.
TraitDefinitionEffectPoints, TraitNodeEntry, ChrRaces, CharBaseInfo identical. CurvePoint: 504 added rows (new
curves 128521-128772), none changed. ItemSparse 108 / ItemEffect 11 rows changed: left to the client update pass.

## Measured

Report requests on CLI e6eaa94706e6 (1000 iterations, 180 s): only rogue__subtlety casts Death Mark (17.4% of
damage, 9.5 casts a fight); combat-swords and mutilate cast none. The finding's DPS size is an estimate.

## Updated

- `findings/mechanics.md`: new mechanics-20261120-2; pass 96 note under mechanics-20261118-1.
- `ref/mechanics/verdicts.md`: 71037 row, "as of" moved to master e6eaa9470, client 71037.
