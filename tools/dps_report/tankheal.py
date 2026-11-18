#!/usr/bin/env python3
"""tankheal.py: the report's Tanking and Healing modes, from tools/tankbench and tools/healbench.

  tankheal.py run [tank|heal ...] [--stale | --presets <stem,...>] [--no-races] [--resume]
                                                 the sims, into $DASH_DIR/tank and $DASH_DIR/heal (the benches' text,
                                                 log lines dropped, and run.json); marks the mode fresh for publish.
                                                 --resume keeps the bench calls a stopped run finished (.progress.json).
                                                 --presets re-sims the race rows of those presets only, --stale of the
                                                 presets stale.mjs calls stale (a mode with none is skipped): the main
                                                 bench still runs whole, and every other preset keeps its race rows
                                                 from raw/bench/<mode>.tgz when its main-run rows there are the same
                                                 bytes as this run's (else it is re-simmed too). Only the presets
                                                 re-simmed are marked fresh (tankheal.py fresh prints <mode>:<stem,...>)
  tankheal.py run <tank|heal> --part [--presets <stem,...>] | --main-only | --presets <stem,...> --carries <part>
                                                 one share of a mode for a worker session (parallel.py): --part the main
                                                 bench and the race sims of the presets named (default: every preset),
                                                 --main-only the main bench alone, --carries the race sims of the presets
                                                 named on another part's main bench outputs (the carries the healing race
                                                 sims start from), checked to be simmed alike. Writes <mode>/part.json, not
                                                 run.json or .fresh; --resume works as for a run
  tankheal.py part <tank|heal> <dest.tgz>        a finished part, packed deterministically (the worker's hand-back)
  tankheal.py merge <tank|heal> [--keep-rest] <part>...
                                                 parts (.tgz or folders) into $DASH_DIR/<mode>, as one run would have
                                                 left it (main bench files, races.jsonl and probes the same bytes), with
                                                 run.json and .fresh; refuses parts of other commits, sim inputs, binary
                                                 stamps or DASH_SPLITS, main outputs that differ, presets raced twice
                                                 differently, and presets no part raced (--keep-rest: those keep the
                                                 pack's rows when their main rows are its bytes, as run --presets does)
  tankheal.py bins                               every report binary into the binaries directory (build_bins.py: the
                                                 benches read every report switch, the racials switches among them)
  tankheal.py build                              the tank_* and heal_* datasets into $DASH_DIR/data (unpacks the raw
                                                 packs first when $DASH_DIR/tank or heal is missing)
  tankheal.py pack                               $DASH_DIR/tank and heal into raw/bench/<mode>.tgz (deterministic)
  tankheal.py unpack                             the packs back into $DASH_DIR/tank and heal
  tankheal.py fresh [--clear]                    the modes run since the last publish (stale.mjs --record-bench them)
  tankheal.py probe [tank|heal ...] [--presets <stem,...>] [--stale-ok]
                                                 each preset's probe into $DASH_DIR/<mode>/probe/<stem>.json, packed with
                                                 the mode: its main bench at ITERS / 10 and its race sims with every pilot
                                                 and ceiling of adaptive.py divided by 10, fingerprinted (probe_cli's
                                                 canon and digest) under a _key of the settings, the real bench binary,
                                                 the requests, boss files and scripts. run makes them after its races
                                                 (only with TANKHEAL_PROBE=1); alone, it refuses stale presets the last run
                                                 here didn't sim (--stale-ok for a throwaway measurement)
  tankheal.py probe check [tank|heal ...] [--presets <stem,...>] [--json]
                                                 probe the presets (default: the stale ones) into probe_check/bench/<mode>/,
                                                 compare with the pack's probes, print the moved and cleared presets and
                                                 one cleared preset to re-sim anyway (the audit, seeded by HEAD), and write
                                                 audit.json there for publish
  tankheal.py probe audit | passed               publish's steps: exit 1 unless the run since kept the cleared presets'
                                                 main rows and the audit preset's race rows as the pack had them; then the
                                                 cleared presets no run re-simmed, as <mode>:<stem,...> (--record-bench --by probe)

Tanking: every preset request in tools/tankbench/req (warrior_protection.json is no preset) on the seven boss files
of tools/encounters, `tankbench -iters 1000 -encounter <boss>`, the requests' own seed (simOptions.randomSeed).
The score is docs/tank_sim.md's objective (2 x SafeDPS - w x 900 x healing needed / healer HPS, w = 1/3, 1, 3) and a
preset's score is the mean of the six Anniversary bosses; Ragnaros is shown, not scored.
Healing: every request in tools/healbench/req, `healbench -carry -role tank|raid -iters 1000 -encounter <boss>`
(files ending _raid are raid healers), at each file's own length, the bench's seed (101): the score is the damage
carried (docs/healing_sim.md "What a healer scores" and "The score in the bench"), its mean over the seven boss files
(docs/healing_sim.md, H3: "The shaman's raid mean over these 7 files").
Races (bench_races.py, into <mode>/races.jsonl): each preset at every race its class can be (run_build.CLASS_RACES),
with every racial left out (off) and each racial part left out (off:<part>), per-fight values from seed 1 with labeled
random streams so every side fights the same fights, sized by adaptive.py's BENCH_GOALS (paired pilots; tank up to
CEIL fights, heal up to HEAL_CEIL). The build moves every race by the preset race's gap to its main-run numbers, keeps
the paired errors between races (pse), and splits each race's gap over its base stats into its racial parts
(<mode>_race_breakdown, split_parts: Shapley for two parts, scaled leave-one-out for three or more).
Text: what players read (preset lines, notes, boss lines, Method, rotation rewording) is written in tankheal_text/
<mode>.md, not taken from the docs; the build stops when a string the page shows has a tool or file word (LEAK).

The binaries (tankbench, healbench) come from $TANKHEAL_BIN, else $DASH_BIN, else $DASH_DIR/bin, and run with the
merged tree ($DASH_TREE, else this checkout when it has no w-dash) as their cwd. build_bins.py builds them, and run
refuses binaries its stamp doesn't match to the code (DASH_ALLOW_BIN_MISMATCH=<reason> to sim anyway, recorded in
<mode>/bin_override.json, which the pack keeps and stale.mjs --record-bench refuses).
Every bench call splits its sims into DASH_SPLITS parts (default 4, as pipeline.sh sets for the DPS steps), so the
outputs are the same on every machine. run.json records the count, and unpack refuses a pack made with another one.
"""
import ast, glob, hashlib, io, json, math, os, platform, random, re, shutil, subprocess, sys, tarfile, threading, time

HERE = os.path.dirname(os.path.abspath(__file__))
sys.path.insert(0, HERE)
import dashpaths  # noqa: E402
import adaptive  # noqa: E402
import bench_races  # noqa: E402
import timing_cli  # noqa: E402

ROOT = os.path.abspath(os.path.join(HERE, '..', '..'))
D = dashpaths.D
TREE = dashpaths.TREE if os.path.isdir(dashpaths.TREE) else ROOT
os.environ['DASH_TREE'] = TREE  # for the modules below that read dashpaths
dashpaths.TREE = TREE
BIN = os.path.abspath(os.environ.get('TANKHEAL_BIN') or dashpaths.BIN)
os.environ.setdefault('DASH_SPLITS', '4')  # the benches' sims split as pipeline.sh's DPS steps do (bench_races.py too)
SPLITS = int(os.environ['DASH_SPLITS'])
RAW = os.path.join(HERE, 'raw', 'bench')  # not raw/*.tgz: rawstore.py prune drops packs with no DPS request
MODES = ('tank', 'heal')
BOSSES = ['magmadar', 'golemagg', 'onyxia', 'broodlord', 'vaelastrasz', 'chromaggus', 'ragnaros']
SCORED = BOSSES[:6]  # the six Anniversary bosses; Ragnaros is shown, not scored (docs/tank_sim.md)
ITERS = 1000
HEAL_SEED = 101  # tools/healbench/main.go: SimOptions RandomSeed 101 (paired across the carry's bisection)
TANK_REQ, HEAL_REQ = 'tools/tankbench/req', 'tools/healbench/req'
NOT_PRESETS = {'warrior_protection'}  # the healing tools' tank, left as it was (docs/tank_warrior.md, Presets)
WEIGHT = {'max_threat': 1 / 3, 'balanced': 1.0, 'max_survival': 3.0}
PRESET_NAME = {'max_threat': 'Max threat', 'balanced': 'Balanced', 'max_survival': 'Max survival'}
TANK_SPEC = {'warrior_protection': ('Protection', 'protection', 'Protection Warrior'), 'paladin_protection': ('Protection', 'protection', 'Protection Paladin'),
             'druid_bear': ('Feral (Bear)', 'bear', 'Bear Druid'), 'shaman_warden': ('Warden', 'warden', 'Warden Shaman')}
HEAL_SPEC = {'priest_holy': 'Holy', 'priest_discipline': 'Discipline', 'paladin_holy': 'Holy', 'druid_restoration': 'Restoration',
             'shaman_restoration': 'Restoration'}
ROLE_NAME = {'tank': 'Tank healer', 'raid': 'Raid healer'}
CLASS_ORDER = ['Warrior', 'Paladin', 'Druid', 'Shaman', 'Priest']
TEXT = os.path.join(HERE, 'tankheal_text')  # the page's written text, a file a mode (tank.md, heal.md)
LOG_LINE = re.compile(r'^\d{4}/\d\d/\d\d \d\d:\d\d:\d\d ')


def bid_of(stem):
    c, rest = stem.split('_', 1)
    return f'{c}__{rest}'


def presets(mode):
    d = os.path.join(TREE, TANK_REQ if mode == 'tank' else HEAL_REQ)
    return sorted(s for s in (os.path.basename(p)[:-5] for p in glob.glob(os.path.join(d, '*.json'))) if s not in NOT_PRESETS)


def req_path(mode, stem):
    return os.path.join(TREE, TANK_REQ if mode == 'tank' else HEAL_REQ, stem + '.json')


def heal_role(stem):
    return 'raid' if stem.endswith('_raid') else 'tank'


def player(req):
    return req['raid']['parties'][0]['players'][0]


def race_of(req):
    return player(req)['race'][4:]


def class_races():
    src = open(os.path.join(HERE, 'run_build.py')).read()
    node = next(n for n in ast.parse(src).body if isinstance(n, ast.Assign) and getattr(n.targets[0], 'id', '') == 'CLASS_RACES')
    return ast.literal_eval(node.value)


# ---------------------------------------------------------------- run

def bench(mode, args, files, out):
    exe = os.path.join(BIN, 'tankbench' if mode == 'tank' else 'healbench')
    t = time.time()
    step = os.environ.get('TANKHEAL_STEP') or mode  # the step timing_cli.py logs it under (a probe's: <mode>_probe)
    r = subprocess.run([exe] + args + files, cwd=TREE, capture_output=True, text=True, env=dict(os.environ, TIMING_STEP=step))
    if r.returncode:
        sys.exit(f'{exe} failed ({r.returncode}):\n{r.stderr[-2000:]}')
    text = ''.join(l + '\n' for l in (r.stdout + r.stderr).split('\n') if l.strip() and not LOG_LINE.match(l))
    os.makedirs(os.path.dirname(out), exist_ok=True)
    open(out, 'w').write(text)
    return time.time() - t


def bins():
    """Every report binary, tankbench and healbench included, into BIN (build_bins.py: a clean worktree at
    ${SIM_COMMIT:-HEAD}, and the stamp run checks)."""
    import build_bins
    t0 = time.time()
    build_bins.build(BIN)
    timing_cli.phase('bins', time.time() - t0)


def main_carries(out, stem):
    """A healing preset's damage carried on each boss in the main run (the race sims' starting scales)."""
    role = heal_role(stem)
    return {b: heal_rows(open(os.path.join(out, f'{role}_{b}.txt')).read())[stem]['carry'] for b in BOSSES}


