#!/usr/bin/env python3
"""Price the Phase 0 BiS list's items in gold per DPS, from our auction house scans.

    python3 tools/rotopt/phase0_prices.py            # print the table
    python3 tools/rotopt/phase0_prices.py --write    # save gold and goldPerDps in phase0_content.json

Only an item that beats the cheaper items of its slot gets the two numbers,
since the line only matters for those.

The Phase 0 list doesn't pick an item that costs too much gold for the DPS it
adds (user, 2026-11-22). bis_phase0.sh makes every item whose goldPerDps is
over MAX_GOLD_PER_DPS (30 by default) a -gear-no-pick item, and marks it
"infeasible" on the list. This script works out the numbers it reads.

An item's gold is the cheapest of three prices. One is what it sells for on
the auction house now: the newest day BayovCore priced it in its scans. Prices
move a lot this early (Brain Hacker went from about 500g to 255g in a week,
and Mooncloth from 21g to 56g), so an older day says little about what we'd
pay today. The second is what its recipe's materials sell for there, since we
can craft it or have it crafted (the recipe's pattern isn't counted).

The third is the effort of farming it in a dungeon (user, 2026-11-22). We
count the hours it takes at 30g an hour (--gold-per-hour). Without it a rare
drop cost nothing, so the list picked Treant's Bane (a 0.38% drop from the
Dire Maul tribute) over Brain Hacker (255g on the auction house). The hours
are the runs we expect times the time of a run. The runs are 1 / (drop chance
x roll share), and the share is 0.5 (--roll-share) because we assume one other
player in the group needs it too. A run takes the dungeon's typical clear
time plus 20 minutes (--run-overhead), for getting there and finding a group
(user, 2026-11-22). When an item drops from several bosses, we take the one
with the fewest hours.

So a bind on pickup drop costs only its effort, and a bind on equip drop costs
the cheaper of its effort and the auction house. An item that is also a quest
reward, a vendor item or a reputation reward gets no effort price yet, and
neither does a world drop outside dungeons (its chance is too small to farm,
so only the auction house counts). An item that has none of these prices
costs nothing, like before.

Its DPS is what it adds over the best item we'd wear instead, which is the
best item of the slot that costs less and is feasible itself. We go through a
slot from the cheapest item up, so a 250g item is measured against the best
of the free and cheap ones, and a 500g item may be measured against that 250g
one when it's worth its price. Rings and trinkets take two slots, so there the
item is measured against the second best cheaper item, the one it would push
out. goldPerDps is (gold - that item's gold) / (DPS gained). An item that
doesn't beat the cheaper one gets no goldPerDps, because the list never picks
it over that one anyway.

The DPS comes from the list's own losses (ui/enhancement_shaman/bis/
level60_phase0.bis.json), and infeasible items are ranked with their real
loss, so their numbers are there too. The picks can move after a rerun, so
run this again after one and see whether anything crosses the line.

The scans are BayovCore's BayovAuctionData saved variables. We take the
newest such file under the WoW folder (--wow, or $WOW_DIR), or the one --ah
names. The recipes are the client's SpellEffect and SpellReagents tables from
wago.tools, in ~/forever-logs/wago/<build>/ (--wago).

The drops come from the WoW folder's BayovUI wiki, the newest copy of it under
bayovui/addons or the game's AddOns. BayovWikiGear says which items bind on
pickup and which dungeon bosses drop them, with the chance from Naowh's
Dungeon Journal. BayovWikiDungeons says which wing each boss is in. When the
wiki doesn't know a chance (it has 0 or no boss), we take AtlasLoot Classic's
droprate.lua from the game's AddOns, for the bosses of that dungeon. That is
where Treant's Bane's 0.38% comes from: AtlasLoot puts the tribute's loot on
Cho'Rush the Observer. A trash drop with no AtlasLoot rate gets no effort
price, because its chance is per mob and not per run. The clear times are
estimates in RUN_MINUTES below.
"""
import argparse
import csv
import glob
import json
import os
import re
import statistics
import sys

HERE = os.path.dirname(os.path.abspath(__file__))
ROOT = os.path.abspath(os.path.join(HERE, '..', '..'))
CONTENT = os.path.join(HERE, 'phase0_content.json')
LIST = os.path.join(ROOT, 'ui', 'enhancement_shaman', 'bis', 'level60_phase0.bis.json')
WAGO = os.path.expanduser('~/forever-logs/wago')

