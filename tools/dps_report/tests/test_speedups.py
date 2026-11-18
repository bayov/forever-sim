"""Tests for the pipeline's timing shim, the probe key and skip, and tankheal's resume decision.

Run: python3 -m unittest discover -s tools/dps_report/tests
No sims run: the benches and the probe are stand-ins.
"""
import json, os, stat, subprocess, sys, tempfile, unittest

HERE = os.path.dirname(os.path.abspath(__file__))
TOOLS = os.path.dirname(HERE)
SCRATCH = tempfile.mkdtemp(prefix='dps_report_test_')
# probe.py and tankheal.py read these when imported
os.environ.update(DASH_DIR=os.path.join(SCRATCH, 'dash'), DASH_BIN=os.path.join(SCRATCH, 'dash', 'bin'),
                  DASH_TREE=os.path.join(SCRATCH, 'tree'))
for d in ('dash/req', 'dash/bin', 'dash/asm_apls', 'tree/assets/database'):
    os.makedirs(os.path.join(SCRATCH, d), exist_ok=True)
sys.path.insert(0, TOOLS)
import timing_cli  # noqa: E402
import probe  # noqa: E402
import tankheal  # noqa: E402


def write(path, text):
    os.makedirs(os.path.dirname(path), exist_ok=True)
    with open(path, 'w') as fh:
        fh.write(text)


