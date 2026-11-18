#!/usr/bin/env python3
"""timing_cli.py: stands in for wowsimcli, wowsimcli_r, wowsimcli_assume, dashstats, tankbench and healbench during
pipeline.sh run, and logs what each call cost.

pipeline.sh links each binary of $DASH_BIN to this file in $DASH_DIR/timing/bin, and the step scripts call it unchanged.
It runs the real binary (TIMING_REAL_BIN) with the same arguments, then appends one line to TIMING_LOG: the step
(TIMING_STEP, which pipeline.sh sets), the binary, the CPU seconds the call used (all its threads: a sim splits its
iterations over DASH_SPLITS goroutines) and its wall seconds, and for a sim what sets its cost: the class, race,
iterations, fight length and target count of the request. timing.py sums the log up. A call costs a few milliseconds
more this way.

tankbench and healbench go through it too (tankheal.py's main runs and bench_races.py's race sims, which set
TIMING_STEP to tank, heal, tank_races or heal_races). For them the line holds the role, the boss, the iterations, the
number of request files and the fight length, read from the bench's flags. The shim prints nothing and passes stdin,
stdout, stderr and the exit code through, because tankheal.py keeps a bench's output as the mode's raw text.

`timing_cli.py phase <name> <start epoch>` (pipeline.sh hydrate and publish) and phase() (tankheal.py bins) log a
stretch of work that is no sim: its wall seconds go to TIMING_LOG when a run is logging, and to timing/phases.jsonl,
which timing.py shows.
"""
import fcntl, json, os, resource, subprocess, sys, time

REAL = os.environ.get('TIMING_REAL_BIN', '')
LOG = os.environ.get('TIMING_LOG', '')
BENCHES = ('tankbench', 'healbench')
BOOL_FLAGS = {'carry', 'floor', 'log', 'actions', 'classic-crit', 'bisect', 'labeled', 'boss-stats'}  # the benches' bool flags


def append(path, rec):
    with open(path, 'a') as fh:
        fcntl.flock(fh, fcntl.LOCK_EX)
        fh.write(json.dumps(rec) + '\n')


def phase(name, wall):
    """Log a stretch of work that is no sim (hydrate, publish, the binary build)."""
    rec = {'phase': name, 'wall': round(wall, 1), 't': round(time.time(), 1)}
    try:
        d = os.environ.get('DASH_DIR') or os.path.dirname(os.path.realpath(__file__))
        os.makedirs(os.path.join(d, 'timing'), exist_ok=True)
        append(os.path.join(d, 'timing', 'phases.jsonl'), rec)
        if os.environ.get('TIMING_LOG'):
            append(os.environ['TIMING_LOG'], rec)
    except OSError:
        pass


def bench_args(argv):
    """A bench's argv as (flags, request files). Go flags are -x v, -x=v or --x, and a bool flag takes no value."""
    flags, files, i = {}, [], 0
    while i < len(argv):
        a = argv[i]
        if a.startswith('-') and len(a) > 1:
            k, eq, v = a.lstrip('-').partition('=')
            if not eq:
                if k in BOOL_FLAGS or i + 1 >= len(argv):
                    v = 'true'
                else:
                    v, i = argv[i + 1], i + 1
            flags[k] = v
        else:
            files.append(a)
        i += 1
    return flags, files


def describe_bench(name, argv):
    """What sets a bench call's cost: its role, boss, iterations, number of requests and fight length."""
    flags, files = bench_args(argv)
    enc = flags.get('encounter', '')
    rec = {'role': flags.get('role') or ('tank' if name == 'tankbench' else 'both'),
           'boss': os.path.basename(enc)[:-5] if enc.endswith('.json') else '',
           'iters': int(flags.get('iters') or (3000 if name == 'tankbench' else 1000)), 'reqs': len(files),
           'carry': flags.get('carry', '') == 'true'}
    req = json.load(open(files[0])) if files else {}
    if len(files) == 1:
        rec['race'] = req['raid']['parties'][0]['players'][0].get('race', '')
    try:
        dur = json.load(open(enc)).get('duration', 0) if enc else 0
    except (OSError, ValueError, AttributeError):
        dur = 0
    rec['dur'] = dur or req.get('encounter', {}).get('duration', 0)
    return rec


def describe(argv, name=''):
    """What sets a sim's cost, from its request (a bench's, from its flags and request files)."""
    if name in BENCHES:
        try:
            return describe_bench(name, argv)
        except (OSError, ValueError, KeyError, IndexError, TypeError, AttributeError):
            return {}
    if argv[:1] != ['sim'] or '--infile' not in argv:
        return {}
    try:
        req = json.load(open(argv[argv.index('--infile') + 1]))
        pl = req['raid']['parties'][0]['players'][0]
        enc, opts = req.get('encounter', {}), req.get('simOptions', {})
        return {'cls': pl.get('class', ''), 'race': pl.get('race', ''), 'iters': int(opts.get('iterations') or 1),
                'dur': enc.get('duration', 0), 'targets': len(enc.get('targets', [])), 'replay': bool(opts.get('replaySeeds'))}
    except (OSError, ValueError, KeyError, IndexError, TypeError):
        return {}


def main():
    if os.path.basename(sys.argv[0]) == 'timing_cli.py' and sys.argv[1:2] == ['phase']:
        phase(sys.argv[2], time.time() - float(sys.argv[3]))
        return 0
    name, argv = os.path.basename(sys.argv[0]), sys.argv[1:]
    info = describe(argv, name)
    t0, c0 = time.monotonic(), resource.getrusage(resource.RUSAGE_CHILDREN)
    rc = subprocess.run([os.path.join(REAL, name), *argv]).returncode
    c1 = resource.getrusage(resource.RUSAGE_CHILDREN)
    rec = {'step': os.environ.get('TIMING_STEP', ''), 'bin': name, **info,
           'cpu': round(c1.ru_utime + c1.ru_stime - c0.ru_utime - c0.ru_stime, 3), 'wall': round(time.monotonic() - t0, 3)}
    if rc:
        rec['rc'] = rc
    if LOG:
        try:
            append(LOG, rec)
        except OSError:
            pass  # silent: tankheal.py keeps a bench's stdout and stderr
    return rc


if __name__ == '__main__':
    sys.exit(main())
