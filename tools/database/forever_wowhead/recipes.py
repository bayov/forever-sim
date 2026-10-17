#!/usr/bin/env python3
# /// script
# dependencies = ["curl_cffi", "websocket-client"]
# ///
# Scrape where every Forever recipe item comes from into
# assets/db_inputs/forever_wowhead_recipes.json.
#
# The sim doesn't read this. bayovui's recipesources.py does, for BayovCore's recipe
# planner, and it lives here because this is where we scrape wowhead. QuestieDB knows
# where most Classic recipes come from, but not the patterns Forever added (the 25xxxx
# to 27xxxx items) or the vendors Forever changed. wowhead's Forever database reads the
# beta client, so it has both.
#
# Each profession's recipe listing gives us the items. Each item's own page then gives
# its tabs: who sells it (with stock and price), who drops it, which chests hold it,
# who it's pickpocketed from, where it's fished, and which quests reward it. The drop
# chances are wowhead's looter counts, which mostly come from Classic players, so they
# say how rare a drop is rather than its exact Forever rate. Last we open each vendor's
# own page for where it stands, because a vendor Forever added has no spawn in QuestieDB.
#
#   uv run tools/database/forever_wowhead/recipes.py
#
# It's about 1,700 pages. wowhead.py asks for one every few seconds, so a full scrape
# takes a couple of hours. It keeps every page, so a run that stops picks up where it
# left off, and a run within a week only fetches what's new. A recipe wowhead lists
# with no source at all has nothing on its page either, so we don't open those.
# wowhead.py also says how we get past wowhead's bot check.
import json, os, sys

sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
from wowhead import Wowhead, listview, listview_items, mapper

OUT = os.path.join(os.path.dirname(__file__), '../../../assets/db_inputs/forever_wowhead_recipes.json')
# wowhead's recipe listings by subclass. Fishing has books, the rest are professions.
PROFESSIONS = ['alchemy', 'blacksmithing', 'cooking', 'enchanting', 'engineering', 'first-aid',
               'fishing', 'leatherworking', 'tailoring']
# A world drop is on hundreds of mobs. We keep the ones it drops from most often.
TOP_DROPS = 12


def pct(row):
    """A loot row's chance in percent, or None when wowhead has no count for it."""
    count, outof = row.get('count'), row.get('outof')
    if not count or not outof or count < 0:
        return None
    return float('%.3g' % (100 * count / outof))


def item_sources(html, npcs, objects, quests):
    """The sources on one recipe item's page, filling in npcs, objects and quests."""
    def npc(r):
        npcs.setdefault(r['id'], {k: r[k] for k in ('name', 'tag', 'minlevel', 'maxlevel', 'react', 'location', 'classification') if r.get(k) is not None})
        return r['id']
    out = {}
    sold = listview(html, 'sold-by')
    if sold:
        out['vendors'] = [{'id': npc(r), 'stock': r.get('stock'), 'cost': r.get('cost')} for r in sold]
    for tab, key in (('dropped-by', 'drops'), ('pick-pocketed-from', 'pickpocket')):
        rows = listview(html, tab)
        if rows:
            rows.sort(key=lambda r: -(pct(r) or 0))
            out[key] = {'n': len(rows), 'npcs': [{'id': npc(r), 'pct': pct(r)} for r in rows[:TOP_DROPS]]}
    rows = listview(html, 'contained-in-object')
    if rows:
        for r in rows:
            objects.setdefault(r['id'], {'name': r['name'], 'location': r.get('location')})
        out['objects'] = [{'id': r['id'], 'pct': pct(r)} for r in rows]
    rows = listview(html, 'contained-in-item')
    if rows:
        out['containers'] = [{'id': r['id'], 'name': r['name'], 'pct': pct(r)} for r in rows]
    rows = listview(html, 'fished-in')
    if rows:
        out['fished'] = [{'zone': r['id'], 'pct': pct(r)} for r in rows]
    rows = listview(html, 'reward-from-q')
    if rows:
        for r in rows:
            quests.setdefault(r['id'], {k: r[k] for k in ('name', 'level', 'reqlevel', 'side', 'reqclass', 'category') if r.get(k) is not None})
        out['quests'] = [r['id'] for r in rows]
    return out


def main():
    wh = Wowhead()
    items = {}
    for prof in PROFESSIONS:
        rows = listview_items(wh.page(f'forever/items/recipes/{prof}') or '')
        if rows is None:
            sys.exit(f'no recipe listing for {prof}, did wowhead change its URLs?')
        if len(rows) >= 1000:
            print(f'{prof}: {len(rows)} rows, wowhead may have cut the listing short', file=sys.stderr)
        for r in rows:
            if 'parentListviewRowItem' in r:
                continue
            item = {k: r[k] for k in ('id', 'name', 'quality', 'subclass', 'skill', 'source') if r.get(k) is not None}
            # What Forever changed about it, like "Skill Tailoring (35) instead of
            # Tailoring (55)" or "Sells: adds ...".
            lines = (r.get('envChange') or {}).get('lines')
            if lines:
                item['changed'] = lines
            items[r['id']] = item
        print(f'{prof}: {len(rows)} recipes', file=sys.stderr)

    npcs, objects, quests = {}, {}, {}
    sourced = [i for i in sorted(items) if items[i].get('source')]
    for n, item_id in enumerate(sourced, 1):
        html = wh.page(f'forever/item={item_id}')
        if html:
            items[item_id].update(item_sources(html, npcs, objects, quests))
        if n % 100 == 0:
            print(f'{n} of {len(sourced)} item pages ({wh.fetched} fetched, {wh.cached} from the cache)',
                  file=sys.stderr)

    vendors = sorted({v['id'] for i in items.values() for v in i.get('vendors', [])})
    for npc_id in vendors:
        html = wh.page(f'forever/npc={npc_id}')
        pins = mapper(html) if html else {}
        # {zone: [[x, y], ...]}, one zone for most vendors.
        coords = {z: [c for p in ps for c in p.get('coords', [])] for z, ps in pins.items()}
        if coords:
            npcs[npc_id]['coords'] = coords
    print(f'{len(items)} recipes, {len(npcs)} npcs ({len(vendors)} vendors), {len(objects)} objects, {len(quests)} quests', file=sys.stderr)

    # One entry per line, so a rescrape diffs recipe by recipe.
    def block(d):
        return ',\n'.join('%s:%s' % (json.dumps(str(k)), json.dumps(v, sort_keys=True, separators=(',', ':'))) for k, v in sorted(d.items()))
    with open(OUT, 'w') as f:
        f.write('{"items": [\n')
        f.write(',\n'.join(json.dumps(i, sort_keys=True, separators=(',', ':')) for i in sorted(items.values(), key=lambda x: x['id'])))
        f.write('\n],\n"npcs": {\n%s\n},\n"objects": {\n%s\n},\n"quests": {\n%s\n}}\n' % (block(npcs), block(objects), block(quests)))
    print(f'written to {OUT}', file=sys.stderr)


if __name__ == '__main__':
    main()
