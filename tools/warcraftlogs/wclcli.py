"""Mine public Warcraft Logs reports for proc cooldowns, proc rates, glancing blows and what
bosses do to their tanks.

Usage: python3 tools/warcraftlogs/wclcli.py [--site sod] [--budget 300] <command> ...
See README.md in this directory for the workflow.
"""
import argparse, bisect, calendar, collections, json, math, os, sys, time

sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
import wcl
import wclraid

DB_PATH = os.path.join(os.path.dirname(os.path.abspath(__file__)), '../../assets/database/db.json')

# hitType codes in Warcraft Logs damage events.
MISS, HIT, CRIT, BLOCK, CRIT_BLOCK, GLANCE, DODGE, PARRY, CRUSH = 0, 1, 2, 4, 5, 6, 7, 8, 15
HIT_NAMES = {0: 'miss', 1: 'hit', 2: 'crit', 3: 'absorb', 4: 'block', 5: 'crit block',
             6: 'glance', 7: 'dodge', 8: 'parry', 10: 'immune', 14: 'resist', 15: 'crush',
             16: 'partial resist', 17: 'crit partial resist'}
LANDED = {HIT, CRIT, BLOCK, CRIT_BLOCK, GLANCE, CRUSH, 16, 17}
WHITE = 1  # "Melee", every white swing of either hand.
PHYSICAL = '1'  # The school of an ability in a report's masterData.
PAIR_MS = 5

# Gear slot indexes in a combatantinfo event.
MAIN_HAND, OFF_HAND = 15, 16

BUFF_TYPES = {'applybuff', 'applybuffstack', 'refreshbuff', 'removebuff', 'removebuffstack',
              'applydebuff', 'applydebuffstack', 'refreshdebuff', 'removedebuff',
              'removedebuffstack'}


def ids(s):
    return {int(x) for x in s.split(',') if x}


def link(code, fight):
    url = f'https://{wcl.SITE}.warcraftlogs.com/reports/{code}'
    return url if fight is None else f'{url}#fight={fight}'


def ms(t):
    return f'{t / 1000:.3f}'


def interval(k, n, z=1.96):
    """Wilson 95% interval of a rate, in percent."""
    if n == 0 or k > n:
        return '-'
    p = k / n
    d = 1 + z * z / n
    c = (p + z * z / (2 * n)) / d
    h = z * math.sqrt(p * (1 - p) / n + z * z / (4 * n * n)) / d
    return f'{100 * (c - h):.1f}-{100 * (c + h):.1f}'


# Samples. A sample file is a JSON list of fights to read, one entry per player:
# {"site", "code", "fight" (null for every fight in the report), "player" (actor id), "name",
#  "class", "spec"}.

def load_sample(path):
    sample = json.load(open(path))
    sites = {s['site'] for s in sample}
    if len(sites) > 1:
        wcl.die(f'{path} mixes sites {sites}')
    if sites:
        wcl.SITE = sites.pop()
    return sample


def cmd_zones(a):
    d = wcl.query('{ worldData { zones { id name frozen expansion { name } '
                  'encounters { id name } partitions { id name default } } } }', ttl=wcl.DAY)
    for z in d['worldData']['zones']:
        parts = ', '.join(f'{p["id"]}={p["name"]}{"*" if p["default"] else ""}'
                          for p in z['partitions'] or [])
        print(f'zone {z["id"]} {z["name"]} ({z["expansion"]["name"]}) partitions: {parts or "-"}')
        for e in z['encounters']:
            print(f'    {e["id"]} {e["name"]}')


def actor_by_name(meta, name):
    for x in meta['masterData']['actors']:
        if x['type'] == 'Player' and x['name'] == name:
            return x
    return None


