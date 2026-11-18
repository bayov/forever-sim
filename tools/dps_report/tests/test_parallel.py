"""Tests for a mode split over worker sessions: tankheal.py's parts and merge, and parallel.py's planner and transport.

Run: python3 -m unittest discover -s tools/dps_report/tests
No sims run: the main bench and the race sims are stand-ins whose outputs depend only on their inputs, as the real ones
do (fixed seeds and splits), and the race sims finish in a shuffled order, as two at a time do.
"""
import contextlib, hashlib, io, json, os, random, shutil, subprocess, sys, tempfile, time, unittest

HERE = os.path.dirname(os.path.abspath(__file__))
TOOLS = os.path.dirname(HERE)
tankheal = bench_races = build_bins = rawstore = parallel = None


def setUpModule():
    # imported when the tests run, not when discover loads this file: other test files set DASH_DIR and the like before
    # their imports, and the pipeline's modules read them once, at import
    global tankheal, bench_races, build_bins, rawstore, parallel
    sys.path.insert(0, TOOLS)
    import tankheal, bench_races, build_bins, rawstore, parallel  # noqa: E401

COLS = ['Healer', 'Carry', 'Death floor', 'HPS', 'EHPS', 'Overheal', 'EH per mana', 'Mana left', 'OOM at', 'Tank DTPS',
        'Raid DTPS', 'Tank died', 'Other died', 'First death', 'Tank scares', 'Raid scares']
STEMS = {'paladin_holy': ('Paladin', 'Human'), 'priest_holy': ('Priest', 'Dwarf'), 'priest_holy_raid': ('Priest', 'Human'),
         'shaman_restoration_raid': ('Shaman', 'Orc')}
STAMP = {'inputHash': 'h1', 'go': 'go1.24', 'binaries': {'healbench': 'b1'}}


def num(*parts):
    return int(hashlib.sha256(':'.join(map(str, parts)).encode()).hexdigest()[:6], 16)


def fake_bench(calls):
    def bench(mode, args, files, out):
        role, boss = args[args.index('-role') + 1], os.path.basename(args[args.index('-encounter') + 1])[:-5]
        rows = ''.join('| ' + ' | '.join([os.path.basename(f)[:-5], f'{1 + num(f, boss) % 300 / 1000:.3f}', '0.5% tank, 1.0% other']
                                          + [str(num(f, boss, c) % 900) for c in COLS[3:8]] + ['(none)'] + ['1'] * 7) + ' |\n' for f in files)
        os.makedirs(os.path.dirname(out), exist_ok=True)
        open(out, 'w').write(f'{boss}, role {role} ×1.00, 180 s, 1000 iterations, a sim. setup\nnotes\n| ' + ' | '.join(COLS)
                             + ' |\n|' + '---|' * len(COLS) + '\n' + rows)
        calls.append((role, boss))
        return 0.0
    return bench


def fake_preset(mode, B, stem, cls, own, races, bosses, role=None, start=None, w=None, each=None, log=print):
    time.sleep(random.random() * 0.03)  # finish in another order each time, as the real pool does
    return [{'build': stem, 'race': r, 'variant': v, 'same': False, 'carry': num(stem, r, v, json.dumps(start, sort_keys=True)) % 10000}
            for r in races for v in ('race', 'off')]


