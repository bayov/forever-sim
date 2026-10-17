# Consumables audit: local reference data (not committed)

The audit compares against Forever client db2 tables from wago.tools. They're 15 to 18 MB a build, so they stay out of
the repo, in the shared cache `tools/sources/wago.py` keeps. Re-create them with:

```sh
python3 tools/sources/wago.py fetch 1.60.1.70338 --set consumables   # the build most of the presets were checked against (2026-10-10)
python3 tools/sources/wago.py fetch 1.60.1.70425 --set consumables   # 2026-10-15/16
python3 tools/sources/wago.py fetch 1.60.1.70450 --set consumables   # newest as of 2026-10-17: the current baseline
```

Tables fetched (wago.py's `consumables` set): Item, ItemSparse, ItemEffect, ItemXItemEffect, SpellEffect, SpellName, Spell, SpellMisc, SpellLevels,
SpellItemEnchantment, SpellCategories, SpellCategory, SpellDuration, SpellAuraOptions.
Item.csv is only needed for ClassID (consumable = 0).

Wowhead Forever pages (item sources) are cached by src.py / srcid.py in `$AUDIT_WORK/wh/<item id>.html`. They're
fetched on demand. Wowhead has returned 403 to the container at times; then say the source couldn't be checked.

Files in this directory (small, committed):

- `scan_<master sha>.txt`: scan.sh output at the last audited master. Clean = nothing between the `==` headers
  above `== counts`.
- `consumes_baseline_<master sha>.txt`: every `*Consumes*` preset of the eleven Level 60 DPS specs at that master,
  one line each. Diff a new master's output against it to see consumable changes at a glance.
- `builddiff_70338_to_70425.txt`, `builddiff_70425_to_70450.txt`: the last two client diffs. 70338 to 70425 changed
  only Lily Root, Flash Bundle, campfire kits, armor re-encodes, and Enchant Bracer - Lesser Intellect +5 to +3
  (fixed in forever_enchants.go). 70425 to 70450 changed nothing this audit owns. Run again since 2026-10-17, it also lists two new ensembles, because builddiff.py used to take Item.csv from the old build.

Re-create the reference files with:

```sh
export AUDIT_TREE=/path/to/forever-sim AUDIT_WORK=/path/to/scratch
M=$(git -C $AUDIT_TREE rev-parse --short origin/master)
tools/audit/consumables/scan.sh > orchestration/audit/ref/consumables/scan_$M.txt
git -C $AUDIT_TREE show origin/master:tools/database/forever_enchants.go > $AUDIT_WORK/fe.go
python3 -I tools/audit/consumables/builddiff.py 70425 70450 $AUDIT_WORK/fe.go
```
