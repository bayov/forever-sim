#!/usr/bin/env python3
"""Rage from melee hits taken by warriors, by stance, from Forever combat logs.

  rage_taken.py [--more] [--shift] LOG.txt.xz [LOG.txt.xz ...]

--more adds blocks, armor and warriors seen in two stances. orchestration/rage_taken_by_stance.md
has what it found.

One sample is a melee hit a warrior took from a creature, when nothing that could move rage happened
between it and the hit before (no cast, no swing of ours, no aura or energize event on us, no other
hit). Its rage is the change in the rage bar between the two hits' snapshots. We give the sample to
the hit we are at, and with --shift to the one before, to see which one the snapshot already holds.

formula is core.ForeverDamageTakenRage for the hit, without any stance: 2.5 x (1 - A) x raw / c.
taken is what the stance did to the damage: amount / raw / what armor leaves. Only plain hits (no
crit, block, absorb, crushing blow).
"""
import collections, lzma, random, statistics, sys

WAR = {100,78,284,285,1608,11564,11565,772,6546,6547,6548,2457,71,2458,6673,5242,6192,7384,6343,8198,
       8204,7386,7405,6572,2687,20252,1680,5308,1464,2565}
STANCE = {'2457': 'Battle', '71': 'Defensive', '2458': 'Berserker'}
ROWS = [(42, .2), (44, .3), (46, .36), (50, .38), (63, .4)]


def armor_share(l):
    if l <= 42:
        return .2
    for (a, x), (b, y) in zip(ROWS, ROWS[1:]):
        if l <= b:
            return x + (y - x) * (l - a) / (b - a)
    return .4


def formula(l):
    return 2.5 * (1 - armor_share(l)) / (0.0091107836 * l * l + 3.225598133 * l + 4.2652911)


def ts(s):
    h, m, x = s.split(' ')[1].split(':')
    return int(h) * 3600 + int(m) * 60 + float(x)


def read(path, warriors, out, shifted):
    level = {}
    prev = {}        # player -> (t, rage, sample dict) of the last hit taken
    clean = {}
    stance = {}
    with lzma.open(path, 'rt', errors='replace') as fh:
        for line in fh:
            if 'Player-' not in line:
                continue
            i = line.find('  ')
            p = line[i + 2:].rstrip('\n').split(',')
            ev = p[0]
            if len(p) < 10:
                continue
            src, dst = p[1], p[5]
            if ev in ('SPELL_CAST_SUCCESS', 'SPELL_AURA_APPLIED', 'SPELL_AURA_REMOVED') and p[9].isdigit():
                if ev == 'SPELL_CAST_SUCCESS' and src.startswith('Player') and int(p[9]) in WAR:
                    warriors.add(src)
                g = dst if ev != 'SPELL_CAST_SUCCESS' else src
                if p[9] in STANCE:
                    if ev == 'SPELL_AURA_REMOVED':
                        stance[g] = None if stance.get(g) == STANCE[p[9]] else stance.get(g)
                    else:
                        stance[g] = STANCE[p[9]]
            if ev == 'SWING_DAMAGE' and src.startswith('Creature') and len(p) > 37 and p[27].lstrip('-').isdigit():
                level[src] = int(p[27])   # SWING_DAMAGE's advanced block is the attacker's
            if (ev == 'SWING_DAMAGE_LANDED' and dst.startswith('Player') and src.startswith('Creature')
                    and len(p) > 37 and p[9] == dst and p[19] == '1'):
                try:
                    t = ts(line[:i]); rage = float(p[20]) / 10; cap = int(p[21]) / 10
                    s = dict(t=t, lvl=int(p[27]), ml=level.get(src), amount=int(p[28]), raw=int(p[29]), blocked=int(p[33]),
                             absorbed=int(p[34]), crit=p[35], crush=p[37] if len(p) > 37 else '', armor=int(p[15]),
                             stance=stance.get(dst), log=path.split('/')[-1], who=dst)
                except ValueError:
                    continue
                a = prev.get(dst)
                if a and clean.get(dst) and t - a[0] < 3.5 and rage < cap - 1:
                    out.append((dst, rage - a[1], s))
                    shifted.append((dst, rage - a[1], a[2]))
                prev[dst] = (t, rage, s)
                clean[dst] = True
                continue
            # Anything else on or by the player may move rage.
            if ev in ('SWING_DAMAGE', 'SPELL_AURA_REFRESH', 'SWING_DAMAGE_LANDED'):
                if ev != 'SWING_DAMAGE_LANDED' and src in clean:
                    clean[src] = False
                continue
            for g in (src, dst):
                if g in clean:
                    clean[g] = False


def med_ci(v, n=1000):
    rnd = random.Random(1)
    m = statistics.median(v)
    if len(v) < 8:
        return m, float('nan'), float('nan')
    b = sorted(statistics.median(rnd.choices(v, k=len(v))) for _ in range(n))
    return m, b[int(n * .025)], b[int(n * .975)]