class TimingCliTest(unittest.TestCase):
    def setUp(self):
        self.tmp = tempfile.mkdtemp(dir=SCRATCH)
        self.real = os.path.join(self.tmp, 'real')
        self.shims = os.path.join(self.tmp, 'shims')
        os.makedirs(self.real)
        os.makedirs(self.shims)
        self.log = os.path.join(self.tmp, 'log.jsonl')
        # a bench that prints to both streams, and a request and boss it can be described from
        for name in ('tankbench', 'healbench'):
            p = os.path.join(self.real, name)
            write(p, '#!/bin/sh\necho "out $*"\necho "err line" >&2\nexit ${BENCH_RC:-0}\n')
            os.chmod(p, os.stat(p).st_mode | stat.S_IXUSR)
            os.symlink(os.path.join(TOOLS, 'timing_cli.py'), os.path.join(self.shims, name))
        self.req = os.path.join(self.tmp, 'warrior_protection.json')
        write(self.req, json.dumps({'raid': {'parties': [{'players': [{'class': 'ClassWarrior', 'race': 'RaceHuman'}]}]},
                                    'encounter': {'duration': 180}}))
        self.req2 = os.path.join(self.tmp, 'druid_bear.json')
        write(self.req2, json.dumps({'raid': {'parties': [{'players': [{'class': 'ClassDruid', 'race': 'RaceNightElf'}]}]}}))
        self.boss = os.path.join(self.tmp, 'magmadar.json')
        write(self.boss, json.dumps({'name': 'Magmadar', 'duration': 90}))

    def call(self, exe_dir, name, *argv, **env):
        e = dict(os.environ, TIMING_REAL_BIN=self.real, **env)
        e.pop('TIMING_LOG', None)
        e.update({k: v for k, v in env.items() if v is not None})
        return subprocess.run([os.path.join(exe_dir, name), *argv], capture_output=True, env=e)

    def test_describe_main_tank_run(self):
        d = timing_cli.describe(['-encounter', self.boss, '-iters', '1000', self.req, self.req2], 'tankbench')
        self.assertEqual((d['role'], d['boss'], d['iters'], d['reqs'], d['dur']), ('tank', 'magmadar', 1000, 2, 90))

    def test_describe_main_heal_run(self):
        d = timing_cli.describe(['-carry', '-role', 'raid', '-encounter', self.boss, '-iters', '1000', self.req], 'healbench')
        self.assertEqual((d['role'], d['boss'], d['iters'], d['reqs'], d['carry'], d['race']),
                         ('raid', 'magmadar', 1000, 1, True, 'RaceHuman'))

    def test_describe_race_sim(self):
        d = timing_cli.describe(['-iters', '40', '-first-seed', '7', '-labeled', '-vectors', '/tmp/v.jsonl', '-encounter', self.boss, self.req],
                                'tankbench')
        self.assertEqual((d['iters'], d['reqs'], d['boss']), (40, 1, 'magmadar'))
        h = timing_cli.describe(['-iters', '40', '-seed', '7', '-labeled', '-vectors', '/tmp/v.jsonl', '-carry=false', '-role', 'tank',
                                 '-encounter', self.boss, '-scales', '1.5', self.req], 'healbench')
        self.assertEqual((h['role'], h['carry'], h['reqs'], h['iters']), ('tank', False, 1, 40))

    def test_describe_defaults_and_bad_input(self):
        self.assertEqual(timing_cli.describe(['-iters', '5', '/no/such/file.json'], 'tankbench'), {})
        self.assertEqual(timing_cli.describe(['sim', '--infile', '/no/such'], 'wowsimcli'), {})

    def test_shim_is_silent_and_passes_output_through(self):
        argv = ('-encounter', self.boss, '-iters', '10', self.req)
        direct = self.call(self.real, 'tankbench', *argv)
        for log in (self.log, None):  # logging, and no log at all
            shim = self.call(self.shims, 'tankbench', *argv, **({'TIMING_LOG': log, 'TIMING_STEP': 'tank'} if log else {}))
            self.assertEqual(shim.stdout, direct.stdout)
            self.assertEqual(shim.stderr, direct.stderr)
            self.assertEqual(shim.returncode, 0)
        rec = json.loads(open(self.log).readline())
        self.assertEqual((rec['step'], rec['bin'], rec['boss'], rec['reqs']), ('tank', 'tankbench', 'magmadar', 1))

    def test_shim_silent_when_the_log_cannot_be_written(self):
        shim = self.call(self.shims, 'healbench', self.req, TIMING_LOG=os.path.join(self.tmp, 'nodir', 'log.jsonl'))
        self.assertEqual(shim.stdout, b'out ' + self.req.encode() + b'\n')
        self.assertEqual(shim.stderr, b'err line\n')

    def test_shim_keeps_the_exit_code(self):
        self.assertEqual(self.call(self.shims, 'tankbench', self.req, BENCH_RC='3', TIMING_LOG=self.log).returncode, 3)
        self.assertEqual(json.loads(open(self.log).readline())['rc'], 3)

    def test_phase_is_logged(self):
        subprocess.run([sys.executable, os.path.join(TOOLS, 'timing_cli.py'), 'phase', 'hydrate', '0'], check=True, env=dict(os.environ, TIMING_LOG=self.log))
        rec = json.loads(open(self.log).readline())
        self.assertEqual(rec['phase'], 'hydrate')
        self.assertGreater(rec['wall'], 1e6)  # since epoch 0
        self.assertTrue(os.path.exists(os.path.join(os.environ['DASH_DIR'], 'timing', 'phases.jsonl')))


