#!/usr/bin/env python3
"""bench_races.py: the Tanking and Healing modes' race and racial sims, paired fight by fight (tankheal.py run).

For every tank and heal preset and every race its class can be (run_build.CLASS_RACES), on the bosses its score averages
(tanks: the six scored bosses; healers: all seven), it sims
  <race>         the preset with that race
  off            the race's base stats only: every racial left out (DASH_RACIALS_OFF=all)
  off:<part>     one racial part left out, for races with more than one part (PARTS, run_races.py's parts plus the
                 defensive ones a tank or healer can feel: Night Elf dodge)
DASH_RACIALS_OFF_CLASS keeps the switch on the preset's class, so the healing raid's tank and damage dealers keep theirs.

Every sim keeps its per-fight values (the benches' -vectors) in adaptive.py's vector store, seed 1 with labeled random
streams, so fight i of every race and racial is the same fight and a comparison's error is the paired one. Fight counts
follow adaptive.py's luck goals, put in each score's unit (GOALS, below): races get a pilot of RACE_PILOT fights, then all
of them the N the widest paired spread between races next to each other asks for; a racial left out gets a pilot of
PILOT fights against its race, then both the N its own spread asks for.

The scores and their paired errors:
  tank   a preset's score at its own weight w: 2 x SafeDPS - w x 900 x healing needed / one healer's HPS, averaged over
         the six scored bosses (tools/tankscore). Healing needed is the 95th percentile of each fight's least healing it
         lives through, so the paired error is a bootstrap: the same resampled fights for both sides, per boss.
  heal   the damage carried: the scale of all the damage at which the deaths reach their line (2% of fights with a tank
         death for tank healers, 5% with any death for raid healers), averaged over the seven bosses. Each fight is
         simmed on a lattice of scales 1.03^j around the crossing, every race and racial on the same lattice, and the
         carry is read between the two lattice scales that bracket the line (healbench's interpolation). Its paired
         error is the same bootstrap. HPS and useful HPS are at the encounter's own damage (scale 1).

Results append to $DASH_DIR/<mode>/races.jsonl, one line a (preset, race, variant), so a rerun only does what is missing;
the vectors stay in $DASH_DIR/vec/bench_<mode>/ (keyed by the request, the switches and the bench binary's hash).
"""
import functools, hashlib, json, math, os, subprocess, sys, tempfile, threading

import numpy as np

HERE = os.path.dirname(os.path.abspath(__file__))
sys.path.insert(0, HERE)
import adaptive  # noqa: E402
from dashpaths import D, RNG  # noqa: E402

# Racial parts a tank or healer can feel, by race: run_races.py's PARTS (the DPS ones) plus Night Elf dodge. The boss
# type parts (beast, insight) are left out: no tank or healing boss has a creature type.
PARTS = {
    'Human': ['spirit', 'ruthless'], 'Dwarf': ['precision', 'chastise'], 'NightElf': ['dodge', 'elune', 'shadowmeld', 'starshards'],
    'Gnome': ['mind', 'eureka'], 'Orc': ['command', 'bloodfury'], 'Undead': ['touch', 'darksacrifice'],
    'Tauren': ['endurance'], 'Troll': ['berserking', 'shadowguard'], 'SkyborneHighOrder': ['windblessed', 'leyline'], 'SkyborneWindshaper': ['windblessed'],
}
PRIEST_ONLY = {'chastise', 'starshards', 'darksacrifice', 'shadowguard'}
PET_ONLY = {'command'}

# The luck goals in each score's unit (adaptive.py's GOALS hold the DPS ones and say how these follow from them).
GOALS = adaptive.BENCH_GOALS
BOOT = 400  # bootstrap resamples: an error good to about 1 / sqrt(2 x 400), 4%
STEP = 1.03  # the healing lattice: scales 1.03^j
LINES = {'tank': (1, 0.02), 'raid': (3, 0.05)}  # heal role: (death flags that fail a fight, the line)
DEATH = 0.05  # the tank's chance of death healing needed aims at (tankbench -death)


