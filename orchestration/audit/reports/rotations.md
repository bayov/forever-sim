# Rotations pass 35 (2026-10-27)

- **Master:** 06a898751 (last audited d3df9247f). Client 1.60.1.70630 (ledger: 92a4d5331).
- **Result:** quiet, 0 new findings. No open findings carried.

## What changed

- `apl_logic_diff.py d3df9247 06a8987`: 0 APL files touched. The req changes are the Ret Shock
  `build` renames in four `.meta.json` files only.
- Non-test sim `.go` that the presets reach:
  - `sim/core/racials_forever.go`: the 2026-10-26 racial pass (Ruthless 1%, Human Spirit 10%, Dwarf
    Artisan's Precision 1% hit in place of Earth Fury, Orc Savage removed and Command +5% pet damage,
    Tauren Endurance health only, Wind Blessed 2%, Read Leyline +50% for 15 s).
  - `sim/mage/ignite.go`, `sim/mage/talents.go`: Arcane Instability and Arcane Power no longer raise
    Ignite ticks; Arcane Missiles' missiles no longer proc Missile Barrage (user ruling 2026-10-27).
  - `sim/warlock/talents.go`: Spellstone is Shadow damage only (client 70630).

## Baseline moves (3k it, seed 424242, vs `baseline_f62cb79`)

No req rotation changed, so every move is the sim. All are explained by the commits above:

| preset | old | new | Δ | cause |
|---|---:|---:|---:|---|
| mage__arcane | 983.3 | 898.0 | −85.3 | Missile Barrage chain from AM gone, AP off Ignite |
| mage__frost | 949.6 | 914.0 | −35.6 | Missile Barrage chain from AM gone, racials |
| mage__fire / frostfire | 1101.6 / 1096.4 | 1088.5 / 1086.2 | −13.0 / −10.2 | racial, Ignite |
| warlock (Spellstone presets) | | | −9.5 to −11.7 | Fire half of Spellstone gone, racials |
| everyone else | | | −3 to −11 | racial pass |

New reference: `ref/rotations/baseline_06a8987_3000it_seed424242.tsv` (f62cb79 file removed).

## Rotation checks after the Missile Barrage change

Missile Barrage now procs only from casts (Arcane Blast 20%/point, Frostbolt/Fireball/Arcane Wrath 10%,
Frostfire Bolt 10% + bonus). The AM-on-Missile-Barrage lines stay right (the missiles are free).
Paired A/B, 20k as 8 blocks, S0 = 1e6 (B−A ± SE at 120 / 180 / 300 s). All lose or trade off, so no
second seed set was needed:

| preset | variant | 120 | 180 | 300 |
|---|---|---:|---:|---:|
| mage__frost (A 911.3 at 180 s) | Arcane Blast stack target 2 → 1 | −16.6 ±0.3 | −16.7 ±0.3 | −14.3 ±0.3 |
| mage__frost | stack target 2 → 3 | +4.4 ±0.5 | −20.9 ±0.6 | −33.9 ±0.4 |
| mage__arcane (A 899.5 at 180 s) | Fireball (10151) filler in place of Frostbolt | −103.7 | −107.4 | −119.7 |

The frost stack target 2 holds (3 is a fight-length trade-off, not a finding).

Read Leyline (now +50% for 15 s) isn't cast by any preset rotation, so the nerf touches no APL.

## Handoffs (not rotation findings)

- Shadow warlock presets (affliction, demonology, destruction and their duty-curse copies) still take
  Spellstone, which lost its Fire damage. Whether Firestone (2% crit) now beats it is an imbue choice
  for the consumables pass; not measured here.

## Optional

None new. The earlier optional list in `passes/rotations.md` stands.
