# Mechanics pass 51 (2026-10-30, master fb1485888)

Range audited: bbbaf2e3c..fb1485888. 4 code commits. TASKS.md / update ledger: MAGE-U-20 to 27 landed (class side, mage worker); no `forever-update/*` merge.

## Get current

- Client: 1.60.1.70707, unchanged (`wago.py latest --probe 300`).
- Dev notes thread 2360696: still 8 posts; post 7 (upd 10-27 16:47:14) and post 8 (upd 10-27 16:46:55) unchanged.
- Blue posts: none newer than 10-29 21:00:50 in category 349's active topics (checked "Thank You For Testing", only Kaivax's 10-23 posts).

## Code range

| Commit | Change | Verdict |
|---|---|---|
| f3017897f | Windfury Totem proc ICD 1.5 s -> 100 ms under Forever | tier 1 match: 70707 10612 aura 42, proc 20%, icd 100; the Icd is the shared proc cooldown in CreateExtraAttackAuraCommon |
| f3017897f | Extra attacks keep the swing timer under Forever | user ruling (wowsims/forever, notes.md to verify); settled, not a finding |
| f3017897f | Eviscerate 4% AP a combo point under Forever | user ruling (wowsims/forever estimate); 31016 has no AP coefficient in the client |
| f3017897f | Stones / oils / scrolls only on weapons that fit | masks 33219 = axe, 2H axe, polearm, sword, 2H sword, dagger; 9264 = mace, 2H mace, staff, fist: match the code's edged / blunt lists |
| e2c54a743 | Subtlety, Instrument of Law, Demonic Brand threat; healing / mana / Battle Shout threat split across enemies; Rockbiter OH threat | 70707 curves: 17118 -10/-20/-30 (3 ranks, Nature + Arcane), 1311085 e1 10/20 (all, RF off), 1293695 e0 -17/-33/-50 (op 2 threat, Searing Pain). Threat only; warrior DPS ~0.01 |
| 31054ead2 | Mage 70707 talents | Precision 29438 curve 2/4/6 aura 55; Accelerant 11069 -100..-500 ms, mask covers Frostfire Bolt (401502, 1237312) and Arcane Wrath (1326079); Critical Mass 3 ranks 2/4/6; Incineration 5 ranks 2..10. All match. Ice Lance shared 8 s CD is a user ruling |
| cf51105c3 | Raid sim table | raid sim only |

## Effect of f3017897f (seed 1, 5000 iterations, CLIs at bbbaf2e3c and f3017897f)

| Build | Before | After | Change |
|---|---|---|---|
| warrior__arms | 905.0 | 995.4 | +10.0% |
| warrior__fury_2h | 963.3 | 1026.8 | +6.6% |
| warrior__fury | 983.1 | 1044.4 | +6.2% |
| hunter__survival | 1028.8 | 1072.7 | +4.3% |
| paladin__level_60_shock_ret | 931.2 | 961.5 | +3.3% |
| rogue__mutilate | 723.8 | 745.0 | +2.9% |
| rogue__combat-swords | 765.2 | 778.5 | +1.7% |

Arms: white swings +13%, Heroic Strike casts 2.7x; most of it is the swing-timer ruling.

## Findings

- New: mechanics-20261030-2 (medium): the DPS report's `wf_totem_icd` and `extra_attack_swing` cards and standards 17 / 174 / 175 still describe 1.5 s and a reset, though the re-simmed report uses 100 ms and no reset.
- Open, unchanged: mechanics-20261030-1 (Tauren Endurance), mechanics-20261027-2 (Amplify Curse / Bane of Agony).

## Verdicts updated

Windfury Totem 100 ms and the extra-attack / Eviscerate rulings replace the 1.5 s settled line; new rows for weapon-fit imbues and the threat talents.
