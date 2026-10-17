#!/usr/bin/env python3
# /// script
# dependencies = ["curl_cffi", "websocket-client"]
# ///
# Scrape wowhead's Forever item listings into assets/db_inputs/forever_wowhead_items.json.
#
# wowhead's Forever database (https://www.wowhead.com/forever/items) reads the beta
# client, so it carries the items Forever added and the Classic items it reworked (a lot
# of low level gear traded Spirit for Spell Power, dungeon weapons hit harder). The
# listing pages embed every row as `listviewitems` and each item's stats as
# `g_items[id].jsonequip`, which is all the sim needs, so we take those instead of
# parsing tooltips. wowhead.py says how we get past wowhead's bot check.
#
# The quest listing comes along too, because a quest row is the only place that says
# which classes and which side can take it, and Forever hands out class gear from quests
# that say nothing about the class on the item itself.
#
#   uv run tools/database/forever_wowhead/scrape.py [max required level]
#   go run ./tools/database/gen_db -outDir=assets -gen=db
#
# The argument is the highest required level to scrape, by default the one the current
# file was scraped with. Items with no required level (Forever quest rewards) come along
# whatever their item level.
import concurrent.futures, json, os, sys, urllib.request

sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
from wowhead import Wowhead, gatherer, listview_items, listview_template
sys.path.insert(0, os.path.join(os.path.dirname(os.path.abspath(__file__)), '../../sources'))
from meta import line, meta

OUT = os.path.join(os.path.dirname(__file__), '../../../assets/db_inputs/forever_wowhead_items.json')

def scraped_level():
    try:
        with open(OUT) as f:
            return json.loads(f.readline() + ']}')['maxReqLevel']
    except (OSError, ValueError, KeyError):
        return 60

MAX_REQ_LEVEL = int(sys.argv[1]) if len(sys.argv) > 1 else scraped_level()
# wowhead inventory slots: every slot a character can wear, shirts and tabards left out.
SLOTS = [1, 2, 3, 16, 5, 9, 10, 6, 7, 8, 11, 12, 13, 14, 17, 21, 22, 23, 15, 25, 26, 28]
# Row fields worth keeping, and jsonequip keys that are not stats (prices, looks).
ROW_KEYS = ['id', 'name', 'quality', 'level', 'reqlevel', 'slot', 'classs', 'subclass', 'reqclass', 'side', 'source', 'sourcemore', 'speed', 'dps', 'armor']
EQ_SKIP = {'appearances', 'displayid', 'sellprice', 'buyprice', 'avgbuyout', 'dura', 'itemSquishEraId', 'sheathtype', 'maxcount'}

TOOLTIP = 'https://nether.wowhead.com/forever/tooltip/item/{}?dataEnv=16&locale=0'

def mark_bind_on_pickup(items):
    """Set bop on the crafted items that bind when picked up.

    Forever's crafted sets (Crusader's Silvered Chain, Totemic Leather, Pristine) are
    bind on pickup, so only the crafter wears them and the item is really gated on the
    profession. The listing does not say how an item binds, the tooltip does, and the
    tooltip endpoint answers plain requests.
    """
    crafted = [i for i in items.values() if 1 in i.get('source', [])]
    def fetch(item):
        try:
            with urllib.request.urlopen(TOOLTIP.format(item['id']), timeout=30) as r:
                return item, json.load(r).get('tooltip', '')
        except Exception as e:
            print('tooltip failed', item['id'], e, file=sys.stderr)
            return item, ''
    with concurrent.futures.ThreadPoolExecutor(8) as pool:
        for item, tooltip in pool.map(fetch, crafted):
            if 'Binds when picked up' in tooltip:
                item['bop'] = True
    print(f"{sum(1 for i in crafted if i.get('bop'))} of {len(crafted)} crafted items bind on pickup", file=sys.stderr)

