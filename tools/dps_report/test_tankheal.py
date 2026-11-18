#!/usr/bin/env python3
"""Tests for tankheal.py's per-preset race reuse: python3 tools/dps_report/test_tankheal.py (no sims run)."""
import contextlib, io, json, os, sys, tempfile, unittest

HERE = os.path.dirname(os.path.abspath(__file__))
sys.path.insert(0, HERE)
import tankheal  # noqa: E402

HEAD = '| Healer | HPS | Carry |\n|---|---|---|\n'


def heal_texts(rows, races=None):
    """A heal pack's files: every boss file of both roles with rows {stem: 'cells'}, and races.jsonl."""
    files = {}
    for role in ('tank', 'raid'):
        for b in tankheal.BOSSES:
            files[f'{role}_{b}.txt'] = 'Onyxia, role ...\n' + HEAD + ''.join(f'| {s} | {c} |\n' for s, c in rows.items() if tankheal.heal_role(s) == role)
    if races is not None:
        files['races.jsonl'] = ''.join(json.dumps({'build': s, 'race': r, 'variant': 'race'}) + '\n' for s, r in races)
    return files


class Reuse(unittest.TestCase):
    stems = ['paladin_holy', 'priest_holy', 'shaman_restoration_raid']
    rows = {'paladin_holy': '177 | 1.018', 'priest_holy': '284 | 0.988', 'shaman_restoration_raid': '293 | 0.946'}
    races = [('priest_holy', 'Human'), ('paladin_holy', 'Human'), ('priest_holy', 'Dwarf'), ('shaman_restoration_raid', 'Orc')]

    def test_same_rows_kept(self):
        old = heal_texts(self.rows, self.races)
        keep = tankheal.reusable('heal', self.stems, set(), old, heal_texts(self.rows))
        self.assertEqual(sorted(keep), self.stems)
        self.assertEqual(len(keep['priest_holy']), 2)

    def test_stale_preset_resimmed(self):
        old = heal_texts(self.rows, self.races)
        keep = tankheal.reusable('heal', self.stems, {'priest_holy'}, old, heal_texts(self.rows))
        self.assertEqual(sorted(keep), ['paladin_holy', 'shaman_restoration_raid'])

    def test_moved_main_rows_resimmed(self):
        # one byte of one boss file's row differs: the preset's sims moved, so its race rows are simmed again
        old = heal_texts(self.rows, self.races)
        new = heal_texts(self.rows)
        new['tank_onyxia.txt'] = new['tank_onyxia.txt'].replace('| 284 | 0.988 |', '| 284 | 0.989 |')
        keep = tankheal.reusable('heal', self.stems, set(), old, new)
        self.assertEqual(sorted(keep), ['paladin_holy', 'shaman_restoration_raid'])

    def test_missing_rows_resimmed(self):
        old = heal_texts(self.rows, [r for r in self.races if r[0] != 'paladin_holy'])
        self.assertNotIn('paladin_holy', tankheal.reusable('heal', self.stems, set(), old, heal_texts(self.rows)))
        old = heal_texts({k: v for k, v in self.rows.items() if k != 'paladin_holy'}, self.races)
        self.assertNotIn('paladin_holy', tankheal.reusable('heal', self.stems, set(), old, heal_texts(self.rows)))

    def test_no_pack(self):
        self.assertEqual(tankheal.reusable('heal', self.stems, set(), None, heal_texts(self.rows)), {})

    def test_tank_rows(self):
        txt = 'Healing: ...\n| Build | Max threat |\n|---|---|\n| druid_bear_balanced | 3038 ± 23 |\n| druid_bear_max_threat | 3033 ± 23 |\n'
        old = {b + '.txt': txt for b in tankheal.BOSSES}
        old['races.jsonl'] = json.dumps({'build': 'druid_bear_balanced'}) + '\n' + json.dumps({'build': 'druid_bear_max_threat'}) + '\n'
        new = dict(old)
        new['onyxia.txt'] = txt.replace('3033', '3034')
        self.assertEqual(sorted(tankheal.reusable('tank', ['druid_bear_balanced', 'druid_bear_max_threat'], set(), old, new)), ['druid_bear_balanced'])

    def test_real_packs_keep_everything_against_themselves(self):
        for mode in tankheal.MODES:
            old = tankheal.pack_files(mode)
            if old is None:
                continue
            stems = tankheal.presets(mode)
            self.assertEqual(sorted(tankheal.reusable(mode, stems, set(), old, old)), stems, mode)