class Sandbox(unittest.TestCase):
    """A tree with four healing presets, fake benches and binaries, and a $DASH_DIR per session."""

    def setUp(self):
        self.tmp = tempfile.mkdtemp(prefix='dps_parallel_test_')
        tree = os.path.join(self.tmp, 'tree')
        for s, (cls, race) in STEMS.items():
            p = os.path.join(tree, tankheal.HEAL_REQ, s + '.json')
            os.makedirs(os.path.dirname(p), exist_ok=True)
            json.dump({'raid': {'parties': [{'players': [{'class': 'Class' + cls, 'race': 'Race' + race}]}]}}, open(p, 'w'))
        os.makedirs(os.path.join(self.tmp, 'bin'))
        open(os.path.join(self.tmp, 'bin', 'healbench'), 'w').write('fake')
        self.calls, self.stamp, self.inputs = [], dict(STAMP), 'fp1'
        self.saved = [(tankheal, k, getattr(tankheal, k)) for k in ('D', 'TREE', 'BIN', 'bench', 'bench_inputs', 'pack_files', 'SPLITS')]
        self.saved += [(bench_races, 'preset', bench_races.preset), (build_bins, 'guard', build_bins.guard), (build_bins, 'summary', build_bins.summary)]
        tankheal.TREE, tankheal.BIN, tankheal.bench = tree, os.path.join(self.tmp, 'bin'), fake_bench(self.calls)
        tankheal.bench_inputs = lambda mode, commit: self.inputs
        tankheal.pack_files = lambda mode: None
        bench_races.preset = fake_preset
        build_bins.guard = lambda *a, **k: None
        build_bins.summary = lambda bin_dir, names: dict(self.stamp)
        self.env = os.environ.pop('TANKHEAL_PROBE', None)

    def tearDown(self):
        for mod, k, v in self.saved:
            setattr(mod, k, v)
        if self.env is not None:
            os.environ['TANKHEAL_PROBE'] = self.env
        shutil.rmtree(self.tmp, ignore_errors=True)

    def session(self, name):
        tankheal.D = os.path.join(self.tmp, name)
        os.makedirs(tankheal.D, exist_ok=True)
        return tankheal.D

    def quiet(self, f, *a, **k):
        with contextlib.redirect_stdout(io.StringIO()), contextlib.redirect_stderr(io.StringIO()):
            return f(*a, **k)

    def part(self, name, presets=None, main_only=False, carries=None):
        """One worker's part, packed: its .tgz path."""
        self.session(name)
        self.quiet(tankheal.run, ['heal'], part={'presets': presets, 'main_only': main_only, 'carries': carries})
        dest = os.path.join(self.tmp, name + '.tgz')
        self.quiet(tankheal.part_pack, 'heal', dest)
        return dest

    def split_run(self):
        main = self.part('main', main_only=True)
        n = len(self.calls)
        a = self.part('w1', presets={'priest_holy_raid'}, carries=main)
        b = self.part('w2', presets={'paladin_holy', 'priest_holy', 'shaman_restoration_raid'}, carries=main)
        self.assertEqual(len(self.calls), n, 'a part with carries ran the main bench')
        return main, a, b

    def files(self, d):
        return {os.path.relpath(p, d): open(p, 'rb').read() for p in (os.path.join(r, f) for r, _, fs in os.walk(d) for f in fs)}

    def rewrite(self, tgz, **info):
        """A copy of a part with part.json fields changed."""
        old, files = tankheal.read_part(tgz)
        files['part.json'] = json.dumps(dict(old, **info)).encode()
        dest = tgz[:-4] + '-x.tgz'
        open(dest, 'wb').write(rawstore.tgz(files))
        return dest


