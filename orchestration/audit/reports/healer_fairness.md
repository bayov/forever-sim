# Healer fairness pass 33 (2026-11-26, master f9fa79917)

Result: findings (1 new: healer_fairness-20261126-1, high). Still open: -20261124-1 (raid filler tracks the DPS report;
ruled, new evidence added), -20261115-1 (raid druid not re-tuned for the 0.066 Wild Growth), -20261104-4 (no healing by
class in the boss files).

Due for 22 watched files in 2af970d74..f9fa79917: the Phase 0 tank/healer merge (7 healbench requests, their bis and
gear_set files, standards 1b/8a/29/35/36/43 updates), healbench `main.go` (Phase 0 search plumbing), Water Shield's 4th
globe with Totem of Sustaining (7de8438a6), bear attacks.

Same terms (step 1): with `.raid.parties[0].players[0].equipment` dropped, all 7 changed requests are byte-identical to
2af970d74 (consumables, buffs, race, professions Engineering/Alchemy, rotation, assignment). Equipment: 16-17 real items
each; both druid requests still wear 15 synthetic pieces (990401-...). Standards 43 allows that until the druid's list
lands "and the page says so"; the page ranks the tank druid first with no note, and `changes.json` says the druid moved
to Phase 0 too, hence -20261126-1.

Numbers (step 2): all 63 cells re-benched (healbench `-tags with_db` in the build_cli.sh worktree of f9fa79917,
`tools/audit/healer_fairness/screen.sh`, seed 101, 1,000 it; `ref/healer_fairness/bench_f9fa799.tsv`). All 63 within
0.015 of healer_rotations' Phase 0 `ref/healer_bench/summary.tsv` (3e9748641), so the Water Shield globe and the sim
changes since then move nothing past noise. Means against pass 31 (synthetic sets) in brackets:

| Preset (role) | Onyxia | Brood | Chrom | Golemagg | Magmadar | Ragnaros | Vael | Mean (pass 31) | Largest move |
|---|---|---|---|---|---|---|---|---|---|
| priest_holy_raid (raid) | 0.971 | 0.904 | 1.108 | 2.516 | 2.000 | 1.330 | 1.253 | 1.440 (1.448) | magmadar -0.058 |
| druid_restoration_raid (raid, synthetic) | 0.953 | 0.893 | 1.095 | 2.494 | 1.824 | 1.238 | 0.963 | 1.351 (1.370) | golemagg -0.042 |
| shaman_restoration_raid (raid) | 0.954 | 0.891 | 1.132 | 2.438 | 1.750 | 1.210 | 1.062 | 1.348 (1.372) | golemagg -0.080 |
| paladin_holy_raid (raid) | 0.947 | 0.899 | 1.110 | 2.510 | 1.727 | 1.227 | 0.979 | 1.343 (1.392) | magmadar -0.145 |
| druid_restoration (tank, synthetic) | 1.387 | 0.885 | 1.011 | 1.902 | 2.049 | 0.623 | 0.931 | 1.255 (1.255) | none |
| priest_discipline (tank) | 1.354 | 0.926 | 0.866 | 1.938 | 2.015 | 0.601 | 0.893 | 1.228 (1.266) | magmadar -0.144 |
| shaman_restoration (tank) | 1.328 | 0.927 | 0.924 | 1.831 | 1.980 | 0.558 | 0.859 | 1.201 (1.276) | golemagg -0.144 |
| priest_holy (tank) | 1.316 | 0.881 | 0.878 | 1.861 | 2.011 | 0.567 | 0.841 | 1.194 (1.267) | golemagg -0.139 |
| paladin_holy (tank) | 1.201 | 0.869 | 0.834 | 1.676 | 1.788 | 0.548 | 0.785 | 1.100 (1.151) | golemagg -0.124 |

Explanations: the Phase 0 lists carry less than the synthetic Phase 1 sets, so every spec on one lost carry, unevenly
(tank shaman -0.075, Holy -0.073, paladin -0.051, Disc -0.038; raid paladin -0.049, shaman -0.024, Holy -0.008). The
druid's tank cells are identical (no request or druid healing change); its raid row's -0.019 is the filler mage's item
swap (evidence under -20261124-1). The druid's first place in the tank role and second in the raid role are its gear
(-20261126-1). Among Phase 0 healers: tank Disc 1.228 leads shaman and Holy; Disc's list is searched tank-only while
Holy's weighs both roles alike (standards 43), which likely explains part of Disc's 0.034 over Holy (owned doc, open
question 9; not a finding, the standard sets it). Raid: Holy first by 0.09; druid, shaman and paladin a tie in noise.

