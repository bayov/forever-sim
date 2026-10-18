#!/usr/bin/env python3
"""The items Forever still hides, and what each of our sources knows about them.

    python3 tools/sources/hidden.py [<build>]

It writes assets/database/hidden_items.json. Run it after dbcache.py and
gen_db (the update-sources skill's last step), and stale.py reads it.

Blizzard took the stats of thousands of Forever items out of the client's
files, so players find them in game rather than from dataminers. The client
still ships an Item row for every one (its class, subclass and slot), but no
ItemSparse row (its name, item level and stats). When the server first sends
an item to a player, it's revealed: the row arrives as a hotfix, and our
client keeps it in its cache (tools/sources/dbcache.py). So an item is hidden
when the client build has an Item row for it, no shipped ItemSparse row, and
no row in our client's cache. At 70450 that gives 7,841 hidden items, where
https://foreverchanges.pro/hidden-items counted 7,843 (12,618 less the 4,775
revealed), so the two agree. An item BayovCore saw revealed in game
(assets/db_inputs/forever_revealed_in_game.json, which bayovui writes) isn't
hidden either, even before our cache snapshot has its row. That file is
optional, and a missing one lists nothing.

We list hidden weapons and armor that matter to the sim:

- hidden, but in the sim with Classic's stats (Destiny, Bloodrazor). Forever
  may have changed them, and the sim can't know. `simStats` is "classic".
- Forever's own hidden items the sim has from a source that knows part of
  them, like Wowhead's armor without stats (the Manaflare set). `simStats` is
  "partial".
- Forever's own hidden items Wowhead lists with stats (Ouroboros Ring,
  Shapeshifter's Helm), so the sim has them all, but our client says the item
  is still hidden and Wowhead doesn't say where the numbers come from. They
  may be from before Blizzard hid them, or from Wowhead's in-game uploader.
  `simStats` is "wowhead". Wowhead's stats for hidden Classic items are
  Classic's, so those stay "classic".
- Forever's own hidden items the sim doesn't have at all.
- Hidden Classic items another source names, like an AtlasLoot drop.

and, apart from those, revealed gear the sim doesn't have yet (our client has
its row, but no source with stats has caught up): `revealed`.

Each item says which sources know it (`sources`): AtlasLoot (a drop), the
foreverchanges.pro loot tables, Wowhead's Forever listing, the sim. The
counts at the top are what stale.py compares, so a run that finds fewer
hidden items means something was revealed.

`reveals` is a separate list, for BayovCore's Reveals page: every weapon and
armor piece Forever hid and has revealed, one line each, the newest last. The
game can't tell an addon when an item was revealed, so the order has to come
from here. We'd like to order them by the hotfix push their row came in, but
a reveal doesn't come in a real push (see dbcache.py), so at 70450 not one
reveal has its own. Only 6 have a push at all, because the server sent their
rows again later (Aegis of Stormwind in push 112864), and that says when it
sent them again, not when they were revealed. So the order is the push when
there is one, then the date BayovCore saw the reveal (or else first saw the
item) for the rest, and the ones with neither go first, by ID. Until
BayovCore has dates for them, that's most of the list. Shirts, tabards, bags,
ammo and items without a slot are left out.
"""
import argparse
import collections
import json
import os
import sys

sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
import meta  # noqa: E402
import wago  # noqa: E402

ROOT = os.path.abspath(os.path.join(os.path.dirname(os.path.abspath(__file__)), '..', '..'))
INPUTS = os.path.join(ROOT, 'assets', 'db_inputs')
DATABASE = os.path.join(ROOT, 'assets', 'database')
OUT = os.path.join(DATABASE, 'hidden_items.json')

