# Healer fairness pass 35 (2026-11-27, master 5ab589d85)

Result: quiet (no new finding). Resolved: -20261127-1 (paladin and shaman requests synced to their gear sets, d74a53261)
and -20261126-1 (the druid's Phase 0 list, eef393de0; page republished, 12fba196b). Still open: -20261124-1 (raid filler
tracks the DPS report; ruled), -20261115-1 (raid druid not re-tuned for the 0.066 Wild Growth), -20261104-4 (no healing by
class in the boss files).

Due for 14 watched files in 34994816e..5ab589d85: the druid's Phase 0 bis list and gear set and both druid requests
(eef393de0, 3a0a00a03), the paladin and shaman requests (d74a53261), the priest tank APLs and bis lists (ff87e66b3).
The one sim change in the range (`sim/common/item_effects.go`, Talisman of Ascendance's charges) reaches no healer
request (item 237283 is in none).

Same terms (step 1): every request's equipment now equals its preset's gear set item for item (all 9 checked with jq).
Once equipment is dropped, the druid, paladin and shaman requests are byte-identical to 34994816e; the two priest tank
requests differ only in `rotation` (ff87e66b3's paced Flash Heal lines). Consumables, buffs, race, professions, level,
heal assignment and reaction settings unchanged. No healer wears a synthetic set any more.

Numbers (step 2): all 63 cells (healbench built `-tags with_db` at 5ab589d85, `tools/audit/healer_fairness/screen.sh`,
seed 101, 1,000 it; `ref/healer_fairness/bench_5ab589d.tsv`) against pass 34's `bench_3499481.tsv`:

| Request (role) | Pass 34 | Pass 35 | Largest cell move | Why |
|---|---|---|---|---|
| priest_discipline (tank) | 1.228 | 1.234 | broodlord +0.015 | APL (paced Flash Heal) |
| shaman_restoration (tank) | 1.201 | 1.206 | golemagg +0.019 | gear set sync |
| priest_holy (tank) | 1.194 | 1.200 | chromaggus +0.034 | APL |
| druid_restoration (tank) | 1.255 | 1.188 | golemagg -0.113 | synthetic -> Phase 0 list |
| paladin_holy (tank) | 1.100 | 1.106 | golemagg +0.028 | gear set sync |
| priest_holy_raid (raid) | 1.440 | 1.440 | none | unchanged |
| shaman_restoration_raid (raid) | 1.348 | 1.346 | golemagg +0.013 | gear set sync |
| paladin_holy_raid (raid) | 1.343 | 1.343 | onyxia +0.008 | gear set sync |
| druid_restoration_raid (raid) | 1.351 | 1.340 | magmadar -0.090 | synthetic -> Phase 0 list |

The paladin and shaman cells equal pass 34's `reqsync_3499481.tsv` exactly, and all nine means equal
`ui/dps_report/data/heal_builds.json` (12fba196b) to the third decimal.

Ranking: tank Disc 1.234, shaman 1.206, Holy 1.200, druid 1.188, paladin 1.106; raid Holy 1.440, shaman 1.346,
paladin 1.343, druid 1.340. The druid's synthetic-set lead is gone (first -> fourth tank), as pass 33 estimated. Spread:
tank first to last 11.5% (the paladin, cast-time bound as before), the middle three within 0.018; raid Holy 7% ahead of a
three-way tie inside 0.006 (noise). Nothing past the 15% mark, so no new explanation owed.

Roles (step 3): unchanged; each request benched only in its own role. Logs (step 2): still blocked by -20261104-4; the
kills' mix (5.1 priests, 3.0 shamans, 1.7 paladins, 1.2 druids per 11) agrees with priests on top and doesn't speak
to the rest.
