#!/usr/bin/env python3
"""The Forever client's own item facts, for gen_db and rotopt.

    python3 tools/sources/itemsparse.py <build>

It writes assets/db_inputs/forever_client_items.csv from the client build's
ItemSparse and ItemSet tables on wago.tools, and forever_client_items.meta.json
next to it with the build (a CSV has no room for a meta block, so it goes
beside it, the way atlasloot_db.json does it). tools/sources/stale.py compares
that build with the client's.

Why it doesn't simply replace wago_db2_items.csv, Classic 1.15.3.55646's
ItemSparse: the Forever client's ItemSparse on wago.tools has only about
19,000 rows, and 6,662 of the 11,356 items in db.json aren't among them
(Destiny, Bloodrazor and most other Classic items). So gen_db reads the
Classic file first and lays this one over it. An item in this file takes its
faction and set from here, and every other item keeps Classic's.

The columns:

    ID, Name, ItemLevel, Flags_1, RequiredPVPRank, ItemSet, DamageType

Name, ItemLevel, Flags_1, RequiredPVPRank and DamageType are blank for an item
that is only in the ItemSet table. DamageType is a weapon's damage school in
the game files' order (0 physical, 1 holy, 2 fire, 3 nature, 4 frost, 5 shadow,
6 arcane), which gen_db gives wands. ItemSet comes from the ItemSet table's member
lists first, because that table also lists Classic items ItemSparse doesn't
have (Defias Leather's five pieces), and from ItemSparse's own column after
that. Forever's ItemSet table and ItemSparse agree on every item both have.
"""
import csv
import json
import os
import sys

sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
import meta  # noqa: E402
import wago  # noqa: E402

ROOT = os.path.abspath(os.path.join(os.path.dirname(os.path.abspath(__file__)), '..', '..'))
OUT = os.path.join(ROOT, 'assets', 'db_inputs', 'forever_client_items.csv')
META = os.path.join(ROOT, 'assets', 'db_inputs', 'forever_client_items.meta.json')
COLUMNS = ['ID', 'Name', 'ItemLevel', 'Flags_1', 'RequiredPVPRank', 'ItemSet', 'DamageType']


def read(build):
    items = {}
    for r in wago.rows('ItemSparse', build):
        items[int(r['ID'])] = {'ID': r['ID'], 'Name': r['Display_lang'], 'ItemLevel': r['ItemLevel'],
                               'Flags_1': r['Flags_1'], 'RequiredPVPRank': r['RequiredPVPRank'],
                               'ItemSet': r['ItemSet'], 'DamageType': r['DamageType']}
    for s in wago.rows('ItemSet', build):
        for k, v in s.items():
            if k.startswith('ItemID_') and v not in ('', '0'):
                row = items.setdefault(int(v), {'ID': v})
                row['ItemSet'] = s['ID']
    return [items[i] for i in sorted(items)]


def main(argv):
    if len(argv) != 2:
        print(__doc__.split('\n\n')[1].strip(), file=sys.stderr)
        return 2
    build = wago.version(argv[1])
    rows = read(build)
    with open(OUT, 'w', encoding='utf-8', newline='') as f:
        w = csv.DictWriter(f, COLUMNS, lineterminator='\n')
        w.writeheader()
        w.writerows(rows)
    m = meta.meta('client beta %s (wago.tools ItemSparse and ItemSet)' % build, 'tools/sources/itemsparse.py',
                  build=build)
    with open(META, 'w', encoding='utf-8') as f:
        json.dump({'meta': m}, f, indent=2, sort_keys=True)
        f.write('\n')
    print('%d items from build %s into %s' % (len(rows), build, os.path.relpath(OUT, ROOT)))
    return 0


if __name__ == '__main__':
    raise SystemExit(main(sys.argv))
