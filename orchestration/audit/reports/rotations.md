# Rotations pass 67 (2026-11-15, master 25ce81de7)

Last audited master: 70ea588bb. Client 1.60.1.70938 data plus the November 13 hotfixes (PR #129, #130). Result: quiet
(no new finding; 3 open findings closed).

## What changed (70ea588bb..25ce81de7, 61 commits)

| path | change | reaches presets? |
|---|---|---|
| `sim/core/buffs.go`, `flags.go`, `windfury_totem_procs_forever.go` (deleted) | Windfury Totem rolls on every landed main-hand melee hit, white or ability, 100 ms ICD | yes: every melee build with the totem |
| `sim/shaman/windfury_weapon.go` | imbue ICD exposed on the "Windfury Imbue" aura (APL can read it) | Enhancement (new Stormstrike gate) |
| `sim/paladin/*` | Judgement of Command on melee defense; Two-Handed Weapon Specialization's Holy part follows defense type | Ret builds |
| `sim/hunter/item_sets_pve.go` | Wildstalker 5pc also shortens Multi-Shot's cooldown | MM, LW (wear 5pc) |
| `sim/warlock/spell_data.go` | Imp Firebolt ranks 2-4 | no (level 60 uses rank 7+) |
| `sim/common/item_effects.go`, `sim/core/item_effects.go` | Silithid Queen's Husk, Dawn items vs Undead | gear only |
| APLs (`apl_logic_diff.py`: 17 files) | Arms WF swing gate, Enh Stormstrike holds for the imbue ICD, Feral Bite 65 + totem gate, Ret twist/Shock re-tune, tank warrior re-tune | yes |

## Baseline (3000 it, seed 424242, all 68 rows re-simmed)

New `ref/rotations/baseline_25ce81d_3000it_seed424242.tsv`. Moves of 0.5 DPS or more vs 70ea588:

| builds | move at the baseline |
|---|---|
| warrior Fury 2H (3) | +70.6 to +73.0 (~+7%) |
| warrior Fury DW (3) | +35.2 to +45.2 |
| paladin Ret (10) | +23.9 to +42.8 |
| druid Feral (2) | +29.7, +31.9 |
| warrior Arms (5) | +25.2 to +28.9 |
| shaman Enhancement (4) | +26.1 to +28.8 |
| hunter Survival (3) | +19.2 to +21.6 |
| rogue Subtlety | +2.3 |

Everything else (casters, MM/BM/LW hunters, the other rogues) moved under 0.5. All moves trace to the round above.

## Open findings checked

| finding | status on 25ce81de7 |
|---|---|
| rotations-20261115-1 Arms WF gate | landed as the swing variant (a8ea43990) on Arms and its duties: resolved. Dragonkin left out, and the gate now loses there (below) |
| rotations-20261105-1 Feral Bite 65 | landed (f9990ec4e): resolved |
| rotations-20261109-1 Fury 2H drop 90-rage HS | now loses 3.0 at 180 s: withdrawn, moved to Resolved |
| rotations-20261112-1 rogue poison swap | still unapplied (reqs keep MH Deadly / OH Instant). Rogue rows barely moved, so not re-measured |

## Windfury timing under the Nov 13 rule (paired ab.py, 20k it, CLI 25ce81de7755, set 1, S0 1e6)

Variants made with `tools/audit/rotations/wf_gate.py` (rewritten this pass; last pass's copy wasn't in the repo). Gate =
the landed Arms form, `autoTimeToNext(MH) <= autoSwingTime(MH) - 0.1s`, on the build's strike lines.

| preset / variant | B-A 120 / 180 / 300 s | A at 180 |
|---|---|---|
| arms_dragonkin, gate on 6 strike lines | -0.15 / -0.38 / -0.84 (SE 0.3-0.5) | 1126.1 |
| fury_2h, gate on 4 lines | +1.42 / -0.52 / +0.96 (SE 0.5-0.8) | 1104.5 |
| fury_2h, delete 90-rage Heroic Strike | -2.23 / -2.99 / -4.09 (SE 0.3-0.6) | 1104.5 |
| fury (DW), gate on Bloodthirst/Whirlwind/etc. | -0.97 / -0.36 / -0.55 (SE 0.4-0.5) | 1092.3 |

None passes the bar, so the second seed set was not run. The workers' round covered the rest of the user's request:
Enhancement holds Stormstrike for Windfury Weapon's ICD (4d0c2ceac, now readable via the aura), Feral and Ret re-tuned,
rogue timing knobs searched and off (057127ac3). The open item "Main-hand ability timing for Windfury" in
`passes/rotations.md` is done and can be removed (this pass may not edit that doc).

## Hunter: Multi-Shot first with Wildstalker 5pc

The 5pc now shortens Multi-Shot's cooldown too. MM with Multi-Shot moved above Aimed Shot: -62.7 / -57.4 / -31.7
(set 1, SE 0.4, A 1234.4). The settled ruling holds; MM/LW baselines didn't move, as Multi-Shot still never fires.

## Optional (under the bar)

- Fury 2H WF swing gate: noisy, +1.4 at 120 but -0.5 at 180 (set 1). Not worth a second set.

## Not done / left open

- Rogue poison swap not re-measured on this master (baseline unchanged for those builds).
- Balance NE Elune's Light, Elemental Mastery hold, warlock duty-curse screens: still open items, unchanged.
