#!/usr/bin/env python3
"""Tests for build_bins.py (the input hash and the guard) and for the removed wowsimcli fallbacks.

  python3 -m unittest discover -s tools/dps_report -p 'test_*.py'

Each test makes a small git repository with the layout the hash reads (sim/, proto/, assets/database, the patches) and
a fake binaries directory, so nothing is built and the real checkout is not touched.
"""
import contextlib, hashlib, importlib, io, json, os, shutil, subprocess, sys, tempfile, unittest, warnings
from unittest import mock

HERE = os.path.dirname(os.path.abspath(__file__))
sys.path.insert(0, HERE)
import build_bins  # noqa: E402


def git(root, *a):
    return subprocess.run(['git', *a], cwd=root, check=True, capture_output=True, text=True).stdout.strip()


class Repo:
    def __init__(self):
        self.root = tempfile.mkdtemp(prefix='bbtest_')
        git(self.root, 'init', '-q')
        git(self.root, 'config', 'user.email', 't@example.com')
        git(self.root, 'config', 'user.name', 't')
        files = {
            'sim/core/a.go': 'package core\n',
            'sim/core/a_test.go': 'package core\n',
            'sim/core/NOTES.md': 'notes\n',
            'sim/warrior/w.go': 'package warrior\n',
            'sim/register.go': 'package sim\n',
            'proto/api.proto': 'syntax = "proto3";\n',
            'cmd/wowsimcli/main.go': 'package main\n',
            'go.mod': 'module x\n',
            'go.sum': '',
            'assets/database/db.json': '{"items": []}\n',
            'assets/database/db.bin': 'bin\n',
            'tools/tankbench/main.go': 'package main\n',
            'tools/tankbench/req/a.json': '{}\n',
            'tools/healbench/main.go': 'package main\n',
            'tools/tankscore/s.go': 'package tankscore\n',
            'README.md': 'top\n',
            'ui/x.ts': 'x\n',
        }
        for n in build_bins.used_patches():
            files[f'{build_bins.PATCH_DIR}/{n}-p.patch'] = f'patch {n}\n'
        files[f'{build_bins.PATCH_DIR}/0099-unused.patch'] = 'unused\n'
        for p, s in files.items():
            self.write(p, s)
        self.commit('init')

    def write(self, path, text):
        p = os.path.join(self.root, path)
        os.makedirs(os.path.dirname(p), exist_ok=True)
        open(p, 'w').write(text)

    def commit(self, msg):
        git(self.root, 'add', '-A')
        git(self.root, 'commit', '-q', '--allow-empty', '-m', msg)
        return git(self.root, 'rev-parse', 'HEAD')

    def hash(self, rev='HEAD'):
        return build_bins.input_hash(rev, self.root)['hash']

    def close(self):
        shutil.rmtree(self.root, ignore_errors=True)


class Case(unittest.TestCase):
    def setUp(self):
        warnings.simplefilter('ignore', ResourceWarning)  # the report scripts open files without closing them


class InputHash(Case):
    def setUp(self):
        super().setUp()
        self.r = Repo()
        self.addCleanup(self.r.close)

    def changed_by(self, path, text):
        before = self.r.hash()
        self.r.write(path, text)
        self.r.commit('change ' + path)
        return before != self.r.hash()

    def test_same_inputs_at_two_commits(self):
        first = git(self.r.root, 'rev-parse', 'HEAD')
        self.r.write('README.md', 'other\n')
        self.r.write('ui/x.ts', 'y\n')
        self.r.write('tools/tankbench/req/a.json', '{"x": 1}\n')  # a bench request is read at run time, not built
        second = self.r.commit('no input changed')
        self.assertNotEqual(first, second)
        self.assertEqual(self.r.hash(first), self.r.hash(second))

    def test_sim_change(self):
        self.assertTrue(self.changed_by('sim/warrior/w.go', 'package warrior // changed\n'))

    def test_new_sim_file(self):
        self.assertTrue(self.changed_by('sim/warrior/new.go', 'package warrior\n'))

    def test_db_change(self):
        self.assertTrue(self.changed_by('assets/database/db.json', '{"items": [1]}\n'))

    def test_db_markdown_counts(self):  # all of assets/database is in
        self.assertTrue(self.changed_by('assets/database/NOTES.md', 'x\n'))

    def test_proto_tool_and_patch_changes(self):
        self.assertTrue(self.changed_by('proto/api.proto', 'syntax = "proto3"; // x\n'))
        self.assertTrue(self.changed_by('tools/tankscore/s.go', 'package tankscore // x\n'))
        self.assertTrue(self.changed_by(f'{build_bins.PATCH_DIR}/0002-p.patch', 'patch 2 changed\n'))
        self.assertTrue(self.changed_by('go.sum', 'x\n'))

    def test_tests_and_docs_dont(self):
        self.assertFalse(self.changed_by('sim/core/a_test.go', 'package core // changed\n'))
        self.assertFalse(self.changed_by('sim/core/NOTES.md', 'changed\n'))
        self.assertFalse(self.changed_by('sim/core/x.results', 'golden\n'))
        self.assertFalse(self.changed_by('sim/core/testdata/t.json', '{}\n'))
        self.assertFalse(self.changed_by(f'{build_bins.PATCH_DIR}/0099-unused.patch', 'changed\n'))

    def test_recipe_counts(self):
        before = self.r.hash()
        with mock.patch.dict(build_bins.BINARIES, {'tankbench': ('./tools/tankbench', ['0002', '0003', '0005', '0006'])}):
            self.assertNotEqual(before, self.r.hash())
        with mock.patch.object(build_bins, 'FLAGS', build_bins.FLAGS + ['-race']):
            self.assertNotEqual(before, self.r.hash())

    def test_missing_patch_file(self):
        os.remove(os.path.join(self.r.root, build_bins.PATCH_DIR, '0004-p.patch'))
        self.r.commit('drop 0004')
        with self.assertRaises(SystemExit):
            self.r.hash()


