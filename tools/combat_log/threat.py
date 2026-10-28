"""Threat per ability, from a combat log and the threat BayovCore recorded during it.

The combat log has no threat. BayovCore's Threat module (bayovui/addons/BayovCore/Modules/
Threat.lua in the WoW repo) records the raw threat the game reports, for us, our pet and two
group members on every mob we fight, and wow-combat-logs' archive.py keeps each log's fights
next to it as <log>.threat.jsonl.xz, plus <log>.threat.2.jsonl.xz and on for fights that
reached the game's file later (its docstring has the format). The addon writes only the
numbers. This lines them up with the log and works out which actions moved them.

    python3 -I tools/combat_log/threat.py <log> [threat file ...] [--who NAME] [--rows]

<log> is the combat log, plain, .xz or .gz. Without threat files we take every
<log>.threat*.jsonl.xz next to it, else in ~/wow-combat-logs/<client>/.

How it matches:

1. Actions. One press can write several lines in the same moment (Sinister Strike's cast,
   its damage and its combo point; a buff's cast and an aura on each party member). A
   player's lines of one spell on one target within ACTION_GAP sec are one action, so an AoE
   is an action on each mob it hits. A cast followed by that spell's damage or miss within
   CAST_LAND sec is the same action (Fireball's flight); a cast with no target (Thunder Clap)
   whose hits follow is dropped, since the hits carry its threat. Lines of a player's totems
   and guardians (SPELL_SUMMON, anywhere in the log) are the player's. An energize belongs to
   whoever gains it, since the gain is what threatens; combo points threaten nobody. Auras a
   player gets without casting anything (procs: Flurry, Clearcasting) threaten nobody either.
2. Which actions can move threat on a mob: the player's actions on that mob, and their
   actions not aimed at a mob (heals, energizes, buffs, shouts), which the game spreads over
   every mob they're on. Actions on other mobs don't move it.
3. The lag. The server sends a threat update a moment after the action it counts, and the
   recorded clock is to a frame (`rough` fights to a second). Each rise in a player's
   threat covers the actions between their previous reading and this one, both moved back
   by the lag. We try every lag from -1 to LAG_MAX sec (-2 to LAG_MAX + 1 when rough), in
   10 ms steps, and keep the middle of the run of lags that make the most rises cover
   exactly one action. A fight with fewer than MIN_RISES rises takes the log's lag, fitted
   over all its exact-clock fights together.
4. A rise with exactly one action is "clean" and gives that action's threat. Falls (the mob
   died or reset, Fade, Feign Death, a knockback's threat cut) and drops off the list (-1)
   are counted but never a sample. A player's first reading on a mob counts from 0 only when
   they did nothing to that mob in the LOOKBACK sec before its window, since a fight cut by a
   /reload, or a pick who pulled before we saw the mob, starts with threat already there.
5. Per player and ability, the clean samples: how many, the median threat, and for damage the
   median threat per point of damage (the amount, which leaves out what was resisted, blocked
   or absorbed). An action not aimed at a mob is split over the mobs the player is on, so its
   sample is the rise times that count ("x mobs"), once per action. The game's raw threat is
   Classic's threat times 100, so we say which scale the white hits point to.
"""
import argparse
import bisect
import datetime
import glob
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
# Lines of one spell from one player on one target this close together are one action.
ACTION_GAP = 0.05
# A cast whose damage or miss lands within this long is one action with it.
CAST_LAND = 3.0
# The lags we try go up to this far.
LAG_MAX = 1.5
# A fight with fewer rises than this takes the log's lag.
MIN_RISES = 10
# How far back we look for a player's earlier actions on a mob, before their first reading.
LOOKBACK = 120
# The most rises a lag is fitted on.
FIT_SAMPLE = 3000

# The lines that can move threat.
THREATENS = ('SWING_DAMAGE', 'SWING_MISSED', 'RANGE_DAMAGE', 'RANGE_MISSED', 'SPELL_DAMAGE',
             'SPELL_MISSED', 'SPELL_PERIODIC_DAMAGE', 'SPELL_PERIODIC_MISSED', 'SPELL_CAST_SUCCESS',
             'SPELL_AURA_APPLIED', 'SPELL_AURA_REFRESH', 'SPELL_AURA_APPLIED_DOSE', 'DAMAGE_SHIELD',
             'DAMAGE_SHIELD_MISSED', 'SPELL_HEAL', 'SPELL_PERIODIC_HEAL', 'SPELL_ENERGIZE',
             'SPELL_PERIODIC_ENERGIZE', 'SPELL_DRAIN', 'SPELL_PERIODIC_DRAIN', 'SPELL_LEECH',
             'SPELL_PERIODIC_LEECH', 'SPELL_DISPEL', 'SPELL_INTERRUPT')
