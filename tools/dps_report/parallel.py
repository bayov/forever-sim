#!/usr/bin/env python3
"""parallel.py: a full DPS Report regeneration over several cloud sessions at once (the regenerate-dps-report skill's
"Parallel full regeneration"). The coordinating session plans, spawns one worker session per part, waits for their
hand-backs, merges, then publishes and deploys as for one run. Workers never publish or deploy.

  parallel.py plan --run <run> --commit <sha> [--heal-groups N] [--heal-presets <stem,...>] [--log <file>...] [--dry]
        the worker sessions: dps (every DPS build, pipeline.sh run), tank (the Tanking mode whole), heal-0 (the Healing
        main bench, handed to the others, then its own presets) and heal-1.. (each a group of Healing presets: every
        priest preset alone, the rest balanced by cost). Prints each session's command and estimated time, and writes
        plan.json to $DASH_DIR/parallel/<run>/ and to the branch dash-raw/<run>/plan (--dry: print only)
  parallel.py worker <run> <session> --commit <sha> [--main] [--presets <stem,...>] [--carries-from <session>]
        one session's whole job, in the background with its log in a file: checks HEAD is <sha>, builds the binaries
        (pipeline.sh bins), sims its part, packs it and pushes it to the branch dash-raw/<run>/<session>. A Healing
        session without --main first waits (polling the remote, no turns spent) for heal-main.tgz on the --carries-from
        branch: the main bench outputs its race sims start from (tankheal.py run heal --carries)
  parallel.py merge <run>
        fetch the plan and every session's branch into $DASH_DIR/parallel/<run>/, check that each planned part is there
        and that every part was simmed at the same commit, sim inputs, binaries (STAMP.json) and DASH_SPLITS as each other
        and as this checkout's binaries would be, then: the DPS pack into $DASH_DIR (pipeline.sh unpack, into emptied
        outputs for a full run, as a fresh run would leave them), and tankheal.py merge for Tanking and Healing.
        Prints the publish command
  parallel.py clean <run>
        delete the run's dash-raw/<run>/* branches (after the deploy). The cloud sessions' git proxy refuses branch
        deletes, so there it names the branches for the user to delete

Transport: each session pushes its part to its own branch, dash-raw/<run>/<session> (the shard branches' names,
orchestration/dash_raw_store.md), one orphan commit a file set made with git plumbing, so the checkout is never touched
and nothing reaches master. A part is small (the 2026-11-18 Healing pack is 13 KB, a full DPS pack a few MB).

Why the outputs are the same bytes as one run's: every sim's seed, iterations and DASH_SPLITS are fixed, a preset's race
sims read only its own request and the main bench's carries, and tankheal.py merge puts races.jsonl in the order one
run writes it (race_lines_sorted). The main bench runs once (heal-0) and every other Healing part reads its outputs.
"""
import argparse, glob, json, os, subprocess, sys, time

HERE = os.path.dirname(os.path.abspath(__file__))
ROOT = os.path.abspath(os.path.join(HERE, '..', '..'))
sys.path.insert(0, HERE)
import dashpaths  # noqa: E402

D = dashpaths.D
REMOTE = os.environ.get('PARALLEL_REMOTE', 'origin')
PREFIX = 'dash-raw'

# ---------------------------------------------------------------- costs and the plan
#
# A preset's race sims cost about its race rows (one a race and racial variant) times a role's seconds a row. The rates
# are the bench probes measured on 4 cores on 2026-11-18 (README "Bench probes": priest_holy_raid 565s for 26 rows,
# shaman_restoration 120s for 10, warrior_protection_balanced 80s for 29), and FULL_X scales a probe to the full run:
# the 2026-11-18 Healing run's race sims took 4,123s wall at 2 presets at a time, about 8,250 preset-seconds, against
# 2,235 for this model of its 9 presets. --log takes measured times instead, from a run's log lines
# '<mode> races <stem>: <n> rows, <s> s' (the same 2-at-a-time seconds).
ROW_S = {('heal', 'raid'): 565 / 26, ('heal', 'tank'): 120 / 10, ('tank', None): 80 / 29}
FULL_X = 3.7
MAIN_S = {'heal': 1090, 'tank': 20}  # the main bench, 2026-11-18 Healing and 2026-11-17 Tanking run.json
MODE_S = {'dps': 80 * 60, 'tank': 3013}  # DPS: pipeline.sh plan's 310 core-minutes on 4 cores; Tanking: 2026-11-17 run.json
SLOTS = 2  # tankheal.race_sims runs 2 presets at a time


