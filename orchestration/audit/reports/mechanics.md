# Mechanics pass 47 (2026-10-29, master dcd927490)

Range audited: 52ec62e99..dcd927490. Two code commits: b2b384c35 (Windfury Totem metrics and labels; `sim/core/buffs.go`, `sim/core/totem_weapon_buffs.go`, `sim/shaman/air_totems.go`) and c282ae5d4 (`sim/warrior/execute.go`). No TASKS.md or update-ledger change, no `forever-update/*` merge.

## Get current

- Client: 1.60.1.70685 is new (was 70658). `wago.py diff 70658 70685`: only Adventurer's Tent (spell 1325145, item 238896, new SpellCategory 2607) and a new spell Camping (1328862). SpellItemEnchantment, TraitDefinitionEffectPoints and CurvePoint identical. No combat row moved, so every verdict holds; verdicts.md client line updated.
- Dev notes thread 2360696: still 8 posts; post 7 (edited 2026-10-27 16:47:14) and post 8 (edited 16:46:55) unchanged.
- Blue posts after 2026-10-28 (forum search, Kaivax, Aggrend, Gnedi, Linxy, Brenoch): nothing newer than Kaivax's 01:28 name-list post, already seen in pass 46.
- Server hotfixes: `tools/audit/mechanics/hotfixes.py` is still missing from the tree; not re-read.

## Commits

| Commit | Change | Mechanic effect | Verdict |
|---|---|---|---|
| b2b384c35 | Raid Windfury Totem aura gets ActionID 8512 / 10613 / 10614 (Era totem spell IDs, rank 3 = 10614); proc AP buff relabelled "Windfury Attack" | None: aura label "Windfury Totem Raid" kept, so its random stream is unchanged; the 1 sec duration lookup follows the new buff label; procs, AP and blocking untouched | Reporting only. Not a finding. |
| c282ae5d4 | Execute drops `SpellFlagPassiveSpell` | Flag is read by metrics and by Revelation's direct-cast filter (`sim/common/enchant_effects.go:292`), which also requires ProcMaskSpellDamage; Execute is melee, so no proc changes | Reporting only. Not a finding. |

No DPS sim run: neither commit changes a number.

## Findings

- No new findings. mechanics-20261027-2 (Amplify Curse vs Bane of Agony) stays open (parked for an in-game test); no warlock commit in range.
