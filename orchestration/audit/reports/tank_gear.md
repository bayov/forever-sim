# tank_gear report, pass 7, master 55cd92b09 (findings)

Forced re-audit after the 2026-11-12 tank fixes (flask re-picks, Warden talents/knobs, paladin Wizard Oil). `git diff --stat
cb0d928fd HEAD` changes no tank synthetic item, ui gear_sets file, gear_targets.md or db_inputs recipes; only tankbench requests
(flasks, imbue, Warden talents/APL) and docs.

## Piece and set checks (tools/audit/tank_gear/{warrior,paladin,druid,shaman}.py)
| Set | Result |
|---|---|
| warrior | pieces 99.4-100.3% linear, norm within 4.2%; set 760.0 (761.0); Sta 31.4 (31.4); Hit +1.1, Crit -1.3 |
| paladin | Sta 32.6 (32.4); all shares within 0.2 of hit-adjusted row; pieces within 1% linear; trinket norm +4.2% |
| bear | set 730.5 (729.3); Sta +1.4, Agi +2.8 (tol 3); trinket linear +1.4% |
| Warden | set 845.1 (844.1); all shares within 0.1; trinket 100.2% |
Unchanged from pass 6: quiet on the set shape.

## Best-for-objective re-check after the flask and talent changes
Method: every request in tools/tankbench/req/{warrior,paladin,bear,warden} x 3 presets, one enchant swapped at a time
(cloak 1219587/1889, wrist 1248665/1886/1885, boots 1887/929, shield 1886/863, gloves 1248640/2564), `abtest` 3,000 seed 1
screen, candidates re-run at 10,000 on seeds 1 and 2 (outputs /tmp/audit-work/tank_gear/screen1.txt, conf.txt). Gloves 2564
(Superior Agility) scores +25 to +79 at Max survival/Balanced but its formula (20731) drops only from AQ40 mobs: not Phase 1, excluded.
Bear and paladin: every enchant holds at every preset (paladin cloak, wrist, shield, boots all win; bear 1885 at Max threat,
1248665 elsewhere). Titans/Supreme Power at Max survival did not flip any enchant. Trinket/talent/flask picks belong to the
rotation and fairness passes and were not re-opened.
Differences found: see findings 20261112-1 (boots), -2 (warrior wrist), -3 (Warden Max threat wrist/cloak, Balanced cloak).
All are low (+3 to +12 on scores 1,100-3,200).