def race_sims(mode, stems, out):
    """Every race of each preset and its racials left out (bench_races.py), into <out>/races.jsonl; a preset already
    there is kept (a resumed run)."""
    from concurrent.futures import ThreadPoolExecutor
    res = os.path.join(out, 'races.jsonl')
    done = {json.loads(l)['build'] for l in open(res)} if os.path.exists(res) else set()
    exe = os.path.join(BIN, 'tankbench' if mode == 'tank' else 'healbench')
    B = bench_races.Bench(mode, exe, TREE, lambda s: req_path(mode, s), lambda b: os.path.join(TREE, 'tools/encounters', b + '.json'))
    cr, lock = class_races(), threading.Lock()

    def one(s):
        req = json.load(open(req_path(mode, s)))
        cls = player(req)['class'][5:]
        t = time.time()
        if mode == 'tank':
            each = {b: (json.load(open(os.path.join(TREE, 'tools/encounters', b + '.json'))).get('healing') or {}).get('hpsEach') or 400 for b in SCORED}
            w = WEIGHT[next(k for k in WEIGHT if s.endswith(k))]
            rows = bench_races.preset(mode, B, s, cls, race_of(req), cr[cls], SCORED, w=w, each=each)
        else:
            start = {b: round(math.log(c) / math.log(bench_races.STEP)) for b, c in main_carries(out, s).items()}
            rows = bench_races.preset(mode, B, s, cls, race_of(req), cr[cls], BOSSES, role=heal_role(s), start=start)
        offs = [r for r in rows if r['variant'] != 'race']
        if offs and all(r['same'] for r in offs) and any(r['variant'] == 'off' for r in offs):
            sys.exit(f'{exe} ignores DASH_RACIALS_OFF: rebuild it from this tree, which has the report switches (pipeline.sh bins)')
        with lock:
            with open(res, 'a') as fh:
                for r in rows:
                    fh.write(json.dumps(r, sort_keys=True) + '\n')
        print(f'{mode} races {s}: {len(rows)} rows, {round(time.time() - t)} s', flush=True)
    with ThreadPoolExecutor(2) as ex:
        list(ex.map(one, [s for s in stems if s not in done]))


def race_lines_sorted(lines, stems):
    """races.jsonl's lines in preset order (stems), each preset's lines in the order they were written: race_sims
    appends them as presets finish, so this makes the pack the same bytes for the same results."""
    order = {s: i for i, s in enumerate(stems)}
    return sorted((l for l in lines if l.strip()), key=lambda l: order.get(json.loads(l)['build'], len(order)))


def main_rows(texts, mode, stem):
    """A preset's rows in the main run's outputs (texts: {file name: text}), one tuple of row lines a boss file, or
    None when a file is missing. The rows hold what the race sims start from (main_carries) and show whether the
    preset's sims moved."""
    names = BOSSES if mode == 'tank' else [f'{heal_role(stem)}_{b}' for b in BOSSES]
    out = []
    for n in names:
        t = texts.get(n + '.txt')
        if t is None:
            return None
        out.append(tuple(l for l in t.split('\n') if l.startswith('|') and l.strip().strip('|').split('|')[0].strip() == stem))
    return tuple(out)


def reusable(mode, stems, stale, old, new):
    """The presets whose race rows a run with a stale list keeps from the pack: {stem: its races.jsonl lines}. old is
    the pack's files ({name: text}, None for no pack), new this run's main outputs. A preset is kept when it is not
    stale, the pack has its race rows, and its main-run rows on every boss are the same bytes as the pack's."""
    if not old:
        return {}
    rows = {}
    for l in old.get('races.jsonl', '').splitlines():
        if l.strip():
            rows.setdefault(json.loads(l)['build'], []).append(l + '\n')
    keep = {}
    for s in stems:
        if s in stale or s not in rows:
            continue
        a = main_rows(old, mode, s)
        if a is not None and all(a) and a == main_rows(new, mode, s):
            keep[s] = rows[s]
    return keep


def pack_files(mode):
    """raw/bench/<mode>.tgz's files as {name: text}, None when there is no pack."""
    p = os.path.join(RAW, mode + '.tgz')
    if not os.path.exists(p):
        return None
    with tarfile.open(p) as tar:
        return {ti.name: tar.extractfile(ti).read().decode() for ti in tar.getmembers() if ti.isfile()}


def check_splits(mode, files):
    """Stop when a pack's run was split into another number of parts than DASH_SPLITS (rawstore.py does the same for
    the DPS packs): its outputs would not be this run's. A run.json without the count is from before the benches read
    DASH_SPLITS, when they split one part per CPU on the 4-CPU machine that made every published pack."""
    if not files or 'run.json' not in files:
        return
    import rawstore
    rawstore.check_splits(f'raw/bench/{mode}.tgz', json.loads(files['run.json']).get('splits', 4), SPLITS)


def bench_inputs(mode, commit):
    """A hash of the mode's sim inputs at a commit (stale.mjs --bench-fingerprint: the same fingerprint publish records),
    or None when it can't be had (no node, or a commit this checkout doesn't know)."""
    r = subprocess.run(['node', os.path.join(HERE, 'stale.mjs'), '--bench-fingerprint', mode, '--commit', commit], cwd=ROOT, capture_output=True, text=True)
    if r.returncode or not r.stdout.strip():
        return None
    return hashlib.sha256(r.stdout.encode()).hexdigest()[:20]


def can_resume(done, commit, inputs):
    """Whether a stopped run's .progress.json (done) still holds for a run at commit whose sim inputs hash to `inputs`: it
    was started at this commit, or at another one with the same sim inputs (master moved on, nothing the sims read did)."""
    if not done:
        return False
    if done.get('commit') == commit:
        return True
    return inputs is not None and done.get('inputs') == inputs


def stale_presets():
    """The presets stale.mjs calls stale, by mode."""
    r = subprocess.run(['node', os.path.join(HERE, 'stale.mjs'), '--json'], cwd=ROOT, capture_output=True, text=True, check=True)
    return {b['id']: set(b['presets']) for b in json.loads(r.stdout)['benches']}


def run_info(mode, commit, secs, seconds, races):
    """run.json (and a part's part.json): the settings and what the run took."""
    return {'iterations': ITERS, 'seed': 'the requests\' simOptions.randomSeed' if mode == 'tank' else HEAL_SEED,
            'raceGoals': {k: v for k, v in bench_races.GOALS[mode].items()}, 'racePilot': [adaptive.RACE_PILOT, adaptive.PILOT],
            'raceCeil': adaptive.CEIL if mode == 'tank' else adaptive.HEAL_CEIL,
            'splits': SPLITS, 'commit': commit, 'at': time.strftime('%Y-%m-%dT%H:%M:%SZ', time.gmtime()), 'seconds': seconds,
            'runSeconds': secs, 'races': races}


def run(modes, races=True, resume=False, only=None, part=None):
    """only: {mode: the stale presets} for a run that re-sims only those presets' race rows (None: every preset).
    part: one share of a mode, simmed in a worker session and combined by merge() (parallel.py): {'presets': the presets
    whose race sims it runs (None: every preset), 'main_only': the main bench alone, 'carries': another part (its .tgz or
    folder) whose main bench outputs it takes instead of running the main bench}. A part writes part.json (what it
    simmed, at which commit, inputs, binaries and DASH_SPLITS), not run.json or .fresh: alone it is not publishable."""
    import build_bins
    override = build_bins.guard(BIN, [build_bins.BENCH_BINS[m] for m in modes], f'tankheal.py run {" ".join(modes)}')
    commit = subprocess.run(['git', 'rev-parse', 'HEAD'], cwd=TREE, capture_output=True, text=True).stdout.strip()
    for mode in modes:
        out = os.path.join(D, mode)
        old = pack_files(mode) if only is not None else None
        prog = os.path.join(out, '.progress.json')
        done = json.load(open(prog)) if resume and os.path.exists(prog) else {}
        inputs = bench_inputs(mode, commit)
        if any(os.path.exists(os.path.join(out, f)) for f in ('run.json', PART)) or not can_resume(done, commit, inputs):
            done = {}  # a finished run, another commit's outputs with other sim inputs: start over
        if done and done['commit'] != commit:
            print(f'{mode}: resuming the run started at {done["commit"][:10]} (HEAD is {commit[:10]}, same sim inputs)', flush=True)
        commit_run = done.get('commit', commit)  # the commit the run started at: what its packs are stamped with
        if not done:
            shutil.rmtree(out, ignore_errors=True)
        os.makedirs(out, exist_ok=True)
        if override:  # packed with the outputs, and stale.mjs --record-bench refuses the mode
            build_bins.mark(os.path.join(out, 'bin_override.json'), override)
        stems = presets(mode)
        secs, t0 = dict(done.get('secs', {})), time.time() - done.get('seconds', 0)
        me = None
        if part is not None:
            me = {'commit': commit_run, 'inputs': inputs, 'splits': SPLITS, 'stamp': build_bins.summary(BIN, [build_bins.BENCH_BINS[mode]])}
        carried = {}
        if part and part.get('carries'):  # the main bench outputs of the part that ran it (the main worker's)
            carried = carries(mode, part['carries'], stems, me)
            for n, data in carried.items():
                open(os.path.join(out, n), 'wb').write(data)
            print(f'{mode}: the main bench outputs from {part["carries"]} ({len(carried)} files)', flush=True)

        def note(k, v):
            secs[k] = v
            json.dump({'commit': commit_run, 'inputs': inputs, 'secs': secs, 'seconds': round(time.time() - t0, 1)}, open(prog, 'w'))
        groups = {'tank': stems} if mode == 'tank' else {r: [s for s in stems if heal_role(s) == r] for r in ('tank', 'raid')}
        for boss in BOSSES:
            enc = ['-encounter', f'tools/encounters/{boss}.json', '-iters', str(ITERS)]
            for g, st in groups.items():
                files = [req_path(mode, s) for s in st]
                name = boss if mode == 'tank' else f'{g}_{boss}'
                args = enc if mode == 'tank' else ['-carry', '-role', g] + enc
                if name not in secs and name + '.txt' not in carried:
                    note(name, round(bench(mode, args, files, os.path.join(out, name + '.txt')), 1))
                    print(f'{mode} {name}: {secs[name]} s', flush=True)
        kept, raced = {}, []
        if races and not (part and part.get('main_only')):
            raced = stems if not part or part.get('presets') is None else [s for s in stems if s in part['presets']]
            t1 = time.time()
            res = os.path.join(out, 'races.jsonl')
            if only is not None:
                new = {n: open(os.path.join(out, n)).read() for n in os.listdir(out) if n.endswith('.txt')}
                kept = reusable(mode, stems, only.get(mode, set()), old, new)
                have = {json.loads(l)['build'] for l in open(res)} if os.path.exists(res) else set()
                with open(res, 'a') as fh:
                    for s in stems:
                        if s in kept and s not in have:
                            fh.writelines(kept[s])
                print(f'{mode} races kept from the pack: {" ".join(kept) or "none"}', flush=True)
            race_sims(mode, raced, out)
            if os.path.exists(res):
                lines = race_lines_sorted(open(res).readlines(), stems)
                open(res, 'w').writelines(lines)
            note('races', round(time.time() - t1, 1))
            if os.environ.get('TANKHEAL_PROBE') == '1':
                # opt-in (TANKHEAL_PROBE=1): the probe of every preset, packed with the mode for the next probe check; a probe the check made or
                # the old pack holds under the key it would get now is copied, and a stopped run keeps what it probed
                t2 = time.time()
                probe_presets(mode, raced if part else stems, os.path.join(out, 'probe'),
                              reuse=(probe_records(os.path.join(CHECK_DIR, mode)), stored_probes(mode)))
                note('probe', round(time.time() - t2, 1))
        info = run_info(mode, commit_run, secs, round(time.time() - t0, 1), bool(raced))
        if part is not None:  # a worker's share: merge() makes run.json and .fresh from every part
            info.update(me, mode=mode, presets=raced, main='carries' if carried else 'here', probes=os.environ.get('TANKHEAL_PROBE') == '1')
            json.dump(info, open(os.path.join(out, PART), 'w'), indent=1)
            os.remove(prog)
            print(f'{mode} part ({len(raced)} presets raced, main bench {info["main"]}): {info["seconds"]} s')
            continue
        if only is not None:
            info['racesKept'] = sorted(kept)  # presets whose race rows came from the pack
        json.dump(info, open(os.path.join(out, 'run.json'), 'w'), indent=1)
        os.remove(prog)
        # the presets simmed here, for stale.mjs --record-bench: a second line names them when some were kept
        simmed = [s for s in stems if s not in kept]
        open(os.path.join(out, '.fresh'), 'w').write(commit_run + '\n' + (','.join(simmed) + '\n' if kept else ''))
        print(f'{mode}: {info["seconds"]} s')