def main():
    shift = '--shift' in sys.argv
    files = [a for a in sys.argv[1:] if not a.startswith('--')]
    warriors, out, shifted = set(), [], []
    for f in files:
        read(f, warriors, out, shifted)
    rows = shifted if shift else out
    rows = [(d, s) for g, d, s in rows if g in warriors and s['ml'] and s['raw'] > 20]
    plain = [(d, s) for d, s in rows if s['crit'] != '1' and s['crush'] != '1' and not s['blocked'] and not s['absorbed']]
    print(f'{len(files)} logs, {len({s["who"] for _, s in rows})} warriors with samples, {len(rows)} hits, {len(plain)} plain, '
          f'{sum(d <= 0 for d, _ in plain)} of them moved rage by 0 or less (left out below)')
    groups = collections.defaultdict(list)
    for d, s in plain:
        if d <= 0:
            continue
        dr = s['armor'] / (s['armor'] + 400 + 85 * s['ml'])
        taken = s['amount'] / s['raw'] / (1 - dr)
        ratio = d / (formula(s['ml']) * s['raw'])
        groups[(s['stance'] or 'unknown', 'all')].append((ratio, taken, s))
        groups[(s['stance'] or 'unknown', f"mob {s['ml'] // 10 * 10}s")].append((ratio, taken, s))
    print('| stance | mobs | hits | warriors | logs | rage / formula, median (95%) | damage / raw after armor | rage / formula / damage share |')
    print('|---|---|---|---|---|---|---|---|')
    for k, v in sorted(groups.items()):
        if len(v) < 8:
            continue
        m, lo, hi = med_ci([x[0] for x in v])
        print(f'| {k[0]} | {k[1]} | {len(v)} | {len({x[2]["who"] for x in v})} | {len({x[2]["log"] for x in v})} | {m:.3f} ({lo:.3f}-{hi:.3f}) | '
              f'{statistics.median(x[1] for x in v):.3f} | {statistics.median(x[0] / x[1] for x in v):.3f} |')


def more(warriors, out):
    def ratio(d, s):
        return d / (formula(s['ml']) * s['raw'])
    rows = [(d, s) for g, d, s in out if g in warriors and s['ml'] and s['raw'] > 20 and d > 0 and s['stance'] and s['crush'] != '1']
    print('\nBlocks, absorbs and crits: rage / formula on the whole hit (raw), by stance')
    print('| stance | hit | hits | rage / formula, median (95%) | amount / raw, after armor |')
    print('|---|---|---|---|---|')
    g = collections.defaultdict(list)
    for d, s in rows:
        kind = 'crit' if s['crit'] == '1' else 'blocked' if s['blocked'] else 'absorbed' if s['absorbed'] else 'plain'
        dr = s['armor'] / (s['armor'] + 400 + 85 * s['ml'])
        g[(s['stance'], kind)].append((ratio(d, s), s['amount'] / s['raw'] / (1 - dr)))
    for k, v in sorted(g.items()):
        if len(v) >= 8:
            m, lo, hi = med_ci([x[0] for x in v])
            print(f'| {k[0]} | {k[1]} | {len(v)} | {m:.3f} ({lo:.3f}-{hi:.3f}) | {statistics.median(x[1] for x in v):.3f} |')
    plain = [(d, s) for d, s in rows if s['crit'] != '1' and not s['blocked'] and not s['absorbed']]
    print('\nArmor: plain Defensive hits by our armor reduction, rage / formula')
    b = collections.defaultdict(list)
    for d, s in plain:
        if s['stance'] == 'Defensive':
            dr = s['armor'] / (s['armor'] + 400 + 85 * s['ml'])
            b[int(dr * 10)].append(ratio(d, s))
    for k, v in sorted(b.items()):
        if len(v) >= 20:
            print(f'| armor takes {k * 10}-{k * 10 + 10}% | {len(v)} hits | {statistics.median(v):.3f} |')
    print('\nSame warrior, two stances, 8+ plain hits in each: median rage / formula')
    w = collections.defaultdict(lambda: collections.defaultdict(list))
    for d, s in plain:
        w[s['who']][s['stance']].append(ratio(d, s))
    n = 0
    for who, st in w.items():
        ok = {k: statistics.median(v) for k, v in st.items() if len(v) >= 8}
        if len(ok) > 1:
            n += 1
            print(f'| warrior {n} | ' + ' | '.join(f'{k} {v:.3f} ({len(st[k])})' for k, v in sorted(ok.items())) + ' |')


warriors, out, shifted = set(), [], []
main()
if '--more' in sys.argv:
    files = [a for a in sys.argv[1:] if not a.startswith('--')]
    for f in files:
        read(f, warriors, out, shifted)
    more(warriors, out)
