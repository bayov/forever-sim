#!/usr/bin/env python3
"""build_bins.py: builds every binary the DPS Report runs (DPS, Tanking and Healing), stamps what they were built from,
and refuses sims whose binaries don't match the code.

  build_bins.py [--force] [--bin DIR]      build every binary in BINARIES into DIR (default $DASH_BIN, else
                                           $DASH_DIR/bin) from a clean worktree at ${SIM_COMMIT:-HEAD}, protos
                                           always regenerated, and write DIR/STAMP.json.
                                           Nothing is built when the stamp already matches. --force builds anyway
  build_bins.py hash [--commit REV]        the input hash at REV (default ${SIM_COMMIT:-HEAD}) and what it covers
  build_bins.py guard [--bin DIR] [--what TEXT] [--need NAME ...] [--mark ID ...] [--clear]
                                           the check before sims (pipeline.sh run and resume call it): exit 1 unless
                                           the binaries in --need match the stamp and the stamp matches the code

The input hash covers exactly what the binaries are built from: sim/ (not _test.go, .md, .results or testdata, which
go build never reads), proto/, cmd/wowsimcli, go.mod, go.sum, the Go sources of the built tools/ packages and those
they import (TOOL_PACKAGES), and all of assets/database (with_db embeds the database),
and the recipe (the table below, the build flags). Like stale.mjs's inputHash it hashes git blob ids at a commit, not
the commit, so a commit that changes none of these gives the same hash and needs no rebuild. The build uses
-trimpath -buildvcs=false, so a binary's sha256 doesn't change with the commit alone either (bench_races.py keys its
race-vector cache by that sha).

The guard (guard() here, from pipeline.sh run and resume, probe.py check, tankheal.py run and events.py) refuses when
the stamp is missing, when its input hash differs from the inputs at ${SIM_COMMIT:-HEAD}, when a binary's sha256
differs from the stamp's, or when the working tree has uncommitted changes under the stamped inputs (they are in no
binary). A missing binary is always an error: no script falls back to plain wowsimcli any more.
DASH_ALLOW_BIN_MISMATCH=<reason> lets a sim start anyway, with a loud warning. The reason is recorded with the
results (bin_override/<id>.json in $DASH_DIR for a build, <mode>/bin_override.json for a Tanking or Healing mode), the
raw packs carry it, and stale.mjs --record and --record-bench refuse to mark those builds or modes current.
"""
import argparse, datetime, glob, hashlib, json, os, re, shutil, subprocess, sys, tempfile, time

HERE = os.path.dirname(os.path.abspath(__file__))
sys.path.insert(0, HERE)
import dashpaths  # noqa: E402

ROOT = os.path.abspath(os.path.join(HERE, '..', '..'))

# Each binary and the package it is built from, all from the same worktree with no patches: the sim reads the report
# switches itself (sim/core/report_switches.go: DASH_NAMES, DASH_SPLITS, DASH_RACIALS_OFF, DASH_RACIALS_OFF_CLASS,
# DASH_ASSUME), so one wowsimcli serves every DPS step and the benches get every switch too.
BINARIES = {
    'wowsimcli': './cmd/wowsimcli',  # every DPS step, each with its switches in the environment
    'dashstats': './tools/dashstats',
    'tankbench': './tools/tankbench',
    'healbench': './tools/healbench',
}
DPS_BINS = ['wowsimcli', 'dashstats']
RETIRED = ['wowsimcli_r', 'wowsimcli_assume']  # copies of wowsimcli the patched recipe made, removed by a build
BENCH_BINS = {'tank': 'tankbench', 'heal': 'healbench'}
FLAGS = ['-tags=with_db', '-trimpath', '-buildvcs=false']
BUILD_ENV = {'CGO_ENABLED': '0'}  # pure Go, so the binary doesn't depend on the machine's C toolchain

INPUTS = ['sim', 'proto', 'cmd/wowsimcli', 'go.mod', 'go.sum', 'assets/database']
# the tools/ packages the binaries build or import (go list -deps, and build() checks no other module package is used): only
# their .go files count, not the bench requests next to them, which are read at run time
TOOL_PACKAGES = ['tools/dashstats', 'tools/tankbench', 'tools/healbench', 'tools/tankscore']
NOT_BUILT = re.compile(r'(_test\.go|\.md|\.results|\.results\.tmp)$|(^|/)testdata/')
STAMP = 'STAMP.json'
OVERRIDE = 'DASH_ALLOW_BIN_MISMATCH'


