"""Raid profiles: what a boss fight does to the raid besides its tanks, and how the raid heals it.

The healing sim's side of the boss files in tools/encounters (docs/healing_sim.md). Per fight it
reads three things past the report's own fights and actors:

- every damage event on a friendly player (DamageTaken), which has the boss's swings on its tanks
  too, so the tanks are the players the bosses swung at,
- the deaths,
- playerDetails (Warcraft Logs' tank, healer and dps roles) and the Healing table by source and by
  target, for the healer count and the healing on the tanks and on the rest.

About 4 points a fight, cached forever like the rest (wcl.py).
"""
import collections, json, statistics, sys

import wcl

WHITE = 1
CAST_GAP_MS = 1500  # Hits of one ability from one source this close are one cast.


def pct(xs, p):
    xs = sorted(xs)
    if not xs:
        return 0
    k = (len(xs) - 1) * p
    lo = int(k)
    hi = min(lo + 1, len(xs) - 1)
    return xs[lo] + (xs[hi] - xs[lo]) * (k - lo)


def med(xs):
    return pct(xs, 0.5)


def table(code, fight, data_type, view_by=None):
    """A fight's Warcraft Logs table (Healing, DamageTaken, ...), its 'data' part."""
    args = {'code': code, 'fights': [fight], 'dt': data_type}
    view_arg, view_use = '', ''
    if view_by:
        args['vb'] = view_by
        view_arg, view_use = ', $vb: ViewType', ', viewBy: $vb'
    meta = wcl.report_meta(code)
    f = next(x for x in meta['fights'] if x['id'] == fight)
    args['start'], args['end'] = f['startTime'], f['endTime']
    d = wcl.query('''query($code: String!, $fights: [Int], $dt: TableDataType, $start: Float, $end: Float'''
                  + view_arg + ''') { reportData { report(code: $code) {
        table(fightIDs: $fights, dataType: $dt, startTime: $start, endTime: $end''' + view_use + ''')
    } } }''', args)['reportData']['report']['table']
    return d.get('data', d)


def player_details(code, fight):
    """Warcraft Logs' roles for a fight's players: {'tanks': [...], 'healers': [...], 'dps': [...]}."""
    meta = wcl.report_meta(code)
    f = next(x for x in meta['fights'] if x['id'] == fight)
    d = wcl.query('''query($code: String!, $fights: [Int], $start: Float, $end: Float) {
        reportData { report(code: $code) {
            playerDetails(fightIDs: $fights, startTime: $start, endTime: $end) } } }''',
                  {'code': code, 'fights': [fight], 'start': f['startTime'], 'end': f['endTime']})
    pd = d['reportData']['report']['playerDetails']
    pd = pd.get('data', pd)
    return pd.get('playerDetails', pd)


def raid_fight(code, fight_id):
    """One boss fight's raid profile, or None when it isn't a boss fight."""
    meta = wcl.report_meta(code)
    fight = next(f for f in meta['fights'] if f['id'] == fight_id)
    if not fight.get('encounterID'):
        return None
    actors = {x['id']: x for x in meta['masterData']['actors']}
    players = {i for i in fight.get('friendlyPlayers') or [] if actors.get(i, {}).get('type') == 'Player'}
    if len(players) < 0.8 * (fight.get('size') or 0):
        return None  # a log of part of the raid: its damage and healing would be part of the fight's
    bosses = {n['id'] for n in fight.get('enemyNPCs') or [] if actors.get(n['id'], {}).get('subType') == 'Boss'}
    taken = [e for e in wcl.events(code, fight_id, data_type='DamageTaken')
             if e['type'] == 'damage' and e.get('targetID') in players]
    swings = collections.Counter(e['targetID'] for e in taken
                                 if e.get('sourceID') in bosses and e.get('abilityGameID') == WHITE)
    if not swings:
        return None
    total = sum(swings.values())
    tanks = {t for t, n in swings.items() if n >= 0.05 * total}
    deaths = [e for e in wcl.events(code, fight_id, data_type='Deaths')
              if e['type'] == 'death' and e.get('targetID') in players]
    roles = player_details(code, fight_id)
    by_source = table(code, fight_id, 'Healing')
    by_target = table(code, fight_id, 'Healing', 'Target')
    return {'code': code, 'fight': fight_id, 'encounter': fight['encounterID'], 'name': fight['name'],
            'kill': fight.get('kill'), 'size': fight.get('size'), 'start': fight['startTime'],
            'seconds': (fight['endTime'] - fight['startTime']) / 1000,
            'players': players, 'tanks': tanks, 'bosses': bosses, 'taken': taken, 'deaths': deaths,
            'roles': roles, 'heal_sources': by_source.get('entries', []),
            'heal_targets': by_target.get('entries', []),
            'names': {x['gameID']: x['name'] for x in meta['masterData']['abilities']},
            'schools': {x['gameID']: x['type'] for x in meta['masterData']['abilities']}, 'actors': actors}


