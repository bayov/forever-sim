# Consumables pass 74 (master 12aa2796f, client 70790 -> 70862)

Findings: consumables-20261107-1 (high, estimated): the four Natural flasks were halved in 70862.

Verified
- changes.sh: no commits to audited paths on master since c67d9e014 (class/consumes branches listed are stale, diverged copies; not audited, gotcha 1).
- scan.sh origin/master: every section above == counts empty; counts as before (minor-armor effects 3 per file text, the earlier baseline pattern; totem set unchanged).
- builddiff 70790 -> 70862: 15 consumable spell changes. Four flasks (1293740-1293743) halved: Stamina 60->30, zone bonus 5/4/5/5->2. Others are new "Endless" elixirs/scrolls (289082-289088, 289321, 289329, 289330, 289334) and a Wormhole item; no preset uses them. Forever enchants: 54 checked, 0 changed.
- Open finding consumables-20261105-1 (feral mksettings Sapper) still open.

Open: the flask finding; no sims were run.
