# Healer rotations pass 1 (findings: 1 low)

- Master: a1b1163d2 (first pass; `last.master` was 36cd929a2, the H3 druid doc commit). Audited 2026-11-04.
- Bench: `tools/healbench` built from this checkout (`make sim/core/proto/api.pb.go`, then
  `go build -tags=with_db -o $AUDIT_WORK/healbench ./tools/healbench`), healbench's default seed, paired runs.
  Each request on each of the 7 boss files at the file's own length in its role (`-role tank` for the tank
  healers, `-role raid` for `*_raid`), `-carry`, 1,000 iterations; the mana side at `-dur 180` and `-dur 300`
  on Golemagg and Ragnaros, 500 iterations.
- New baseline: `ref/healer_bench/summary.tsv` (63 rows: carry, EHPS, overheal, EH per mana, mana left,
  OOM time, deaths, death floor).

## What changed

36cd929a2..a1b1163d2 is one commit: the audit loop's state and findings files for the healer auditors. No
healer APL, request, `sim/**`, encounter or healbench change. All 9 requests embed exactly their UI APL
(priority list and prepull, notes aside).

## The baseline (carry, 1,000 it; length tank / raid)

The 63 cells equal the H3 carry table in `docs/healing_sim.md` to the third decimal: nothing drifted since H3.

| Boss | Tank: shaman | Holy priest | Disc priest | Paladin | Druid | Raid: shaman | Holy priest | Paladin | Druid |
|---|---|---|---|---|---|---|---|---|---|
| Onyxia (35 s / 137 s) | 1.141 | 1.230 | 1.195 | 1.081 | 1.165 | 0.678 | 0.709 | 0.702 | 0.711 |
| Broodlord (41 s / 44 s) | 0.837 | 0.891 | 0.852 | 0.750 | 0.731 | 0.817 | 0.879 | 0.838 | 0.828 |
| Chromaggus (115 s / 115 s) | 0.884 | 1.006 | 0.824 | 0.825 | 0.870 | 0.906 | 0.911 | 0.970 | 0.953 |
| Golemagg (56 s / 56 s) | 1.716 | 1.915 | 1.826 | 1.625 | 1.648 | 1.913 | 2.064 | 2.067 | 2.021 |
| Magmadar (53 s / 54 s) | 1.798 | 2.089 | 1.941 | 1.645 | 1.699 | 1.540 | 1.693 | 1.500 | 1.605 |
| Ragnaros (71 s / 97 s) | 0.521 | 0.582 | 0.562 | 0.521 | 0.517 | 1.171 | 1.255 | 1.244 | 1.240 |
| Vaelastrasz (17 s / 31 s) | 0.775 | 0.903 | 0.865 | 0.723 | 0.688 | 1.110 | 1.139 | 1.002 | 1.092 |

| Request | Overheal (range over 7 files) | Mana left (range) |
|---|---|---|
| shaman_restoration | 1.9-10.6% | 3041-6769 |
| priest_holy | 35.1-49.6% | 3130-6789 |
| priest_discipline | 3.1-16.3% | 422-6374 |
| paladin_holy | 15.1-22.8% | 4304-7243 |
| druid_restoration | 9.7-23.8% | 4598-6343 |
| shaman_restoration_raid | 15.1-49.7% | 6712-7516 |
| priest_holy_raid | 32.6-53.7% | 6283-7357 |
| paladin_holy_raid | 16.2-22.3% | 6104-7879 |
| druid_restoration_raid | 45.7-77.1% | 3743-6520 |

Mana side, 180 s, 500 it, -carry:

