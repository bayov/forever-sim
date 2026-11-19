#!/usr/bin/env python3
"""parallel.py: a full DPS Report regeneration over several cloud sessions at once (the regenerate-dps-report skill's
"Parallel full regeneration"). The coordinating session plans, spawns one worker session per part, waits for their
hand-backs, merges, then publishes and deploys as for one run. Workers never publish or deploy.

  parallel.py plan --run <run> --commit <sha> [--dps-shards N] [--ids <id,...>] [--heal-groups N]
                   [--heal-presets <stem,...>] [--modes dps,tank,heal] [--log <file>...] [--dry]
        the worker sessions: dps-0.. (the DPS builds split by cost, pipeline.sh run on each share; N by default the
        fewest shards that none outlasts the longest other session), tank (the Tanking mode whole), heal-0 (the
        Healing main bench, handed to the others, then its own presets) and heal-1.. (each a group of Healing presets:
        every priest and raid preset alone, the rest balanced by cost). Prints each session's command and estimated
        time, and writes plan.json to $DASH_DIR/parallel/<run>/ and to the slot dash-raw/slot/plan (--dry: print only)
  parallel.py worker <run> <session> --commit <sha> [--shard K/N] [--ids <id,...>] [--main] [--presets <stem,...>]
                     [--carries-from <session>]
        one session's whole job, in the background with its log in a file: checks HEAD is <sha>, builds the binaries
        (pipeline.sh bins), sims its part, packs it and pushes it to its slot dash-raw/slot/<session>. A Healing
        session without --main first waits (polling the remote, no turns spent) for this run's heal-main.tgz on the
        --carries-from slot: the main bench outputs its race sims start from (tankheal.py run heal --carries)
  parallel.py merge <run>
        fetch the plan and every session's slot into $DASH_DIR/parallel/<run>/, check that each slot holds this run at
        the plan's commit with every file planned, and that every part was simmed at the same commit, sim inputs,
        binaries (STAMP.json) and DASH_SPLITS as each other and as this checkout's binaries would be, then: every DPS
        pack into $DASH_DIR (pipeline.sh unpack, into emptied outputs for a full run, as a fresh run would leave them),
        and tankheal.py merge for Tanking and Healing. Prints the publish command
  parallel.py shard <k> <n> <ids...>
        the ids of DPS shard k of n (pipeline.sh shard): each build with its variants, balanced by cost
  parallel.py slots
        each slot's tip: the run, session and commit it holds

Transport: a fixed set of slot branches, dash-raw/slot/<session> (plan, dps-0.., tank, heal-0..), reused by every run,
because the cloud sessions can't delete a branch (the git proxy and the GitHub API both refuse). Each push is one new
commit made with git plumbing (the checkout and master are never touched) whose parent is the slot's remote tip, so it
is always a fast-forward, never a force-push. Its tree holds only this run's files and slot.json, which names the run,
the session and the commit simmed; the commit message names the run too. A second push of the same run (heal-0: its
main bench, then its presets) keeps the first one's files. Whoever reads a slot (merge, a Healing worker waiting for
heal-main.tgz) refuses a tip that names another run or another commit: one run at a time uses the slots. A part is
small (the 2026-11-18 Healing pack is 13 KB, a DPS shard's pack a few MB).

Why the outputs are the same bytes as one run's: every sim's seed, iterations and DASH_SPLITS are fixed, a build's sims
read only its own request (a DPS shard keeps each build with its variants), a preset's race sims read only its own
request and the main bench's carries, pipeline.sh unpack leaves $DASH_DIR as a run of the same builds would, and
tankheal.py merge puts races.jsonl in the order one run writes it (race_lines_sorted). The main bench runs once
(heal-0) and every other Healing part reads its outputs.
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
# A preset's cost is the seconds of its race sims in a full run, 2 presets at a time on 4 cores. HEAL_S holds the
# 2026-11-19 run at d82ab73 (after the race sims skip what their pilots show changes nothing): the log lines
# '<mode> races <stem>: <n> rows, <s> s' of heal-6 for the tank healers, and for the presets that ran alone their
# seconds over ALONE_X. A preset not in it costs its race rows (one a race and racial variant) times a role's seconds a
# row, the same run's rates (raid: Druid 1203s for 13 rows, Shaman 842s for 10; tank: Druid 457s for 13, Shaman 293s
# for 10, Paladin 272s for 8), and Tanking the 2026-11-18 bench probe (warrior_protection_balanced 80s for 29) times
# FULL_X. --log takes measured times instead, from a run's log.
HEAL_S = {'druid_restoration': 457, 'shaman_restoration': 293, 'paladin_holy': 272,
          'priest_holy_raid': 913 / 0.6, 'druid_restoration_raid': 722 / 0.6, 'shaman_restoration_raid': 505 / 0.6,
          'priest_discipline': 486 / 0.6, 'priest_holy': 427 / 0.6, 'paladin_holy_raid': 328 / 0.6}
ROW_S = {('heal', 'raid'): 90, ('heal', 'tank'): 33, ('tank', None): 80 / 29 * 3.7}
FULL_X = 1
ALONE_X = 0.6  # a preset alone has all 4 cores: the priests, alone on 2026-11-18, took about 0.6 of the model
MAIN_S = {'heal': 231, 'tank': 20}  # the main bench, the 2026-11-18 Healing run at 87ce834 and 2026-11-17 Tanking run.json
MODE_S = {'dps': 80 * 60, 'tank': 940}  # DPS: pipeline.sh plan's 310 core-minutes on 4 cores; Tanking: the
# 2026-11-19 run at d82ab73 (its session took 17 min with the build)
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
    table = HEAL_S if mode == 'heal' else {}
    return {s: round(got.get(s) or table.get(s) or preset_rows(mode, s) * ROW_S[(mode, role(s))] * FULL_X) for s in stems}


def wall(cs, slots=SLOTS):
    """The wall time of presets of these costs run `slots` at a time, longest first (race_sims' pool). One preset
    alone has the whole machine (ALONE_X)."""
    if len(cs) == 1:
        return cs[0] * ALONE_X
    load = [0.0] * slots
    for c in sorted(cs, reverse=True):
        load[load.index(min(load))] += c
    return max(load) if cs else 0.0


def split(cost, groups=None, alone=lambda s: s.startswith('priest_') or s.endswith('_raid')):
    """The presets in groups, one a worker session: each `alone` preset (the priests, with the most race rows, and the
    raid presets, whose rows cost about 3 times a tank healer's) in a group of its own, the rest spread over the other
    groups longest first, each to the group that would end soonest. groups: how many in all (default: as few as keep
    every group within the longest lone preset's time, at least one)."""
    lone = sorted((s for s in cost if alone(s)), key=lambda s: (-cost[s], s))
    rest = sorted((s for s in cost if not alone(s)), key=lambda s: (-cost[s], s))
    if groups is not None and groups <= len(lone):  # too few sessions for the lone presets: one balanced spread
        rest, lone = sorted(cost, key=lambda s: (-cost[s], s)), []
        n = max(1, groups)
    elif groups is not None:
        n = groups - len(lone)
    else:
        cap = max([wall([cost[s]]) for s in lone] or [cost[s] for s in rest])  # no group longer than the longest lone preset
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


# ---------------------------------------------------------------- DPS shards
#
# A build's cost is pipeline.sh's table: core-minutes per base build of every step on the 2026-11-04 full run, a duty
# preset or pet variant 0.35 of its class's. MODE_S['dps'] is the whole set on one 4-core session, so a shard's time is
# its share of the set's cost. A shard keeps each build with its variants (the presets whose baseBuild it is): run
# re-sims those with their base anyway (rawstore.py variants), so splitting them would sim them twice.
DPS_W = {'hunter': 12.1, 'warrior': 9.6, 'rogue': 6.5, 'warlock': 5.4, 'shaman': 4.7, 'mage': 4.4, 'druid': 4.3, 'paladin': 4.2,
         'priest': 2.5}
DPS_ALONE_S = 40 * 60  # a DPS-only plan's shard target, with no other session to keep within


def dps_meta(bid):
    return json.load(open(os.path.join(D, 'req', bid + '.meta.json')))


def dps_cost(bid):
    return DPS_W.get(bid.split('__')[0], 4) * (0.35 if any(dps_meta(bid).get(k) for k in ('duty', 'pet')) else 1)


def dps_shards(ids, n):
    """The ids (with their variants) in n shards: each family (a build and its variants) to the shard that is least
    loaded, longest family first. Each shard's ids most expensive first, as pipeline.sh runs them."""
    every = req_ids()
    ids = list(dict.fromkeys(list(ids) + [b for b in every if dps_meta(b).get('baseBuild') in ids]))
    fam = {}
    for b in ids:
        base = dps_meta(b).get('baseBuild')
        fam.setdefault(base if base in ids else b, []).append(b)
    fcost = {k: sum(dps_cost(b) for b in v) for k, v in fam.items()}
    load, out = [0.0] * n, [[] for _ in range(n)]
    for k in sorted(fam, key=lambda k: (-fcost[k], k)):
        j = load.index(min(load))
        load[j] += fcost[k]
        out[j] += fam[k]
    return [sorted(g, key=lambda b: (-dps_cost(b), b)) for g in out]


def dps_seconds(ids):
    return MODE_S['dps'] * sum(dps_cost(b) for b in ids) / sum(dps_cost(b) for b in req_ids())


def dps_split(ids, cap, shards=None):
    """(n, shards): n given, or the fewest shards none of which outlasts cap seconds (at most one a family)."""
    if shards:
        return shards, dps_shards(ids, shards)
    n, out = 1, dps_shards(ids, 1)
    while max(dps_seconds(g) for g in out) > cap:
        more = dps_shards(ids, n + 1)
        if any(not g for g in more):
            break
        n, out = n + 1, more
    return n, out


def plan(run, commit, heal_groups=None, heal_presets=None, logs=(), modes=('dps', 'tank', 'heal'), dps_ids=None, dps_n=None):
    """The worker sessions of a run: [{'session', 'mode', 'presets' or 'ids', 'main', 'carriesFrom', 'estimate', 'command'}]."""
    import tankheal
    sessions = []
    cmd = lambda name, *more: ' '.join(['python3 tools/dps_report/parallel.py worker', run, name, '--commit', commit, *more])  # noqa: E731
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
    if 'dps' in modes:
        ids = list(dps_ids) if dps_ids else req_ids()
        cap = max([s['estimate'] for s in sessions] or [DPS_ALONE_S])  # no DPS shard longer than the longest other session
        n, shards = dps_split(ids, cap, dps_n)
        more = ['--ids', ','.join(ids)] if dps_ids else []
        sessions[:0] = [{'session': f'dps-{k}', 'mode': 'dps', 'shard': [k, n], 'ids': g, 'estimate': round(dps_seconds(g)),
                         'command': cmd(f'dps-{k}', '--shard', f'{k}/{n}', *more)} for k, g in enumerate(shards)]
    return {'run': run, 'commit': commit, 'modes': list(modes), 'healAll': heal_presets is None, 'sessions': sessions}


def req_ids():
    return sorted(os.path.basename(p)[:-10] for p in glob.glob(os.path.join(D, 'req', '*.meta.json')))


# ---------------------------------------------------------------- git transport: the slots

SLOT = 'slot'  # dash-raw/slot/<session>
MARK = 'slot.json'  # in every slot commit: {'run', 'session', 'commit', 'files'}


def git(*args, data=None, env=None, check=True):
    r = subprocess.run(['git', *args], cwd=ROOT, input=data, capture_output=True, env=env)
    if check and r.returncode:
        sys.exit(f'git {" ".join(args)} failed: {r.stderr.decode(errors="replace")[-1500:]}')
    return r.stdout


def ref_of(session):
    return f'refs/heads/{PREFIX}/{SLOT}/{session}'


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


def same_commit(a, b):
    """Two spellings of one commit (a sha and its abbreviation)."""
    return bool(a and b) and (a.startswith(b) or b.startswith(a))


def slot_info(sha):
    """A slot commit's slot.json, or {} (none: a slot from before slot.json, or not a slot commit)."""
    blob = tree_files(sha).get(MARK) if sha else None
    return json.loads(git('cat-file', 'blob', blob)) if blob else {}


def slot_of(run, session, commit):
    """The slot's tip and files when it holds this run at this commit: (sha, {name: blob}, None), else
    (sha, None, why not)."""
    sha = fetch(ref_of(session))
    if not sha:
        return None, None, 'its slot is empty'
    info = slot_info(sha)
    if info.get('run') != run:
        return sha, None, f'its slot holds run {info.get("run")}, not {run}'
    if not same_commit(info.get('commit'), commit):
        return sha, None, f'its slot holds {run} simmed at {info.get("commit")}, not {commit}'
    return sha, {n: b for n, b in tree_files(sha).items() if n != MARK}, None


def push(run, session, files, commit):
    """files {name: local path} onto the slot dash-raw/slot/<session>: one new commit, made with git plumbing (the
    checkout and its index are never touched), whose parent is the slot's remote tip, so the push is a fast-forward.
    Its tree: slot.json naming the run, session and commit, the files an earlier push of this same run and commit left
    on the slot, and these. Another run's files are never carried over. Retried on a fresh tip when the push fails."""
    ref = ref_of(session)
    env = dict(os.environ)
    if not git('config', 'user.email', check=False).strip():
        env.update(GIT_AUTHOR_NAME='dps report worker', GIT_AUTHOR_EMAIL='noreply@anthropic.com',
                   GIT_COMMITTER_NAME='dps report worker', GIT_COMMITTER_EMAIL='noreply@anthropic.com')
    blobs = {name: git('hash-object', '-w', path).decode().strip() for name, path in files.items()}
    for i in range(5):
        parent, ents, _ = slot_of(run, session, commit)
        ents = dict(ents or {}, **blobs)
        info = {'run': run, 'session': session, 'commit': commit, 'files': sorted(ents)}
        ents[MARK] = git('hash-object', '-w', '--stdin', data=(json.dumps(info, indent=1) + '\n').encode()).decode().strip()
        tree = git('mktree', data=''.join(f'100644 blob {b}\t{n}\n' for n, b in sorted(ents.items())).encode()).decode().strip()
        msg = f'dps report {run}: {session} at {commit[:10]} ({", ".join(sorted(files))})'
        new = git('commit-tree', tree, *(['-p', parent] if parent else []), '-m', msg, env=env).decode().strip()
        r = subprocess.run(['git', 'push', '-q', REMOTE, f'{new}:{ref}'], cwd=ROOT, capture_output=True)
        if r.returncode == 0:
            print(f'pushed {", ".join(sorted(files))} of {run} to {ref[len("refs/heads/"):]}', flush=True)
            return new
        time.sleep(5 * 2 ** i)
    sys.exit(f'git push to {ref} failed five times: {r.stderr.decode(errors="replace")[-1000:]}')


def wait_for(run, session, name, dest, commit, timeout=4 * 3600, every=60):
    """Wait until the slot dash-raw/slot/<session> holds `name` of this run and commit, then write it to dest. A tip of
    another run (the slot's last run, or this run's session not started yet) is waited past; of another commit, refused."""
    t0, ref = time.time(), ref_of(session)
    while True:
        sha, files, why = slot_of(run, session, commit)
        if why and 'simmed at' in why:
            sys.exit(f'{ref[len("refs/heads/"):]}: {why}: was that session started at another commit?')
        if files and name in files:
            open(dest, 'wb').write(git('cat-file', 'blob', files[name]))
            print(f'{name} of {run} from {ref[len("refs/heads/"):]} after {round(time.time() - t0)} s', flush=True)
            return dest
        if time.time() - t0 > timeout:
            sys.exit(f'{name} of {run} never reached {ref} ({timeout} s; {why or "not pushed yet"}): is that session stuck?')
        time.sleep(every)


def slots():
    for ref, sha in sorted(remote_tips(f'refs/heads/{PREFIX}/{SLOT}/*').items()):
        fetch(ref)
        info = slot_info(sha)
        print(f'{ref[len("refs/heads/"):]:28} {sha[:10]}  run {info.get("run")}  at {(info.get("commit") or "")[:10]}  '
              f'{" ".join(info.get("files") or [])}')


# ---------------------------------------------------------------- the worker

def sh(*cmd, env=None):
    print('$ ' + ' '.join(cmd), flush=True)
    t = time.time()
    r = subprocess.run(list(cmd), cwd=ROOT, env=env)
    if r.returncode:
        sys.exit(f'{" ".join(cmd[:3])} failed ({r.returncode})')
    print(f'  ({round(time.time() - t)} s)', flush=True)


def dps_manifest(ids, shard, seconds):
    import build_bins
    return {'mode': 'dps', 'commit': head(), 'splits': int(os.environ.get('DASH_SPLITS', '4')), 'ids': ids, 'shard': shard,
            'seconds': seconds, 'stamp': build_bins.summary(os.environ.get('DASH_BIN') or dashpaths.BIN, ['wowsimcli', 'dashstats'])}


def head():
    return git('rev-parse', 'HEAD').decode().strip()


def worker(run, session, commit, main=False, presets=None, carries_from=None, shard=None, ids=None):
    t0 = time.time()
    if not head().startswith(commit):
        sys.exit(f'HEAD is {head()[:10]}, not the run\'s commit {commit}: git fetch origin && git checkout --detach {commit}')
    if os.environ.get('DASH_ALLOW_BIN_MISMATCH'):
        sys.exit('a worker never sims with DASH_ALLOW_BIN_MISMATCH: its part would be refused at the merge')
    os.environ.setdefault('DASH_SPLITS', '4')
    # the tree is this checkout, at the run's commit (checked above): a fresh cloud session has no ../w-dash
    os.environ.setdefault('DASH_TREE', os.path.abspath(os.path.join(HERE, '..', '..')))
    out = os.path.join(D, 'parallel', run)
    os.makedirs(out, exist_ok=True)
    pl, th = os.path.join(HERE, 'pipeline.sh'), os.path.join(HERE, 'tankheal.py')
    sha = head()
    if session.startswith('dps-'):  # its share of the DPS builds, in this empty $DASH_DIR's outputs
        k, n = (int(x) for x in (shard or '').split('/')) if shard else (None, None)
        if k is None or session != f'dps-{k}':
            sys.exit(f'{session}: --shard {k}/<n> names its share of the DPS builds')
        mine = dps_shards(ids.split(',') if ids else req_ids(), n)[k]
    sh(pl, 'bins')
    env = dict(os.environ, TANKHEAL_BIN=os.environ.get('DASH_BIN') or dashpaths.BIN)
    if session.startswith('dps-'):
        sh(pl, 'run', *mine)
        tgz, man = os.path.join(out, session + '.tgz'), os.path.join(out, session + '.json')
        sh(pl, 'pack', tgz, *mine)
        json.dump(dps_manifest(mine, [k, n], round(time.time() - t0)), open(man, 'w'), indent=1)
        push(run, session, {session + '.tgz': tgz, session + '.json': man}, sha)
    elif session.startswith('tank'):
        sh(sys.executable, th, 'run', 'tank', '--part', *(['--presets', presets] if presets else []), env=env)
        sh(sys.executable, th, 'part', 'tank', os.path.join(out, session + '.tgz'))
        push(run, session, {session + '.tgz': os.path.join(out, session + '.tgz')}, sha)
    elif session.startswith('heal'):
        carries = os.path.join(out, 'heal-main.tgz')
        if main:  # the main bench first, pushed at once: the other Healing sessions wait for it
            sh(sys.executable, th, 'run', 'heal', '--main-only', env=env)
            sh(sys.executable, th, 'part', 'heal', carries)
            push(run, session, {'heal-main.tgz': carries}, sha)
        else:
            wait_for(run, carries_from, 'heal-main.tgz', carries, sha)
        if presets:
            sh(sys.executable, th, 'run', 'heal', '--presets', presets, '--carries', carries, env=env)
            sh(sys.executable, th, 'part', 'heal', os.path.join(out, session + '.tgz'))
            push(run, session, {session + '.tgz': os.path.join(out, session + '.tgz')}, sha)
    else:
        sys.exit(f'no such session: {session} (dps-<k>, tank, heal-<k>)')
    print(f'worker {session} of {run} done in {round((time.time() - t0) / 60)} min: slot {PREFIX}/{SLOT}/{session}', flush=True)


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


def wanted(s):
    """The files a planned session pushes."""
    if s['mode'] == 'dps':
        return [s['session'] + '.tgz', s['session'] + '.json']
    if s['mode'] == 'tank':
        return [s['session'] + '.tgz']
    return (['heal-main.tgz'] if s.get('main') else []) + ([s['session'] + '.tgz'] if s.get('presets') else [])


def gather(run, dest):
    """The plan and every planned session's files into dest: (plan, {name: path}, manifests). Refuses a slot whose tip
    holds another run or another commit than the plan's, and a file missing from a slot."""
    import tankheal
    os.makedirs(dest, exist_ok=True)
    sha = fetch(ref_of('plan'))
    info = slot_info(sha) if sha else {}
    if info.get('run') != run:
        sys.exit(f'the slot {PREFIX}/{SLOT}/plan holds run {info.get("run")}, not {run}: parallel.py plan --run {run} ... first')
    got = {}

    def take(name, blob):
        got[name] = os.path.join(dest, name)
        open(got[name], 'wb').write(git('cat-file', 'blob', blob))
    take('plan.json', tree_files(sha)['plan.json'])
    pl = json.load(open(got['plan.json']))
    missing, refused = [], []
    for s in pl['sessions']:
        _, files, why = slot_of(run, s['session'], pl['commit'])
        if why:
            (missing if 'simmed at' not in why else refused).append(f'{s["session"]}: {why}')
            continue
        for n in wanted(s):
            if n in files:
                take(n, files[n])
            else:
                missing.append(f'{s["session"]}: {n} not pushed yet')
    if refused:
        sys.exit('slots of this run simmed at another commit than the plan\'s:\n' + '\n'.join('  - ' + m for m in refused))
    if missing:
        sys.exit('parts not pushed yet (is a session still running or stuck?):\n' + '\n'.join('  - ' + m for m in missing))
    manifests = []
    for s in pl['sessions']:
        if s['mode'] == 'dps':
            m = json.load(open(got[s['session'] + '.json']))
            if m.get('ids') != s['ids']:
                sys.exit(f'{s["session"]} simmed {m.get("ids")}, the plan gave it {s["ids"]}')
            manifests.append((s['session'], m))
    for name in sorted(got):
        if name.endswith('.tgz') and not name.startswith('dps-'):
            info, _ = tankheal.read_part(got[name])
            if info is None:
                sys.exit(f'{name}: no part.json')
            manifests.append((name, info))
            s = next((x for x in pl['sessions'] if x['session'] + '.tgz' == name), None)
            if s and s['mode'] == 'heal' and sorted(info.get('presets') or []) != sorted(s['presets']):
                sys.exit(f'{name} raced {info.get("presets")}, the plan gave it {s["presets"]}')
    probs = stamp_problems(manifests) if manifests else []
    if manifests and not same_commit(manifests[0][1].get('commit'), pl['commit']):
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
    dps = [s for s in pl['sessions'] if s['mode'] == 'dps']
    ids = [b for s in dps for b in s['ids']]
    if dps and sorted(ids) == req_ids():
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
    for s in dps:  # in shard order: unpack leaves the outputs as one run of the shards' builds in this order would
        sh(os.path.join(HERE, 'pipeline.sh'), 'unpack', got[s['session'] + '.tgz'])
    th = os.path.join(HERE, 'tankheal.py')
    for mode in ('tank', 'heal'):
        ss = [s for s in pl['sessions'] if s['mode'] == mode]
        if not ss:
            continue
        parts = [got[n] for n in sorted(got) if n.endswith('.tgz') and n.startswith(mode)]
        keep = mode == 'heal' and not pl.get('healAll', True)
        sh(sys.executable, th, 'merge', mode, *(['--keep-rest'] if keep else []), *parts)
    print(f'\nnext: SIM_COMMIT={pl["commit"]} tools/dps_report/pipeline.sh publish {" ".join(ids)}\n'
          f'then check, commit, push and deploy as for one run. The slots stay: the next run reuses them')


def main(argv):
    ap = argparse.ArgumentParser(description='A DPS Report regeneration over several cloud sessions (module docstring).')
    sub = ap.add_subparsers(dest='cmd', required=True)
    p = sub.add_parser('plan')
    p.add_argument('--run', required=True)
    p.add_argument('--commit', required=True)
    p.add_argument('--dps-shards', type=int)
    p.add_argument('--ids')
    p.add_argument('--heal-groups', type=int)
    p.add_argument('--heal-presets')
    p.add_argument('--modes', default='dps,tank,heal')
    p.add_argument('--log', action='append', default=[])
    p.add_argument('--dry', action='store_true')
    w = sub.add_parser('worker')
    w.add_argument('run')
    w.add_argument('session')
    w.add_argument('--commit', required=True)
    w.add_argument('--shard')
    w.add_argument('--ids')
    w.add_argument('--main', action='store_true')
    w.add_argument('--presets')
    w.add_argument('--carries-from', default='heal-0')
    sub.add_parser('merge').add_argument('run')
    s = sub.add_parser('shard')
    s.add_argument('k', type=int)
    s.add_argument('n', type=int)
    s.add_argument('ids', nargs='*')
    sub.add_parser('slots')
    a = ap.parse_args(argv)
    if a.cmd == 'plan':
        if '/' in a.run or a.run == SLOT:
            sys.exit(f'a run id is one name, e.g. <date>-<sha7>: {a.run}')
        r = subprocess.run(['git', 'rev-parse', '--verify', '-q', a.commit + '^{commit}'], cwd=ROOT, capture_output=True, text=True)
        commit = r.stdout.strip() if r.returncode == 0 else a.commit
        pl = plan(a.run, commit, a.heal_groups, set(a.heal_presets.split(',')) if a.heal_presets else None, a.log, a.modes.split(','),
                  a.ids.split(',') if a.ids else None, a.dps_shards)
        for s in pl['sessions']:
            print(f'{s["session"]:8} about {round(s["estimate"] / 60):3} min  {" ".join(s.get("presets") or []) if s["mode"] != "dps" else str(len(s["ids"])) + " builds"}')
            print(f'         {s["command"]}')
        print(f'wall: about {round(max(s["estimate"] for s in pl["sessions"]) / 60)} min of sims, plus the binaries\' build in each session')
        if not a.dry:
            d = os.path.join(D, 'parallel', a.run)
            os.makedirs(d, exist_ok=True)
            json.dump(pl, open(os.path.join(d, 'plan.json'), 'w'), indent=1)
            push(a.run, 'plan', {'plan.json': os.path.join(d, 'plan.json')}, commit)
    elif a.cmd == 'worker':
        worker(a.run, a.session, a.commit, a.main, a.presets, a.carries_from, a.shard, a.ids)
    elif a.cmd == 'merge':
        merge(a.run)
    elif a.cmd == 'shard':
        print(' '.join(dps_shards(a.ids or req_ids(), a.n)[a.k]))
    elif a.cmd == 'slots':
        slots()


if __name__ == '__main__':
    main(sys.argv[1:])