def parts_for(cls, race):
    return [p for p in PARTS.get(race, []) if (p not in PRIEST_ONLY or cls == 'Priest') and p not in PET_ONLY]


def variants(cls, race):
    ps = parts_for(cls, race)
    return ['off'] + (['off:' + p for p in ps] if len(ps) > 1 else [])


def sha(path):
    return hashlib.sha1(open(path, 'rb').read()).hexdigest()[:12]


# ---------------------------------------------------------------- the benches

class Bench:
    """Runs one bench binary on a preset at a race with racials left out, and keeps its per-fight values."""

    def __init__(self, mode, exe, tree, req_path, encounters):
        self.mode, self.exe, self.tree, self.req_path, self.enc = mode, exe, tree, req_path, encounters
        self.binsha = sha(exe)
        self.lock = threading.Lock()

    def _run(self, stem, race, off, args, seed, count):
        req = json.load(open(self.req_path(stem)))
        req['raid']['parties'][0]['players'][0]['race'] = 'Race' + race
        cls = req['raid']['parties'][0]['players'][0]['class'][5:]
        env = dict(os.environ)
        for k in ('DASH_RACIALS_OFF', 'DASH_RACIALS_OFF_CLASS', 'DASH_NAMES'):
            env.pop(k, None)
        if off:
            env.update(DASH_RACIALS_OFF=off, DASH_RACIALS_OFF_CLASS=cls)
        with tempfile.TemporaryDirectory() as d:
            p, out = os.path.join(d, stem + '.json'), os.path.join(d, 'v.jsonl')
            json.dump(req, open(p, 'w'))
            r = subprocess.run([self.exe, '-iters', str(count), '-seed', str(seed), '-labeled', '-vectors', out] + args + [p],
                               cwd=self.tree, env=env, capture_output=True, text=True)
            if r.returncode or not os.path.exists(out):
                raise RuntimeError(f'{self.exe} {stem} {race} {off}: {r.stderr[-1500:]}')
            return [json.loads(l) for l in open(out)]

    def key(self, stem, race, off, boss, extra=None):
        req = json.load(open(self.req_path(stem)))
        req['raid']['parties'][0]['players'][0]['race'] = 'Race' + race
        return {'req': req, 'boss': boss, 'encounter': json.load(open(self.enc(boss))), 'extra': extra}, {'off': off or '', 'bin': self.binsha}

    def tank(self, stem, race, off, boss, n):
        """[(safeDps, surviveHps (inf: none), dtps, tps)] for the first n fights."""
        def sim(seed, count):
            v = self._run(stem, race, off, ['-encounter', self.enc(boss)], seed, count)[0]
            s = [x if x >= 0 else math.inf for x in v['surviveHps']]
            if len(s) != count or len(v['safeDps']) != count:
                raise RuntimeError(f'{stem} {race} {off} {boss}: {len(v["safeDps"])} fights of {count}')
            return [y for i in range(count) for y in (v['safeDps'][i], s[i], v['dtps'][i], v['tps'][i])]
        k, e = self.key(stem, race, off, boss)
        return np.array(adaptive.values('bench_tank', k, e, n, sim, width=4)).reshape(-1, 4)

    def heal(self, stem, race, off, boss, role, scale, n):
        """[(death flags, hps, useful hps, mana left)] for the first n fights at this damage scale."""
        def sim(seed, count):
            v = self._run(stem, race, off, ['-carry=false', '-role', role, '-encounter', self.enc(boss), '-scales', repr(scale)], seed, count)[0]
            return [y for i in range(count) for y in (v['deaths'][i], v['hps'][i], v['ehps'][i], v['manaEnd'][i])]
        k, e = self.key(stem, race, off, boss, {'role': role, 'scale': scale})
        return np.array(adaptive.values('bench_heal', k, e, n, sim, width=4)).reshape(-1, 4)