LANDS = ('SPELL_DAMAGE', 'SPELL_MISSED', 'RANGE_DAMAGE', 'RANGE_MISSED')
AURAS = ('SPELL_AURA_APPLIED', 'SPELL_AURA_REFRESH', 'SPELL_AURA_APPLIED_DOSE')
ENERGIZES = ('SPELL_ENERGIZE', 'SPELL_PERIODIC_ENERGIZE')
COMBO_POINTS = '4'


def threat_paths(log):
    """Every threat file of a log: next to it, else in ~/wow-combat-logs/<client>/."""
    base = os.path.basename(log)
    for ext in ('.txt.xz', '.txt.gz', '.txt'):
        if base.endswith(ext):
            base = base[:-len(ext)]
            break
    pattern = glob.escape(base) + '.threat*.jsonl.xz'
    found = glob.glob(os.path.join(glob.escape(os.path.dirname(log) or '.'), pattern))
    if not found:
        found = glob.glob(os.path.join(glob.escape(os.path.expanduser('~/wow-combat-logs')), '*', pattern))
    return sorted(found)


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
    """What a line adds to its action's name: a miss type, a tick, a heal."""
    kind = e.ev.split('_', 1)[1]
    if kind.endswith('MISSED'):
        try:
            return cl.miss(e).lower()
        except IndexError:
            return 'miss'
    return {'PERIODIC_DAMAGE': 'tick', 'HEAL': 'heal', 'PERIODIC_HEAL': 'heal tick', 'ENERGIZE': 'energize',
            'PERIODIC_ENERGIZE': 'energize', 'CAST_SUCCESS': 'cast', 'AURA_APPLIED': 'aura',
            'AURA_REFRESH': 'aura', 'AURA_APPLIED_DOSE': 'aura stack', 'DISPEL': 'dispel',
            'INTERRUPT': 'interrupt', 'DRAIN': 'drain', 'PERIODIC_DRAIN': 'drain',
            'LEECH': 'leech', 'PERIODIC_LEECH': 'leech'}.get(kind)


class Action:
    """One press: a player's lines of one spell on one target at one moment."""

    def __init__(self, e, by, key):
        self.t, self.lines, self.by, self.key = e.t, [e], by, key
        self.mob = e.dstGUID if is_mob(e) else None

    def spell(self):
        return self.key[1]

    def label(self):
        e = self.lines[0]
        damage = [x for x in self.lines if amount(x) is not None]
        main = damage[0] if damage else next((x for x in self.lines if x.ev in LANDS), e)
        if self.spell() == 'melee':
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


def summons_of(evs):
    """Summoned GUID -> (owner GUID, name), from every SPELL_SUMMON in the log."""
    return {e.dstGUID: (e.srcGUID, e.dst) for e in evs if e.ev == 'SPELL_SUMMON'}


