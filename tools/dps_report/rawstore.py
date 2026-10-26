#!/usr/bin/env python3
"""rawstore.py: the DPS Report's raw sim outputs, kept per build in tools/dps_report/raw/<id>.tgz on master.

Every dataset is rebuilt by build_data.py from the raw outputs of every build, and some datasets span builds (ranks,
race_summary, race_notes, assumptions). Keeping each build's raw outputs next to the datasets lets a regeneration
re-sim only the builds that changed (or none, for a text change) and still rebuild every dataset exactly.
orchestration/dash_raw_store.md has the design; pipeline.sh wires these commands together.

  rawstore.py pack [ids...]       write raw/<id>.tgz for these builds (default: every build in $DASH_DIR/out),
                                  after checking each has all its outputs
  rawstore.py unpack              load every raw/<id>.tgz into an empty $DASH_DIR
  rawstore.py clean <ids...>      drop these builds from every raw output in $DASH_DIR, so every step re-sims them
  rawstore.py variants <ids...>   the ids plus every preset whose baseBuild is one of them (duty and encounter copies)
  rawstore.py check [ids...]      rebuild nothing: compare $DASH_DIR/data (a fresh build_data.py) with the published
                                  datasets. A build outside [ids] whose sim numbers differ is drift: exit 1.
  rawstore.py prune-events        delete the sample fights and raw packs of builds that have no request any more
  rawstore.py copy                copy $DASH_DIR/data into ui/dps_report/data, stamping `updated` in datasets.json only
                                  for files that changed, and drop the events, raw packs and provenance records of
                                  builds that have no request any more

A pack holds one build's out/<id>.json, names/<id>.tsv, weights/<id>.json and mana/<id>.json where it has them, its
lines of rv/results.jsonl and asm/results.jsonl (rv.jsonl, asm.jsonl: sorted by key, each key once, the last line
winning as in build_races.py and asm_data.py), its abl/ablate.json entry with that file's iters and seed (abl.json),
and fingerprint.json: the sha256 of its request and assumption rotations, the sim split count and the commit simmed at.
Packs are deterministic (sorted members, zero mtimes and owners, gzip without a name or time), so an unchanged
build's pack is byte-identical and git stores it once.
"""
import glob, gzip, hashlib, io, json, os, subprocess, sys, tarfile
from datetime import datetime, timezone

sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
from dashpaths import D  # noqa: E402

HERE = os.path.dirname(os.path.abspath(__file__))
ROOT = os.path.abspath(os.path.join(HERE, '..', '..'))
RAW = os.path.join(HERE, 'raw')
REPORT = os.path.join(ROOT, 'ui', 'dps_report')
SPLITS = int(os.environ.get('DASH_SPLITS', '4'))
FILES = (('out', '.json'), ('names', '.tsv'), ('weights', '.json'), ('mana', '.json'))
RV, ASM, ABL = os.path.join(D, 'rv', 'results.jsonl'), os.path.join(D, 'asm', 'results.jsonl'), os.path.join(D, 'abl', 'ablate.json')
RV_KEY = ('build', 'race', 'variant')
ASM_KEY = ('build', 'combo', 'secs', 'rot')


def meta(bid):
    p = os.path.join(D, 'req', bid + '.meta.json')
    return json.load(open(p)) if os.path.exists(p) else None


def all_ids():
    return sorted(os.path.basename(p)[:-10] for p in glob.glob(os.path.join(D, 'req', '*.meta.json')))


def sha(path):
    return hashlib.sha256(open(path, 'rb').read()).hexdigest()


def lines_by_build(path, key):
    """{build: {key: line}} with the last line winning for a repeated key."""
    out = {}
    if os.path.exists(path):
        for line in open(path):
            if line.strip():
                x = json.loads(line)
                out.setdefault(x['build'], {})[tuple(x[k] for k in key)] = json.dumps(x, sort_keys=True)
    return out


def write_lines(path, by_build):
    os.makedirs(os.path.dirname(path), exist_ok=True)
    with open(path, 'w') as fh:
        for b in sorted(by_build):
            for k in sorted(by_build[b], key=repr):
                fh.write(by_build[b][k] + '\n')