# ---------------------------------------------------------------- scores

def needed(surv, death=DEATH):
    """tankscore's healing needed (core.TankScoreSummary): the (1 - death) quantile of the fights' least healing it lives
    through, NaN past the death floor (more than `death` of the fights no healing saves). Along the last axis."""
    n = surv.shape[-1]
    floor = np.isinf(surv).mean(axis=-1)
    i = max(0, min(n - 1, math.ceil((1 - death) * n) - 1))
    q = np.partition(surv, i, axis=-1)[..., i]
    return np.where(floor > death, np.nan, q)


@functools.lru_cache(maxsize=4)
def _boot_idx(bosses, n):
    rng = np.random.default_rng(n)
    return {b: rng.integers(0, n, size=(BOOT, n), dtype=np.int32) for b in bosses}


def boot_idx(bosses, n):
    """The resampled fights, the same for every side of a comparison (fixed seed): {boss: (BOOT, n)}."""
    return _boot_idx(tuple(bosses), n)


class TankSide:
    """A race or racial of a tank preset on the scored bosses: {boss: (fights, 4) array}, each boss's healer HPS."""

    def __init__(self, per_boss, each, w):
        self.b, self.each, self.w = per_boss, each, w

    def n(self):
        return min(len(a) for a in self.b.values())

    def score(self, n=None, w=None):
        """{score, safeDps, healingNeeded, dtps, tps}: means over the bosses of the first n fights (score None when no
        healing saves it on a boss)."""
        n, w = n or self.n(), self.w if w is None else w
        safe = [self.b[b][:n, 0].mean() for b in self.b]
        need = [float(needed(self.b[b][:n, 1])) for b in self.b]
        sc = [2 * s - w * 900 / self.each[b] * q for b, s, q in zip(self.b, safe, need)]
        f = lambda xs: None if any(math.isnan(x) for x in xs) else float(np.mean(xs))  # noqa: E731
        return {'score': f(sc), 'safeDps': f(safe), 'healingNeeded': f(need), 'dtps': f([self.b[b][:n, 2].mean() for b in self.b]),
                'tps': f([self.b[b][:n, 3].mean() for b in self.b]), 'none': sum(math.isnan(x) for x in need)}

    def boot(self, n):
        """The score over each resample of the first n fights (NaN where a resample has no score)."""
        if getattr(self, '_boot', (None,))[0] == n:
            return self._boot[1]
        idx = boot_idx(list(self.b), n)
        out = 0
        for b, a in self.b.items():
            ii = idx[b]
            safe = a[:n, 0][ii].mean(axis=1)
            out = out + 2 * safe - self.w * 900 / self.each[b] * needed(a[:n, 1][ii])
        self._boot = (n, out / len(self.b))
        return self._boot[1]


def crossing(ks, fs, line):
    """The scale where the failure share reaches the line: linear between the two lattice scales (healbench's
    interpolation), and out along that line when a resample's share has moved past one of them."""
    (k0, k1), (f0, f1) = ks, fs
    if f1 <= f0:
        return k0 if f0 >= line else k1
    return k0 + (k1 - k0) * (line - f0) / (f1 - f0)


