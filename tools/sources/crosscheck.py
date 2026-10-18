#!/usr/bin/env python3
"""Check AtlasLoot's item sources against Wowhead's, and list where they disagree.

    python3 tools/sources/crosscheck.py [--all]

We have two sources for where an item drops: the AtlasLoot Forever addon
(assets/db_inputs/atlasloot_db.json, from `gen_db -gen=atlasloot`) and
Wowhead's Forever gear listing (assets/db_inputs/forever_wowhead_items.json,
the scrape-wowhead skill). Neither is complete, and each has been wrong
before. So the update-sources skill runs this after refreshing either one,
and reads what it prints:

- conflict: both name the NPCs an item drops from, and no NPC is in both.
- AtlasLoot only: AtlasLoot has a drop and Wowhead knows no drop for the
  item (it may say quest or vendor, or nothing).
- Wowhead only: Wowhead names a boss AtlasLoot has loot tables for, and
  AtlasLoot doesn't list the item under any boss. That is usually loot
  AtlasLoot hasn't added yet, like a new Forever dungeon's.

It only compares items both files cover, and Wowhead's listing is gear only.

A last section compares our client's hotfix cache (forever_client_cache.json,
tools/sources/dbcache.py) with the sim's database: Classic items whose
primary stats differ. The cache stores stats as shares of a budget that
RandPropPoints gives, which holds for Classic's items only (see dbcache.py),
so Forever's own items are left out. The sim leaves out negative stats, so
those don't count either. A difference is often a cache row from an earlier
push, so it's for a look, not a fix.
Wowhead often names a zone but no NPC (an item with several bosses), and
that isn't a conflict. It prints the first 10 of each with their names, or
every one with --all.
"""
import argparse
import json
import os

ROOT = os.path.abspath(os.path.join(os.path.dirname(os.path.abspath(__file__)), '..', '..'))
ATLASLOOT = os.path.join(ROOT, 'assets', 'db_inputs', 'atlasloot_db.json')
WOWHEAD = os.path.join(ROOT, 'assets', 'db_inputs', 'forever_wowhead_items.json')

# Wowhead's source codes, and its sourcemore type for an NPC.
DROP = 2
NPC = 1


def wowhead_drops(item):
    """(whether Wowhead says it drops, the NPC IDs it names for the drop)."""
    src = item.get('source') or []
    more = item.get('sourcemore') or []
    npcs = {m['ti'] for k, m in zip(src, more) if k == DROP and m.get('t') == NPC and 'ti' in m}
    return DROP in src, npcs


def main():
    ap = argparse.ArgumentParser(description=__doc__.split('\n\n')[0])
    ap.add_argument('--all', action='store_true', help='list every item, not the first 10')
    a = ap.parse_args()

    with open(ATLASLOOT, encoding='utf-8') as f:
        al = json.load(f)
    with open(WOWHEAD, encoding='utf-8') as f:
        wh = {i['id']: i for i in json.load(f)['items']}
    npc_names = {n['id']: n.get('name', '') for n in al.get('npcs', []) if 'id' in n}
    al_items = {i['id']: i for i in al['items'] if 'id' in i}

    def al_npcs(item):
        return {s['drop']['npcId'] for s in item.get('sources', []) if 'drop' in s and s['drop'].get('npcId')}

    def al_drops(item):
        return any('drop' in s for s in item.get('sources', []))

    conflict, al_only, wh_only = [], [], []
    for iid, w in sorted(wh.items()):
        drops, wn = wowhead_drops(w)
        item = al_items.get(iid)
        if item is None:
            # Only a boss AtlasLoot knows, so world drops don't count.
            if wn & set(npc_names):
                wh_only.append((iid, w['name'], 'Wowhead: %s' % names(wn, npc_names)))
            continue
        an = al_npcs(item)
        if an and wn and not an & wn:
            conflict.append((iid, w['name'], 'AtlasLoot: %s, Wowhead: %s' % (
                names(an, npc_names), names(wn, npc_names))))
        elif al_drops(item) and not drops:
            al_only.append((iid, w['name'], 'AtlasLoot: %s, Wowhead sources %s' % (
                names(an, npc_names) or 'a drop', w.get('source') or 'none')))
        elif wn & set(npc_names) and not al_drops(item):
            wh_only.append((iid, w['name'], 'Wowhead: %s' % names(wn, npc_names)))

    cache = cache_differences()
    print('AtlasLoot %d items, Wowhead %d, both %d' % (len(al_items), len(wh), len(set(al_items) & set(wh))))
    for title, rows in (('conflict', conflict), ('AtlasLoot only', al_only), ('Wowhead only', wh_only),
                        ('client cache vs the sim', cache)):
        print('\n%s: %d' % (title, len(rows)))
        for iid, name, why in rows if a.all else rows[:10]:
            print('  %d %s (%s)' % (iid, name, why))
    return 0