# ---------------------------------------------------------------- parts: a mode split over worker sessions
#
# A full regeneration can run each mode, and Healing's presets, in separate cloud sessions at once (parallel.py, the
# regenerate-dps-report skill's "Parallel"). Each session runs a part (run --part, --main-only or --carries), packs it
# (tankheal.py part), and the coordinating session merges the parts into $DASH_DIR/<mode> as one run would have left it:
# the same main bench files, races.jsonl and probes byte for byte (a preset's race sims are independent of the other
# presets', and every sim's seed and split count are fixed), then run.json and .fresh. merge refuses parts simmed at
# other commits, inputs, binaries or DASH_SPLITS, main bench outputs that differ, and presets missing or simmed twice
# differently.

PART = 'part.json'


def main_names(mode, stems):
    """The main bench's output files of a mode: one a boss (tank), one a boss and role with presets (heal)."""
    if mode == 'tank':
        return [b + '.txt' for b in BOSSES]
    roles = [r for r in ('tank', 'raid') if any(heal_role(s) == r for s in stems)]
    return [f'{r}_{b}.txt' for b in BOSSES for r in roles]


def read_part(src):
    """A part's part.json (None when it has none) and its files {name: bytes}, from its .tgz or its folder."""
    if os.path.isdir(src):
        files = {os.path.relpath(p, src): open(p, 'rb').read() for p in sorted(glob.glob(os.path.join(src, '**', '*'), recursive=True))
                 if os.path.isfile(p) and not os.path.basename(p).startswith('.')}
    else:
        with tarfile.open(src) as tar:
            files = {ti.name: tar.extractfile(ti).read() for ti in tar.getmembers() if ti.isfile()}
    info = json.loads(files.pop(PART)) if PART in files else None
    return info, files


def part_mismatch(a, b):
    """Why two parts' outputs can't be put together (or a part used as another's carries): [] when they can."""
    a, b = json.loads(json.dumps(a)), json.loads(json.dumps(b))  # as part.json holds them (tuples are lists)
    out = []
    for k, what in (('mode', 'mode'), ('commit', 'commit'), ('inputs', 'sim inputs (stale.mjs --bench-fingerprint)'),
                    ('splits', 'DASH_SPLITS')):
        if a.get(k) != b.get(k):
            out.append(f'other {what}: {a.get(k)} and {b.get(k)}')
    sa, sb = a.get('stamp'), b.get('stamp')
    if not sa or not sb:
        out.append('a part has no binary stamp (bins not built by build_bins.py)')
    elif sa.get('inputHash') != sb.get('inputHash') or sa.get('binaries') != sb.get('binaries'):
        out.append(f'other binaries (STAMP.json): {sa} and {sb}' + (' (other Go versions)' if sa.get('go') != sb.get('go') else ''))
    for k in ('iterations', 'seed', 'raceGoals', 'racePilot', 'raceCeil'):
        if a.get(k) != b.get(k):
            out.append(f'other {k}: {a.get(k)} and {b.get(k)}')
    return out


def carries(mode, src, stems, me):
    """The main bench files of the part at src, for a part that runs only race sims: refused unless it was simmed like
    this run (part_mismatch) and has every main file."""
    info, files = read_part(src)
    if info is None:
        sys.exit(f'{src}: no {PART}, not a part (tankheal.py part)')
    probs = part_mismatch(info, dict(run_info(mode, me['commit'], {}, 0, True), mode=mode, **me))
    if probs:
        sys.exit(f'{src} can\'t be this run\'s main bench:\n' + '\n'.join('  - ' + p for p in probs))
    names = main_names(mode, stems)
    missing = [n for n in names if n not in files]
    if missing:
        sys.exit(f'{src} lacks main bench outputs: {" ".join(missing)}')
    return {n: files[n] for n in names}


def part_pack(mode, dest):
    """$DASH_DIR/<mode>, a finished part, into a deterministic .tgz (the worker's hand-back)."""
    import rawstore
    src = os.path.join(D, mode)
    if not os.path.exists(os.path.join(src, PART)):
        sys.exit(f'{src} holds no finished part ({PART}): tankheal.py run {mode} --part | --main-only | --carries <part> first')
    _, files = read_part(src)
    files[PART] = open(os.path.join(src, PART), 'rb').read()
    os.makedirs(os.path.dirname(os.path.abspath(dest)), exist_ok=True)
    open(dest, 'wb').write(rawstore.tgz(files))
    print(f'{mode} part: {len(files)} files into {dest}')


def merge(mode, sources, keep_rest=False):
    """The parts at sources (.tgz or folders) into $DASH_DIR/<mode>, as one run of the mode would have left it. Every
    preset must be raced by a part, unless keep_rest (a split of run --presets): then the presets no part raced keep the
    pack's race rows when their main rows are the pack's bytes, as run --presets keeps them."""
    import rawstore
    stems = presets(mode)
    parts = []
    for src in sources:
        info, files = read_part(src)
        if info is None:
            sys.exit(f'{src}: no {PART}, not a part (tankheal.py part)')
        if info.get('mode') != mode:
            sys.exit(f'{src}: a {info.get("mode")} part, not {mode}')
        rawstore.check_splits(src, info.get('splits'), SPLITS)
        parts.append((src, info, files))
    if not parts:
        sys.exit(f'merge {mode}: no parts')
    ref = parts[0][1]
    for src, info, _ in parts[1:]:
        probs = part_mismatch(ref, info)
        if probs:
            sys.exit(f'{src} and {parts[0][0]} were not simmed alike:\n' + '\n'.join('  - ' + p for p in probs))
    here = bench_inputs(mode, head())
    if here is not None and here != ref['inputs']:
        sys.exit(f'the parts were simmed on other {mode} sim inputs than HEAD has ({ref["inputs"]}, HEAD {here}): '
                 f'check out their commit {ref["commit"][:10]}, or re-sim')
    names, main = main_names(mode, stems), {}
    for src, info, files in parts:
        for n in names:
            if n in files:
                if n in main and main[n] != files[n]:
                    sys.exit(f'{n} differs between the parts ({src}): the main bench is not the same run')
                main.setdefault(n, files[n])
    if [n for n in names if n not in main]:
        sys.exit(f'no part has the main bench outputs {" ".join(n for n in names if n not in main)}')
    rows, by = {}, {}
    for src, info, files in parts:
        got = {}
        for l in files.get('races.jsonl', b'').decode().splitlines():
            if l.strip():
                got.setdefault(json.loads(l)['build'], []).append(l + '\n')
        said = set(info.get('presets') or [])
        if set(got) - said or said - set(got):
            sys.exit(f'{src}: its races.jsonl holds {sorted(got)}, its {PART} names {sorted(said)}')
        for s, ls in got.items():
            if s in rows and rows[s] != ls:
                sys.exit(f'{s}: its race rows differ between {by[s]} and {src}')
            rows[s], by[s] = ls, src
    if set(rows) - set(stems):
        sys.exit(f'the parts sim presets this checkout doesn\'t have: {" ".join(sorted(set(rows) - set(stems)))}')
    kept, rest = {}, [s for s in stems if s not in rows]
    if rest and not keep_rest:
        sys.exit(f'no part raced {",".join(rest)} (--keep-rest keeps the pack\'s rows of presets no part was to sim)')
    if rest:  # presets no part raced: their pack rows, when their main rows are the pack's bytes (run --presets's rule)
        old = pack_files(mode)
        check_splits(mode, old)
        kept = reusable(mode, rest, set(), old, {n: t.decode() for n, t in main.items()})
        if [s for s in rest if s not in kept]:
            miss = [s for s in rest if s not in kept]
            sys.exit(f'no part raced {",".join(miss)}, and the pack\'s rows can\'t be kept (none, or the main bench rows moved): '
                     f'sim them (tankheal.py run {mode} --presets {",".join(miss)} --carries <a part>, then tankheal.py part) and merge again')
    out = os.path.join(D, mode)
    shutil.rmtree(out, ignore_errors=True)
    os.makedirs(out)
    for n, data in main.items():
        open(os.path.join(out, n), 'wb').write(data)
    lines = race_lines_sorted([l for s in stems for l in rows.get(s) or kept.get(s) or []], stems)
    if lines:
        open(os.path.join(out, 'races.jsonl'), 'w').writelines(lines)
    probes = {}
    for src, info, files in parts:
        for n, data in files.items():
            if n.startswith('probe/'):
                if n in probes and probes[n] != data:
                    sys.exit(f'{n} differs between the parts ({src})')
                probes[n] = data
    if probes and kept:  # a run with TANKHEAL_PROBE=1 keeps a probe of every preset: the kept ones' from the pack
        for n, t in (pack_files(mode) or {}).items():
            if n.startswith('probe/') and n[6:-5] in kept:
                probes.setdefault(n, t.encode())
    for n, data in probes.items():
        os.makedirs(os.path.dirname(os.path.join(out, n)), exist_ok=True)
        open(os.path.join(out, n), 'wb').write(data)
    over = [files['bin_override.json'] for _, _, files in parts if 'bin_override.json' in files]
    if over:  # stale.mjs --record-bench refuses the mode, as for one run with DASH_ALLOW_BIN_MISMATCH
        open(os.path.join(out, 'bin_override.json'), 'wb').write(over[0])
        print(f'{mode}: a part was simmed with DASH_ALLOW_BIN_MISMATCH: the mode is marked (bin_override.json)', file=sys.stderr)
    mains = [i for _, i, f in parts if i.get('main') == 'here']
    secs = {k: v for k, v in (mains[0] if mains else ref)['runSeconds'].items() if k != 'races' and k != 'probe'}
    raced = [i['runSeconds'].get('races', 0) for _, i, _ in parts if i.get('presets')]
    if raced:
        secs['races'] = max(raced)  # the parts ran at once: the longest
    info = {k: ref[k] for k in run_info(mode, '', {}, 0, True)}
    info.update(at=max(i['at'] for _, i, _ in parts), seconds=max(i['seconds'] for _, i, _ in parts), runSeconds=secs, races=bool(rows) or bool(kept),
                parts=[{'presets': i.get('presets'), 'main': i.get('main'), 'seconds': i['seconds']} for _, i, _ in parts])
    if kept:
        info['racesKept'] = sorted(kept)
    json.dump(info, open(os.path.join(out, 'run.json'), 'w'), indent=1)
    open(os.path.join(out, '.fresh'), 'w').write(info['commit'] + '\n' + (','.join(s for s in stems if s not in kept) + '\n' if kept else ''))
    print(f'{mode}: merged {len(parts)} parts, {len(rows)} presets raced' + (f', {len(kept)} kept from the pack' if kept else '') + f' into {out}')


# ---------------------------------------------------------------- probes
#
# A preset's probe is its sims at fixed reduced sizes, a tenth of the full run's: the main bench at ITERS / 10 and the
# race sims with every pilot and ceiling of adaptive.py divided by 10. It is compared with the probe the last full run
# made at the same sizes, never with the full run itself: the carry's bisection and the race lattice walk pick their
# scales from the death shares seen so far, so a fight prefix of the full run is not what fewer fights sim.

