#!/usr/bin/env python3
"""probe.py: tells which DPS Report builds a code or data change really moves, without re-simming them all.

stale.mjs flags a build when one of its inputs changed (the sim's packages, its class, its request, the item database
entries it wears, the step scripts), which says "may have moved". A probe says whether it did. It runs every step of
the pipeline for a build (run_build, weights, races, mana, ablation, assumptions; the steps pipeline.sh runs for it) in
a scratch folder of its own, with probe_cli.py standing in for the binaries: every sim runs at a tenth of its
iterations (one-fight runs, replays and runs of 50 fights or fewer as asked) and the mana step logs 10 fights instead
of 100. It records a fingerprint of every sim's full result keyed by the request the script asked for, and a
fingerprint of every output the steps write (out, names, weights, mana, the race, assumption and ablation rows), so a
change to a script's own post-processing shows too.

The sim reseeds every fight (seed + i), so a probe of unchanged code and data gives the same fingerprints whatever else
changed, and `pipeline.sh run` stores each re-simmed build's probe in its raw pack. When the probe of the code being
checked equals the pack's, the build's full sims would come out the same: it is recorded as current (stale.mjs --by
probe) instead of re-simmed. Only fights the probe doesn't sim can hide a change (a code path one fight in thousands
reaches), so every check re-sims a few cleared builds in full (the audit) and publish refuses if one moved.

  probe.py run [ids]          probe these builds (default: every request) into $DASH_DIR/probe/<id>.json
  probe.py check [ids]        probe these builds (default: the ones stale.mjs lists) into $DASH_DIR/probe_check/,
                              compare each with the probe in its raw pack, and write $DASH_DIR/probe_audit.json:
                              the builds it cleared, the ones it picked to re-sim anyway (the audit) and the ones to re-sim
  probe.py audit              publish's check: every audit build's fresh sims must equal its old pack, and none of the
                              cleared builds may have been re-simmed for another reason
  probe.py passed             the builds probe_audit.json cleared and no run re-simmed (publish records them --by probe)

Environment: PROBE_P builds at a time (default 4), PROBE_SCALE (10), PROBE_MIN (50), PROBE_FIGHTS (10), PROBE_AUDIT
(builds to re-sim anyway, default 3), DASH_DIR, DASH_BIN (the real binaries), DASH_TREE and DASH_SPLITS as pipeline.sh
sets them. A probe from another split count, scale or machine architecture isn't compared.
"""
import glob, hashlib, io, json, os, platform, random, shutil, subprocess, sys, tarfile
from concurrent.futures import ThreadPoolExecutor

sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
from dashpaths import D, BIN  # noqa: E402

HERE = os.path.dirname(os.path.abspath(__file__))
ROOT = os.path.abspath(os.path.join(HERE, '..', '..'))
RAW = os.path.join(HERE, 'raw')
P = int(os.environ.get('PROBE_P', 4))
SETTINGS = {
    'format': 1,
    'scale': int(os.environ.get('PROBE_SCALE', 10)),
    'min': int(os.environ.get('PROBE_MIN', 50)),
    'fights': int(os.environ.get('PROBE_FIGHTS', 10)),
    'splits': int(os.environ.get('DASH_SPLITS', 4)),
    'arch': platform.machine(),
}
AUDIT_N = int(os.environ.get('PROBE_AUDIT', 3))
AUDIT_FILE = os.path.join(D, 'probe_audit.json')
BINARIES = ('wowsimcli', 'wowsimcli_r', 'wowsimcli_assume', 'dashstats')


def meta(bid):
    return json.load(open(os.path.join(D, 'req', bid + '.meta.json')))


def all_ids():
    return sorted(os.path.basename(p)[:-10] for p in glob.glob(os.path.join(D, 'req', '*.meta.json')))


def steps(bid):
    """The steps pipeline.sh runs for a build, in its order."""
    m = meta(bid)
    base = not (m.get('duty') or m.get('pet'))
    out = [('run_build', ['run_build.py', bid])]
    if base:
        out.append(('weights', ['weights.py', bid]))
        out.append(('races', ['run_races.py', bid]))
    if not bid.startswith(('warrior__', 'rogue__')):
        out.append(('mana', ['mana_curves.py', '--force', bid]))
    if base:
        out.append(('ablate', ['ablate.py', bid]))
        out.append(('assumptions', ['run_assumptions.py', bid]))
    return out


def h(data):
    return hashlib.sha256(data if isinstance(data, bytes) else data.encode()).hexdigest()[:20]


def canon_json(x):
    sys.path.insert(0, HERE)
    from probe_cli import canon
    return json.dumps(canon(x), sort_keys=True, separators=(',', ':'))