| Request | Golemagg carry / mana left / OOM | Ragnaros carry / mana left / OOM |
|---|---|---|
| shaman_restoration | 1.469 / 5377 / (811 s) | 0.466 / 6725 / (2118 s) |
| priest_holy | 1.483 / 5675 / (831 s) | 0.517 / 5697 / (1330 s) |
| priest_discipline | 1.098 / 2348 / (266 s) | 0.405 / 5717 / (1300 s) |
| paladin_holy | 1.406 / 6738 / (1499 s) | 0.456 / 6709 / (2091 s) |
| druid_restoration | 1.535 / 5905 / (1477 s) | 0.502 / 6133 / (1885 s) |
| shaman_restoration_raid | 1.526 / 7530 / (3484 s) | 1.048 / 7332 / (3106 s) |
| priest_holy_raid | 1.652 / 7396 / (3175 s) | 1.162 / 6740 / (1880 s) |
| paladin_holy_raid | 1.625 / 7915 / (3550 s) | 1.152 / 7720 / (3250 s) |
| druid_restoration_raid | 1.653 / 6415 / (2411 s) | 1.147 / 5046 / (1120 s) |

Mana side, 300 s, 500 it, -carry:

| Request | Golemagg carry / mana left / OOM | Ragnaros carry / mana left / OOM |
|---|---|---|
| shaman_restoration | 1.242 / 4597 / (892 s) | 0.432 / 6725 / (2721 s) |
| priest_holy | 1.201 / 4520 / (842 s) | 0.438 / 5697 / (1901 s) |
| priest_discipline | 0.891 / 752 / 227 s | 0.325 / 5717 / (1900 s) |
| paladin_holy | 1.324 / 6484 / (1913 s) | 0.441 / 6709 / (2580 s) |
| druid_restoration | 1.322 / 5711 / (1929 s) | 0.457 / 6133 / (2531 s) |
| shaman_restoration_raid | 1.478 / 7524 / (3575 s) | 1.015 / 7322 / (3453 s) |
| priest_holy_raid | 1.540 / 7390 / (3442 s) | 1.106 / 6673 / (2498 s) |
| paladin_holy_raid | 1.562 / 7882 / (3581 s) | 1.128 / 7701 / (3472 s) |
| druid_restoration_raid | 1.559 / 6286 / (2698 s) | 1.042 / 2651 / (682 s) |

At the boss files' own lengths (17-137 s) no healer is short of mana: every row ends with 3,000-7,900 mana and
a projected OOM past the fight. The one OOM is the Discipline tank healer on Golemagg at 300 s (227 s); at 180 s
it ends with 2,348.

## Cooldowns (one logged iteration near the carry scale, Golemagg 180 s)

| Preset | Cooldowns cast |
|---|---|
| Shaman tank | Nature's Swiftness 1, Mana Tide 1 (both on 3 min cooldowns) |
| Holy priest tank | Inner Focus 1; Power Infusion 0 (no talent, see the finding) |
| Disc priest tank | Inner Focus 1, Power Infusion 1 |
| Paladin tank / raid | Divine Favor 2 / 1, Lay on Hands 0 / 1 |
| Druid tank / raid | Innervate 1 / 1, Nature's Swiftness 1 / 1 (consumed only by Healing Touch or Regrowth, `sim/druid/healing.go:735`) |
| Shaman raid | Nature's Swiftness 1 (autocast), Water Shield 1, Mana Spring 1 |

Raid healers are idle most of a fight at the carry scale (30-70 casts in 180 s): the raid score is set by
burst (who dies between heals), not throughput or mana.

## Changes tried (one line each; carry B-A, paired seeds)

At the files' own lengths, 1,000 it, mean of the 7 files (largest rows):