class HealSide:
    """A race or racial of a healing preset on the seven bosses: per boss the fights at its two lattice scales around
    the line, and at the encounter's own damage."""

    def __init__(self, role, per_boss, at_one):
        self.role, self.b, self.one = role, per_boss, at_one  # b: {boss: {scale: flags}}, one: {boss: (n, 4) array}
        self.mask, self.line = LINES[role]

    def n(self):
        return min(len(v) for pts in self.b.values() for v in pts.values())

    def _fails(self, boss, n):
        pts = sorted(self.b[boss].items())
        return [k for k, _ in pts], [((v[:n].astype(np.int64) & self.mask) != 0) for _, v in pts]

    def score(self, n=None):
        n = n or self.n()
        cs = []
        for b in self.b:
            ks, fl = self._fails(b, n)
            cs.append(crossing(ks, [f.mean() for f in fl], self.line))
        one = list(self.one.values())
        hps, ehps = float(np.mean([a[:, 1].mean() for a in one])), float(np.mean([a[:, 2].mean() for a in one]))
        return {'carry': float(np.mean(cs)), 'hps': hps, 'ehps': ehps, 'overheal': 100 * (1 - ehps / hps) if hps else None,
                'manaLeft': float(np.mean([a[:, 3].mean() for a in one]))}

    def boot(self, n):
        if getattr(self, '_boot', (None,))[0] == n:
            return self._boot[1]
        idx = boot_idx(list(self.b), n)
        out = 0
        for b in self.b:
            ks, fl = self._fails(b, n)
            f0, f1 = fl[0][idx[b]].mean(axis=1), fl[1][idx[b]].mean(axis=1)
            d = np.where(f1 > f0, f1 - f0, np.nan)
            c = ks[0] + (ks[1] - ks[0]) * (self.line - f0) / d
            c = np.where(np.isnan(d), np.where(f0 >= self.line, ks[0], ks[1]), c)
            out = out + c
        self._boot = (n, out / len(self.b))
        return self._boot[1]


def value(side, n=None):
    s = side.score(n)
    return s['score'] if isinstance(side, TankSide) else s['carry']


def pse(a, b, n):
    """The paired error of b - a over their first n fights: the spread of the difference over the same resamples."""
    ra, rb = a.boot(n), b.boot(n)
    ok = ~(np.isnan(ra) | np.isnan(rb))
    if ok.mean() < 0.9:
        return None
    return float(np.std(rb[ok] - ra[ok]))


def same(a, b, n):
    """Whether two sides' first n fights came out exactly alike: the racial does nothing on this preset."""
    xs = [(a.b[k], b.b[k]) for k in a.b] if isinstance(a, TankSide) else [(a.b[k][s], b.b[k].get(s)) for k in a.b for s in a.b[k]]
    return all(y is not None and len(x) >= n and len(y) >= n and np.array_equal(x[:n], y[:n]) for x, y in xs)


# ---------------------------------------------------------------- the walk on the healing lattice

def lattice(j):
    return round(STEP ** j, 6)


def heal_side(bench, stem, race, off, role, bosses, start, n, one_n):
    """A healer's side at n fights: per boss, the lattice scales from start[boss] (j of the lower of the preset race's
    two scales) out to the two around this side's line, so most sides of a preset compare at the very same scales.
    HPS, useful HPS and mana left at the encounter's own damage are from one_n fights (the pilot's)."""
    mask, line = LINES[role]
    per, one, js = {}, {}, {}
    for boss in bosses:
        j, pts = start[boss], {}

        def F(j):
            k = lattice(j)
            if k not in pts:
                pts[k] = bench.heal(stem, race, off, boss, role, k, n)[:, 0]
            return ((pts[k].astype(np.int64) & mask) != 0).mean()
        if F(j) >= line:
            while F(j - 1) >= line and j > -60:
                j -= 1
            j -= 1
        else:
            while F(j + 1) < line and j < 60:
                j += 1
        F(j), F(j + 1)
        per[boss], js[boss] = {lattice(j): pts[lattice(j)], lattice(j + 1): pts[lattice(j + 1)]}, j
        one[boss] = bench.heal(stem, race, off, boss, role, 1.0, one_n)
    return HealSide(role, per, one), js


# ---------------------------------------------------------------- a preset

