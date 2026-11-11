# tank_gear report, pass 6, master 707ae92cc5 (quiet)

Changed since 27d0d59d67 on the watched paths: only orchestration/standards.md (standard 25 amendment of 2026-11-11:
weapon imbue and weapon totem now both work; Enhancement/Flametongue wording). No change to tools/database tank sets,
ui gear_sets, gear_targets.md or db_inputs recipes. Tank standards (1b, 2, 4, 5a, 6, 7, 8a, 36) untouched, so the
earlier per-piece and per-set tables stand.

Re-ran tools/audit/tank_gear/{warrior,paladin,druid,shaman}.py on master:
- warrior: pieces within 1% linear, norm within 4%; set 760.0 (761.0); Sta 31.4 (31.4), Hit +1.1, Crit -1.3 (in tolerance).
- paladin: Sta 32.6 (32.4), all shares within 0.2 of hit-adjusted row; trinket norm +4.2%; pieces within 1% linear.
- druid: set 730.5 (729.3); Sta +1.4 (tol 1.5), Agi +2.8 (tol 3); trinket linear +1.4%, norm +2.6%.
- shaman: set 845.1 (844.1); all shares within 0.1; trinket 100.2% linear.
No findings; none open.