def landed(e):
    return e.get('amount', 0) + e.get('absorbed', 0)


def casts(evs):
    """Split one ability's hits from one source into casts: [(time, [events])]."""
    out = []
    for e in sorted(evs, key=lambda e: e['timestamp']):
        if out and e['timestamp'] - out[-1][1][-1]['timestamp'] <= CAST_GAP_MS:
            out[-1][1].append(e)
        else:
            out.append((e['timestamp'], [e]))
    return out


def heal_total(entry):
    return entry.get('total', 0) + entry.get('overheal', 0)


def summarize(name, enc, fs, top):
    """Print one encounter's raid profile and return its numbers."""
    secs = [f['seconds'] for f in fs]
    tot_secs = sum(secs)
    print(f'\n== {name} (encounter {enc}): {len(fs)} fights, raid size {sorted({f["size"] for f in fs})}')
    print(f'fight length: median {med(secs):.0f} s (10th to 90th percentile {pct(secs, .1):.0f} to {pct(secs, .9):.0f})')
    ntank = [len(f['tanks']) for f in fs]
    nplay = [len(f['players']) for f in fs]
    print(f'players: median {med(nplay):.0f}; tanks the bosses swung at (5%+ of swings): median {med(ntank):.0f} '
          f'(range {min(ntank)}-{max(ntank)})')

    # Damage on the players who aren't tanks, by ability.
    others_secs = sum(f['seconds'] * len(f['players'] - f['tanks']) for f in fs)
    dmg_others = sum(landed(e) for f in fs for e in f['taken'] if e['targetID'] not in f['tanks'])
    dmg_tanks = sum(landed(e) for f in fs for e in f['taken'] if e['targetID'] in f['tanks'])
    print(f'damage taken: {dmg_others / tot_secs:.0f} a second on the non-tanks together, '
          f'{dmg_others / max(others_secs, 1):.1f} a second each; {dmg_tanks / tot_secs:.0f} a second on the tanks together')
    abil = collections.defaultdict(list)  # (ability, source name, tick) -> [(fight, event)]
    for f in fs:
        for e in f['taken']:
            if e['targetID'] in f['tanks']:
                continue
            src = f['actors'].get(e.get('sourceID'), {})
            sname = src.get('name') or ('Environment' if e.get('sourceID', -1) == -1 else '?')
            if src.get('type') == 'Player' or src.get('petOwner') or e.get('sourceID') == e['targetID']:
                sname = 'friendly'
            abil[(e.get('abilityGameID'), sname, bool(e.get('tick')))].append((f, e))
    # A player's own damage on themselves (Seal of Martyrdom, Life Tap) shows up as from the
    # environment: an Environment or friendly row whose hits are all on one class is left out.
    rows, dropped = [], []
    for (ab, src, tick), fe in abil.items():
        amt = sum(landed(e) for _, e in fe)
        cls = collections.Counter(f['actors'].get(e['targetID'], {}).get('subType') for f, e in fe)
        if src in ('Environment', 'friendly', '?') and cls.most_common(1)[0][1] >= 0.95 * len(fe):
            dropped.append(f'{fs[0]["names"].get(ab, ab)} ({cls.most_common(1)[0][0]}, {100 * amt / max(dmg_others, 1):.1f}%)')
            continue
        rows.append((amt, ab, src, tick, fe))
    if dropped:
        print('left out as self damage: ' + ', '.join(dropped))
    rows.sort(key=lambda r: -r[0])
    print(f'abilities on the non-tanks ({100 * sum(r[0] for r in rows[:top]) / max(dmg_others, 1):.0f}% of their damage '
          f'in the top {top}): share, per fight, players hit a cast (of the non-tanks), first and gap between casts, '
          'landed and unmitigated amounts')
    out_abil = []
    for amt, ab, src, tick, fe in rows[:top]:
        fights_with = {id(f) for f, _ in fe}
        per_fight_casts, targets, firsts, gaps = [], [], [], []
        hits_per_target, tick_gaps, ncasts = [], [], 0
        for f in fs:
            evs = [e for g, e in fe if g is f]
            if not evs:
                continue
            by_src = collections.defaultdict(list)
            for e in evs:
                by_src[e.get('sourceID')].append(e)
            cs = sorted((c for v in by_src.values() for c in casts(v)), key=lambda c: c[0])
            per_fight_casts.append(len(cs))
            ncasts += len(cs)
            firsts.append((cs[0][0] - f['start']) / 1000)
            gaps += [(b[0] - a[0]) / 1000 for a, b in zip(cs, cs[1:])]
            for _, c in cs:
                n = len({e['targetID'] for e in c})
                targets.append(n / max(len(f['players'] - f['tanks']), 1))
                hits_per_target.append(len(c) / n)
                if tick:
                    on = collections.defaultdict(list)
                    for e in c:
                        on[e['targetID']].append(e['timestamp'])
                    for ts in on.values():
                        tick_gaps += [(b - a) / 1000 for a, b in zip(ts, ts[1:])]
        evs = [e for _, e in fe]
        land = [landed(e) for e in evs if landed(e) > 0]
        raw = [e['unmitigatedAmount'] for e in evs if e.get('unmitigatedAmount')]
        hit_types = collections.Counter(e.get('hitType') for e in evs)
        name_ = fs[0]['names'].get(ab, '?')
        print(f'  {ab:>8} {name_[:24]:<24} {"tick " if tick else "     "}{src[:18]:<18} {100 * amt / max(dmg_others, 1):5.1f}%  '
              f'{len(fights_with)}/{len(fs)} fights, {med(per_fight_casts):.0f} casts a fight, '
              f'{100 * med(targets):.0f}% hit a cast ({100 * pct(targets, .1):.0f}-{100 * pct(targets, .9):.0f}), '
              f'{med(hits_per_target):.1f} hits each, first {med(firsts):.0f} s, gap {med(gaps):.1f} s '
              f'({pct(gaps, .1):.1f}-{pct(gaps, .9):.1f}), landed median {med(land):.0f} '
              + (f'unmitigated {pct(raw, .05):.0f}-{pct(raw, .95):.0f} (median {med(raw):.0f})' if raw else ''))
        out_abil.append({'ability': ab, 'name': name_, 'source': src, 'tick': tick,
                         'school': next((f['schools'].get(ab) for f in fs if f['schools'].get(ab)), None),
                         'share': amt / max(dmg_others, 1), 'dtps_each': amt / max(others_secs, 1),
                         'fights': len(fights_with), 'casts_per_fight': med(per_fight_casts),
                         'casts_per_second': ncasts / tot_secs, 'tick_seconds': med(tick_gaps),
                         'landed_share': sum(1 for e in evs if landed(e) > 0) / len(evs),
                         'hit_fraction': med(targets), 'hit_fraction_p10': pct(targets, .1),
                         'hit_fraction_p90': pct(targets, .9), 'hits_per_target': med(hits_per_target),
                         'first': med(firsts), 'gap': med(gaps), 'gap_p10': pct(gaps, .1), 'gap_p90': pct(gaps, .9),
                         'landed_median': med(land), 'raw_p05': pct(raw, .05), 'raw_median': med(raw),
                         'raw_p95': pct(raw, .95), 'hit_types': {str(k): v for k, v in hit_types.items()}})

    # Deaths.
    d_tank = [sum(1 for e in f['deaths'] if e['targetID'] in f['tanks']) for f in fs]
    d_other = [sum(1 for e in f['deaths'] if e['targetID'] not in f['tanks']) for f in fs]
    killers = collections.Counter()
    for f in fs:
        for e in f['deaths']:
            killers[f['names'].get(e.get('killingAbilityGameID'), str(e.get('killingAbilityGameID')))] += 1
    print(f'deaths a fight: non-tanks mean {statistics.mean(d_other):.2f} (median {med(d_other):.0f}, max {max(d_other)}), '
          f'tanks mean {statistics.mean(d_tank):.2f}; fights with no death {sum(1 for a, b in zip(d_tank, d_other) if a + b == 0)}/{len(fs)}; '
          'killed by ' + ', '.join(f'{k} {v}' for k, v in killers.most_common(6)))

    # Healers and healing.
    heal_counts, wcl_healers, hps_each, heal_eff, heal_over = [], [], [], 0, 0
    on_tanks_eff, on_tanks_all, on_others_eff, on_others_all = 0, 0, 0, 0
    classes = collections.Counter()
    healer_share = []
    for f in fs:
        src = [x for x in f['heal_sources'] if x.get('id') in f['players']]
        tot = sum(heal_total(x) for x in src)
        heal_counts.append(sum(1 for x in src if heal_total(x) >= 0.05 * tot))
        # The healers are the players Warcraft Logs gives the healer role (by spec).
        role_ids = {x.get('id') for x in (f['roles'].get('healers', []) if isinstance(f['roles'], dict) else [])}
        healers = [x for x in src if x.get('id') in role_ids]
        wcl_healers.append(len(role_ids))
        hps_each += [heal_total(x) / f['seconds'] for x in healers]
        classes.update(x.get('type', '?') for x in healers)
        healer_share.append(sum(heal_total(x) for x in healers) / max(tot, 1))
        heal_eff += sum(x.get('total', 0) for x in src)
        heal_over += sum(x.get('overheal', 0) for x in src)
        for x in f['heal_targets']:
            if x.get('id') in f['tanks']:
                on_tanks_eff += x.get('total', 0)
                on_tanks_all += heal_total(x)
            elif x.get('id') in f['players']:
                on_others_eff += x.get('total', 0)
                on_others_all += heal_total(x)
    print(f'healers (Warcraft Logs\' healer role): median {med(wcl_healers):.0f} (range {min(wcl_healers)}-{max(wcl_healers)}), '
          f'{100 * med(healer_share):.0f}% of the healing; players with 5%+ of the healing: median {med(heal_counts):.0f}; '
          + ', '.join(f'{k} {v / len(fs):.1f}' for k, v in classes.most_common()))
    print(f'  each healer {med(hps_each):.0f} HPS with overheal (10th to 90th {pct(hps_each, .1):.0f} to {pct(hps_each, .9):.0f}); '
          f'raid {heal_eff / tot_secs:.0f} HPS effective, {(heal_eff + heal_over) / tot_secs:.0f} with overheal '
          f'({100 * heal_over / max(heal_eff + heal_over, 1):.0f}% overheal)')
    print(f'  on the tanks {on_tanks_eff / tot_secs:.0f} effective, {on_tanks_all / tot_secs:.0f} with overheal; '
          f'on the rest {on_others_eff / tot_secs:.0f} effective, {on_others_all / tot_secs:.0f} with overheal '
          f'({100 * (1 - on_others_eff / max(on_others_all, 1)):.0f}% overheal), '
          f'{on_others_eff / max(others_secs, 1):.1f} effective a second each')
    return {'name': name, 'fights': len(fs), 'reports': sorted({f'{f["code"]}:{f["fight"]}' for f in fs}),
            'seconds_median': med(secs), 'seconds_p10': pct(secs, .1), 'seconds_p90': pct(secs, .9),
            'players_median': med(nplay), 'tanks_median': med(ntank),
            'dtps_others_total': dmg_others / tot_secs, 'dtps_others_each': dmg_others / max(others_secs, 1),
            'dtps_tanks_total': dmg_tanks / tot_secs, 'abilities': out_abil,
            'deaths_others_mean': statistics.mean(d_other), 'deaths_tanks_mean': statistics.mean(d_tank),
            'deathless_fights': sum(1 for a, b in zip(d_tank, d_other) if a + b == 0),
            'killers': dict(killers.most_common(10)),
            'healers_median': med(wcl_healers), 'healers_min': min(wcl_healers), 'healers_max': max(wcl_healers),
            'healers_share': med(healer_share), 'healers_5pct_median': med(heal_counts), 'healer_classes': {k: v / len(fs) for k, v in classes.items()},
            'hps_each_median': med(hps_each), 'hps_effective': heal_eff / tot_secs,
            'hps_total': (heal_eff + heal_over) / tot_secs,
            'hps_tanks_effective': on_tanks_eff / tot_secs, 'hps_tanks_total': on_tanks_all / tot_secs,
            'hps_others_effective': on_others_eff / tot_secs, 'hps_others_total': on_others_all / tot_secs}


