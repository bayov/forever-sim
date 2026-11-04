#!/usr/bin/env python3
"""timing_cli.py: stands in for wowsimcli, wowsimcli_r, wowsimcli_assume and dashstats during pipeline.sh run, and logs
what each call cost.

pipeline.sh links each binary of $DASH_BIN to this file in $DASH_DIR/timing/bin, and the step scripts call it unchanged.
It runs the real binary (TIMING_REAL_BIN) with the same arguments, then appends one line to TIMING_LOG: the step
(TIMING_STEP, which pipeline.sh sets), the binary, the CPU seconds the call used (all its threads: a sim splits its
iterations over DASH_SPLITS goroutines) and its wall seconds, and for a sim what sets its cost: the class, race,
iterations, fight length and target count of the request. timing.py sums the log up. A call costs a few milliseconds
more this way.
"""
import fcntl, json, os, resource, subprocess, sys, time

REAL = os.environ.get('TIMING_REAL_BIN', '')
LOG = os.environ.get('TIMING_LOG', '')


def describe(argv):
    """What sets a sim's cost, from its request."""
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
    name, argv = os.path.basename(sys.argv[0]), sys.argv[1:]
    info = describe(argv)
    t0, c0 = time.monotonic(), resource.getrusage(resource.RUSAGE_CHILDREN)
    rc = subprocess.run([os.path.join(REAL, name), *argv]).returncode
    c1 = resource.getrusage(resource.RUSAGE_CHILDREN)
    rec = {'step': os.environ.get('TIMING_STEP', ''), 'bin': name, **info,
           'cpu': round(c1.ru_utime + c1.ru_stime - c0.ru_utime - c0.ru_stime, 3), 'wall': round(time.monotonic() - t0, 3)}
    if rc:
        rec['rc'] = rc
    if LOG:
        with open(LOG, 'a') as fh:
            fcntl.flock(fh, fcntl.LOCK_EX)
            fh.write(json.dumps(rec) + '\n')
    return rc


if __name__ == '__main__':
    sys.exit(main())
