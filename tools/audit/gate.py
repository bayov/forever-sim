#!/usr/bin/env python3
"""Decide which auditors need a pass, without a model: git diffs on each auditor's watched paths,
new Forever client builds, and periodic checks of outside sources.

usage: gate.py [--advance] [--force ID[,ID]] [--all] [--json]

Reads orchestration/audit/auditors.json (the registry) and orchestration/audit/state/<id>.json (what
each auditor last audited). Run it after `git fetch origin --prune`.

An auditor is due when one of these holds:
- a watched path changed on origin/master since the master commit it last audited;
- a watched branch (a glob in "branches") is ahead of master, has a commit from the last
  "branch_active_days" days (default 2), touches a watched path, and its head moved since the last pass;
- "wago_build" is in its "external" list and wago.tools has a newer 1.60 build than it last audited;
- "external_every_h" hours passed since its last pass (for sources git can't see, like dev notes);
- "max_quiet_days" days passed since its last pass (a safety net).

--advance moves each auditor that isn't due up to the current heads, so its next diff starts there.
It changes state files only; commit them with commit_state.sh. Prints a short table, or JSON with --json.
"""
import fnmatch, json, os, re, shutil, subprocess, sys
from datetime import datetime, timezone

# The cloud runs use git. A local checkout that jj manages asks jj instead, because agents
# there aren't allowed to call git (a wrapper refuses it), so the gate couldn't run when
# someone wants to check which auditors a change would wake. Run `jj git fetch` first then.
JJ = subprocess.run(['jj', 'root'], capture_output=True, text=True) if shutil.which('jj') else None
if JJ and JJ.returncode == 0:
    REPO = JJ.stdout.strip()
else:
    JJ = None
    REPO = subprocess.check_output(['git', 'rev-parse', '--show-toplevel'], text=True).strip()
AUD = os.path.join(REPO, 'orchestration/audit')
NOW = datetime.now(timezone.utc)


def git(*a):
    return subprocess.run(['git', '-C', REPO, *a], capture_output=True, text=True).stdout.strip()


def jj(*a):
    return subprocess.run(['jj', '-R', REPO, '--ignore-working-copy', *a], capture_output=True, text=True).stdout.strip()


def jj_one(revset, template):
    return jj('log', '--no-graph', '-r', revset, '-T', template)


def origin(branch):
    """A branch on origin as a jj revset. Names like forever-update/2026-10-16 need quotes."""
    return f'"{branch}"@origin'


def glob_re(pattern):
    """A git :(glob) pathspec as a regex: ** crosses folders, * and ? don't."""
    out, i = '', 0
    while i < len(pattern):
        if pattern.startswith('**/', i):
            out, i = out + '(?:.*/)?', i + 3
        elif pattern.startswith('**', i):
            out, i = out + '.*', i + 2
        elif pattern[i] == '*':
            out, i = out + '[^/]*', i + 1
        elif pattern[i] == '?':
            out, i = out + '[^/]', i + 1
        else:
            out, i = out + re.escape(pattern[i]), i + 1
    return re.compile(out + r'\Z')


def changed(old, new, paths, exclude, three_dot=False):
    if not paths:
        return []
    if not old:
        return ['(no previous commit recorded)']
    if old == new:
        return []
    if JJ:
        base = jj_one(f'fork_point({old} | {new})', 'commit_id') if three_dot else old
        names = jj('diff', '--name-only', '--from', base, '--to', new).splitlines()
        want, skip = [glob_re(p) for p in paths], [glob_re(p) for p in exclude]
        return [n for n in names if any(w.match(n) for w in want) and not any(x.match(n) for x in skip)]
    spec = [f':(glob){p}' for p in paths] + [f':(glob,exclude){p}' for p in exclude]
    rng = f'{old}...{new}' if three_dot else f'{old}..{new}'
    return [l for l in git('diff', '--name-only', rng, '--', *spec).splitlines() if l]


def hours_since(ts):
    if not ts:
        return 1e9
    return (NOW - datetime.fromisoformat(ts.replace('Z', '+00:00'))).total_seconds() / 3600


args = sys.argv[1:]
force = set()
if '--force' in args:
    force = set(args[args.index('--force') + 1].split(','))
