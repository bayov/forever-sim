# Healer fairness pass 12 (2026-11-13, master 97da1cca5)

Result: quiet. No new finding. Still open: -20261104-4 (low; the boss files' `raidHealing` holds healer counts by
class but no HPS by class, checked on all 7 files). Due because master changed `sim/paladin/unsealed_wrath.go`
(Windfury Totem swings store no Unsealed Wrath echoes) and `orchestration/standards.md` item 17 (no extra attack
rolls Windfury Totem; `sim/core/attack.go`, `buffs.go`). Both are melee rules; no healer preset melees.

Check: healbench built from 97da1cca5 (`-tags with_db`, build_cli worktree), all 9 requests in their own role on
the 7 boss files, seed 101, 1,000 iterations (`tools/audit/healer_fairness/screen.sh`): all 63 carry cells identical
to `ref/healer_fairness/bench_71562a8.tsv`. `tools/healbench/req/` unchanged since 71562a8b0, so sections 1 and 3
below still hold. The tables below are pass 11's, re-confirmed.

## 1. Same terms (the 9 requests)

Everything but the healer row is byte-identical in all 9 (md5 of the request with the player row nulled): encounter,
raid, buffs, debuffs, sim options. Player rows: level 60, Engineering + Alchemy, Blessing of Kings and Improved Wisdom,
reaction 200 ms (healbench overrides), 20 yd, channel clip 50 ms (null for the paladin, who channels nothing).
Consumables: Supreme Power, Greater Mageblood, Greater Cleric's Elixir, Greater Intellect, Sages, Cerebral Cortex,
Major Mana Potion, Demonic Rune in all 9; food and oil each preset's own pick by objective (Wicked Smoothie for the
druid, Runn Tum Tuber Surprise for the rest; Lesser Mana Oil for paladin raid and Holy tank, Wizard Oil elsewhere;
`docs/healing_sim.md`, Consumables). Race: Human priest/paladin, Skyborne Windshaper shaman/druid. Talent strings
match the shipped presets in `ui/*/presets.ts`. No unexplained difference.

## 2. The numbers (healbench -carry, 1,000 iterations, seed 101, each file at its own length)

Healbench built from 71562a8b0 (`-tags with_db`, in the build_cli worktree); rows in
`ref/healer_fairness/bench_71562a8.tsv` (summary.tsv format, tools/audit/healer_rotations/baseline.sh).

| Request | Onyxia | Brood | Chrom | Golem | Magma | Rag | Vael | Mean | Rank |
|---|---|---|---|---|---|---|---|---|---|
| priest_holy (tank) | 1.378 | .931 | .970 | 2.000 | 2.154 | .600 | .910 | 1.278 | 1 |
| priest_discipline (tank) | 1.363 | .958 | .871 | 2.027 | 2.143 | .602 | .938 | 1.272 | 2 |
| druid_restoration (tank) | 1.403 | .893 | 1.023 | 1.951 | 2.045 | .634 | .932 | 1.269 | 3 |
| shaman_restoration (tank) | 1.340 | .978 | .984 | 1.938 | 2.062 | .587 | .911 | 1.257 | 4 |
| paladin_holy (tank) | 1.241 | .883 | .908 | 1.806 | 1.867 | .568 | .789 | 1.152 | 5 |
| priest_holy_raid | .988 | .906 | 1.065 | 2.607 | 2.062 | 1.340 | 1.262 | 1.461 | 1 |
| paladin_holy_raid | 1.018 | .912 | 1.146 | 2.702 | 1.914 | 1.283 | 1.035 | 1.430 | 2 |
| druid_restoration_raid | .991 | .924 | 1.094 | 2.630 | 1.829 | 1.278 | 1.152 | 1.414 | 3 |
| shaman_restoration_raid | .946 | .885 | 1.115 | 2.674 | 1.887 | 1.236 | 1.138 | 1.412 | 4 |

Against `ref/healer_bench/summary.tsv` (8f3f74f3f): the 49 cells of the 7 unchanged presets are identical, so the
seen-health values, the Windfury rule and the rest of the merge move no healer. The two re-tuned raid presets land on
pass 9's predictions exactly (`retune_25f65da_s101.tsv`): Holy raid 1.440 -> 1.461 (Twilight Focus, +0.137 on
Vaelastrasz), paladin raid 1.400 -> 1.430 (Flash of Light, gains on Onyxia, Chromaggus, Vaelastrasz).

Spread: tank 1.152-1.278 (paladin 10% behind the leader, cast-time bound with mana to spare, as before; the other
four within 2%), raid 1.412-1.461 (3.5%). Nothing past the 15% line. Only the Disc tank preset runs out of mana
(Chromaggus).

## 3. What a role rewards

Each preset benched only in its own role (`_raid` in raid, the rest in tank), each file at its own length. Unchanged.

## Against the logs

Still can't be checked: the boss files keep no healing by class (-20261104-4). The logs' composition (about 5
priests, 3 shamans, 2 paladins, 1 druid per 11 healers) is not a ranking.

## Open items

- -20261104-4 (low): HPS by healer class in the boss files.
- `ref/healer_bench/summary.tsv` (healer_rotations' baseline) predates the two raid re-tunes; its next pass should
  regenerate it (this pass's rows are the same format).