# What a build should have, so a pack never stores a build with a step missing or a failed sim.
def expected(bid):
    m = meta(bid)
    duty, enc = m.get('duty'), m.get('encounter')
    mana = not bid.startswith(('warrior__', 'rogue__'))
    want = {'out': True, 'names': True, 'weights': not duty, 'mana': mana, 'abl': not duty}
    rv = set()
    if not duty and not enc:
        import run_races
        rv = {k for k, _ in run_races.tasks(bid)}
    asm = set()
    if not duty:
        import run_assumptions as ra
        if ra.keys_for(bid):
            for b, combo, kind in ra.tasks(bid):
                lens = [180] + ([120, 300] if kind in ('base', 'single') else [])
                asm |= {(b, ra.combo_str(combo), secs, rot) for secs in lens for rot in ra.rots(combo)}
    return want, rv, asm


def pack(ids, commit):
    rv, asm = lines_by_build(RV, RV_KEY), lines_by_build(ASM, ASM_KEY)
    abl = json.load(open(ABL)) if os.path.exists(ABL) else {'builds': {}}
    os.makedirs(RAW, exist_ok=True)
    problems = []
    for bid in ids:
        if meta(bid) is None:
            problems.append(f'{bid}: no request')
            continue
        want, rv_want, asm_want = expected(bid)
        members = {}
        for d, ext in FILES:
            p = os.path.join(D, d, bid + ext)
            if os.path.exists(p):
                data = open(p, 'rb').read()
                if d == 'names':  # patch 0001 appends one copy per sim split, in the order the splits finish
                    data = ''.join(sorted(set(data.decode().splitlines(keepends=True)))).encode()
                members[f'{d}/{bid}{ext}'] = data
            elif want[d]:
                problems.append(f'{bid}: no {d}/{bid}{ext}')
        got_rv = rv.get(bid, {})
        missing = rv_want - set(got_rv)
        if missing:
            problems.append(f'{bid}: {len(missing)} race scenario sims missing, e.g. {sorted(missing, key=repr)[0]}')
        if got_rv:
            members['rv.jsonl'] = ''.join(got_rv[k] + '\n' for k in sorted(got_rv, key=repr) if k in rv_want).encode()
        got_asm = asm.get(bid, {})
        missing = asm_want - set(got_asm)
        if missing:
            problems.append(f'{bid}: {len(missing)} assumption sims missing, e.g. {sorted(missing, key=repr)[0]}')
        if got_asm:  # combos no longer in assumptions_def.py are left behind
            members['asm.jsonl'] = ''.join(got_asm[k] + '\n' for k in sorted(got_asm, key=repr) if k in asm_want).encode()
        entry = abl['builds'].get(bid)
        if entry is not None:
            if any('error' in x for x in entry.get('drop', []) + entry.get('pairs', [])):
                problems.append(f'{bid}: a setup ablation sim failed')
            members['abl.json'] = json.dumps({'iters': abl['iters'], 'seed': abl['seed'], 'entry': entry}, sort_keys=True).encode()
        elif want['abl']:
            problems.append(f'{bid}: no setup ablation entry')
        apls = sorted(glob.glob(os.path.join(HERE, 'asm_apls', bid + '.*')))
        members['fingerprint.json'] = json.dumps({
            'request': sha(os.path.join(D, 'req', bid + '.json')),
            'asmApls': {os.path.basename(p): sha(p) for p in apls},
            'splits': SPLITS,
            'commit': commit,
        }, indent=1, sort_keys=True).encode()
        buf = io.BytesIO()
        with tarfile.open(fileobj=buf, mode='w', format=tarfile.PAX_FORMAT) as tar:
            for name in sorted(members):
                ti = tarfile.TarInfo(name)
                ti.size, ti.mtime, ti.mode, ti.uid, ti.gid, ti.uname, ti.gname = len(members[name]), 0, 0o644, 0, 0, '', ''
                tar.addfile(ti, io.BytesIO(members[name]))
        out = io.BytesIO()
        with gzip.GzipFile(filename='', mode='wb', fileobj=out, mtime=0, compresslevel=9) as gz:
            gz.write(buf.getvalue())
        path = os.path.join(RAW, bid + '.tgz')
        old = open(path, 'rb').read() if os.path.exists(path) else None
        if old != out.getvalue():
            open(path, 'wb').write(out.getvalue())
            print(f'packed {bid} ({len(out.getvalue()) // 1024} KB)')
    if problems:
        print('\n'.join(['Not complete:'] + problems), file=sys.stderr)
        sys.exit(1)


