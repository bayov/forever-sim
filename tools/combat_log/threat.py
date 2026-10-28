"""Threat per ability, from a combat log and the threat BayovCore recorded during it.

The combat log has no threat. BayovCore's Threat module (bayovui/addons/BayovCore/Modules/
Threat.lua in the WoW repo) records the raw threat the game reports, for us, our pet and two
group members on every mob we fight, and wow-combat-logs' archive.py keeps each log's fights
next to it as <log>.threat.jsonl.xz (its docstring has the format). The addon writes only the
numbers. This lines them up with the log and works out which actions moved them.

    python3 -I tools/combat_log/threat.py <log> [threat file] [--who NAME] [--rows]

<log> is the combat log, plain, .xz or .gz. Without a threat file we look for
<log>.threat.jsonl.xz next to it, then in ~/wow-combat-logs/<client>/.

How it matches:

1. Actions. One press can write several lines in the same moment (Sinister Strike's cast,
   its damage and its combo point; a buff's cast and an aura on each party member). A
   player's lines of one spell within ACTION_GAP sec are one action. A cast followed by that
   spell's damage or miss within CAST_LAND sec is the same action (Fireball's flight), and
   combo point gains move no threat, so neither counts on its own. Lines from a player's
   totems and guardians (SPELL_SUMMON) are the player's.
2. Which actions can move threat on a mob: the player's hits, misses, casts and auras on that
   mob, and everything they do that isn't aimed at a mob (heals, energizes, buffs, shouts),
   which the game spreads over every mob they're on. Actions on other mobs don't move it.
3. The lag. The server sends a threat update a moment after the action it counts, and the
   recorded clock is to a frame (`rough` fights to a second). For each rise in a player's
   threat we take the time back to their latest action that could have moved it, and use the
   median as the fight's lag.
4. Each rise in a player's threat on a mob covers the actions between their previous reading
   and this one, shifted by the lag. A rise with exactly one action is "clean" and gives that
   action's threat. Falls (the mob died or reset, Fade, Feign Death, a knockback's threat
   cut) and drops off the list (-1) are counted but never a sample.
5. Per player and ability, the clean samples: how many, the median threat, and for damage the
   median threat per point of damage (the amount, which leaves out what was resisted, blocked
   or absorbed). An action not aimed at a mob is split over the mobs the player is on, so its
   samples are multiplied back by that count ("x mobs"). The game's raw threat is Classic's
   threat times 100, so we say which scale the white hits point to.

Threat starts at 0, so a player's first reading counts from the pull (the actions up to PULL
sec before the fight's first reading). A -1 reading is a player dropping off the mob's list.
Their threat is 0 after it.
"""
import argparse
import bisect
import datetime
import gzip
import json
import lzma
import os
import shutil
import statistics
import sys
import tempfile
from collections import defaultdict

sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
import cl  # noqa: E402

# How long before a fight's first reading its first actions can be.
PULL = 3
# Lines of one spell from one player this close together are one action.
ACTION_GAP = 0.05
# A cast whose damage or miss lands within this long is one action with it.
CAST_LAND = 3.0
# The furthest back an action can be from the rise it caused, for the lag.
LAG_MAX = 1.5

# The lines of a player that can move threat.
THREATENS = ('SWING_DAMAGE', 'SWING_MISSED', 'RANGE_DAMAGE', 'RANGE_MISSED', 'SPELL_DAMAGE',
             'SPELL_MISSED', 'SPELL_PERIODIC_DAMAGE', 'SPELL_PERIODIC_MISSED', 'SPELL_CAST_SUCCESS',
             'SPELL_AURA_APPLIED', 'SPELL_AURA_REFRESH', 'SPELL_AURA_APPLIED_DOSE', 'DAMAGE_SHIELD',
             'DAMAGE_SHIELD_MISSED', 'SPELL_HEAL', 'SPELL_PERIODIC_HEAL', 'SPELL_ENERGIZE',
             'SPELL_PERIODIC_ENERGIZE', 'SPELL_DISPEL', 'SPELL_INTERRUPT')