def cmd_raid(a, load_sample, fight_ranges):
    """The raid side of each boss fight in a sample: damage on the non-tanks, deaths, healers and healing."""
    sample = load_sample(a.sample)
    fights, seen = [], set()
    want = set(a.encounter or [])
    stopped = False
    for entry in sample[:a.limit]:
        try:
            meta = wcl.report_meta(entry['code'])
            for fid in fight_ranges(meta, entry):
                f = next(x for x in meta['fights'] if x['id'] == fid)
                if (entry['code'], fid) in seen or (want and f.get('encounterID') not in want):
                    continue
                if not (f.get('kill') or a.wipes) or (f.get('size') or 0) < a.min_size:
                    continue
                if a.difficulty and f.get('difficulty') not in a.difficulty:
                    continue
                seen.add((entry['code'], fid))
                try:
                    r = raid_fight(entry['code'], fid)
                except RuntimeError as err:
                    if 'archived' not in str(err):
                        raise
                    print(f'skipped {entry["code"]}: archived', file=sys.stderr)
                    break
                if r:
                    fights.append(r)
        except wcl.BudgetError as err:
            print(f'stopped after {len(fights)} fights: {err}', file=sys.stderr)
            stopped = True
            break
    by_enc = collections.defaultdict(list)
    for f in fights:
        by_enc[(f['encounter'], f['name'])].append(f)
    out = {}
    for (enc, name), fs in sorted(by_enc.items()):
        out[str(enc)] = summarize(name, enc, fs, a.top)
    out['_site'] = wcl.SITE
    if stopped:
        out['_stopped'] = True
    if a.out:
        json.dump(out, open(a.out, 'w'), indent=1)
        print(f'\nwrote {a.out}', file=sys.stderr)


def add_parser(analysis):
    s = analysis('raid', "each boss fight's damage on the raid besides its tanks, deaths, healers and healing")
    s.add_argument('--wipes', action='store_true', help='count wipes too (kills only by default)')
    s.add_argument('--encounter', type=int, action='append', help='only these encounter IDs')
    s.add_argument('--min-size', type=int, default=10, help='skip fights with a smaller raid')
    s.add_argument('--difficulty', type=int, action='append', help="only this difficulty (3 Normal, 4 Heroic, 5 Mythic: SoD's heat levels)")
    s.add_argument('--top', type=int, default=12, help='abilities to list')
    s.add_argument('-o', '--out', help='also write the numbers to this JSON file')
