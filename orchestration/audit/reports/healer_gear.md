# healer_gear report, pass 6, master 55cd92b09

Trigger: forced re-audit after the Restoration druid fixes (git diff --stat cb0d928fd HEAD); run on the local branch tankheal.

## Gear (rules 1b, 5, 5a, 35)
Between 33db32b36 (last audited) and 55cd92b09 nothing changed in tools/database forever_synthetic_items_*, ui/*/gear_sets,
forever_wowhead_recipes.json, gear_targets.md or sim/*/items.go (git diff --stat empty). The only healer changes are the Restoration
druid talents (ui/restoration_druid/presets.ts, 05302001--5253005153113051, Heart of the Wild 2) and APLs; neither touches a set.
Per-piece recompute skipped; the pass 3 result stands (no findings open).

## Consumables, judged at each preset's objective (damage carried, healbench -role tank -carry)
All five presets share one set: Supreme Power, Lobster Stew, Greater Arcane Elixir, Greater Intellect, Mageblood, Sages,
Brilliant Mana Oil, Major Mana Potion, Demonic Rune, Cerebral Cortex. Method: tools healer_gear/sweep.py (in AUDIT_WORK), 2000 iterations,
paired seeds; carry noise about 0.01.

Greater Arcane Elixir and Lobster Stew add SpellDamage, which healing ignores: removing the elixir gives identical carry. Swap = Greater
Cleric's Elixir + Sage's Tea in their place (carry; base to swap):

| boss | druid | shaman | paladin | Holy | Disc |
|---|---|---|---|---|---|
| onyxia | 1.370 to 1.373 | 1.345 to 1.366 | 1.251 to 1.251 | 1.385 to 1.378 | 1.363 to 1.363 |
| chromaggus | 1.051 to 1.047 | 0.965 to 0.977 | 0.915 to 0.929 | 0.953 to 0.970 | 0.873 to 0.893 |
| golemagg | 1.953 to 1.988 | 1.916 to 1.950 | 1.911 to 1.922 | 1.985 to 2.006 | 1.993 to 1.995 |

Further, on top of the swap (carry): Brilliant Wizard Oil (swapBW) and Distilled Wisdom (swapDW):

| boss | druid BW/DW | shaman BW/DW | paladin BW/DW | Holy BW/DW | Disc BW/DW |
|---|---|---|---|---|---|
| onyxia | 1.380 / 1.359 | 1.377 / 1.347 | 1.252 / 1.230 | 1.391 / 1.382 | 1.359 / 1.342 |
| chromaggus | 1.057 / 1.049 | 0.974 / 0.987 | 0.930 / 0.914 | 0.970 / 1.035 | 0.884 / 0.971 |
| golemagg | 1.996 / 1.981 | 1.950 / 1.881 | 1.945 / 1.892 | 2.024 / 1.977 | 2.035 / 1.964 |

Other single swaps at the base set, chromaggus (carry): Wicked Regeneration and Whale elixirs: no gain (druid 1.062 and 1.042 noise, others lower or equal).
The raid role was run only once (chromaggus, 1000 iterations) and its output lost to a pipeline error; not used.

Findings: -1 (Cleric's Elixir, Sage's Tea), -2 (wizard oil), -3 (Distilled Wisdom for the priests; needs a ruling).
Druid: the new talents/APL leave its consumables the same as the others; no druid-specific consumable gain over 1% was measured.
