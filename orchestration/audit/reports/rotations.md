# Rotations pass 33 (2026-10-26)

- **Master:** f62cb7971 (last audited f796d9953). Client 1.60.1.70532 (unchanged).
- **Result:** quiet, 0 new findings. No open findings carried.

## What changed

- No Level 60 APL or req file changed (`apl_logic_diff.py f796d9953 f62cb7971`: 0 apl files).
- Sim: 32a6848f9, the wowsims/forever assumptions of standard 28: Multi-Shot 0.5 s cast hasted by
  ranged speed and keeping the Auto Shot timer, live DoT/channel ticks, hunter base RAP 110,
  failed Ferocious Bite keeps energy and combo points, regen held at the casting rate through a
  mana-paid channel. The rest of the range is Replay/DPS Report UI.

## Regression baseline (3000 it, seed 424242, shared CLI at f62cb7971)

New file `ref/rotations/baseline_f62cb79_3000it_seed424242.tsv` (replaces `baseline_74b3a5f`).
Rows that moved more than 0.5 DPS, all explained by the mechanics above:

| preset | old | new | Δ |
|---|---:|---:|---:|
| warlock affliction | 1034.9 | 1029.0 | −6.0 |
| hunter lone-wolf | 1126.9 | 1121.6 | −5.2 |
| priest shadow | 961.8 | 956.9 | −5.0 |
| priest shadow-holy | 919.1 | 914.4 | −4.7 |
| warlock demonology | 1064.8 | 1060.1 | −4.7 |
| druid balance | 856.7 | 852.5 | −4.2 |
| balance wrath duty-ff | 784.8 | 780.9 | −3.8 |
| affliction duty-curse | 1024.4 | 1020.7 | −3.7 |
| balance duty-ff | 825.1 | 821.5 | −3.6 |
| demonology duty-curse | 1051.6 | 1048.2 | −3.4 |
| hunter marksmanship | 1138.7 | 1135.7 | −3.0 |
| priest smite | 766.5 | 763.9 | −2.6 |
| balance wrath | 813.1 | 810.7 | −2.4 |
| destruction duty-curse | 999.4 | 1002.3 | +2.9 |
| hunter beast-mastery | 1141.2 | 1143.5 | +2.3 |
| destruction fire duty-curse | 976.5 | 978.6 | +2.1 |

Matches the loop's `vs_published.txt` (23 of 57 moved > 0.5).

## Rotation checks against the new mechanics

Paired A/B, 20k it in 8 blocks, B−A ± SE (`ab.py`, shared CLI at f62cb7971).

**Multi-Shot cast (MM, LW).** It fires about 24 times a fight in MM and LW (not "never", as the
settled list in `passes/rotations.md` still says; that line is stale, BM fires it ~0 times).

| variant | preset | s0 | 120 s | 180 s | 300 s |
|---|---|---|---:|---:|---:|
| drop Multi-Shot | MM | 1e6 / 2e8 | −130.7 / −130.9 | −123.4 / −123.0 | −96.4 / −95.9 |
| drop Multi-Shot | LW | 1e6 / 2e8 | −144.9 / −145.1 | −139.0 / −139.1 | −111.7 / −111.8 |
| only if `autoTimeToNext(Ranged) ≥ castTime` | MM | 1e6 / 2e8 | −2.2 / −2.7 | −2.6 / −2.2 | −0.3 / −0.1 |
| only if `autoTimeToNext(Ranged) ≥ castTime` | LW | 1e6 / 2e8 | −1.9 / −1.7 | −2.2 / −2.1 | −0.5 / −0.4 |

The current Multi-Shot-on-cooldown line holds. Avoiding the auto clip loses.

**Channel five-second rule (Shadow).** Wand window 10→20% varied to 10→30%, 15→30%, 5→15% and
removed: all exactly +0.00 at 120/180/300 s (s0 1e6). Shadow never reaches 10% mana in these
fights, so the line stays inert and the channel rule costs no rotation change.

**Live DoT ticks.** No APL line times a spell DoT to a buff or trinket (notes grep for snapshot,
refresh-on-buff; only the Subtlety Rupture note, a bleed, which already ticked live). Nothing to test.

**Bite miss refund.** Feral −1.3 at 3000 it; not screened (mechanic is small and the Bite gates are
energy thresholds the refund barely touches). Open item for a later pass if Feral moves again.

## Optional

None new. The pass-doc Optional list carries over.

## Note for the pass doc

`passes/rotations.md` "Settled: Multi-Shot never fires in BM, MM or LW" is out of date for MM and LW
(it fires ~24 times a fight and is worth 96–145 DPS). Suggest the orchestrator reword it.
