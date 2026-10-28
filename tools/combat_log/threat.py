"""Threat per ability, from a combat log and the threat BayovCore recorded during it.

The combat log has no threat. BayovCore's Threat module (bayovui/addons/BayovCore/Modules/
Threat.lua in the WoW repo) records the raw threat the game reports, for us, our pet and two
group members on every mob we fight, and wow-combat-logs' archive.py keeps each log's fights
next to it as <log>.threat.jsonl.xz (its docstring has the format). The addon writes only the
numbers. This lines them up with the log and works out which actions moved them.

    python3 -I tools/combat_log/threat.py <log> [threat file] [--who NAME] [--rows]

<log> is the combat log, plain or .xz. Without a threat file we look for
<log>.threat.jsonl.xz next to it, then in ~/wow-combat-logs/<client>/.

How it matches:

1. The clock. Each fight has `at`, the local wall clock of its first reading, to about a frame
   (`rough` fights only to a second). The server also sends a threat update a moment after the
   hit it counts. So we try lags from -1 to +1 sec (-2 to +2 when rough) and keep the one that
   puts the most rises in a player's threat right after one of their own lines on that mob.
2. Each change in a player's threat on a mob covers the lines since their previous reading:
   their hits, misses and casts on that mob, and their heals and energizes (which threaten
   every mob). A change with exactly one such line is "clean", and gives that line's threat.
3. Per player and ability, the clean samples: how many, the median threat, and for damage the
   median threat per point of damage (the amount, which leaves out what was resisted, blocked
   or absorbed). The game's raw threat is Classic's threat times 100, so we say which scale
   the white hits point to.

Threat starts at 0, so a player's first reading counts from the pull (the lines up to PULL sec
before the fight's first reading). A -1 reading is a player dropping off the mob's list. Their
threat is 0 after it.
"""
import argparse
import datetime
import json
import lzma
import os
import statistics
import sys
from collections import defaultdict

sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
import cl  # noqa: E402

# The lines of a player that can move their threat on one mob, and those that move it on all.
ON_MOB = ('SWING_DAMAGE', 'SWING_MISSED', 'RANGE_DAMAGE', 'RANGE_MISSED', 'SPELL_DAMAGE',
          'SPELL_MISSED', 'SPELL_PERIODIC_DAMAGE', 'SPELL_CAST_SUCCESS', 'SPELL_AURA_APPLIED',
          'SPELL_AURA_REFRESH', 'SPELL_AURA_APPLIED_DOSE', 'DAMAGE_SHIELD')
# How long before a fight's first reading its first lines can be.
PULL = 3

ON_ALL = ('SPELL_HEAL', 'SPELL_PERIODIC_HEAL', 'SPELL_ENERGIZE', 'SPELL_PERIODIC_ENERGIZE')


def threat_path(log):
    base = os.path.basename(log)
    for ext in ('.txt.xz', '.txt'):
        if base.endswith(ext):
            base = base[:-len(ext)]
    name = base + '.threat.jsonl.xz'
    here = os.path.join(os.path.dirname(log), name)
    if os.path.exists(here):
        return here
    archive = os.path.expanduser('~/wow-combat-logs')
    for client in sorted(os.listdir(archive)) if os.path.isdir(archive) else []:
        p = os.path.join(archive, client, name)
        if os.path.exists(p):
            return p
    return None


def load_fights(path):
    opener = lzma.open if path.endswith('.xz') else open
    with opener(path, 'rt', encoding='utf-8') as f:
        return [json.loads(line) for line in f if line.strip()]


def epoch(at):
    """A fight's `at` (local wall clock) as the same kind of number cl.Ev.t is."""
    return datetime.datetime.strptime(at, '%Y-%m-%d %H:%M:%S.%f').timestamp()


def label(e):
    if e.ev.startswith('SWING'):
        name = 'Melee'
    elif e.ev.startswith('RANGE'):
        name = 'Ranged ' + (e.spell or '')
    else:
        name = e.spell or '?'
    kind = e.ev.split('_', 1)[1]
    if kind.endswith('MISSED'):
        return '%s (%s)' % (name, cl.miss(e).lower())
    if kind in ('DAMAGE', 'PERIODIC_DAMAGE'):
        return name + (' (tick)' if kind.startswith('PERIODIC') else '')
    return '%s (%s)' % (name, kind.lower().replace('_', ' '))


def amount(e):
    if e.ev in ('SWING_DAMAGE', 'RANGE_DAMAGE', 'SPELL_DAMAGE', 'SPELL_PERIODIC_DAMAGE', 'DAMAGE_SHIELD'):
        try:
            return cl.dmg(e)['amount']
        except (IndexError, ValueError):
            return None
    return None


class Series:
    """One player's readings on one mob: [(time, threat)], with -1 drops as threat 0."""

    def __init__(self, who, mob):
        self.who, self.mob = who, mob
        self.points = []


def series_of(fight):
    out = {}
    start = epoch(fight['at'])
    for r in fight['readings']:
        t, m, w = start + r[0] / 1000.0, r[1], r[2]
        key = (w, m)
        if key not in out:
            out[key] = Series(fight['who'][w - 1], fight['mobs'][m - 1])
            # Threat starts at 0, so a first reading is a change from the pull.
            out[key].points.append((start - PULL, 0))
        out[key].points.append((t, 0 if r[3] == -1 else r[6]))
    return out


