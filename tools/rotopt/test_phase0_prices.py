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
        # Only we roll (share 1, user 2026-11-25): 1 / 0.38% = 263 runs of 30 + 20 min, 219h at 30g an hour.
        gold, how = phase0_prices.drop_effort(self.TRIBUTE, 30, 1, 20)
        self.assertAlmostEqual(gold, 100 / 0.38 * 50 / 60 * 30, places=3)
        self.assertEqual(how, 'effort: DireMaul North 0.38%, 263 runs, 219h')

    def test_default_roll_share_is_one(self):
        out = io.StringIO()
        with contextlib.redirect_stdout(out), self.assertRaises(SystemExit):
            sys.argv = ['phase0_prices.py', '--help']
            phase0_prices.main()
        self.assertIn('1 when only we roll', out.getvalue())

    def test_share_and_overhead(self):
        # One other player rolling (share 0.5) doubles the runs; no overhead makes them 30 min, not 50.
        shared, _ = phase0_prices.drop_effort(self.TRIBUTE, 30, 0.5, 0)
        self.assertAlmostEqual(shared, phase0_prices.drop_effort(self.TRIBUTE, 30, 1, 20)[0] * 2 * 30 / 50)

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


class ApplyPrices(unittest.TestCase):
    def row(self, name, gold, gpd):
        return {'name': name, 'gold': gold, 'gpd': gpd}

    def test_keeps_items_off_the_list(self):
        content = {
            '1263': {'name': 'Brain Hacker', 'gold': 255.0, 'goldPerDps': {'enh': 10.7}},
            '22403': {'name': 'Nacreous Shell Necklace', 'avail': 'infeasible', 'gold': 282.6, 'goldPerDps': {'enh': 217.4}},
            '13098': {'name': 'Painweaver Band', 'gold': 347.8, 'goldPerDps': {'enh': 483.1}},
        }
        priced = {'enh': {'1263': self.row('Brain Hacker', 255.0, 7.4), '2915': self.row('Taran Icebreaker', 99.04, 21.9)}}
        changed = phase0_prices.apply_prices(content, priced, {'enh': {'1263', '2915', '13098'}})
        self.assertEqual(changed, 3)
        self.assertEqual(content['1263'], {'name': 'Brain Hacker', 'gold': 255.0, 'goldPerDps': {'enh': 7.4}})
        self.assertEqual(content['2915'], {'name': 'Taran Icebreaker', 'gold': 99.0, 'goldPerDps': {'enh': 21.9}})
        # On the list with no goldPerDps now: its old one goes.
        self.assertEqual(content['13098'], {'name': 'Painweaver Band'})
        # Not on the list: price and avail stay.
        self.assertEqual(content['22403'], {'name': 'Nacreous Shell Necklace', 'avail': 'infeasible',
                                            'gold': 282.6, 'goldPerDps': {'enh': 217.4}})

    def test_every_list_takes_its_own_number(self):
        # Elemental Mage Staff: over the line for Fire, under it for Frost, not on the Arcane list's top, and
        # the Enhancement number from before the lists were priced by spec stays when enh doesn't run.
        content = {'944': {'name': 'Elemental Mage Staff', 'gold': 2000.0, 'goldPerDps': 7.0}}
        priced = {'mage_fire': {'944': self.row('Elemental Mage Staff', 249.0, 19.9)},
                  'mage_frost': {'944': self.row('Elemental Mage Staff', 249.0, 41.5)},
                  'mage_arcane': {'944': self.row('Elemental Mage Staff', 249.0, None)}}
        on_list = {'mage_fire': {'944'}, 'mage_frost': {'944'}, 'mage_arcane': {'944'}}
        self.assertEqual(phase0_prices.apply_prices(content, priced, on_list), 1)
        self.assertEqual(content['944'], {'name': 'Elemental Mage Staff', 'gold': 249.0,
                                          'goldPerDps': {'enh': 7.0, 'mage_fire': 19.9, 'mage_frost': 41.5}})
        self.assertEqual(phase0_prices.spec_gpd(content['944'], 'mage_frost'), 41.5)
        self.assertIsNone(phase0_prices.spec_gpd(content['944'], 'mage_arcane'))
        # Its last number gone, the gold goes too.
        phase0_prices.apply_prices(content, {'enh': {}}, {'enh': {'944'}, 'mage_fire': {'944'}, 'mage_frost': {'944'}})
        self.assertEqual(content['944'], {'name': 'Elemental Mage Staff'})

    def test_write_keeps_blackblade_of_shahram_price(self):
        # The ret list gives Blackblade no number (Destiny is better and cheaper), so a --write pass drops it;
        # the hand price in HAND_PRICES puts it back, and the committed content already has it.
        import json
        with open(phase0_prices.CONTENT) as f:
            content = json.load(f)
        self.assertEqual(phase0_prices.apply_prices(content, {'ret': {}}, {'ret': {'12592'}}), 1)
        self.assertNotIn('goldPerDps', content['12592'])
        self.assertEqual(phase0_prices.apply_hand_prices(content), 1)
        self.assertEqual(content['12592']['gold'], 2500.0)
        self.assertEqual(phase0_prices.spec_gpd(content['12592'], 'ret'), 424.4)
        # idempotent, and other specs' numbers on the item stay
        self.assertEqual(phase0_prices.apply_hand_prices(content), 0)
        content['12592']['goldPerDps']['enh'] = 9.0
        phase0_prices.apply_hand_prices(content)
        self.assertEqual(content['12592']['goldPerDps'], {'enh': 9.0, 'ret': 424.4})

    def test_committed_content_matches_hand_prices(self):
        import json
        with open(phase0_prices.CONTENT) as f:
            content = json.load(f)
        self.assertEqual(phase0_prices.apply_hand_prices(content), 0)

    def test_every_spec_has_a_list(self):
        import json
        with open(phase0_prices.SPECS) as f:
            specs = json.load(f)['specs']
        for spec, row in specs.items():
            self.assertTrue(os.path.exists(phase0_prices.list_path(spec, row)), spec)

    def test_content_text_one_entry_per_line(self):
        text = phase0_prices.content_text({'22403': {'name': 'Nacreous Shell Necklace'}, '1263': {'name': 'Brain Hacker'}})
        self.assertEqual(text, '{\n  "1263": {"name": "Brain Hacker"},\n  "22403": {"name": "Nacreous Shell Necklace"}\n}\n')


