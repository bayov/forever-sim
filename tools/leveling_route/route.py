#!/usr/bin/env python3
# /// script
# dependencies = ["curl_cffi", "websocket-client"]
# ///
# Turn our RestedXP route into the quest rewards and dungeon runs the leveling gear plan takes.
#
# The route is the "Bayov On-Release Guide" that BayovCore registers in RestedXP, cut into
# sections in bayovui/addons/BayovCore/RXPGuides. We walk it as one character (an Orc Shaman
# by default) and keep each quest we turn in with the level we're at then, and each dungeon we
# run. A quest row on wowhead says which items the quest gives, as a choice or all of them, and
# how much XP, so the quest listing comes from wowhead (through the page cache wowhead.py
# keeps).
#
#   uv run tools/leveling_route/route.py [guides dir] > ui/enhancement_shaman/leveling_route.json
#
# The output holds only ids and levels, not the guide's text, because levels 22 to 60 come
# from RestedXP's paid guides and this repo is public.
#
# The level comes from each quest's XP (less for a quest far below us, like the game does),
# and the guide's "grind to" steps, where we're at least at their level. Each section's name
# says the level the guide expects at its start, and we're at least there too. We don't count
# the XP of the mobs we kill on the way, so the level runs up to a level behind the guide's.
import json, os, re, sys

sys.path.insert(0, os.path.join(os.path.dirname(os.path.abspath(__file__)), '../database/forever_wowhead'))
from scrape import scrape_quests
from wowhead import Wowhead

ROOT = os.path.join(os.path.dirname(os.path.abspath(__file__)), '../..')
GUIDES = '/mnt/d/Gaming/World of Warcraft/bayovui/addons/BayovCore/RXPGuides'
# Who walks the route, as the RestedXP step filters name it.
ME = {'orc', 'shaman', 'horde'}
# The dungeons BayovCore's Modules/RXPGuides.lua turns on and off for our route.
DUNGEON_TAGS = {'RFC': True, 'ROL': True, 'WC': False, 'DM': False, 'SM': True, 'RFD': True, 'ULDA': True,
                'ZF': True, 'MARA': True, 'ST': True, 'BRD': True}
# The places a guide sends us into a dungeon, and the dungeon's zone. Most are the dungeon's own
# zone, but Zul'Farrak's and Blackfathom Deeps' steps name the place outside.
DUNGEON_SUBZONES = {2437: 2437, 16611: 16611, 209: 209, 2797: 719, 491: 491, 721: 721, 722: 722, 796: 796,
                    1337: 1337, 978: 1176, 2100: 2100, 1477: 1477, 1584: 1584, 16732: 16732, 16560: 16560}
# Wailing Caverns' section has no step that names its zone.
DUNGEON_SECTIONS = {'Wailing Caverns': 718}
DUNGEON_NAMES = {2437: 'Ragefire Chasm', 16611: 'Ruins of Lordaeron', 718: 'Wailing Caverns', 209: 'Shadowfang Keep',
                 719: 'Blackfathom Deeps', 491: 'Razorfen Kraul', 721: 'Gnomeregan', 722: 'Razorfen Downs',
                 796: 'Scarlet Monastery', 1337: 'Uldaman', 1176: "Zul'Farrak", 2100: 'Maraudon',
                 1477: "The Temple of Atal'Hakkar", 1584: 'Blackrock Depths', 16732: 'Excavation Site: Wetlands',
                 16560: 'City of Dalaran'}
# The XP to the next level, by level.
XP_NEXT = [0, 400, 900, 1400, 2100, 2800, 3600, 4500, 5400, 6500, 7600, 8800, 10100, 11400, 12900, 14400,
           16000, 17700, 19400, 21300, 23200, 25200, 27300, 29400, 31700, 34000, 36400, 38900, 41400, 44300,
           47400, 50800, 54500, 58600, 62800, 67100, 71600, 76100, 80800, 85700, 90700, 95800, 101000,
           106300, 111800, 117500, 123200, 129100, 135100, 141200, 147500, 153900, 160400, 167100, 173900,
           180800, 187900, 195000, 202300, 209800]


def applies(expr):
    """Whether a RestedXP '<< ...' filter is true for us. A '/' is or, a space is and, '!' is not."""
    for alt in expr.split('/'):
        if all((tok.lstrip('!').lower() in ME) != tok.startswith('!') for tok in alt.split()):
            return True
    return False


def split_filter(line):
    if ' << ' in line:
        body, expr = line.rsplit(' << ', 1)
        return body, expr
    return line, None


def sections(guides):
    """The route's sections in order, by their level in the file name."""
    for f in sorted(os.listdir(guides)):
        if f.endswith('.lua'):
            body = open(os.path.join(guides, f)).read().split('[==[', 1)[1].split(']==]', 1)[0]
            yield re.search(r'^#name (.+)$', body, re.M).group(1).strip(), body


