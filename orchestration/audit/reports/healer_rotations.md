# Healer rotations pass 1 (findings: 1 medium, 1 low)

- Master: 1e9eb71f7 (first pass; `last.master` was 36cd929a2). Done first on a1b1163d2, then redone for the raid
  role on 1e9eb71f7 after the raid damage fix (b2c16f6, merged at c8f5858bd, healer_mechanics-20261104-1). The
  numbers below are from 1e9eb71f7 unless marked. Audited 2026-11-04.
- Bench: `tools/healbench` built from this checkout (`make sim/core/proto/api.pb.go`, then
  `go build -tags=with_db -o $AUDIT_WORK/healbench ./tools/healbench`), healbench's default seed, paired runs.
  Each request on each of the 7 boss files at the file's own length in its role (`-role tank` for the tank
  healers, `-role raid` for `*_raid`), `-carry`, 1,000 iterations. The mana side was run at `-dur 180` and
  `-dur 300` on Golemagg and Ragnaros, 500 iterations.
- New baseline: `ref/healer_bench/summary.tsv` on 1e9eb71f7 (63 rows: carry, EHPS, overheal, EH per mana, mana
  left, OOM time, deaths, death floor).

## What changed

- 36cd929a2..a1b1163d2: the audit loop's healer state files only. All 9 requests embed exactly their UI APL.
- a1b1163d2..1e9eb71f7: no healer APL, request, encounter or healbench change. Two things moved the bench:
  - **Raid role**: b2c16f6 (a share of a raid event under one player hits nobody). It moves every raid row
    up, by as much as +0.73 (Golemagg). That is the intended fix, not a finding.
  - **Tank role**: rows moved by up to 0.05, for example the Holy priest on Golemagg 1.915 -> 1.865 and the
    shaman on Golemagg 1.716 -> 1.752. The bench's tank is the Protection warrior (`tools/tankbench/req/warrior_protection.json`,
    unchanged). Its sim changed in the tank branch merge 6de9e64 (Forever threat, Bloodrage's and Reckless
    Swings' health costs on a tank, 20ad1aa and earlier). Healbench built at 20ad1aa^ already gives the Holy
    priest's 1.865. Then b2c16f6 moved the shaman's tank row 1.775 -> 1.752 (Golemagg), even though it only
    meant to change raid damage. These are tank-side changes, and the healer rotations are unchanged by them,
    so I filed no finding. The largest shift per preset is in the last column below.

## The baseline (carry, 1,000 it, 1e9eb71f7; length tank / raid)

| Boss | Tank: shaman | Holy priest | Disc priest | Paladin | Druid | Raid: shaman | Holy priest | Paladin | Druid |
|---|---|---|---|---|---|---|---|---|---|
| Onyxia (35 s / 137 s) | 1.154 | 1.240 | 1.225 | 1.097 | 1.164 | 0.923 | 0.947 | 0.984 | 0.977 |
| Broodlord (41 s / 44 s) | 0.839 | 0.891 | 0.859 | 0.754 | 0.732 | 0.879 | 0.925 | 0.906 | 0.906 |
| Chromaggus (115 s / 115 s) | 0.878 | 1.002 | 0.818 | 0.837 | 0.857 | 1.082 | 1.035 | 1.081 | 1.103 |
| Golemagg (56 s / 56 s) | 1.752 | 1.865 | 1.802 | 1.647 | 1.649 | 2.645 | 2.542 | 2.699 | 2.656 |
| Magmadar (53 s / 54 s) | 1.793 | 2.062 | 1.930 | 1.637 | 1.691 | 1.772 | 2.007 | 1.919 | 1.869 |
| Ragnaros (71 s / 97 s) | 0.511 | 0.578 | 0.553 | 0.510 | 0.517 | 1.203 | 1.304 | 1.302 | 1.287 |
| Vaelastrasz (17 s / 31 s) | 0.793 | 0.905 | 0.854 | 0.711 | 0.672 | 1.110 | 1.139 | 1.002 | 1.092 |

| Request | Overheal (range over 7 files) | Mana left (range) | Carry vs H3 table (a1b1163), largest |
|---|---|---|---|
| shaman_restoration | 1.9-10.7% | 2997-6732 | +0.036 (Golemagg) |
| priest_holy | 34.9-49.5% | 3058-6774 | -0.050 (Golemagg) |
| priest_discipline | 3.2-16.4% | 435-6316 | +0.030 (Onyxia) |
| paladin_holy | 14.8-22.6% | 4286-7249 | +0.022 (Golemagg) |
| druid_restoration | 8.9-23.3% | 4566-6398 | -0.016 (Vaelastrasz) |
| shaman_restoration_raid | 15.1-70.8% | 6928-7619 | +0.732 (Golemagg) |
| priest_holy_raid | 32.6-66.7% | 6956-7589 | +0.478 (Golemagg) |
| paladin_holy_raid | 16.9-22.8% | 7081-7971 | +0.632 (Golemagg) |
| druid_restoration_raid | 45.7-79.9% | 5121-7018 | +0.635 (Golemagg) |

Mana side, raid healers on the new raid damage (500 it, carry / mana left): no raid healer runs low.

| Request | Golemagg 180 s | Ragnaros 180 s | Golemagg 300 s | Ragnaros 300 s |
|---|---|---|---|---|
| shaman_restoration_raid | 2.090 / 7603 | 1.058 / 7411 | 1.865 / 7604 | 0.994 / 7403 |
| priest_holy_raid | 2.161 / 7601 | 1.215 / 6910 | 2.016 / 7587 | 1.168 / 6859 |
| paladin_holy_raid | 2.188 / 7969 | 1.199 / 7820 | 2.016 / 7963 | 1.170 / 7807 |
| druid_restoration_raid | 2.172 / 6942 | 1.178 / 5366 | 2.021 / 6909 | 1.127 / 3952 |

Tank healers at 180 / 300 s (a1b1163d2, 500 it, carry / mana left / OOM), Golemagg: shaman 1.469 / 5377 and
1.242 / 4597; Holy priest 1.483 / 5675 and 1.201 / 4520; Disc 1.098 / 2348 and 0.891 / 752, OOM 227 s; paladin
1.406 / 6738 and 1.324 / 6484; druid 1.535 / 5905 and 1.322 / 5711. On Ragnaros every tank healer ends with
5,700-6,700 mana.

The only healer that runs out of mana at a boss's own length is the Discipline tank healer on Chromaggus
(115 s): 431 mana left, out at 107 s (4,000 it). Finding 2.

## Cooldowns (one logged iteration near the carry scale, Golemagg 180 s, a1b1163d2)

| Preset | Cooldowns cast |
|---|---|
| Shaman tank | Nature's Swiftness 1, Mana Tide 1 (both on 3 min cooldowns) |
| Holy priest tank | Inner Focus 1; Power Infusion 0 (no talent, finding 1) |
| Disc priest tank | Inner Focus 1, Power Infusion 1 |
| Paladin tank / raid | Divine Favor 2 / 1, Lay on Hands 0 / 1 |
| Druid tank / raid | Innervate 1 / 1, Nature's Swiftness 1 / 1 (consumed only by Healing Touch or Regrowth, `sim/druid/healing.go:735`) |
| Shaman raid | Nature's Swiftness 1 (autocast), Water Shield 1, Mana Spring 1 |

## Changes tried (one line each; carry B-A, paired seeds, 1,000 it, mean of the 7 files at their lengths)

| Preset | Change | Mean | Rows |
|---|---|---|---|
| Holy priest raid | Inner Focus ahead of Flash Heal (when Flash Heal 7 fits), old line out | -0.008 | Ragnaros -0.023 |
| Holy priest raid | Prayer of Healing 5 ahead of Flash Heal | -0.094 | Magmadar -0.346, Golemagg -0.222, Vael +0.017 |
| Holy priest raid | Renew ahead of Flash Heal | -0.202 | Golemagg -0.563, Magmadar -0.438 |
| Shaman raid | Water Shield whenever off (was under 60% mana) | -0.010 | Vael -0.020 |
| Paladin raid | Divine Favor only above 50% mana | -0.000 | |
| Paladin raid | no Lay on Hands | -0.004 | Onyxia -0.010 |
| Disc priest tank | Power Word: Shield mana gate 0.7 (was 0.5) | +0.005 | Chromaggus +0.020 (4,000 it: +0.008) |
| Disc priest tank | Power Word: Shield mana gate 0.9 | +0.008 | Chromaggus +0.031 (4,000 it: +0.021), Vael +0.022, Golemagg -0.009 |
| *a1b1163d2:* Holy priest tank | Renew only once 60% of it is missing (was whenever off) | -0.097 | Magmadar -0.151; 300 s Golemagg +0.032 |
| *a1b1163d2:* Disc priest tank | the same Renew gate | -0.065 | Vael -0.165 |
| *a1b1163d2:* Disc priest tank | Flash Heal first gate 0.9 (was 0.7) / no Flash Heal first | -0.008 / -0.012 | OOM unchanged |
| *a1b1163d2:* Paladin tank | Divine Favor only above 50% mana | 0.000 | |
| *a1b1163d2:* Paladin tank | no Lay on Hands | -0.004 | Broodlord -0.023 |
| *a1b1163d2:* Paladin tank | Holy Light 9/8 and Divine Favor once 80% of Holy Light is missing | +0.005 | Golemagg +0.069, Vael -0.025; 180/300 s Golemagg -0.031/-0.036 |

The tank-role tests weren't re-run on the new master: their rows moved by no more than 0.05, and the orchestrator
said they stand. Mana side (a1b1163d2, 500 it, Golemagg and Ragnaros): Disc's shield gates 0.7 / 0.9 changed carry by
-0.001 / -0.003 at 180 s and -0.003 / +0.012 at 300 s, and it still ran dry on Golemagg at 300 s. The Holy tank
Renew gate gave -0.003 / +0.016, and the paladin's Holy Light at 80% gave -0.012 / -0.014.

No change wins by 0.02 or more on the mean. The Disc shield gate 0.9 is the best one measured (Chromaggus +0.021
at 4,000 it). It is the stopgap proposed in finding 2.

## Conclusions

- Apart from the two findings, every preset's rotation does what its notes say in game terms, and its
  cooldowns are used. The always-on tank Renew, Lay on Hands, the shaman raid healer's 60% Water Shield gate
  and Flash Heal first on the Holy raid healer all earn their place. Raid healers are idle most of a fight
  (30-70 casts in 180 s): their score is set by burst, not throughput or mana.
- `healer_rotations-20261104-2` (medium): the Discipline tank healer runs out of mana on Chromaggus at its
  own length (107 of 115 s), and on Golemagg at 300 s (227 s). No gate tried stops it.
- `healer_rotations-20261104-1` (low): the Holy presets carry lines that never fire. Power Infusion is there
  without the talent, and the raid healer's Prayer of Healing, Renew, Greater Heal and Inner Focus sit behind
  Flash Heal 4. The order is right; the notes and `docs/healing_priest.md` overstate what it does.
- Nit: the shaman raid healer's Riptide note still says "an instant heal". The notes' Riptide (the default)
  has none.

## Helpers (scratch, not committed)

- `bench.py`: every request × boss in its role, `-carry`, to a TSV.
- `variant.py`: one-line edits of a request's priority list, written to a copy.
- `cmp.py`: paired carry deltas against the baseline.
- `cdlog.sh`: one logged iteration's casts by the healer.

They take minutes to rebuild from this report. They belong in `tools/audit/healer_rotations/` once the
orchestrator allows it.