class ProbeTest(unittest.TestCase):
    def setUp(self):
        self.d = os.environ['DASH_DIR']
        self.scripts = os.path.join(SCRATCH, 'scripts')
        os.makedirs(self.scripts, exist_ok=True)
        probe.HERE = self.scripts
        for f in ('probe.py', 'run_build.py', 'weights.py', 'assumptions_def.py'):
            write(os.path.join(self.scripts, f), f'# {f} v1\n')
        for n in probe.BINARIES:
            write(os.path.join(self.d, 'bin', n), n + ' binary v1')
        write(os.path.join(self.d, 'req', 'a__x.json'), '{"req": 1}')
        write(os.path.join(self.d, 'req', 'a__x.meta.json'), '{"class": "a"}')
        write(os.path.join(self.d, 'req', 'b__y.json'), '{"req": 2}')
        write(os.path.join(self.d, 'req', 'b__y.meta.json'), '{"class": "b"}')
        write(os.path.join(self.d, 'asm_apls', 'one.simc'), 'apl v1')
        write(os.path.join(os.environ['DASH_TREE'], 'assets', 'database', 'db.json'), '{}')
        probe._hashes.clear()
        self.calls = []
        self.orig = probe.probe_one
        probe.probe_one = lambda bid, base, pbin: (self.calls.append(bid), {'settings': probe.SETTINGS, 'sims': {'s:k:1': bid}, 'outputs': {}})[1]
        self.dest = os.path.join(SCRATCH, 'probe_dest_' + str(id(self)))
        self.check = os.path.join(SCRATCH, 'probe_check_' + str(id(self)))

    def tearDown(self):
        probe.probe_one = self.orig

    def key(self, bid='a__x'):
        probe._hashes.clear()
        return probe.probe_key(bid)

    def test_key_follows_every_input(self):
        base = self.key()
        self.assertEqual(base, self.key())
        self.assertNotEqual(base, self.key('b__y'))
        changes = [(os.path.join(self.d, 'req', 'a__x.json'), '{"req": 9}'),
                   (os.path.join(self.d, 'req', 'a__x.meta.json'), '{"class": "z"}'),
                   (os.path.join(self.d, 'bin', 'dashstats'), 'new binary'),
                   (os.path.join(self.d, 'asm_apls', 'one.simc'), 'apl v2'),
                   (os.path.join(self.d, 'asm_apls', 'two.simc'), 'a new apl'),
                   (os.path.join(self.d, 'req', 'races', 'a__x.Orc.json'), '{}'),
                   (os.path.join(self.scripts, 'run_build.py'), '# changed\n'),
                   (os.path.join(self.scripts, 'assumptions_def.py'), '# changed\n'),
                   (os.path.join(os.environ['DASH_TREE'], 'assets', 'database', 'db.json'), '{"x": 1}')]
        for path, text in changes:
            old = open(path).read() if os.path.exists(path) else None
            write(path, text)
            self.assertNotEqual(base, self.key(), path)
            if old is None:
                os.remove(path)
            else:
                write(path, old)
            self.assertEqual(base, self.key(), path)

    def test_key_changes_with_settings(self):
        base = self.key()
        probe.SETTINGS['scale'] += 1
        try:
            self.assertNotEqual(base, self.key())
        finally:
            probe.SETTINGS['scale'] -= 1

    def test_key_follows_a_timing_shim_to_the_real_binary(self):
        real = os.path.join(SCRATCH, 'realbin')
        write(os.path.join(real, 'wowsimcli'), 'wowsimcli binary v1')
        link = os.path.join(self.d, 'bin', 'wowsimcli')
        os.remove(link)
        os.symlink(os.path.join(TOOLS, 'timing_cli.py'), link)
        os.environ['TIMING_REAL_BIN'] = real
        try:
            with_shim = self.key()
            write(os.path.join(real, 'wowsimcli'), 'wowsimcli binary v2')
            self.assertNotEqual(with_shim, self.key())
        finally:
            del os.environ['TIMING_REAL_BIN']
            os.remove(link)
            write(link, 'wowsimcli binary v1')

    def test_run_probes_then_skips(self):
        probe.run(['a__x', 'b__y'], self.dest)
        self.assertEqual(sorted(self.calls), ['a__x', 'b__y'])
        rec = json.load(open(os.path.join(self.dest, 'a__x.json')))
        self.assertEqual(rec['_key'], self.key())
        self.calls.clear()
        probe.run(['a__x', 'b__y'], self.dest)  # a restart: nothing left to do
        self.assertEqual(self.calls, [])
        write(os.path.join(self.d, 'req', 'b__y.json'), '{"req": 3}')  # one request changed
        probe._hashes.clear()  # (a new process in real use)
        probe.run(['a__x', 'b__y'], self.dest)
        self.assertEqual(self.calls, ['b__y'])

    def test_run_copies_a_matching_probe_check(self):
        probe.run(['a__x', 'b__y'], self.check)
        self.calls.clear()
        write(os.path.join(self.d, 'req', 'b__y.json'), '{"req": 4}')  # the check's probe of b__y is out of date
        probe._hashes.clear()  # (a new process in real use)
        res = probe.run(['a__x', 'b__y'], self.dest, reuse_from=self.check)
        self.assertEqual(self.calls, ['b__y'])
        self.assertEqual(json.load(open(os.path.join(self.dest, 'a__x.json'))), json.load(open(os.path.join(self.check, 'a__x.json'))))
        self.assertEqual(res['a__x']['_key'], self.key())

    def test_run_reprobes_a_record_without_a_key_or_a_failed_one(self):
        os.makedirs(self.dest)
        write(os.path.join(self.dest, 'a__x.json'), json.dumps({'settings': probe.SETTINGS, 'sims': {}, 'outputs': {}}))
        write(os.path.join(self.dest, 'b__y.json'), json.dumps({'settings': probe.SETTINGS, 'sims': {}, 'outputs': {}, 'failed': {'weights': 1},
                                                                 '_key': self.key('b__y')}))
        probe.run(['a__x', 'b__y'], self.dest)
        self.assertEqual(sorted(self.calls), ['a__x', 'b__y'])

    def test_force_probes_everything(self):
        probe.run(['a__x'], self.dest)
        self.calls.clear()
        os.environ['PROBE_FORCE'] = '1'
        try:
            probe.run(['a__x'], self.dest)
        finally:
            del os.environ['PROBE_FORCE']
        self.assertEqual(self.calls, ['a__x'])

    def test_compare_ignores_the_key(self):
        a = {'settings': probe.SETTINGS, 'sims': {'s:k:1': 'x'}, 'outputs': {}, '_key': 'one'}
        b = dict(a, _key='two')
        self.assertEqual(probe.compare(a, b), [])


