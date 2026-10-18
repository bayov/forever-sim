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
# Each profession's spell listing then gives us what a trainer asks for each recipe it
# teaches and for each rank, like Journeyman (the money and the skill), the specialization a recipe asks for (Armorsmith,
# Gnomish Engineer, Dragonscale Leatherworking...), and the reagents every recipe uses.
# A trainer recipe's own page says which trainers teach it, since not every trainer
# teaches every recipe: a Journeyman trainer stops early, and only a Goblin Engineer
# teaches Goblin Sapper Charge. BayovCore only learns a trainer's prices when a character
# opens one, so until then it plans with these. Each reagent's page says who sells it
# and whether in limited stock, so the planner knows which reagents a vendor restocks,
# like Coarse Thread, and where, before a character has opened that vendor.
#
#   uv run tools/database/forever_wowhead/recipes.py
#
# It's about 2,900 pages. wowhead.py asks for one every few seconds, so a full scrape
# takes a couple of hours. It keeps every page, so a run that stops picks up where it
# left off, and a run within a week only fetches what's new. A recipe wowhead lists
# with no source at all has nothing on its page either, so we don't open those.
# wowhead.py also says how we get past wowhead's bot check.
import datetime, json, os, sys, time

sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
from wowhead import Wowhead, listview, listview_items, listview_spells, mapper
sys.path.insert(0, os.path.join(os.path.dirname(os.path.abspath(__file__)), '../../sources'))
from meta import line, meta

OUT = os.path.join(os.path.dirname(__file__), '../../../assets/db_inputs/forever_wowhead_recipes.json')
# wowhead's recipe listings by subclass. Fishing has books, the rest are professions.
PROFESSIONS = ['alchemy', 'blacksmithing', 'cooking', 'enchanting', 'engineering', 'first-aid',
               'fishing', 'leatherworking', 'tailoring']
# wowhead's spell listings of the professions, with each one's skill line. Mining's are
# the smelting recipes, and herbalism, skinning and fishing have only their ranks.
# Cooking, first aid and fishing are secondary skills, and a wrong path doesn't 404 but
# gives every profession's spells, cut at 1,000. So we keep only the rows of the skill
# line we asked for. Rogue poisons aren't here: neither professions/poisons nor
# class-skills/rogue/poisons is their listing.
SPELL_PROFESSIONS = {
    'professions/alchemy': 171, 'professions/blacksmithing': 164, 'professions/enchanting': 333,
    'professions/engineering': 202, 'professions/leatherworking': 165, 'professions/mining': 186,
    'professions/tailoring': 197, 'professions/herbalism': 182, 'professions/skinning': 393,
    'secondary-skills/cooking': 185, 'secondary-skills/first-aid': 129, 'secondary-skills/fishing': 356,
}
# A world drop is on hundreds of mobs. We keep the ones it drops from most often.
TOP_DROPS = 12


def pct(row):
    """A loot row's chance in percent, or None when wowhead has no count for it."""
    count, outof = row.get('count'), row.get('outof')
    if not count or not outof or count < 0:
        return None
    return float('%.3g' % (100 * count / outof))


def keep_npc(npcs, r):
    """Keeps what an NPC row of a listing says about the NPC, and answers its id."""
    npcs.setdefault(r['id'], {k: r[k] for k in ('name', 'tag', 'minlevel', 'maxlevel', 'react', 'location', 'classification') if r.get(k) is not None})
    return r['id']