def git(root, *args):
    return subprocess.run(['git', *args], cwd=root, check=True, capture_output=True, text=True).stdout


def rev_parse(rev, root=ROOT):
    return git(root, 'rev-parse', '--verify', f'{rev}^{{commit}}').strip()


def sim_rev():
    return os.environ.get('SIM_COMMIT') or 'HEAD'


def recipe():
    """What else decides a binary besides the files: the table and the flags."""
    return {'binaries': dict(sorted(BINARIES.items())), 'flags': FLAGS, 'env': BUILD_ENV}


def is_input(path):
    if path.startswith('assets/database/'):
        return True
    if NOT_BUILT.search(path):
        return False
    if any(path.startswith(t + '/') for t in TOOL_PACKAGES):
        return path.endswith('.go')
    return any(path == p or path.startswith(p + '/') for p in INPUTS)


def group(path):
    """The name a changed input is reported under: sim/<package>, a tools/ package, or the top path."""
    parts = path.split('/')
    if parts[0] == 'sim':
        return '/'.join(parts[:2]) if len(parts) > 2 else 'sim/*.go'
    if path.startswith('tools/'):
        return '/'.join(parts[:2])
    return next(p for p in INPUTS if path == p or path.startswith(p + '/'))


def input_files(rev, root=ROOT):
    """[(blob, path)] of every input at a revision, sorted by path."""
    out = git(root, 'ls-tree', '-r', '-z', rev, '--', *INPUTS, *TOOL_PACKAGES)
    files = []
    for entry in out.split('\0'):
        if entry:
            meta, path = entry.split('\t', 1)
            if is_input(path):
                files.append((meta.split()[2], path))
    return sorted(files, key=lambda f: f[1])


def input_hash(rev, root=ROOT):
    """{'hash': the input hash, 'groups': {group: hash}} at a revision, in stale.mjs's inputHash style (sha256 of
    "<blob> <path>" lines) with the recipe as a last line."""
    files = input_files(rev, root)
    rec = 'recipe ' + json.dumps(recipe(), sort_keys=True)
    sha = lambda lines: hashlib.sha256('\n'.join(lines).encode()).hexdigest()  # noqa: E731
    by = {}
    for blob, path in files:
        by.setdefault(group(path), []).append(f'{blob} {path}')
    groups = {g: sha(lines) for g, lines in sorted(by.items())}
    groups['recipe'] = sha([rec])
    return {'hash': sha([f'{b} {p}' for b, p in files] + [rec]), 'groups': groups}


def dirty_inputs(root=ROOT):
    """Inputs with uncommitted changes in the working tree (untracked files included): no binary has them."""
    out = git(root, 'status', '--porcelain=v1', '-z', '--untracked-files=all', '--no-renames', '--', *INPUTS, *TOOL_PACKAGES)
    return sorted({e[3:] for e in out.split('\0') if len(e) > 3 and is_input(e[3:])})


def file_sha(path):
    h = hashlib.sha256()
    with open(path, 'rb') as f:
        for chunk in iter(lambda: f.read(1 << 20), b''):
            h.update(chunk)
    return h.hexdigest()


SHIM = 'timing_cli.py'  # pipeline.sh run's timing shims ($DASH_DIR/timing/bin, timing_cli.py)


def real_bin(path):
    """The binary a path runs. pipeline.sh run points DASH_BIN at $DASH_DIR/timing/bin, whose entries are links to
    timing_cli.py that run the binary of the same name in TIMING_REAL_BIN: the stamp, the guard and every key hashed from
    a binary (bench_races.py's result cache, probe.py's _key) follow the shim to that binary."""
    real = os.environ.get('TIMING_REAL_BIN')
    if real and os.path.basename(os.path.realpath(path)) == SHIM:
        return os.path.join(real, os.path.basename(path))
    return path


