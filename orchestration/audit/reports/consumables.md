# Consumables report, pass 46 (master d3df9247f)

Result: quiet, no findings.

## Moved since 99bd53ada
- changes.sh lists no moved watched files on master. Class and consumes branches show the usual stale diverged diffs (gotcha 1); nothing new to audit.

## New client build 1.60.1.70532 -> 1.60.1.70630
- builddiff.py: consumable spell changes 0; Forever enchants checked 53, changed 0.

## Verified
- scan.sh on master: banned items, Revelation, non-synthetic trinkets, double crit elixirs all empty.
- Counts unchanged: raid totems (Grace of Air + Windfury for melee, Flametongue for enhancement). Minor-armor count 3 is the known unused ForeverDustCloudAura (scan's expected 2 is stale).
- wago latest probe: 1.60.1.70630.

No open items.