class MergeEqualsOneRun(Sandbox):
    def test_split_run_merges_to_the_single_run(self):
        one = self.session('one')
        self.quiet(tankheal.run, ['heal'])
        want = self.files(os.path.join(one, 'heal'))
        parts = self.split_run()
        for order in (parts, parts[::-1]):
            self.session('coord')
            self.quiet(tankheal.merge, 'heal', list(order))
            got = self.files(os.path.join(tankheal.D, 'heal'))
            self.assertEqual(sorted(got), sorted(want))
            for n in want:
                if n != 'run.json':
                    self.assertEqual(got[n], want[n], n)
            a, b = json.loads(got['run.json']), json.loads(want['run.json'])
            for k in ('at', 'seconds', 'runSeconds'):
                a.pop(k), b.pop(k)
            self.assertEqual(sorted(x['main'] for x in a.pop('parts')), ['carries', 'carries', 'here'])
            self.assertEqual(list(a), list(b))  # the same keys in the same order
            self.assertEqual(a, b)
        self.assertEqual(got['.fresh'], want['.fresh'])
        self.assertIn('races', json.loads(got['run.json'])['runSeconds'])

    def test_main_worker_carries_its_own_main(self):
        # heal-0: the main bench, then its own presets on the part it just packed (no second main bench)
        main = self.part('w0', main_only=True)
        n = len(self.calls)
        self.part('w0', presets={'priest_holy'}, carries=main)
        self.assertEqual(len(self.calls), n)
        info, files = tankheal.read_part(os.path.join(self.tmp, 'w0.tgz'))
        self.assertEqual((info['presets'], info['main']), (['priest_holy'], 'carries'))
        self.assertEqual({json.loads(l)['build'] for l in files['races.jsonl'].decode().splitlines()}, {'priest_holy'})

    def test_keep_rest_keeps_the_packs_rows(self):
        one = self.session('one')
        self.quiet(tankheal.run, ['heal'])
        pack = {n.decode() if isinstance(n, bytes) else n: v.decode() for n, v in self.files(os.path.join(one, 'heal')).items()}
        main, a, _ = self.split_run()
        self.session('coord')
        with self.assertRaises(SystemExit):  # strict by default: paladin_holy ... were raced by no part
            self.quiet(tankheal.merge, 'heal', [main, a])
        tankheal.pack_files = lambda mode: pack
        self.quiet(tankheal.merge, 'heal', [main, a], keep_rest=True)
        got = self.files(os.path.join(tankheal.D, 'heal'))
        self.assertEqual(got['races.jsonl'].decode(), pack['races.jsonl'])
        self.assertEqual(got['.fresh'].decode().split('\n')[1], 'priest_holy_raid')
        self.assertEqual(json.loads(got['run.json'])['racesKept'], ['paladin_holy', 'priest_holy', 'shaman_restoration_raid'])


class MergeRefuses(Sandbox):
    def merge_fails(self, parts, why):
        self.session('coord')
        with self.assertRaises(SystemExit) as e:
            self.quiet(tankheal.merge, 'heal', parts)
        self.assertIn(why, str(e.exception))

    def test_other_commit(self):
        main, a, b = self.split_run()
        self.merge_fails([main, a, self.rewrite(b, commit='other')], 'commit')

    def test_other_binaries(self):
        main, a, b = self.split_run()
        self.merge_fails([main, a, self.rewrite(b, stamp=dict(STAMP, binaries={'healbench': 'b2'}))], 'binaries')
        self.merge_fails([main, a, self.rewrite(b, stamp=dict(STAMP, inputHash='h2'))], 'binaries')

    def test_other_splits(self):
        main, a, b = self.split_run()
        self.merge_fails([main, a, self.rewrite(b, splits=8)], 'splits')
        tankheal.SPLITS = 8  # this checkout's DASH_SPLITS against the parts' 4
        self.merge_fails([main, a, b], 'DASH_SPLITS=8')

    def test_other_inputs(self):
        main, a, b = self.split_run()
        self.merge_fails([main, a, self.rewrite(b, inputs='fp2')], 'sim inputs')
        self.inputs = 'fp2'  # HEAD's bench inputs moved since the parts were simmed
        self.merge_fails([main, a, b], 'HEAD')

    def test_carries_of_other_binaries(self):
        main = self.part('main', main_only=True)
        self.stamp = dict(STAMP, binaries={'healbench': 'b2'})
        with self.assertRaises(SystemExit):
            self.part('w1', presets={'priest_holy'}, carries=main)

    def test_main_outputs_differ(self):
        main, a, b = self.split_run()
        info, files = tankheal.read_part(b)
        files['raid_onyxia.txt'] = files['raid_onyxia.txt'].replace(b'| ', b'|  ', 1)
        files['part.json'] = json.dumps(info).encode()
        open(b, 'wb').write(rawstore.tgz(files))
        self.merge_fails([main, a, b], 'raid_onyxia.txt differs')

    def test_preset_twice_differently(self):
        main, a, b = self.split_run()
        main2 = self.part('main2', main_only=True)
        c = self.part('w3', presets={'priest_holy_raid'}, carries=main2)
        info, files = tankheal.read_part(c)
        files['races.jsonl'] = files['races.jsonl'].replace(b'"carry": ', b'"carry": 1', 1)
        files['part.json'] = json.dumps(info).encode()
        open(c, 'wb').write(rawstore.tgz(files))
        self.merge_fails([main, a, b, c], 'race rows differ')

    def test_rows_of_a_preset_it_was_not_given(self):
        main, a, b = self.split_run()
        self.merge_fails([main, a, self.rewrite(b, presets=['priest_holy'])], 'races.jsonl holds')

    def test_no_main(self):
        _, a, b = self.split_run()
        _, fa = tankheal.read_part(a)
        _, fb = tankheal.read_part(b)
        for t, f in ((a, fa), (b, fb)):
            info, _ = tankheal.read_part(t)
            f = {n: v for n, v in f.items() if not n.endswith('.txt')}
            f['part.json'] = json.dumps(info).encode()
            open(t, 'wb').write(rawstore.tgz(f))
        self.merge_fails([a, b], 'no part has the main bench outputs')


