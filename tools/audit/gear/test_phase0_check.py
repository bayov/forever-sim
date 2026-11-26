#!/usr/bin/env python3
"""Tests for phase0_check.py: python3 tools/audit/gear/test_phase0_check.py"""
import contextlib
import io
import json
import os
import sys
import tempfile
import unittest

HERE = os.path.dirname(os.path.abspath(__file__))
sys.path.insert(0, HERE)
import phase0_check  # noqa: E402


def write(path, data):
    os.makedirs(os.path.dirname(path), exist_ok=True)
    with open(path, 'w') as f:
        json.dump(data, f)


class GoldLine(unittest.TestCase):
    def test_dps_row(self):
        self.assertEqual(phase0_check.gold_line({'dir': 'x'}, 30, 3000), 30)

    def test_bench_row(self):
        # 3000g per 1% of a 3000 score is 100g per point
        self.assertEqual(phase0_check.gold_line({'bench': 'tank', 'scoreRef': 3000}, 30, 3000), 100)


class Noise(unittest.TestCase):
    def test_dps_row(self):
        self.assertEqual(phase0_check.noise_of({'dir': 'x'}, 1.0), 1.0)

    def test_bench_row(self):
        # 0.1% of the scoreRef per 1.0 of noise
        self.assertAlmostEqual(phase0_check.noise_of({'bench': 'tank', 'scoreRef': 3000}, 1.0), 15.0)
        self.assertAlmostEqual(phase0_check.noise_of({'bench': 'heal', 'scoreRef': 130}, 2.0), 3.9)


class Run(unittest.TestCase):
    """One DPS and one tank spec, each picking an item at 50g per unit of its list."""

    def run_check(self, tank_gpd, dps_gpd, tank_list=True):
        with tempfile.TemporaryDirectory() as d:
            specs = {'enh': {'dir': 'enh'},
                     'tank_a': {'dir': 'tank', 'bench': 'tank', 'scoreRef': 3000, 'reqFile': 'req/tank_a.json'}}
            write(f'{d}/tools/rotopt/phase0_specs.json', {'specs': specs})
            write(f'{d}/tools/rotopt/phase0_content.json', {
                '1': {'name': 'Tank Pick', 'goldPerDps': {'tank_a': tank_gpd}},
                '2': {'name': 'Dps Pick', 'goldPerDps': {'enh': dps_gpd}}})
            write(f'{d}/tools/rotopt/phase0_enchants.json', {})
            for spec, iid in (('enh', 2), ('tank_a', 1)):
                if spec == 'tank_a' and not tank_list:
                    continue
                dr = specs[spec]['dir']
                write(f'{d}/ui/{dr}/bis/level60_phase0_{spec}.bis.json',
                      {'slots': {'head': [{'id': iid, 'name': 'x', 'loss': 0, 'bis': True}]}})
                write(f'{d}/ui/{dr}/gear_sets/level60_phase0_{spec}.gear.json', {'items': [{'id': iid}]})
            write(f'{d}/req/tank_a.json', {'raid': {'parties': [{'players': [{'equipment': {'items': [{'id': 1}]}}]}]}})
            cwd, argv = os.getcwd(), sys.argv
            out = io.StringIO()
            try:
                os.chdir(d)
                sys.argv = ['phase0_check.py']
                with contextlib.redirect_stdout(out):
                    phase0_check.main()
            finally:
                os.chdir(cwd)
                sys.argv = argv
            return [l for l in out.getvalue().splitlines() if l.startswith('!')]

    def test_tank_under_its_line(self):
        # 50g per point is over the 30g DPS line but under the tank's 100g line
        self.assertEqual(self.run_check(50, 20), [])

    def test_tank_over_its_line(self):
        found = self.run_check(150, 20)
        self.assertEqual(len(found), 1)
        self.assertIn('per point of score, over the 100g line', found[0])

    def test_dps_line_unchanged(self):
        found = self.run_check(50, 50)
        self.assertEqual(len(found), 1)
        self.assertIn('Dps Pick (2) costs 50g per DPS, over the 30g line', found[0])

    def test_missing_bench_list(self):
        self.assertEqual(self.run_check(50, 20, tank_list=False), ['! tank_a: no list at ui/tank/bis/level60_phase0_tank_a.bis.json'])


if __name__ == '__main__':
    unittest.main()
