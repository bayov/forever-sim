#!/usr/bin/env python3
# /// script
# dependencies = ["curl_cffi", "websocket-client"]
# ///
# Scrape who starts each of the quests Forever added, and where that giver stands, into
# assets/db_inputs/forever_wowhead_new_quests.json.
#
# The sim doesn't read this. BayovCore does, through bayovui, to draw a "!" on the map
# for quests QuestieDB doesn't know (Forever's new quests, like The Legashi Threat from
# Ekkal in Azshara). QuestieDB has nothing on them, but wowhead's Forever database reads
# the beta client, so it has their givers and where they stand.
#
# Each quest's own page gives its name, levels and side (from g_quests), its starters
# and enders (from the infobox), and usually a map with every spawn of each. We open each
# NPC starter's own page once for how it reacts to each faction, since the quest's side
# can be wrong. When the map has no pin for a starter (it's in an instance, or wowhead
# didn't draw it), we take its spawns from its own page too. A quest an item
# starts has no position: the item drops somewhere, and that isn't a giver on the map.
#
#   uv run tools/database/forever_wowhead/new_quests.py quest_ids.txt
#   uv run tools/database/forever_wowhead/new_quests.py     # refresh the quests in the file
#
# The ID file has one quest ID per line. Without one we refresh the quests the current
# output already lists, found or missing. It's one page a quest and one a giver the map
# misses, so about 1,500 quests take five hours or more at wowhead.py's pace. Every page is
# cached, so a run that stops picks up where it left off.
import datetime, html as htmllib, json, os, re, sys, time

sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
from wowhead import Wowhead, _js_value, listview, mapper
sys.path.insert(0, os.path.join(os.path.dirname(os.path.abspath(__file__)), '../../sources'))
from meta import meta

OUT = os.path.join(os.path.dirname(__file__), '../../../assets/db_inputs/forever_wowhead_new_quests.json')
# A giver like a guard or a wandering vendor can have dozens of spawns. The map icon
# only needs a few, so we keep at most this many, spread over the list.
MAX_SPOTS = 20
SIDES = {1: 'alliance', 2: 'horde'}
# The mapper's object types, the way wowhead numbers them.
TYPES = {1: 'npc', 2: 'object', 3: 'item'}


def infobox(page):
    """The quick facts box of a page, as wowhead markup with its escapes undone."""
    m = re.search(r'WH\.markup\.printHtml\((")\[ul\]', page)
    if not m:
        return ''
    return json.JSONDecoder().raw_decode(page, m.start(1))[0]


def givers(box, label):
    """The NPCs, objects and items after "Start:" or "End:" in the infobox."""
    out = []
    for li in box.split('[li]'):
        m = re.search(r'\]%s:' % label, li)
        if not m:
            continue
        for t, i, name in re.findall(r'\[url=/forever/(npc|object|item)=(\d+)[^\]]*\]([^\[]*)\[/url\]', li[m.end():]):
            out.append({'type': t, 'id': int(i), 'name': htmllib.unescape(name)})
    return out


def quest_pins(page):
    """The quest map's start and end pins: {(point, type, id): [(zone, x, y)]}."""
    i = page.find('new Mapper(')
    if i < 0:
        return {}
    try:
        data = _js_value(page, i + len('new Mapper('))
    except ValueError:
        return {}
    pins = {}
    for zone, obj in (data.get('objectives') or {}).items():
        if not str(zone).lstrip('-').isdigit():
            continue
        for level in obj.get('levels') or []:
            for p in level or []:
                if p.get('point') not in ('start', 'end') or p.get('type') not in TYPES:
                    continue
                spots = pins.setdefault((p['point'], TYPES[p['type']], p.get('id')), [])
                coords = p.get('coords') or ([p['coord']] if p.get('coord') else [])
                spots += [(int(zone), c[0], c[1]) for c in coords if is_point(c)]
    return pins


def is_point(c):
    """Whether a pin's coordinate is an [x, y] pair. A pin with no position has [""],
    like Rugged Traveler on quest 82062."""
    return isinstance(c, list) and len(c) == 2 and all(isinstance(v, (int, float)) for v in c)


def page_pins(page):
    """Every spawn on an NPC's or object's own page: [(zone, x, y)]."""
    return [(int(zone), c[0], c[1]) for zone, ps in mapper(page).items() if str(zone).lstrip('-').isdigit()
            for p in ps for c in p.get('coords') or [] if is_point(c)]


def react(page, npc_id):
    """How an NPC reacts to each faction: 1 friendly, 0 neutral, -1 hostile.

    The quest's side isn't enough to tell who can take it. Wowhead says The Legashi
    Threat is Alliance, but a Horde orc took it from Ekkal. So we keep the NPC's own
    react from its g_npcs entry. It's None for a faction wowhead has no data on yet,
    which is common for NPCs Forever added (Ekkal is [1, None]).
    """
    m = re.search(r'g_npcs\[%d\], ' % npc_id, page)
    r = (_js_value(page, m.end()).get('react') if m else None) or [None, None]
    return {'alliance': r[0], 'horde': r[1]}


def spread(spots):
    """The spots without repeats, at most MAX_SPOTS of them spread over the list."""
    spots = list(dict.fromkeys(spots))
    if len(spots) > MAX_SPOTS:
        step = len(spots) / MAX_SPOTS
        spots = [spots[int(k * step)] for k in range(MAX_SPOTS)]
    return [{'zone': z, 'x': x, 'y': y} for z, x, y in spots]