class ResumeTest(unittest.TestCase):
    def test_same_commit_resumes(self):
        self.assertTrue(tankheal.can_resume({'commit': 'aaa', 'inputs': 'f1'}, 'aaa', 'f2'))
        self.assertTrue(tankheal.can_resume({'commit': 'aaa'}, 'aaa', None))  # an older progress file, or no fingerprint

    def test_other_commit_same_fingerprint_resumes(self):
        self.assertTrue(tankheal.can_resume({'commit': 'aaa', 'inputs': 'f1'}, 'bbb', 'f1'))

    def test_other_commit_other_fingerprint_starts_over(self):
        self.assertFalse(tankheal.can_resume({'commit': 'aaa', 'inputs': 'f1'}, 'bbb', 'f2'))

    def test_other_commit_without_a_fingerprint_starts_over(self):
        self.assertFalse(tankheal.can_resume({'commit': 'aaa', 'inputs': 'f1'}, 'bbb', None))
        self.assertFalse(tankheal.can_resume({'commit': 'aaa'}, 'bbb', None))  # None must not match a missing one
        self.assertFalse(tankheal.can_resume({'commit': 'aaa'}, 'bbb', 'f1'))

    def test_nothing_to_resume(self):
        self.assertFalse(tankheal.can_resume({}, 'aaa', 'f1'))

    def test_bench_inputs_is_stable_at_head_and_none_for_an_unknown_commit(self):
        head = subprocess.run(['git', 'rev-parse', 'HEAD'], cwd=tankheal.ROOT, capture_output=True, text=True).stdout.strip()
        a = tankheal.bench_inputs('tank', head)
        self.assertTrue(a)
        self.assertEqual(a, tankheal.bench_inputs('tank', head))
        self.assertNotEqual(a, tankheal.bench_inputs('heal', head))
        self.assertIsNone(tankheal.bench_inputs('tank', '0' * 40))


if __name__ == '__main__':
    unittest.main()