class FactionSets(unittest.TestCase):
    def test_set_paths_and_gold(self):
        with tempfile.TemporaryDirectory() as root:
            folder = os.path.join(root, 'ui', 'warrior', 'gear_sets')
            os.makedirs(folder)
            table = {'x': {'dir': 'warrior', 'gearOut': 'ui/warrior/gear_sets/level60_phase0_x.gear.json'},
                     'x_dw': {'dir': 'warrior'}}
            for name in ('x', 'x_alliance', 'x_shock', 'x_dw', 'x_dw_alliance'):
                open(os.path.join(folder, f'level60_phase0_{name}.gear.json'), 'w').close()
            got = phase0_prices.set_paths(root, 'x', table['x'], table)
            self.assertEqual([label for label, _ in got], ['x', 'x_alliance', 'x_shock'])
            self.assertEqual(got[0][1], os.path.join(root, table['x']['gearOut']))
            # x_dw's sets are its own, found from its default path
            self.assertEqual([label for label, _ in phase0_prices.set_paths(root, 'x_dw', table['x_dw'], table)],
                             ['x_dw', 'x_dw_alliance'])
        items = [{'id': 1}, {'id': 2}, {'id': 3}]
        total, parts = phase0_prices.set_gold(items, {1: 10.0}, {2: [(1, 2)]}, {3: (5.0, 'effort: run')},
                                              names={'1': 'A'})
        self.assertEqual(total, 35.0)
        self.assertEqual(parts, ['A 10g (AH)', '2 20g (materials)', '3 5g (effort)'])

    def test_variant_lists(self):
        with tempfile.TemporaryDirectory() as root:
            folder = os.path.join(root, 'ui', 'warrior', 'bis')
            os.makedirs(folder)
            table = {'x': {'dir': 'warrior'}, 'x_dw': {'dir': 'warrior'}}
            for name in ('x', 'x_alliance', 'x_horde', 'x_dw', 'x_dw_alliance'):
                open(os.path.join(folder, f'level60_phase0_{name}.bis.json'), 'w').close()
            old, phase0_prices.ROOT = phase0_prices.ROOT, root
            try:
                got = phase0_prices.variant_lists(root, 'x', table['x'], table)
                self.assertEqual([l for l, _ in got], ['x_alliance', 'x_horde'])
                self.assertEqual([l for l, _ in phase0_prices.variant_lists(root, 'x_dw', table['x_dw'], table)],
                                 ['x_dw_alliance'])
            finally:
                phase0_prices.ROOT = old


if __name__ == '__main__':
    unittest.main()