# The two lists whose two picks share one list (docs/bis_lists.md).
TWO_SLOTS = {'finger', 'trinket'}
# Avail values we don't price: we can't get the item, or not before the raids.
SKIP_AVAIL = {'doubtful', 'out'}
# A material's SpellEffect "create item" effect.
CREATE_ITEM = '24'

# A dungeon's typical clear time at level 60 in minutes, by its BayovWikiDungeons key, and by
# wing where a wing is its own run. They are estimates of a group that knows the dungeon, for
# the boss in question rather than a full clear.
#
# We checked them against the user's own runs in the Forever Dungeon Journal's run log (Bayov-Oom,
# November 2026): Dire Maul East took 13 and 19 min, North 29, Stratholme's Service Entrance 31
# to 36, Main Gate 44, Lower Blackrock Spire 41, Alcaz Prison 18 and 28, and Shaper's Terrace 68.
# Its Scholomance and Blackrock Depths runs (an hour to almost four) were while leveling, with
# quests and wipes, so we don't take those.
RUN_MINUTES = {
    'DireMaul': {'East': 20, 'West': 35, 'North': 30},
    'Stratholme': {'Main Gate': 45, 'Service Entrance': 35},
    'BlackrockDepths': {'Detention Block': 20, 'Ring of Law': 30, 'Shadowforge City': 75},
    'Scholomance': 75,
    'LowerBlackrockSpire': 45,
    'UpperBlackrockSpire': 60,
    'BlackmawHold': 45,
    'ShapersTerrace': 70,
    'AlcazPrison': 25,
    'SunkenTemple': 60,
    'Maraudon': 60,
    'ZulFarrak': 45,
    'Uldaman': 50,
}
# A dungeon RUN_MINUTES doesn't have.
DEFAULT_RUN_MINUTES = 60


# How much newer another account's prices have to be before we say so.
MUCH_NEWER = 6 * 3600


def account_of(path):
    return os.path.basename(os.path.dirname(os.path.dirname(path)))


def newest_ah_file(wow):
    """The main's newest BayovAuctionData.lua, or another account's when the main has none.

    Two accounts of ours play the same client, often at once, and the main's is
    the one that ends in #1. We say so when another one's prices are much newer.
    """
    paths = glob.glob(os.path.join(wow, '*', 'WTF', 'Account', '*', 'SavedVariables', 'BayovAuctionData.lua'))
    if not paths:
        return None
    ours = max([p for p in paths if account_of(p).endswith('#1')] or paths, key=os.path.getmtime)
    newest = max(paths, key=os.path.getmtime)
    if os.path.getmtime(newest) - os.path.getmtime(ours) > MUCH_NEWER:
        print(f"note: {account_of(newest)} has newer prices than {account_of(ours)}'s, which we use. "
              f'--ah reads them: {newest}')
    return ours


def middle(values):
    return statistics.median(values) if values else None


def parse_price(record):
    """An item's price in gold out of BayovCore's string (AuctionDB.lua, WHAT WE KEEP).

    The newest day's price (days come newest first), or the last look's when no
    day has one. A day still in progress keeps its looks as "p/p/p", and we take
    their middle.
    """
    parts = record.split(';')
    for day in parts[1:]:
        f = day.split(',')
        if len(f) > 1 and f[1]:
            return middle([int(p) for p in f[1].split('/')]) / 10000
    head = parts[0].split(',')
    for p in (head[1:2] + head[3:4]):
        if p:
            return int(p) / 10000
    return None


def load_prices(path):
    """{item ID: gold} from a BayovAuctionData.lua, over every realm in it.

    Random suffix keys ("id:suffix") give the item the cheapest of its suffixes.
    """
    with open(path, encoding='utf-8', errors='replace') as f:
        text = f.read()
    prices = {}
    for key, record in re.findall(r'\["(\d+(?::\d+)?)"\] = "([^"]*)"', text):
        gold = parse_price(record)
        if gold is None:
            continue
        iid = int(key.split(':')[0])
        if iid not in prices or gold < prices[iid]:
            prices[iid] = gold
    return prices


def latest_wago(base):
    builds = [d for d in glob.glob(os.path.join(base, '*'))
              if os.path.exists(os.path.join(d, 'SpellEffect.csv')) and os.path.exists(os.path.join(d, 'SpellReagents.csv'))]
    return max(builds, key=lambda d: [int(x) for x in re.findall(r'\d+', os.path.basename(d))]) if builds else None


