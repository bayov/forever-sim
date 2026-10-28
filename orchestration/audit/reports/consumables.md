# Consumables pass 49 (master 5d27b0015, client 1.60.1.70634)

Result: quiet. No findings.

Moved since 37a8a8ccd: orchestration/standards.md and nine ui/<spec>/presets.ts (race re-picks per standards 29 and 16,
rotation comments and tooltips, one prettier reflow of EncounterLevel60SapperOnlyHumanoid). No Consumes, RaidBuffs,
Debuffs or IndividualBuffs object, gear set, mksettings or sim file changed.

Checked: scan.sh on master: banned, Revelation, non-synthetic trinkets and multi-crit-elixir sections empty; main.go
default MobTypeUnknown, Humanoid only as the alternate, Salvation and Sunder on; melee raid totems as ruled (Enhancement
Flametongue); wago latest is still 70634 (no new build).
Note: scan counts 3 MinorArmorReduction exclusive effects (ref says 2). The third is the hunter pet Dust Cloud
(debuffs.go ~942), untouched this period; the scan's expectation is stale, not a sim change.
Class branches listed "ahead" by changes.sh carry unmerged history (thousands of commits); not audited.

No open items.
