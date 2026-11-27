# Phase 0 cleanup (2026-11-27)

The DPS, tank and healer reports are all on Phase 0 gear and deployed (master 2a919a193). These are the
small items the last three rounds left open. Notes: `phase0_prices_1125.md`, `phase0_rerank_1125.md` and
the `phase0_tankheal*.md` files in `orchestration/round_notes_2026-11-11/`.

## Fix

1. **Failing tests on master.** `TestWarriorWfGate` (tools/rotopt) and `TestShamanEnhStormstrikeWindfuryHold`.
   Find out why each fails since the sim batches, and fix the code or the test's stale expectation. Never
   skip or loosen a test to get green.
2. **Balance template trinket hold.** The druid_balance rotopt template still writes a trinket hold. Make
   it stop, and check the Balance APL.
3. **Rogue effects files.** These weren't refreshed after the second re-rank round. Refresh them and re-sim
   any rogue row that moves.
4. **Ret Shock Talisman.** Talisman of Ascendance gives +6.3 in a sim but -0.4 on the list score. Find which
   is wrong and fix the list or the sim.
5. **Blackblade of Shahram price.** It was set by hand (424 g/DPS) because phase0_prices.py gives none, so
   a `--write` run may drop it. Make the script price it, or keep a hand price the script respects, with a
   test.
6. **Holy and disc priests go OOM on Chromaggus.** Find out whether that's the gear, the rotation or the
   bench, and fix what's wrong. If it's real at Phase 0 gear, say so in the notes.
7. **Holy lists sit about 1.4% over bench.** Find out why the list score and the bench differ, and fix.

Each fix gets a test where it can have one. Leave open, and only report: Shield Slam 2x block value (a user
ruling) and the unsure or unpriced items (standards 42 keeps them).

## Publish

`stale.mjs`, the pipelines for any rows that moved, `build_data.py`, `changes.py` and one Changes tab entry
if numbers moved. Run the auditors for what changed, merge to master (fetch and merge first, never
force-push), `make dps-report-deploy`, check the live page.

## Hand back

One message to the parent, under 250 words: master commit, deploy check, each item's answer, what moved.
Then stop.