reg = json.load(open(os.path.join(AUD, 'auditors.json')))
if JJ:
    master = jj_one('master@origin', 'commit_id')
    remote_branches = jj_one('remote_bookmarks(remote=exact:"origin")',
                             'remote_bookmarks.map(|b| if(b.remote() == "origin", b.name())).join(" ") ++ " "').split()
else:
    master = git('rev-parse', 'origin/master')
    remote_branches = [b.strip().removeprefix('origin/') for b in git('branch', '-r', '--format=%(refname:short)').splitlines()]
client = None
if any('wago_build' in a.get('external', []) for a in reg['auditors']):
    probe = ['--probe', str(reg['wago_probe'])] if reg.get('wago_probe') else []
    r = subprocess.run([sys.executable, os.path.join(REPO, 'tools/audit/wago_latest.py'), *probe], capture_output=True, text=True)
    client = r.stdout.strip() if r.returncode == 0 else None

out = []
for a in reg['auditors']:
    if not a.get('enabled', True):
        continue
    sp = os.path.join(AUD, 'state', a['id'] + '.json')
    st = json.load(open(sp)) if os.path.exists(sp) else {}
    last = st.get('last', {})
    reasons, files = [], []
    f = changed(last.get('master'), master, a['watch'], a.get('exclude', []))
    if f:
        reasons.append(f'master: {len(f)} watched file(s) changed')
        files += f
    heads = {'master': master}
    for b in remote_branches:
        if b in ('master', 'HEAD') or not any(fnmatch.fnmatch(b, g) for g in a.get('branches', [])):
            continue
        if JJ:
            head = jj_one(origin(b), 'commit_id')
            if not jj_one(f'master@origin..{origin(b)}', '"x"'):
                continue
            age_h = (NOW.timestamp() - int(jj_one(head, 'committer.timestamp().format("%s")') or 0)) / 3600
        else:
            head = git('rev-parse', f'origin/{b}')
            if git('rev-list', '--count', f'origin/master..origin/{b}') in ('', '0'):
                continue
            age_h = (NOW.timestamp() - int(git('log', '-1', '--format=%ct', head) or 0)) / 3600
        if age_h > 24 * reg.get('branch_active_days', 2):  # old unmerged branches aren't live work
            continue
        heads[b] = head
        if last.get(b) == head:
            continue
        bf = changed(master, head, a['watch'], a.get('exclude', []), three_dot=True)
        if bf:
            reasons.append(f'{b}: {len(bf)} watched file(s) ahead of master')
            files += [f'{b}:{x}' for x in bf]
    if client and 'wago_build' in a.get('external', []) and client != st.get('client_build'):
        reasons.append(f'client build {st.get("client_build")} -> {client}')
    h = hours_since(st.get('last_pass_at'))
    if a.get('external_every_h') and h >= a['external_every_h']:
        reasons.append(f'periodic outside-source check ({h:.0f} h since last pass)')
    if h >= 24 * a.get('max_quiet_days', 7):
        reasons.append(f'no pass for {h / 24:.1f} days')
    if a['id'] in force or '--all' in args:
        reasons.append('forced')
    due = bool(reasons)
    if not due and '--advance' in args:
        st.setdefault('last', {})
        st['last'].update(heads)
        if client:
            st['client_build'] = st.get('client_build') or client
        st['checked_at'] = NOW.strftime('%Y-%m-%dT%H:%M:%SZ')
        os.makedirs(os.path.dirname(sp), exist_ok=True)
        json.dump(st, open(sp, 'w'), indent=1)
        open(sp, 'a').write('\n')
    out.append({'id': a['id'], 'due': due, 'reasons': reasons, 'files': files[:40], 'model': a.get('model', 'sonnet'),
                'sims': a.get('sims', False), 'last_master': last.get('master'), 'heads': heads})

if '--json' in args:
    print(json.dumps({'master': master, 'client_build': client, 'auditors': out}, indent=1))
else:
    print(f'master {master[:10]}  client {client or "unknown"}')
    for o in out:
        print(f"{'DUE ' if o['due'] else 'quiet'} {o['id']:12s} {'; '.join(o['reasons']) or '-'}")