class Guard(Case):
    NEED = ['wowsimcli', 'wowsimcli_r']

    def setUp(self):
        super().setUp()
        self.r = Repo()
        self.addCleanup(self.r.close)
        self.bin = tempfile.mkdtemp(prefix='bbbin_')
        self.addCleanup(shutil.rmtree, self.bin, True)
        self.dash = tempfile.mkdtemp(prefix='bbdash_')
        self.addCleanup(shutil.rmtree, self.dash, True)
        bins = {}
        for n in build_bins.BINARIES:
            p = os.path.join(self.bin, n)
            open(p, 'wb').write(b'binary ' + n.encode())
            bins[n] = {'sha256': hashlib.sha256(open(p, 'rb').read()).hexdigest()}
        h = build_bins.input_hash('HEAD', self.r.root)
        self.stamp = {'format': 1, 'inputHash': h['hash'], 'inputs': h['groups'], 'commit': git(self.r.root, 'rev-parse', 'HEAD'), 'binaries': bins}
        self.write_stamp()
        env = mock.patch.dict(os.environ)
        env.start()
        self.addCleanup(env.stop)
        for k in (build_bins.OVERRIDE, 'SIM_COMMIT'):
            os.environ.pop(k, None)

    def write_stamp(self):
        json.dump(self.stamp, open(os.path.join(self.bin, build_bins.STAMP), 'w'))

    def guard(self, need=None):
        err = io.StringIO()
        with contextlib.redirect_stderr(err):
            rec = build_bins.guard(self.bin, need or self.NEED, 'test', root=self.r.root)
        return rec, err.getvalue()

    def refused(self, need=None):
        with self.assertRaises(SystemExit) as cm:
            self.guard(need)
        return str(cm.exception.code)

    def test_ok(self):
        self.assertEqual(self.guard(), (None, ''))

    def test_ok_at_another_commit_with_the_same_inputs(self):
        self.r.write('README.md', 'other\n')
        self.r.commit('docs')
        self.assertEqual(self.guard(), (None, ''))

    def test_sim_commit(self):
        first = git(self.r.root, 'rev-parse', 'HEAD')
        self.r.write('sim/core/a.go', 'package core // later\n')
        self.r.commit('later')
        self.assertIn('other inputs', self.refused())
        os.environ['SIM_COMMIT'] = first
        self.assertEqual(self.guard(), (None, ''))

    def test_missing_stamp(self):
        os.remove(os.path.join(self.bin, build_bins.STAMP))
        self.assertIn('no STAMP.json', self.refused())

    def test_mismatch_names_what_changed(self):
        self.r.write('sim/warrior/w.go', 'package warrior // changed\n')
        self.r.write('assets/database/db.json', '{"items": [2]}\n')
        self.r.commit('sim and db')
        msg = self.refused()
        self.assertIn('sim/warrior', msg)
        self.assertIn('assets/database', msg)
        self.assertNotIn('sim/core', msg)

    def test_binary_replaced(self):
        open(os.path.join(self.bin, 'wowsimcli_r'), 'wb').write(b'hand built')
        self.assertIn('wowsimcli_r is not the binary', self.refused())

    def test_binary_not_needed_is_not_checked(self):
        open(os.path.join(self.bin, 'healbench'), 'wb').write(b'hand built')
        self.assertEqual(self.guard(), (None, ''))
        self.assertIn('healbench', self.refused(['healbench']))

    def test_missing_binary_is_an_error_even_with_override(self):
        os.remove(os.path.join(self.bin, 'wowsimcli_r'))
        os.environ[build_bins.OVERRIDE] = 'testing'
        self.assertIn('wowsimcli_r missing', self.refused())

    def test_dirty_tree(self):
        self.r.write('sim/core/a.go', 'package core // uncommitted\n')
        self.assertIn('uncommitted changes', self.refused())

    def test_untracked_sim_file(self):
        self.r.write('sim/warrior/new.go', 'package warrior\n')
        self.assertIn('sim/warrior/new.go', self.refused())

    def test_dirty_outside_inputs_is_fine(self):
        self.r.write('sim/core/a_test.go', 'package core // uncommitted\n')
        self.r.write('sim/core/NOTES.md', 'uncommitted\n')
        self.r.write('README.md', 'uncommitted\n')
        self.r.write('tools/tankbench/req/b.json', '{}\n')
        self.assertEqual(self.guard(), (None, ''))

    def test_override(self):
        self.r.write('sim/core/a.go', 'package core // uncommitted\n')
        os.environ[build_bins.OVERRIDE] = 'checking a hunch'
        rec, err = self.guard()
        self.assertEqual(rec['reason'], 'checking a hunch')
        self.assertTrue(any('uncommitted' in p for p in rec['problems']))
        self.assertIn('DASH_ALLOW_BIN_MISMATCH=checking a hunch', err)
        build_bins.mark_builds(['warrior__fury'], rec, self.dash)
        p = build_bins.override_path('warrior__fury', self.dash)
        self.assertEqual(json.load(open(p))['reason'], 'checking a hunch')
        build_bins.clear_builds(['warrior__fury'], self.dash)
        self.assertFalse(os.path.exists(p))

    def test_cli_marks_and_clears(self):
        def cli(*args):
            with mock.patch.object(build_bins, 'ROOT', self.r.root), mock.patch.object(build_bins.dashpaths, 'D', self.dash), \
                    mock.patch.object(build_bins.guard, '__defaults__', (None, self.r.root)), contextlib.redirect_stderr(io.StringIO()):
                build_bins.main(['guard', '--bin', self.bin, '--need', 'wowsimcli', *args])
        p = build_bins.override_path('mage__fire', self.dash)
        self.r.write('sim/core/a.go', 'package core // uncommitted\n')
        os.environ[build_bins.OVERRIDE] = 'why not'
        cli('--mark', 'mage__fire', '--clear')
        self.assertTrue(os.path.exists(p))  # a mismatch marks, even for a run
        git(self.r.root, 'checkout', '--', 'sim/core/a.go')
        cli('--mark', 'mage__fire')
        self.assertTrue(os.path.exists(p))  # a resume keeps it: some of its results may be from the override
        cli('--mark', 'mage__fire', '--clear')
        self.assertFalse(os.path.exists(p))  # a whole clean run drops it


