# Gear pass 80, master bfbee78fcc

Trigger: 2026-11-25 Phase 0 price rerank touched all 27 DPS lists, gear sets, content checks, requests.

- phase0_check.py: 2 flags, 27 specs read. Destruction main hand Darkspear Insurgent's Spellblade (+1.75 over pick;
  rerank_warlock notes it as a -0.5 close call not taken). Enhancement main hand The Blackrock Slicer (+1.77; shaman_ret notes: 1.5 at 20000 iterations, unconfirmed, Brain Hacker stays). Both explained, not findings.
- No standards.md commits since last pass. No faction, set or request/gear-set mismatches flagged.
- List dps vs shared run: requests sit 0.1-1.6% (1-16 DPS) under the list dps for nearly all specs
  (Shadow -16.5, MM -14, Affliction -14.3, Feral -13.5; Enhancement -0.9, Fury -2). No sim/ change since list sim 34fe148 beyond class
  rerank commits. Balance rerank notes state the list dps is the search fight (use-adjusted) and the request matches its plain sim (802.1; I reran 20k iterations: 802.29). Treated as the settled search-fight difference; worth a worker note if the gap should be closed.
- Not done this pass: per-pick source rechecks, set-bonus counts, hit totals (checker clean; no new rulings).