def unpack():
    for d in ('out', 'names', 'weights', 'mana', 'rv', 'asm', 'abl', 'data'):
        if glob.glob(os.path.join(D, d, '*')):
            sys.exit(f'unpack: {D}/{d} is not empty; hydrate starts from an empty $DASH_DIR')
    rv, asm, abl = {}, {}, {'builds': {}}
    for path in sorted(glob.glob(os.path.join(RAW, '*.tgz'))):
        bid = os.path.basename(path)[:-4]
        with tarfile.open(path) as tar:
            for ti in tar.getmembers():
                data = tar.extractfile(ti).read()
                if ti.name == 'fingerprint.json':
                    fp = json.loads(data)
                    if fp['splits'] != SPLITS:
                        sys.exit(f'{bid}: simmed with {fp["splits"]} splits, this run uses DASH_SPLITS={SPLITS}')
                elif ti.name == 'rv.jsonl':
                    rv[bid] = data.decode()
                elif ti.name == 'asm.jsonl':
                    asm[bid] = data.decode()
                elif ti.name == 'abl.json':
                    a = json.loads(data)
                    for k in ('iters', 'seed'):
                        if abl.setdefault(k, a[k]) != a[k]:
                            sys.exit(f'{bid}: setup ablation at {k} {a[k]}, other builds at {abl[k]}')
                    abl['builds'][bid] = a['entry']
                else:
                    p = os.path.join(D, ti.name)
                    os.makedirs(os.path.dirname(p), exist_ok=True)
                    open(p, 'wb').write(data)
    for path, by in ((RV, rv), (ASM, asm)):
        os.makedirs(os.path.dirname(path), exist_ok=True)
        open(path, 'w').write(''.join(by[b] for b in sorted(by)))
    if abl['builds']:
        os.makedirs(os.path.dirname(ABL), exist_ok=True)
        json.dump(abl, open(ABL, 'w'), indent=1)
    print(f'unpacked {len(glob.glob(os.path.join(RAW, "*.tgz")))} builds into {D}')


def clean(ids):
    ids = set(ids)
    for d, ext in FILES:
        for bid in ids:
            p = os.path.join(D, d, bid + ext)
            if os.path.exists(p):
                os.remove(p)
    for path, key in ((RV, RV_KEY), (ASM, ASM_KEY)):
        if os.path.exists(path):
            by = lines_by_build(path, key)
            write_lines(path, {b: v for b, v in by.items() if b not in ids})
    if os.path.exists(ABL):
        a = json.load(open(ABL))
        a['builds'] = {b: x for b, x in a['builds'].items() if b not in ids}
        json.dump(a, open(ABL, 'w'), indent=1)
    print(f'cleaned {len(ids)} builds')


def variants(ids):
    ids = list(dict.fromkeys(ids))
    extra = [b for b in all_ids() if b not in ids and (meta(b) or {}).get('baseBuild') in ids]
    if extra:
        print(f'adding the variants of the given builds: {" ".join(extra)}', file=sys.stderr)
    print(' '.join(ids + extra))


# The datasets made from sims alone. A build's rows in them only change when it is re-simmed.
SIM_DATASETS = ('builds', 'races', 'lengths', 'hist', 'weights', 'mana', 'mana_summary', 'race_variants', 'assumption_dps')
NOT_SIM = {'rank', 'rankAll', 'of', 'vsBest', 'vsFactionBest', 'factionRank', 'dutyCostPct', 'dutyCost'}


def numbers(row):
    return tuple(sorted((k, round(v, 2)) for k, v in row.items()
                        if isinstance(v, (int, float)) and not isinstance(v, bool) and k not in NOT_SIM))