def outputs(work, bid):
    """Fingerprints of what the steps wrote for the build, the way rawstore.py packs it."""
    fp = {}
    p = os.path.join(work, 'out', bid + '.json')
    if os.path.exists(p):
        o = json.load(open(p))
        o.pop('meta', None)  # the meta file's text: not a sim output
        fp['out'] = h(canon_json(o))
    for d, ext in (('names', '.tsv'), ('weights', '.json'), ('mana', '.json')):
        p = os.path.join(work, d, bid + ext)
        if os.path.exists(p):
            data = open(p).read()
            fp[d] = h(''.join(sorted(set(data.splitlines(keepends=True)))) if d == 'names' else canon_json(json.loads(data)))
    for name, rel in (('rv', 'rv/results.jsonl'), ('asm', 'asm/results.jsonl')):
        p = os.path.join(work, rel)
        if os.path.exists(p):
            fp[name] = h(''.join(sorted(set(canon_json(json.loads(l)) + '\n' for l in open(p) if l.strip()))))
    p = os.path.join(work, 'abl', 'ablate.json')
    if os.path.exists(p):
        a = json.load(open(p))
        fp['abl'] = h(canon_json({'iters': a.get('iters'), 'seed': a.get('seed'), 'entry': a['builds'].get(bid)}))
    return fp


def probe_bin(base):
    b = os.path.join(base, 'bin')
    os.makedirs(b, exist_ok=True)
    for name in BINARIES:
        link = os.path.join(b, name)
        if os.path.lexists(link):
            os.remove(link)
        if os.path.exists(os.path.join(BIN, name)):  # the scripts fall back to wowsimcli when a variant is missing
            os.symlink(os.path.join(HERE, 'probe_cli.py'), link)
    return b


def probe_one(bid, base, pbin):
    work = os.path.join(base, bid)
    shutil.rmtree(work, ignore_errors=True)
    os.makedirs(work)
    for f in ('req', 'asm_apls'):
        if os.path.exists(os.path.join(D, f)):
            os.symlink(os.path.realpath(os.path.join(D, f)), os.path.join(work, f))
    logp = os.path.join(work, 'probe.log')
    env = dict(os.environ, DASH_DIR=work, DASH_BIN=pbin, PROBE_REAL_BIN=BIN, PROBE_LOG=logp, PAR='1', ABL_MERGE='1',
               FIGHTS=str(SETTINGS['fights']), PROBE_SCALE=str(SETTINGS['scale']), PROBE_MIN=str(SETTINGS['min']),
               DASH_SPLITS=str(SETTINGS['splits']))
    for k in ('ITERS', 'MAX_ITERS', 'SEED', 'WEIGHTS_FORCE', 'ASM_REUSE_OUT', 'RACES_ALL_TARGETS'):
        env.pop(k, None)  # the probe runs each step with its defaults, as pipeline.sh does
    failed = {}
    for step, cmd in steps(bid):
        r = subprocess.run([sys.executable, os.path.join(HERE, cmd[0]), *cmd[1:]], env=dict(env, PROBE_STEP=step),
                           cwd=work, capture_output=True, text=True)
        if r.returncode:
            failed[step] = r.returncode
            print(f'{bid}: {step} failed (exit {r.returncode}): {r.stderr.strip().splitlines()[-1:] or ""}', file=sys.stderr)
    sims = {}
    for line in open(logp) if os.path.exists(logp) else []:
        x = json.loads(line)
        sims[f"{x['step']}:{x['kind']}:{x['key']}"] = x['fp']
    rec = {'settings': SETTINGS, 'sims': dict(sorted(sims.items())), 'outputs': outputs(work, bid)}
    if failed:
        rec['failed'] = failed
    shutil.rmtree(work, ignore_errors=True)
    return rec


def run(ids, dest):
    base = os.path.join(D, 'probe_work')
    pbin = probe_bin(base)
    os.makedirs(dest, exist_ok=True)

    def one(bid):
        rec = probe_one(bid, base, pbin)
        json.dump(rec, open(os.path.join(dest, bid + '.json'), 'w'), indent=0, sort_keys=True)
        return bid, rec

    with ThreadPoolExecutor(P) as ex:
        res = dict(ex.map(one, ids))
    shutil.rmtree(base, ignore_errors=True)
    n = sum(len(r['sims']) for r in res.values())
    print(f'probed {len(ids)} builds ({n} sims) into {dest}')
    return res


def stored(bid):
    """The probe in the build's raw pack, or None."""
    p = os.path.join(RAW, bid + '.tgz')
    if not os.path.exists(p):
        return None
    with tarfile.open(p) as tar:
        try:
            return json.load(tar.extractfile('probe.json'))
        except KeyError:
            return None


def compare(old, new):
    """What differs between two probes, by step: [] when they're the same."""
    if old is None:
        return ['no stored probe (the build was last simmed before probes were kept)']
    if old['settings'] != new['settings']:
        return [f'the stored probe has other settings ({old["settings"]})']
    out = []
    if old.get('failed') or new.get('failed'):
        out.append(f'failed steps: {new.get("failed") or {}}')
    steps_ = sorted({k.split(':')[0] for k in old['sims']} | {k.split(':')[0] for k in new['sims']})
    for s in steps_:
        o = {k: v for k, v in old['sims'].items() if k.split(':')[0] == s}
        n = {k: v for k, v in new['sims'].items() if k.split(':')[0] == s}
        changed = sum(1 for k in o.keys() & n.keys() if o[k] != n[k])
        gone, added = len(o.keys() - n.keys()), len(n.keys() - o.keys())
        if changed or gone or added:
            out.append(f'{s}: ' + ', '.join(x for x in (f'{changed} of {len(o.keys() & n.keys())} sims differ' if changed else '',
                                                          f'{gone} sims no longer run' if gone else '',
                                                          f'{added} new sims' if added else '') if x))
    outs = sorted(k for k in old['outputs'].keys() | new['outputs'].keys() if old['outputs'].get(k) != new['outputs'].get(k))
    if outs:
        out.append('outputs differ: ' + ', '.join(outs))
    return out