PROBE_SCALE = 10
PROBE_P = int(os.environ.get('TANKHEAL_PROBE_P', 2))  # presets at a time, each sim split DASH_SPLITS ways
CHECK_DIR = os.path.join(D, 'probe_check', 'bench')  # a check's probes and audit.json, by mode (hydrate empties it)
# the report switches sim/core/report_switches.go reads: each one set in the caller's environment changes the sims
SWITCH_VARS = ('DASH_NAMES', 'DASH_SPLITS', 'DASH_RACIALS_OFF', 'DASH_RACIALS_OFF_CLASS', 'DASH_ASSUME')
PROBE_SCRIPTS = ('tankheal.py', 'bench_races.py', 'adaptive.py', 'dashpaths.py', 'run_build.py', 'probe_cli.py')
HEAL_RAID = ('tools/tankbench/req/warrior_protection.json', 'tools/dps_report/req/mage__frost.json')  # healbench -tank, -filler
_hashes = {}


def probe_settings():
    """The probe's sizes: a tenth of the full run's, recorded in every probe (a probe of other sizes isn't compared)."""
    s = PROBE_SCALE
    return {'format': 1, 'scale': s, 'iters': ITERS // s, 'racePilot': adaptive.RACE_PILOT // s, 'pilot': adaptive.PILOT // s,
            'ceil': adaptive.CEIL // s, 'healCeil': adaptive.HEAL_CEIL // s, 'splits': SPLITS, 'arch': platform.machine()}


def file_hash(path):
    """sha256 of a file ('-' when it is missing), kept for the process: the binaries are tens of megabytes."""
    if path not in _hashes:
        try:
            _hashes[path] = hashlib.sha256(open(path, 'rb').read()).hexdigest()
        except OSError:
            _hashes[path] = '-'
    return _hashes[path]


def probe_key(mode, stem, settings):
    """What a preset's probe is made from (probe.py's _key for the benches): the settings, the bench binary behind any
    timing link (build_bins.real_bin), the preset's request, the boss files, healbench's tank and filler requests, and
    the scripts. When it is equal, the probe is too."""
    import build_bins
    exe = build_bins.BENCH_BINS[mode]
    k = {'settings': settings, 'bin': file_hash(build_bins.real_bin(os.path.join(BIN, exe))), 'req': file_hash(req_path(mode, stem)),
         'encounters': {b: file_hash(os.path.join(TREE, 'tools/encounters', b + '.json')) for b in BOSSES},
         'scripts': {f: file_hash(os.path.join(HERE, f)) for f in PROBE_SCRIPTS},
         'switches': {v: os.environ[v] for v in SWITCH_VARS if v in os.environ}}
    if mode == 'heal':
        k['raid'] = {p: file_hash(os.path.join(TREE, p)) for p in HEAL_RAID}
    return hashlib.sha256(json.dumps(k, sort_keys=True).encode()).hexdigest()[:20]


def probe_env(mode, settings, work):
    """The environment of one preset's probe: its own DASH_DIR (so the race vectors start empty) and the reduced sizes."""
    return dict(os.environ, DASH_DIR=work, DASH_TREE=TREE, TANKHEAL_BIN=BIN, DASH_SPLITS=str(settings['splits']),
                ADAPT_PILOT=str(settings['pilot']), ADAPT_RACE_PILOT=str(settings['racePilot']), ADAPT_MAX=str(settings['ceil']),
                ADAPT_HEAL_MAX=str(settings['healCeil']), TANKHEAL_PROBE_ITERS=str(settings['iters']), TANKHEAL_STEP=mode + '_probe')


def boss_files(mode, stem):
    return list(BOSSES) if mode == 'tank' else [f'{heal_role(stem)}_{b}' for b in BOSSES]


def probe_one(mode, stem, work):
    """One preset's probe, in a process of its own (probe_env): the main bench on each boss at the probe's iterations,
    then its race sims (race_sims at adaptive.py's reduced sizes), fingerprinted into <work>/probe.json."""
    from probe_cli import digest
    for boss, name in zip(BOSSES, boss_files(mode, stem)):
        args = ['-encounter', f'tools/encounters/{boss}.json', '-iters', os.environ['TANKHEAL_PROBE_ITERS']]
        bench(mode, args if mode == 'tank' else ['-carry', '-role', heal_role(stem)] + args, [req_path(mode, stem)], os.path.join(work, name + '.txt'))
    race_sims(mode, [stem], work)
    texts = {n: open(os.path.join(work, n)).read() for n in os.listdir(work) if n.endswith('.txt')}
    rows = main_rows(texts, mode, stem)
    if rows is None or not all(rows):
        sys.exit(f'{mode} {stem}: the main bench printed no row for it')
    score = 'score' if mode == 'tank' else 'carry'
    rec = {'main': {n: digest(list(r)) for n, r in zip(boss_files(mode, stem), rows)}, 'races': {}, 'values': {}}
    for l in open(os.path.join(work, 'races.jsonl')):
        r = json.loads(l)
        k = f'{r["race"]}:{r["variant"]}'
        rec['races'][k] = digest(r)
        rec['values'][k] = r.get(score)  # to say how far a moved probe moved; compare reads the fingerprints
    json.dump(rec, open(os.path.join(work, 'probe.json'), 'w'))


def probe_records(d):
    """The probe records in a folder: {stem: record}."""
    out = {}
    for p in glob.glob(os.path.join(d, '*.json')):
        if os.path.basename(p) not in ('audit.json', 'bin_override.json'):
            try:
                out[os.path.basename(p)[:-5]] = json.load(open(p))
            except ValueError:
                pass
    return out


def stored_probes(mode):
    """The probes in raw/bench/<mode>.tgz: {stem: record} (none in a pack made before probes were kept)."""
    files = pack_files(mode) or {}
    return {n[len('probe/'):-5]: json.loads(t) for n, t in files.items() if n.startswith('probe/') and n.endswith('.json')}


def probe_presets(mode, stems, dest, reuse=()):
    """Probe these presets into dest/<stem>.json: {stem: record}. A record already there with the key it would get now is
    kept (a stopped run restarts where it was), and one in a reuse source ({stem: record}) with it is copied (a full run
    after a check, or a pack of the same binary and inputs). PROBE_FORCE=1 probes everything."""
    from concurrent.futures import ThreadPoolExecutor
    settings, force = probe_settings(), os.environ.get('PROBE_FORCE') == '1'
    os.makedirs(dest, exist_ok=True)
    base = os.path.join(D, 'probe_work', 'bench_' + mode)

    def save(path, rec):
        with open(path, 'w') as fh:
            json.dump(rec, fh, indent=0, sort_keys=True)
            fh.write('\n')

    def one(s):
        key, path = probe_key(mode, s, settings), os.path.join(dest, s + '.json')
        if not force:
            have = probe_records(dest).get(s) if os.path.exists(path) else None
            for i, rec in enumerate([have] + [r.get(s) for r in reuse]):
                if rec and rec.get('_key') == key and not rec.get('failed'):
                    if i:
                        save(path, rec)
                    print(f'{mode} probe {s}: {"kept" if not i else "copied"} (same key)', flush=True)
                    return s, rec
        work = os.path.join(base, s)
        shutil.rmtree(work, ignore_errors=True)
        os.makedirs(work)
        t = time.time()
        r = subprocess.run([sys.executable, os.path.abspath(__file__), '_probe1', mode, s, work], env=probe_env(mode, settings, work),
                           capture_output=True, text=True)
        rec = {'settings': settings, '_key': key}
        if r.returncode or not os.path.exists(os.path.join(work, 'probe.json')):
            rec['failed'] = r.returncode or 1
            print(f'{mode} probe {s} FAILED (exit {r.returncode}):\n{(r.stdout + r.stderr)[-1500:]}', file=sys.stderr, flush=True)
        else:
            rec.update(json.load(open(os.path.join(work, 'probe.json'))))
        save(path, rec)
        shutil.rmtree(work, ignore_errors=True)
        print(f'{mode} probe {s}: {len(rec.get("races", {}))} race rows, {round(time.time() - t)} s', flush=True)
        return s, rec

    with ThreadPoolExecutor(PROBE_P) as ex:
        res = dict(ex.map(one, stems))
    shutil.rmtree(base, ignore_errors=True)
    return res


def probe_compare(old, new):
    """What differs between two probes of a preset: [] when they're the same."""
    if old is None:
        return ['no stored probe (the mode was last simmed before bench probes were kept)']
    if old.get('settings') != new.get('settings'):
        return [f'the stored probe has other settings ({old.get("settings")})']
    if old.get('failed') or new.get('failed'):
        return [f'a probe failed (stored: {old.get("failed") or "no"}, new: {new.get("failed") or "no"})']
    out = []
    main = sorted(k for k in old['main'].keys() | new['main'].keys() if old['main'].get(k) != new['main'].get(k))
    if main:
        out.append('main bench: ' + ', '.join(main))
    o, n = old['races'], new['races']
    changed, gone, added = [k for k in o.keys() & n.keys() if o[k] != n[k]], o.keys() - n.keys(), n.keys() - o.keys()
    if changed or gone or added:
        moves = [abs(new['values'][k] / old['values'][k] - 1) for k in changed
                 if (old.get('values') or {}).get(k) and (new.get('values') or {}).get(k) is not None]
        out.append('races: ' + ', '.join(x for x in (
            f'{len(changed)} of {len(o.keys() & n.keys())} rows differ' + (f' (score up to {100 * max(moves):.2f}% apart)' if moves else '') if changed else '',
            f'{len(gone)} rows gone' if gone else '', f'{len(added)} new rows' if added else '') if x))
    return out


def audit_pick(cleared, commit, mode, n=1):
    """The cleared presets a check re-sims in full anyway, seeded by the commit and the mode (probe.py's audit)."""
    rng = random.Random(f'{commit}:{mode}')
    return sorted(rng.sample(sorted(cleared), min(n, len(cleared))))


def head():
    return subprocess.run(['git', 'rev-parse', 'HEAD'], cwd=ROOT, capture_output=True, text=True).stdout.strip()


def fresh_simmed(mode):
    """The presets the mode's last run simmed (its .fresh), None when it hasn't run since the last publish."""
    p = os.path.join(D, mode, '.fresh')
    if not os.path.exists(p):
        return None
    lines = open(p).read().split('\n')
    return set(lines[1].split(',')) if len(lines) > 1 and lines[1].strip() else set(presets(mode))


def probe_guard(mode, what, marker):
    """The bin check before a probe: a probe on binaries that don't match the code is marked (marker), so stale.mjs
    --record-bench refuses what it vouches for."""
    import build_bins
    rec = build_bins.guard(BIN, [build_bins.BENCH_BINS[mode]], what)
    if rec:
        build_bins.mark(marker, rec)
    elif os.path.exists(marker) and 'probe_check' in marker:
        os.remove(marker)


def probe_write(modes, want=None, stale_ok=False):
    """tankheal.py probe: the probes the packs keep, into $DASH_DIR/<mode>/probe/ (pack after). A probe must describe the
    sims of the pack it goes into, so a preset stale.mjs calls stale is refused unless this $DASH_DIR's last run simmed
    it (.fresh) or --stale-ok (a throwaway measurement)."""
    stale = {} if stale_ok else stale_presets()
    for mode in modes:
        stems = [s for s in presets(mode) if want is None or s in want]
        simmed = fresh_simmed(mode) or set()
        bad = [s for s in stems if s in stale.get(mode, ()) and s not in simmed]
        if bad:
            print(f'{mode}: not probing {" ".join(bad)}: stale, so the pack\'s sims are not this code\'s (re-sim them, or --stale-ok)', file=sys.stderr)
            stems = [s for s in stems if s not in bad]
        if not stems:
            continue
        probe_guard(mode, f'tankheal.py probe {mode}', os.path.join(D, mode, 'bin_override.json'))
        t = time.time()
        probe_presets(mode, stems, os.path.join(D, mode, 'probe'))
        print(f'{mode}: probed {len(stems)} presets in {round(time.time() - t)} s')


def probe_check(modes, want=None, as_json=False):
    """tankheal.py probe check: probe the presets (default: the stale ones) into probe_check/bench/<mode>/, compare each
    with the probe in the mode's pack, and write audit.json there: the presets it cleared, the one it re-sims anyway (the
    audit) and the ones that moved."""
    stale = stale_presets() if want is None else None
    commit, report, no_probe = head(), {}, []
    for mode in modes:
        if not stored_probes(mode):
            print(f'{mode}: the pack has no probe to compare with (made before probes were kept, or run without TANKHEAL_PROBE=1): '
                  f're-sim the mode with tankheal.py run {mode}, or make probes with tankheal.py probe {mode} afterwards', file=sys.stderr)
            no_probe.append(mode)
            continue
        stems = sorted(want & set(presets(mode)) if want is not None else stale.get(mode, ()))
        if not stems:
            print(f'{mode}: nothing to probe')
            continue
        dest = os.path.join(CHECK_DIR, mode)
        os.makedirs(dest, exist_ok=True)
        probe_guard(mode, f'tankheal.py probe check {mode}', os.path.join(dest, 'bin_override.json'))
        t = time.time()
        new, old = probe_presets(mode, stems, dest), stored_probes(mode)
        moved, cleared = {}, []
        for s in stems:
            why = probe_compare(old.get(s), new[s])
            if why:
                moved[s] = why
            else:
                cleared.append(s)
        audit = audit_pick(cleared, commit, mode)
        passed = sorted(set(cleared) - set(audit))
        json.dump({'commit': commit, 'cleared': passed, 'audit': audit, 'moved': sorted(moved), 'why': moved},
                  open(os.path.join(dest, 'audit.json'), 'w'), indent=1)
        report[mode] = {'cleared': passed, 'audit': audit, 'moved': moved}
        if as_json:
            continue
        print(f'{mode}: {len(cleared)} of {len(stems)} presets come out the same: {" ".join(cleared) or "none"} ({round(time.time() - t)} s)')
        for s, why in sorted(moved.items()):
            print(f'  {s}: {"; ".join(why)}')
        print(f'  audit (cleared, but its races re-simmed in full): {" ".join(audit) or "none"}')
        resim = sorted(moved) + audit
        print(f'  next: tankheal.py run {mode} --presets {",".join(resim)}' if resim else '  next: nothing to sim')
    if as_json:
        print(json.dumps(report, indent=1))
    elif report:
        print('then: pipeline.sh publish: it checks the audit and records the cleared presets (stale.mjs --record-bench --by probe)')
    if no_probe:
        sys.exit(1)


def probe_audit():
    """publish's check, before the pack is rewritten: for every mode a check left an audit.json for, the run since must
    have re-simmed the audit preset, its main rows and race rows must be the pack's, and every cleared preset's main
    rows (the full main bench, which ran whole) must be the pack's too. Exit 1 otherwise."""
    for mode in MODES:
        p = os.path.join(CHECK_DIR, mode, 'audit.json')
        if not os.path.exists(p):
            continue
        a = json.load(open(p))
        simmed = fresh_simmed(mode)
        missing = [s for s in a['audit'] if simmed is None or s not in simmed]
        if simmed is None or missing:
            sys.exit(f'{mode} probe audit: run the main bench and re-sim the audit preset first (tankheal.py run {mode} --presets '
                     f'{",".join(sorted(set(a["moved"]) | set(a["audit"])))}), or delete {p} to record nothing by probe')
        old, out = pack_files(mode) or {}, os.path.join(D, mode)
        new = {n: open(os.path.join(out, n)).read() for n in os.listdir(out) if n.endswith('.txt') or n == 'races.jsonl'}
        bad = [s for s in a['audit'] if main_rows(old, mode, s) != main_rows(new, mode, s) or race_rows_of(old, s) != race_rows_of(new, s)]
        bad += [s for s in a['cleared'] if main_rows(old, mode, s) != main_rows(new, mode, s)]
        if bad:
            sys.exit(f'{mode} probe audit FAILED: {" ".join(sorted(set(bad)))} passed the probe, but the full sims moved. The probe '
                     f'missed a change, so its other clearances can\'t be trusted: re-sim them (tankheal.py run {mode} --presets '
                     f'{",".join(sorted(set(a["cleared"]) | set(a["audit"]) | set(a["moved"])))}), then delete {p} and publish')
        print(f'{mode} probe audit: the main bench unchanged for {len(a["cleared"])} cleared presets, the audit preset\'s races '
              f'({" ".join(a["audit"]) or "none"}) unchanged')


def race_rows_of(files, stem):
    """A preset's races.jsonl rows in a mode's files, in order."""
    return [json.loads(l) for l in (files.get('races.jsonl') or '').splitlines() if l.strip() and json.loads(l)['build'] == stem]


def probe_passed():
    """The presets the checks cleared and no run since re-simmed, as stale.mjs --record-bench takes them (<mode>:<stem,...>)."""
    out, commit = [], head()
    for mode in MODES:
        p = os.path.join(CHECK_DIR, mode, 'audit.json')
        if not os.path.exists(p):
            continue
        a = json.load(open(p))
        if a['commit'] != commit:
            print(f'{p} is from {a["commit"][:10]}, not HEAD: its presets are not recorded (probe again)', file=sys.stderr)
            continue
        simmed = fresh_simmed(mode) or set()
        cleared = [s for s in a['cleared'] if s not in simmed]
        if cleared:
            out.append(f'{mode}:{",".join(cleared)}')
    return out


# ---------------------------------------------------------------- packs

def pack():
    import rawstore
    os.makedirs(RAW, exist_ok=True)
    for mode in MODES:
        src = os.path.join(D, mode)
        if not os.path.isdir(src):
            continue
        members = {}
        for p in sorted(glob.glob(os.path.join(src, '**', '*'), recursive=True)):
            if os.path.isfile(p) and not os.path.basename(p).startswith('.'):
                members[os.path.relpath(p, src)] = open(p, 'rb').read()
        data = rawstore.tgz(members)
        dst = os.path.join(RAW, mode + '.tgz')
        same = os.path.exists(dst) and open(dst, 'rb').read() == data
        if not same:
            open(dst, 'wb').write(data)
        print(f'{mode}: {len(members)} files, pack {"unchanged" if same else "written"}')


def unpack(modes=MODES):
    for mode in modes:
        p = os.path.join(RAW, mode + '.tgz')
        if not os.path.exists(p):
            print(f'no pack for {mode}')
            continue
        check_splits(mode, pack_files(mode))
        dst = os.path.join(D, mode)
        shutil.rmtree(dst, ignore_errors=True)
        with tarfile.open(p) as tar:
            for ti in tar.getmembers():
                if ti.isfile():
                    q = os.path.join(dst, ti.name)
                    os.makedirs(os.path.dirname(q), exist_ok=True)
                    open(q, 'wb').write(tar.extractfile(ti).read())
        print(f'unpacked {mode} into {dst}')


# ---------------------------------------------------------------- parsing

def _num(s):
    m = re.match(r'^\(?(-?[\d.]+)', s.replace(',', ''))
    return float(m.group(1)) if m else None


def _pm(s):
    """'4648 ± 15' -> (4648, 15); 'none' -> (None, None)."""
    m = re.match(r'^(-?[\d.]+) ± ([\d.]+)', s)
    return (float(m.group(1)), float(m.group(2))) if m else (None, None)


def table(text):
    """The bench's markdown table: [{header: cell}]. 'Took: avoid | block | crit | crush' is four columns."""
    rows, head = [], None
    for ln in text.split('\n'):
        if not ln.startswith('|'):
            continue
        cells = [c.strip() for c in ln.strip().strip('|').split('|')]
        if head is None:
            head = cells
        elif not set(ln) <= set('|- '):
            rows.append(dict(zip(head, cells)))
    return rows


def tank_rows(text):
    out = {}
    for r in table(text):
        sc = {k: _pm(r[n]) for k, n in (('scoreMaxThreat', 'Max threat'), ('scoreBalanced', 'Balanced'), ('scoreMaxSurvival', 'Max survival'))}
        safe = _pm(r['SafeDPS'])
        need = _pm(r['Healing needed'])
        mx = re.search(r'\(([\d.]+)x DTPS\)', r['Healing needed'])
        bh = re.match(r'([\d.]+) \(([\d.]+) to kill\)', r['Biggest hit'])
        ar = re.match(r'([\d.]+) \(([\d.]+)%\)', r['Armor'])
        row = {}
        for k, (v, e) in sc.items():
            row[k], row[k + 'Se'] = v, e
        row.update({'safeDps': safe[0], 'safeDpsSe': safe[1], 'healingNeeded': need[0], 'healingNeededSe': need[1],
                    'healingNeededX': float(mx.group(1)) if mx else None,
                    'healingNone': need[0] is None or any(v is None for v, _ in sc.values()),
                    'deathFloor': _num(r['Death floor']), 'tps': _num(r['TPS']), 'dps': _num(r['DPS']), 'dtps': _num(r['DTPS']),
                    'tmi': _num(r['TMI']), 'death': _num(r['Death']), 'health': _num(r['Health']), 'eh': _num(r['EH']),
                    'biggestHit': float(bh.group(1)) if bh else _num(r['Biggest hit']), 'hitsToKill': float(bh.group(2)) if bh else None,
                    'armor': float(ar.group(1)) if ar else _num(r['Armor']), 'armorReduction': float(ar.group(2)) if ar else None,
                    'def': _num(r['Def']), 'noCritAt': _num(r['No crit at']), 'miss': _num(r['Miss']), 'dodge': _num(r['Dodge']),
                    'parry': _num(r['Parry']), 'block': _num(r['Block']), 'crit': _num(r['Crit']), 'crush': _num(r['Crush']),
                    'crushGap': _num(r['Crush gap']), 'tookAvoid': _num(r['Took: avoid']), 'tookBlock': _num(r['block']),
                    'tookCrit': _num(r['crit']), 'tookCrush': _num(r['crush'])})
        out[r['Build']] = row
    return out


def tank_header(text):
    h = {}
    for ln in text.split('\n'):
        for k, pre in (('healingLine', 'Healing:'), ('notes', 'Encounter:'), ('standIn', 'Stand-in healer:'), ('objective', 'Objective:')):
            if ln.startswith(pre):
                h[k] = ln[len(pre):].strip() if k != 'healingLine' else ln.strip()
    if 'notes' in h:  # "Encounter: Magmadar. <notes>"
        h['notes'] = re.sub(r'^[^.]+\.\s*', '', h['notes'], count=1)
    return h


def heal_rows(text):
    out = {}
    for r in table(text):
        fl = re.match(r'([\d.]+)% tank, ([\d.]+)% other', r.get('Death floor', ''))
        oom = r['OOM at']
        out[r['Healer']] = {
            'carry': _num(r['Carry']), 'deathFloorTank': float(fl.group(1)) if fl else None, 'deathFloorOther': float(fl.group(2)) if fl else None,
            'hps': _num(r['HPS']), 'ehps': _num(r['EHPS']), 'overheal': _num(r['Overheal']), 'ehPerMana': _num(r['EH per mana']),
            'manaLeft': _num(r['Mana left']), 'oomAt': _num(oom), 'oomInFight': not oom.startswith('('),
            'tankDtps': _num(r['Tank DTPS']), 'raidDtps': _num(r['Raid DTPS']), 'tankDied': _num(r['Tank died']), 'otherDied': _num(r['Other died']),
            'firstDeath': _num(r['First death']), 'tankScares': _num(r['Tank scares']), 'raidScares': _num(r['Raid scares'])}
    return out


def heal_header(text):
    lines = [l for l in text.split('\n') if l.strip()]
    m = re.match(r'^(.*?), role (\w+) ×[\d.]+, (\d+) s, (\d+) iterations', lines[0])
    return {'name': m.group(1), 'duration': int(m.group(3)), 'iterations': int(m.group(4)), 'notes': lines[1] if not lines[1].startswith('|') else '',
            'setup': lines[0].split('a sim. ', 1)[-1]}


# ---------------------------------------------------------------- the DPS pipeline's helpers

def dps_helpers():
    """build_data.py's name, icon, gear and setup helpers, without running its build: its statements up to where it
    reads the outputs, and the few functions it defines later that only use those."""
    src = open(os.path.join(HERE, 'build_data.py')).read()
    body = ast.parse(src).body
    cut = next(i for i, n in enumerate(body) if isinstance(n, ast.Assign) and getattr(n.targets[0], 'id', '') == 'outs')
    late = {'icon_of', 'nm', 'unship', '_SHIP', '_SUBS', '_clean_note', 'fetch_url', 'icon_color'}
    keep = body[:cut] + [n for n in body[cut:] if (isinstance(n, ast.FunctionDef) and n.name in late) or
                         (isinstance(n, ast.Assign) and getattr(n.targets[0], 'id', '') in late)]
    ns = {'__file__': os.path.join(HERE, 'build_data.py'), '__name__': 'build_data_helpers'}
    exec(compile(ast.Module(body=keep, type_ignores=[]), 'build_data.py', 'exec'), ns)
    import base64, colorsys, urllib.error, urllib.request
    from PIL import Image
    ns.update(names={}, base64=base64, colorsys=colorsys, urllib=urllib, Image=Image, time=time)
    return ns


def common(name):
    """A common dataset: this run's ($DASH_DIR/data, build_data.py's) or else the published one."""
    p = os.path.join(D, 'data', name + '.json')
    return json.load(open(p if os.path.exists(p) else os.path.join(ROOT, 'ui', 'dps_report', 'data', name + '.json')))


def icon_rows(H, want, large):
    """The icons the common icons and icons_more datasets lack, as their rows (data URLs; build_data.py's fetch)."""
    from concurrent.futures import ThreadPoolExecutor
    have = {r['icon'] for r in common('icons') + common('icons_more')}
    os.makedirs(os.path.join(D, 'icons'), exist_ok=True)

    def fetch(a):
        ic, size = a
        path = os.path.join(D, 'icons', ic + ('.large' if size == 'large' else '') + '.jpg')
        if not os.path.exists(path) and not H['fetch_url'](f'https://wow.zamimg.com/images/wow/icons/{size}/{ic}.jpg', path):
            return None
        return {'icon': ic + (':l' if size == 'large' else ''), 'url': 'data:image/jpeg;base64,' + H['base64'].b64encode(open(path, 'rb').read()).decode()}
    todo = [(ic, 'medium') for ic in sorted(filter(None, want)) if ic not in have] + [(ic, 'large') for ic in sorted(filter(None, large)) if ic + ':l' not in have]
    with ThreadPoolExecutor(16) as ex:
        got = list(ex.map(fetch, todo))
    missing = [a[0] for a, g in zip(todo, got) if not g]
    if missing:
        print('missing icons:', missing)
    return [g for g in got if g]


def words(k):
    return re.sub(r'(?<=[a-z0-9])(?=[A-Z])', ' ', k[0].upper() + k[1:])


def gear_setup(H, bid, req):
    """The DPS gear and setup rows (build_data.py) for a request."""
    import setup_active
    pl = player(req)
    gear, setup = [], []
    for i, it in enumerate(pl.get('equipment', {}).get('items', [])[:len(H['SLOTS'])]):
        if not it or not it.get('id'):
            continue
        info = H['item_names'].get(it['id'], {})
        gear.append({'build': bid, 'slot': H['SLOTS'][i], 'item': info.get('name', 'Item %d' % it['id']), 'ilvl': info.get('ilvl'),
                     'enchant': H['enchant_name'](it.get('enchant'), i, it['id']), 'itemId': it['id'],
                     'icon': info.get('icon'), 'quality': H['QUALITY'].get(info.get('quality'), 'common')})
    for cat, obj in (('Consumes', pl.get('consumes', {})), ('Self buffs', pl.get('buffs', {})), ('Raid buffs', req['raid'].get('buffs', {})),
                     ('Party buffs', req['raid']['parties'][0].get('buffs', {})), ('Debuffs', req['raid'].get('debuffs', {}))):
        for k, v in (obj or {}).items():
            if v in (False, 0, '', None) or v == 'TristateEffectMissing':
                continue
            if cat == 'Consumes':
                txt = words(k) + ': ' + (words(str(v)) if isinstance(v, str) else str(v))
            else:
                txt = words(k) + (' (improved)' if v == 'TristateEffectImproved' else '' if v in (True, 'TristateEffectRegular') else f' ({v})')
            slot, so = H['CONSUME_SLOT'].get(k, ('Other', 90)) if cat == 'Consumes' else (None, 0)
            setup.append({'build': bid, 'category': cat, 'item': txt, 'icon': H['setup_icon'](k, v, cat), 'group': H['SETUP_GROUP'][cat], 'slot': slot,
                          'slotOrder': so, 'hand': 'MH' if k == 'mainHandImbue' else 'OH' if k == 'offHandImbue' else None,
                          'level': 'Improved' if v == 'TristateEffectImproved' else (str(v) if isinstance(v, int) and not isinstance(v, bool) and v > 1 else None)})
    for p in (pl.get('profession1', ''), pl.get('profession2', '')):
        if p:
            setup.append({'build': bid, 'category': 'Professions', 'item': words(p), 'icon': H['setup_icon'](p, p, 'Professions'), 'group': 'Professions',
                          'slot': None, 'slotOrder': 0, 'hand': None, 'level': None})
    # no setup ablation for the benches: every entry counts as active, the sim's rules still say which are gone
    for r in setup:
        k = setup_active.key_of(r)
        r['active'], r['why'], r['note'] = True, None, None
        if k in setup_active.GONE:
            r['active'], r['why'] = False, setup_active.GONE[k]
        elif '(improved)' in r['item'] and k in setup_active.IMPROVED_IGNORED:
            r['note'] = 'WoW Forever has no Improved version. The sim uses the regular one.'
    return gear, setup


def identity(H, bid, cls, req):
    pl = player(req)
    race = H['race_name'](race_of(req))
    tal = pl.get('talentsString', '')
    fac = 'Horde' if race in ('Orc', 'Undead', 'Tauren', 'Troll', 'Windshaper') else 'Alliance'
    return {'race': race, 'faction': fac, 'talents': tal, 'treePoints': H['tree_points'](tal),
            'trees': ' / '.join(t['name'] for t in H['TREE_JSON'][cls]), 'classIcon': 'classicon_' + cls.lower(),
            'specIcon': H['spec_icon'](cls, tal) if tal else None, 'raceIcon': H['RACE_ICON'].get(race)}


# spells build_data.py's name sources miss (ranks the sim keys in struct literals, not a *SpellId array): name, rank
SPELL_NAMES = {1235827: ('Lacerate', 5), 1235826: ('Lacerate', 4), 1238070: ('Primal Bite', 3), 1238069: ('Primal Bite', 2),
              1238215: ('Wild Growth', 3), 1238214: ('Wild Growth', 2), 1240827: ('Prayer of Mending', 3), 1240826: ('Prayer of Mending', 2),
              1239243: ('Riptide', 3), 1239242: ('Riptide', 2)}


def tank_nm(H):
    """build_data's nm(), with SPELL_NAMES and its '<Name> Rank' reading of rank arrays (SealOfFuryRankSpellId) fixed."""
    def nm(key):
        names = H['names']
        if key not in names:
            m = re.match(r'^s(\d+)$', key or '')
            if m and int(m.group(1)) in SPELL_NAMES:
                n, rk = SPELL_NAMES[int(m.group(1))]
                names[key] = {'key': key, 'name': n, 'rank': rk, 'icon': H['wiki_icon'].get(n.lower()) or H['spell_by_name'].get(n.lower())}
            else:
                H['nm'](key)
                n = names[key]['name']
                if n.endswith(' Rank'):
                    n = re.sub(r' (Of|The|And) ', lambda x: x.group(0).lower(), n[:-5])
                    names[key]['name'] = n
                    names[key]['icon'] = names[key]['icon'] or H['wiki_icon'].get(n.lower()) or H['spell_by_name'].get(n.lower())
        return names[key]['name']
    return nm


def rotation_rows(H, rot_mod, talents, bid, cls, req, rules=()):
    pl = player(req)
    nm = tank_nm(H)
    Hm = rot_mod._humaniser(nm, cls, talents, pl.get('talentsString'))
    rows = rot_mod.apl_rows(bid, pl.get('rotation') or {}, H['race_name'](race_of(req)), nm, H['names'], {}, H['RACIAL_RACE'], Hm)
    for r in rows:
        r['casts'] = None  # the benches print no casts
        if r['key'] and not r['icon']:
            r['icon'] = H['names'].get(r['key'], {}).get('icon')
        if r['key'] == 'oOtherActionPotion':
            pot = next((v for k, v in (pl.get('consumes') or {}).items() if k == 'defaultPotion'), None)
            if pot:
                r['name'] = words(pot)
                r['icon'] = H['setup_icon']('defaultPotion', pot, 'Consumes')
        r['notes'] = rot_mod.link_talents(reword(H['_clean_note'](r['notes']), rules), cls, talents, r['name'])
        r['why'] = r['notes']  # the bench presets have no rotation_text.json lines: their own notes, reworded, are the why
    return rows, Hm.fallbacks


# ---------------------------------------------------------------- the page's text

def page_text(mode):
    """tankheal_text/<mode>.md: {'Presets': {id: md}, 'Notes': [(head, md)], 'Bosses': {boss: md}, 'Method': [(title, md)],
    'Rotation': [(regex, replacement)]}. The file's header comment says what each section holds."""
    md = re.sub(r'<!--.*?-->', '', open(os.path.join(TEXT, mode + '.md')).read(), flags=re.S)
    secs = {}
    for sec in re.split(r'^# ', md, flags=re.M)[1:]:
        name, _, body = sec.partition('\n')
        if name.strip() == 'Rotation':
            secs['Rotation'] = [(a, b.strip()) for a, b in re.findall(r'^from: (.*)\nto:(.*)$', body, re.M)]
            continue
        ents = []
        for ent in re.split(r'^## ', body, flags=re.M)[1:]:
            head, _, txt = ent.partition('\n')
            ents.append((head.strip(), txt.strip()))
        secs[name.strip()] = ents
    for k in ('Presets', 'Bosses'):
        secs[k] = dict(secs.get(k, []))
    return secs


def reword(t, rules):
    """A rotation note through the text file's Rotation rules, in order."""
    for a, b in rules:
        t = re.sub(a, b, t, flags=re.M) if t else t
    return t


def need(mode, what, have, want):
    miss = [w for w in want if w not in have]
    if miss:
        sys.exit(f'tankheal_text/{mode}.md: no {what} entry for {", ".join(miss)}')


def notes_rows(md2html, T, builds):
    """The "Good to know" notes: each to every preset its heading names (a class: all of the class's presets)."""
    rows = []
    for head, body in T['Notes']:
        who, _, title = head.partition(':')
        names = [w.strip() for w in who.split(',')]
        ids = [b['id'] for b in builds if b['class'] in names or b['id'] in names]
        bad = [n for n in names if not any(b['class'] == n or b['id'] == n for b in builds)]
        if bad or not title.strip():
            sys.exit(f'tankheal_text: note "{head}" names no preset or class: {bad or "(no title)"}')
        for cls in dict.fromkeys(b['class'] for b in builds if b['id'] in ids):
            rows.append({'build': None, 'class': cls, 'kind': 'note', 'title': title.strip(), 'html': md2html.convert(body),
                         'builds': [i for i in ids if i.split('__')[0] == cls.lower()]})
    return rows


def secs(n):
    n = int(round(n or 0))
    return f'{n} sec' if n < 60 else f'{n // 60} min {n % 60} sec' if n % 60 else f'{n // 60} min'


def method_rows(md2html, mode, T, encs):
    if mode == 'tank':
        lst = '\n'.join(f"- **{e['name']}** ({secs(e['duration'])}): {T['Bosses'][e['boss']]}" for e in encs)
    else:
        lst = '\n'.join(f"- **{e['name']}** (tank {secs(e['tankDuration'])}, raid {secs(e['raidDuration'])}): {T['Bosses'][e['boss']]}" for e in encs)
    return [{'order': i, 'title': t, 'html': md2html.convert(body.replace('{bosses}', lst))} for i, (t, body) in enumerate(T['Method'])]


# The words that mean the page shows our tools, files or working notes instead of the game: the build stops on any.
LEAK = re.compile(r'\btools/|\bdocs/|orchestration/|\.(?:json|go|md|py)\b|\brotopt\b|tankbench|healbench|\bbench|\bthe user\b|\bAPLs?\b|'
                  r'cmangos|\biterations?\b|\bseeds?\b|\bknobs?\b|\btemplates?\b|\brequests?\b|<code>|`|scoring run|stand-in|'
                  r'\bstandards? \d|\bclient \d{5}|\b[a-z]+_[a-z]+-\d{8}\b', re.I)
LEAK_SKIP = {'icon', 'url', 'key', 'build', 'builds', 'id', 'talents', 'raceIcon', 'classIcon', 'specIcon', 'boss', 'preset', 'specKey',
             'role', 'talentsString', 'bg', 'href'}


def leak_check(mode, sets):
    """Every string the page can show from the mode's datasets, against LEAK; the build exits with the hits."""
    hits = []

    def walk(v, where):
        if isinstance(v, dict):
            for k, x in v.items():
                if k not in LEAK_SKIP:
                    walk(x, f'{where}.{k}')
        elif isinstance(v, list):
            for i, x in enumerate(v):
                walk(x, f'{where}[{i}]')
        elif isinstance(v, str) and not v.startswith('data:'):
            for m in LEAK.finditer(re.sub(r'<a [^>]*>', '', v)):
                hits.append(f'{where}: "{m.group(0)}" in ...{v[max(0, m.start() - 50):m.end() + 50]}...')
    for name, rows in sets.items():
        walk(rows, name)
    if hits:
        sys.exit(f'{mode}: {len(hits)} internal words on the page (tankheal.py LEAK; fix the text in tankheal_text/{mode}.md):\n  ' + '\n  '.join(hits[:60]))


# ---------------------------------------------------------------- build

def mean(xs):
    return round(sum(xs) / len(xs), 2) if xs and all(x is not None for x in xs) else None


def mean_se(xs):
    return round(math.sqrt(sum(x * x for x in xs)) / len(xs), 2) if xs and all(x is not None for x in xs) else None


def dump(name, rows):
    os.makedirs(os.path.join(D, 'data'), exist_ok=True)
    json.dump(rows, open(os.path.join(D, 'data', name + '.json'), 'w'), separators=(',', ':'))
    print(f'{name}: {len(rows)} rows')


RACE_KEYS = {'tank': ('scoreMaxThreat', 'scoreBalanced', 'scoreMaxSurvival', 'safeDps', 'healingNeeded', 'dtps', 'tps'),
             'heal': ('carry', 'hps', 'ehps', 'overheal', 'manaLeft')}
SCORE_KEY = {'max_threat': 'scoreMaxThreat', 'balanced': 'scoreBalanced', 'max_survival': 'scoreMaxSurvival'}
# What each racial part does: the DPS report's text (build_races.py), and Night Elf Quickness, which only a tank feels.
PART_DODGE = ('Quickness', 'Dodge chance increased by 1%.', 'ability_racial_shadowmeld')


def race_results(src):
    """<src>/races.jsonl (race_sims) by preset: {race: {variant: row}}."""
    out = {}
    p = os.path.join(src, 'races.jsonl')
    if os.path.exists(p):
        for line in open(p):
            r = json.loads(line)
            out.setdefault(r['build'], {}).setdefault(r['race'], {})[r['variant']] = r
    return out


def split_parts(cls, race, v, key):
    """Each racial part's share of the race's gap over its base stats, from its race sims (v: variant -> row), and its
    error: a part alone is the race less "off"; two parts are split by their average contribution (Shapley), so they
    add up exactly; three or more get their leave-one-out loss, scaled so they add up to the whole gap (build_races.py
    does the same for DPS). Returns {part: (value, error, does something)}."""
    ps = bench_races.parts_for(cls, race)
    base, off = v['race'][key], v['off'][key]
    # a sim that moved no fight's number (the part does nothing here, e.g. dodge for a healer) is idle like a same one
    idle = lambda var: v[var].get('same') or (v[var][key] is not None and abs(v[var][key] - base) < 1e-12 and not v[var].get('pseBase'))
    if base is None or off is None:
        return {}
    one = lambda var: (v[var][key], v[var].get('pseBase') or 0.0)
    if len(ps) == 1:
        return {ps[0]: (base - off, v['off'].get('pseBase') or 0.0, not idle('off'))}
    live = [p for p in ps if not idle('off:' + p) and v['off:' + p][key] is not None]
    out = {p: (0.0, 0.0, False) for p in ps}
    if len(live) == 1:
        out[live[0]] = (base - off, v['off'].get('pseBase') or 0.0, True)
    elif len(live) == 2:
        a, c = live
        oa, oc = one('off:' + a), one('off:' + c)
        e = (oa[1] + oc[1] + (v['off'].get('pseBase') or 0.0)) / 2
        out[a] = (0.5 * ((base - oa[0]) + (oc[0] - off)), e, True)
        out[c] = (0.5 * ((base - oc[0]) + (oa[0] - off)), e, True)
    elif live:
        loo = {p: base - v['off:' + p][key] for p in live}
        scale = (base - off) / sum(loo.values()) if sum(loo.values()) else 0
        for p in live:
            out[p] = (loo[p] * scale, max(abs(scale) * (v['off:' + p].get('pseBase') or 0.0), v['off'].get('pseBase') or 0.0), True)
    return out


def race_rows(mode, H, bid, s, cls, req, cr, main, rs):
    """The preset's race rows and their racial parts. Every race comes from the race sims (one set of fights), moved by
    the preset race's gap to its main-run numbers, so the preset race's row is the main run's and the gaps between
    races are the paired ones. Each row's pse holds the paired error of its gap to every other race."""
    import build_races
    own = race_of(req)
    rname = lambda rc: H['race_name'](rc)
    keys = RACE_KEYS[mode]
    skey = next(v for k, v in SCORE_KEY.items() if s.endswith(k)) if mode == 'tank' else 'carry'
    ref = (rs.get(own) or {}).get('race')
    shift = {k: (main[k] - ref[k]) if ref and main.get(k) is not None and ref.get(k) is not None else 0.0 for k in keys}
    races, parts = [], []
    for rc in cr[cls]:
        v = rs.get(rc) or {}
        r = v.get('race')
        if rc != own and not r:
            continue
        nm = rname(rc)
        row = {'build': bid, 'race': nm, 'raceIcon': H['RACE_ICON'].get(nm), 'faction': 'Horde' if nm in ('Orc', 'Undead', 'Tauren', 'Troll', 'Windshaper') else 'Alliance',
               'isPreset': rc == own}
        for k in keys:
            row[k] = main.get(k) if rc == own or not r else (r[k] + shift[k] if r.get(k) is not None else None)
        po = (r or {}).get('pse', {}).get(own) if rc != own else None
        if mode == 'tank':
            for k in ('scoreMaxThreat', 'scoreBalanced', 'scoreMaxSurvival', 'safeDps', 'healingNeeded'):
                row[k + 'Se'] = main.get(k + 'Se') if rc == own else (math.hypot(main.get(k + 'Se') or 0, po or 0) if k == skey else main.get(k + 'Se'))
            row['healingNone'] = main['healingNone'] if rc == own or not r else r['none']
        else:
            row['role'] = heal_role(s)
            row['carrySe'] = po
        row['pse'] = {rname(o): round(e, 5) for o, e in ((r or {}).get('pse') or {}).items() if e is not None}
        row['iters'] = (r or {}).get('iters')
        races.append(row)
        if not r or 'off' not in v:
            continue
        lo = v['off'][skey] + shift[skey] if v['off'].get(skey) is not None else None
        if lo is None:
            continue
        parts.append({'build': bid, 'race': nm, 'part': 'base', 'label': build_races.PART_INFO['base'][0], 'v': round(lo, 4),
                      'se': round(v['off'].get('pseBase') or 0.0, 4), 'active': True, 'note': '', 'text': build_races.PART_INFO['base'][1]})
        for p, (d, e, live) in split_parts(cls, rc, v, skey).items():
            title, text, icon = PART_DODGE if p == 'dodge' else (*build_races.PART_INFO[p], build_races.PART_ICON.get(p))
            text = build_races.PART_SHORT.get(p, text)
            parts.append({'build': bid, 'race': nm, 'part': p, 'label': title, 'v': round(d, 4), 'se': round(e, 4), 'active': live,
                          'note': '' if live else 'no effect on this preset', 'text': text, 'icon': icon})
    return races, parts


def build():
    for mode in MODES:
        if not os.path.isdir(os.path.join(D, mode)):
            unpack([mode])
    import asm_data
    from assumptions_def import HEAL
    import md2html
    import rotation as rot_mod
    import tooltips as tt_mod
    H = dps_helpers()
    RR = {}
    for _race, _lst in json.load(open(os.path.join(TREE, 'ui/core/forever/racials.json'))).items():
        for _r in _lst:
            for _v in _r.get('variants', []):
                RR.setdefault(_v['spellId'], H['race_name'](_race.replace('Race', '')))
            RR.setdefault(_r['name'], H['race_name'](_race.replace('Race', '')))
    H['RACIAL_RACE'] = {k: {'SkyborneHighOrder': 'High Order', 'SkyborneWindshaper': 'Windshaper'}.get(v, v) for k, v in RR.items()}
    talents = rot_mod.load_talents(TREE)
    cr = class_races()
    common_keys = {r['key'] for r in common('tooltips')}
    fallbacks = []

    for mode in MODES:
        src = os.path.join(D, mode)
        if not os.path.isdir(src):
            print(f'{mode}: no outputs, skipped')
            continue
        info = json.load(open(os.path.join(src, 'run.json')))
        stems = presets(mode)
        reqs = {s: json.load(open(req_path(mode, s))) for s in stems}
        T = page_text(mode)
        need(mode, 'Presets', T['Presets'], [bid_of(s) for s in stems])
        need(mode, 'Bosses', T['Bosses'], BOSSES)
        builds, bosses, encs, races, rparts, gear, setup, rot = [], [], [], [], [], [], [], []
        outs_tt = []
        RS = race_results(src)
        per = {}  # stem -> boss -> row
        for bi, boss in enumerate(BOSSES):
            e = json.load(open(os.path.join(TREE, 'tools/encounters', boss + '.json')))
            if mode == 'tank':
                txt = open(os.path.join(src, boss + '.txt')).read()
                hd = tank_header(txt)
                for k, v in tank_rows(txt).items():
                    per.setdefault(k, {})[boss] = v
                h = e.get('healing') or {}
                encs.append({'boss': boss, 'name': e['name'], 'scored': boss in SCORED, 'order': bi, 'duration': e['encounter'].get('duration'),
                             'durationVariation': e['encounter'].get('durationVariation'), 'healers': h.get('healers'), 'healerHps': h.get('hpsEach'),
                             'healCast': h.get('castSeconds'), 'notes': T['Bosses'][boss]})
            else:
                er = {'boss': boss, 'name': e['name'], 'scored': True, 'order': bi}
                for role in ('tank', 'raid'):
                    txt = open(os.path.join(src, f'{role}_{boss}.txt')).read()
                    hd = heal_header(txt)
                    er[role + 'Duration'] = hd['duration']
                    er['notes'] = T['Bosses'][boss]
                    for k, v in heal_rows(txt).items():
                        per.setdefault(k, {})[boss] = dict(v, duration=hd['duration'])
                rh = e.get('raidHealing') or {}
                er.update({'raidHealers': rh.get('healers'), 'raidHealerClasses': rh.get('healerClasses'), 'raidHps': rh.get('raidHps'),
                           'tankHealers': (e.get('healing') or {}).get('healers')})
                encs.append(er)

        for s in stems:
            req = reqs[s]
            bid = bid_of(s)
            cls = player(req)['class'][5:]
            idn = identity(H, bid, cls, req)
            rows = per.get(s, {})
            if mode == 'tank':
                spec_k = s.rsplit('_', 2)[0] if s.endswith(('max_threat', 'max_survival')) else s.rsplit('_', 1)[0]
                preset = s[len(spec_k) + 1:]
                spec, specKey, label = TANK_SPEC[spec_k]
                sc = [rows[b] for b in SCORED if b in rows]
                agg = {}
                for k in ('scoreMaxThreat', 'scoreBalanced', 'scoreMaxSurvival', 'safeDps', 'healingNeeded'):
                    agg[k] = mean([r[k] for r in sc]) if len(sc) == 6 else None
                    agg[k + 'Se'] = mean_se([r[k + 'Se'] for r in sc]) if len(sc) == 6 else None
                for k in ('deathFloor', 'death', 'tps', 'dps', 'dtps', 'tmi'):
                    agg[k] = mean([r[k] for r in sc]) if len(sc) == 6 else None
                agg['healingNone'] = sum(1 for r in sc if r['healingNone'])
                seed = int(req.get('simOptions', {}).get('randomSeed', 0))
                builds.append({'id': bid, 'class': cls, 'spec': spec, 'specKey': specKey, 'preset': preset, 'presetName': PRESET_NAME[preset],
                               'weight': round(WEIGHT[preset], 4), 'label': label, **idn, 'summary': ' '.join(T['Presets'][bid].split()), **agg,
                               'iterations': info['iterations'], 'seed': seed, 'bosses': len(sc)})
                for b in BOSSES:
                    if b in rows:
                        bosses.append({'build': bid, 'boss': b, 'scored': b in SCORED, **rows[b]})
            else:
                role = heal_role(s)
                spec = HEAL_SPEC[s[:-5] if role == 'raid' else s]
                sc = [rows[b] for b in BOSSES if b in rows]
                agg = {k: mean([r[k] for r in sc]) for k in ('carry', 'hps', 'ehps', 'overheal', 'ehPerMana', 'manaLeft', 'oomAt', 'tankDied', 'otherDied')}
                agg['carry'] = round(sum(r['carry'] for r in sc) / len(sc), 4) if len(sc) == 7 else None
                agg['oomInFight'] = sum(1 for r in sc if r['oomInFight'])
                builds.append({'id': bid, 'class': cls, 'spec': spec, 'role': role, 'roleName': ROLE_NAME[role], 'label': f'{spec} {cls}',
                               **idn, 'summary': ' '.join(T['Presets'][bid].split()), **agg, 'bosses': len(sc), 'iterations': info['iterations'], 'seed': HEAL_SEED})
                for b in BOSSES:
                    if b in rows:
                        bosses.append({'build': bid, 'boss': b, 'role': role, **rows[b]})
            # races: every race from the race sims, put on the main run's numbers (race_rows)
            rr_, rp_ = race_rows(mode, H, bid, s, cls, req, cr, agg, RS.get(s, {}))
            races += rr_
            rparts += rp_
            g, st = gear_setup(H, bid, req)
            gear += g
            setup += st
            rr_, fb = rotation_rows(H, rot_mod, talents, bid, cls, req, T['Rotation'])
            rot += rr_
            fallbacks += fb
            outs_tt.append({'id': bid, 'meta': {'class': cls}, 'races': [{'race': rc} for rc in cr[cls]],
                            'setup': {'equipment': player(req).get('equipment', {}).get('items', [])}})

        # ranks
        if mode == 'tank':
            for b in builds:
                key = {'max_threat': 'scoreMaxThreat', 'balanced': 'scoreBalanced', 'max_survival': 'scoreMaxSurvival'}[b['preset']]
                peers = [x for x in builds if x['preset'] == b['preset'] and x[key] is not None]
                b['rank'] = 1 + sum(1 for x in peers if x[key] > (b[key] if b[key] is not None else -1e18)) if b[key] is not None else None
            builds.sort(key=lambda b: (CLASS_ORDER.index(b['class']), list(WEIGHT).index(b['preset'])))
        else:
            for b in builds:
                peers = [x for x in builds if x['role'] == b['role'] and x['carry'] is not None]
                b['rank'] = 1 + sum(1 for x in peers if x['carry'] > b['carry']) if b['carry'] is not None else None
            builds.sort(key=lambda b: (b['role'] != 'tank', CLASS_ORDER.index(b['class']), b['spec']))

        # tooltips the common dataset lacks
        abil = [{'build': r['build'], 'key': r['key']} for r in rot if r['key']]
        tt_rows, _miss, _rep = tt_mod.build(TREE, {k: v for k, v in H['names'].items()}, gear, setup, outs_tt, abil, [], {}, builds, H['item_names'],
                                             H['enchant_for'], H['TREE_JSON'],
                                             os.path.join(TREE, 'ui/core/forever/racials.json'), H['SLOTS'])
        tt_rows = [r for r in tt_rows if r['key'] not in common_keys]

        # notes and method: the written text (tankheal_text/<mode>.md)
        notes = notes_rows(md2html, T, builds)
        method = method_rows(md2html, mode, T, encs)
        p = 'tank_' if mode == 'tank' else 'heal_'
        # the Healing page's assumptions (assumptions_def.HEAL): one-value rows, the DPS panel's shape
        asm = asm_data.asm_rows(HEAL, {b['id'] for b in builds}) if mode == 'heal' else None
        leak_check(mode, {p + 'builds': builds, p + 'bosses': bosses, p + 'encounters': encs, p + 'races': races, p + 'race_breakdown': rparts, p + 'gear': gear,
                          p + 'rotation': rot, p + 'setup': setup, p + 'tooltips': tt_rows, p + 'notes': notes, p + 'method': method,
                          **({p + 'assumptions': asm} if asm else {})})
        dump(p + 'builds', builds)
        dump(p + 'bosses', bosses)
        dump(p + 'encounters', encs)
        dump(p + 'races', races)
        dump(p + 'race_breakdown', rparts)
        dump(p + 'gear', gear)
        dump(p + 'rotation', rot)
        dump(p + 'setup', setup)
        dump(p + 'tooltips', tt_rows)
        dump(p + 'notes', notes)
        dump(p + 'method', method)
        if asm:
            dump(p + 'assumptions', asm)
        # names (with their colour) and icons the common datasets lack
        known = {r['key'] for r in common('names')}
        mine = sorted({r['key'] for r in rot if r['key']} | {r['key'] for r in tt_rows if r['kind'] in ('spell', 'item')})
        icons = icon_rows(H, {r.get('icon') for r in gear + setup + rot + tt_rows + races + rparts} | {H['names'][k]['icon'] for k in mine if k in H['names']} |
                          {b[k] for b in builds for k in ('classIcon', 'specIcon', 'raceIcon')}, {b['specIcon'] for b in builds})
        dump(p + 'icons', icons)
        nrows = []
        for k in mine:
            if k in known or k not in H['names']:
                continue
            n = dict(H['names'][k])
            if n.get('icon'):
                path = os.path.join(D, 'icons', n['icon'] + '.jpg')
                if not os.path.exists(path):
                    H['fetch_url'](f"https://wow.zamimg.com/images/wow/icons/medium/{n['icon']}.jpg", path)
            n['color'] = H['icon_color'](n['icon']) if n.get('icon') else None
            nrows.append(n)
        dump(p + 'names', nrows)
    if fallbacks:
        print('rotation conditions shown in compact form:', sorted(set(fallbacks)))


def fresh(clear=False):
    """The modes simmed since the last publish: <mode> when the run simmed every preset, <mode>:<stem,...> when it kept
    some presets' race rows from the pack (stale.mjs --record-bench records only the presets named)."""
    got = [m for m in MODES if os.path.exists(os.path.join(D, m, '.fresh'))]
    if clear:
        for m in got:
            os.remove(os.path.join(D, m, '.fresh'))
    else:
        out = []
        for m in got:
            lines = open(os.path.join(D, m, '.fresh')).read().split('\n')
            out.append(f'{m}:{lines[1]}' if len(lines) > 1 and lines[1].strip() else m)
        print(' '.join(out))


if __name__ == '__main__':
    cmd, args = (sys.argv[1], sys.argv[2:]) if len(sys.argv) > 1 else ('', [])
    if cmd == 'run':
        modes, only = [a for a in args if a in MODES], None
        part = None
        if {'--part', '--main-only', '--carries'} & set(args):  # a worker's share of one mode (parallel.py), for merge
            if len(modes) != 1 or '--stale' in args or '--no-races' in args:
                sys.exit('a part is one mode (run tank|heal --part), with --presets, --main-only or --carries, not --stale or --no-races')
            want = set(args[args.index('--presets') + 1].split(',')) if '--presets' in args else None
            if want and want - set(presets(modes[0])):
                sys.exit(f'no such {modes[0]} preset: {" ".join(sorted(want - set(presets(modes[0]))))}')
            part = {'presets': want, 'main_only': '--main-only' in args,
                    'carries': os.path.abspath(args[args.index('--carries') + 1]) if '--carries' in args else None}
            if part['main_only'] and (want or part['carries']):
                sys.exit('--main-only runs the main bench alone: no --presets or --carries')
            run(modes, resume='--resume' in args, part=part)
            sys.exit(0)
        if '--stale' in args:
            only = stale_presets()
        elif '--presets' in args:
            want = set(args[args.index('--presets') + 1].split(','))
            only = {m: want & set(presets(m)) for m in MODES}
            if want - set().union(*only.values()):
                sys.exit(f'no such preset: {" ".join(sorted(want - set().union(*only.values())))}')
        if only is not None:
            # a mode with no preset to re-sim is left as it is
            modes = [m for m in (modes or MODES) if only.get(m)]
            print(f'presets to re-sim: {only}, modes: {" ".join(modes) or "none"}', flush=True)
        elif not modes:
            modes = list(MODES)
        run(modes, races='--no-races' not in args, resume='--resume' in args, only=only)
    elif cmd == 'part':
        if len(args) != 2 or args[0] not in MODES:
            sys.exit('tankheal.py part <tank|heal> <dest.tgz>')
        part_pack(args[0], args[1])
    elif cmd == 'merge':
        srcs = [a for a in args[1:] if a != '--keep-rest']
        if not args or args[0] not in MODES or not srcs:
            sys.exit('tankheal.py merge <tank|heal> [--keep-rest] <part.tgz | folder> ...')
        merge(args[0], srcs, keep_rest='--keep-rest' in args)
    elif cmd == 'build':
        build()
    elif cmd == 'bins':
        bins()
    elif cmd == 'pack':
        pack()
    elif cmd == 'unpack':
        unpack()
    elif cmd == 'fresh':
        fresh('--clear' in args)
    elif cmd == 'probe':
        sub = args[0] if args and args[0] in ('check', 'audit', 'passed') else ''
        rest = args[1:] if sub else args
        modes = [a for a in rest if a in MODES] or list(MODES)
        want = set(rest[rest.index('--presets') + 1].split(',')) if '--presets' in rest else None
        if want and want - set().union(*(presets(m) for m in MODES)):
            sys.exit(f'no such preset: {" ".join(sorted(want - set().union(*(presets(m) for m in MODES))))}')
        if sub == 'check':
            probe_check(modes, want, '--json' in rest)
        elif sub == 'audit':
            probe_audit()
        elif sub == 'passed':
            print(' '.join(probe_passed()))
        else:
            probe_write(modes, want, '--stale-ok' in rest)
    elif cmd == '_probe1':  # one preset's probe, in the process probe_presets starts (probe_env)
        probe_one(*args)
    else:
        print(__doc__)
        sys.exit(1)