def steps(body):
    step = None
    for raw in body.split('\n'):
        line = raw.strip()
        if line.startswith('step'):
            if step:
                yield step
            step = {'filter': split_filter(line)[1], 'lines': []}
        elif step is not None and line and not line.startswith('--'):
            step['lines'].append(line)
    if step:
        yield step


def skipped(flags, lines, reached):
    """Whether we skip a step: another class's or race's, hardcore only, Season of Discovery,
    for a faster XP rate, a dungeon we don't run, or for other levels than ours. `reached`
    says whether we're at a (level, xp) yet."""
    for f in flags:
        parts = f.split()
        if parts[0] in ('#hardcore', '#hardcoreserver'):
            return True
        if parts[0] == '#season' and parts[1:2] == ['2']:
            return True
        if parts[0] == '#xprate' and len(parts) > 1:
            rate = float(parts[1][1:])
            if (parts[1][0] == '<') != (1.0 < rate):
                return True
    for l in lines:
        m = re.match(r'\.dungeon (!?)(\w+)', l)
        if m and DUNGEON_TAGS.get(m.group(2), True) == bool(m.group(1)):
            return True
        # A level check (".xp <16,1" or ".xp >18,1", with the step to skip to after the comma)
        # isn't a grind target. "<16" skips the step below level 16, and ">18" (or a bare
        # "18") skips it from 18 on, the way RestedXP's functions.xp reads it with its
        # default "skip steps by level" setting on.
        m = re.match(r'\.xp ([<>]?)(\d+)(?:\+(\d+))?,', l)
        if m and reached((int(m.group(2)), int(m.group(3) or 0))) != (m.group(1) == '<'):
            return True
    return False


def main():
    guides = sys.argv[1] if len(sys.argv) > 1 else GUIDES
    quests = scrape_quests(Wowhead())
    items = {i['id']: i for i in json.load(open(os.path.join(ROOT, 'assets/database/db.json')))['items']}
    level, xp = 1, 0
    turnins, runs = [], {}

    def reach(lv, plus=0):
        nonlocal level, xp
        if (level, xp) < (lv, plus):
            level, xp = lv, plus

    for name, body in sections(guides):
        m = re.match(r'(\d+)-', name)
        if m:
            reach(int(m.group(1)))
        for zone_name, zone in DUNGEON_SECTIONS.items():
            if zone_name in name:
                runs.setdefault(zone, level)
        section = list(steps(body))
        for i, step in enumerate(section):
            if step['filter'] and not applies(step['filter']):
                continue
            # A guide sometimes sends us into a dungeon only to die there and get back to town
            # (Ragefire Chasm at 11), and that's no run.
            deathskip = i + 1 < len(section) and any(l.startswith('.deathskip') for l in section[i + 1]['lines'])
            flags = [l for l in step['lines'] if l.startswith('#')]
            lines = [body for body, expr in map(split_filter, step['lines']) if expr is None or applies(expr)]
            if skipped(flags, lines, lambda at: (level, xp) >= at):
                continue
            for l in lines:
                m = re.match(r'\.xp (\d+)(?:\+(\d+))?(?![,\d])', l)
                if m:
                    reach(int(m.group(1)), int(m.group(2) or 0))
                m = re.match(r'\.subzone (\d+)', l)
                if m and int(m.group(1)) in DUNGEON_SUBZONES and not deathskip:
                    runs.setdefault(DUNGEON_SUBZONES[int(m.group(1))], level)
                m = re.match(r'\.turnin (\d+)', l)
                if m:
                    q = quests.get(int(m.group(1)))
                    if q is None:
                        print('no wowhead row for quest', m.group(1), file=sys.stderr)
                        continue
                    if any(t['id'] == q['id'] for t in turnins):
                        continue
                    below = level - q['level']
                    share = 1 if below <= 5 else max(0.1, 1 - 0.2 * (below - 5))
                    xp += round(q['xp'] * share)
                    while level < 60 and xp >= XP_NEXT[level]:
                        xp -= XP_NEXT[level]
                        level += 1
                    turnins.append({'id': q['id'], 'name': q['name'], 'level': level,
                                    'choiceIds': [c[0] for c in q['itemchoices'] if c[0] in items],
                                    'rewardIds': [r[0] for r in q['itemrewards'] if r[0] in items]})

    out = {
        'name': 'Bayov On-Release Guide (Orc Shaman)',
        'quests': [t for t in turnins if t['choiceIds'] or t['rewardIds']],
        'dungeons': [{'zoneId': zone, 'name': DUNGEON_NAMES[zone], 'level': lv} for zone, lv in sorted(runs.items(), key=lambda r: r[1])],
    }
    print(f"{len(turnins)} quests, {len(out['quests'])} with items, {len(out['dungeons'])} dungeons", file=sys.stderr)
    json.dump(out, sys.stdout, indent='\t')
    print()


main()
