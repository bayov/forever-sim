# Phase 0 gear, step 3 rerun (2026-11-23)

Every Phase 0 list was searched before these landed on master, so all 23 DPS specs run step 3 again:

- fdafddb1e: Dodge, Block, Parry and haste read from the game and Wowhead for 238 items.
- aff06f74e: Blackhand's Breadth crit is melee and ranged only, Eye of the Beast's is spells only.
- 57ea8cb6b: weapon skill from the game's readings (Distracting Dagger and Mugger's Belt daggers +1,
  Expert Goldminer's Helmet axes +1, nine added, three cleared), and 1% ranged hit on Marksman Bands,
  Sash of the Grand Hunt and Stoneshatter.
- 6e128175e and 9457811ae: the search drops the other faction's rewards by the build's race
  (`tools/rotopt/faction.go`), and unsure items were flipped to include per standards 42.

## Task

Follow `briefs/phase0_step3_1122.md` steps 1 to 5 for each spec in your group, on master 9457811ae or
later. Read your group's earlier notes, `phase0_step3_<group>.md`, first. The mage re-tune is done, so
run all four mage specs. The other brief's coordination section no longer applies.

Standards 42 holds: unsure availability is "include" with a note, never a call for the user. Name only
real problems in the hand-back: a pick that is wrong, a source that is known to be closed, a bug.

## When done

Update `phase0_step3_<group>.md` in place: replace the picks, flags, comparison and calibration with
the new ones, and add one short line per spec on what changed from the earlier list and why.

Push it (merge to master, never force-push, then your own branch). Send one message to the parent
session with the master commit and, in under 250 words, each spec's new Phase 0 DPS against synthetic
and against its earlier Phase 0 list, and the picks that changed. Then stop.
