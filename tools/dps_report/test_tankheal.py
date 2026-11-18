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


if __name__ == '__main__':
    unittest.main()
