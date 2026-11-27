# Results audit report

## Pass 113 (branch phase0-regear 313069ce, 2026-11-27; scoped to the 29 builds re-simmed since 046d0e4b8)
- Run: wowsimcli at 313069ce, the 29 requests one at a time, 20k, seed 1, 180 s (`/tmp/claude-0/aud-res/sum29.txt`).
- builds.json vs run: 29 of 29 within 1.3 (max Fury 2H duty-sunder 957.0 vs 958.3). Provenance: all 29 at ee28c064, by sim, 20k; the other 39 are by hand at the same commit (as briefed). sim/proto/mksettings/database unchanged since ee28c064 (only a Talisman comment in 8d45b49e).
- Datasets: abilities, lengths (180 s equals the main run within 1.3), hist (20k), weights, mana, assumption_dps (base rows equal builds.json) are current for the 15 non-duty builds; duty rows carry main run, lengths and mana only, as documented. The three Armaments of Battle items are worn by no build in gear.json and appear in no meta text except verify notes.
- Drift report equals the reference (115 lines); unread fields: none.
- Races: builds.json race equals the meta race on all 29 (display names only differ). Warlock Demonology, Destruction and Affliction Shadow Bolt keep Orc against a best race of Undead (+2.6 to +6.7), unchanged since 046d0e4b8.
- Moves vs 046d0e4b8 (builds.json): Affliction -17.4, Affliction SB -7.2, Demonology -6.8, Balance -6.2, Feral -3.5, Fury 2H -2.9, Enhancement +5.9 to +6.3, Elemental +1.9, Smite and Destruction ~0. All explained by the gear change (313069ce) and the Talisman charge rule.
- Findings: metas lag (results-20261127-1, now 29 builds, 27 off by more than 0.5, max 18.4), duty verify blocks stale (results-20261127-2), Changes line without DPS (results-20261127-3).
- Not checked: Tanking and Healing modes (stale, not re-simmed) and the 39 hand-recorded builds.