def cmd_sample(a):
    out = []
    seen = set()
    if a.encounter:
        for enc in a.encounter:
            got = 0
            for page in range(1, a.pages + 1):
                args = {'e': enc, 'c': a.cls, 's': a.spec, 'p': page, 'part': a.partition,
                        'd': a.difficulty}
                r = wcl.query('''query($e: Int!, $c: String, $s: String, $p: Int, $part: Int, $d: Int) {
                    worldData { encounter(id: $e) { characterRankings(className: $c, specName: $s,
                        page: $p, partition: $part, difficulty: $d) } } }''',
                              args, ttl=wcl.DAY)['worldData']['encounter']['characterRankings']
                for rk in r['rankings']:
                    if got >= a.per_encounter:
                        break
                    code, fight = rk['report']['code'], rk['report']['fightID']
                    if (code, fight, rk['name']) in seen:
                        continue
                    actor = actor_by_name(wcl.report_meta(code), rk['name'])
                    if not actor:
                        continue
                    seen.add((code, fight, rk['name']))
                    out.append({'site': wcl.SITE, 'code': code, 'fight': fight,
                                'player': actor['id'], 'name': rk['name'], 'class': rk['class'],
                                'spec': rk['spec']})
                    got += 1
                if got >= a.per_encounter or not r['hasMorePages']:
                    break
            print(f'encounter {enc}: {got} fights', file=sys.stderr)
    if a.zone:
        # Public reports of a zone, newest first. This also finds trash and dungeon runs,
        # which matter when a new game has no rankings yet.
        # --since and --until pick a window, newest first within it: the Anniversary realms'
        # Molten Core of its own phase rather than this year's runs by higher level characters.
        reports = []
        page = 1
        window = {}
        if a.since:
            window['s'] = calendar.timegm(time.strptime(a.since, '%Y-%m-%d')) * 1000
        if a.until:
            window['e'] = calendar.timegm(time.strptime(a.until, '%Y-%m-%d')) * 1000
        while len(reports) < a.reports:
            r = wcl.query('''query($z: Int, $p: Int''' + (', $s: Float' if 's' in window else '') + (', $e: Float' if 'e' in window else '') + ''') { reportData {
                reports(zoneID: $z, page: $p, limit: 100''' + (', startTime: $s' if 's' in window else '') + (', endTime: $e' if 'e' in window else '') + ''') {
                    data { code guild { server { slug } } } has_more_pages } } }''',
                          {'z': a.zone, 'p': page, **window}, ttl=wcl.DAY)['reportData']['reports']
            # --server keeps the reports of guilds on those realms, --exclude-server drops them:
            # the vanilla site has the Anniversary realms and Classic Era together.
            for x in r['data']:
                slug = ((x.get('guild') or {}).get('server') or {}).get('slug')
                if a.server and slug not in a.server:
                    continue
                if a.exclude_server and (slug is None or slug in a.exclude_server):
                    continue
                reports.append(x['code'])
            # The API lists 25 pages at most: a busy zone needs a narrower window.
            if not r['has_more_pages'] or page == 25:
                break
            page += 1
        for code in reports[:a.reports]:
            meta = wcl.report_meta(code)
            for x in meta['masterData']['actors']:
                if x['type'] == 'Player' and x['subType'] == a.cls:
                    out.append({'site': wcl.SITE, 'code': code, 'fight': None, 'player': x['id'],
                                'name': x['name'], 'class': x['subType'], 'spec': None})
        print(f'zone {a.zone}: {len(reports[:a.reports])} reports', file=sys.stderr)
    for code_fight in a.report or []:
        code, _, rest = code_fight.partition(':')
        fight, _, name = rest.partition(':')
        actor = actor_by_name(wcl.report_meta(code), name)
        if not actor:
            wcl.die(f'no player {name} in {code}')
        out.append({'site': wcl.SITE, 'code': code, 'fight': int(fight) if fight else None,
                    'player': actor['id'], 'name': name, 'class': actor['subType'], 'spec': None})
    if os.path.exists(a.out) and a.append:
        old = json.load(open(a.out))
        keys = {(s['code'], s['fight'], s['player']) for s in old}
        out = old + [s for s in out if (s['code'], s['fight'], s['player']) not in keys]
    json.dump(out, open(a.out, 'w'), indent=1)
    print(f'{len(out)} entries in {a.out}', file=sys.stderr)


def fight_ranges(meta, entry):
    """{fight id: (start, end)} for the fights an entry covers."""
    return {f['id']: (f['startTime'], f['endTime']) for f in meta['fights']
            if entry['fight'] is None or f['id'] == entry['fight']}


def entry_events(entry):
    """Events of one sample entry, split by fight. None when we stop early on the budget."""
    meta = wcl.report_meta(entry['code'])
    evs = wcl.player_events(entry['code'], entry['fight'], entry['player'])
    by_fight = collections.defaultdict(list)
    ranges = fight_ranges(meta, entry)
    for e in evs:
        if e.get('fight') in ranges:
            by_fight[e['fight']].append(e)
    return meta, ranges, by_fight


def each_entry(sample, a):
    """Yield (entry, meta, fight ranges, events by fight) and stop cleanly on the budget."""
    done = 0
    for entry in sample[:a.limit]:
        try:
            meta, ranges, by_fight = entry_events(entry)
        except wcl.BudgetError as err:
            print(f'stopped after {done} of {len(sample[:a.limit])} entries: {err}', file=sys.stderr)
            return
        done += 1
        yield entry, meta, ranges, by_fight


def ability_names(meta):
    return {x['gameID']: x['name'] for x in meta['masterData']['abilities']}


def cmd_abilities(a):
    """Every ability a player shows up with, by event type, to find a game's spell IDs."""
    code, _, rest = a.target.partition(':')
    fight, _, name = rest.partition(':')
    meta = wcl.report_meta(code)
    actor = actor_by_name(meta, name)
    if not actor:
        wcl.die(f'no player {name} in {code}. Players: ' + ', '.join(
            sorted(x['name'] for x in meta['masterData']['actors'] if x['type'] == 'Player')))
    names = ability_names(meta)
    evs = wcl.player_events(code, int(fight) if fight else None, actor['id'])
    c = collections.Counter()
    for e in evs:
        if 'abilityGameID' not in e:
            continue
        who = 'own' if e['sourceID'] == actor['id'] else 'from others'
        c[(e['abilityGameID'], e['type'], who)] += 1
    print(f'{name} ({actor["subType"]}) in {code} fight {fight or "all"}: {len(evs)} events')
    for (ab, typ, who), n in sorted(c.items(), key=lambda x: (x[0][0], x[0][1])):
        print(f'{ab:>8} {names.get(ab, "?"):<32} {typ:<18} {who:<12} {n}')