def real_dir(bin_dir):
    """bin_dir, or TIMING_REAL_BIN when bin_dir is the timing shims' directory (it has no stamp of its own)."""
    real = os.environ.get('TIMING_REAL_BIN')
    if real and os.path.isdir(bin_dir) and not os.path.exists(os.path.join(bin_dir, STAMP)) and any(
            os.path.basename(os.path.realpath(os.path.join(bin_dir, n))) == SHIM for n in os.listdir(bin_dir)):
        return os.path.abspath(real)
    return bin_dir


def load_stamp(bin_dir):
    p = os.path.join(bin_dir, STAMP)
    if not os.path.exists(p):
        return None
    with open(p) as f:
        return json.load(f)


def few(xs, n=6):
    return ', '.join(xs[:n]) + (f' and {len(xs) - n} more' if len(xs) > n else '')


def problems(bin_dir, need, rev=None, root=ROOT):
    """Why the binaries in `need` can't be trusted for the code at rev (default ${SIM_COMMIT:-HEAD}), [] when they can."""
    rev = rev or sim_rev()
    bin_dir = real_dir(bin_dir)
    stamp = load_stamp(bin_dir)
    if stamp is None:
        return [f'{bin_dir} has no {STAMP}: these binaries were not built by build_bins.py']
    out = []
    now = input_hash(rev, root)
    if stamp.get('inputHash') != now['hash']:
        old = stamp.get('inputs', {})
        changed = sorted(g for g in set(old) | set(now['groups']) if old.get(g) != now['groups'].get(g))
        out.append(f'built from other inputs than {rev} has (built at {stamp.get("commit", "?")[:10]}): {few(changed) or "the recipe"} changed')
    for name in need:
        rec = stamp.get('binaries', {}).get(name)
        path = os.path.join(bin_dir, name)
        if rec is None:
            out.append(f'{name} is not in the stamp')
        elif os.path.exists(path) and file_sha(path) != rec.get('sha256'):
            out.append(f'{name} is not the binary the stamp records (its sha256 differs): rebuilt or replaced by hand')
    dirty = dirty_inputs(root)
    if dirty:
        out.append(f'uncommitted changes under the stamped inputs, which no binary has: {few(dirty)}')
    return out


def guard(bin_dir, need, what, rev=None, root=ROOT):
    """Refuse a sim run (exit 1) unless the binaries match the code. Returns None when they do, or the override record
    when DASH_ALLOW_BIN_MISMATCH is set (mark it on what the run makes: mark_builds, mark). Timing shims in bin_dir
    (pipeline.sh run) are followed to the real binaries and their stamp (real_dir)."""
    bin_dir = real_dir(os.path.abspath(bin_dir))
    missing = [n for n in need if not os.path.isfile(os.path.join(bin_dir, n))]
    if missing:
        sys.exit(f'{what}: {few(missing)} missing from {bin_dir}. Build them: tools/dps_report/pipeline.sh bins')
    probs = problems(bin_dir, need, rev, root)
    if not probs:
        return None
    reason = os.environ.get(OVERRIDE, '').strip()
    lines = '\n'.join('  - ' + p for p in probs)
    if not reason:
        sys.exit(f'{what}: the binaries in {bin_dir} don\'t match the code at {rev or sim_rev()}:\n{lines}\n'
                 f'Commit or drop the changes, then rebuild: tools/dps_report/pipeline.sh bins (build_bins.py). '
                 f'To sim anyway, set {OVERRIDE}=<reason>: the results are marked and stale.mjs won\'t record them current.')
    bar = '!' * 100
    print(f'{bar}\n{what}: SIMMING WITH BINARIES THAT DON\'T MATCH THE CODE ({OVERRIDE}={reason})\n{lines}\n'
          f'The reason goes with the results and their raw packs, and stale.mjs --record will refuse them.\n{bar}', file=sys.stderr, flush=True)
    return {'reason': reason, 'what': what, 'problems': probs, 'bin': bin_dir, 'rev': rev or sim_rev(),
            'at': datetime.datetime.now(datetime.timezone.utc).strftime('%Y-%m-%dT%H:%M:%SZ')}


def mark(path, rec):
    os.makedirs(os.path.dirname(path), exist_ok=True)
    with open(path, 'w') as f:
        json.dump(rec, f, indent=1, sort_keys=True)


