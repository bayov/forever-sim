# Consumables pass 71, master 6b480b060 (since fb584eafa)

Findings: none new. No open items.

Verified:
- sim/core/buffs.go: Power Infusion now also raises healing dealt by 20% (client 10060, auras 79 and 136); the spell side stays 1.2. No DPS effect on the audited presets.
- tools/dps_report/assumptions_def.py: only the Paladin Two-Handed Spec text (hotfix push 113542); no buff or consumable assumption.
- ui/elemental_shaman, rogue, warrior presets.ts: comment and talent-string changes only; no Consumes, RaidBuffs or Debuffs change.
- healing_priest and holy_paladin presets and level60 gear: out of scope.
- standards.md: new rule 40 (Enhancement imbue pairings); nothing for this audit.
- No change in mksettings, level60 gear sets of the audited specs, or the other watched code files.
- scan.sh on master: banned, Revelation, trinkets and crit-elixir sections empty; counts as last pass (minor-armor effects 3, Salvation absent).
- wago latest: 1.60.1.70755, nothing newer. Class branches stale, not audited.
- Pass doc standing check "Salvation on" is stale: it should say Salvation off (Forever removed it in the 2026-10-30 hotfix).