def preset(mode, bench, stem, cls, own, races, bosses, w=None, each=None, role=None, start=None, log=print):
    """Every race of a preset and its racials left out, at the fights the goals ask for: result rows."""
    g = GOALS[mode]
    ceil = adaptive.CEIL if mode == 'tank' else g['ceil']
    rows = []

    def side(race, off, n, at=None):
        if mode == 'tank':
            return TankSide({b: bench.tank(stem, race, off, b, n) for b in bosses}, each, w), None
        return heal_side(bench, stem, race, off, role, bosses, at or start, n, adaptive.PILOT)

    # the preset race first: on the healing lattice its crossing is where every other side starts
    sides, js = {}, start
    s0, js = side(own, None, adaptive.RACE_PILOT)
    sides[own] = s0
    for rc in races:
        if rc != own:
            sides[rc] = side(rc, None, adaptive.RACE_PILOT, js)[0]
    order = sorted(races, key=lambda rc: -(value(sides[rc]) if value(sides[rc]) is not None else -1e18))
    spread = max([(pse(sides[a], sides[b], adaptive.RACE_PILOT) or 0) * math.sqrt(adaptive.RACE_PILOT) for a, b in zip(order, order[1:])] or [0.0])
    n = adaptive.fights(spread, g['race'][0], adaptive.RACE_PILOT, ceil)
    if n > adaptive.RACE_PILOT:
        sides[own], js = side(own, None, n, js)
        for rc in races:
            if rc != own:
                sides[rc] = side(rc, None, n, js)[0]
    log(f'{mode} {stem}: races at {n} fights (pilot spread {spread:.3g})')
    for rc in races:
        row = {'build': stem, 'race': rc, 'variant': 'race', 'iters': n, **sides[rc].score(n),
               'pse': {o: pse(sides[rc], sides[o], n) for o in races if o != rc}}
        if mode == 'tank':
            for k, ww in (('scoreMaxThreat', 1 / 3), ('scoreBalanced', 1.0), ('scoreMaxSurvival', 3.0)):
                row[k] = sides[rc].score(n, ww)['score']
        rows.append(row)

    # each race's racials left out, against the race
    for rc in races:
        for v in variants(cls, rc):
            off = 'all' if v == 'off' else v[4:]
            p = adaptive.PILOT
            sv = side(rc, off, p, js)[0]
            d = value(sv, p) - value(sides[rc], p) if value(sv, p) is not None and value(sides[rc], p) is not None else None
            alike = same(sides[rc], sv, p)
            if alike or d is None:
                m = p
            else:
                e = pse(sides[rc], sv, p) or 0
                floor, share = g['racial']
                m = adaptive.fights(e * math.sqrt(p), max(floor, share * abs(d)), p, ceil)
            base = sides[rc]
            if m > p:
                sv = side(rc, off, m, js)[0]
                if m > n:
                    base = side(rc, None, m, js)[0]
            k = min(m, base.n(), sv.n())
            # the race's numbers at its own fights plus the paired difference: a racial that changes nothing is the race
            race_row = next(r for r in rows if r['race'] == rc and r['variant'] == 'race')
            a, z = base.score(k), sv.score(k)
            row = {'build': stem, 'race': rc, 'variant': v, 'iters': m, 'same': alike, 'pseBase': 0.0 if alike else pse(base, sv, k)}
            for key, x in race_row.items():
                if key in a and isinstance(x, float):
                    row[key] = None if alike is False and (a[key] is None or z[key] is None) else x + ((z[key] - a[key]) if not alike else 0.0)
            if mode == 'tank':
                row['none'] = z['none']
                for key, ww in (('scoreMaxThreat', 1 / 3), ('scoreBalanced', 1.0), ('scoreMaxSurvival', 3.0)):
                    za, zz = base.score(k, ww)['score'], sv.score(k, ww)['score']
                    row[key] = race_row[key] + (zz - za) if None not in (za, zz, race_row[key]) else None
            row['diff'] = row.get('score' if mode == 'tank' else 'carry') - race_row['score' if mode == 'tank' else 'carry'] \
                if row.get('score' if mode == 'tank' else 'carry') is not None and race_row['score' if mode == 'tank' else 'carry'] is not None else None
            rows.append(row)
    return rows
