# healer_mechanics pass 32

Branch phase0-tankheal 3e9748641 (last: master 85691655a, client 71091). Quiet.

## Get current

- Client: newest build 1.60.1.71139. `wago.py diff --set mechanics 1.60.1.71091 1.60.1.71139`: SpellEffect ~23 (Sayge's
  Dark Fortunes, Fengus' Ferocity 230 -> 115, Enveloping Winds and Fist of Shahram targets, Reinforce, Reagent Economy,
  Weight of Duty, Explosive Runeshard), SpellMisc ~67 (rogue poisons and Vanish Attributes_5, world buffs Attributes_6,
  Holy Fire DurationIndex 21 -> 0), ItemEffect (Nutritious Food), ItemSparse (bonding and sell prices), CurvePoint 5.
  No heal, healing talent, Mana Spring / Mana Tide / Innervate or healer item is touched. None of the changed spells is
  in sim/ for a healer. Holy Fire is a damage spell (mechanics' scope). (`--by-spell` crashes on a table without
  SpellID: wago.py diff_by_spell KeyError 'SpellID'; the plain diff works.)
- update_ledger.md: no 71139 entry yet; nothing healing-related since the last pass.

## What changed on the watched paths

- eac7ba8b5: healbench now builds its raid through tools/healscore (shared with rotopt's healer objective). The
  druid_restoration tank rows of the new baseline are identical to the last one row by row, so the refactor kept the
  bench's results. docs/healing_sim.md gains "The rotopt objective" (search-side, no game rule).
- tools/healbench/req/*: equipment only (Phase 0 sets); healer_gear's scope.
- sim/druid/bear_attacks.go: bear threat, not healing.

## Spells, engine, tests

Nothing changed, so the per-spec tables of pass 31 stand. Open findings unchanged (Innervate on Mana Well,
Riptide cost, untested ranks, Vampiric Embrace 10%, Wild Growth coefficient, seen vs true health, priest golden,
NS and Wild Growth).

No new findings.

---

## Master's pass, carried over from origin/master: Healer mechanics: pass 33

Merged in from master's audit loop (synthetic-gear tank/healer rows; numbers there are master's baseline, not the
Phase 0 one above). Kept for the verdicts and finding notes that exist only on master.

2026-11-26T06:51:17Z, master ab3fff7007 (last bfbee78fcc), client 1.60.1.71139 (unchanged). Result: quiet.

### Why due

One watched file changed: `sim/core/spell_result.go` (1223e764e, leveling sim big pulls).

### Get current

| Source | Seen | Since last pass |
|---|---|---|
| Client build (`wago.py latest`) | 1.60.1.71139 | unchanged |
| Blizzard hotfix thread t/2403040 | last post 2026-11-25T23:25:08Z | nothing new (= dev_notes_seen) |
| `orchestration/update_ledger.md` | build 71139 section, Nov 25 hotfix rows | already covered by pass 32 (findings -18, -19); no new healer rows |

### What changed

| File | Change | Healer reach | Verdict |
|---|---|---|---|
| sim/core/spell_result.go:571-578 (dealDamageInternal) | under `sim.Encounter.leveling`, an enemy hit on the leveling player calls `onHitFromBehind` when `hitBehind` (daze roll); enemy-target onHit unchanged | none: `Encounter.leveling` is only set by `newLevelingSession` (environment.go:150), never in healing fights; heal and absorb paths untouched | no change |

No heal, healing talent or engine healing file changed; spell tables unchanged, so no per-spec tables re-checked this pass
(see pass 32's report in git history and `ref/healer_mechanics/verdicts.md`, unchanged).

### Open findings

None fixed on master: -19 (Oath of Light 60 min), -18 (Genesis on Tranquility), -17 (Mana Well energize under Innervate),
-11 (Riptide costs) and the older open items stay open.

### New findings

None.
