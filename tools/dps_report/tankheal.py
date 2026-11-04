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
# one line a tank preset, from its class doc's Presets section (docs/tank_<class>.md)
SUMMARY = {
    'warrior__protection_max_threat': 'Improved Heroic Strike over Deflection; no Thunder Clap, Heroic Strike from 30 Rage, Revenge first.',
    'warrior__protection_balanced': 'Deflection 5 and 31 Protection with Shield Slam; Shield Block whenever its buff is down, Thunder Clap kept up.',
    'warrior__protection_max_survival': "Balanced's build and rotation: no talent or rotation change moved it past the noise.",
    'paladin__protection_max_threat': "Improved Seals 5 in place of Anticipation and a point of Toughness; the template's rotation.",
    'paladin__protection_balanced': 'Seal of Fury, Holy Shield first and before the pull, Consecration above 60% mana.',
    'paladin__protection_max_survival': "Balanced's build and rotation: its own talent search found nothing better.",
    'druid__bear_max_threat': 'Balanced with Maul from 20 Rage.',
    'druid__bear_balanced': 'Lacerate to 5 stacks, Primal Bite on cooldown, Maul from 30 Rage, Barkskin as the filler; no Furor.',
    'druid__bear_max_survival': 'Balanced with Frenzied Regeneration at 100 Rage and Enrage 3 s before the pull.',
    'shaman__warden_max_threat': 'Elemental Weapons 3 over Shamanistic Focus, the attack power trinket and Mana Spring Totem.',
    'shaman__warden_balanced': 'Tidal Focus 5 and Improved Reincarnation 2; Stoneskin, Lightning Shield and Earth Shock on cooldown.',
    'shaman__warden_max_survival': "Balanced's talents and rotation with the armor on-use trinket on a timer.",
}
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
                     'enchant': H['enchant_names'].get(it.get('enchant'), '') if it.get('enchant') else '', 'itemId': it['id'],
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


def rotation_rows(H, rot_mod, talents, bid, cls, req):
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
        r['notes'] = rot_mod.link_talents(H['_clean_note'](r['notes']), cls, talents, r['name'])
        r['why'] = r['notes']  # the bench presets have no rotation_text.json lines: their APL notes are the why
    return rows, Hm.fallbacks


# ---------------------------------------------------------------- doc text

NEUTRAL = [(r"\bThe user's rulings\b", 'Rulings'), (r"\bthe user's rulings?\b", 'the rulings'), (r"\bthe user's\b", "the project's"),
           (r'\bThe user:', 'The project:'), (r'\bthe user\b', 'the project'), (r'\bThe user\b', 'The project'),
           (r'\s*Say the word and it changes\.', '')]


def neutral(t):
    for a, b in NEUTRAL:
        t = re.sub(a, b, t)
    return t


def doc(name):
    return open(os.path.join(ROOT, 'docs', name)).read()


def numbered_items(md):
    """A section's top-level numbered list: [(title, markdown)], the title its bold lead."""
    items, cur = [], None
    for ln in md.split('\n'):
        m = re.match(r'^\d+\.\s+(.*)$', ln)
        if m:
            cur = [m.group(1)]
            items.append(cur)
        elif cur is not None and (ln.startswith(' ') or not ln.strip()):
            cur.append(ln.strip())
        elif cur is not None:
            cur = None
    out = []
    for it in items:
        text = '\n'.join(it).strip()
        t = re.match(r'^\*\*(.+?)\*\*', text)
        title = re.sub(r'[.:]$', '', t.group(1)) if t else re.split(r'(?<=[.:])\s', text, 1)[0]
        out.append((re.sub(r'`', '', title), text))
    return out


def notes_rows(md2html, cls, docname, preset_sections, preset_builds):
    md = doc(docname)
    secs = md2html.sections(md)
    rows = []
    for title, body in secs:
        if any(title.startswith(p) for p in preset_sections):
            rows.append({'build': None, 'class': cls, 'kind': 'preset', 'title': neutral(re.sub(r'\s*\(.*\)$', '', title)),
                         'html': md2html.convert(neutral(body.strip())), 'builds': preset_builds})
        elif title.startswith('Calls for the user'):
            items = numbered_items(body)
            if not items:
                rows.append({'build': None, 'class': cls, 'kind': 'call', 'title': 'Open questions', 'html': md2html.convert(neutral(body.strip()))})
            for t, txt in items:
                rows.append({'build': None, 'class': cls, 'kind': 'call', 'title': neutral(t), 'html': md2html.convert(neutral(txt))})
    return rows


