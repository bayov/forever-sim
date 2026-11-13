# Consumables pass 82, master 71562a8b0

No new findings.

- Moved since 8306f1d52 (DPS specs): hunter/presets.ts (comment only), rogue/presets.ts (default gear and talents switched from Combat Daggers to Combat Swords, rotation preset order swapped; Consumes untouched), warrior/presets.ts (comment only), enhancement bis level60_phase0 (out of scope), hunter sv.apl and warrior fury APL (rotation, not this audit). No change in mksettings, sim/core consumes/buffs/debuffs, sim/common, forever_enchants.go, assumptions or the mage imbue. Moved level60 gear sets are all tank specs (out of scope).
- scan.sh on master: banned, Revelation, trinket and crit-elixir sections empty. Totem counts match the baseline rulings.
- Scan note: "minor-armor exclusive effects" prints 3 against the expected 2. The third is ForeverDustCloudAura (since dbafa638b, 2026-11-11), not a preset issue. The scan's expected count is stale.
- wago latest: 1.60.1.70888, same as last pass; no new build.
- Open items: the two low findings in findings/consumables.md are unchanged.