class Planner(unittest.TestCase):
    """On this checkout's healing presets (their requests only: the cost model counts race rows)."""

    def setUp(self):
        self.tree = tankheal.TREE
        tankheal.TREE = os.path.dirname(os.path.dirname(TOOLS))

    def tearDown(self):
        tankheal.TREE = self.tree

    def test_priests_alone_rest_balanced(self):
        cost = parallel.costs('heal', tankheal.presets('heal'))
        groups = parallel.split(cost)
        self.assertEqual(sorted(s for g in groups for s in g), sorted(cost))  # every preset once
        for g in groups:
            if any(s.startswith('priest_') for s in g):
                self.assertEqual(len(g), 1, g)
        cap = max(cost[s] for s in cost if s.startswith('priest_'))
        self.assertLessEqual(max(parallel.wall([cost[s] for s in g]) for g in groups), cap)
        self.assertEqual(max(cost, key=cost.get), 'priest_holy_raid')

    def test_more_groups_never_slower(self):
        cost = parallel.costs('heal', tankheal.presets('heal'))
        walls = [max(parallel.wall([cost[s] for s in g]) for g in parallel.split(cost, n)) for n in range(4, 10)]
        self.assertEqual(walls, sorted(walls, reverse=True))
        self.assertEqual(len(parallel.split(cost, 9)), 9)

    def test_too_few_groups_balances_everything(self):
        cost = {'priest_a': 50, 'priest_b': 40, 'druid': 30, 'paladin': 20, 'shaman': 10}
        groups = parallel.split(cost, 2)
        self.assertEqual(len(groups), 2)
        self.assertEqual(sorted(s for g in groups for s in g), sorted(cost))

    def test_balance_by_cost(self):
        cost = {'priest_x': 100, 'a': 60, 'b': 50, 'c': 40, 'd': 30, 'e': 20, 'f': 10}
        groups = parallel.split(cost)
        self.assertEqual(groups[0], ['priest_x'])
        loads = [parallel.wall([cost[s] for s in g]) for g in groups[1:]]
        self.assertLessEqual(max(loads), 100)
        self.assertEqual(parallel.wall([60, 50, 40]), 90)  # two at a time: 60 | 50 + 40

    def test_logged_times_win(self):
        with tempfile.NamedTemporaryFile('w', suffix='.log', delete=False) as f:
            f.write('heal races paladin_holy: 8 rows, 4000 s\nheal tank_onyxia: 29.2 s\n')
        cost = parallel.costs('heal', ['paladin_holy', 'priest_holy'], [f.name])
        os.remove(f.name)
        self.assertEqual(cost['paladin_holy'], 4000)
        self.assertGreater(cost['priest_holy'], 0)

    def test_plan_commands(self):
        pl = parallel.plan('r1', 'abc', heal_groups=5)
        names = [s['session'] for s in pl['sessions']]
        self.assertEqual(names, ['dps', 'tank', 'heal-0', 'heal-1', 'heal-2', 'heal-3', 'heal-4'])
        h0, h1 = pl['sessions'][2], pl['sessions'][3]
        self.assertIn('--main', h0['command'])
        self.assertIn('--carries-from heal-0', h1['command'])
        self.assertNotIn('--main', h1['command'])