def proc_finder(a):
    """Returns procs(events, player) -> [(time, events in the proc)] for the --ability procs."""
    want = ids(a.ability)
    types = set(a.type.split(','))

    def match(e, player):
        if e.get('abilityGameID') not in want or e['type'] not in types:
            return False
        # A buff counts when it lands on the player. Everything else counts when the player
        # did it, so a shaman's totem buff on a warrior counts for the warrior only.
        if e['type'] in BUFF_TYPES:
            return e.get('targetID') == player
        return e['sourceID'] == player

    def procs(evs, player):
        hits = [e for e in evs if match(e, player)]
        # Buff procs are counted event by event, because two procs can land on the same ms
        # (SoD's two Windfury attacks can each add a Maelstrom Weapon stack). A stack added
        # logs applybuffstack and then refreshbuff 0 or 1 ms later, and a proc at full stacks
        # logs only refreshbuff. So a refreshbuff up to PAIR_MS after an unpaired stack event
        # is the same proc, and we drop it.
        buff = []
        open_stacks = collections.defaultdict(list)
        for e in hits:
            if e['type'] not in BUFF_TYPES:
                continue
            k = e['abilityGameID']
            if e['type'] in ('refreshbuff', 'refreshdebuff'):
                pending = [t for t in open_stacks[k] if e['timestamp'] - t <= PAIR_MS]
                if pending:
                    open_stacks[k] = pending[1:]
                    continue
            if e['type'] in ('applybuffstack', 'applydebuffstack'):
                open_stacks[k].append(e['timestamp'])
            buff.append([e['timestamp'], 1])
        # Other procs merge events closer than --merge ms into one proc (SoD's Windfury Weapon
        # logs two attacks at the same ms, sometimes up to 21 ms apart).
        other = []
        for e in hits:
            if e['type'] in BUFF_TYPES:
                continue
            if other and e['timestamp'] - other[-1][0] <= a.merge:
                other[-1][1] += 1
            else:
                other.append([e['timestamp'], 1])
        return sorted(buff + other)
    return procs