| Preset | Change | Mean | Rows |
|---|---|---|---|
| Holy priest raid | Inner Focus ahead of Flash Heal (when Flash Heal 7 fits), old line out | -0.003 | Magmadar -0.025, Broodlord +0.006 |
| Holy priest raid | Prayer of Healing 5 ahead of Flash Heal | -0.094 | Golemagg -0.302, Magmadar -0.255, Vael +0.017 |
| Holy priest raid | Renew ahead of Flash Heal | -0.182 | Golemagg -0.476, Magmadar -0.337 |
| Holy priest tank | Renew only once 60% of it is missing (was whenever off) | -0.097 | Magmadar -0.151, Vael -0.140; 300 s Golemagg +0.032 |
| Disc priest tank | the same Renew gate | -0.065 | Vael -0.165; 300 s Golemagg +0.005 |
| Disc priest tank | Power Word: Shield mana gate 0.7 (was 0.5) | +0.003 | Chromaggus +0.020 |
| Disc priest tank | Power Word: Shield mana gate 0.9 | +0.008 | Onyxia +0.029, Broodlord +0.024, Ragnaros -0.013 |
| Disc priest tank | Flash Heal first gate 0.9 (was 0.7) | -0.008 | Broodlord -0.024 |
| Disc priest tank | no Flash Heal first | -0.012 | Magmadar -0.028 |
| Paladin tank | Divine Favor only above 50% mana (Holy Light's gate) | 0.000 | all 7 within 0.000 |
| Paladin raid | the same | -0.000 | |
| Paladin tank | no Lay on Hands | -0.004 | Broodlord -0.023 |
| Paladin raid | no Lay on Hands | -0.003 | |
| Paladin tank | Holy Light 9/8 and Divine Favor once 80% of Holy Light is missing | +0.005 | Golemagg +0.069, Vael -0.025; 180/300 s Golemagg -0.031/-0.036 |
| Shaman raid | Water Shield whenever off (was under 60% mana) | -0.016 | Golemagg -0.045, Onyxia -0.030, Chromaggus +0.023 |

Mana side (500 it, Golemagg and Ragnaros, B-A mean of the two): Disc's shield gates 0.7 / 0.9 -0.001 / -0.003 at
180 s, -0.003 / +0.012 at 300 s, and it still runs dry on Golemagg 300 s (235 s / 245 s); its Flash Heal gates
change nothing there (OOM 228 s). Holy tank Renew gate -0.003 / +0.016; paladin Holy Light at 80% -0.012 / -0.014.

None wins at both the boss lengths and the mana side by 0.02 or more, so none is a finding. Settling the two
small Disc numbers (+0.008, +0.012) would take 4,000 it; they're under the bar either way.

## Conclusions

- Every preset's rotation does what its notes say in game terms, its cooldowns are used, and none runs out
  of mana at its boss's length. The always-on tank Renew, Lay on Hands, the shaman raid healer's 60% Water
  Shield gate and Flash Heal first on the Holy raid healer all earn their place.
- Finding `healer_rotations-20261104-1` (low): the Holy presets carry lines that never fire: Power Infusion
  without the talent, and the raid healer's Prayer of Healing, Renew, Greater Heal and Inner Focus behind Flash
  Heal 4. The order is right (moving them up loses up to 0.18); the APL notes and `docs/healing_priest.md` overstate it.
- Open item (not a finding, nothing tried fixes it): the Discipline tank healer goes OOM at 227 s on Golemagg
  at 300 s (carry 0.891 against the Holy tank healer's 1.201; 1.098 against 1.483 at 180 s). One logged
  iteration at 300 s ×0.9: Disc spends 38.7k mana (Flash Heal 24.6k, Renew 6.3k, Shield 5.0k), Holy 34.7k on
  more healing. Healing per mana of the kit, not a line. Worth a look if long tank fights ever count.
- Nit: the shaman raid healer's Riptide note still says "an instant heal"; the notes' Riptide (the default)
  has none.

## Helpers (scratch, not committed)

`bench.py` (every request × boss in its role, `-carry`, to a TSV), `variant.py` (one-line edits of a request's
priority list into a copy), `cmp.py` (paired carry deltas against the baseline), `cdlog.sh` (one logged
iteration's casts by the healer). A next pass rebuilds them from this report in a few minutes; they belong in
`tools/audit/healer_rotations/` when the orchestrator allows it.