LANDS = ('SPELL_DAMAGE', 'SPELL_MISSED', 'RANGE_DAMAGE', 'RANGE_MISSED')
COMBO_POINTS = '4'


def threat_path(log):
    base = os.path.basename(log)
    for ext in ('.txt.xz', '.txt.gz', '.txt'):
        if base.endswith(ext):
            base = base[:-len(ext)]
            break
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


def load_log(path):
    """cl.load, with a .gz log unpacked to a temporary file first."""
    if not path.endswith('.gz'):
        return cl.load(path)
    with tempfile.NamedTemporaryFile('wb', suffix='.txt', delete=False) as out, gzip.open(path, 'rb') as f:
        shutil.copyfileobj(f, out)
    try:
        return cl.load(out.name)
    finally:
        os.remove(out.name)


def load_fights(path):
    opener = lzma.open if path.endswith('.xz') else open
    with opener(path, 'rt', encoding='utf-8') as f:
        return [json.loads(line) for line in f if line.strip()]


def epoch(at):
    """A fight's `at` (local wall clock) as the same kind of number cl.Ev.t is."""
    return datetime.datetime.strptime(at, '%Y-%m-%d %H:%M:%S.%f').timestamp()


def is_mob(e):
    """Whether a line is aimed at a mob: an NPC we can fight (hostile or neutral)."""
    if not e.dstGUID.startswith(('Creature-', 'Vehicle-')) or len(e.f) < 8:
        return False
    try:
        flags = int(e.f[7], 16)
    except ValueError:
        return False
    return flags & 0x60 != 0 and flags & 0x100 == 0


def amount(e):
    if e.ev in ('SWING_DAMAGE', 'RANGE_DAMAGE', 'SPELL_DAMAGE', 'SPELL_PERIODIC_DAMAGE', 'DAMAGE_SHIELD'):
        try:
            return cl.dmg(e)['amount']
        except (IndexError, ValueError):
            return None
    return None


def part(e):
    """What a line adds to its action's name: Melee, a miss type, a tick, a heal."""
    kind = e.ev.split('_', 1)[1]
    if kind.endswith('MISSED'):
        try:
            return cl.miss(e).lower()
        except IndexError:
            return 'miss'
    return {'PERIODIC_DAMAGE': 'tick', 'HEAL': 'heal', 'PERIODIC_HEAL': 'heal tick', 'ENERGIZE': 'energize',
            'PERIODIC_ENERGIZE': 'energize', 'CAST_SUCCESS': 'cast', 'AURA_APPLIED': 'aura',
            'AURA_REFRESH': 'aura', 'AURA_APPLIED_DOSE': 'aura stack', 'DISPEL': 'dispel',
            'INTERRUPT': 'interrupt'}.get(kind)


class Action:
    """One press: a player's lines of one spell at one moment."""

    def __init__(self, e, by):
        self.t, self.lines, self.by = e.t, [e], by
        self.key = 'melee' if e.ev.startswith('SWING') else (e.spellId, e.spell)
        self.mob = e.dstGUID if is_mob(e) else None

    def label(self):
        e = self.lines[0]
        damage = [x for x in self.lines if amount(x) is not None]
        main = damage[0] if damage else next((x for x in self.lines if x.ev in LANDS), e)
        if self.key == 'melee':
            name = 'Melee'
        elif main.ev.startswith('RANGE'):
            name = 'Ranged ' + (main.spell or '')
        else:
            name = main.spell or '?'
        if self.by:
            name = '%s: %s' % (self.by, name)
        extra = part(main)
        if extra and not (damage and extra in ('cast', 'aura')):
            name += ' (%s)' % extra
        return name

    def damage(self):
        n = [amount(x) for x in self.lines if amount(x) is not None]
        return sum(n) if n else None