def cmd_gaps(a):
    sample = load_sample(a.sample)
    find_procs = proc_finder(a)
    gaps = []
    sizes = collections.Counter()
    nprocs = 0
    players = 0
    for entry, meta, ranges, by_fight in each_entry(sample, a):
        names = ability_names(meta)
        schools = {x['gameID']: x['type'] for x in meta['masterData']['abilities']}
        players += 1
        p = entry['player']
        for fight, evs in by_fight.items():
            procs = find_procs(evs, p)
            nprocs += len(procs)
            sizes.update(n for _, n in procs)
            # Only physical casts (white swings and weapon strikes) count as triggers, so a
            # Lightning Bolt cast between a swing and its proc doesn't hide the swing.
            casts = [(e['timestamp'], e['abilityGameID']) for e in evs
                     if e['type'] == 'cast' and e['sourceID'] == p
                     and e.get('abilityGameID') not in ids(a.ability)
                     and schools.get(e.get('abilityGameID')) == PHYSICAL]
            cast_times = [t for t, _ in casts]

            def trigger(t):
                # The player's last physical cast before the proc. A proc lands 20 to 70 ms
                # after the hit that caused it in SoD logs.
                i = bisect.bisect_right(cast_times, t) - 1
                if i < 0 or t - cast_times[i] > a.trigger_window:
                    return '?'
                return f'{names.get(casts[i][1], casts[i][1])} {t - cast_times[i]} ms before'
            f0 = ranges[fight][0]
            for (t1, _), (t2, _) in zip(procs, procs[1:]):
                gaps.append((t2 - t1, entry, fight, t1 - f0, t2 - f0, trigger(t1), trigger(t2)))
    print(f'{players} players, {nprocs} procs, {len(gaps)} gaps between procs in the same fight')
    print('events per proc: ' + ', '.join(f'{k}: {v}' for k, v in sorted(sizes.items())))
    if not gaps:
        return
    gaps.sort(key=lambda g: g[0])
    print(f'\nshortest gap {gaps[0][0]} ms. Gaps under each limit:')
    for lim in (100, 250, 500, 1000, 1500, 2000, 3000):
        print(f'  < {lim:>4} ms: {sum(1 for g in gaps if g[0] < lim)}')
    print(f'\nhistogram, {a.bin} ms bins up to {a.max} ms:')
    hist = collections.Counter(g[0] // a.bin for g in gaps if g[0] < a.max)
    top = max(hist.values()) if hist else 1
    for b in range(a.max // a.bin):
        n = hist.get(b, 0)
        print(f'  {b * a.bin:>5}-{(b + 1) * a.bin - 1:<5} {n:>6} {"#" * round(50 * n / top)}')
    print(f'\n{a.top} shortest gaps (time from the fight start, and what came before each proc):')
    for g, entry, fight, t1, t2, tr1, tr2 in gaps[:a.top]:
        print(f'  {g:>5} ms  {entry["name"]:<14} {ms(t1)} ({tr1}) -> {ms(t2)} ({tr2})  '
              f'{link(entry["code"], fight)}')


def cmd_glance(a):
    sample = load_sample(a.sample)
    total = collections.Counter()
    glance_dmg, hit_dmg = [], []
    rows = []
    for entry, meta, ranges, by_fight in each_entry(sample, a):
        actors = {x['id']: x for x in meta['masterData']['actors']}
        c = collections.Counter()
        for evs in by_fight.values():
            for e in evs:
                if e['type'] != 'damage' or e.get('abilityGameID') != WHITE or e['sourceID'] != entry['player']:
                    continue
                tgt = actors.get(e.get('targetID'), {})
                if a.targets == 'boss' and tgt.get('subType') != 'Boss':
                    continue
                c[e['hitType']] += 1
                if e['hitType'] == GLANCE:
                    glance_dmg.append((e['amount'], e.get('unmitigatedAmount')))
                elif e['hitType'] == HIT:
                    hit_dmg.append((e['amount'], e.get('unmitigatedAmount')))
        total.update(c)
        rows.append((entry, c))
    print(f'white swings on {"bosses" if a.targets == "boss" else "every target"}, by player:')
    for entry, c in rows:
        n = sum(c.values())
        if n:
            print(f'  {entry["name"]:<14} {n:>5} swings, {c[GLANCE]:>4} glances '
                  f'({100 * c[GLANCE] / n:.1f}%)  {link(entry["code"], entry["fight"])}')
    n = sum(total.values())
    print(f'\nall players: {n} white swings')
    for k, v in sorted(total.items()):
        print(f'  {HIT_NAMES.get(k, k):<8} {v:>6}  {100 * v / n:.2f}% (95%: {interval(v, n)})')
    if glance_dmg and hit_dmg:
        # A rough check of glancing damage. Main hand and off hand swings are mixed in both
        # groups, and so are targets with different armor. The final amount has the target's
        # armor in it. unmitigatedAmount should be the log's raw field (before armor), but in
        # SoD logs it doesn't add up on crits, so we print both.
        print(f'\nglances {len(glance_dmg)}, normal hits {len(hit_dmg)}. Average glance over average hit:')
        for i, what in ((0, 'final amount'), (1, 'unmitigatedAmount')):
            g = [x[i] for x in glance_dmg if x[i] is not None]
            h = [x[i] for x in hit_dmg if x[i] is not None]
            if g and h:
                print(f'  {what:<18} {sum(g) / len(g):>9.1f} / {sum(h) / len(h):>9.1f} = '
                      f'{sum(g) / len(g) / (sum(h) / len(h)):.3f}')


def weapon_speeds(overrides):
    try:
        db = json.load(open(DB_PATH))
        speeds = {i['id']: i['weaponSpeed'] for i in db['items'] if i.get('weaponSpeed')}
    except OSError:
        speeds = {}
    for kv in overrides or []:
        k, v = kv.split('=')
        speeds[int(k)] = float(v)
    return speeds


def cmd_procrate(a):
    sample = load_sample(a.sample)
    find_procs = proc_finder(a)
    triggers = ids(a.triggers)
    speeds = weapon_speeds(a.speed)
    tot_procs = tot_weight = tot_hits = tot_ms = 0
    missing = set()
    print('player          procs  landed hits  per hit   per min  weapons            PPM')
    for entry, meta, ranges, by_fight in each_entry(sample, a):
        p = entry['player']
        procs = hits_white = hits_yellow = 0
        gear = None
        dur = 0
        for fight, evs in by_fight.items():
            dur += ranges[fight][1] - ranges[fight][0]
            procs += len(find_procs(evs, p))
            for e in evs:
                if e['type'] == 'combatantinfo' and e['sourceID'] == p:
                    gear = e.get('gear')
                if e['type'] != 'damage' or e['sourceID'] != p or e.get('hitType') not in LANDED:
                    continue
                if e.get('abilityGameID') not in triggers:
                    continue
                if e['abilityGameID'] == WHITE:
                    hits_white += 1
                else:
                    hits_yellow += 1
        hits = hits_white + hits_yellow
        mh = oh = None
        if gear and len(gear) > OFF_HAND:
            mh, oh = gear[MAIN_HAND]['id'], gear[OFF_HAND]['id']
        smh, soh = speeds.get(mh), speeds.get(oh)
        weight = None
        if smh:
            # A PPM proc has a chance of PPM * speed / 60 on each hit. White swings of both
            # hands are mixed in the log. Each hand swings in proportion to 1 / speed, so the
            # main hand takes soh / (smh + soh) of the white hits.
            if soh:
                mh_white = hits_white * soh / (smh + soh)
                weight = mh_white * smh + hits_yellow * smh
                if not a.main_hand_only:
                    weight += (hits_white - mh_white) * soh
            else:
                weight = hits * smh
        else:
            missing.add(mh)
        ppm = f'{procs * 60 / weight:.2f}' if weight else '?'
        wtxt = f'{mh}:{smh or "?"}' + (f' {oh}:{soh}' if soh else '')
        print(f'{entry["name"]:<14} {procs:>6} {hits:>12} {procs / hits if hits else 0:>8.3f} '
              f'{procs * 60000 / dur if dur else 0:>9.2f}  {wtxt:<18} {ppm}')
        tot_procs += procs
        tot_hits += hits
        tot_ms += dur
        if weight:
            tot_weight += weight
    print(f'\nall: {tot_procs} procs on {tot_hits} landed hits ({interval(tot_procs, tot_hits)}% a hit), '
          f'{tot_procs * 60000 / tot_ms if tot_ms else 0:.2f} a minute of fight')
    if tot_weight:
        print(f'PPM over the players with known weapon speeds: {tot_procs * 60 / tot_weight:.2f}')
    if missing - {None}:
        print(f'no weapon speed for items {sorted(missing - {None})}. Pass --speed ID=SPEED.')


# Tank profiles: what a boss does to its main tank and what the raid heals it for.

# Two boss swings on the main tank this far apart or less are one stretch of tanking. Longer gaps
# are a tank swap, an air phase or the boss on someone else, and don't count as tanking time.
TANKING_GAP_MS = 6000


def median(xs):
    xs = sorted(xs)
    if not xs:
        return 0
    m = len(xs) // 2
    return xs[m] if len(xs) % 2 else (xs[m - 1] + xs[m]) / 2


def pctl(xs, p):
    """The p-th percentile (0 to 100) of xs, nearest rank."""
    xs = sorted(xs)
    if not xs:
        return 0
    return xs[min(len(xs) - 1, max(0, round(p / 100 * (len(xs) - 1))))]


def tanking_windows(swing_times):
    """The stretches the boss swung at the tank: [(start, end)] in ms, each swing's gap included."""
    out = []
    for t in sorted(swing_times):
        if out and t - out[-1][1] <= TANKING_GAP_MS:
            out[-1][1] = t
        else:
            out.append([t, t])
    # A stretch runs from its first swing to one median gap past its last.
    gaps = [y - x for x, y in zip(sorted(swing_times), sorted(swing_times)[1:]) if y - x <= TANKING_GAP_MS]
    pad = median(gaps) if gaps else 2000
    return [(s, e + pad) for s, e in out]


def inside(t, windows):
    return any(s <= t < e for s, e in windows)


def tank_fight(code, fight_id):
    """One boss fight's tank profile, or None when it isn't a boss fight or the report is archived.

    Two queries a fight past the report's own: the bosses' damage done (as enemies: the API's
    default hostility is friendlies, which returns nothing for a boss), then every event on the
    main tank, the player the bosses swung at most, which has its damage and healing taken (heal
    events with overheal, and absorbed events). Checked on Season of Discovery Onyxia, 2026-11-02.
    """
    meta = wcl.report_meta(code)
    fight = next(f for f in meta['fights'] if f['id'] == fight_id)
    if not fight.get('encounterID'):
        return None
    actors = {x['id']: x for x in meta['masterData']['actors']}
    bosses = [n['id'] for n in fight.get('enemyNPCs') or [] if actors.get(n['id'], {}).get('subType') == 'Boss']
    if not bosses:
        return None
    names = ability_names(meta)
    schools = {a['gameID']: a.get('type') for a in meta['masterData']['abilities']}
    boss_evs = []
    try:
        for b in bosses:
            boss_evs += [e for e in wcl.events(code, fight_id, data_type='DamageDone', source_id=b, hostility='Enemies')
                         if e.get('sourceID') == b and e['type'] == 'damage']
    except RuntimeError as err:
        # An archived report keeps its fights and actors but refuses its events.
        if 'archived' in str(err):
            return None
        raise
    swings = collections.Counter(e['targetID'] for e in boss_evs if e.get('abilityGameID') == WHITE)
    if not swings:
        return None
    mt = swings.most_common(1)[0][0]
    on_mt = [e for e in wcl.events(code, fight_id, data_type='All', target_id=mt) if e.get('targetID') == mt]
    windows = tanking_windows([e['timestamp'] for e in boss_evs if e['targetID'] == mt and e.get('abilityGameID') == WHITE])
    taken = [e for e in on_mt if e['type'] == 'damage' and inside(e['timestamp'], windows)]
    heals = [e for e in on_mt if e['type'] in ('heal', 'absorbed') and inside(e['timestamp'], windows)]
    return {'code': code, 'fight': fight_id, 'encounter': fight['encounterID'], 'name': fight['name'],
            'kill': fight.get('kill'), 'size': fight.get('size'), 'difficulty': fight.get('difficulty'),
            'start': fight['startTime'], 'date': meta['startTime'] + fight['startTime'], 'seconds': (fight['endTime'] - fight['startTime']) / 1000,
            'tanking': sum(e - s for s, e in windows) / 1000,
            'tank': actors.get(mt, {}).get('name'), 'tank_class': actors.get(mt, {}).get('subType'),
            'boss': boss_evs, 'mt': mt, 'windows': windows,
            'on_tank': [e for e in boss_evs if e['targetID'] == mt and inside(e['timestamp'], windows)],
            'taken': taken, 'heals': heals, 'names': names, 'schools': schools, 'actors': actors,
            'stints': [(e - s) / 1000 for s, e in windows]}


def ability_timers(fs, ab):
    """When a boss ability is used: its first use after the pull and the gaps between uses, in s.

    A use is a cast on anyone. Hits closer than a second are one use (a breath or a cleave hits
    several players at once)."""
    firsts, gaps = [], []
    for f in fs:
        ts = sorted(e['timestamp'] for e in f['boss'] if e.get('abilityGameID') == ab)
        uses = []
        for t in ts:
            if not uses or t - uses[-1] > 1000:
                uses.append(t)
        if uses:
            firsts.append((uses[0] - f['start']) / 1000)
        gaps += [(y - x) / 1000 for x, y in zip(uses, uses[1:])]
    return firsts, gaps


def cmd_tank(a):
    """What each boss does to its main tank, and the healing the tank takes, by encounter."""
    sample = load_sample(a.sample)
    fights = []
    seen = set()
    skipped = 0
    for entry in sample[:a.limit]:
        try:
            meta = wcl.report_meta(entry['code'])
        except wcl.BudgetError as err:
            print(f'stopped after {len(fights)} fights: {err}', file=sys.stderr)
            break
        for fid in fight_ranges(meta, entry):
            if (entry['code'], fid) in seen:
                continue
            seen.add((entry['code'], fid))
            fight = next(x for x in meta['fights'] if x['id'] == fid)
            # Only the encounters asked for, and kills unless --wipes, before spending a query.
            if a.encounter and fight.get('encounterID') not in a.encounter:
                continue
            if a.difficulty and fight.get('difficulty') not in a.difficulty:
                continue
            # A boss soloed or duoed by characters of a later expansion isn't a raid.
            if (fight.get('size') or len(fight.get('friendlyPlayers') or [])) < a.min_size:
                continue
            if not (fight.get('kill') or a.wipes):
                continue
            try:
                f = tank_fight(entry['code'], fid)
            except wcl.BudgetError as err:
                print(f'stopped after {len(fights)} fights: {err}', file=sys.stderr)
                break
            if f is None:
                skipped += 1
            elif f['kill'] or a.wipes:
                fights.append(f)
        else:
            continue
        break
    if skipped:
        print(f'{skipped} fights skipped (no boss swings, not a boss, or an archived report)', file=sys.stderr)
    by_enc = collections.defaultdict(list)
    for f in fights:
        by_enc[(f['encounter'], f['name'])].append(f)
    out = {}
    for (enc, name), fs in sorted(by_enc.items()):
        secs = sum(f['tanking'] for f in fs)
        lengths = [f['seconds'] for f in fs]
        print(f'\n== {name} (encounter {enc}): {len(fs)} fights, raid size {sorted({f["size"] for f in fs})}, '
              f'difficulty {sorted({f["difficulty"] for f in fs})}')
        print(f'fight length: median {median(lengths):.0f} s (10th to 90th percentile {pctl(lengths, 10):.0f} to {pctl(lengths, 90):.0f})')
        print(f'main tank: {collections.Counter(f["tank_class"] for f in fs).most_common()}, '
              f'tanking {100 * secs / sum(lengths):.0f}% of the fight time, median {median([f["tanking"] for f in fs]):.0f} s a fight, '
              f'longest stretch median {median([max(f["stints"]) for f in fs]):.0f} s')
        # The white swings on the main tank: rhythm, outcomes and size.
        outcomes = collections.Counter()
        gaps, hits, raws, raw_crush, raw_crit, rel = [], [], [], [], [], []
        for f in fs:
            sw = [e for e in f['on_tank'] if e.get('abilityGameID') == WHITE]
            # Each normal hit over its own kill's median: the boss's roll without the spread
            # between raids (their debuffs on him differ).
            own = [e['unmitigatedAmount'] for e in sw if e['hitType'] == HIT and e.get('unmitigatedAmount')]
            if len(own) >= 3:
                rel += [x / median(own) for x in own]
            outcomes.update(e['hitType'] for e in sw)
            gaps += [(y['timestamp'] - x['timestamp']) / 1000 for x, y in zip(sw, sw[1:])
                     if y['timestamp'] - x['timestamp'] <= TANKING_GAP_MS and y['sourceID'] == x['sourceID']]
            hits += [e['amount'] + e.get('absorbed', 0) for e in sw if e['hitType'] == HIT]
            raws += [e['unmitigatedAmount'] for e in sw if e['hitType'] == HIT and e.get('unmitigatedAmount')]
            raw_crush += [e['unmitigatedAmount'] for e in sw if e['hitType'] == CRUSH and e.get('unmitigatedAmount')]
            raw_crit += [e['unmitigatedAmount'] for e in sw if e['hitType'] == CRIT and e.get('unmitigatedAmount')]
        n = sum(outcomes.values())
        print(f'white swings on the tank: {n}, {n / max(secs, 1):.2f} a second while tanking, '
              f'median gap {median(gaps):.2f} s (25th to 75th {pctl(gaps, 25):.2f} to {pctl(gaps, 75):.2f})')
        for k, v in sorted(outcomes.items()):
            print(f'  {HIT_NAMES.get(k, k):<12} {v:>6} {100 * v / n:6.2f}% (95%: {interval(v, n)})')
        if hits:
            print(f'  normal hit after mitigation: median {median(hits):.0f}, min {min(hits):.0f}, max {max(hits):.0f}')
        if raws:
            print(f'  normal hit before mitigation (unmitigatedAmount): min {min(raws):.0f}, 5th {pctl(raws, 5):.0f}, '
                  f'median {median(raws):.0f}, 95th {pctl(raws, 95):.0f}, max {max(raws):.0f} ({len(raws)} hits)')
        if rel:
            print(f'  normal hit over its own kill\'s median: 5th {pctl(rel, 5):.3f}, 95th {pctl(rel, 95):.3f} '
                  f'({len(rel)} hits), so {median(raws) * pctl(rel, 5):.0f} to {median(raws) * pctl(rel, 95):.0f} around the median')
        if raw_crush or raw_crit:
            print(f'  unmitigatedAmount of crushing blows median {median(raw_crush):.0f} ({len(raw_crush)}), '
                  f'crits median {median(raw_crit):.0f} ({len(raw_crit)}): the same as a hit\'s means the field is before the multiplier')
        # The boss's other abilities on the tank.
        # A DoT's ticks are their own line: their size and rhythm aren't the hit's.
        abil = collections.defaultdict(list)
        for f in fs:
            for e in f['on_tank']:
                if e.get('abilityGameID') != WHITE:
                    abil[(e['abilityGameID'], bool(e.get('tick')))].append(e)
        if abil:
            print('abilities on the tank (a minute of tanking, outcomes, after mitigation, before it, '
                  'first use and gap between uses on anyone):')
        abilities = {}
        for (ab, tick), evs in sorted(abil.items(), key=lambda x: -len(x[1])):
            oc = collections.Counter(HIT_NAMES.get(e['hitType'], e['hitType']) for e in evs)
            landed = [e['amount'] + e.get('absorbed', 0) for e in evs if e['hitType'] in LANDED]
            raw = [e['unmitigatedAmount'] for e in evs if e['hitType'] == HIT and e.get('unmitigatedAmount')]
            firsts, ugaps = ability_timers(fs, ab)
            school = next((f['schools'].get(ab) for f in fs if f['schools'].get(ab)), None)
            print(f'  {ab:>8} {fs[0]["names"].get(ab, "?")[:20] + (" tick" if tick else ""):<26} {60 * len(evs) / max(secs, 1):5.2f}/min  '
                  + ' '.join(f'{k} {v}' for k, v in oc.most_common())
                  + (f'  landed median {median(landed):.0f} max {max(landed):.0f}' if landed else '')
                  + (f'  unmitigated hit {pctl(raw, 5):.0f} to {pctl(raw, 95):.0f}' if raw else '')
                  + (f'  first {median(firsts):.0f} s, every {median(ugaps):.1f} s ({pctl(ugaps, 10):.1f} to {pctl(ugaps, 90):.1f})' if ugaps else ''))
            abilities[f'{ab} tick' if tick else str(ab)] = {
                                  'id': ab, 'tick': tick, 'school': school,
                                  'name': fs[0]['names'].get(ab), 'per_minute': 60 * len(evs) / max(secs, 1),
                                  'outcomes': dict(oc), 'landed_median': median(landed),
                                  'landed_p05': pctl(landed, 5), 'landed_p95': pctl(landed, 95),
                                  'unmitigated_p05': pctl(raw, 5), 'unmitigated_median': median(raw),
                                  'unmitigated_p95': pctl(raw, 95), 'first_median': median(firsts),
                                  'gap_median': median(ugaps), 'gap_p10': pctl(ugaps, 10), 'gap_p90': pctl(ugaps, 90)}
        # What the main tank took and was healed for, while tanking.
        dmg = sum(e['amount'] + e.get('absorbed', 0) for f in fs for e in f['taken'])
        boss_dmg = sum(e['amount'] + e.get('absorbed', 0) for f in fs for e in f['on_tank'])
        deaths = sum(1 for f in fs for e in f['taken'] if e.get('overkill', 0) > 0)
        heal_evs = [e for f in fs for e in f['heals'] if e['type'] == 'heal']
        absorbs = sum(e['amount'] for f in fs for e in f['heals'] if e['type'] == 'absorbed')
        eff = sum(e['amount'] for e in heal_evs)
        over = sum(e.get('overheal', 0) for e in heal_evs)
        healers = collections.Counter()
        spells = collections.defaultdict(list)
        for f in fs:
            for e in f['heals']:
                if e['type'] != 'heal':
                    continue
                healers[(f['code'], f['fight'], e['sourceID'])] += e['amount'] + e.get('overheal', 0)
                spells[e.get('abilityGameID')].append(e)
        per_fight = collections.defaultdict(list)
        for (code, fid, src), amt in healers.items():
            per_fight[(code, fid)].append(amt)
        counts = [sum(1 for x in v if x >= 0.05 * sum(v)) for v in per_fight.values()]
        # The healing the tank's own spells did (a warlock's drains, a paladin's Holy Shield heals).
        own = sum(e['amount'] + e.get('overheal', 0) for f in fs for e in f['heals']
                  if e['type'] == 'heal' and e['sourceID'] == f['mt'])
        raw_hps = (eff + over + absorbs) / max(secs, 1)
        heals_ps = len(heal_evs) / max(secs, 1)
        # The healers' part, for an encounter file: the tank's own heals are the sim's to make.
        others = [e for f in fs for e in f['heals'] if e['type'] == 'heal' and e['sourceID'] != f['mt']]
        others_hps = (sum(e['amount'] + e.get('overheal', 0) for e in others) + absorbs) / max(secs, 1)
        others_ps = len(others) / max(secs, 1)
        print(f'main tank while tanking: {dmg / max(secs, 1):.0f} damage taken a second '
              f'({boss_dmg / max(secs, 1):.0f} from the boss), {deaths} deaths in {len(fs)} fights')
        print(f'  healing taken {eff / max(secs, 1):.0f} a second effective, {(eff + over) / max(secs, 1):.0f} with overheal '
              f'({100 * over / max(eff + over, 1):.0f}% overheal), absorbs {absorbs / max(secs, 1):.0f} a second, '
              f'{(eff + over + absorbs) / max(dmg, 1):.2f}x the damage taken with absorbs, '
              f'{100 * own / max(eff + over, 1):.0f}% of it the tank\'s own')
        print(f'  {heals_ps:.2f} heals a second ({sum(1 for e in heal_evs if e.get("tick")) / max(secs, 1):.2f} of them HoT ticks), '
              f'mean heal {(eff + over) / max(len(heal_evs), 1):.0f}, '
              f'healers with 5%+ of it: median {median(counts):.0f} (range {min(counts, default=0)}-{max(counts, default=0)})')
        for ab, evs in sorted(spells.items(), key=lambda x: -sum(e['amount'] + e.get('overheal', 0) for e in x[1]))[:a.top]:
            tot = sum(e['amount'] + e.get('overheal', 0) for e in evs)
            print(f'    {ab:>8} {fs[0]["names"].get(ab, "?"):<26} {100 * tot / max(eff + over, 1):5.1f}% of it, '
                  f'{60 * len(evs) / max(secs, 1):5.1f}/min, median {median([e["amount"] + e.get("overheal", 0) for e in evs]):.0f}')
        hc = median(counts)
        print(f'  as an encounter file\'s healing (the tank\'s own left out): {hc:.0f} healers at '
              f'{others_hps / max(hc, 1):.0f} HPS each (with overheal and absorbs), {hc / max(others_ps, 1e-9):.2f} s heals')
        out[str(enc)] = {'name': name, 'fights': len(fs), 'reports': [f'{f["code"]}:{f["fight"]}' for f in fs],
                         'tank_classes': dict(collections.Counter(f['tank_class'] for f in fs)),
                         'difficulty': sorted({f['difficulty'] for f in fs}),
                         'length_median': median(lengths), 'length_p10': pctl(lengths, 10), 'length_p90': pctl(lengths, 90),
                         'tanking_seconds': secs, 'tanking_median': median([f['tanking'] for f in fs]),
                         'stint_median': median([max(f['stints']) for f in fs]),
                         'lengths': lengths, 'tankings': [f['tanking'] for f in fs],
                         'stints': [max(f['stints']) for f in fs],
                         'dates': [f['date'] for f in fs], 'swings': {HIT_NAMES.get(k, k): v for k, v in outcomes.items()},
                         'swing_gap_median': median(gaps), 'hit_median': median(hits),
                         'unmitigated_min': min(raws, default=0), 'unmitigated_p05': pctl(raws, 5),
                         'unmitigated_median': median(raws), 'unmitigated_p95': pctl(raws, 95),
                         'unmitigated_max': max(raws, default=0), 'unmitigated_hits': len(raws),
                         'unmitigated_rel_p05': pctl(rel, 5), 'unmitigated_rel_p95': pctl(rel, 95), 'unmitigated_rel_hits': len(rel),
                         'abilities': abilities, 'dtps': dmg / max(secs, 1), 'boss_dtps': boss_dmg / max(secs, 1),
                         'hps_effective': eff / max(secs, 1), 'hps_total': (eff + over) / max(secs, 1),
                         'absorbs_ps': absorbs / max(secs, 1), 'hps_raw': raw_hps, 'own_share': own / max(eff + over, 1),
                         'heals_ps': heals_ps, 'healers_median': hc, 'deaths': deaths,
                         'healers_hps': others_hps, 'healers_heals_ps': others_ps,
                         'file_healing': {'healers': hc, 'hpsEach': others_hps / max(hc, 1),
                                          'castSeconds': hc / max(others_ps, 1e-9)}}
    if a.out:
        json.dump(out, open(a.out, 'w'), indent=1)
        print(f'\nwrote {a.out}', file=sys.stderr)


def main():
    ap = argparse.ArgumentParser(description=__doc__, formatter_class=argparse.RawDescriptionHelpFormatter)
    ap.add_argument('--site', default=wcl.SITE, help='Warcraft Logs subdomain (sod, vanilla, fresh, classic, ...)')
    ap.add_argument('--budget', type=float, default=wcl.BUDGET, help='most API points this run may spend')
    sub = ap.add_subparsers(dest='cmd', required=True)

    sub.add_parser('zones', help="list the site's zones, encounters and partitions")

    s = sub.add_parser('sample', help='pick players and fights to read')
    s.add_argument('-o', '--out', required=True)
    s.add_argument('--append', action='store_true', help='add to the file instead of replacing it')
    s.add_argument('--class', dest='cls', default='Shaman')
    s.add_argument('--spec', help='rankings only, like Enhancement')
    s.add_argument('--encounter', type=int, action='append', help='from the rankings of this encounter')
    s.add_argument('--partition', type=int)
    s.add_argument('--difficulty', type=int)
    s.add_argument('--per-encounter', type=int, default=20)
    s.add_argument('--pages', type=int, default=1)
    s.add_argument('--zone', type=int, help='from the newest public reports of this zone')
    s.add_argument('--reports', type=int, default=10)
    s.add_argument('--since', help='with --zone: reports from this date on (YYYY-MM-DD, UTC)')
    s.add_argument('--until', help='with --zone: reports before this date (YYYY-MM-DD, UTC)')
    s.add_argument('--server', action='append', help="with --zone: only guilds of this realm (its slug, like dreamscythe); repeat for more")
    s.add_argument('--exclude-server', action='append', help='with --zone: not guilds of this realm, nor reports with no guild')
    s.add_argument('--report', action='append', metavar='CODE:FIGHT:NAME',
                   help='one player in one report (leave FIGHT empty for every fight)')

    s = sub.add_parser('abilities', help="list a player's abilities by event type")
    s.add_argument('target', metavar='CODE:FIGHT:NAME')

    def analysis(name, help):
        s = sub.add_parser(name, help=help)
        s.add_argument('sample')
        s.add_argument('--limit', type=int, default=None, help='read only the first N entries')
        return s

    def proc_args(s, types):
        s.add_argument('--ability', required=True, help='proc ability IDs, comma separated')
        s.add_argument('--type', default=types, help='event types that count as a proc')
        s.add_argument('--merge', type=int, default=50,
                       help='damage or extra attack events this many ms apart or closer count as one proc')

    s = analysis('gaps', 'gaps between procs, to find a proc cooldown')
    proc_args(s, 'damage,extraattacks')
    s.add_argument('--trigger-window', type=int, default=400)
    s.add_argument('--bin', type=int, default=50)
    s.add_argument('--max', type=int, default=3000)
    s.add_argument('--top', type=int, default=20)

    s = analysis('glance', 'white swing outcomes, glances in particular')
    s.add_argument('--targets', choices=['boss', 'all'], default='boss')

    s = analysis('procrate', 'procs per landed hit and PPM')
    proc_args(s, 'applybuff,applybuffstack,refreshbuff')
    s.add_argument('--triggers', default='1', help='ability IDs whose landed hits can proc (1 is white swings)')
    s.add_argument('--main-hand-only', action='store_true', help='off hand swings never proc')
    s.add_argument('--speed', action='append', metavar='ITEM=SPEED', help='weapon speed the item database lacks')

    s = analysis('tank', "each boss's damage on its main tank and the healing the tank takes")
    s.add_argument('--wipes', action='store_true', help='count wipes too (kills only by default)')
    s.add_argument('--encounter', type=int, action='append', help='only this encounter (repeat for more)')
    s.add_argument('--difficulty', type=int, action='append',
                   help="only this difficulty (repeat for more): 3 Normal, 4 Heroic, 5 Mythic (SoD's heat levels)")
    s.add_argument('--min-size', type=int, default=10,
                   help='only fights with this many players or more (default 10, which drops soloed bosses)')
    s.add_argument('--top', type=int, default=8, help='heal spells to list')
    s.add_argument('-o', '--out', help='also write the numbers to this JSON file')

    wclraid.add_parser(analysis)

    a = ap.parse_args()
    wcl.SITE = a.site
    wcl.BUDGET = a.budget
    try:
        {'zones': cmd_zones, 'sample': cmd_sample, 'abilities': cmd_abilities, 'gaps': cmd_gaps,
         'glance': cmd_glance, 'procrate': cmd_procrate, 'tank': cmd_tank,
         'raid': lambda a: wclraid.cmd_raid(a, load_sample, fight_ranges)}[a.cmd](a)
    except wcl.BudgetError as err:
        print(f'stopped: {err}', file=sys.stderr)
    print(wcl.summary(), file=sys.stderr)


if __name__ == '__main__':
    main()
