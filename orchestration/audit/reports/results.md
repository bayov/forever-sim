# Results pass 109 (master 2af970d74)

Range ab3fff700..2af970d74 (47 commits, 176 watched files). Sim/proto/mksettings changes: Barov servant swing wait guard (barov_servant.go, no request moves),
proto weights_build field (UI only), mksettings Land Mine (Survival, Enhancement, feral) and Sapper (feral). Rest: Phase 0 requests and metas.

- Shared run (68, 20k, seed 1) vs metas: 14 of 68 differ by more than 0.5, largest +1.4 (warrior fury_2h-duty-sunder), all noise-level; the rest within 0.5.
- Drift (regen.py): identical to the reference. The Land Mine and feral sapper lines are gone, so results-20261126-1 and results-20261105-1 are resolved.
- builds.json vs shared run: 68 builds checked (compare.py builds against the shared summary; two fury_2h duty rows are missing from the reference summary, a reference gap).
- Unread fields: none printed.
- results-20261104-2 (metas and builds.json lag) is effectively closed: metas now within 1.4.