def preset_rows(mode, stem):
    import tankheal, bench_races
    req = json.load(open(tankheal.req_path(mode, stem)))
    cls = tankheal.player(req)['class'][5:]
    return sum(1 + len(bench_races.variants(cls, r)) for r in tankheal.class_races()[cls])


def logged_costs(paths, mode):
    """{stem: seconds} from run logs' '<mode> races <stem>: <n> rows, <s> s' lines (the last of each preset)."""
    import re
    out = {}
    for p in paths or ():
        for ln in open(p, errors='replace'):
            m = re.match(rf'^{mode} races (\w+): \d+ rows, (\d+) s', ln)
            if m:
                out[m.group(1)] = float(m.group(2))
    return out


def costs(mode, stems, logs=()):
    """Each preset's race sims in seconds of a full run, 2 presets at a time: measured (--log) or modelled."""
    import tankheal
    got = logged_costs(logs, mode)
    role = (lambda s: tankheal.heal_role(s)) if mode == 'heal' else (lambda s: None)
    return {s: got.get(s) or round(preset_rows(mode, s) * ROW_S[(mode, role(s))] * FULL_X) for s in stems}


def wall(cs, slots=SLOTS):
    """The wall time of presets of these costs run `slots` at a time, longest first (race_sims' pool)."""
    load = [0.0] * slots
    for c in sorted(cs, reverse=True):
        load[load.index(min(load))] += c
    return max(load) if cs else 0.0


def split(cost, groups=None, alone=lambda s: s.startswith('priest_')):
    """The presets in groups, one a worker session: each `alone` preset (the priests: the most race rows) in a group of
    its own, the rest spread over the other groups longest first, each to the group that would end soonest. groups:
    how many in all (default: as few as keep every group within the longest lone preset's time, at least one)."""
    lone = sorted((s for s in cost if alone(s)), key=lambda s: (-cost[s], s))
    rest = sorted((s for s in cost if not alone(s)), key=lambda s: (-cost[s], s))
    if groups is not None and groups <= len(lone):  # too few sessions for the priests alone: one balanced spread
        rest, lone = sorted(cost, key=lambda s: (-cost[s], s)), []
        n = max(1, groups)
    elif groups is not None:
        n = groups - len(lone)
    else:
        cap = max([cost[s] for s in lone] or [cost[s] for s in rest])  # no group longer than the longest lone preset
        n = 1
        while rest and n < len(rest) and max(wall([cost[s] for s in g]) for g in spread(rest, cost, n)) > cap:
            n += 1
    out = [[s] for s in lone] + (spread(rest, cost, n) if rest else [])
    return [g for g in out if g]


def spread(stems, cost, n):
    gs = [[] for _ in range(n)]
    for s in stems:  # longest first, each to the group that would end soonest with it
        i = min(range(n), key=lambda j: (wall([cost[x] for x in gs[j] + [s]]), j))
        gs[i].append(s)
    return [sorted(g) for g in gs]


def plan(run, commit, heal_groups=None, heal_presets=None, logs=(), modes=('dps', 'tank', 'heal')):
    """The worker sessions of a run: [{'session', 'mode', 'presets', 'main', 'carriesFrom', 'estimate', 'command'}]."""
    import tankheal
    sessions = []
    cmd = lambda name, *more: ' '.join(['python3 tools/dps_report/parallel.py worker', run, name, '--commit', commit, *more])  # noqa: E731
    if 'dps' in modes:
        sessions.append({'session': 'dps', 'mode': 'dps', 'ids': req_ids(), 'estimate': MODE_S['dps'], 'command': cmd('dps')})
    if 'tank' in modes:
        sessions.append({'session': 'tank', 'mode': 'tank', 'presets': tankheal.presets('tank'), 'main': True,
                         'estimate': MODE_S['tank'], 'command': cmd('tank')})
    if 'heal' in modes:
        stems = [s for s in tankheal.presets('heal') if heal_presets is None or s in heal_presets]
        cost = costs('heal', stems, logs)
        for k, g in enumerate(split(cost, heal_groups)):
            name = f'heal-{k}'
            more = ['--main', '--presets', ','.join(g)] if k == 0 else ['--presets', ','.join(g), '--carries-from', 'heal-0']
            sessions.append({'session': name, 'mode': 'heal', 'presets': g, 'main': k == 0, 'carriesFrom': None if k == 0 else 'heal-0',
                             'raceSeconds': {s: cost[s] for s in g},
                             'estimate': round(MAIN_S['heal'] + wall([cost[s] for s in g])), 'command': cmd(name, *more)})
    return {'run': run, 'commit': commit, 'modes': list(modes), 'healAll': heal_presets is None, 'sessions': sessions}