def actions_of(evs, guid, summons, players):
    """A player's actions (theirs, their summons' and the energizes they gain), by the mob
    they're on (None for the actions not aimed at a mob). `players` are the GUIDs the fight
    follows, whose summons (a pet) are theirs and not their owner's."""
    acts = []
    last = {}
    for e in evs:
        if e.ev not in THREATENS:
            continue
        if e.ev in ENERGIZES:
            # The gain threatens, so it's the receiver's; combo points threaten nobody.
            if e.dstGUID != guid or (len(e.f) > 2 and e.f[-2] == COMBO_POINTS):
                continue
            by = None
        elif e.srcGUID == guid:
            by = None
        else:
            owner = summons.get(e.srcGUID)
            if not owner or owner[0] != guid or e.srcGUID in players:
                continue
            by = owner[1] or 'summon'
        spell = 'melee' if e.ev.startswith('SWING') else (e.spellId, e.spell)
        target = e.dstGUID if is_mob(e) else None
        k = (e.srcGUID, spell, target)
        a = last.get(k)
        # A cast is joined by its own landing within CAST_LAND, anything else within ACTION_GAP.
        landing = a is not None and e.ev in LANDS and any(x.ev == 'SPELL_CAST_SUCCESS' for x in a.lines) \
            and not any(x.ev in LANDS for x in a.lines) and e.t - a.t <= CAST_LAND
        if a is not None and (e.t - a.lines[-1].t <= ACTION_GAP or landing):
            a.lines.append(e)
            if landing:
                a.t = e.t  # the threat comes when it lands
            continue
        a = Action(e, by, k)
        last[k] = a
        acts.append(a)
    kept = []
    for a in acts:
        casts = any(x.ev == 'SPELL_CAST_SUCCESS' for x in a.lines)
        if a.mob is None:
            # A proc's aura on a friend or ourselves, with no cast: no threat.
            a.lines = [x for x in a.lines if casts or x.ev not in AURAS]
            if not a.lines:
                continue
            # An untargeted cast whose hits follow (an AoE): the hits carry the threat.
            if all(x.ev == 'SPELL_CAST_SUCCESS' for x in a.lines) and any(
                    b.mob and b.key[:2] == a.key[:2] and 0 <= b.t - a.t <= CAST_LAND for b in acts):
                continue
        kept.append(a)
    kept.sort(key=lambda a: a.t)
    on = defaultdict(list)
    for a in kept:
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
    """(from, to, delta, dropped, first) for each change in a series."""
    prev = None
    first = True
    for t, v, dropped in s.points:
        if prev is not None and v != prev[1]:
            yield prev[0], t, v - prev[1], dropped, first
            first = False
        prev = (t, v)


class Lines:
    """A sorted list of actions, cut by time with bisect."""

    def __init__(self, acts):
        self.acts = sorted(acts, key=lambda a: a.t)
        self.ts = [a.t for a in self.acts]

    def between(self, a, b):
        return self.acts[bisect.bisect_right(self.ts, a):bisect.bisect_right(self.ts, b)]

    def count(self, a, b):
        return bisect.bisect_right(self.ts, b) - bisect.bisect_right(self.ts, a)


def rises(series, lines_for):
    """(lines, from, to) of each rise worth matching."""
    out = []
    for s in series.values():
        lines = lines_for(s)
        for a, b, d, dropped, first in changes(s):
            if d > 0 and not dropped:
                out.append((lines, a, b))
    return out


