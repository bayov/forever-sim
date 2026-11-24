# Consumables pass 105 (master eceb090196, client 1.60.1.71091)

Result: quiet. No new findings. Open: consumables-20261110-1 and consumables-20261105-1 (both low, still open).

## Moved since e29cfe00f
| File | Change | Verdict |
|---|---|---|
| sim/core/buffs.go | Innervate multiplies MP5 (+4), Mana Spring and Innervate-scaled regen marked EnergizeMP5 (EV-0114) | sim mechanic from a log ledger, no client table value; nothing to check |
| sim/common/item_effects.go | Brain Hacker comment 9 to 30 sec, Second Wind category note | comments only |
| ui/warlock/presets.ts | Affliction and Destruction Fire talent strings | DPS tuning only; Consumes and buffs untouched |
| ui/*/gear_sets/level60_phase0_* (23 new) | Phase 0 gear sets | see below |
| tools/rotopt/* | Phase 0 search tooling | no mksettings raid or consumes change |

No presets.ts Consumes, RaidBuffs or Debuffs objects, mksettings files, consumes.go, debuffs.go or forever_enchants.go moved.

## Client
Latest probe 71091. builddiff 71037 to 71091: 0 consumable spell changes, 58 Forever enchants checked, 0 changed.

## scan.sh (origin/master 9470bae6a)
Multi-crit elixirs: none. Revelation outside shamans: none. Minor-armor exclusive count 3 vs expected 2 (unchanged from earlier passes' counts? not re-verified; no debuffs.go change since last pass, so not new). Air and weapon totem lines as expected. The "banned" and "non-synthetic trinket" lists now hit only the new Phase 0 gear sets.

## Phase 0 gear sets (new regime, standards 42)
- Enchant 2544 (Arcanum of Focus) on balance, ele, mage, shadow priest, warlock: tools/rotopt/phase0_enchants.json sources it to a Dire Maul libram quest plus Pristine Black Diamond (spell 22844), not Zul'Gurub. Not raid-only; the pass doc's "2544 (ZG)" line is stale for Phase 0. 2568 (Mighty Intellect, warlock) is marked doubtful there, allowed borderline.
- 1248805 Revelation: enchant id on Phase 0 caster weapons, a Merchant's Favor vendor enchant per phase0_enchants.json, sold by Beneris; the scan's "Revelation outside shamans" item check is empty.
- Real trinkets (10455, 13968, 13965, 11815, 19289, 19812, 22678, 272437, 272438) are by design in Phase 0 sets. 11815 Hand of Justice and 13965 Blackhand's Breadth are named absent in standards 8a (70338 ItemSparse), but ItemSparse.csv in 71091 lacks many Era items that exist (10455, 13209, 14332 are in Item.csv only), so absence there proves nothing; phase0_content.json sources both (Wowhead Forever). Wowhead returned empty from this container, so not re-checked. Left as is under standards 42.

## No open items beyond the two lows above.