def load_recipes(wago):
    """{item ID: [(material ID, count)]} for every item a spell creates."""
    creates = {}
    with open(os.path.join(wago, 'SpellEffect.csv'), encoding='utf-8') as f:
        for row in csv.DictReader(f):
            if row['Effect'] == CREATE_ITEM and row['EffectItemType'] != '0':
                creates.setdefault(int(row['SpellID']), int(row['EffectItemType']))
    recipes = {}
    with open(os.path.join(wago, 'SpellReagents.csv'), encoding='utf-8') as f:
        for row in csv.DictReader(f):
            item = creates.get(int(row['SpellID']))
            if item is None:
                continue
            mats = [(int(row[f'Reagent_{i}']), int(row[f'ReagentCount_{i}'])) for i in range(8)
                    if row.get(f'Reagent_{i}', '0') not in ('0', '')]
            if mats:
                recipes.setdefault(item, mats)
    return recipes


def newest_file(wow, pattern):
    """The newest file matching pattern under the WoW folder, in BayovUI's source or the game's AddOns."""
    paths = (glob.glob(os.path.join(wow, 'bayovui', 'addons', pattern))
             + glob.glob(os.path.join(wow, '*', 'Interface', 'AddOns', pattern)))
    return max(paths, key=os.path.getmtime) if paths else None


def load_drops(paths):
    """{item ID: [(dungeon key, npc or None, chance)]} from BayovWikiGear's Data*.lua files.

    Only items we'd farm: ones a dungeon drops that aren't also a quest reward,
    a vendor item or a reputation reward (WikiGear.lua's header has the format).
    """
    drops = {}
    for path in paths:
        with open(path, encoding='utf-8', errors='replace') as f:
            text = f.read()
        for iid, record in re.findall(r'G\[(\d+)\] = "([^"]*)"', text):
            fields = record.split('|')
            kinds, sources = fields[0], fields[-1]
            if 'd' not in kinds or set(kinds) & set('qvr'):
                continue
            for s in sources.split(';'):
                f = s.split(':')
                if f[0] == 'd' and len(f) == 4:
                    drops.setdefault(int(iid), []).append((f[1], int(f[2]) if f[2] else None, float(f[3] or 0)))
    return drops


def load_bosses(path):
    """{npc or chest ID: (dungeon key, wing or None)} from BayovWikiDungeons' Data.lua."""
    with open(path, encoding='utf-8', errors='replace') as f:
        text = f.read()
    bosses = {}
    for key, record in re.findall(r'^    (\w+) = \[\[(\{ name = .*?)\]\],?$', text, re.M):
        wings = re.search(r'wings = \{ ([^}]*) \}', record)
        wings = re.findall(r'"([^"]*)"', wings.group(1)) if wings else []
        # A boss's own fields come before its first nested table, so [^{}]* stays on that boss.
        for m in re.finditer(r'\{ (?:npc = (\d+), )?name = "(?:[^"\\]|\\.)*"([^{}]*)', record):
            chest = re.search(r'\bchest = (\d+)', m.group(2))
            npc = m.group(1) or (chest and chest.group(1))
            wing = re.search(r'\bwing = (\d+)', m.group(2))
            if npc:
                bosses.setdefault(int(npc), (key, wings[int(wing.group(1)) - 1] if wing and wings else None))
    return bosses


def load_droprates(path):
    """{item ID: {npc: chance}} from AtlasLoot Classic's droprate.lua, its data part only.

    The lines after the data change some rates for Classic's early phases, and we
    keep the final ones, like BayovUI's wiki.py does.
    """
    with open(path, encoding='utf-8', errors='replace') as f:
        text = f.read()
    start, end = text.find('--##START-DATA##'), text.find('--##END-DATA##')
    rates = {}
    for npc, body in re.findall(r'\[(\d+)\] = \{([^{}]*)\}', text[start:end]):
        for iid, rate in re.findall(r'\[(\d+)\] = ([\d.]+)', body):
            rates.setdefault(int(iid), {})[int(npc)] = float(rate)
    return rates