def req_ids():
    return sorted(os.path.basename(p)[:-10] for p in glob.glob(os.path.join(D, 'req', '*.meta.json')))


# ---------------------------------------------------------------- git transport

def git(*args, data=None, env=None, check=True):
    r = subprocess.run(['git', *args], cwd=ROOT, input=data, capture_output=True, env=env)
    if check and r.returncode:
        sys.exit(f'git {" ".join(args)} failed: {r.stderr.decode(errors="replace")[-1500:]}')
    return r.stdout


def ref_of(run, session):
    return f'refs/heads/{PREFIX}/{run}/{session}'


def remote_tips(pattern):
    """{ref: sha} on the remote matching a ref pattern."""
    out = {}
    for ln in git('ls-remote', REMOTE, pattern).decode().splitlines():
        sha, ref = ln.split('\t')
        out[ref] = sha
    return out


def fetch(ref):
    """The remote ref's tip, fetched: its sha, or None when the branch isn't there."""
    sha = remote_tips(ref).get(ref)
    if sha:
        for i in range(5):
            if subprocess.run(['git', 'fetch', '-q', REMOTE, ref], cwd=ROOT, capture_output=True).returncode == 0:
                break
            time.sleep(5 * 2 ** i)
        else:
            sys.exit(f'git fetch {REMOTE} {ref} failed five times')
    return sha


def tree_files(sha):
    """{name: blob sha} at the top of a commit."""
    out = {}
    for ent in git('ls-tree', '-z', sha).split(b'\0'):
        if ent:
            meta, name = ent.split(b'\t', 1)
            out[name.decode()] = meta.split()[2].decode()
    return out


def push(run, session, files):
    """files {name: local path} onto the branch dash-raw/<run>/<session>: one commit on its tip (an orphan for a new
    branch) holding its files and these, made with git plumbing, so the checkout and its index are never touched."""
    ref = ref_of(run, session)
    parent = fetch(ref)
    ents = {n: b for n, b in (tree_files(parent) if parent else {}).items()}
    for name, path in files.items():
        ents[name] = git('hash-object', '-w', path).decode().strip()
    tree = git('mktree', data=''.join(f'100644 blob {b}\t{n}\n' for n, b in sorted(ents.items())).encode()).decode().strip()
    env = dict(os.environ)
    if not git('config', 'user.email', check=False).strip():
        env.update(GIT_AUTHOR_NAME='dps report worker', GIT_AUTHOR_EMAIL='noreply@anthropic.com',
                   GIT_COMMITTER_NAME='dps report worker', GIT_COMMITTER_EMAIL='noreply@anthropic.com')
    commit = git('commit-tree', tree, *(['-p', parent] if parent else []), '-m', f'dps report {run}: {session} ({", ".join(sorted(files))})',
                 env=env).decode().strip()
    for i in range(5):
        r = subprocess.run(['git', 'push', '-q', REMOTE, f'{commit}:{ref}'], cwd=ROOT, capture_output=True)
        if r.returncode == 0:
            print(f'pushed {", ".join(sorted(files))} to {ref[len("refs/heads/"):]}', flush=True)
            return commit
        time.sleep(5 * 2 ** i)
    sys.exit(f'git push to {ref} failed five times: {r.stderr.decode(errors="replace")[-1000:]}')


