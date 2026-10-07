#!/usr/bin/env python3
# Scrape the random suffixes each item can roll from https://foreverchanges.pro/item/ID into
# assets/db_inputs/foreverchanges_suffixes.json.
#
# Our suffix lists come from Classic wowhead, which only has the suffixes players saw on
# each item. Cutthroat's Mantle there rolls 12 suffixes and none of them is "of the
# Gorilla" (Strength and Intellect), but the beta client gives it 31, and the Gorilla one is
# what ForeverChanges puts on its level 30 enhancement shaman. The item pages list every
# suffix from the client with its stats, so we add the ones wowhead never saw. When wowhead
# has a suffix name for the item we keep wowhead's values, because the site computes its
# numbers from the item level and is sometimes 1 off (Unearthed Bands of Intellect is +10 on
# wowhead and +9 on the site).
#
# Forever also added suffixes Classic never had ("of Magic" spell power, "of the Knight",
# "of the Bandit"). Those, and the ones whose stats match no wowhead suffix, get new IDs
# from 990000 up.
#
# We scrape the items in assets/database/db.json that already roll suffixes and need level
# MAX_LEVEL or lower. The pages are cached in CACHE_DIR, so a rerun only fetches what is new.
#
#   python3 tools/database/foreverchanges/suffixes.py 30 /tmp/fc_items
#   go run ./tools/database/gen_db -outDir=assets -gen=db
import html, json, os, re, sys, time, urllib.request

ROOT = os.path.join(os.path.dirname(__file__), '../../..')
OUT = os.path.join(ROOT, 'assets/db_inputs/foreverchanges_suffixes.json')
FIRST_NEW_ID = 990000

# wowhead's suffix stat keys and the site's stat names, as indices of proto.Stat.
WOWHEAD_STATS = {'str': 0, 'agi': 1, 'sta': 2, 'int': 3, 'spi': 4, 'spldmg': 5, 'arcsplpwr': 6, 'firsplpwr': 7,
                 'frosplpwr': 8, 'holsplpwr': 9, 'natsplpwr': 10, 'shasplpwr': 11, 'manargn': 12, 'mlecritstrkpct': 19,
                 'mleatkpwr': 17, 'armor': 26, 'rgdatkpwr': 27, 'def': 28, 'blockpct': 29, 'dodgepct': 31, 'arcres': 35,
                 'firres': 36, 'frores': 37, 'natres': 38, 'shares': 39, 'splheal': 41}
SITE_STATS = {'Strength': [0], 'Agility': [1], 'Stamina': [2], 'Intellect': [3], 'Spirit': [4], 'Spell Power': [5],
              'Arcane Spell Damage': [6], 'Fire Spell Damage': [7], 'Frost Spell Damage': [8], 'Holy Spell Damage': [9],
              'Nature Spell Damage': [10], 'Shadow Spell Damage': [11], 'Mana Regeneration': [12],
              'Attack Power': [17, 27], 'Defense Rating': [28], 'Arcane Resistance': [35], 'Fire Resistance': [36],
              'Frost Resistance': [37], 'Nature Resistance': [38], 'Shadow Resistance': [39], 'Healing': [41],
              'Spell Damage': [42]}


def fetch(item_id, cache_dir):
    path = os.path.join(cache_dir, f'{item_id}.html')
    if os.path.exists(path) and os.path.getsize(path) > 0:
        return open(path).read()
    req = urllib.request.Request(f'https://foreverchanges.pro/item/{item_id}', headers={'User-Agent': 'Mozilla/5.0'})
    with urllib.request.urlopen(req, timeout=60) as r:
        page = r.read().decode()
    open(path, 'w').write(page)
    time.sleep(0.3)
    return page


def site_suffixes(page):
    # Each suffix is <li><b>of the Gorilla</b><span><span class="it-suffix-set">+7 Strength, +7 Intellect</span>...
    out = []
    for name, body in re.findall(r'<li><b class="q\d">([^<]*)</b><span>(.*?)</span></li>', page):
        for stat_set in re.findall(r'<span class="it-suffix-set">(?:<!-- -->)?([^<]*)</span>', body):
            out.append((html.unescape(name), [p.strip() for p in html.unescape(stat_set).split(',')]))
    return out


def site_stats(parts):
    stats = {}
    for part in parts:
        m = re.match(r'\+(\d+) (.*)$', part)
        if m and m.group(2) in SITE_STATS:
            for stat in SITE_STATS[m.group(2)]:
                stats[stat] = int(m.group(1))
            continue
        m = re.match(r'increases your chance to dodge an attack by (\d+)%', part)
        if m:
            stats[31] = int(m.group(1))
            continue
        # The sim has no health regeneration stat.
        if 'Health Regeneration' in part:
            continue
        sys.exit(f'unknown suffix stat: {part}')
    return stats


def main():
    max_level, cache_dir = int(sys.argv[1]), sys.argv[2]
    os.makedirs(cache_dir, exist_ok=True)

    planner = open(os.path.join(ROOT, 'assets/db_inputs/wowhead_gearplannerdb.txt')).read()
    start = planner.index('wow.gearPlanner.classic.randomEnchant",') + len('wow.gearPlanner.classic.randomEnchant",')
    wowhead, _ = json.JSONDecoder().raw_decode(planner, start)
    known = {}
    for key, suffix in wowhead.items():
        stats = {WOWHEAD_STATS[k]: v for k, v in suffix['stats'].items() if v and k in WOWHEAD_STATS}
        known.setdefault((suffix['name'], json.dumps(sorted(stats.items()))), int(key))

    db = json.load(open(os.path.join(ROOT, 'assets/database/db.json')))
    suffix_names = {s['id']: s['name'] for s in db['randomSuffixes']}
    items = [i for i in db['items']
             if i.get('randomSuffixOptions') and (i.get('requiredLevel') or i.get('ilvl', 0) - 6) <= max_level]

    new, options = {}, {}
    for item in sorted(items, key=lambda i: i['id']):
        have = {suffix_names[o] for o in item['randomSuffixOptions']}
        added = []
        for name, parts in site_suffixes(fetch(item['id'], cache_dir)):
            if name in have:
                continue
            key = (name, json.dumps(sorted(site_stats(parts).items())))
            if key in known:
                suffix_id = known[key]
            else:
                suffix_id = new.setdefault(key, FIRST_NEW_ID + len(new))
            if suffix_id not in added:
                added.append(suffix_id)
        if added:
            options[str(item['id'])] = added

    out = {
        'suffixes': [{'id': i, 'name': k[0], 'stats': {str(s): v for s, v in json.loads(k[1])}} for k, i in new.items()],
        'items': options,
    }
    json.dump(out, open(OUT, 'w'), indent=None, separators=(',', ':'))
    print(f'{len(items)} items, {len(options)} gain suffixes, {sum(map(len, options.values()))} in all, {len(new)} new suffixes')


if __name__ == '__main__':
    main()