def drop_sources(iid, drops, bosses, rates):
    """[(dungeon key, wing, chance)] an item can be farmed from.

    The wiki's chance when it knows one. Else AtlasLoot's, from the bosses it
    puts the item on in the same dungeon (it has Treant's Bane on Cho'Rush the
    Observer, where the wiki has the Gordok Tribute chest with no chance). A
    boss the wiki doesn't know counts when the item drops in one dungeon only.
    """
    wiki = drops.get(iid, [])
    found = {npc: (key, chance) for key, npc, chance in wiki if npc and chance > 0}
    keys = {key for key, _, _ in wiki}
    for npc, rate in rates.get(iid, {}).items():
        key = bosses.get(npc, (next(iter(keys)) if len(keys) == 1 else None, None))[0]
        if npc not in found and key in keys and rate > 0:
            found[npc] = (key, rate)
    return [(key, bosses.get(npc, (key, None))[1], chance) for npc, (key, chance) in found.items()]


def run_minutes(key, wing):
    m = RUN_MINUTES.get(key, DEFAULT_RUN_MINUTES)
    if isinstance(m, dict):
        m = m.get(wing) or max(m.values())
    return m


def drop_effort(sources, gold_per_hour, roll_share, run_overhead):
    """(gold, how) for farming an item from its easiest source, or None with no source.

    Runs = 1 / (chance x roll share), and each run takes the dungeon's clear time
    plus run_overhead minutes (see the docstring at the top).
    """
    best = None
    for key, wing, chance in sources:
        runs = 100 / (chance * roll_share)
        hours = runs * (run_minutes(key, wing) + run_overhead) / 60
        if best is None or hours < best[0]:
            best = (hours, runs, key, wing, chance)
    if best is None:
        return None
    hours, runs, key, wing, chance = best
    where = key + (f' {wing}' if wing else '')
    return hours * gold_per_hour, f'effort: {where} {chance:g}%, {runs:.0f} runs, {hours:.0f}h'


def item_gold(iid, prices, recipes, effort=None):
    """(gold, how) for an item, the cheapest of the auction house, its materials and
    its effort ({item ID: (gold, how)}), or (None, None)."""
    options = []
    if iid in prices:
        options.append((prices[iid], 'AH'))
    mats = recipes.get(iid)
    if mats and all(m in prices for m, _ in mats):
        options.append((sum(prices[m] * n for m, n in mats), 'materials'))
    if effort and iid in effort:
        options.append(effort[iid])
    return min(options) if options else (None, None)


def price_slot(slot, entries, content, prices, recipes, max_gpd, effort=None):
    """Price one list's entries: [{id, name, gold, how, gpd, over}] from the cheapest up."""
    rows = []
    for e in entries:
        c = content.get(str(e['id']), {})
        if c.get('avail') in SKIP_AVAIL:
            continue
        gold, how = item_gold(e['id'], prices, recipes, effort)
        rows.append({'id': e['id'], 'name': e['name'], 'loss': e['loss'], 'gold': gold, 'how': how,
                     'gpd': None, 'over': None, 'pick': bool(e.get('bis')),
                     # An item infeasible for its reputation stays so whatever it costs.
                     'rep': c.get('avail') == 'infeasible'})
    # Same item twice (a ring with two suffixes): keep the better.
    feasible = []
    for r in sorted(rows, key=lambda r: (r['gold'] or 0, r['loss'])):
        cheaper = sorted((f for f in feasible if (f['gold'] or 0) < (r['gold'] or 0)), key=lambda f: f['loss'])
        k = 1 if slot in TWO_SLOTS else 0
        base = cheaper[k] if len(cheaper) > k else (cheaper[-1] if cheaper else None)
        if r['gold'] and base is not None:
            gain = base['loss'] - r['loss']
            if gain > 0:
                r['gpd'] = round((r['gold'] - (base['gold'] or 0)) / gain, 1)
                r['over'] = base['name']
        if not r['rep'] and (r['gpd'] is None or r['gpd'] <= max_gpd):
            feasible.append(r)
    return rows


