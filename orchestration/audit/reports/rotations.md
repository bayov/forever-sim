# Rotations pass 79 (2026-11-25, master 0404484fb)

Result: quiet (sim changed, no new finding). Range 5536db4bc..0404484fb (24 commits).

## Checked

| Check | Result |
|---|---|
| `apl_logic_diff.py 5536db4bc 0404484fb` | 0 APL files touched |
| `tools/dps_report/req/*.json` changed | none |
| Non-test sim `.go` changed | rage.go + warrior/stances.go (Rage from hits taken: stance modifier, x2 for spells, ticks and specials, b4cab8498); world buffs and Spirit of Zandalar do nothing in a raid instance (3275a60b1, 9701cff08); buff party/raid scope vs client 71091 (5dbc01412); AI / Divine Spirit overwrite their stat's elixir (f1942793d); Armor Shatter and Puncture Armor share the minor armor slot (75dee7458); compsearch duty fillers (not read by the presets) |
| Regression baseline (68 presets, 3000 it, seed 424242) | refreshed: `ref/rotations/baseline_0404484_3000it_seed424242.tsv` replaces the eceb090 one |
| Resource check (shared 20k run, `secondsOomAvg`) | no build starves: worst warlock Affliction SB 4.5 s OOM, other warlocks about 3 s, Enhancement 2.5-2.9 s, feral 1.7 s, mages 0.6-1.1 s, all others under 0.4 s |
| Warrior rage (more Rage from raid hits): Fury Heroic Strike at rage >= 60 (req line 14) vs 50 and 70 | 60 holds, see below |

### Baseline moves vs eceb090 (3000 it, seed 424242)

Mostly the world-buff removal in raids: rogue Combat Swords -11.1%, hunters BM/MM -9.9 to -10.0%, the other physical
and caster builds down less. Warriors gain +0.4 to +0.8% (Berserker Stance x1.1 and the x2 for spell hits on the raid
hits' Rage); feral +0.3%. Against the published report (`runs/vs_published.txt`) Enhancement moves -2.2 to -2.6%,
paladin -1.0 to -1.7%, mage Arcane/Frostfire -1.5 to -1.7%, warriors +0.2 to +0.6%. These are sim mechanic changes,
not rotation faults.

### Fury Heroic Strike threshold (paired, 20k it, 8 blocks; B-A ± SE)

| Variant | Seed set | 120 s | 180 s | 300 s |
|---|---|---|---|---|
| HS >= 50 | 1e6 | -1.91 ±0.31 | -1.34 ±0.54 | -0.78 ±0.44 |
| HS >= 50 | 2e8 | -0.62 ±0.71 | -0.57 ±0.48 | +0.52 ±0.32 |
| HS >= 70 | 1e6 | -2.40 ±0.69 | -2.28 ±0.38 | -2.05 ±0.36 |
| HS >= 70 | 2e8 | -1.01 ±0.68 | -0.89 ±0.41 | -1.91 ±0.70 |

Baseline Fury at 180 s: 1015. Both directions lose; the extra Rage from raid hits doesn't move the threshold. The
full warrior Heroic Strike / Cleave re-search (open item, ruling 2026-11-17) is still owed to the next warrior search.

## Open findings

Unchanged: rotations-20261122-1 (Shock Ret ordering) and rotations-20261112-1 (rogue poison hands) have not landed
(no APL or request change in range).

## Optional

None new.

## Open items

Carried over from passes/rotations.md "Open items and pending rulings"; none other advanced this pass (the Phase 0
trinket check and Windfury timing remain to do).
