#!/usr/bin/env python3
"""Tests for phase0_prices.py: python3 tools/rotopt/test_phase0_prices.py"""
import contextlib
import io
import os
import sys
import tempfile
import unittest

HERE = os.path.dirname(os.path.abspath(__file__))
sys.path.insert(0, HERE)
import phase0_prices  # noqa: E402


def entry(iid, name, loss, pick=False):
    return {'id': iid, 'name': name, 'loss': loss, **({'bis': True} if pick else {})}


def gpd(slot, entries, prices, content=None, recipes=None, line=30, effort=None):
    rows = phase0_prices.price_slot(slot, entries, content or {}, prices, recipes or {}, line, effort)
    return {r['name']: (r['gpd'], r['over']) for r in rows if r['gpd'] is not None}


class NewestAhFile(unittest.TestCase):
    def test_main_account_first(self):
        with tempfile.TemporaryDirectory() as wow:
            def save(account, at):
                folder = os.path.join(wow, '_forever_', 'WTF', 'Account', account, 'SavedVariables')
                os.makedirs(folder, exist_ok=True)
                path = os.path.join(folder, 'BayovAuctionData.lua')
                open(path, 'w').close()
                os.utime(path, (at, at))
                return path

            def find():
                out = io.StringIO()
                with contextlib.redirect_stdout(out):
                    return phase0_prices.newest_ah_file(wow), out.getvalue()
            self.assertEqual(find(), (None, ''))
            alt = save('1#2', 100000)
            self.assertEqual(find(), (alt, ''))
            main = save('1#1', 99000)
            self.assertEqual(find(), (main, ''))
            save('1#2', 99000 + 7 * 3600)
            path, said = find()
            self.assertEqual(path, main)
            self.assertIn('1#2 has newer prices', said)


class ParsePrice(unittest.TestCase):
    def test_newest_day(self):
        # Brain Hacker's string on 2026-11-22: 255g today, 215g and up to 498g before.
        rec = '1794942479,2550000,1794942479,2550000,2,1794942479;2512,2550000,2550000,2;2511,,1870000,4;2510,2150000,2150000,4'
        self.assertEqual(phase0_prices.parse_price(rec), 255.0)

    def test_today_in_progress(self):
        self.assertEqual(phase0_prices.parse_price('1,0,1,0,5,1;2514,10000/30000/20000,10000,5'), 2.0)

    def test_skips_a_day_without_a_price(self):
        self.assertEqual(phase0_prices.parse_price('1,,1,,0,1;2512,,500,0;2511,40000,40000,3'), 4.0)

    def test_last_look_when_no_day(self):
        self.assertEqual(phase0_prices.parse_price('1,70000,1,70000,1,1'), 7.0)


class PriceSlot(unittest.TestCase):
    def test_measured_against_the_best_cheaper_item(self):
        entries = [entry(1, 'Belt', 0, True), entry(2, 'Free Belt', 4), entry(3, 'Worse Free Belt', 6)]
        self.assertEqual(gpd('waist', entries, {1: 533.0}), {'Belt': (133.2, 'Free Belt')})

    def test_cheaper_of_auction_house_and_materials(self):
        entries = [entry(1, 'Helm', 0, True), entry(2, 'Coif', 2.5)]
        recipes = {1: [(10, 2), (11, 1)]}
        got = gpd('head', entries, {1: 300.0, 10: 50.0, 11: 25.0}, recipes=recipes)
        self.assertEqual(got, {'Helm': (50.0, 'Coif')})

    def test_unpriced_material_leaves_the_auction_house_price(self):
        got = gpd('head', [entry(1, 'Helm', 0), entry(2, 'Coif', 2)], {1: 100.0}, recipes={1: [(10, 1)]})
        self.assertEqual(got, {'Helm': (50.0, 'Coif')})

    def test_an_item_over_the_line_is_no_base_for_the_next(self):
        # The 200g legs are over the line, so the 300g legs are measured against the free ones.
        entries = [entry(1, 'Best', 0), entry(2, 'Mid', 3), entry(3, 'Free', 4)]
        got = gpd('legs', entries, {1: 300.0, 2: 200.0})
        self.assertEqual(got, {'Mid': (200.0, 'Free'), 'Best': (75.0, 'Free')})

    def test_a_feasible_item_is_the_base_for_the_next(self):
        entries = [entry(1, 'Best', 0), entry(2, 'Mid', 3), entry(3, 'Free', 4)]
        got = gpd('legs', entries, {1: 90.0, 2: 20.0})
        self.assertEqual(got, {'Mid': (20.0, 'Free'), 'Best': (23.3, 'Mid')})

    def test_rings_measure_against_the_second_cheaper(self):
        entries = [entry(1, 'Gold Ring', 0), entry(2, 'Ring A', 1, True), entry(3, 'Ring B', 2, True)]
        self.assertEqual(gpd('finger', entries, {1: 40.0}), {'Gold Ring': (20.0, 'Ring B')})

    def test_no_number_when_a_cheaper_item_is_as_good(self):
        self.assertEqual(gpd('chest', [entry(1, 'Free', 0), entry(2, 'Dear', 0.7)], {2: 270.0}), {})

    def test_doubtful_and_reputation_items(self):
        content = {'2': {'avail': 'doubtful'}, '3': {'avail': 'infeasible'}}
        entries = [entry(1, 'Dear', 0), entry(2, 'Doubtful', 1), entry(3, 'Exalted', 2), entry(4, 'Free', 4)]
        # Neither the doubtful item nor the Exalted one can be the base, so Dear is measured against Free.
        self.assertEqual(gpd('back', entries, {1: 40.0}, content=content), {'Dear': (10.0, 'Free')})



