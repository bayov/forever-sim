# Consumables pass 85 (master 1bf5feb4b)

Moved since 3a86d4a6b: ui/enhancement_shaman/presets.ts only (plus magma.apl.json, bis/level60_phase0.bis.json, not watched). wago latest 1.60 build is still 70938 (already audited).

- Change: new gear preset "Enh Imp GW (hit-1)" (level60_phase1_synthetic_hit-1.gear.json) and the Imp GW build uses it; a magma APL flameShockFirst comment. No Consumes, buff, debuff or enchant object touched.
- hit-1 gear vs base set: one slot differs, a synthetic 99xxxx item (990040 vs 990007), same enchant 1248640. No real non-relic items.
- scan.sh origin/master: banned, Revelation, trinket and crit-elixir sections empty.
- Class branches: stale copies, none audited (gotcha 1).

No findings. Open items: the two low nits in findings/consumables.md (enchant 41 name, feral mksettings Sapper) remain open.
