#!/usr/bin/env python3
"""probe_cli.py: stands in for wowsimcli, wowsimcli_r, wowsimcli_assume and dashstats while probe.py runs the step scripts.

probe.py links each binary name of $DASH_BIN to this file in a bin folder of its own, so the step scripts run unchanged
and every sim they ask for comes through here. Each sim runs with its iterations cut to a tenth (PROBE_SCALE, at least
PROBE_MIN, never more than asked; one-fight runs and replays run as asked), and the result's fingerprint is appended to
PROBE_LOG keyed by the request as the script asked for it (its own iteration count) and the DASH_ env knobs that change
a sim. The sim is seeded per fight (sim.go: RandomSeed + i), so the first fights of a run are the same fights at any
iteration count, and a probe of unchanged code gives the same fingerprints.

The fingerprint is a sha256 of the result with every list of objects sorted (the sim emits actions in Go map order) and
every zero, empty or false field dropped (so a new result field nobody fills doesn't move every build).
"""
import fcntl, hashlib, json, os, subprocess, sys, tempfile

REAL = os.environ.get('PROBE_REAL_BIN', '')
LOG = os.environ.get('PROBE_LOG', '')
SCALE = int(os.environ.get('PROBE_SCALE', 10))
MIN = int(os.environ.get('PROBE_MIN', 50))
KNOBS = ('DASH_ASSUME', 'DASH_RACIALS_OFF', 'DASH_SPLITS')


def canon(x):
    if isinstance(x, dict):
        out = {}
        for k, v in x.items():
            v = canon(v)
            if v not in (0, 0.0, '', False, None, [], {}):
                out[k] = v
        return out
    if isinstance(x, list):
        x = [canon(v) for v in x]
        if any(isinstance(v, (dict, list)) for v in x):
            x.sort(key=lambda v: json.dumps(v, sort_keys=True))
        return x
    return x


def digest(x):
    return hashlib.sha256(json.dumps(canon(x), sort_keys=True, separators=(',', ':')).encode()).hexdigest()[:20]


def log(kind, key, fp, asked, ran):
    line = json.dumps({'step': os.environ.get('PROBE_STEP', ''), 'kind': kind, 'key': key, 'fp': fp, 'asked': asked, 'ran': ran}) + '\n'
    with open(LOG, 'a') as fh:
        fcntl.flock(fh, fcntl.LOCK_EX)
        fh.write(line)


def sim(name, argv):
    infile, outfile = argv[argv.index('--infile') + 1], argv[argv.index('--outfile') + 1]
    req = json.load(open(infile))
    opts = req.setdefault('simOptions', {})
    asked = int(opts.get('iterations') or 1)
    key = digest({'req': req, 'env': {k: os.environ.get(k, '') for k in KNOBS}})
    ran = asked if asked <= MIN or opts.get('replaySeeds') else max(MIN, asked // SCALE)
    opts['iterations'] = ran
    with tempfile.TemporaryDirectory() as d:
        p = os.path.join(d, 'i.json')
        json.dump(req, open(p, 'w'))
        args = list(argv)
        args[args.index('--infile') + 1] = p
        r = subprocess.run([os.path.join(REAL, name), *args])  # args starts with the sim subcommand
    if r.returncode:
        log('sim', key, f'exit {r.returncode}', asked, ran)
    else:
        log('sim', key, digest(json.load(open(outfile))), asked, ran)
    return r.returncode


def stats(name, argv):
    req = json.load(open(argv[0]))
    r = subprocess.run([os.path.join(REAL, name), *argv], capture_output=True, text=True)
    sys.stdout.write(r.stdout)
    sys.stderr.write(r.stderr)
    out = hashlib.sha256(f'{r.returncode}\n{r.stdout}'.encode()).hexdigest()[:20]
    log('dashstats', digest({'req': req}), out, 0, 0)
    return r.returncode


if __name__ == '__main__':
    name, argv = os.path.basename(sys.argv[0]), sys.argv[1:]
    sys.exit(stats(name, argv) if name.startswith('dashstats') else sim(name, argv) if argv[:1] == ['sim'] else
             subprocess.run([os.path.join(REAL, name), *argv]).returncode)