def wait_for(run, session, name, dest, timeout=4 * 3600, every=60):
    """Wait until the branch dash-raw/<run>/<session> holds `name`, then write it to dest."""
    t0, ref = time.time(), ref_of(run, session)
    while True:
        sha = fetch(ref)
        if sha and name in tree_files(sha):
            open(dest, 'wb').write(git('cat-file', 'blob', tree_files(sha)[name]))
            print(f'{name} from {ref[len("refs/heads/"):]} after {round(time.time() - t0)} s', flush=True)
            return dest
        if time.time() - t0 > timeout:
            sys.exit(f'{name} never reached {ref} ({timeout} s): is that session stuck?')
        time.sleep(every)


# ---------------------------------------------------------------- the worker

def sh(*cmd, env=None):
    print('$ ' + ' '.join(cmd), flush=True)
    t = time.time()
    r = subprocess.run(list(cmd), cwd=ROOT, env=env)
    if r.returncode:
        sys.exit(f'{" ".join(cmd[:3])} failed ({r.returncode})')
    print(f'  ({round(time.time() - t)} s)', flush=True)


def dps_manifest(ids, seconds):
    import build_bins
    return {'mode': 'dps', 'commit': head(), 'splits': int(os.environ.get('DASH_SPLITS', '4')), 'ids': ids, 'seconds': seconds,
            'stamp': build_bins.summary(os.environ.get('DASH_BIN') or dashpaths.BIN, ['wowsimcli', 'dashstats'])}


def head():
    return git('rev-parse', 'HEAD').decode().strip()


def worker(run, session, commit, main=False, presets=None, carries_from=None):
    t0 = time.time()
    if not head().startswith(commit):
        sys.exit(f'HEAD is {head()[:10]}, not the run\'s commit {commit}: git fetch origin && git checkout --detach {commit}')
    if os.environ.get('DASH_ALLOW_BIN_MISMATCH'):
        sys.exit('a worker never sims with DASH_ALLOW_BIN_MISMATCH: its part would be refused at the merge')
    os.environ.setdefault('DASH_SPLITS', '4')
    out = os.path.join(D, 'parallel', run)
    os.makedirs(out, exist_ok=True)
    pl, th = os.path.join(HERE, 'pipeline.sh'), os.path.join(HERE, 'tankheal.py')
    sh(pl, 'bins')
    env = dict(os.environ, TANKHEAL_BIN=os.environ.get('DASH_BIN') or dashpaths.BIN)
    if session == 'dps':
        ids = req_ids()
        sh(pl, 'run', *ids)
        tgz = os.path.join(out, 'dps.tgz')
        sh(pl, 'pack', tgz, *ids)
        man = os.path.join(out, 'dps.json')
        json.dump(dps_manifest(ids, round(time.time() - t0)), open(man, 'w'), indent=1)
        push(run, session, {'dps.tgz': tgz, 'dps.json': man})
    elif session.startswith('tank'):
        sh(sys.executable, th, 'run', 'tank', '--part', *(['--presets', presets] if presets else []), env=env)
        sh(sys.executable, th, 'part', 'tank', os.path.join(out, session + '.tgz'))
        push(run, session, {session + '.tgz': os.path.join(out, session + '.tgz')})
    elif session.startswith('heal'):
        carries = os.path.join(out, 'heal-main.tgz')
        if main:  # the main bench first, pushed at once: the other Healing sessions wait for it
            sh(sys.executable, th, 'run', 'heal', '--main-only', env=env)
            sh(sys.executable, th, 'part', 'heal', carries)
            push(run, session, {'heal-main.tgz': carries})
        else:
            wait_for(run, carries_from, 'heal-main.tgz', carries)
        if presets:
            sh(sys.executable, th, 'run', 'heal', '--presets', presets, '--carries', carries, env=env)
            sh(sys.executable, th, 'part', 'heal', os.path.join(out, session + '.tgz'))
            push(run, session, {session + '.tgz': os.path.join(out, session + '.tgz')})
    else:
        sys.exit(f'no such session: {session} (dps, tank, heal-<k>)')
    print(f'worker {session} of {run} done in {round((time.time() - t0) / 60)} min: branch {PREFIX}/{run}/{session}', flush=True)


# ---------------------------------------------------------------- the coordinator's merge

