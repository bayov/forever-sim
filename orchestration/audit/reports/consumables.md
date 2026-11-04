# Consumables pass 70, master fb584eafa (since f331f0cff)

Findings: none new. consumables-20261104-1 (Salvation ticked in presets.ts) is resolved: all 12 presets.ts dropped `salvation: true`; mksettings and the requests already had.

Verified:
- ui/*/presets.ts diffs (12 specs): the only Consumes/raid change is the Salvation removal; warrior also has APL/talent comment and talent-string changes (Fury 2H), no consumes.
- standards.md: new rules 38 (Forever only) and 39 (cast-time coefficient); nothing for this audit.
- scan.sh on master: banned, Revelation, trinkets and crit-elixir sections empty. Counts as last pass (minor-armor effects 3, the third is Dust Cloud, unused by presets; totems as expected).
- wago latest: 1.60.1.70755, nothing newer. Class branches are stale, not audited.
- Pass doc standing check "Salvation on" is now stale: it should say Salvation off (Forever removed it, 2026-10-30 hotfix).

No open items.
