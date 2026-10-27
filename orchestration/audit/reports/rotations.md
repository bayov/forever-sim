# Rotations pass 34 (2026-10-27)

- **Master:** 99bd53ada (last audited f62cb7971). Client 1.60.1.70532 (unchanged).
- **Result:** quiet, 0 new findings. No open findings carried.

## What changed

- New preset Level 60 Shock (Ret side): `ui/retribution_paladin/apls/level60_shock_ret.apl.json`
  (+ `tools/dps_report/req/paladin__level_60_shock_ret.json`, talents 23/0/28, `tools/rotopt/paladin_presets.sh`).
  The req's embedded rotation equals the APL file.
- No non-test sim `.go` changed; the rest of the range is DPS Report tooling and state. Baseline
  `ref/rotations/baseline_f62cb79_3000it_seed424242.tsv` kept (no sim change).

## Audit of Level 60 Shock (Ret side)

Shared run (20k, seed 1, 180 s): 932.4 DPS, matches published. Per fight: Consecration 22.4,
Judgement of Command 22.4, Holy Strike 17.9, Holy Shock 17.5, Hammer of Wrath 4.9, Seal of Command
6.5 recasts (30 s seal, no lapse). OOM 0.01 s on average, so the mana is spent, not short.
Exorcism is inert against a non-Undead target, by design.

Paired A/B, 20k as 8 blocks, S0 = 1e6 (B−A ± SE at 120 / 180 / 300 s). All lose, so no second
seed set was needed:

| variant | 120 | 180 | 300 |
|---|---:|---:|---:|
| Holy Shock above Holy Strike | −1.9 ±0.3 | −1.6 ±0.5 | −0.7 ±0.4 |
| Holy Shock above Consecration | −4.1 ±0.4 | −1.2 ±0.4 | −2.2 ±0.5 |
| Judgement above Consecration | −1.9 | −1.8 | −1.1 |
| Consecration below Judgement/Seal/HoW/Strike/Shock | −28.4 | −26.3 | −13.0 |

The rotopt search's order holds. No correctness issues (spell ids, seal guard, trinket gate with
integer ×4÷5 arithmetic, pre-pull at −1.5 s landing before 0).

## Optional

None new. The earlier optional list in `passes/rotations.md` stands.
