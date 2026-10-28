# Consumables report, pass 52 (master dd663e6522)

Quiet. No findings.

- standards.md: added rule 31 (rotation intent) and a Report 30 sourcing note. No consumable, buff or debuff content.
- assumptions_def.py: wording-only rewrite of labels, short/desc/src texts. oil_proc_damage, hunter_tsa_hm_rank and every other buff-related setting keep their values (diff checked).
- scan.sh on master: banned, Revelation, non-synthetic trinket and multi-crit-elixir sections empty. Totems as expected.
- Latest wago build is 1.60.1.70658 (already recorded); nothing newer.
- Note: scan.sh's "minor-armor exclusive effects (expect 2)" prints 3 because ForeverDustCloudAura (a pet ability, debuffs.go) also shares MinorArmorReduction. Faerie Fire and Curse of Recklessness still share the slot; the scan expectation is stale, not a bug.
- Class branches show as ahead of master only through stale copies of shared files; not audited (gotcha 1).

No open items.