def item_sources(html, npcs, objects, quests):
    """The sources on one recipe item's page, filling in npcs, objects and quests."""
    def npc(r):
        return keep_npc(npcs, r)
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

    # What a trainer asks for each recipe, by spell. A recipe a trainer doesn't teach
    # has no trainingcost, and the reagents of every recipe go in `reagents`.
    # `specializations` is the spell a recipe asks the crafter to know, by recipe.
    trainers, specializations, reagents = {}, {}, {}
    for prof, skill in SPELL_PROFESSIONS.items():
        rows = listview_spells(wh.page(f'forever/spells/{prof}') or '')
        rows = [r for r in rows or [] if skill in (r.get('skill') or [])]
        if not rows:
            sys.exit(f'no spell listing for {prof}, did wowhead change its URLs?')
        if len(rows) >= 1000:
            print(f'{prof}: {len(rows)} spells, wowhead may have cut the listing short', file=sys.stderr)
        taught = 0
        for r in rows:
            # A training cost is the test. Some specialization recipes have one but no
            # `source` at all, like the armorsmith's and the goblin engineer's.
            if r.get('trainingcost') is not None:
                trainers[r['id']] = {'cost': r['trainingcost'], 'req': r.get('learnedat') or 1}
                # A rank of the profession itself, like Journeyman Tailoring, is in the
                # listing too, with the skill line it raises.
                if r.get('rank'):
                    trainers[r['id']].update(rank=r['rank'].strip(), skill=skill)
                taught += 1
            if r.get('specialization'):
                specializations[r['id']] = r['specialization']
            for reagent, _ in r.get('reagents') or []:
                reagents.setdefault(reagent, {})
        print(f'{prof}: {len(rows)} spells, {taught} from a trainer', file=sys.stderr)

    # Which trainers teach each recipe. The NPCs go in `npcs`, like the vendors.
    for n, spell in enumerate(sorted(trainers), 1):
        html = wh.page(f'forever/spell={spell}')
        taught = listview(html, 'taught-by-npc') if html else []
        trainers[spell]['npcs'] = sorted(keep_npc(npcs, r) for r in taught)
        if n % 100 == 0:
            print(f'{n} of {len(trainers)} trainer recipe pages ({wh.fetched} fetched, {wh.cached} from the cache)',
                  file=sys.stderr)

    # Who sells each reagent. `stock` is -1 when the vendor never runs out, and `cost`
    # is [[copper, [[currency, n]], [[item, n]]]], since some vendors ask for currency
    # or items, like Lokhtos' Essences for Firelands Embers.
    for n, item_id in enumerate(sorted(reagents), 1):
        html = wh.page(f'forever/item={item_id}')
        sold = listview(html, 'sold-by') if html else []
        if sold:
            reagents[item_id]['vendors'] = [{'id': keep_npc(npcs, r), 'stock': r.get('stock'), 'cost': r.get('cost')}
                                            for r in sold]
        if n % 100 == 0:
            print(f'{n} of {len(reagents)} reagent pages ({wh.fetched} fetched, {wh.cached} from the cache)',
                  file=sys.stderr)
    reagents = {i: r for i, r in reagents.items() if r}
    print(f'{len(trainers)} recipes from a trainer, {len(specializations)} that ask for a specialization, '
          f'{len(reagents)} reagents a vendor sells', file=sys.stderr)

    # One entry per line, so a rescrape diffs recipe by recipe.
    def block(d):
        return ',\n'.join('%s:%s' % (json.dumps(str(k)), json.dumps(v, sort_keys=True, separators=(',', ':'))) for k, v in sorted(d.items()))
    with open(OUT, 'w') as f:
        # `oldestPage` is when the oldest page we used was fetched. A page comes from the
        # cache for a week, so it can be older than `fetched`. bayovui weighs what BayovCore
        # saw in game against it.
        oldest = datetime.datetime.fromtimestamp(wh.oldest or time.time(), datetime.timezone.utc)
        f.write('{' + line(meta('www.wowhead.com/forever', 'tools/database/forever_wowhead/recipes.py', dataEnv=16,
                                oldestPage=oldest.strftime('%Y-%m-%dT%H:%MZ'))))
        f.write('"items": [\n')
        f.write(',\n'.join(json.dumps(i, sort_keys=True, separators=(',', ':')) for i in sorted(items.values(), key=lambda x: x['id'])))
        f.write('\n],\n"npcs": {\n%s\n},\n"objects": {\n%s\n},\n"quests": {\n%s\n},\n"trainers": {\n%s\n},\n"specializations": {\n%s\n},\n"reagents": {\n%s\n}}\n'
                % (block(npcs), block(objects), block(quests), block(trainers), block(specializations), block(reagents)))
    print(f'written to {OUT}', file=sys.stderr)


if __name__ == '__main__':
    main()
