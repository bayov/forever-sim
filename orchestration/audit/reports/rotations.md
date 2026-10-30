# Rotations pass 40 (2026-10-30)

- **Master:** fb1485888 (last audited bbbaf2e3c). Client 1.60.1.70707 mage talents.
- **Result:** findings. Two new: rotations-20261030-1 (high, Ret paladin twist window) and
  rotations-20261030-2 (medium, Arms / Fury 2H Hamstring rage gate). rotations-20261028-1 (hunter
  Trueshot/Hawk caveats) still open.

## What changed

- `apl_logic_diff.py bbbaf2e3c origin/master`: only the mage files.
  - Arcane: the single-stack Arcane Blast fallback below 20% mana is gone (rotopt, +1.6).
  - Frost: Evocation dropped, and the Arcane Blast stack target is now 3 (it was 2), with Ice Lance's new 8 s cooldown.
- Sim `.go` that the presets reach:
  - f3017897f: Forever extra attacks keep the swing timer.
  - Windfury Totem's ICD is 100 ms.
  - Eviscerate takes 4% AP a combo point.
  - Imbues apply only to weapons they fit.
  - The mage 70707 talents: Accelerant, Precision covering all spells, Critical Mass at 3, Ice Lance cooldown.
  - The rest is threat only (mana-gain threat, Subtlety, Instrument of Law, Demonic Brand, shout threat) or raid sim / compsearch.

## Regression baseline (3000 it, seed 424242) vs baseline_d50909a

| group | change |
|---|---|
| warrior arms (4) | +88.6 to +91.3 (+9.6 to +10.3%) |
| warrior fury / fury_2h (6) | +62.0 to +65.6 (+6.3 to +6.9%) |
| hunter survival (3) | +43 to +44 (+4.3%) |
| paladin non-shock and shock (11) | +15 to +29.5 (+1.6 to +3.2%) |
| rogue (12) | +13.6 to +21.9 (+1.8 to +2.9%) |
| feral cat / -duty-ff | +22.6 / +26.0 |
| mage fire / frostfire / frost / arcane | -26.3 / -38.6 / -4.6 / +12.3 |

Everything else moved by less than 1. New baseline: `ref/rotations/baseline_fb14858_3000it_seed424242.tsv`.
The moves match the mechanic changes (extra attacks for warriors, rogues, paladins and SV; 70707 talents
for mages). Enhancement didn't move.

## A/B screens (paired, 8x2500, s0=1e6, 180 s unless noted)

| preset | variant | B-A |
|---|---|---:|
| warrior arms | Hamstring rage 30 -> 10 / 15 / 20 / 25 / 40 | +4.3 / +4.3 / +3.5 / +1.9 / -5.6 (finding 2) |
| warrior arms | Heroic Strike 90 -> 80 / 70 | -6.2 / -17.5 |
| warrior arms-duty-sunder | Hamstring 30 -> 10 | +2.9 |
| warrior fury_2h | Hamstring 20 -> 10 / 30 | +3.2 / -4.2 (finding 2) |
| warrior fury | Heroic Strike 80 -> 70 / 90; Hamstring 10 -> 25 | -1.1 / -2.3; -0.1 |
| mage frost | re-add the old Evocation line (<20% mana, >90 s left) / >30 s left | -0.8 / -17.8 (rotopt's drop holds) |
| paladin level_60 | twist window 1.5 s -> 1.4 / 1.6 / 1.7 / 1.8 / 2.0 / 2.5 / 3.0 / 10 | -0.2 / +2.2 / +2.7 / +3.6 / +4.9 / +8.3 / +6.6 / +13.9 (finding 1) |
| paladin sub_holy | twist 10 s | +14.9 |

The two findings were confirmed on s0=2e8. The pooled numbers at 120/180/300 s are in `findings/rotations.md`.

## Notes and open items

- **Paladin twist:** the "1.5 s sharp optimum" settled item no longer holds. The bbbaf2e CLI still has the
  10 s window losing (-4.0), so a sim change in this range flipped it. Most likely it's the extra-attack
  swing timer, although the Ret req has no Windfury. Worth a look from mechanics before the window is
  widened. The duty and undead variants weren't measured separately.
- **Rogue:** no rotation screen for Eviscerate's 4% AP. The finisher lines don't choose between Eviscerate
  and another finisher, so a gate shift is unlikely. Left for the next pass.
- **Fire, Frostfire and Arcane:** the 70707 rotopt re-searches were taken as they are, and no A/B was run.
- **Open items from earlier passes:** carried unchanged (Balance NE, holding Elemental Mastery, the warlock
  duty variants).

## Optional (under the bar)

- None new.
