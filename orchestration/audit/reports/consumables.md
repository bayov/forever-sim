# Consumables pass 73, master a07941531

Changed since 0facd9428: sim/core/consumes.go, feral/hunter/ret presets.ts, hunter synthetic gear json, standards.md (tank text only).

| Check | Result |
|---|---|
| scan.sh on master | all banned/Revelation/trinket/crit-elixir sections empty |
| wago latest | 1.60.1.70790, no new build |
| consumes.go explosive-in-form change | refactor into explosivesUsableInForm; no value changes (Sapper, Land Mine, Dark Bomb, Thorium Grenade usable in form) |
| feral presets.ts | adds Goblin Sapper Charge; mksettings spec_feral.go does not (finding consumables-20261105-1) |
| hunter presets.ts | Survival talents string only |
| ret presets.ts | default gear/APL/talents switch (user ruling); consumes untouched |
| standards.md | rule 40 tank consumables text only |

Open: consumables-20261105-1 (low).