Logs: still blocked by -20261104-4. Roles (step 3): unchanged; tank presets benched in the tank role, raid in the raid role.

## Pass 31 (2026-11-25, master 85691655a)

Result: quiet. Still open: -20261124-1 (high, raid filler on Phase 0 gear; user ruled to keep it; new evidence added),
-20261115-1 (high, raid druid not re-tuned for the 0.066 Wild Growth), -20261104-4 (low, no healing by class in the
boss files).

Due because 23 watched files changed in c3bbe33ea..85691655a: the wowsims/forever evidence batches (22d2cf952,
b663145ad, 295206ddf, efeb06b95, 2f42f4a44, 60fde111f): spell resistances, item/enchant/set procs, consumables
(stun explosives, swing timer), Windfury Totem on extra attacks, Frost mage rules, warrior/bear tank threat and
damage taken, shaman Clearcasting (Elemental Focus, Lightning Bolt/Chain Lightning only).

Same terms (step 1): none of the 9 healer requests, the raid filler (`tools/dps_report/req/mage__frost.json`) or the
tank request changed (git diff empty). Consumable, buff and flag changes are DPS-side (explosives, Spell Piercing,
weapon proc rolls) and touch no healer request.

Re-benched all 63 cells (healbench `go build -tags with_db` in the build_cli.sh worktree of 85691655a,
`tools/audit/healer_fairness/screen.sh`, seed 101, 1,000 it; `ref/healer_fairness/bench_8569165.tsv`):

| Preset (role) | Onyxia | Brood | Chrom | Golemagg | Magmadar | Ragnaros | Vael | Mean (pass 30) | Largest move |
|---|---|---|---|---|---|---|---|---|---|
| priest_holy_raid (raid) | 0.988 | 0.938 | 1.071 | 2.508 | 2.058 | 1.355 | 1.216 | 1.448 (1.425) | magmadar +0.055 |
| paladin_holy_raid (raid) | 0.937 | 0.918 | 1.113 | 2.580 | 1.872 | 1.281 | 1.042 | 1.392 (1.388) | onyxia -0.051 |
| shaman_restoration_raid (raid) | 0.932 | 0.928 | 1.097 | 2.518 | 1.806 | 1.223 | 1.102 | 1.372 (1.354) | broodlord +0.076 |
| druid_restoration_raid (raid) | 0.948 | 0.916 | 1.113 | 2.536 | 1.815 | 1.267 | 0.995 | 1.370 (1.362) | vaelastrasz +0.031 |
| shaman_restoration (tank) | 1.395 | 0.960 | 1.005 | 1.975 | 2.091 | 0.578 | 0.928 | 1.276 (1.274) | vaelastrasz +0.008 |
| priest_holy (tank) | 1.385 | 0.922 | 0.914 | 2.000 | 2.144 | 0.589 | 0.914 | 1.267 (1.266) | vaelastrasz +0.010 |
| priest_discipline (tank) | 1.359 | 0.938 | 0.828 | 2.033 | 2.159 | 0.616 | 0.931 | 1.266 (1.263) | vaelastrasz +0.007 |
| druid_restoration (tank) | 1.387 | 0.885 | 1.011 | 1.902 | 2.049 | 0.623 | 0.931 | 1.255 (1.253) | magmadar +0.011 |
| paladin_holy (tank) | 1.253 | 0.872 | 0.898 | 1.800 | 1.876 | 0.569 | 0.786 | 1.151 (1.146) | golemagg +0.022 |

Tank means move +0.001 to +0.005, order unchanged (shaman, Holy, Disc, druid, paladin). Raid means move +0.004 to
+0.023, unevenly (Holy +0.023, shaman +0.018, druid +0.008, paladin +0.004); the raid shaman (1.372) now passes the
raid druid (1.370) by 0.002, inside seed noise, so the bottom two are a tie. Since only code moved and the raid row
moved more than the tank row, the likeliest cause is the raid side (filler mage rules, boss-spell resists against the
raid); not attributed (no per-commit builds). Recorded as evidence under -20261124-1, whose "shaman last" no longer
holds. No healer gained or lost a place past noise, so no new finding.

Logs (step 2): still blocked by -20261104-4. Roles (step 3): unchanged.
