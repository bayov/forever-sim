# tank_gear pass 8 (master 71562a8b03)

No file under ui/*/gear_sets or tools/database/forever_synthetic_items_*_tank.go changed since 8306f1d52 (git diff). Changes were db_inputs/Wowhead/seen-in-game data only; enchant ids used by the sets unaffected.

Re-ran tools/audit/tank_gear/{warrior,paladin,druid,shaman}.py (outputs in $AUDIT_WORK/tg_*.txt).

| set | pieces lin vs target | norm worst | set total vs target | stat shares vs Tanks row | trinket |
|---|---|---|---|---|---|
| warrior | 99.2-100.4% | Finger II +5.3% | 760.0 / 761.0 | max diff 1.3 (crit), Sta +0.0 | 46.0 lin, norm +4.2% |
| paladin | 99.1-100.5% | Finger +4.0% | 841.8 / 844.1 | max diff 0.2 | 46.0 lin, norm +4.2% |
| bear | 0.1-1.4% | Finger II -4.6% | 730.5 / 729.3 | Agi +2.8, Sta +1.4 (tol 3 / 1.5) | +1.4% lin, +2.6% norm |
| Warden | 99.6-100.5% | +3.3% | 845.1 / 844.1 | max diff 0.1 | 100.2% lin |

All within tolerance; no open findings; none newly resolved. Not rechecked: Tanks-row derivation (gear_targets.md unchanged).
