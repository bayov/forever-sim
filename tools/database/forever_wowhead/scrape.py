#!/usr/bin/env python3
# Scrape wowhead's Forever item listings into assets/db_inputs/forever_wowhead_items.json.
#
# wowhead's Forever database (https://www.wowhead.com/forever/items) reads the beta
# client, so it carries the items Forever added and the Classic items it reworked (a lot
# of low level gear traded Spirit for Spell Power, dungeon weapons hit harder). The
# listing pages embed every row as `listviewitems` and each item's stats as
# `g_items[id].jsonequip`, which is all the sim needs, so we take those instead of
# parsing tooltips. The pages need a real browser (curl gets a 403), so this drives
# headless Chrome over the DevTools protocol.
#
#   pip install websocket-client
#   python3 tools/database/forever_wowhead/scrape.py 25
#   go run ./tools/database/gen_db -outDir=assets -gen=db
#
# The argument is the highest required level to scrape. Items with no required level
# (Forever quest rewards) come along whatever their item level.
import concurrent.futures, json, os, re, subprocess, sys, time, urllib.request
import websocket

MAX_REQ_LEVEL = int(sys.argv[1]) if len(sys.argv) > 1 else 25
OUT = os.path.join(os.path.dirname(__file__), '../../../assets/db_inputs/forever_wowhead_items.json')
# wowhead inventory slots: every slot a character can wear, shirts and tabards left out.
SLOTS = [1, 2, 3, 16, 5, 9, 10, 6, 7, 8, 11, 12, 13, 14, 17, 21, 22, 23, 15, 25, 26, 28]
PORT = 9334
# Row fields worth keeping, and jsonequip keys that are not stats (prices, looks).
ROW_KEYS = ['id', 'name', 'quality', 'level', 'reqlevel', 'slot', 'classs', 'subclass', 'reqclass', 'side', 'source', 'sourcemore', 'speed', 'dps', 'armor']
EQ_SKIP = {'appearances', 'displayid', 'sellprice', 'buyprice', 'avgbuyout', 'dura', 'itemSquishEraId', 'sheathtype', 'maxcount'}

EXPR = '''(() => {
    const s = [...document.scripts].map(x => x.textContent).join("\\n");
    const i = s.indexOf("listviewitems = ");
    if (i < 0) return document.body && document.body.innerText.includes("No items") ? {rows: [], g: {}} : null;
    const j = s.indexOf("];", i);
    const rows = eval(s.slice(i + 16, j + 1));
    const g = {};
    for (const r of rows) { const e = g_items[r.id]; if (e) g[r.id] = {icon: e.icon, eq: e.jsonequip}; }
    return {rows, g};
})()'''

TOOLTIP = 'https://nether.wowhead.com/forever/tooltip/item/{}?dataEnv=17&locale=0'
QUEST = 'https://www.wowhead.com/forever/quest={}'
BROWSER = 'Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 Chrome/120 Safari/537.36'
# Seconds between quest pages. Under a second gets the scrape blocked for an hour.
QUEST_PAUSE = 2

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

def mark_quest_classes(items):
    """Set qclass on the items that only one class's quest hands out.

    Forever added quests like Friend of the Library that only a mage can take, and their
    rewards carry no class restriction of their own, so an item like Erudite's Amulet
    looks wearable by everyone. The quest's own page says which class it is for, in the
    Quick Facts list it builds from a `[class=8]` markup tag.

    Only items that come from nothing but quests get the restriction. When a drop or a
    vendor also hands the item out, or one of its quests is open to every class, anyone
    can wear it.

    The pages come one at a time with a long pause between them. wowhead answers a
    browser user agent happily, but it cut us off for over an hour when we asked eight at
    a time, and asking again while cut off only made it last longer. An in-page fetch()
    from the headless browser is no way around it, that gets a 403 where curl gets a 200.
    """
    by_quest = {}
    for item in items.values():
        if set(item.get('source') or []) != {4}:
            continue
        for s in item.get('sourcemore') or []:
            if s.get('t') == 5 and s.get('ti'):
                by_quest.setdefault(s['ti'], []).append(item)
    classes = {}
    for n, quest in enumerate(sorted(by_quest), 1):
        for wait in (QUEST_PAUSE, 60, 300):
            time.sleep(wait)
            page = subprocess.run(['curl', '-sfL', '-A', BROWSER, QUEST.format(quest)],
                capture_output=True, text=True, errors='replace')
            if page.returncode == 0:
                break
            print('quest', quest, 'curl', page.returncode, 'backing off', file=sys.stderr)
        if page.returncode != 0:
            print('quest failed', quest, file=sys.stderr)
            continue
        facts = page.stdout[page.stdout.find('Quick Facts'):][:2000]
        classes[quest] = sorted({int(c) for c in re.findall(r'\[class=(\d+)\]', facts)})
        if n % 50 == 0:
            print(f'quests {n}/{len(by_quest)}', file=sys.stderr)
    for quest, quest_items in by_quest.items():
        if not classes.get(quest):
            continue
        for item in quest_items:
            others = [s['ti'] for s in item['sourcemore'] if s.get('t') == 5 and s.get('ti') != quest]
            if any(not classes.get(q) for q in others):
                continue
            item['qclass'] = sorted(set(item.get('qclass', [])) | set(classes[quest]))
    print(f"{sum(1 for i in items.values() if i.get('qclass'))} items come from a class's own quest", file=sys.stderr)