def actions_of(evs, guid):
    """A player's actions (theirs and their summons'), by time, and by the mob they're on (None
    for the actions not aimed at a mob)."""
    summons = {e.dstGUID: e.dst for e in evs if e.ev == 'SPELL_SUMMON' and e.srcGUID == guid}
    acts = []
    last = {}
    for e in evs:
        if e.ev not in THREATENS:
            continue
        mine = e.srcGUID == guid
        if not mine and e.srcGUID not in summons:
            continue
        if e.ev in ('SPELL_ENERGIZE', 'SPELL_PERIODIC_ENERGIZE') and len(e.f) > 2 and e.f[-2] == COMBO_POINTS:
            continue
        key = 'melee' if e.ev.startswith('SWING') else (e.spellId, e.spell)
        k = (e.srcGUID, key)
        a = last.get(k)
        # A cast is joined by its own landing within CAST_LAND, anything else within ACTION_GAP.
        landing = a is not None and e.ev in LANDS and any(x.ev == 'SPELL_CAST_SUCCESS' for x in a.lines) \
            and not any(x.ev in LANDS for x in a.lines) and e.t - a.t <= CAST_LAND
        if a is not None and (e.t - a.lines[-1].t <= ACTION_GAP or landing):
            a.lines.append(e)
            if landing:
                a.t = e.t  # the threat comes when it lands
            if a.mob is None and is_mob(e):
                a.mob = e.dstGUID
            continue
        a = Action(e, None if mine else summons.get(e.srcGUID) or 'summon')
        last[k] = a
        acts.append(a)
    acts.sort(key=lambda a: a.t)
    on = defaultdict(list)
    for a in acts:
        on[a.mob].append(a)
    return on


class Series:
    """One player's readings on one mob: [(time, threat, dropped)]."""

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
            out[key].points.append((start - PULL, 0, False))
        dropped = r[3] == -1
        value = 0 if dropped else r[6]
        if value is None:  # a value the archive couldn't read
            continue
        out[key].points.append((t, value, dropped))
    return out


def changes(s):
    """(from, to, delta, dropped) for each change in a series."""
    prev = None
    for t, v, dropped in s.points:
        if prev is not None and v != prev[1]:
            yield prev[0], t, v - prev[1], dropped
        prev = (t, v)


class Lines:
    """A sorted list of actions, cut by time with bisect."""

    def __init__(self, acts):
        self.acts = sorted(acts, key=lambda a: a.t)
        self.ts = [a.t for a in self.acts]

    def between(self, a, b):
        return self.acts[bisect.bisect_right(self.ts, a):bisect.bisect_right(self.ts, b)]

    def before(self, b):
        i = bisect.bisect_right(self.ts, b)
        return self.acts[i - 1] if i else None


def fit_lag(series, lines_for, rough):
    """The median time from a rise back to the player's latest action that could have caused it."""
    most = LAG_MAX + (1.0 if rough else 0.0)
    gaps = []
    for s in series.values():
        lines = lines_for(s)
        for a, b, d, dropped in changes(s):
            if d <= 0 or dropped:
                continue
            act = lines.before(b + (1.0 if rough else 0.0))
            if act and -1.0 <= b - act.t <= most:
                gaps.append(b - act.t)
    return (statistics.median(gaps), len(gaps)) if gaps else (0.0, 0)


def mobs_on(series, guid, t):
    """How many mobs the player has threat on at t."""
    n = 0
    for s in series.values():
        if s.who['guid'] != guid:
            continue
        v = 0
        for pt, value, _ in s.points:
            if pt > t:
                break
            v = value
        n += v > 0
    return max(n, 1)


