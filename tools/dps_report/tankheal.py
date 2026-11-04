#!/usr/bin/env python3
"""tankheal.py: the report's Tanking and Healing modes, from tools/tankbench and tools/healbench.

  tankheal.py run [tank|heal ...] [--no-races] [--resume]
                                                 the sims, into $DASH_DIR/tank and $DASH_DIR/heal (the benches' text,
                                                 log lines dropped, and run.json); marks the mode fresh for publish.
                                                 --resume keeps the bench calls a stopped run finished (.progress.json)
  tankheal.py build                              the tank_* and heal_* datasets into $DASH_DIR/data (unpacks the raw
                                                 packs first when $DASH_DIR/tank or heal is missing)
  tankheal.py pack                               $DASH_DIR/tank and heal into raw/bench/<mode>.tgz (deterministic)
  tankheal.py unpack                             the packs back into $DASH_DIR/tank and heal
  tankheal.py fresh [--clear]                    the modes run since the last publish (stale.mjs --record-bench them)

Tanking: every preset request in tools/tankbench/req (warrior_protection.json is no preset) on the seven boss files
of tools/encounters, `tankbench -iters 1000 -encounter <boss>`, the requests' own seed (simOptions.randomSeed).
The score is docs/tank_sim.md's objective (2 x SafeDPS - w x 900 x healing needed / healer HPS, w = 1/3, 1, 3) and a
preset's score is the mean of the six Anniversary bosses; Ragnaros is shown, not scored.
Healing: every request in tools/healbench/req, `healbench -carry -role tank|raid -iters 1000 -encounter <boss>`
(files ending _raid are raid healers), at each file's own length, the bench's seed (101): the score is the damage
carried (docs/healing_sim.md "What a healer scores" and "The score in the bench"), its mean over the seven boss files
(docs/healing_sim.md, H3: "The shaman's raid mean over these 7 files").
Races: each preset again at every race its class can be (run_build.CLASS_RACES), the request's race changed.
Text: what players read (preset lines, notes, boss lines, Method, rotation rewording) is written in tankheal_text/
<mode>.md, not taken from the docs; the build stops when a string the page shows has a tool or file word (LEAK).

The binaries (tankbench, healbench) come from $TANKHEAL_BIN, else $DASH_BIN, else $DASH_DIR/bin, and run with the
merged tree ($DASH_TREE, else this checkout when it has no w-dash) as their cwd; README.md has the build recipe.
"""
import ast, glob, io, json, math, os, re, shutil, subprocess, sys, tarfile, tempfile, time

HERE = os.path.dirname(os.path.abspath(__file__))
sys.path.insert(0, HERE)
import dashpaths  # noqa: E402

ROOT = os.path.abspath(os.path.join(HERE, '..', '..'))
D = dashpaths.D
TREE = dashpaths.TREE if os.path.isdir(dashpaths.TREE) else ROOT
os.environ['DASH_TREE'] = TREE  # for the modules below that read dashpaths
dashpaths.TREE = TREE
BIN = os.path.abspath(os.environ.get('TANKHEAL_BIN') or dashpaths.BIN)
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
    r = subprocess.run([exe] + args + files, cwd=TREE, capture_output=True, text=True)
    if r.returncode:
        sys.exit(f'{exe} failed ({r.returncode}):\n{r.stderr[-2000:]}')
    text = ''.join(l + '\n' for l in (r.stdout + r.stderr).split('\n') if l.strip() and not LOG_LINE.match(l))
    os.makedirs(os.path.dirname(out), exist_ok=True)
    open(out, 'w').write(text)
    return time.time() - t


def race_files(mode, stems, tmp):
    """A copy of each preset at every other race its class can be: <tmp>/<stem>@<Race>.json."""
    cr, files = class_races(), {}
    for s in stems:
        req = json.load(open(req_path(mode, s)))
        cls = player(req)['class'][5:]
        for rc in cr[cls]:
            if rc == race_of(req):
                continue
            r2 = json.loads(json.dumps(req))
            player(r2)['race'] = 'Race' + rc
            p = os.path.join(tmp, f'{s}@{rc}.json')
            json.dump(r2, open(p, 'w'))
            files.setdefault(heal_role(s) if mode == 'heal' else 'tank', []).append(p)
    return files