def override_path(bid, d=None):
    return os.path.join(d or dashpaths.D, 'bin_override', bid + '.json')


def mark_builds(ids, rec, d=None):
    for bid in ids:
        mark(override_path(bid, d), rec)


def clear_builds(ids, d=None):
    for bid in ids:
        if os.path.exists(override_path(bid, d)):
            os.remove(override_path(bid, d))


def require(path):
    """A binary a script runs: missing is an error, never a fallback to another binary. A timing shim counts only when
    the binary behind it is there (real_bin)."""
    if not os.path.isfile(real_bin(path)):
        sys.exit(f'{path} is missing. Build every report binary: tools/dps_report/pipeline.sh bins (build_bins.py)')
    return path


# ---------------------------------------------------------------- build

def tool_env():
    gobin = os.path.join(subprocess.run(['go', 'env', 'GOPATH'], capture_output=True, text=True, check=True).stdout.strip(), 'bin')
    return dict(os.environ, PATH=os.environ['PATH'] + os.pathsep + gobin, GOMAXPROCS=os.environ.get('BUILD_P', '2'), **BUILD_ENV)


def protoc_cmd():
    if os.environ.get('PROTOC'):
        return [os.environ['PROTOC']]
    return ['protoc'] if shutil.which('protoc') else ['npx', '--yes', 'protoc']


def run(cmd, cwd, env=None):
    r = subprocess.run(cmd, cwd=cwd, env=env, capture_output=True, text=True)
    if r.returncode:
        sys.exit(f'build_bins: {" ".join(cmd)} failed in {cwd}:\n{(r.stdout + r.stderr)[-3000:]}')
    return r.stdout.strip()


def check_deps(wt, pkgs, env):
    """Every package of this module the binaries use is in the stamped inputs (else the stamp would miss a change)."""
    out = run(['go', 'list', '-deps', '-tags=with_db', '-f', '{{if not .Standard}}{{.ImportPath}}|{{.Dir}}|{{join .EmbedFiles ","}}{{end}}', *pkgs], wt, env)
    bad, top = [], os.path.realpath(wt)
    for line in filter(None, out.splitlines()):
        imp, d, embeds = line.split('|')
        d = os.path.realpath(d)
        if not d.startswith(top + os.sep):
            continue  # a module dependency, pinned by go.sum
        rel = os.path.relpath(d, top)
        files = [f'{rel}/x.go'] + [f'{rel}/{e}' for e in embeds.split(',') if e]
        if not all(is_input(f) for f in files):
            bad.append(imp)
    if bad:
        sys.exit(f'build_bins: {few(bad)} are built in but not in the stamped inputs: add them to INPUTS or TOOL_PACKAGES')


