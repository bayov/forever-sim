# Healer mechanics: pass 33

2026-11-26T06:51:17Z, master ab3fff7007 (last bfbee78fcc), client 1.60.1.71139 (unchanged). Result: quiet.

## Why due

One watched file changed: `sim/core/spell_result.go` (1223e764e, leveling sim big pulls).

## Get current

| Source | Seen | Since last pass |
|---|---|---|
| Client build (`wago.py latest`) | 1.60.1.71139 | unchanged |
| Blizzard hotfix thread t/2403040 | last post 2026-11-25T23:25:08Z | nothing new (= dev_notes_seen) |
| `orchestration/update_ledger.md` | build 71139 section, Nov 25 hotfix rows | already covered by pass 32 (findings -18, -19); no new healer rows |

## What changed

| File | Change | Healer reach | Verdict |
|---|---|---|---|
| sim/core/spell_result.go:571-578 (dealDamageInternal) | under `sim.Encounter.leveling`, an enemy hit on the leveling player calls `onHitFromBehind` when `hitBehind` (daze roll); enemy-target onHit unchanged | none: `Encounter.leveling` is only set by `newLevelingSession` (environment.go:150), never in healing fights; heal and absorb paths untouched | no change |

No heal, healing talent or engine healing file changed; spell tables unchanged, so no per-spec tables re-checked this pass
(see pass 32's report in git history and `ref/healer_mechanics/verdicts.md`, unchanged).

## Open findings

None fixed on master: -19 (Oath of Light 60 min), -18 (Genesis on Tranquility), -17 (Mana Well energize under Innervate),
-11 (Riptide costs) and the older open items stay open.

## New findings

None.
