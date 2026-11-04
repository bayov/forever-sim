"""adaptive.py: fight counts sized to the report's luck goals, and the per-fight DPS they are measured on.

A comparison the report shows (a race against the next race, a racial left out against the race, an assumption against
the build) sims a pilot of PILOT fights for each side, then measures how far apart the two sides' fights land: every
sim uses seed 1, and fight i of a seed-1 run is the fight a one-fight run with seed 1+i sims, so the two sides of a
comparison fight the same fights and the error of their difference is the paired one, sd(b_i - a_i) / sqrt(N). With
labeled random streams (dashpaths.RNG) that is far smaller than two independent runs' error. From the pilot's paired sd
it picks N = SAFETY x (sd / goal)^2 fights, rounded up to a whole number of pilots, up to CEIL, and tops every side up to N once:
fights PILOT+1..N are a run with seed 1+PILOT, so the topped-up values are exactly a seed-1 run of N fights.

The Tanking and Healing modes (bench_races.py) size their race and racial sims the same way, on their own scores, with
BENCH_GOALS: the DPS goals put in each score's unit.
  tank   the score is raid DPS already (2 x SafeDPS less the DPS of the healer slots the tank's healing needs), so the
         goals are the DPS ones as they are: a race 0.35, a racial max(0.25, 10%), in score points. The score's paired
         spread is about 250 to 300 points a fight (healing needed is a 95th percentile), so a race comparison stops at
         CEIL fights (an error of about 1.5 points) and a racial at what 10% of its difference asks for.
  heal   the score is the damage carried, a scale of the encounter's damage (about 1.15 for the typical preset), so the
         goals keep the DPS goals' relative precision at the report's typical DPS: a race 0.35 / TYPICAL_DPS of
         TYPICAL_CARRY, a racial max(0.25 / TYPICAL_DPS of TYPICAL_CARRY, 10%). The carry's paired spread is about 0.2 a
         fight (whether a fight dies at a given scale), so those goals would need some 300k fights; healing sims stop at
         HEAL_CEIL instead (an error of about 0.0025), the most the full run's time allows (about 1.3 ms a fight over
         seven bosses and two lattice scales). The errors shown are the ones reached, and the report's ties use them.

N is set from the paired sd, not from the error, so it comes out the same at a tenth of the fights (probe.py's runs).

Goals (orchestration/followups.md, "DPS report luck goals"), one standard error of the difference the reader compares:
  race      a race against another race of the same build: 0.35 DPS
  racial    a racial or every racial left out, or another target type, against the race: max(0.25 DPS, 10%)
  asm       an assumption setting against the build: max(0.25 DPS, 5%)

Vectors: each sim's per-fight DPS is kept in $DASH_DIR/vec/<build>/<key>.f64 (and .json: the fights it asked for), the
key a hash of the request and the switches it ran with, so any step asking for the same sim reads the same fights and
tops them up. run_build.py clears a build's vectors when it sims the build again; a missing vector is simmed again
(the same fights, bit for bit). They aren't packed: only the numbers made from them are.
"""
import array, fcntl, hashlib, json, math, os

from dashpaths import D, RNG

PILOT = int(os.environ.get('ADAPT_PILOT', 1000))
RACE_PILOT = int(os.environ.get('ADAPT_RACE_PILOT', 2000))
CEIL = int(os.environ.get('ADAPT_MAX', 40000))
SAFETY = 1.2
GOALS = {'race': (0.35, 0.0), 'racial': (0.25, 0.10), 'asm': (0.25, 0.05)}  # (DPS floor, share of the difference)
TYPICAL_DPS = 1047  # the median DPS of the report's generic-boss presets (2026-11-04)
TYPICAL_CARRY = 1.15  # the median damage carried of the healing presets (2026-11-04)
HEAL_CEIL = int(os.environ.get('ADAPT_HEAL_MAX', 6000))
BENCH_GOALS = {
    'tank': {'race': GOALS['race'], 'racial': GOALS['racial']},
    'heal': {'race': (GOALS['race'][0] * TYPICAL_CARRY / TYPICAL_DPS, 0.0),
             'racial': (GOALS['racial'][0] * TYPICAL_CARRY / TYPICAL_DPS, GOALS['racial'][1]), 'ceil': HEAL_CEIL},
}


def goal(kind, diff=0.0):
    floor, share = GOALS[kind]
    return max(floor, share * abs(diff))


def fights(sd, g, pilot, ceil=CEIL):
    """Fights to ask for so a paired sd of `sd` gives an error under the goal `g`: a whole number of pilots."""
    need = SAFETY * (sd / g) ** 2
    return min(ceil, max(pilot, pilot * math.ceil(need / pilot)))


def paired(a, b):
    """(mean, sd, se) of b_i - a_i over the fights both have."""
    n = min(len(a), len(b))
    if n < 2:
        return (b[0] - a[0] if n else 0.0), 0.0, 0.0
    d = [b[i] - a[i] for i in range(n)]
    m = sum(d) / n
    sd = math.sqrt(max(0.0, sum((x - m) ** 2 for x in d) / (n - 1)))
    return m, sd, sd / math.sqrt(n)


def mean(v):
    return sum(v) / len(v)


def sd(v):
    m = mean(v)
    return math.sqrt(max(0.0, sum((x - m) ** 2 for x in v) / max(1, len(v) - 1)))


def key(req, env=None):
    """The identity of a sim's fights: the request less its simOptions, the random stream options, and the switches."""
    r = {k: v for k, v in req.items() if k != 'simOptions'}
    blob = json.dumps([r, RNG, sorted((env or {}).items())], sort_keys=True, separators=(',', ':'))
    return hashlib.sha1(blob.encode()).hexdigest()[:20]


def clear(bid):
    d = os.path.join(D, 'vec', bid)
    if os.path.isdir(d):
        for f in os.listdir(d):
            os.remove(os.path.join(d, f))


def values(bid, req, env, n, sim, have=None, width=1):
    """Per-fight DPS of the first n fights of this sim (seed 1), from the stored vector topped up as needed.

    width: values a fight (the benches keep several, fight by fight: tankheal.py); sim then returns width values a fight.

    sim(seed, count) runs `count` fights from `seed` and returns their DPS. `have` seeds the store with values a step
    already simmed (the first len(have) fights asked as `have[0]` fights: (asked, values)). Under probe_cli.py a sim
    returns fewer fights than it asks for, so the result can be shorter than n; the store keeps what was asked."""
    d = os.path.join(D, 'vec', bid)
    os.makedirs(d, exist_ok=True)
    base = os.path.join(d, key(req, env))
    with open(base + '.lock', 'w') as lk:
        fcntl.flock(lk, fcntl.LOCK_EX)
        asked, vals = 0, array.array('d')
        if os.path.exists(base + '.json') and os.path.exists(base + '.f64'):
            asked = json.load(open(base + '.json'))['asked']
            with open(base + '.f64', 'rb') as fh:
                vals.frombytes(fh.read())
        if have and have[0] > asked:
            asked, vals = have[0], array.array('d', have[1])
            save = True
        else:
            save = False
        if asked < n:
            vals.extend(sim(1 + asked, n - asked))
            asked = n
            save = True
        if save:
            with open(base + '.f64', 'wb') as fh:
                vals.tofile(fh)
            json.dump({'asked': asked, 'fights': len(vals)}, open(base + '.json', 'w'))
        # under probe_cli.py each run returned a fraction of what it asked; keep the share of fights n asks for
        got = len(vals) // width
        cut = got if asked <= n else max(1, round(got * n / asked))
        return list(vals[:cut * width])
