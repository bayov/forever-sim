# tank_mechanics pass 7 (2026-11-08, master 9ba00c2dc, client 1.60.1.70862+hf)

Due on 11 watched shaman files changed (73d5bfa9e..9ba00c2dc). Result: findings (one, low).

**Getting current:** wago's latest is still 1.60.1.70862 (pass 6 covered it); the loop checked the dev notes today (state `checked_at`). No ledger entries in range.

**What changed** (4 commits touching sim/shaman; read the hunks plus the helpers they call):

| Commit | Change | Tank effect | Verdict | Tier |
|---|---|---|---|---|
| 5a870eccc | Stormstrike mark applied after CalcDamage (`applyStormstrike`, stormstrike.go:144), threat redone via ThreatFromDamage | Warden shocks: same damage and threat (multiplicative, same formula as spell_result.go:428) | match | 1 (live logs) |
| 2a60326a1 | `sim.Roll` rounds to whole points; pushback rolls moved to RandomFloat | No tank proc/probability uses sim.Roll (grep of core, warrior, druid, paladin, shaman); Dragonbreath Chili 61-69 fine | match | 1 (live logs) |
| 0193721bc | Downranking penalty on heals only | Warden shock ranks keep full coefficients | settled ruling | user 2026-11-08 |
| 9ba00c2dc | Windfury/Flametongue Totem share one weapon slot, last placed wins | Warden presets: Rockbiter + raid Windfury Totem, own Searing/Fire Nova/Mana Spring; no Flametongue Totem | not affected | beta (user) |

Golden deltas (TestWardenShaman, TestProtection): dps/tps move under 0.01, consistent with rounding only.

**Engine:** sim/core changes in range are cast.go (RNG label), downrank.go (scope), sim.go (Roll rounding), totem_weapon_buffs.go (slot). None touches the attack table, armor, block, Rage, threat stances/forms, the tank sheet or the score replay.

**Tests:** Stormstrike's own threat is pinned (warden/forever_test.go:263-271, :402); a shock under the mark is not. New finding tank_mechanics-20261108-1 (low).

**Still open:** tank_mechanics-20261107-1 (stale 30%/50% labels in stances.go:110, bear.go:57, docs/tank_druid.md:22/:55) not yet fixed.
