#!/usr/bin/env python3
"""Tests for apply_checks.py: python3 tools/rotopt/test_apply_checks.py"""
import contextlib
import io
import json
import os
import sys
import tempfile
import unittest

HERE = os.path.dirname(os.path.abspath(__file__))
sys.path.insert(0, HERE)
import apply_checks  # noqa: E402

AB = 'Rutherford Twing · Arathi Basin in Forever unconfirmed'
CONTENT = {
    '20150': {'avail': 'doubtful', 'name': "Defiler's Chain Girdle", 'note': f'Defilers Honored · {AB}'},
    '20158': {'avail': 'infeasible', 'name': "Defiler's Chain Pauldrons", 'note': f'Defilers Exalted · {AB}'},
    '275968': {'avail': 'doubtful', 'name': "Initiate's Signet of Ferocity", 'note': 'Watchers Friendly · no rep source found yet'},
    '275981': {'avail': 'infeasible', 'name': "Warden's Signet of Ferocity", 'note': 'Watchers Exalted · no rep source found yet'},
    '19287': {'avail': 'doubtful', 'name': 'Darkmoon Card: Heroism', 'note': 'Darkmoon Faire deck · Forever card drops unconfirmed'},
    '275740': {'avail': 'doubtful', 'name': "Ring of Nature's Cycles", 'note': 'Quest: Work Well Done'},
    '19325': {'avail': 'infeasible', 'name': "Don Julio's Band", 'note': 'Frostwolf Exalted · AV queue unconfirmed'},
    '50000': {'avail': 'out', 'name': 'Raid Thing', 'note': 'Quest: Work Well Done'},
    '272399': {'name': "Stormseeker's Girdle", 'note': 'Pix Xizzix · 100 Reals'},
}


def flipped(checks, content=None):
    return {iid: part for iid, _, _, part in apply_checks.flips(content or CONTENT, checks)}


class Flips(unittest.TestCase):
    def test_nothing_seen_flips_nothing(self):
        self.assertEqual(flipped({}), {})

    def test_arathi_basin_queued(self):
        got = flipped({'bgs': {'Arathi Basin': {'queued': '2026-11-20', 'listed': '2026-11-19', 'canEnter': True}}})
        self.assertEqual(got, {'20150': 'AB queued Nov 20'})

    def test_listed_but_closed_is_no_answer(self):
        self.assertEqual(flipped({'bgs': {'Arathi Basin': {'listed': '2026-11-19', 'canEnter': False}}}), {})

    def test_defilers_rep_gain_opens_ab_items(self):
        reps = {'The Defilers': {'r': 4, 'v': 200, 'v0': 0, 'top': 4, 'd': '2026-11-21'}}
        self.assertEqual(flipped({'reps': reps}), {'20150': 'Defilers rep gained Nov 21'})

    def test_standing_reached_opens_infeasible_too(self):
        reps = {'The Defilers': {'r': 8, 'v': 42000, 'v0': 0, 'top': 8, 'd': '2026-12-01'}}
        got = flipped({'reps': reps})
        self.assertEqual(got['20158'], 'Exalted reached Dec 1')
        self.assertEqual(got['20150'], 'Honored reached Dec 1')

    def test_watchers_source(self):
        reps = {'The Watchers': {'r': 4, 'v': 0, 'v0': 0, 'top': 4, 'd0': '2026-11-20'}}
        self.assertEqual(flipped({'reps': reps}), {}, 'a faction that never moved is no source')
        reps['The Watchers'].update(v=150, d='2026-11-22')
        self.assertEqual(flipped({'reps': reps}), {'275968': 'Watchers rep gained Nov 22'})

    def test_card_and_quest(self):
        checks = {'cards': {'19262': {'name': 'Five of Warlords', 'd': '2026-11-20'}},
                  'quests': {'275000': {'name': 'Work Well Done', 'took': '2026-11-18'}}}
        self.assertEqual(flipped(checks), {'19287': 'card looted Nov 20', '275740': 'quest seen Nov 18'})

    def test_av_gain_leaves_infeasible(self):
        self.assertEqual(flipped({'bgs': {'Alterac Valley': {'inside': '2026-11-20'}}}), {})


class Write(unittest.TestCase):
    def test_write_keeps_notes_and_layout(self):
        with tempfile.TemporaryDirectory() as d:
            content = os.path.join(d, 'content.json')
            with open(content, 'w') as f:
                f.write(apply_checks.dump(CONTENT))
            export = os.path.join(d, 'export.json')
            with open(export, 'w') as f:
                json.dump({'who': {'name': 'Bayov'}, 'checks': {
                    'date': '2026-11-20', 'since': '2026-11-20', 'reals': {'n': 80, 'top': 80, 'd0': '2026-11-20'},
                    'bgs': {'Arathi Basin': {'inside': '2026-11-20'}}}}, f)
            out = io.StringIO()
            with contextlib.redirect_stdout(out):
                self.assertEqual(apply_checks.main([export, '--write', '--content', content]), 0)
            self.assertIn('20150 Defiler\'s Chain Girdle: doubtful -> open · Arathi Basin inside (2026-11-20)', out.getvalue())
            self.assertIn('Reals: 80 held', out.getvalue())
            with open(content) as f:
                raw = f.read()
            got = json.loads(raw)
            self.assertNotIn('avail', got['20150'])
            self.assertEqual(got['20150']['note'], f'Defilers Honored · {AB} · AB inside Nov 20')
            self.assertEqual(got['20158']['avail'], 'infeasible')
            self.assertEqual(raw, apply_checks.dump(got))

    def test_no_checks_section(self):
        with tempfile.TemporaryDirectory() as d:
            export = os.path.join(d, 'export.json')
            with open(export, 'w') as f:
                json.dump({'who': {}}, f)
            with contextlib.redirect_stdout(io.StringIO()):
                self.assertEqual(apply_checks.main([export]), 1)


if __name__ == '__main__':
    unittest.main()
