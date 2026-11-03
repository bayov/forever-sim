# Rotations pass 49 (no findings)

- Master: f5c99ad01 (from 9da90fcde), audited 2026-11-03. Client 1.60.1.70753 plus the 2026-11-02 development notes
  (round 70753, phase 3 audit; standards 37: notes-only values are modelled, not findings by themselves).
- Not a quiet pass: 28 Level 60 APL files changed in logic and 82 non-test sim `.go` files changed.
- New baseline: `ref/rotations/baseline_f5c99ad_3000it_seed424242.tsv` (replaces `baseline_a0b3069_...`).

## What changed

Every class was reworked this round by its class worker, with talent and knob searches at 120/180/300 s on two seeds
(`orchestration/round_70753/<class>.md`). APL logic changes (`apl_logic_diff.py 9da90fcde f5c99ad01`):

| Class | APL change |
|---|---|
| Warrior (10 APLs) | Reckless Swings (1399001, notes-only) on cooldown everywhere; Arms Slam ahead of Mortal Strike; Fury Sunder duty HS from 70 and Bloodthirst first; Fury 2H Sunder duty filler gap |
| Rogue (6) | Shank (1399248, notes-only) at 75 Energy or less; Subtlety finisher rework; Shadowmeld (462707) dropped from the Ambush gates; Combat Daggers Adrenaline Rush at <49 |
| Shaman | Lava Burst lines removed (spell removed); Elemental mana-saver bolt rank 8 -> rank 9; Magma totems 0.5 s earlier |
| Hunter | Aimed Shot mana gate dropped (MM, LW); BM weave slack; MM Serpent Sting line moved |
| Mage | Arcane: two Arcane Blast (1239700) lines added, filler Frostbolt -> Fireball (and the end-of-fight Fire Blast gate on its cast time); Fire: Combustion condition |
| Warlock | New preset Affliction Wrack (`level60_affliction_wrack`) |
| Paladin | Undead Ret: two lines reordered |

Baseline vs a0b3069 (3,000 it, seed 424242): warrior +76..+127, rogue +122..+209, shaman +201..+275, hunter +16..+224,
druid +8..+191, paladin +48..+69, warlock -1..+66, mage 0..+25, priest 0..+1. These are the round's intended changes;
`regress.py` was not run (old CLI vs new CLI on these req files mixes the client data, the notes values and the new
talent trees, so it can't isolate anything the round files don't already state).

## Checks

1. **Dead lines** (new helper `tools/audit/rotations/dead_lines.py` on the shared 20k run): every APL castSpell id with
   under 0.05 casts an iteration. All are inert by design: race lines (Blood Fury, Berserking, Shadowmeld, Elune's Light,
   Eureka, Dark Sacrifice, Exorcism on a humanoid), talent-less lines (Summon Hawk, Sniper Shot in BM, Death Wish in
   Arms, Arcane Blast/Missiles in Frost), AQ ranks (Backstab 25300, Eviscerate 31016), end-of-fight lines, the mage
   "never" gem lines, Multi-Shot (settled). Two read as zero only because of how the metrics count them: Immolation
   Trap (14305, 0 casts but 14.1k damage a fight) and Judgement (20271; the seal's judgement ids carry the casts).
2. **Talent strings** (new helper `tools/audit/rotations/talent_check.py`): every rogue, warrior and shaman req string
   is legal in `ui/core/talents/trees/<class>.json` (ranks, 5 points a row, prerequisites maxed, 51 points), and those
   trees match `tools/forever_talents/notes/<class>.json` (max ranks, 1-based rows and columns, prerequisites,
   removals, the three renames). No mismatch.
3. **Duties** (target and self aura uptime, shared run, 180 s): Sunder 179.7-180.0, Battle Shout 180.0, Curse of
   Recklessness 179.6-179.7, Faerie Fire 179.2-180.0, Judgement of the Crusader / Wisdom 180.0. None lapses.
4. **Resources**: Energy overcap 0.0-0.6% on every rogue and cat; Rage overcap 0.5-1.3% Fury, 3.4-5.7% Fury 2H and
   4.2-4.7% Arms (mostly white-hit rage at the cap, about 10% of Reckless Swings' rage). OOM seconds: MM/LW hunters
   2.6-3.3 s, Enhancement Imp GW 2.3 s, Fire 1.4 s, others under 1 s. Probed below; nothing pays.
5. **Spot-check of a worker change**: Fury Sunder duty HS 70 (was 80) holds: reverting to 80 is -1.75 / -1.70 at
   180 s on the two seed sets (+0.5 / +0.6 at 120, -3.5 / -2.8 at 300).
6. Open findings before this pass: none.

### Probes (paired 20k, 8 blocks, B-A ± SE at 120 / 180 / 300 s, s0 1e6; baseline A at 180 s)

| Preset | Variant | B-A | Verdict |
|---|---|---|---|
| Arms (1097.8) | Heroic Strike at 75 Rage (was 90) | -2.47 / -3.21 / -1.37 | loses |
| Arms (1097.8) | Reckless Swings only at 70 Rage or less | +0.67 / -0.65 / +0.51 | noise |
| Fury 2H (1169.5) | same | -0.19 / -0.47 / -0.72 | loses |
| Fury (1152.1) | same | +0.26 / -0.33 / -0.13 | noise |
| Fire (1057.9) | Evocation with 30 s left (was 90 s; never fires now) | -0.05 / -9.71 / +1.81 | fight-length trade-off, loses at 180 |
| Fire (1057.9) | Evocation under 10% mana with 20 s left | -0.02 / -4.48 / +3.23 | same |

## Optional (under the bar)

None new. The pass-26 optional list stands.

## Notes (not findings)

- The mages and priests "cast" Shoot (5019) 2.7-25 times a fight with no wand damage: the wand line is reached right
  after an instant while the GCD runs and is a zero-time no-op (debug log, Fire: Shoot at 7.50 after Fire Blast, next
  cast at 9.00). No time is lost.
- Affliction Wrack ends with about 2,000 mana where the other warlocks end near 650; its Life Tap line almost never
  fires (0.01 an iteration). Mana isn't its limit, so nothing to convert. Not probed further.

## Left for a later pass

- A full threshold re-scan per class was not repeated: the class workers ran talent and knob searches at three fight
  lengths and two seeds this round. A later pass with time could screen the new structural lines (Shank's place in the
  priority, Arcane Blast lines, the Subtlety finisher rework) with ordering variants the knob searches don't reach.