def stamp_problems(manifests):
    """Why parts made in different sessions can't go together: [] when they were simmed alike."""
    out, ref = [], manifests[0]
    for m in manifests[1:]:
        for k in ('commit', 'splits'):
            if m[1].get(k) != ref[1].get(k):
                out.append(f'{m[0]} and {ref[0]}: other {k} ({m[1].get(k)}, {ref[1].get(k)})')
        a, b = m[1].get('stamp') or {}, ref[1].get('stamp') or {}
        if not a or not b:
            out.append(f'{m[0] if not a else ref[0]}: no binary stamp')
        elif a.get('inputHash') != b.get('inputHash') or a.get('go') != b.get('go'):
            out.append(f'{m[0]} and {ref[0]}: binaries of other inputs or Go ({a.get("inputHash", "")[:12]} {a.get("go")}, '
                       f'{b.get("inputHash", "")[:12]} {b.get("go")})')
    return out


def gather(run, dest):
    """The plan and every session's files into dest: (plan, {name: path})."""
    import tankheal
    os.makedirs(dest, exist_ok=True)
    tips = remote_tips(f'refs/heads/{PREFIX}/{run}/*')
    got, where = {}, {}
    for ref in sorted(tips):
        fetch(ref)
        for name, blob in tree_files(tips[ref]).items():
            if name in where:
                sys.exit(f'{name} is on both {where[name]} and {ref}')
            where[name] = ref
            got[name] = os.path.join(dest, name)
            open(got[name], 'wb').write(git('cat-file', 'blob', blob))
    if 'plan.json' not in got:
        sys.exit(f'no plan on {PREFIX}/{run}/plan: parallel.py plan --run {run} ... first')
    pl = json.load(open(got['plan.json']))
    missing = []
    for s in pl['sessions']:
        want = ['dps.tgz', 'dps.json'] if s['mode'] == 'dps' else [s['session'] + '.tgz'] if s['mode'] == 'tank' else \
            (['heal-main.tgz'] if s.get('main') else []) + ([s['session'] + '.tgz'] if s.get('presets') else [])
        missing += [f'{s["session"]}: {n}' for n in want if n not in got]
    if missing:
        sys.exit('parts not pushed yet (is a session still running or stuck?):\n' + '\n'.join('  - ' + m for m in missing))
    manifests = []
    if 'dps.json' in got:
        manifests.append(('dps', json.load(open(got['dps.json']))))
    for name in sorted(got):
        if name.endswith('.tgz') and name != 'dps.tgz':
            info, _ = tankheal.read_part(got[name])
            if info is None:
                sys.exit(f'{name}: no part.json')
            manifests.append((name, info))
            s = next((x for x in pl['sessions'] if x['session'] + '.tgz' == name), None)
            if s and s['mode'] == 'heal' and sorted(info.get('presets') or []) != sorted(s['presets']):
                sys.exit(f'{name} raced {info.get("presets")}, the plan gave it {s["presets"]}')
    probs = stamp_problems(manifests)
    if manifests and manifests[0][1].get('commit') != pl['commit'] and not manifests[0][1].get('commit', '').startswith(pl['commit']):
        probs.append(f'the parts were simmed at {manifests[0][1].get("commit")}, the plan says {pl["commit"]}')
    if probs:
        sys.exit('the parts were not simmed alike:\n' + '\n'.join('  - ' + p for p in probs))
    return pl, got, manifests


OUTPUTS = ('out', 'names', 'weights', 'mana', 'rv', 'asm', 'abl', 'probe', 'probe_check', 'data', 'vec', 'tank', 'heal', 'bin_override')