def lines_of(evs, guid):
    """A player's lines, by mob GUID, and those that threaten every mob."""
    on, every = defaultdict(list), []
    for e in evs:
        if e.srcGUID != guid:
            continue
        if e.ev in ON_ALL:
            every.append(e)
        elif e.ev in ON_MOB:
            on[e.dstGUID].append(e)
    return on, every


def window(lines, a, b):
    return [e for e in lines if a < e.t <= b]


def changes(s):
    """(from, to, delta) for each change in a series."""
    prev = None
    for t, v in s.points:
        if prev is not None and v != prev[1]:
            yield prev[0], t, v - prev[1]
        prev = (t, v)


def fit_lag(series, by_who, rough):
    span = 2.0 if rough else 1.0
    best = (-1, 0.0)
    for step in range(int(-span * 100), int(span * 100) + 1, 2):
        lag = step / 100.0
        hits = 0
        for s in series.values():
            on, _ = by_who.get(s.who['guid'], ({}, []))
            lines = on.get(s.mob['guid'], [])
            for a, b, d in changes(s):
                if d > 0 and window(lines, b - lag - 0.4, b - lag):
                    hits += 1
        if hits > best[0] or (hits == best[0] and abs(lag) < abs(best[1])):
            best = (hits, lag)
    return best[1], best[0]


def main():
    ap = argparse.ArgumentParser(description=__doc__.split('\n')[0])
    ap.add_argument('log')
    ap.add_argument('threat', nargs='?')
    ap.add_argument('--who', help='only this player (name)')
    ap.add_argument('--rows', action='store_true', help='print every change and the lines it covers')
    args = ap.parse_args()

    tpath = args.threat or threat_path(args.log)
    if not tpath:
        sys.exit('no threat file for %s (archive.py writes <log>.threat.jsonl.xz)' % args.log)
    fights = load_fights(tpath)
    evs = cl.load(args.log)
    print('%s: %d fights, %d lines in the log' % (os.path.basename(tpath), len(fights), len(evs)))

    # (who guid, ability) -> list of (threat, amount or None)
    samples = defaultdict(list)
    names = {}
    mixed = defaultdict(int)
    for fight in fights:
        series = series_of(fight)
        start = epoch(fight['at'])
        end = max((p[0] for s in series.values() for p in s.points), default=start)
        near = [e for e in evs if start - 3 <= e.t <= end + 3]
        by_who = {w['guid']: lines_of(near, w['guid']) for w in fight['who']}
        lag, matched = fit_lag(series, by_who, fight.get('rough'))
        who = ', '.join('%s (%s)' % (w.get('name') or '?', (w.get('class') or w.get('kind') or '?').lower())
                        for w in fight['who'])
        print('\n%s  %s, %d readings, %d mobs, lag %+.2f sec (%d rises matched)%s%s' % (
            fight['at'], fight.get('zone') or '?', len(fight['readings']), len(fight['mobs']), lag, matched,
            ', %d hidden' % fight['secret'] if fight.get('secret') else '',
            ', rough clock' if fight.get('rough') else ''))
        print('  ' + who)
        for s in series.values():
            g = s.who['guid']
            names[g] = '%s (%s)' % (s.who.get('name') or '?', (s.who.get('class') or s.who.get('kind') or '?').lower())
            if args.who and s.who.get('name') != args.who:
                continue
            on, every = by_who.get(g, ({}, []))
            lines = sorted(on.get(s.mob['guid'], []) + every, key=lambda e: e.t)
            for a, b, d in changes(s):
                got = window(lines, a - lag, b - lag)
                if args.rows:
                    print('    %s %s on %s: %+g  %s' % (
                        datetime.datetime.fromtimestamp(b).strftime('%H:%M:%S.%f')[:-3],
                        s.who.get('name'), s.mob.get('name'), d,
                        ', '.join(label(e) + ('' if amount(e) is None else ' %d' % amount(e)) for e in got)
                        or '(no line)'))
                if len(got) == 1:
                    samples[(g, label(got[0]))].append((d, amount(got[0])))
                elif got:
                    mixed[g] += 1

    print('\nClean samples (one line between two readings):')
    melee = []
    for (g, ability), got in sorted(samples.items(), key=lambda kv: (names.get(kv[0][0], ''), kv[0][1])):
        if args.who and not names.get(g, '').startswith(args.who + ' '):
            continue
        threat = statistics.median(d for d, _ in got)
        per = [d / a for d, a in got if a]
        if ability == 'Melee' and per:
            melee.extend(per)
        print('  %-24s %-34s %4d  threat %10.1f%s' % (
            names.get(g, g), ability, len(got), threat,
            '  per damage %.2f' % statistics.median(per) if per else ''))
    for g, n in mixed.items():
        print('  %s: %d changes covered more than one line (not used)' % (names.get(g, g), n))
    if melee:
        m = statistics.median(melee)
        print('\nWhite hits make %.1f threat per point of damage, so the raw threat looks %s.' % (
            m, 'scaled by 100 (Classic threat x 100)' if 50 < m < 200 else 'unscaled' if 0.5 < m < 2 else 'unusual'))


if __name__ == '__main__':
    main()