class NoFallback(Case):
    """run_races.py and run_assumptions.py use their own binary or stop, and probe.py stands in for every binary."""

    def setUp(self):
        super().setUp()
        self.bin = tempfile.mkdtemp(prefix='bbbin_')
        self.addCleanup(shutil.rmtree, self.bin, True)
        self.dash = tempfile.mkdtemp(prefix='bbdash_')
        self.addCleanup(shutil.rmtree, self.dash, True)
        os.makedirs(os.path.join(self.dash, 'req'))
        open(os.path.join(self.bin, 'wowsimcli'), 'w').write('plain')
        env = mock.patch.dict(os.environ, DASH_BIN=self.bin, DASH_DIR=self.dash)
        env.start()
        self.addCleanup(env.stop)

    def load(self, name):
        for m in (name, 'dashpaths'):
            sys.modules.pop(m, None)
        mod = importlib.import_module(name)
        self.addCleanup(sys.modules.pop, name, None)
        self.addCleanup(sys.modules.pop, 'dashpaths', None)
        return mod

    def test_run_races(self):
        m = self.load('run_races')
        self.assertEqual(m.CLI, os.path.join(self.bin, 'wowsimcli_r'))
        with self.assertRaises(SystemExit) as cm:
            m.main([])
        self.assertIn('wowsimcli_r', str(cm.exception.code))

    def test_run_assumptions(self):
        m = self.load('run_assumptions')
        self.assertEqual(m.CLI, os.path.join(self.bin, 'wowsimcli_assume'))
        with self.assertRaises(SystemExit) as cm:
            m.main([])
        self.assertIn('wowsimcli_assume', str(cm.exception.code))

    def test_probe_bin(self):
        m = self.load('probe')
        with self.assertRaises(SystemExit) as cm:
            m.probe_bin(self.dash)
        self.assertIn('wowsimcli_r', str(cm.exception.code))
        for n in m.BINARIES:
            open(os.path.join(self.bin, n), 'w').write('x')
        b = m.probe_bin(self.dash)
        self.assertEqual(sorted(os.listdir(b)), sorted(m.BINARIES))


if __name__ == '__main__':
    unittest.main()