class Transport(unittest.TestCase):
    """push, wait_for and gather against a bare repository standing in for origin."""

    def setUp(self):
        self.tmp = tempfile.mkdtemp(prefix='dps_parallel_git_')
        self.saved = (parallel.ROOT, parallel.REMOTE, parallel.D)
        remote, work = os.path.join(self.tmp, 'remote.git'), os.path.join(self.tmp, 'work')
        subprocess.run(['git', 'init', '-q', '--bare', remote], check=True)
        subprocess.run(['git', 'init', '-q', work], check=True)
        parallel.ROOT, parallel.REMOTE, parallel.D = work, remote, os.path.join(self.tmp, 'dash')

    def tearDown(self):
        parallel.ROOT, parallel.REMOTE, parallel.D = self.saved
        shutil.rmtree(self.tmp, ignore_errors=True)

    def put(self, name, data):
        p = os.path.join(self.tmp, name)
        open(p, 'wb').write(data)
        return p

    def test_push_twice_then_wait_for(self):
        with contextlib.redirect_stdout(io.StringIO()):
            parallel.push('r1', 'heal-0', {'heal-main.tgz': self.put('a', b'main')})
            parallel.push('r1', 'heal-0', {'heal-0.tgz': self.put('b', b'races')})
            got = parallel.wait_for('r1', 'heal-0', 'heal-main.tgz', os.path.join(self.tmp, 'got'), timeout=5, every=1)
        with open(got, 'rb') as f:
            self.assertEqual(f.read(), b'main')
        tip = parallel.remote_tips(parallel.ref_of('r1', 'heal-0'))[parallel.ref_of('r1', 'heal-0')]
        self.assertEqual(sorted(parallel.tree_files(tip)), ['heal-0.tgz', 'heal-main.tgz'])
        with self.assertRaises(SystemExit), contextlib.redirect_stdout(io.StringIO()):
            parallel.wait_for('r1', 'heal-1', 'heal-main.tgz', os.path.join(self.tmp, 'x'), timeout=0, every=0)

    def test_clean_deletes_only_this_runs_branches(self):
        with contextlib.redirect_stdout(io.StringIO()):
            parallel.push('r1', 'tank', {'tank.tgz': self.put('a', b'x')})
            parallel.push('r1', 'plan', {'plan.json': self.put('b', b'{}')})
            parallel.push('r2', 'tank', {'tank.tgz': self.put('c', b'y')})
            parallel.clean('r1')
        self.assertEqual(sorted(parallel.remote_tips(f'refs/heads/{parallel.PREFIX}/*')), [parallel.ref_of('r2', 'tank')])

    def test_gather_wants_every_planned_part(self):
        pl = {'run': 'r1', 'commit': 'c1', 'sessions': [{'session': 'tank', 'mode': 'tank'}, {'session': 'heal-0', 'mode': 'heal', 'main': True, 'presets': ['x']}]}
        info = {'mode': 'tank', 'commit': 'c1', 'splits': 4, 'stamp': STAMP, 'presets': ['x']}
        with contextlib.redirect_stdout(io.StringIO()):
            parallel.push('r1', 'plan', {'plan.json': self.put('plan.json', json.dumps(pl).encode())})
            parallel.push('r1', 'tank', {'tank.tgz': self.put('t.tgz', rawstore.tgz({'part.json': json.dumps(info).encode()}))})
            with self.assertRaises(SystemExit) as e:
                parallel.gather('r1', os.path.join(self.tmp, 'g'))
            self.assertIn('heal-0: heal-main.tgz', str(e.exception))
            heal = dict(info, mode='heal')
            parallel.push('r1', 'heal-0', {'heal-main.tgz': self.put('m.tgz', rawstore.tgz({'part.json': json.dumps(dict(heal, presets=[])).encode()})),
                                           'heal-0.tgz': self.put('h.tgz', rawstore.tgz({'part.json': json.dumps(dict(heal, stamp=dict(STAMP, go='go2'))).encode()}))})
            with self.assertRaises(SystemExit) as e:
                parallel.gather('r1', os.path.join(self.tmp, 'g'))
            self.assertIn('other inputs or Go', str(e.exception))


if __name__ == '__main__':
    unittest.main()
