# Consumables pass 61 (master c23ef0440)

No new findings. Open: consumables-20261101-1 (Sapper on ranged hunters, low; unchanged, Sapper logic not touched).

Last audited master dd053d24c; client 1.60.1.70707 still newest (wago probe). No class branch ahead.
Moved watched files, all checked:
- sim/core/buffs.go: under Forever, Moonkin Aura and Leader of the Pack are each 3% melee+spell crit, raid wide, non-stacking (user ruling 2026-11-01, hotfix 113283 on 70707). Implemented as one 3% when either is on; Era split kept. Every preset and mksettings still has one of the two on (main.go raid has both; Balance and Feral switch both off and get their own aura from sim/druid/druid.go:137/141, which now also gives the other crit type, harmless; Warlock keeps Moonkin on, so the demon now gets 3% melee crit too, consistent with the ruling). This supersedes the pass doc's "Moonkin Aura / Leader of the Pack" settled bullet (spell-only / melee-only), which needs updating by the orchestrator.
- sim/common/item_effects.go: Kang the Decapitator proc added (Rend 100 x 5 ticks Forever); no preset or level60 gear set uses item 2291.
- tools/dps_report/assumptions_def.py, ui/hunter/presets.ts, ui/enhancement_shaman/presets.ts, standards.md: comment/text only (Cobra Reflexes wording, leveling-plan route, Rage crit rule). No consumes, buffs, debuffs or enchants changed.
- scan.sh on master: every ban section empty; minor-armor count 3 (known: FF, CoR, Tallstrider Dust Cloud; ref file's 2 predates Dust Cloud).