def main():
    ap = argparse.ArgumentParser(description=__doc__.split('\n')[0])
    ap.add_argument('log')
    ap.add_argument('threat', nargs='?')
    ap.add_argument('--who', help='only this player (name)')
    ap.add_argument('--rows', action='store_true', help='print every change and the actions it covers')
    args = ap.parse_args()

    tpath = args.threat or threat_path(args.log)
    if not tpath:
        sys.exit('no threat file for %s (archive.py writes <log>.threat.jsonl.xz)' % args.log)
    fights = load_fights(tpath)
    evs = load_log(args.log)
    evs.sort(key=lambda e: e.t)
    times = [e.t for e in evs]
    print('%s: %d fights, %d lines in the log' % (os.path.basename(tpath), len(fights), len(evs)))

    # (who guid, ability) -> list of (threat, damage or None)
    samples = defaultdict(list)
    names = {}
    mixed, falls = defaultdict(int), defaultdict(int)
    for fight in fights:
        series = series_of(fight)
        if not series:
            continue
        start = epoch(fight['at'])
        end = max(p[0] for s in series.values() for p in s.points)
        span = LAG_MAX + (1.0 if fight.get('rough') else 0.0)
        near = evs[bisect.bisect_left(times, start - PULL - span):bisect.bisect_right(times, end + span)]
        acts = {w['guid']: actions_of(near, w['guid']) for w in fight['who']}
        cache = {}

        def lines_for(s):
            key = (s.who['guid'], s.mob['guid'])
            if key not in cache:
                on = acts.get(s.who['guid'], {})
                cache[key] = Lines(on.get(s.mob['guid'], []) + on.get(None, []))
            return cache[key]

        lag, matched = fit_lag(series, lines_for, fight.get('rough'))
        who = ', '.join('%s (%s)' % (w.get('name') or '?', (w.get('class') or w.get('kind') or '?').lower())
                        for w in fight['who'])
        print('\n%s  %s, %d readings, %d mobs, lag %+.3f sec (from %d rises)%s%s' % (
            fight['at'], fight.get('zone') or '?', len(fight['readings']), len(fight['mobs']), lag, matched,
            ', %d hidden' % fight['secret'] if fight.get('secret') else '',
            ', rough clock' if fight.get('rough') else ''))
        print('  ' + who)
        for s in series.values():
            g = s.who['guid']
            names[g] = '%s (%s)' % (s.who.get('name') or '?', (s.who.get('class') or s.who.get('kind') or '?').lower())
            if args.who and s.who.get('name') != args.who:
                continue
            lines = lines_for(s)
            for a, b, d, dropped in changes(s):
                got = lines.between(a - lag, b - lag)
                if args.rows:
                    print('    %s %s on %s: %+g%s  %s' % (
                        datetime.datetime.fromtimestamp(b).strftime('%H:%M:%S.%f')[:-3],
                        s.who.get('name'), s.mob.get('name'), d, ' (dropped)' if dropped else '',
                        ', '.join(x.label() + ('' if x.damage() is None else ' %d' % x.damage()) for x in got)
                        or '(no action)'))
                if d <= 0 or dropped:
                    falls[g] += 1
                elif len(got) == 1:
                    act = got[0]
                    if act.mob is None:
                        k = mobs_on(series, g, b)
                        samples[(g, act.label() + ' x mobs')].append((d * k, None))
                    else:
                        samples[(g, act.label())].append((d, act.damage()))
                elif got:
                    mixed[g] += 1

    print('\nClean samples (one action between two readings):')
    melee = []
    for (g, ability), got in sorted(samples.items(), key=lambda kv: (names.get(kv[0][0], ''), kv[0][1])):
        if args.who and not names.get(g, '').startswith(args.who + ' '):
            continue
        threat = statistics.median(d for d, _ in got)
        per = [d / a for d, a in got if a]
        if ability == 'Melee' and per:
            melee.extend(per)
        print('  %-24s %-40s %4d  threat %10.1f%s' % (
            names.get(g, g), ability, len(got), threat,
            '  per damage %.2f' % statistics.median(per) if per else ''))
    for g in sorted(set(mixed) | set(falls), key=lambda g: names.get(g, g)):
        print('  %s: %d rises covered more than one action, %d falls and drops (not used)'
              % (names.get(g, g), mixed[g], falls[g]))
    if melee:
        m = statistics.median(melee)
        print('\nWhite hits make %.1f threat per point of damage, so the raw threat looks %s.' % (
            m, 'scaled by 100 (Classic threat x 100)' if 50 < m < 200 else 'unscaled' if 0.5 < m < 2 else 'unusual'))


if __name__ == '__main__':
    main()