def main():
    chrome = subprocess.Popen(['google-chrome', '--headless=new', '--no-sandbox', '--disable-gpu',
        f'--remote-debugging-port={PORT}', '--remote-allow-origins=*', '--window-size=1600,1200',
        '--user-data-dir=/tmp/forever-wowhead-chrome', 'about:blank'], stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)
    try:
        for _ in range(50):
            try:
                tabs = json.load(urllib.request.urlopen(f'http://127.0.0.1:{PORT}/json')); break
            except Exception:
                time.sleep(0.2)
        ws = websocket.create_connection([t for t in tabs if t['type'] == 'page'][0]['webSocketDebuggerUrl'])
        ws.settimeout(90)
        seq = [0]
        def send(method, **params):
            seq[0] += 1
            ws.send(json.dumps({'id': seq[0], 'method': method, 'params': params}))
            while True:
                m = json.loads(ws.recv())
                if m.get('id') == seq[0]: return m.get('result', {})
        def fetch(url):
            send('Page.navigate', url=url)
            for _ in range(60):
                time.sleep(0.5)
                try:
                    val = send('Runtime.evaluate', expression=EXPR, returnByValue=True).get('result', {}).get('value')
                except Exception:
                    val = None
                if val is not None: return val
            return None
        send('Page.enable')
        items = {}
        for slot in SLOTS:
            # wowhead stops a listing somewhere above 1000 rows, so a range that comes back
            # that big is split by required level and fetched again.
            ranges = [(0, MAX_REQ_LEVEL)]
            while ranges:
                lo, hi = ranges.pop()
                url = f'https://www.wowhead.com/forever/items/slot:{slot}/min-req-level:{lo}/max-req-level:{hi}'
                val = fetch(url)
                if val is None:
                    print('failed', url, file=sys.stderr); continue
                if len(val['rows']) >= 1000 and lo < hi:
                    mid = (lo + hi) // 2
                    ranges += [(lo, mid), (mid + 1, hi)]
                    continue
                n = 0
                for r in val['rows']:
                    # Random suffix variants repeat the base row.
                    if 'parentListviewRowItem' in r or r['id'] in items: continue
                    g = val['g'].get(str(r['id']))
                    if not g: continue
                    item = {k: r[k] for k in ROW_KEYS if k in r}
                    item['icon'] = g['icon']
                    item['eq'] = {k: v for k, v in (g['eq'] or {}).items() if k not in EQ_SKIP}
                    items[r['id']] = item
                    n += 1
                print(f'slot {slot} levels {lo}-{hi}: {n} items', file=sys.stderr)
        mark_bind_on_pickup(items)
        mark_quest_classes(items)
        out = sorted(items.values(), key=lambda x: x['id'])
        # One item per line, so a rescrape diffs item by item.
        with open(OUT, 'w') as f:
            f.write('{"maxReqLevel": %d, "items": [\n' % MAX_REQ_LEVEL)
            f.write(',\n'.join(json.dumps(i, sort_keys=True, separators=(',', ':')) for i in out))
            f.write('\n]}\n')
        print(f'{len(out)} items written to {OUT}', file=sys.stderr)
    finally:
        chrome.terminate()

if __name__ == '__main__':
    main()
