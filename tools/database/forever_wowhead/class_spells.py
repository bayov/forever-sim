#!/usr/bin/env python3
# /// script
# dependencies = ["curl_cffi", "websocket-client"]
# ///
# Scrape what a class trainer asks for every class spell, and where each rank comes
# from, into assets/db_inputs/forever_wowhead_class_spells.json.
#
# The sim doesn't read this. bayovui's classspells.py does, for BayovCore's Training
# module, and it lives here because this is where we scrape wowhead. The client's game
# data says which spells a class has and the level of each rank, but not what a trainer
# charges, since that's the server's. BayovCore learns the prices when a character opens
# a class trainer, so until then it plans with these.
#
# wowhead's ability listing of each class (spells/abilities/<class>) has every rank,
# with its level, its `trainingcost` and its `source`. Source 6 is a trainer, 4 a quest
# (the shaman's first rank of each element's totem) and 2 a drop (Healing Wave 10, from
# a book). A rank with no source at all is one wowhead doesn't know yet, like Forever's
# own Lava Burst 4. It's nine pages, one per class.
#
#   uv run tools/database/forever_wowhead/class_spells.py
#
# wowhead.py says how we get past wowhead's bot check, and keeps every page for a week.
import datetime, json, os, sys, time

sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
from wowhead import Wowhead, listview_spells
sys.path.insert(0, os.path.join(os.path.dirname(os.path.abspath(__file__)), '../../sources'))
from meta import line, meta

OUT = os.path.join(os.path.dirname(__file__), '../../../assets/db_inputs/forever_wowhead_class_spells.json')
# wowhead's class names in its URLs, by the game's class file name.
CLASSES = {
    'WARRIOR': 'warrior', 'PALADIN': 'paladin', 'HUNTER': 'hunter', 'ROGUE': 'rogue',
    'PRIEST': 'priest', 'SHAMAN': 'shaman', 'MAGE': 'mage', 'WARLOCK': 'warlock', 'DRUID': 'druid',
}


def main():
    wh = Wowhead()
    out = {}
    for cls, path in CLASSES.items():
        rows = listview_spells(wh.page(f'forever/spells/abilities/{path}') or '')
        if not rows:
            sys.exit(f'no ability listing for {cls}, did wowhead change its URLs?')
        if len(rows) >= 1000:
            print(f'{cls}: {len(rows)} spells, wowhead may have cut the listing short', file=sys.stderr)
        spells = {}
        for r in rows:
            s = {'name': r['name'], 'level': r.get('level') or 0}
            if r.get('rank'):
                s['rank'] = r['rank'].strip()
            if r.get('trainingcost') is not None:
                s['cost'] = r['trainingcost']
            if r.get('source'):
                s['source'] = r['source']
            spells[r['id']] = s
        out[cls] = spells
        priced = sum(1 for s in spells.values() if 'cost' in s)
        print(f'{cls}: {len(spells)} spells, {priced} with a trainer cost', file=sys.stderr)

    with open(OUT, 'w') as f:
        oldest = datetime.datetime.fromtimestamp(wh.oldest or time.time(), datetime.timezone.utc)
        f.write('{' + line(meta('www.wowhead.com/forever', 'tools/database/forever_wowhead/class_spells.py',
                                dataEnv=16, oldestPage=oldest.strftime('%Y-%m-%dT%H:%MZ'))))
        # One spell per line, so a rescrape diffs rank by rank.
        blocks = []
        for cls in sorted(out):
            body = ',\n'.join('%s:%s' % (json.dumps(str(k)), json.dumps(v, sort_keys=True, separators=(',', ':')))
                              for k, v in sorted(out[cls].items()))
            blocks.append('%s: {\n%s\n}' % (json.dumps(cls), body))
        f.write('"classes": {\n%s\n}}\n' % ',\n'.join(blocks))
    print(f'written to {OUT}', file=sys.stderr)


if __name__ == '__main__':
    main()