def method_rows(md2html, docname, wanted):
    secs = dict(md2html.sections(doc(docname)))
    subs = dict(md2html.sections(doc(docname), 3))
    rows = []
    for i, (title, shown) in enumerate(wanted):
        body = secs.get(title, subs.get(title))
        if body is None:
            print(f'WARNING {docname}: no section "{title}"')
            continue
        rows.append({'order': i, 'title': shown, 'html': md2html.convert(neutral(body.strip()))})
    return rows


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
                             'healCast': h.get('castSeconds'), 'notes': hd.get('notes'), 'standIn': hd.get('standIn'),
                             'healingLine': hd.get('healingLine'), 'objective': hd.get('objective'), 'healingNotes': h.get('notes')})
            else:
                er = {'boss': boss, 'name': e['name'], 'scored': True, 'order': bi}
                for role in ('tank', 'raid'):
                    txt = open(os.path.join(src, f'{role}_{boss}.txt')).read()
                    hd = heal_header(txt)
                    er[role + 'Duration'] = hd['duration']
                    er['notes'] = hd['notes']
                    er['setup'] = hd['setup']
                    for k, v in heal_rows(txt).items():
                        per.setdefault(k, {})[boss] = dict(v, duration=hd['duration'])
                    rp = os.path.join(src, 'races', f'{role}_{boss}.txt')
                    if os.path.exists(rp):
                        for k, v in heal_rows(open(rp).read()).items():
                            rper.setdefault(k, {})[boss] = v
                rh = e.get('raidHealing') or {}
                er.update({'raidHealers': rh.get('healers'), 'raidHealerClasses': rh.get('healerClasses'), 'raidHps': rh.get('raidHps'),
                           'tankHealers': (e.get('healing') or {}).get('healers'), 'healingNotes': rh.get('notes')})
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
                               'weight': round(WEIGHT[preset], 4), 'label': label, **idn, 'summary': SUMMARY.get(bid, ''), **agg,
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
                               **idn, **agg, 'bosses': len(sc), 'iterations': info['iterations'], 'seed': HEAL_SEED})
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
            rr_, fb = rotation_rows(H, rot_mod, talents, bid, cls, req)
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
                                             {e['effectId']: e for e in H['db']['enchants']}, H['TREE_JSON'],
                                             os.path.join(TREE, 'ui/core/forever/racials.json'), H['SLOTS'])
        tt_rows = [r for r in tt_rows if r['key'] not in common_keys]

        # notes and method
        md2 = md2html
        notes = []
        for cls in CLASS_ORDER:
            ids = [b['id'] for b in builds if b['class'] == cls]
            if not ids:
                continue
            if mode == 'tank':
                notes += notes_rows(md2, cls, f'tank_{cls.lower()}.md', ('Presets',), ids)
            else:
                pre = {'Priest': ('The presets',), 'Paladin': ('The rotations', 'The numbers'), 'Druid': ('Tuning and the presets',),
                       'Shaman': ('Presets',)}[cls]
                notes += notes_rows(md2, cls, f'healing_{cls.lower()}.md', pre, ids)
        if mode == 'tank':
            method = method_rows(md2, 'tank_sim.md', [("The user's rulings (2026-11-02)", 'Rulings'), ('Encounters', 'Encounters'),
                                                       ('The bench', 'The bench'), ('The stand-in healer', 'The stand-in healer'),
                                                       ('Scoring survival', 'Scoring survival'), ("The optimizer's objective", 'The score')])
        else:
            method = method_rows(md2, 'healing_sim.md', [('What a healer scores', 'The score'), ('The score in the bench (2026-11-02)', 'The bench and the roles'),
                                                          ('What deals the damage', 'Damage'), ('Raid damage from the logs (2026-11-02)', 'Raid damage from the logs'),
                                                          ('Who heals whom', 'Who heals whom'), ("The user's rulings (2026-11-02)", 'Rulings')])
            method.append({'order': len(method), 'title': 'The mean', 'html': md2.convert(
                f'A healer\'s score is its **damage carried** (`healbench -carry`), {info["iterations"]:,} iterations a sim at each boss file\'s own '
                f'length, on the bench\'s seed ({HEAL_SEED}). Its mean is over the seven boss files, Ragnaros included, as the healing docs '
                'average them; tank healers (one healer, the tank\'s damage, fails at 2% tank deaths) and raid healers (the raid\'s damage, '
                'fails at 5% deaths) are ranked apart.')})
        p = 'tank_' if mode == 'tank' else 'heal_'
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
