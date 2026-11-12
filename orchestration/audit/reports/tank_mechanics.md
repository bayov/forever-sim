# tank_mechanics pass 11 (2026-11-12, master cb0d928fd, client 1.60.1.70888+hf)

Due on 2 watched files changed (33db32b36..cb0d928fd): sim/druid/talents.go, sim/shaman/talents.go, both from
cb0d928fd "Talent tooltips: bring stale text in line with what the sim already does". Result: quiet.

**Getting current:** wago's latest is still 1.60.1.70888 (no client diff). No Blizzard post newer than pass 10's
(Nov 10 class tuning) recorded; no update_ledger entry in range.

**Changed code:**

| Change | Tank reach | Client / source | Sim | Verdict |
|---|---|---|---|---|
| druid/talents.go: Rend and Tear comment says the tree now shows 3% via notes/druid.json | cat abilities only (Claw, Shred, Rake, Rip, Bite); no bear spell | notes 2026-11-02 3%; client 2% | comment only, multiplier unchanged | n/a (comment) |
| shaman/talents.go: Talents.ImprovedFireNova -> TrialByFire (proto field 9 renamed, number unchanged) | Fire Nova / Magma totem only | client row unchanged | same field number, so talent strings and presets keep their value; fire_totems.go still calls the same helpers | no behaviour change |

Tooltip-only items in the same commit that touch tank talents, checked against the sim: Defiance 7/13/20% threat
(stances.go:118 uses 0.07/0.13/0.20 under Forever; tier 1 notes, client still 15% at rank 3, as pass 6);
Thick Hide 1/2/3 armor per level (bear.go:111 `thickHide * Level`). Both agree; the sim did not change.

**Tests:** `go test -tags=with_db` on sim/druid/..., sim/shaman/..., sim/warrior/..., sim/paladin/...: all pass
except sim/warrior/dps_warrior TestP1DPSWarrior (stale stats / DPS goldens, e.g. 64.344 -> 63.242; not a tank,
already noted in pass 10).

**Engine:** no sim/core tank file, tools/tankbench, tankscore, docs/tank_*.md, stances.go or bear.go change in
range; pass 7's checks stand.

**Findings:** none new. Still open: 20261111-1 (docs/tank_shaman.md stale Stormstrike / Mental Dexterity),
20261108-1 (no shock-under-mark threat test), 20261107-1 (stale "client disagrees" labels).
