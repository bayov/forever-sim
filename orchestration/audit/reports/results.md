# Results pass 74, master f23ea25d2 (2026-11-03)

- Shared run (68 requests, 20k, seed 1): all within 0.15 of published except the six shaman metas (still the 2026-11-02 values on purpose; results-20261102-1 update).
- Warlock: Affliction is now the Wrack build (Imp sacrificed, Siphon Life + Wrack in the APL): 1034.1 vs published 1034.1; Affliction Shadow Bolt 1015.6 (renamed from affliction_wrack/old Wrack 999.9); duty variants 1025.1 and 1019.9 match. mksettings spec_warlock.go and the requests agree (0 drift).
- Drift report vs reference: only two differences, both explained: rogue__subtlety now regenerates (5b766ca2b) and warlock__affliction_wrack renamed affliction_shadow_bolt. Expected duty drift unchanged.
- sim/compsearch/pool.go leftOut key renamed to match.
- results-20261103-1 resolved. results-20261103-2 (hunter ungated shares) and results-20261102-1 (shaman metas) remain open.
