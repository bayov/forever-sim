# Consumables pass 41 (master 016daa654, client 1.60.1.70532)

Result: quiet. No new findings, nothing open.

## Checked
- Moved watched files since 54878fa3d: ui/hunter/presets.ts, tools/rotopt/mksettings/spec_hunter.go, orchestration/standards.md (standards 23 hunter line). Nothing else under the watch list moved; the class branches show only stale copies of other classes' files (gotcha 1).
- Hunter duty change (user ruling 2026-10-25): the Hunter's Mark and Scorpid Sting duty presets, their encounters and builds are removed. Every hunter now keeps up its own mark (raid Hunter's Mark Missing in presets.ts, hunterRangedRaid and hunterMeleeRaid, and the four tools/dps_report/req/hunter__*.json requests) and the raid keeps Scorpid Sting on (true everywhere). All seven hunter APLs still cast Hunter's Mark. presets.ts and mksettings agree. Matches the pass doc's standards 23 line.
- No consumes, enchant, gear or trinket change in the hunter diff. scan.sh: banned items, Revelation, trinkets, double crit elixir all empty. Minor-armor exclusive effects: 3 (Dust Cloud, expected since pass 40).
- Client: no newer build probed this pass beyond 1.60.1.70532 in state.

## Open items
None. Housekeeping: pass doc Hunter's Mark +143 and minor-armor count 3 (from pass 40); stale drift/rank_tuning refs still name hunter *_mark/_scorpid specs (other auditors' refs).
