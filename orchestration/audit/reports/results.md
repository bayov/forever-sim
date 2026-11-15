# Results pass 90 (master 70ea588bb)

Shared run (68 builds, 20k, seed 1) vs published:
- builds.json: 68 of 68 within 1.5 of the run. The 8c75031e2 republish carries the Feral re-sim (feral-cat 1155.8, duty-ff 1150.7).
- Metas: 41 of 68 off by more than 0.5, the same families as last pass (results-20261104-1/-2: warrior, Affliction, hunters, feral +16.8, rogues, Ret Shock crusader). No new family.
- Feral move explained: 0cc2c3c8b (an avoided Ferocious Bite refunds 80% under Forever; Sharpened Claws also adds spell crit, forms.go). Feral was 1151.5 after the natural-flask change (pass notes of 2026-11-11), now 1155.8 (+0.4%), so the move is small and in the expected direction. Only the feral builds changed.
- Drift (mksettings regen vs committed): equals the reference except the known feral sapper line (results-20261105-1), now on both feral requests.
- Unread fields: none besides notes (no output).
- sim/testdata/report_hashes.txt and sim/druid/feral test results changed with the Bite commit; no other watched path moved.