def scrape_quests(wh):
    """Every quest wowhead knows, by id.

    A quest row says which classes and which side can take the quest, and that is the
    only place the restriction is written down when the reward carries none of its own.
    """
    quests = {}
    ranges = [(0, 60)]
    while ranges:
        lo, hi = ranges.pop()
        url = f'https://www.wowhead.com/forever/quests/min-req-level:{lo}/max-req-level:{hi}'
        html = wh.page(url)
        rows = listview_template(html, 'quest') if html else None
        if rows is None:
            print('no quest listing at', url, file=sys.stderr)
            continue
        # Same 1000 row ceiling as the item listing, so a range that comes back that big
        # is split by required level and fetched again.
        if len(rows) >= 1000 and lo < hi:
            mid = (lo + hi) // 2
            ranges += [(lo, mid), (mid + 1, hi)]
            continue
        for r in rows:
            quests[r['id']] = r
        print(f'quests {lo}-{hi}: {len(rows)}', file=sys.stderr)
    print(f'{len(quests)} quests', file=sys.stderr)
    return quests

def mark_quest_restrictions(items, quests):
    """Set qclassmask and qside on the items only one class, or one side, can quest for.

    Some class quest rewards carry no restriction of their own and would look wearable
    by anyone. wowhead gets some quests wrong too: it has the Library quests (Friend of
    the Library) as mage only, when every class can take them. ToProto has a list of
    those in foreverOpenQuests.

    Only items that come from nothing but quests get a restriction, and only when every
    one of those quests agrees. When a drop or a vendor also hands the item out, or one
    of the quests is open to everyone, anyone can wear it. A quest wowhead has no row for
    counts as open, we would rather miss a restriction than invent one.
    """
    classes = sides = 0
    for item in items.values():
        if set(item.get('source') or []) != {4}:
            continue
        qs = [quests.get(s['ti']) for s in item.get('sourcemore') or [] if s.get('t') == 5]
        if not qs or not all(qs):
            continue
        if all(q['reqclass'] for q in qs):
            mask = 0
            for q in qs:
                mask |= q['reqclass']
            item['qclassmask'] = mask
            classes += 1
        # 1 Alliance, 2 Horde. wowhead writes 3 for a quest both sides can take, and 0
        # for one it has not worked out, so only those two values restrict anything.
        side = {q['side'] for q in qs}
        if side in ({1}, {2}):
            item['qside'] = side.pop()
            sides += 1
    print(f"{classes} items come from a class's own quest, {sides} from one side's", file=sys.stderr)

def main():
    wh = Wowhead()
    items = {}
    for slot in SLOTS:
        # wowhead stops a listing somewhere above 1000 rows, so a range that comes back
        # that big is split by required level and fetched again.
        ranges = [(0, MAX_REQ_LEVEL)]
        while ranges:
            lo, hi = ranges.pop()
            url = f'https://www.wowhead.com/forever/items/slot:{slot}/min-req-level:{lo}/max-req-level:{hi}'
            html = wh.page(url)
            rows = listview_items(html) if html else None
            if rows is None:
                print('no item listing at', url, file=sys.stderr); continue
            if len(rows) >= 1000 and lo < hi:
                mid = (lo + hi) // 2
                ranges += [(lo, mid), (mid + 1, hi)]
                continue
            n = 0
            g = gatherer(html, 3)
            for r in rows:
                # Random suffix variants repeat the base row.
                if 'parentListviewRowItem' in r or r['id'] in items: continue
                e = g.get(str(r['id']))
                if not e: continue
                item = {k: r[k] for k in ROW_KEYS if k in r}
                # wowhead writes a nonsense side on the rows that carry a bonustree
                # (Willow Vest and 145 others), so only the three real values count.
                if item.get('side') not in (1, 2, 3):
                    item.pop('side', None)
                item['icon'] = e['icon']
                item['eq'] = {k: v for k, v in (e.get('jsonequip') or {}).items() if k not in EQ_SKIP}
                items[r['id']] = item
                n += 1
            print(f'slot {slot} levels {lo}-{hi}: {n} items', file=sys.stderr)
    mark_bind_on_pickup(items)
    mark_quest_restrictions(items, scrape_quests(wh))
    out = sorted(items.values(), key=lambda x: x['id'])
    # One item per line, so a rescrape diffs item by item.
    with open(OUT, 'w') as f:
        f.write('{' + line(meta('www.wowhead.com/forever', 'tools/database/forever_wowhead/scrape.py', dataEnv=16)))
        f.write('"maxReqLevel": %d, "items": [\n' % MAX_REQ_LEVEL)
        f.write(',\n'.join(json.dumps(i, sort_keys=True, separators=(',', ':')) for i in out))
        f.write('\n]}\n')
    print(f'{len(out)} items written to {OUT}', file=sys.stderr)

if __name__ == '__main__':
    main()
