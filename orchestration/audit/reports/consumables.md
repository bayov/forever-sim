# Consumables pass 87 (master cfc7f9382)

No new findings.

## Moved since 70ea588bb
- ui/{enhancement_shaman,feral_druid,retribution_paladin,warrior}/presets.ts: comments, APL tooltips only (rotation re-tunes for Windfury Totem and Judgement of Command). Consumes, raid buffs, debuffs, gear untouched.
- sim/core/buffs.go: Windfury Totem now rolls on every landed main hand melee hit (not a consumable value; totem AP +95/+179/+246 kept per ruling 2026-11-15).
- sim/common/item_effects.go: Silithid Queen's Husk (286590) +29 AP vs Humanoids, Rune of the Dawn 46 and Seal of the Dawn 78 (Forever values, cited from client 70938 in the code comments). None is in a preset, level60 gear set or mksettings (git grep).
- tools/dps_report/assumptions_def.py: new Blood Fury weapon and Execute Windfury rows, paladin Holy rows; no buff/oil assumption changed.
- orchestration/standards.md item 17 text only.

## Verified
- scan.sh on master: banned, Revelation, non-synthetic trinkets, double crit elixir sections all empty. Raid air/weapon totems match the standing check (melee: GoA + WF; hunter: GoA + WF per scan; enhancement: GoA only, own totems).
- Scan count "minor-armor exclusive effects" prints 3 (ref 2), but debuffs.go is unchanged since the last pass, so it is a scan-pattern artifact, not a change.
- wago latest build 70938, already audited. Class branches are stale copies, not audited.

## Open items
The two older low findings stay open (Inferior Stamina naming, feral mksettings Sapper).