def run(modes, races=True, resume=False):
    commit = subprocess.run(['git', 'rev-parse', 'HEAD'], cwd=TREE, capture_output=True, text=True).stdout.strip()
    for mode in modes:
        out = os.path.join(D, mode)
        prog = os.path.join(out, '.progress.json')
        done = json.load(open(prog)) if resume and os.path.exists(prog) else {}
        if done.get('commit') != commit or os.path.exists(os.path.join(out, 'run.json')):
            done = {}  # another commit's outputs, or a finished run: start over
        if not done:
            shutil.rmtree(out, ignore_errors=True)
        os.makedirs(os.path.join(out, 'races'), exist_ok=True)
        stems = presets(mode)
        secs, t0 = dict(done.get('secs', {})), time.time() - done.get('seconds', 0)

        def note(k, v):
            secs[k] = v
            json.dump({'commit': commit, 'secs': secs, 'seconds': round(time.time() - t0, 1)}, open(prog, 'w'))
        groups = {'tank': stems} if mode == 'tank' else {r: [s for s in stems if heal_role(s) == r] for r in ('tank', 'raid')}
        tmp = tempfile.mkdtemp(prefix=f'{mode}_races_')
        rgroups = race_files(mode, stems, tmp) if races else {}
        for boss in BOSSES:
            enc = ['-encounter', f'tools/encounters/{boss}.json', '-iters', str(ITERS)]
            for g, st in groups.items():
                files = [req_path(mode, s) for s in st]
                name = boss if mode == 'tank' else f'{g}_{boss}'
                args = enc if mode == 'tank' else ['-carry', '-role', g] + enc
                if name not in secs:
                    note(name, round(bench(mode, args, files, os.path.join(out, name + '.txt')), 1))
                    print(f'{mode} {name}: {secs[name]} s', flush=True)
                k = 'races/' + name
                if g in rgroups and k not in secs:
                    note(k, round(bench(mode, args, rgroups[g], os.path.join(out, 'races', name + '.txt')), 1))
                    print(f'{mode} {k}: {secs[k]} s ({len(rgroups[g])} race copies)', flush=True)
        shutil.rmtree(tmp, ignore_errors=True)
        info = {'iterations': ITERS, 'seed': 'the requests\' simOptions.randomSeed' if mode == 'tank' else HEAL_SEED,
                'commit': commit, 'at': time.strftime('%Y-%m-%dT%H:%M:%SZ', time.gmtime()), 'seconds': round(time.time() - t0, 1),
                'runSeconds': secs, 'races': races}
        json.dump(info, open(os.path.join(out, 'run.json'), 'w'), indent=1)
        os.remove(prog)
        open(os.path.join(out, '.fresh'), 'w').write(commit + '\n')
        print(f'{mode}: {info["seconds"]} s')


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
        builds, bosses, encs, races, gear, setup, rot = [], [], [], [], [], [], []
        outs_tt = []
        per = {}  # stem -> boss -> row
        rper = {}  # stem@Race -> boss -> row
        for bi, boss in enumerate(BOSSES):
            e = json.load(open(os.path.join(TREE, 'tools/encounters', boss + '.json')))
            if mode == 'tank':
                txt = open(os.path.join(src, boss + '.txt')).read()
                hd = tank_header(txt)
                for k, v in tank_rows(txt).items():
                    per.setdefault(k, {})[boss] = v
                rp = os.path.join(src, 'races', boss + '.txt')
                if os.path.exists(rp):
                    for k, v in tank_rows(open(rp).read()).items():
                        rper.setdefault(k, {})[boss] = v
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
                    rp = os.path.join(src, 'races', f'{role}_{boss}.txt')
                    if os.path.exists(rp):
                        for k, v in heal_rows(open(rp).read()).items():
                            rper.setdefault(k, {})[boss] = v
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
            # races: the preset's own race from the main run, the others from the race copies
            for rc in cr[cls]:
                rname = H['race_name'](rc)
                own = rc == race_of(req)
                rr = rows if own else rper.get(f'{s}@{rc}', {})
                if not rr:
                    continue
                row = {'build': bid, 'race': rname, 'raceIcon': H['RACE_ICON'].get(rname), 'faction': 'Horde' if rname in ('Orc', 'Undead', 'Tauren', 'Troll', 'Windshaper') else 'Alliance',
                       'isPreset': own}
                if mode == 'tank':
                    sc = [rr[b] for b in SCORED if b in rr]
                    for k in ('scoreMaxThreat', 'scoreBalanced', 'scoreMaxSurvival', 'safeDps', 'healingNeeded'):
                        row[k] = mean([r[k] for r in sc]) if len(sc) == 6 else None
                        row[k + 'Se'] = mean_se([r[k + 'Se'] for r in sc]) if len(sc) == 6 else None
                    row['dtps'] = mean([r['dtps'] for r in sc]) if len(sc) == 6 else None
                    row['tps'] = mean([r['tps'] for r in sc]) if len(sc) == 6 else None
                    row['healingNone'] = sum(1 for r in sc if r['healingNone'])
                else:
                    sc = [rr[b] for b in BOSSES if b in rr]
                    row['role'] = heal_role(s)
                    row['carry'] = round(sum(r['carry'] for r in sc) / len(sc), 4) if len(sc) == 7 else None
                    for k in ('hps', 'ehps', 'overheal', 'ehPerMana', 'manaLeft'):
                        row[k] = mean([r[k] for r in sc]) if len(sc) == 7 else None
                races.append(row)
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
        leak_check(mode, {p + 'builds': builds, p + 'bosses': bosses, p + 'encounters': encs, p + 'races': races, p + 'gear': gear,
                          p + 'rotation': rot, p + 'setup': setup, p + 'tooltips': tt_rows, p + 'notes': notes, p + 'method': method,
                          **({p + 'assumptions': asm} if asm else {})})
        dump(p + 'builds', builds)
        dump(p + 'bosses', bosses)
        dump(p + 'encounters', encs)
        dump(p + 'races', races)
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
        icons = icon_rows(H, {r.get('icon') for r in gear + setup + rot + tt_rows + races} | {H['names'][k]['icon'] for k in mine if k in H['names']} |
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
    got = [m for m in MODES if os.path.exists(os.path.join(D, m, '.fresh'))]
    if clear:
        for m in got:
            os.remove(os.path.join(D, m, '.fresh'))
    else:
        print(' '.join(got))


if __name__ == '__main__':
    cmd, args = (sys.argv[1], sys.argv[2:]) if len(sys.argv) > 1 else ('', [])
    if cmd == 'run':
        run([a for a in args if a in MODES] or list(MODES), races='--no-races' not in args, resume='--resume' in args)
    elif cmd == 'build':
        build()
    elif cmd == 'pack':
        pack()
    elif cmd == 'unpack':
        unpack()
    elif cmd == 'fresh':
        fresh('--clear' in args)
    else:
        print(__doc__)
        sys.exit(1)