class Effort(unittest.TestCase):
    # Dire Maul North is a 30 min clear in RUN_MINUTES.
    TRIBUTE = [('DireMaul', 'North', 0.38)]

    def test_hours_at_the_default_share_and_overhead(self):
        # 1 / (0.38% x 0.5) = 526 runs of 30 + 20 min, 439h at 30g an hour.
        gold, how = phase0_prices.drop_effort(self.TRIBUTE, 30, 0.5, 20)
        self.assertAlmostEqual(gold, 100 / 0.19 * 50 / 60 * 30, places=3)
        self.assertEqual(how, 'effort: DireMaul North 0.38%, 526 runs, 439h')

    def test_share_and_overhead(self):
        # Alone in the group (share 1) and straight in (no overhead) is half the runs of 30 min, not 50.
        alone, _ = phase0_prices.drop_effort(self.TRIBUTE, 30, 1, 0)
        self.assertAlmostEqual(alone, phase0_prices.drop_effort(self.TRIBUTE, 30, 0.5, 20)[0] / 2 * 30 / 50)

    def test_the_easiest_source(self):
        # A dungeon RUN_MINUTES doesn't know is a 60 min clear. 20% there is 10 runs of 80 min (13.3h),
        # and that beats 10% in the 35 min Service Entrance, 20 runs of 55 min (18.3h).
        sources = [('Stratholme', 'Service Entrance', 10), ('Nowhere', None, 20)]
        self.assertTrue(phase0_prices.drop_effort(sources, 30, 0.5, 20)[1].startswith('effort: Nowhere'))

    def test_no_source(self):
        self.assertIsNone(phase0_prices.drop_effort([], 30, 0.5, 20))

    def test_bind_on_pickup_drop_costs_its_effort(self):
        # Only the effort prices the drop, so the 50g weapon is measured against the free one, and the
        # drop against the 50g one.
        entries = [entry(1, 'Treant', 0, True), entry(2, 'Brain', 2), entry(3, 'Free', 4)]
        effort = {1: phase0_prices.drop_effort(self.TRIBUTE, 30, 0.5, 20)}
        got = gpd('mainhand', entries, {2: 50.0}, effort=effort)
        self.assertEqual(got['Brain'], (25.0, 'Free'))
        self.assertEqual(got['Treant'][1], 'Brain')
        self.assertGreater(got['Treant'][0], 1000)

    def test_bind_on_equip_drop_takes_the_cheaper(self):
        effort = {1: (900.0, 'effort: somewhere')}
        self.assertEqual(phase0_prices.item_gold(1, {1: 255.0}, {}, effort), (255.0, 'AH'))
        self.assertEqual(phase0_prices.item_gold(1, {1: 1000.0}, {}, effort), (900.0, 'effort: somewhere'))

    def test_atlasloot_fills_a_chance_the_wiki_lacks(self):
        # The wiki has Treant's Bane on the tribute chest with no chance, and AtlasLoot on Cho'Rush.
        drops = {18538: [('DireMaul', 179564, 0)]}
        bosses = {179564: ('DireMaul', 'North'), 14324: ('DireMaul', 'North'), 11501: ('DireMaul', 'North')}
        rates = {18538: {14324: 0.38}}
        self.assertEqual(phase0_prices.drop_sources(18538, drops, bosses, rates), [('DireMaul', 'North', 0.38)])

    def test_the_wiki_chance_wins_for_its_boss(self):
        drops = {1: [('Scholomance', 1853, 14)]}
        rates = {1: {1853: 9.0}}
        self.assertEqual(phase0_prices.drop_sources(1, drops, {1853: ('Scholomance', None)}, rates),
                         [('Scholomance', None, 14)])


if __name__ == '__main__':
    unittest.main()