def stale_ids():
    r = subprocess.run(['node', os.path.join(HERE, 'stale.mjs'), '--json'], cwd=ROOT, capture_output=True, text=True, check=True)
    return [s['id'] for s in json.loads(r.stdout)['stale']]


def head():
    return subprocess.run(['git', 'rev-parse', 'HEAD'], cwd=ROOT, capture_output=True, text=True).stdout.strip()


def check(ids, as_json=False):
    ids = ids or stale_ids()
    if not ids:
        print('nothing is stale')
        return
    new = run(ids, os.path.join(D, 'probe_check'))
    same, differ = [], {}
    for bid in ids:
        why = compare(stored(bid), new[bid])
        (differ.__setitem__(bid, why) if why else same.append(bid))
    # the audit: a few cleared builds are re-simmed anyway, and publish checks they come out unchanged
    rng = random.Random(head())
    audit = sorted(rng.sample(same, min(AUDIT_N, len(same))))
    passed = sorted(set(same) - set(audit))
    json.dump({'commit': head(), 'passed': passed, 'audit': audit, 'resim': sorted(differ), 'why': differ},
              open(AUDIT_FILE, 'w'), indent=1)
    if as_json:
        print(json.dumps({'same': sorted(same), 'audit': audit, 'differ': differ}, indent=1))
        return
    print(f'{len(same)} of {len(ids)} builds come out the same: {" ".join(sorted(same)) or "none"}')
    for bid, why in sorted(differ.items()):
        print(f'  {bid}: {"; ".join(why)}')
    resim = sorted(differ) + audit
    print(f'audit (cleared, but re-simmed to check the probe): {" ".join(audit) or "none"}')
    print(f'next: pipeline.sh run {" ".join(resim)}' if resim else 'next: nothing to sim')
    print('then: pipeline.sh publish <the same ids>: it checks the audit and records the cleared builds (--by probe)')


def members(bid, data):
    """A pack's sim members, comparable across packs: no fingerprint or probe, no meta text in out."""
    out = {}
    with tarfile.open(fileobj=io.BytesIO(data)) as tar:
        for ti in tar.getmembers():
            if ti.name in ('fingerprint.json', 'probe.json'):
                continue
            x = tar.extractfile(ti).read()
            if ti.name == f'out/{bid}.json':
                o = json.loads(x)
                o.pop('meta', None)
                x = json.dumps(o, sort_keys=True).encode()
            out[ti.name] = x
    return out


def audit(resimmed):
    """publish's check, before it writes anything: exit 1 when an audit build moved."""
    if not os.path.exists(AUDIT_FILE):
        return
    a = json.load(open(AUDIT_FILE))
    import rawstore
    bad, missing = [], [b for b in a['audit'] if b not in resimmed]
    for bid in a['audit']:
        if bid not in resimmed:
            continue
        old = os.path.join(RAW, bid + '.tgz')
        new = rawstore.pack_bytes(bid, 'audit')
        if not os.path.exists(old) or members(bid, open(old, 'rb').read()) != members(bid, new):
            bad.append(bid)
    if missing:
        sys.exit(f'probe audit: re-sim the audit builds too (pipeline.sh run {" ".join(missing)}), or delete {AUDIT_FILE} to '
                 f'record nothing by probe')
    if bad:
        sys.exit(f'probe audit FAILED: {" ".join(bad)} passed the probe, but their full sims moved. The probe missed a change, '
                 f'so its other clearances can\'t be trusted: re-sim them (pipeline.sh run {" ".join(a["passed"])}), then '
                 f'delete {AUDIT_FILE} and publish')
    print(f'probe audit: {len(a["audit"])} cleared builds re-simmed in full and unchanged')


def passed(resimmed):
    if not os.path.exists(AUDIT_FILE):
        return []
    a = json.load(open(AUDIT_FILE))
    if a['commit'] != head():
        print(f'probe_audit.json is from {a["commit"][:10]}, not HEAD: its builds are not recorded (probe again)', file=sys.stderr)
        return []
    return [b for b in a['passed'] if b not in resimmed]


if __name__ == '__main__':
    cmd, args = (sys.argv[1], sys.argv[2:]) if len(sys.argv) > 1 else ('', [])
    if cmd == 'run':
        run(args or all_ids(), os.path.join(D, 'probe'))
    elif cmd == 'check':
        check([a for a in args if not a.startswith('--')], '--json' in args)
    elif cmd == 'audit':
        audit(set(args))
    elif cmd == 'passed':
        print(' '.join(passed(set(args))))
    else:
        sys.exit(__doc__)
