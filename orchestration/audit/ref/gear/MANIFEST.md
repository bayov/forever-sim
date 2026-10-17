# Gear audit reference files

Baseline: origin/master d2633f4a4 (2026-10-17), game client 1.60.1.70425. Gear on it was last audited clean in pass 36. Its preset gear is identical to 6c64ef928, which pass 32 audited in full.

| File | Made by |
|---|---|
| lin2_d2633f4a4.txt | `tools/audit/gear/lin2.py` on `extract.sh d2633f4a4` |
| std_d2633f4a4.txt | `tools/audit/gear/std.py` on the same extract (after `gen_db.py` at d2633f4a4) |
| weapons_d2633f4a4.txt | `tools/audit/gear/wepcheck.py` on the same extract, without an old dir |
| trinkets_d2633f4a4.txt | `tools/audit/gear/trinkets.py d2633f4a4` |

Paths in the outputs are relative to `$AUDIT_WORK`. Diff them after replacing your extract dir name with `sX`. Example: `diff <(sed 's/s37/sX/' ref) <(sed 's/sNN/sX/' new)`.

## Kept locally only (not committed), and how to re-create

- **`$AUDIT_WORK/enchants.json` and `$AUDIT_WORK/items.json`:** lookup tables for std.py.
  ```
  AUDIT_TREE=. AUDIT_WORK=$AUDIT_WORK AUDIT_GITREF=origin/master python3 tools/audit/gear/gen_db.py
  ```
- **`$AUDIT_WORK/tt/<id>.json`:** the wowhead Era tooltip cache, filled on demand by `tools/audit/gear/tt.py ID`. It reads `https://nether.wowhead.com/classic/tooltip/item/ID`.
- **`wago/*.csv`:** client db2 tables at build 1.60.1.70338: ItemSet, ItemSetSpell, SpellEffect, Spell and SpellName. They were used once to check the 2-piece bonuses against real Forever item sets.
  - Every synthetic set bonus was settled long ago, so you only need them if a new set bonus appears.
  - Re-create one with `curl -sSL "https://wago.tools/db2/<Table>/csv?build=1.60.1.70425" -o wago/<Table>.csv`. Use the current client build; 70425 has been live since 2026-10-15.
- **Old per-pass extracts (`s2`..`s37`):** throwaway. Re-extract any sha with `extract.sh <sha> <dir>`.

No built binaries are needed: the gear audit is static (git show plus Python 3). It never runs the sim.
