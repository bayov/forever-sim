#!/usr/bin/env python3
# Scrape the dungeon loot tables of https://foreverchanges.pro into
# assets/db_inputs/foreverchanges_loot.json.
#
# wowhead's Forever database knows the items Forever added, but not where most of them
# drop, so the gear search used to leave out every new dungeon drop as unconfirmed
# (First Mate Band, Lookie's Spyglass, the whole Ruins of Lordaeron). ForeverChanges
# lists every boss of each dungeon with its loot, from the beta client and from what
# players loot in the beta, so gen_db takes the drop sources from here. We scrape the
# dungeons up to the level 37 to 46 ones, because a level 30 character wears their low
# level drops. The new Forever dungeons of that range (Excavation Site, City of Dalaran,
# The Drowned City) have no loot tables on the site yet.
#
# The pages are Next.js server rendered, and the loot sits in the React payload
# (`self.__next_f.push`) as one JSON object per boss:
# {"name": "Witherfang", "kind": "boss", "level": 17, "items": [{"i": 271201, "n": ..., "r": 15, "x": [tooltip lines]}]}.
#
#   python3 tools/database/foreverchanges/scrape.py
#   go run ./tools/database/gen_db -outDir=assets -gen=db
import json, os, re, sys, time, urllib.request

OUT = os.path.join(os.path.dirname(__file__), '../../../assets/db_inputs/foreverchanges_loot.json')
DUNGEONS = ['ragefire-chasm', 'the-deadmines', 'wailing-caverns', 'hall-of-thanes', 'ruins-of-lordaeron', 'shadowfang-keep',
            'blackfathom-deeps', 'the-stockade', 'gnomeregan', 'razorfen-kraul',
            'scarlet-monastery-graveyard', 'scarlet-monastery-library', 'scarlet-monastery-armory',
            'scarlet-monastery-cathedral', 'razorfen-downs']
BOSS = re.compile(r'\{"name":"[^"]*","kind":"[^"]*"')


def fetch(path):
    req = urllib.request.Request('https://foreverchanges.pro' + path, headers={'User-Agent': 'Mozilla/5.0'})
    with urllib.request.urlopen(req, timeout=60) as r:
        return r.read().decode()


def payload(page):
    chunks = re.findall(r'self\.__next_f\.push\(\[1,(".*?")\]\)</script>', page, re.S)
    return ''.join(json.loads(c) for c in chunks)


def main():
    rows = []
    dec = json.JSONDecoder()
    for slug in DUNGEONS:
        text = payload(fetch('/dungeons/' + slug))
        n = 0
        for m in BOSS.finditer(text):
            try:
                boss, _ = dec.raw_decode(text[m.start():])
            except ValueError:
                continue
            for item in boss.get('items') or []:
                rows.append({
                    'dungeon': slug, 'boss': boss['name'], 'kind': boss['kind'],
                    'id': item['i'], 'name': item['n'], 'reqlevel': item.get('r', 0),
                    # new, changed or same as Classic.
                    'status': item.get('t', ''), 'tooltip': item.get('x', []),
                })
                n += 1
        print(f'{slug}: {n} drops', file=sys.stderr)
        time.sleep(1)
    with open(OUT, 'w') as f:
        f.write('{"items": [\n')
        f.write(',\n'.join(json.dumps(r, ensure_ascii=False, separators=(',', ':')) for r in rows))
        f.write('\n]}\n')
    print(f'{len(rows)} drops written to {OUT}', file=sys.stderr)


if __name__ == '__main__':
    main()
