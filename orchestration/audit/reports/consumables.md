# Consumables pass 75 (master 9ba00c2dc, client 1.60.1.70862, unchanged)

Findings: none new. Open: consumables-20261107-1 (Natural flasks halved, sim/core/consumes.go:83-115 still 60 Stamina and 4-5 bonuses) and consumables-20261105-1 (feral mksettings Sapper; spec_feral.go still has no SapperExplosive).

Verified since 73d5bfa9e
- changes.sh: audited master files moved: sim/common/item_effects.go, sim/core/totem_weapon_buffs.go, tools/dps_report/assumptions_def.py, tools/rotopt/mksettings/spec_enh.go, ui/enhancement_shaman/presets.ts (plus orchestration/standards.md). Class/consumes branches listed as "ahead" are stale diverged copies (gotcha 1); not audited.
- Enhancement raid change (user saw on the beta 2026-11-08 that Windfury and Flametongue Totem replace each other): presets.ts and spec_enh.go both drop the raid's Flametongue Totem and keep Grace of Air; they agree. No enhancement dps_report request carries totemWeaponBuff. Other melee (warrior, rogue, Ret, Feral, Survival) keep totemWeaponBuff Windfury in both presets and mksettings. standards.md 25 and 40 and the sham_ftt_stack assumption describe the new rule.
- sim/core/totem_weapon_buffs.go: slot is now last-placed-wins; no number changed. item_effects.go: one RNG call changed (Sulfuras-style six-spell roll), no value change.
- scan.sh on master: every section above "== counts" empty.
- wago: no build newer than 1.60.1.70862; no builddiff needed.

Note for the pass doc (not edited, not an owned doc): the standing check "Enhancement gets ... Flametongue Totem (rank 4)" is out of date; Enhancement now gets Grace of Air only.

Open items: the two findings above. No sims were run.
