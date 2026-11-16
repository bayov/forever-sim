# Consumables pass 88

Master f30078da1 (last audited cfc7f9382). Client 1.60.1.70938, no newer build.

## Findings
None new.

## Verified
- forever_enchants.json: meta build and fetch date only.
- standards.md 17: Windfury Weapon and Totem separate cooldowns ruling (no consumable effect).
- assumptions_def.py: wf_weapon_totem_icd text only.
- ui/warrior/presets.ts: default gear, APL and talents switched to two-hand Fury; consumes unchanged.
- scan.sh on master: banned, Revelation, trinket and crit-elixir sections empty; totem counts as expected.
- minor-armor exclusive-effect count is 3 (ref says 2); debuffs.go did not change since the last pass, so not new.

## Open items
consumables-20261110-1 and consumables-20261105-1 (both low) stay open.