def check(scope):
    scope, drift, new = set(scope), set(), set()
    for ds in SIM_DATASETS:
        p_new, p_old = os.path.join(D, 'data', ds + '.json'), os.path.join(REPORT, 'data', ds + '.json')
        if not (os.path.exists(p_new) and os.path.exists(p_old)):
            continue
        col = 'id' if ds == 'builds' else 'build'
        sig = {}
        for which, p in (('new', p_new), ('old', p_old)):
            for r in json.load(open(p)):
                sig.setdefault(r[col], {'new': [], 'old': []})[which].append(numbers(r))
        for b, s in sig.items():
            if not s['old']:
                new.add(b)
            elif s['new'] and sorted(s['new']) != sorted(s['old']) and b not in scope:
                drift.add(b)
                print(f'drift: {b} in {ds}', file=sys.stderr)
    if new:
        print('not published yet:', ' '.join(sorted(new)))
    if drift:
        print(f'{len(drift)} builds outside the scope have sim numbers that differ from the published ones: '
              f'{" ".join(sorted(drift))}. Re-sim them (add them to the scope), or find why the packs and the report disagree.',
              file=sys.stderr)
        sys.exit(1)
    print('sim numbers match the published report for every build outside the scope')


def prune_events():
    """Delete the sample fights and raw packs of builds with no request (a removed preset)."""
    live, gone = set(all_ids()), []
    for p in glob.glob(os.path.join(REPORT, 'data', 'events', '*.json.gz')) + glob.glob(os.path.join(RAW, '*.tgz')):
        bid = os.path.basename(p).split('.')[0]
        if bid not in live:
            os.remove(p)
            gone.append(bid)
    return gone


def copy():
    manifest_p = os.path.join(REPORT, 'datasets.json')
    manifest = json.load(open(manifest_p))
    now = datetime.now(timezone.utc).isoformat(timespec='milliseconds').replace('+00:00', 'Z')
    changed = []
    for p in sorted(glob.glob(os.path.join(D, 'data', '*.json'))):
        ds = os.path.basename(p)[:-5]
        if ds not in manifest:
            print(f'skipping {ds}: not a dataset in datasets.json (add it there first)', file=sys.stderr)
            continue
        new = open(p, 'rb').read()
        target = os.path.join(REPORT, 'data', ds + '.json')
        if not os.path.exists(target) or open(target, 'rb').read() != new:
            open(target, 'wb').write(new)
            manifest[ds]['updated'] = {'at': now, 'by': 'Claude (forever-sim)'}
            changed.append(ds)
    open(manifest_p, 'w').write(json.dumps(manifest, indent=1, ensure_ascii=False) + '\n')  # as merge.mjs writes it
    live = set(all_ids())
    gone = prune_events()
    prov_p = os.path.join(REPORT, 'provenance.json')
    prov = json.load(open(prov_p))
    stale = [b for b in prov['builds'] if b not in live]
    if stale:
        prov['builds'] = {b: v for b, v in prov['builds'].items() if b in live}
        open(prov_p, 'w').write(json.dumps(prov, indent='\t', ensure_ascii=False) + '\n')  # as stale.mjs writes it
    print(f'{len(changed)} datasets changed: {" ".join(changed) or "none"}')
    if gone or stale:
        print('removed builds:', ' '.join(sorted(set(gone) | set(stale))))


if __name__ == '__main__':
    cmd, args = (sys.argv[1], sys.argv[2:]) if len(sys.argv) > 1 else ('', [])
    if cmd == 'pack':
        commit = os.environ.get('SIM_COMMIT') or subprocess.run(['git', 'rev-parse', 'HEAD'], cwd=ROOT, capture_output=True, text=True).stdout.strip()
        pack(args or sorted(os.path.basename(p)[:-5] for p in glob.glob(os.path.join(D, 'out', '*.json'))), commit)
    elif cmd == 'unpack':
        unpack()
    elif cmd == 'clean' and args:
        clean(args)
    elif cmd == 'variants' and args:
        variants(args)
    elif cmd == 'check':
        check(args)
    elif cmd == 'copy':
        copy()
    elif cmd == 'prune-events':
        prune_events()
    else:
        print(__doc__)
        sys.exit(1)