# Item stat numbers of the primary stats, to their index in the sim's stats.
PRIMARY = {4: 0, 3: 1, 7: 2, 5: 3, 6: 4}
STAT_NAMES = ['Strength', 'Agility', 'Stamina', 'Intellect', 'Spirit']
# RandPropPoints' column for each quality, and its slot index for each inventory type.
QUALITY = {2: 'Good', 3: 'Superior', 4: 'Epic'}
BUDGET_SLOT = {1: 0, 4: 0, 5: 0, 7: 0, 17: 0, 20: 0, 3: 1, 6: 1, 8: 1, 10: 1, 12: 1, 2: 2, 9: 2, 11: 2,
               14: 2, 16: 2, 23: 2, 13: 3, 21: 3, 22: 3, 15: 4, 25: 4, 26: 4, 28: 4}
SOD_FIRST_ID = 200000


def cache_differences():
    """(id, name, why) for the Classic items whose cached primary stats differ from the sim's."""
    import sys
    sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
    import wago
    path = os.path.join(ROOT, 'assets', 'db_inputs', 'forever_client_cache.json')
    if not os.path.exists(path):
        return []
    with open(path, encoding='utf-8') as f:
        cache = json.load(f)
    with open(os.path.join(ROOT, 'assets', 'database', 'db.json'), encoding='utf-8') as f:
        sim = {i['id']: i for i in json.load(f)['items']}
    points = wago.by_id('RandPropPoints', cache['meta']['build']) or {}
    out = []
    for key, e in sorted(cache['items'].items(), key=lambda kv: int(kv[0])):
        iid = int(key)
        if iid >= SOD_FIRST_ID or iid not in sim or 'statShares' not in e:
            continue
        shares = [(t, v) for t, v in e['statShares'] if t in PRIMARY and v > 0]
        column = QUALITY.get(e.get('quality'))
        slot = BUDGET_SLOT.get(e.get('inventoryType'))
        row = points.get(str(e.get('ilvl')))
        if not shares or column is None or slot is None or row is None:
            continue
        budget = float(row['%s_%d' % (column, slot)])
        cached = {PRIMARY[t]: int(budget * v / 10000 + 0.5) for t, v in shares}
        theirs = {k: sim[iid]['stats'][k] for k in PRIMARY.values() if sim[iid]['stats'][k] > 0}
        cached = {k: v for k, v in cached.items() if v}
        if cached != theirs:
            fmt = lambda d: ' '.join('%s %g' % (STAT_NAMES[k], v) for k, v in sorted(d.items())) or 'none'
            out.append((iid, sim[iid]['name'], 'client cache: %s, sim: %s' % (fmt(cached), fmt(theirs))))
    return out


def names(ids, npc_names):
    return ', '.join('%s %d' % (npc_names[i], i) if npc_names.get(i) else str(i) for i in sorted(ids))


if __name__ == '__main__':
    raise SystemExit(main())
