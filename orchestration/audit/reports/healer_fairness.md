# Healer fairness pass 32 (2026-11-26, master ab3fff700)

Result: quiet. Due for one watched file, `sim/core/spell_result.go` (1223e764e, leveling big pulls): the new branch
(`onHitFromBehind`) sits inside `if leveling := sim.Encounter.leveling; leveling != nil`, which is set only when the
request has an `Encounter.Leveling` (`sim/core/environment.go:149`). No healbench request, boss file or the raid filler
has one, and no healer request, the filler or healbench changed in bfbee78fc..ab3fff700 (git diff empty), so the code
is inert for the bench; not re-benched. Table below is pass 31's, still current. Open findings unchanged:
-20261124-1, -20261115-1, -20261104-4.

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