def fit_lag(found, rough):
    """The middle of the run of lags that make the most rises cover exactly one action, and
    how many it makes so."""
    if not found:
        return None, 0
    # An even sample is as good a fit as every rise, and a raid log has tens of thousands.
    if len(found) > FIT_SAMPLE:
        found = found[::-(-len(found) // FIT_SAMPLE)]
    lo, hi = (-2.0, LAG_MAX + 1.0) if rough else (-1.0, LAG_MAX)
    scores = []
    for step in range(int(round(lo * 100)), int(round(hi * 100)) + 1):
        lag = step / 100.0
        scores.append((sum(1 for lines, a, b in found if lines.count(a - lag, b - lag) == 1), lag))
    top = max(sc for sc, _ in scores)
    best = [lag for sc, lag in scores if sc == top]
    return best[len(best) // 2], top


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


class Fight:
    """A fight with its series and each player's actions, for matching."""

    def __init__(self, fight, evs, times, summons):
        self.fight = fight
        self.series = series_of(fight)
        self.start = epoch(fight['at'])
        self.rough = bool(fight.get('rough'))
        end = max((p[0] for s in self.series.values() for p in s.points), default=self.start)
        span = LAG_MAX + (1.0 if self.rough else 0.0)
        near = evs[bisect.bisect_left(times, self.start - PULL - span - LOOKBACK):
                   bisect.bisect_right(times, end + span)]
        players = {w['guid'] for w in fight['who']}
        self.acts = {w['guid']: actions_of(near, w['guid'], summons, players) for w in fight['who']}
        self.cache = {}

    def lines_for(self, s):
        key = (s.who['guid'], s.mob['guid'])
        if key not in self.cache:
            on = self.acts.get(s.who['guid'], {})
            self.cache[key] = Lines(on.get(s.mob['guid'], []) + on.get(None, []))
        return self.cache[key]

    def mob_lines(self, s):
        return Lines(self.acts.get(s.who['guid'], {}).get(s.mob['guid'], []))


def main():
    ap = argparse.ArgumentParser(description=__doc__.split('\n')[0])
    ap.add_argument('log')
    ap.add_argument('threat', nargs='*')
    ap.add_argument('--who', help='only this player (name)')
    ap.add_argument('--rows', action='store_true', help='print every change and the actions it covers')
    args = ap.parse_args()

    paths = args.threat or threat_paths(args.log)
    if not paths:
        sys.exit('no threat file for %s (archive.py writes <log>.threat.jsonl.xz)' % args.log)
    fights = [f for p in paths for f in load_fights(p)]
    fights.sort(key=lambda f: f['at'])
    evs = load_log(args.log)
    evs.sort(key=lambda e: e.t)
    times = [e.t for e in evs]
    summons = summons_of(evs)
    print('%s: %d fights, %d lines in the log' % (', '.join(os.path.basename(p) for p in paths),
                                                    len(fights), len(evs)))

    matched = [Fight(f, evs, times, summons) for f in fights if f.get('readings')]
    # The log's lag, over every exact-clock fight, for the fights with few rises.
    pooled = [r for m in matched if not m.rough for r in rises(m.series, m.lines_for)]
    log_lag, _ = fit_lag(pooled, False)

    # (who guid, ability) -> list of (threat, damage or None)
    samples = defaultdict(list)
    names = {}
    mixed, falls, unsure = defaultdict(int), defaultdict(int), defaultdict(int)
    for m in matched:
        fight, series = m.fight, m.series
        found = rises(series, m.lines_for)
        lag, top = fit_lag(found, m.rough)
        how = 'fitted'
        if (lag is None or len(found) < MIN_RISES) and not m.rough and log_lag is not None:
            lag, how = log_lag, "the log's"
        if lag is None:
            lag, how = 0.0, 'none to fit'
        who = ', '.join('%s (%s)' % (w.get('name') or '?', (w.get('class') or w.get('kind') or '?').lower())
                        for w in fight['who'])
        print('\n%s  %s, %d readings, %d mobs, lag %+.2f sec (%s, %d of %d rises clean)%s%s' % (
            fight['at'], fight.get('zone') or '?', len(fight['readings']), len(fight['mobs']), lag, how,
            top, len(found), ', %d hidden' % fight['secret'] if fight.get('secret') else '',
            ', rough clock' if m.rough else ''))
        print('  ' + who)
        used = set()
        for s in series.values():
            g = s.who['guid']
            names[g] = '%s (%s)' % (s.who.get('name') or '?', (s.who.get('class') or s.who.get('kind') or '?').lower())
            if args.who and s.who.get('name') != args.who:
                continue
            lines = m.lines_for(s)
            for a, b, d, dropped, first in changes(s):
                got = lines.between(a - lag, b - lag)
                if args.rows:
                    print('    %s %s on %s: %+g%s  %s' % (
                        datetime.datetime.fromtimestamp(b).strftime('%H:%M:%S.%f')[:-3],
                        s.who.get('name'), s.mob.get('name'), d, ' (dropped)' if dropped else '',
                        ', '.join(x.label() + ('' if x.damage() is None else ' %d' % x.damage()) for x in got)
                        or '(no action)'))
                if d <= 0 or dropped:
                    falls[g] += 1
                    continue
                if first and m.mob_lines(s).count(a - lag - LOOKBACK, a - lag):
                    # They hit this mob before the fight's first reading: the threat they
                    # had then isn't 0.
                    unsure[g] += 1
                    continue
                if len(got) != 1:
                    if got:
                        mixed[g] += 1
                    continue
                act = got[0]
                if act.mob is None:
                    if id(act) in used:
                        continue
                    used.add(id(act))
                    k = mobs_on(series, g, b)
                    samples[(g, act.label() + ' x mobs')].append((d * k, None))
                else:
                    samples[(g, act.label())].append((d, act.damage()))

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
    for g in sorted(set(mixed) | set(falls) | set(unsure), key=lambda g: names.get(g, g)):
        print('  %s: %d rises covered more than one action, %d falls and drops, %d first readings '
              'after earlier hits (not used)' % (names.get(g, g), mixed[g], falls[g], unsure[g]))
    if melee:
        mid = statistics.median(melee)
        print('\nWhite hits make %.1f threat per point of damage, so the raw threat looks %s.' % (
            mid, 'scaled by 100 (Classic threat x 100)' if 50 < mid < 200 else 'unscaled' if 0.5 < mid < 2
            else 'unusual'))


if __name__ == '__main__':
    main()