def merge(run):
    import shutil
    import build_bins
    dest = os.path.join(D, 'parallel', run)
    pl, got, manifests = gather(run, dest)
    here = build_bins.input_hash(build_bins.sim_rev())['hash']
    stamp = (manifests[0][1].get('stamp') or {}) if manifests else {}
    if stamp and stamp.get('inputHash') != here:
        sys.exit(f'the parts\' binaries were built from other sim inputs than {build_bins.sim_rev()} has: check out the run\'s commit '
                 f'{pl["commit"]} (or a commit with the same sim inputs), or set SIM_COMMIT')
    dps = next((s for s in pl['sessions'] if s['mode'] == 'dps'), None)
    if dps and sorted(json.load(open(got['dps.json']))['ids']) == req_ids():
        # a full run's DASH_DIR: the outputs emptied (hydrate's list), the inputs linked when $DASH_DIR is elsewhere
        if os.path.realpath(D) != os.path.realpath(HERE):
            for f in ('req', 'forever_tt_cache', 'fortt_ids.json', 'talent_ids.json', 'asm_apls', 'bin'):
                if not os.path.exists(os.path.join(D, f)):
                    os.symlink(os.path.join(HERE, f), os.path.join(D, f))
        for d in OUTPUTS:
            shutil.rmtree(os.path.join(D, d), ignore_errors=True)
        print(f'emptied the outputs in {D} for the full run\'s parts', flush=True)
    elif not os.path.isdir(os.path.join(D, 'out')):
        sys.exit(f'a partial DPS run (or none) merges into a hydrated $DASH_DIR: pipeline.sh hydrate first')
    if dps:
        sh(os.path.join(HERE, 'pipeline.sh'), 'unpack', got['dps.tgz'])
    th = os.path.join(HERE, 'tankheal.py')
    for mode in ('tank', 'heal'):
        ss = [s for s in pl['sessions'] if s['mode'] == mode]
        if not ss:
            continue
        parts = [got[n] for n in sorted(got) if n.endswith('.tgz') and n.startswith(mode)]
        keep = mode == 'heal' and not pl.get('healAll', True)
        sh(sys.executable, th, 'merge', mode, *(['--keep-rest'] if keep else []), *parts)
    ids = ' '.join(json.load(open(got['dps.json']))['ids']) if dps else ''
    print(f'\nnext: SIM_COMMIT={pl["commit"]} tools/dps_report/pipeline.sh publish {ids}\n'
          f'then check, commit, push and deploy as for one run; after the deploy: parallel.py clean {run}')


def clean(run):
    refs = sorted(remote_tips(f'refs/heads/{PREFIX}/{run}/*'))
    if refs:
        git('push', REMOTE, '--delete', *refs, check=False)
        left = sorted(remote_tips(f'refs/heads/{PREFIX}/{run}/*'))
        if left:  # the cloud sessions' git proxy refuses branch deletes (HTTP 403): the user deletes them on GitHub
            sys.exit(f'the remote kept these branches; ask the user to delete them: {" ".join(r[11:] for r in left)}')
    print(f'deleted {len(refs)} branches of {run}')


def main(argv):
    ap = argparse.ArgumentParser(description='A DPS Report regeneration over several cloud sessions (module docstring).')
    sub = ap.add_subparsers(dest='cmd', required=True)
    p = sub.add_parser('plan')
    p.add_argument('--run', required=True)
    p.add_argument('--commit', required=True)
    p.add_argument('--heal-groups', type=int)
    p.add_argument('--heal-presets')
    p.add_argument('--modes', default='dps,tank,heal')
    p.add_argument('--log', action='append', default=[])
    p.add_argument('--dry', action='store_true')
    w = sub.add_parser('worker')
    w.add_argument('run')
    w.add_argument('session')
    w.add_argument('--commit', required=True)
    w.add_argument('--main', action='store_true')
    w.add_argument('--presets')
    w.add_argument('--carries-from', default='heal-0')
    for c in ('merge', 'clean'):
        sub.add_parser(c).add_argument('run')
    a = ap.parse_args(argv)
    if a.cmd == 'plan':
        pl = plan(a.run, a.commit, a.heal_groups, set(a.heal_presets.split(',')) if a.heal_presets else None, a.log, a.modes.split(','))
        for s in pl['sessions']:
            print(f'{s["session"]:8} about {round(s["estimate"] / 60):3} min  {" ".join(s.get("presets") or []) if s["mode"] != "dps" else str(len(s["ids"])) + " builds"}')
            print(f'         {s["command"]}')
        print(f'wall: about {round(max(s["estimate"] for s in pl["sessions"]) / 60)} min of sims, plus the binaries\' build in each session')
        if not a.dry:
            d = os.path.join(D, 'parallel', a.run)
            os.makedirs(d, exist_ok=True)
            json.dump(pl, open(os.path.join(d, 'plan.json'), 'w'), indent=1)
            push(a.run, 'plan', {'plan.json': os.path.join(d, 'plan.json')})
    elif a.cmd == 'worker':
        worker(a.run, a.session, a.commit, a.main, a.presets, a.carries_from)
    elif a.cmd == 'merge':
        merge(a.run)
    elif a.cmd == 'clean':
        clean(a.run)


if __name__ == '__main__':
    main(sys.argv[1:])