def main():
    if len(sys.argv) > 1:
        with open(sys.argv[1]) as f:
            ids = sorted({int(l) for l in f.read().split() if l.strip().isdigit()})
    else:
        with open(OUT) as f:
            old = json.load(f)
        ids = sorted({int(i) for i in old['quests']} | set(old.get('missing', [])))

    wh = Wowhead()
    quests, missing = {}, []
    # Each giver's own page's spawns and react, so a giver of many quests is opened once.
    giver_page = {}
    for n, qid in enumerate(ids, 1):
        page = wh.page(f'forever/quest={qid}')
        m = page and re.search(r'g_quests\[%d\], ' % qid, page)
        if not m:
            missing.append(qid)
            continue
        info = _js_value(page, m.end())
        box = infobox(page)
        pins = quest_pins(page)
        q = {'name': info.get('name'), 'level': info.get('level'), 'reqLevel': info.get('reqlevel'),
             'side': SIDES.get(info.get('side'), 'both')}
        for point, key in (('start', 'Start'), ('end', 'End')):
            entries = []
            for g in givers(box, key):
                if g['type'] != 'item':
                    spots = pins.get((point, g['type'], g['id']))
                    # A starter's own page tells us how it reacts to each faction, and
                    # its spawns when the quest map has no pin for it. We only open it
                    # for starters (and an NPC always): enders are a bonus.
                    if point == 'start' and (g['type'] == 'npc' or not spots):
                        k = (g['type'], g['id'])
                        if k not in giver_page:
                            gp = wh.page(f'forever/{g["type"]}={g["id"]}')
                            giver_page[k] = ((page_pins(gp), react(gp, g['id']), g['name'],
                                              [r['id'] for r in listview(gp, 'starts')]) if gp else ([], None, g['name'], []))
                        spots = spots or giver_page[k][0]
                        if g['type'] == 'npc':
                            g['react'] = giver_page[k][1]
                    g['spots'] = spread(spots or [])
                entries.append(g)
            if entries or point == 'start':
                q[point] = entries
        quests[str(qid)] = q
        if n % 50 == 0:
            print(f'{n} of {len(ids)} quests ({wh.fetched} pages fetched, {wh.cached} from the cache)', file=sys.stderr)

    # Some quest pages have no "Start:" at all, while a giver's own page lists the quest
    # in its "Starts" tab. The Pact (93773) is one: only Dendrite Starblaze's page says
    # he starts it. So we fill those in from the giver pages we already opened.
    for (t, i), (spots, r, name, starts) in giver_page.items():
        for qid in starts:
            q = quests.get(str(qid))
            if q is not None and not q['start']:
                g = {'type': t, 'id': i, 'name': name}
                if t == 'npc':
                    g['react'] = r
                g['spots'] = spread(spots)
                q['start'].append(g)

    out = {'meta': meta('www.wowhead.com/forever', 'tools/database/forever_wowhead/new_quests.py', dataEnv=16),
           'source': 'wowhead forever',
           'scraped': datetime.datetime.now(datetime.timezone.utc).strftime('%Y-%m-%dT%H:%MZ'),
           'missing': missing}
    # One quest per line, so a rescrape diffs quest by quest.
    with open(OUT, 'w') as f:
        head = json.dumps(out, sort_keys=True)[:-1]
        body = ',\n'.join('%s:%s' % (json.dumps(k), json.dumps(v, separators=(',', ':')))
                          for k, v in sorted(quests.items(), key=lambda kv: int(kv[0])))
        f.write(head + ',\n"quests": {\n' + body + '\n}}\n')

    starts = [s for q in quests.values() for s in q['start']]
    givers_ = [q for q in quests.values() if any(s['type'] != 'item' for s in q['start'])]
    placed = [q for q in quests.values() if any(s.get('spots') for s in q['start'])]
    items = [q for q in quests.values() if q['start'] and all(s['type'] == 'item' for s in q['start'])]
    nogiver = [k for k, q in quests.items() if not q['start']]
    unplaced = [k for k, q in quests.items() if any(s['type'] != 'item' for s in q['start'])
                and not any(s.get('spots') for s in q['start'])]
    print(f'{len(quests)} quests found, {len(givers_)} with an NPC or object giver, {len(placed)} with positions, '
          f'{len(items)} item-started, {len(missing)} missing ({wh.fetched} pages fetched, {wh.cached} from the cache)',
          file=sys.stderr)
    if nogiver:
        print(f'{len(nogiver)} with no starter: {" ".join(nogiver[:40])}', file=sys.stderr)
    if unplaced:
        print(f'{len(unplaced)} with a giver but no position: {" ".join(unplaced[:40])}', file=sys.stderr)
    npcs = {s['id']: s.get('react') or {} for s in starts if s['type'] == 'npc'}
    print(f'{len(npcs)} NPC starters, {sum(r.get("alliance") is None for r in npcs.values())} with no alliance react, '
          f'{sum(r.get("horde") is None for r in npcs.values())} with no horde react', file=sys.stderr)
    zones = sorted({sp['zone'] for s in starts for sp in s.get('spots', [])})
    print(f'{len(zones)} zones: {" ".join(map(str, zones))}', file=sys.stderr)
    print(f'written to {OUT}', file=sys.stderr)


if __name__ == '__main__':
    main()
