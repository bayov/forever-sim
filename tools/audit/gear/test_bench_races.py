#!/usr/bin/env python3
"""Every tank and healer bench request plays a race its class can be in Forever (run_build.py CLASS_RACES, from
the client's CharBaseInfo). tank_fairness-20261105-2: the Warden presets played Night Elf and Undead.

Run from anywhere: python3 tools/audit/gear/test_bench_races.py (no sims run)."""
import glob, json, os, sys, unittest

ROOT = os.path.abspath(os.path.join(os.path.dirname(os.path.abspath(__file__)), '..', '..', '..'))
sys.path.insert(0, os.path.join(ROOT, 'tools', 'dps_report'))
from run_build import CLASS_RACES  # noqa: E402


class BenchRaces(unittest.TestCase):
    def test_every_bench_request_race_is_legal(self):
        files = sorted(glob.glob(os.path.join(ROOT, 'tools', 'tankbench', 'req', '*.json')) +
                       glob.glob(os.path.join(ROOT, 'tools', 'healbench', 'req', '*.json')))
        self.assertTrue(files)
        for f in files:
            with self.subTest(req=os.path.relpath(f, ROOT)):
                p = json.load(open(f))['raid']['parties'][0]['players'][0]
                cls = p['class'].removeprefix('Class')
                race = p['race'].removeprefix('Race')
                self.assertIn(race, CLASS_RACES[cls], '%s can\'t be %s' % (cls, race))


if __name__ == '__main__':
    unittest.main()