def build(bin_dir, rev=None, force=False, root=ROOT):
    rev = rev or sim_rev()
    commit = rev_parse(rev, root)
    bin_dir = real_dir(os.path.abspath(bin_dir))  # never over the timing shims
    want = input_hash(commit, root)
    stamp = load_stamp(bin_dir)
    if not force and stamp and stamp.get('inputHash') == want['hash'] and all(
            os.path.isfile(os.path.join(bin_dir, n)) and file_sha(os.path.join(bin_dir, n)) == stamp['binaries'].get(n, {}).get('sha256')
            for n in BINARIES):
        print(f'binaries in {bin_dir} are up to date (inputs {want["hash"][:12]}, built at {stamp["commit"][:10]})')
        return False
    dirty = dirty_inputs(root)
    if dirty:
        print(f'note: uncommitted changes under the inputs are not built (commit them first): {few(dirty)}', file=sys.stderr)
    t0 = time.time()
    env = tool_env()
    tmp = tempfile.mkdtemp(prefix='binbuild_')
    built, info = {}, {}
    try:
        wt = os.path.join(tmp, 'wt')
        git(root, 'worktree', 'add', '--detach', '--quiet', wt, commit)
        try:
            for f in glob.glob(os.path.join(wt, 'sim', 'core', 'proto', '*.pb.go')):
                os.remove(f)
            protos = sorted(os.path.relpath(p, wt) for p in glob.glob(os.path.join(wt, 'proto', '*.proto')))
            run(protoc_cmd() + ['-I=proto', '--go_out=sim/core'] + protos, wt, env)
            if not glob.glob(os.path.join(wt, 'sim', 'core', 'proto', '*.pb.go')):
                sys.exit('build_bins: protoc wrote no sim/core/proto/*.pb.go')
            pkgs = sorted(set(BINARIES.values()))
            check_deps(wt, pkgs, env)
            for pkg in pkgs:
                out = os.path.join(tmp, 'out', os.path.basename(pkg))
                t = time.time()
                run(['nice', '-n', '10', 'go', 'build', '-p', env['GOMAXPROCS'], *FLAGS, '-o', out, pkg], wt, env)
                print(f'built {pkg} in {time.time() - t:.0f}s', flush=True)
                for n, p in BINARIES.items():
                    if p == pkg:
                        built[n] = out
            info['protoc'] = run(protoc_cmd() + ['--version'], wt, env)
            info['protoc-gen-go'] = run(['protoc-gen-go', '--version'], wt, env)
        finally:
            subprocess.run(['git', 'worktree', 'remove', '--force', wt], cwd=root, capture_output=True)
        os.makedirs(bin_dir, exist_ok=True)
        bins = {}
        for n in BINARIES:  # each next to its final path, then renamed over it, so a running sim keeps its old file
            part = os.path.join(bin_dir, f'.{n}.part')
            shutil.copyfile(built[n], part)
            os.chmod(part, 0o755)
            os.replace(part, os.path.join(bin_dir, n))
            bins[n] = {'package': BINARIES[n], 'sha256': file_sha(os.path.join(bin_dir, n))}
        for n in RETIRED:  # binaries of the old recipe, which no script runs any more
            if os.path.exists(os.path.join(bin_dir, n)):
                os.remove(os.path.join(bin_dir, n))
    finally:
        shutil.rmtree(tmp, ignore_errors=True)
        subprocess.run(['git', 'worktree', 'prune'], cwd=root, capture_output=True)
    stamp = {'format': 2, 'inputHash': want['hash'], 'inputs': want['groups'], 'commit': commit,
             'builtAt': datetime.datetime.now(datetime.timezone.utc).strftime('%Y-%m-%dT%H:%M:%SZ'),
             'go': run(['go', 'version'], root, env), 'protoc': info.get('protoc'), 'protocGenGo': info.get('protoc-gen-go'),
             'recipe': recipe(), 'binaries': bins}
    with open(os.path.join(bin_dir, STAMP), 'w') as f:
        json.dump(stamp, f, indent=1, sort_keys=True)
        f.write('\n')
    print(f'{len(bins)} binaries into {bin_dir} at {commit[:10]} in {time.time() - t0:.0f}s (inputs {want["hash"][:12]})')
    return True


def main(argv):
    ap = argparse.ArgumentParser(description='Build, stamp and check the DPS Report binaries (see the module docstring).')
    ap.add_argument('cmd', nargs='?', default='build', choices=['build', 'hash', 'guard'])
    ap.add_argument('--bin', default=dashpaths.BIN, help='the binaries directory (default $DASH_BIN, else $DASH_DIR/bin)')
    ap.add_argument('--force', action='store_true')
    ap.add_argument('--commit', help='hash: the revision (default ${SIM_COMMIT:-HEAD})')
    ap.add_argument('--what', default='sims', help='guard: what is starting, for the messages')
    ap.add_argument('--need', nargs='*', default=[], help='guard: the binaries it will run')
    ap.add_argument('--mark', nargs='*', default=[], help='guard: the builds whose results get the override record')
    ap.add_argument('--clear', action='store_true', help='guard: drop the override records of --mark when the binaries match (a run that re-sims them whole)')
    a = ap.parse_args(argv)
    if a.cmd == 'hash':
        h = input_hash(a.commit or sim_rev())
        print(json.dumps(h, indent=1))
    elif a.cmd == 'guard':
        rec = guard(a.bin, a.need, a.what)
        if rec:
            mark_builds(a.mark, rec)
        elif a.clear:
            clear_builds(a.mark)
    else:
        build(a.bin, force=a.force)


if __name__ == '__main__':
    main(sys.argv[1:])