def main():
    ap = argparse.ArgumentParser(description=__doc__.split('\n')[0])
    ap.add_argument('--write', action='store_true', help='save gold and goldPerDps in phase0_content.json')
    ap.add_argument('--ah', help="a BayovAuctionData.lua (default: the main account's newest under --wow)")
    ap.add_argument('--wow', default=os.environ.get('WOW_DIR', '/mnt/d/Gaming/World of Warcraft'))
    ap.add_argument('--wago', help='a wago table folder with SpellEffect.csv and SpellReagents.csv (default: the newest build)')
    ap.add_argument('--list', default=LIST)
    ap.add_argument('--max', type=float, default=float(os.environ.get('MAX_GOLD_PER_DPS', 30)),
                    help='the line, for the table and the feasible chain (bis_phase0.sh has its own)')
    ap.add_argument('--gold-per-hour', type=float, default=30, help='what an hour of farming is worth')
    ap.add_argument('--roll-share', type=float, default=0.5, help='our share of a drop, 0.5 when one other player needs it')
    ap.add_argument('--run-overhead', type=float, default=20,
                    help="minutes a run takes on top of the clear, for getting there and finding a group")
    args = ap.parse_args()

    ah = args.ah or newest_ah_file(args.wow)
    wago = args.wago or latest_wago(WAGO)
    if not ah or not wago:
        sys.exit(f'no auction data ({ah}) or wago tables ({wago})')
    prices, recipes = load_prices(ah), load_recipes(wago)
    with open(args.list) as f:
        bis = json.load(f)
    with open(CONTENT) as f:
        content = json.load(f)

    gear = sorted(set(glob.glob(os.path.join(os.path.dirname(newest_file(args.wow, 'BayovWikiGear/Data.lua') or ''),
                                             'Data*.lua'))))
    dungeons = newest_file(args.wow, 'BayovWikiDungeons/Data.lua')
    atlas = newest_file(args.wow, 'AtlasLootClassic_DungeonsAndRaids/droprate.lua')
    drops = load_drops(gear) if gear else {}
    bosses = load_bosses(dungeons) if dungeons else {}
    rates = load_droprates(atlas) if atlas else {}
    effort, unknown = {}, set()
    for entries in bis['slots'].values():
        for e in entries:
            if e['id'] not in drops:
                continue
            r = drop_effort(drop_sources(e['id'], drops, bosses, rates),
                            args.gold_per_hour, args.roll_share, args.run_overhead)
            if r:
                effort[e['id']] = r
            elif content.get(str(e['id']), {}).get('avail') not in SKIP_AVAIL:
                unknown.add(e['name'])

    print(f'prices from {ah}, recipes from {wago}, line {args.max:g}g per DPS')
    print(f'drops from {gear[0] if gear else None}, wings from {dungeons}, AtlasLoot rates from {atlas}')
    print(f'effort at {args.gold_per_hour:g}g per hour, roll share {args.roll_share:g}, run overhead {args.run_overhead:g} min')
    priced = {}
    for slot, entries in bis['slots'].items():
        for r in price_slot(slot, entries, content, prices, recipes, args.max, effort):
            if not r['gold']:
                continue
            priced[str(r['id'])] = r
            over = r['gpd'] is not None and r['gpd'] > args.max
            if r['gpd'] is not None:
                print(f"{slot:9s} {r['name'][:30]:30s} {r['gold']:8.1f}g {r['gpd']:8.1f}g/DPS over "
                      f"{r['over'][:28]:28s}{' PICK' if r['pick'] else '     '}{' OUT' if over else '    '}  {r['how']}")
    if unknown:
        print(f"drops with no known chance, so no effort price: {', '.join(sorted(unknown))}")

    if not args.write:
        return
    priced = {iid: r for iid, r in priced.items() if r['gpd'] is not None}
    changed = 0
    for iid, c in content.items():
        r = priced.get(iid)
        new = {'gold': round(r['gold'], 1), 'goldPerDps': r['gpd']} if r else {}
        old = {k: c[k] for k in ('gold', 'goldPerDps') if k in c}
        if old != new:
            changed += 1
            for k in ('gold', 'goldPerDps'):
                c.pop(k, None)
            c.update(new)
    # A priced list item the content check hasn't seen yet gets an entry of its own.
    for iid, r in priced.items():
        if iid not in content:
            content[iid] = {'name': r['name'], 'gold': round(r['gold'], 1), 'goldPerDps': r['gpd']}
            changed += 1
    with open(CONTENT, 'w') as f:
        f.write('{\n')
        keys = sorted(content, key=int)
        for i, k in enumerate(keys):
            f.write(f'  {json.dumps(k)}: {json.dumps(content[k], ensure_ascii=False)}{"," if i < len(keys) - 1 else ""}\n')
        f.write('}\n')
    print(f'wrote {CONTENT}, {changed} entries changed')


if __name__ == '__main__':
    main()