# source: Item.ClassID and ItemSparse.InventoryType numbers, client beta 1.60.1.70450
WEAPON, ARMOR = '2', '4'
# Forever's own items have IDs from here up. Below are Classic's and Season of
# Discovery's (200000 on), which the Forever client still carries.
FOREVER_FIRST_ID = 250000
SOD_FIRST_ID = 200000
# Wowhead's names for the primary stats in a listing's "eq".
PRIMARY = ('str', 'agi', 'sta', 'int', 'spi')
SLOTS = {1: 'head', 2: 'neck', 3: 'shoulder', 4: 'shirt', 5: 'chest', 6: 'waist', 7: 'legs',
         8: 'feet', 9: 'wrist', 10: 'hands', 11: 'finger', 12: 'trinket', 13: 'one-hand',
         14: 'shield', 15: 'ranged', 16: 'back', 17: 'two-hand', 18: 'bag', 19: 'tabard',
         20: 'chest', 21: 'main hand', 22: 'off hand', 23: 'held in off hand', 24: 'ammo',
         25: 'thrown', 26: 'ranged', 28: 'relic'}
# Gear that isn't worth a line in the reveals: shirts, bags, tabards and ammo.
NOT_REVEALS = (4, 18, 19, 24)


def load(name, folder=INPUTS):
    with open(os.path.join(folder, name), encoding='utf-8') as f:
        return json.load(f)


def in_game():
    """{item: entry} of the reveals BayovCore saw in game, {} until there's a file."""
    path = os.path.join(INPUTS, 'forever_revealed_in_game.json')
    if not os.path.exists(path):
        return {}
    return {int(k): v for k, v in load(path)['items'].items()}


def reveals(item, shipped, revealed, seen):
    """Every weapon and armor piece Forever hid and has revealed, the newest last.

    The order is the push, then the date BayovCore saw it, and the ones with
    neither go first, by ID, since we can't tell when they came (the module
    docstring says why so few have a push).
    """
    out = []
    for i in sorted(set(revealed) | set(seen)):
        row = item.get(str(i))
        if not row or str(i) in shipped or row['ClassID'] not in (WEAPON, ARMOR):
            continue
        r, g = revealed.get(i, {}), seen.get(i, {})
        if not (r.get('name') or g):
            continue
        slot = r.get('inventoryType') or int(row['InventoryType'])
        if not slot or slot in NOT_REVEALS:
            continue
        entry = collections.OrderedDict([
            ('id', i), ('name', r.get('name') or g.get('name')), ('slot', SLOTS.get(slot, slot)),
            ('inventoryType', slot), ('ilvl', r.get('ilvl')), ('quality', r.get('quality'))])
        if r.get('name'):
            # dbcache.py leaves out zeros, so a row with data and no required level has none.
            entry['requiredLevel'] = r.get('requiredLevel', 0)
        entry['push'] = r.get('push')
        entry['date'] = g.get('revealed') or g.get('seen')
        out.append(entry)

    def order(e):
        if e['push'] is not None:
            return (1, e['push'], '', e['id'])
        if e['date']:
            return (2, 0, e['date'], e['id'])
        return (0, 0, '', e['id'])
    return sorted(out, key=order)


