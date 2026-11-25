# Healer fairness pass 30 (2026-11-25, master 0404484fb)

Result: quiet. Still open: -20261124-1 (high, raid filler on Phase 0 gear; user ruled to keep it, re-bench when the
healer report moves to Phase 0), -20261115-1 (high, raid druid not re-tuned for the 0.066 Wild Growth),
-20261104-4 (low, no healing by class in the boss files).

Due because standards.md, `sim/core/spell_result.go` and the shaman air/fire totem files changed since 5536db4bc.
What reaches the healing bench in 5536db4bc..0404484fb:

- b4cab8498: warrior Rage from hits taken follows the stance and doubles for spells, ticks and specials
  (`PreAbsorbDamage` in spell_result.go is read only by `core/rage.go`). Healbench's tank is
  `tools/tankbench/req/warrior_protection.json`, so the tank's Rage, rotation and random stream move.
- f1942793d: Arcane Intellect and Divine Spirit overwrite elixirs of their stat. All 9 healer requests wear Elixir of
  Greater Intellect and Elixir of Sages under both buffs, so every healer loses them alike (same terms).
- 3275a60b1 / 9701cff08: world buffs do nothing in a raid; no healer request ticks one (checked with jq), inert.
- 5dbc01412 and the totem files: scope check and comments only, nothing moves.
- standards.md: the Phase 0 gold-line rule and the Rage-taken text; no healer rule changed.

Re-benched all 63 cells (healbench `go build -tags with_db` at 0404484fb, `tools/audit/healer_fairness/screen.sh`,
seed 101, 1,000 it; `ref/healer_fairness/bench_0404484.tsv`):

| Preset (role) | Onyxia | Brood | Chrom | Golemagg | Magmadar | Ragnaros | Vael | Mean (pass 28) |
|---|---|---|---|---|---|---|---|---|
| shaman_restoration (tank) | 1.393 | 0.959 | 1.006 | 1.976 | 2.085 | 0.580 | 0.920 | 1.274 (1.280) |
| priest_holy (tank) | 1.384 | 0.920 | 0.911 | 2.000 | 2.148 | 0.597 | 0.904 | 1.266 (1.278) |
| priest_discipline (tank) | 1.354 | 0.935 | 0.828 | 2.034 | 2.153 | 0.613 | 0.924 | 1.263 (1.272) |
| druid_restoration (tank) | 1.386 | 0.884 | 1.011 | 1.902 | 2.038 | 0.620 | 0.931 | 1.253 (1.268) |
| paladin_holy (tank) | 1.250 | 0.873 | 0.898 | 1.778 | 1.875 | 0.569 | 0.781 | 1.146 (1.152) |
| priest_holy_raid (raid) | 0.969 | 0.927 | 1.055 | 2.484 | 2.003 | 1.337 | 1.203 | 1.425 (1.432) |
| paladin_holy_raid (raid) | 0.988 | 0.882 | 1.149 | 2.542 | 1.871 | 1.264 | 1.021 | 1.388 (1.387) |
| druid_restoration_raid (raid) | 0.961 | 0.901 | 1.133 | 2.537 | 1.790 | 1.246 | 0.964 | 1.362 (1.372) |
| shaman_restoration_raid (raid) | 0.921 | 0.852 | 1.107 | 2.507 | 1.781 | 1.212 | 1.096 | 1.354 (1.344) |

Means move -0.015 to +0.010; both rankings are unchanged (tank: shaman, Holy, Disc, druid, paladin; raid: Holy,
paladin, druid, shaman). Cells move both ways (largest: tank Holy Chromaggus -0.059, tank druid Golemagg -0.050,
tank Disc Chromaggus -0.043), the pattern of a changed tank (warrior Rage) plus a small uniform mana loss (the
elixirs), not of one healer singled out. Not attributed between the two (no revert build); no healer gained or lost
a place, so no finding.

Same terms (step 1): the 9 healer requests are unchanged since pass 11. Logs (step 2): still blocked by
-20261104-4. Roles (step 3): unchanged.

Note for healer_rotations / consumables: Elixir of Greater Intellect and Elixir of Sages are now dead in every healer
preset (Arcane Intellect and Divine Spirit are on). Equal for all healers, so not a fairness finding here.