class Sort(unittest.TestCase):
    def test_preset_order_stable_within(self):
        lines = [json.dumps({'build': b, 'n': i}) + '\n' for i, b in enumerate(['b', 'a', 'b', 'c', 'a'])]
        got = [json.loads(l) for l in tankheal.race_lines_sorted(lines, ['a', 'b', 'c'])]
        self.assertEqual([(r['build'], r['n']) for r in got], [('a', 1), ('a', 4), ('b', 0), ('b', 2), ('c', 3)])


class Fresh(unittest.TestCase):
    def test_fresh_names_partial_runs(self):
        with tempfile.TemporaryDirectory() as d:
            old, tankheal.D = tankheal.D, d
            try:
                for m, txt in (('tank', 'abc\n'), ('heal', 'abc\npriest_holy,paladin_holy\n')):
                    os.makedirs(os.path.join(d, m))
                    open(os.path.join(d, m, '.fresh'), 'w').write(txt)
                buf = io.StringIO()
                with contextlib.redirect_stdout(buf):
                    tankheal.fresh()
                self.assertEqual(buf.getvalue().strip(), 'tank heal:priest_holy,paladin_holy')
            finally:
                tankheal.D = old


class ProbeKey(unittest.TestCase):
    """The bench probe's _key: every input a preset's probe is made from moves it, nothing else does."""

    def setUp(self):
        self.tmp = tempfile.mkdtemp()
        self.saved = (tankheal.BIN, tankheal.TREE, tankheal.HERE)
        tankheal.BIN, tankheal.TREE, tankheal.HERE = (os.path.join(self.tmp, d) for d in ('bin', 'tree', 'scripts'))
        self.put('bin/healbench', 'healbench v1')
        self.put('tree/tools/healbench/req/priest_holy.json', '{"a": 1}')
        self.put('tree/tools/healbench/req/paladin_holy.json', '{"a": 2}')
        for b in tankheal.BOSSES:
            self.put(f'tree/tools/encounters/{b}.json', b)
        for p in tankheal.HEAL_RAID:
            self.put('tree/' + p, p)
        for f in tankheal.PROBE_SCRIPTS:
            self.put('scripts/' + f, f)
        self.settings = tankheal.probe_settings()

    def tearDown(self):
        tankheal.BIN, tankheal.TREE, tankheal.HERE = self.saved
        tankheal._hashes.clear()
        os.environ.pop('TIMING_REAL_BIN', None)

    def put(self, rel, text):
        p = os.path.join(self.tmp, rel)
        os.makedirs(os.path.dirname(p), exist_ok=True)
        if os.path.lexists(p):
            os.remove(p)
        open(p, 'w').write(text)

    def key(self, stem='priest_holy', settings=None):
        tankheal._hashes.clear()
        return tankheal.probe_key('heal', stem, settings or self.settings)

    def test_stable(self):
        self.assertEqual(self.key(), self.key())
        self.assertNotEqual(self.key(), self.key('paladin_holy'))

    def test_each_input_moves_it(self):
        base = self.key()
        for rel in ('bin/healbench', 'tree/tools/healbench/req/priest_holy.json', 'tree/tools/encounters/onyxia.json',
                    'tree/' + tankheal.HEAL_RAID[0], 'tree/' + tankheal.HEAL_RAID[1], 'scripts/bench_races.py', 'scripts/tankheal.py'):
            old = open(os.path.join(self.tmp, rel)).read()
            self.put(rel, old + ' changed')
            self.assertNotEqual(self.key(), base, rel)
            self.put(rel, old)
            self.assertEqual(self.key(), base, rel)

    def test_settings_move_it(self):
        for k in ('iters', 'racePilot', 'pilot', 'ceil', 'healCeil', 'splits', 'arch'):
            s = dict(self.settings)
            s[k] = s[k] + 1 if isinstance(s[k], int) else s[k] + 'x'
            self.assertNotEqual(self.key(settings=s), self.key(), k)

    def test_another_preset_does_not_move_it(self):
        base = self.key()
        self.put('tree/tools/healbench/req/paladin_holy.json', '{"a": 3}')
        self.assertEqual(self.key(), base)

    def test_follows_a_timing_shim_to_the_real_binary(self):
        base = self.key()
        real = os.path.join(self.tmp, 'real')
        os.makedirs(real)
        open(os.path.join(real, 'healbench'), 'w').write('healbench v1')
        link = os.path.join(self.tmp, 'bin', 'healbench')
        os.remove(link)
        os.symlink(os.path.join(HERE, 'timing_cli.py'), link)
        os.environ['TIMING_REAL_BIN'] = real
        self.assertEqual(self.key(), base)
        open(os.path.join(real, 'healbench'), 'w').write('healbench v2')
        self.assertNotEqual(self.key(), base)

    def test_sizes_are_a_tenth(self):
        import adaptive
        s = self.settings
        self.assertEqual((s['iters'], s['racePilot'], s['pilot'], s['ceil'], s['healCeil']),
                         (tankheal.ITERS // 10, adaptive.RACE_PILOT // 10, adaptive.PILOT // 10, adaptive.CEIL // 10, adaptive.HEAL_CEIL // 10))
        env = tankheal.probe_env('heal', s, '/w')
        self.assertEqual((env['DASH_DIR'], env['ADAPT_HEAL_MAX'], env['ADAPT_RACE_PILOT'], env['TANKHEAL_PROBE_ITERS']),
                         ('/w', str(s['healCeil']), str(s['racePilot']), str(s['iters'])))


def probe_rec(main=None, races=None, values=None, **kw):
    rec = {'settings': {'format': 1, 'scale': 10}, '_key': 'k', 'main': main or {'tank_onyxia': 'a', 'tank_ragnaros': 'b'},
           'races': races or {'Human:race': 'r1', 'Human:off': 'r2'}, 'values': values or {'Human:race': 1.0, 'Human:off': 0.99}}
    rec.update(kw)
    return rec


class ProbeCompare(unittest.TestCase):
    def test_same(self):
        self.assertEqual(tankheal.probe_compare(probe_rec(), probe_rec(_key='other')), [])  # the key isn't compared

    def test_no_stored_probe(self):
        self.assertIn('no stored probe', tankheal.probe_compare(None, probe_rec())[0])

    def test_other_settings(self):
        self.assertIn('other settings', tankheal.probe_compare(probe_rec(settings={'scale': 5}), probe_rec())[0])

    def test_failed(self):
        self.assertIn('failed', tankheal.probe_compare(probe_rec(), probe_rec(failed=1))[0])

    def test_main_moved(self):
        why = tankheal.probe_compare(probe_rec(), probe_rec(main={'tank_onyxia': 'x', 'tank_ragnaros': 'b'}))
        self.assertEqual(why, ['main bench: tank_onyxia'])

    def test_races_moved(self):
        new = probe_rec(races={'Human:race': 'r1', 'Human:off': 'r3', 'Dwarf:race': 'r4'}, values={'Human:race': 1.0, 'Human:off': 0.9801})
        why = tankheal.probe_compare(probe_rec(), new)
        self.assertEqual(why, ['races: 1 of 2 rows differ (score up to 1.00% apart), 1 new rows'])
        self.assertIn('1 rows gone', tankheal.probe_compare(new, probe_rec())[0])


class AuditPick(unittest.TestCase):
    cleared = ['druid_restoration', 'paladin_holy', 'priest_holy', 'shaman_restoration']

    def test_one_cleared_preset_seeded_by_the_commit_and_mode(self):
        a = tankheal.audit_pick(self.cleared, 'c1', 'heal')
        self.assertEqual(len(a), 1)
        self.assertIn(a[0], self.cleared)
        self.assertEqual(a, tankheal.audit_pick(list(reversed(self.cleared)), 'c1', 'heal'))  # order doesn't matter
        picks = {tankheal.audit_pick(self.cleared, f'c{i}', 'heal')[0] for i in range(40)}
        self.assertEqual(picks, set(self.cleared))  # every cleared preset gets its turn over commits
        self.assertNotEqual([tankheal.audit_pick(self.cleared, f'c{i}', 'heal') for i in range(10)],
                            [tankheal.audit_pick(self.cleared, f'c{i}', 'tank') for i in range(10)])

    def test_none_cleared(self):
        self.assertEqual(tankheal.audit_pick([], 'c1', 'heal'), [])


class ProbeAudit(unittest.TestCase):
    """publish's checks after a probe check: tankheal.py probe audit and probe passed, on stand-in packs."""
    stems = ['paladin_holy', 'priest_holy', 'shaman_restoration']
    rows = {'paladin_holy': '177 | 1.018', 'priest_holy': '284 | 0.988', 'shaman_restoration': '293 | 0.946'}

    def setUp(self):
        self.d = tempfile.mkdtemp()
        self.saved = (tankheal.D, tankheal.CHECK_DIR, tankheal.pack_files, tankheal.presets, tankheal.head)
        tankheal.D, tankheal.CHECK_DIR = self.d, os.path.join(self.d, 'probe_check', 'bench')
        races = [(s, r) for s in self.stems for r in ('Human', 'Dwarf')]
        self.old = heal_texts(self.rows, races)
        tankheal.pack_files = lambda mode: self.old if mode == 'heal' else None
        tankheal.presets = lambda mode: self.stems if mode == 'heal' else []
        tankheal.head = lambda: 'c1'
        self.out = os.path.join(self.d, 'heal')
        self.write_run(self.old, simmed=['priest_holy'])
        self.check({'commit': 'c1', 'cleared': ['paladin_holy', 'shaman_restoration'], 'audit': ['priest_holy'], 'moved': [], 'why': {}})

    def tearDown(self):
        tankheal.D, tankheal.CHECK_DIR, tankheal.pack_files, tankheal.presets, tankheal.head = self.saved

    def write_run(self, files, simmed):
        os.makedirs(self.out, exist_ok=True)
        for n, t in files.items():
            open(os.path.join(self.out, n), 'w').write(t)
        open(os.path.join(self.out, '.fresh'), 'w').write('c1\n' + ','.join(simmed) + '\n')

    def check(self, a):
        os.makedirs(os.path.join(tankheal.CHECK_DIR, 'heal'), exist_ok=True)
        json.dump(a, open(os.path.join(tankheal.CHECK_DIR, 'heal', 'audit.json'), 'w'))

    def audit(self):
        with contextlib.redirect_stdout(io.StringIO()):
            tankheal.probe_audit()

    def test_unchanged_passes_and_records_the_cleared(self):
        self.audit()
        self.assertEqual(tankheal.probe_passed(), ['heal:paladin_holy,shaman_restoration'])

    def test_audit_preset_races_moved(self):
        new = dict(self.old)
        new['races.jsonl'] = new['races.jsonl'].replace('"priest_holy", "race": "Dwarf"', '"priest_holy", "race": "Dwarf", "carry": 1')
        self.write_run(new, simmed=['priest_holy'])
        with self.assertRaisesRegex(SystemExit, 'FAILED: priest_holy'):
            self.audit()

    def test_cleared_preset_main_rows_moved(self):
        new = dict(self.old)
        new['tank_onyxia.txt'] = new['tank_onyxia.txt'].replace('| 177 | 1.018 |', '| 177 | 1.019 |')
        self.write_run(new, simmed=['priest_holy', 'paladin_holy'])
        with self.assertRaisesRegex(SystemExit, 'FAILED: paladin_holy'):
            self.audit()

    def test_audit_preset_not_resimmed(self):
        self.write_run(self.old, simmed=['paladin_holy'])
        with self.assertRaisesRegex(SystemExit, 'audit preset first'):
            self.audit()
        os.remove(os.path.join(self.out, '.fresh'))
        with self.assertRaisesRegex(SystemExit, 'audit preset first'):
            self.audit()

    def test_passed_leaves_out_what_was_simmed_and_another_commits_check(self):
        self.write_run(self.old, simmed=['priest_holy', 'shaman_restoration'])
        self.assertEqual(tankheal.probe_passed(), ['heal:paladin_holy'])
        tankheal.head = lambda: 'c2'
        with contextlib.redirect_stderr(io.StringIO()):
            self.assertEqual(tankheal.probe_passed(), [])


class ProbeReuse(unittest.TestCase):
    """probe_presets keeps a record with the current key, copies one from a reuse source, and probes the rest."""

    def setUp(self):
        self.d = tempfile.mkdtemp()
        self.saved = (tankheal.D, tankheal.probe_key, tankheal.subprocess)
        tankheal.D = self.d
        tankheal.probe_key = lambda mode, stem, settings: 'key-' + stem
        self.ran = []
        test = self

        class Sub:
            @staticmethod
            def run(cmd, **kw):
                test.ran.append(cmd[4])
                json.dump({'main': {'b': cmd[4]}, 'races': {}, 'values': {}}, open(os.path.join(cmd[5], 'probe.json'), 'w'))
                return type('R', (), {'returncode': 0, 'stdout': '', 'stderr': ''})()
        tankheal.subprocess = Sub

    def tearDown(self):
        tankheal.D, tankheal.probe_key, tankheal.subprocess = self.saved

    def probe(self, stems, reuse=()):
        with contextlib.redirect_stdout(io.StringIO()):
            return tankheal.probe_presets('heal', stems, os.path.join(self.d, 'heal', 'probe'), reuse)

    def test_probes_then_keeps_then_copies(self):
        got = self.probe(['a', 'b'])
        self.assertEqual(sorted(self.ran), ['a', 'b'])
        self.assertEqual(got['a']['_key'], 'key-a')
        self.assertEqual(json.load(open(os.path.join(self.d, 'heal', 'probe', 'a.json'))), got['a'])
        self.ran.clear()
        self.probe(['a', 'b'])
        self.assertEqual(self.ran, [])  # same keys: kept
        src = {'c': dict(got['a'], _key='key-c'), 'd': dict(got['a'], _key='old')}
        self.probe(['c', 'd'], reuse=(src,))
        self.assertEqual(self.ran, ['d'])  # c copied under its key, d's source has another key
        self.assertEqual(json.load(open(os.path.join(self.d, 'heal', 'probe', 'c.json')))['_key'], 'key-c')


class ShortVectors(unittest.TestCase):
    """bench_races.fight_values: a bench handing back fewer fights than asked no longer fails."""

    def test_full(self):
        v = {'a': [1, 2], 'b': [3, 4]}
        self.assertEqual(tankheal.bench_races.fight_values(v, ('a', 'b'), 2, 'x'), [1, 3, 2, 4])

    def test_short_is_the_fights_returned(self):
        v = {'a': [1, 2], 'b': [3, 4]}
        with contextlib.redirect_stderr(io.StringIO()) as err:
            self.assertEqual(tankheal.bench_races.fight_values(v, ('a', 'b'), 10, 'x'), [1, 3, 2, 4])
        self.assertIn('2 fights of 10', err.getvalue())

    def test_bad_vectors_raise(self):
        f = tankheal.bench_races.fight_values
        for v, n in (({'a': [1, 2], 'b': [3]}, 2), ({'a': [], 'b': []}, 2), ({'a': [1, 2, 3], 'b': [1, 2, 3]}, 2), ({'a': [1]}, 2)):
            with self.assertRaises(RuntimeError, msg=str(v)):
                f(v, ('a', 'b'), n, 'x')

    def test_adaptive_values_keeps_the_share_asked(self):
        import adaptive
        old = adaptive.D
        adaptive.D = tempfile.mkdtemp()
        try:
            def sim(seed, count):  # a run that hands back half its fights
                k = count // 2
                return tankheal.bench_races.fight_values({'a': [seed + i for i in range(k)], 'b': [0.0] * k}, ('a', 'b'), count, 'x')
            with contextlib.redirect_stderr(io.StringIO()):
                got = adaptive.values('bench_t', {'r': 1}, {}, 10, sim, width=2)
                self.assertEqual(len(got), 10)  # 5 fights of 2 values
                got = adaptive.values('bench_t', {'r': 1}, {}, 20, sim, width=2)
            self.assertEqual(len(got), 20)
        finally:
            adaptive.D = old


if __name__ == '__main__':
    unittest.main()