def read(build):
    item = wago.by_id('Item', build)
    shipped = wago.by_id('ItemSparse', build)
    cache = load('forever_client_cache.json')
    revealed = {int(k): v for k, v in cache['items'].items() if v['state'] == 'valid'}
    seen = in_game()
    sim = {x['id']: x for x in load('db.json', DATABASE)['items']}
    sim.update({x['id']: x for x in load('leftover_db.json', DATABASE)['items'] if x['id'] not in sim})
    wowhead = {x['id']: x for x in load('forever_wowhead_items.json')['items']}
    atlasloot = {x['id']: x for x in load('atlasloot_db.json')['items'] if 'id' in x}
    npcs = {n['id']: n.get('name', '') for n in load('atlasloot_db.json').get('npcs', []) if 'id' in n}
    fc = {x['id']: x for x in load('foreverchanges_loot.json')['items']}

    hidden = {int(i) for i in item if i not in shipped and int(i) not in revealed and int(i) not in seen}
    out = {}
    for i in sorted(hidden):
        row = item[str(i)]
        if row['ClassID'] not in (WEAPON, ARMOR):
            continue
        sources = []
        if i in atlasloot:
            drops = [npcs.get(s['drop'].get('npcId')) or str(s['drop'].get('npcId'))
                     for s in atlasloot[i].get('sources', []) if 'drop' in s and s['drop'].get('npcId')]
            sources.append('AtlasLoot' + (': ' + ', '.join(dict.fromkeys(drops)) if drops else ''))
        if i in fc:
            sources.append('foreverchanges.pro: %s, %s' % (fc[i].get('dungeon'), fc[i].get('boss')))
        if i in wowhead:
            sources.append('Wowhead')
        forever = i >= FOREVER_FIRST_ID
        if not (i in sim or forever or sources) or (SOD_FIRST_ID <= i < FOREVER_FIRST_ID and i not in sim):
            continue
        entry = {'state': 'hidden', 'slot': SLOTS.get(int(row['InventoryType']), row['InventoryType'])}
        name = (sim.get(i) or wowhead.get(i) or fc.get(i) or {}).get('name')
        if name:
            entry['name'] = name
        if i in sim:
            if not forever:
                entry['simStats'] = 'classic'
            elif any(k in wowhead.get(i, {}).get('eq', {}) for k in PRIMARY):
                entry['simStats'] = 'wowhead'
            else:
                entry['simStats'] = 'partial'
        entry['sources'] = sources
        out[i] = entry
    for i, r in sorted(revealed.items()):
        if r.get('classID') in (2, 4) and i not in sim and r.get('quality', 0) >= 2:
            out[i] = {'state': 'revealed', 'slot': SLOTS.get(r.get('inventoryType'), r.get('inventoryType')),
                      'name': r.get('name'), 'ilvl': r.get('ilvl'), 'quality': r.get('quality'), 'sources': []}

    gear = [e for e in out.values() if e['state'] == 'hidden']
    shown = reveals(item, shipped, revealed, seen)
    counts = collections.OrderedDict([
        ('hidden', len(hidden)),
        ('hiddenGear', len(gear)),
        ('inSimWithClassicStats', sum(1 for e in gear if e.get('simStats') == 'classic')),
        ('inSimPartly', sum(1 for e in gear if e.get('simStats') == 'partial')),
        ('inSimFromWowhead', sum(1 for e in gear if e.get('simStats') == 'wowhead')),
        ('foreverNotInSim', sum(1 for i, e in out.items() if e['state'] == 'hidden'
                                and i >= FOREVER_FIRST_ID and 'simStats' not in e)),
        ('revealedNotInSim', sum(1 for e in out.values() if e['state'] == 'revealed')),
        ('revealedGear', len(shown)),
    ])
    return counts, out, shown, cache['meta'], bool(seen)


def main():
    ap = argparse.ArgumentParser(description=__doc__.split('\n\n')[0])
    ap.add_argument('build', nargs='?', help='the client build (default: the client cache\'s)')
    a = ap.parse_args()
    build = wago.version(a.build or load('forever_client_cache.json')['meta']['build'])
    counts, items, shown, cache_meta, seen = read(build)
    previous = None
    if os.path.exists(OUT):
        with open(OUT, encoding='utf-8') as f:
            previous = json.load(f)
    m = meta.meta('client beta %s Item and ItemSparse tables, and our client cache of %s%s' % (
        build, cache_meta.get('fetched'), ', and the reveals BayovCore saw in game' if seen else ''),
        'tools/sources/hidden.py', build=build)
    with open(OUT, 'w', encoding='utf-8') as f:
        f.write('{\n' + meta.line(m) + '"counts": %s,\n"items": {\n' % json.dumps(counts))
        f.write(',\n'.join('%s: %s' % (json.dumps(str(i)), json.dumps(e, sort_keys=True, ensure_ascii=False))
                           for i, e in sorted(items.items())))
        f.write('\n},\n"reveals": [\n')
        f.write(',\n'.join(json.dumps(e, ensure_ascii=False) for e in shown))
        f.write('\n]}\n')
    print(', '.join('%s %d' % kv for kv in counts.items()))
    if previous:
        gone = [i for i, e in previous['items'].items() if e['state'] == 'hidden'
                and items.get(int(i), {}).get('state') != 'hidden']
        if gone:
            print('revealed since the last run (%d): %s' % (len(gone), ', '.join(
                previous['items'][i].get('name') or i for i in gone[:20])))
    print('wrote %s' % os.path.relpath(OUT, ROOT))
    return 0


if __name__ == '__main__':
    raise SystemExit(main())
