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

    print('AtlasLoot %d items, Wowhead %d, both %d' % (len(al_items), len(wh), len(set(al_items) & set(wh))))
    for title, rows in (('conflict', conflict), ('AtlasLoot only', al_only), ('Wowhead only', wh_only)):
        print('\n%s: %d' % (title, len(rows)))
        for iid, name, why in rows if a.all else rows[:10]:
            print('  %d %s (%s)' % (iid, name, why))
    return 0


def names(ids, npc_names):
    return ', '.join('%s %d' % (npc_names[i], i) if npc_names.get(i) else str(i) for i in sorted(ids))


if __name__ == '__main__':
    raise SystemExit(main())
