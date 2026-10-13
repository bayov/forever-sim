"""Mine public Warcraft Logs reports for proc cooldowns, proc rates and glancing blows.

Usage: python3 tools/warcraftlogs/wclcli.py [--site sod] [--budget 300] <command> ...
See README.md in this directory for the workflow.
"""
import argparse, bisect, collections, json, math, os, sys

sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
import wcl

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
        reports = []
        page = 1
        while len(reports) < a.reports:
            r = wcl.query('''query($z: Int, $p: Int) { reportData {
                reports(zoneID: $z, page: $p, limit: 100) { data { code } has_more_pages } } }''',
                          {'z': a.zone, 'p': page}, ttl=wcl.DAY)['reportData']['reports']
            reports += [x['code'] for x in r['data']]
            if not r['has_more_pages']:
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

    a = ap.parse_args()
    wcl.SITE = a.site
    wcl.BUDGET = a.budget
    try:
        {'zones': cmd_zones, 'sample': cmd_sample, 'abilities': cmd_abilities, 'gaps': cmd_gaps,
         'glance': cmd_glance, 'procrate': cmd_procrate}[a.cmd](a)
    except wcl.BudgetError as err:
        print(f'stopped: {err}', file=sys.stderr)
    print(wcl.summary(), file=sys.stderr)


if __name__ == '__main__':
    main()
