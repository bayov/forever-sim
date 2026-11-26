#!/usr/bin/env python3
"""The gear pass's mechanical checks on the Phase 0 lists (orchestration/standards.md 43, passes/gear.md).

Run from the repo root: python3 tools/audit/gear/phase0_check.py [--noise 1.0] [--max-gpd 30] [--max-gold-per-pct 3000]

For every spec in tools/rotopt/phase0_specs.json it reads the list (ui/<dir>/bis/...), its gear set and the content
checks (tools/rotopt/phase0_content.json, phase0_enchants.json), and for every DPS request in tools/dps_report/req
and every race override in req/races the gear the report sims. Each line starting with '!' is a finding to read:

  - a pick or enchant the content check doesn't list, marks "out" or "infeasible", or puts over the gold line;
  - a pick or enchant of the other faction (the content check's "faction" against the list's faction, or a race
    override's race);
  - the gear set not matching the list's picks and enchants;
  - an open item more than --noise DPS above the slot's pick: the search stopped short there. An item tagged
    "infeasible" or "out" (by the content check today or on the list's own row) or over the gold line ranks but is
    never a pick, so it may sit above the pick (Nacreous Shell Necklace above the casters' neck picks);
  - a request whose gearSet isn't a Phase 0 set, has a hit-trade step (_hit), or whose equipment isn't that set's.

A tank or healer row (bench in phase0_specs.json, 2026-11-25) is scored in its report's score, not DPS, so its line
is bis_phase0.sh's: --max-gold-per-pct (3000g) per 1% of the row's scoreRef, 3000 * 100 / scoreRef gold per point of
score, against its own goldPerDps. Its request (the row's reqFile) must wear its gear set, like a DPS request.

It doesn't check sources against the game, set bonuses, hit or the list's dps: the pass does those by hand.
"""
import argparse
import glob
import json
import os
import re

ALLIANCE = {'Human', 'Dwarf', 'Gnome', 'NightElf', 'SkyborneHighOrder'}
HORDE = {'Orc', 'Troll', 'Tauren', 'Undead', 'SkyborneWindshaper'}


def faction_of(race):
    race = re.sub(r'^Race', '', race or '').replace(' ', '')
    return 'Alliance' if race in ALLIANCE else 'Horde' if race in HORDE else None


def noise_of(row, noise):
    """The noise an open item may sit above the pick: noise DPS, or for a bench row noise * scoreRef / 1000 points."""
    if row.get('bench'):
        return noise * row['scoreRef'] / 1000
    return noise


def gold_line(row, max_gpd, max_gold_per_pct):
    """The most gold a spec's list may pay per unit it adds: max_gpd per DPS, or for a tank or healer row (bench)
    max_gold_per_pct per 1% of its scoreRef, as bis_phase0.sh's nopick_of has it."""
    if row.get('bench'):
        return max_gold_per_pct * 100 / row['scoreRef']
    return max_gpd


def unit_of(row):
    return 'point of score' if row.get('bench') else 'DPS'


def load(p):
    with open(p) as f:
        return json.load(f)


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument('--noise', type=float, default=1.0, help='DPS an open item may sit above the pick (default 1.0). A tank or healer row (bench) is scored in '
                    'points of score, so it gets the same noise relative to its scoreRef: --noise * scoreRef / 1000, '
                    '0.1%% of scoreRef per 1.0 (a ~1000 DPS spec has 1 DPS of noise): 3.45 points on the 3454 warrior, '
                    '0.13 on a healer at 130')
    ap.add_argument('--max-gpd', type=float, default=float(os.environ.get('MAX_GOLD_PER_DPS', 30)))
    ap.add_argument('--max-gold-per-pct', type=float, default=float(os.environ.get('MAX_GOLD_PER_PCT', 3000)),
                    help="a tank or healer list's line: gold per 1%% of its scoreRef (default 3000)")
    a = ap.parse_args()
    specs = load('tools/rotopt/phase0_specs.json')['specs']
    content = load('tools/rotopt/phase0_content.json')
    enchants = load('tools/rotopt/phase0_enchants.json')
    bad = 0

    def flag(msg):
        nonlocal bad
        bad += 1
        print('! ' + msg)

    def gpd(c, spec):
        # goldPerDps is one number per spec's list (phase0_prices.py)
        g = c.get('goldPerDps')
        return (g.get(spec) if isinstance(g, dict) else None) or 0

    def line(spec):
        return gold_line(specs.get(spec) or {}, a.max_gpd, a.max_gold_per_pct)

    def item_ok(where, iid, faction, spec):
        c = content.get(str(iid))
        if c is None:
            return flag(f'{where}: item {iid} is not in phase0_content.json')
        if c.get('avail') in ('out', 'infeasible'):
            flag(f"{where}: {c.get('name')} ({iid}) is {c['avail']}: {c.get('note', '')}")
        if gpd(c, spec) > line(spec):
            flag(f"{where}: {c.get('name')} ({iid}) costs {gpd(c, spec)}g per {unit_of(specs.get(spec) or {})}, "
                 f"over the {line(spec):g}g line")
        if faction and c.get('faction') and c['faction'] != faction:
            flag(f"{where}: {c.get('name')} ({iid}) is {c['faction']} only, the build is {faction}")

    def ench_ok(where, e, faction):
        t = enchants.get(str(e.get('spell') or e.get('id')))
        if t is None:
            return flag(f"{where}: enchant {e.get('name')} (spell {e.get('spell')}) is not in phase0_enchants.json")
        if t.get('avail') == 'out':
            flag(f"{where}: enchant {e.get('name')} is out: {t.get('note', '')}")
        if faction and t.get('faction') and t['faction'] != faction:
            flag(f"{where}: enchant {e.get('name')} is {t['faction']} only, the build is {faction}")

    sets = {}  # gear set name -> its items
    for p in glob.glob('ui/*/gear_sets/level60_phase0*.gear.json'):
        sets[os.path.basename(p)[:-len('.gear.json')]] = load(p)['items']
    set_spec = {}  # gear set name -> the spec whose list it wears, for the spec's goldPerDps
    for spec, row in specs.items():
        gp = row.get('gearOut') or f"ui/{row['dir']}/gear_sets/level60_phase0_{spec}.gear.json"
        set_spec[os.path.basename(gp)[:-len('.gear.json')]] = spec
    for spec, row in sorted(specs.items()):
        lp = row.get('out') or f"ui/{row['dir']}/bis/level60_phase0_{spec}.bis.json"
        gp = row.get('gearOut') or f"ui/{row['dir']}/gear_sets/level60_phase0_{spec}.gear.json"
        if not os.path.exists(lp):
            flag(f'{spec}: no list at {lp}')
            continue
        lst = load(lp)
        fac = lst.get('faction')
        picks, ench = {}, {}
        for slot, rows in lst['slots'].items():
            pk = [r for r in rows if r.get('bis')]
            picks[slot] = sorted(r['id'] for r in pk)
            for r in pk:
                item_ok(f'{spec} {slot}', r['id'], fac, spec)
            best = min((r['loss'] for r in pk), default=0)
            for r in rows:
                c = content.get(str(r['id'])) or {}  # today's tag, and the one the list copied when it ran
                # an item tagged infeasible (Exalted reputation, or over the gold line when the list ran) ranks but is
                # never a pick, by design (bis_phase0.sh -gear-no-pick): sitting above the pick is not a finding
                if (r.get('bis') or c.get('avail') in ('infeasible', 'out') or r.get('avail') in ('infeasible', 'out')
                        or gpd(c, spec) > line(spec)):
                    continue
                if r['loss'] < best - noise_of(row, a.noise):
                    flag(f"{spec} {slot}: open {r['name']} ({r['id']}) sits {best - r['loss']:.2f} above the pick")
        for slot, rows in (lst.get('enchants') or {}).items():
            for r in rows:
                if r.get('bis'):
                    ench[slot] = r['id']
                    ench_ok(f'{spec} {slot}', r, fac)
        if not os.path.exists(gp):
            flag(f'{spec}: no gear set at {gp}')
            continue
        items = load(gp)['items']
        have = sorted(it['id'] for it in items if it and it.get('id'))
        want = sorted(i for v in picks.values() for i in v)
        if have != want:
            flag(f'{spec}: gear set {gp} differs from the list picks: only in set {sorted(set(have) - set(want))}, only in list {sorted(set(want) - set(have))}')
        set_ench = {it.get('enchant') for it in items if it and it.get('enchant')}
        missing = sorted(set(ench.values()) - set_ench)
        if missing:
            flag(f'{spec}: gear set {gp} lacks the list enchants {missing}')
        # a tank or healer request (the row's reqFile) wears its gear set, as the DPS requests below do
        rq = row.get('reqFile') if row.get('bench') else None
        if rq and os.path.exists(rq):
            eq = load(rq)['raid']['parties'][0]['players'][0].get('equipment', {}).get('items') or []
            if [(i or {}).get('id') for i in eq] != [(i or {}).get('id') for i in items]:
                flag(f'{spec}: its request {rq} does not wear {gp}')
        if row.get('bench'):
            print(f"{spec}: {lst.get('race')} {fac}, list score {lst.get('dps')}, line {line(spec):.1f}g per point of score, "
                  f"made {lst.get('made')} on sim {lst.get('sim')}")
        else:
            print(f"{spec}: {lst.get('race')} {fac}, list dps {lst.get('dps')}, made {lst.get('made')} on sim {lst.get('sim')}")

    for mp in sorted(glob.glob('tools/dps_report/req/*.meta.json')):
        bid = os.path.basename(mp)[:-len('.meta.json')]
        meta = load(mp)
        gs = meta.get('gearSet') or ''
        if not gs.startswith('level60_phase0') or '_hit' in gs:
            flag(f'{bid}: gearSet {gs} is not a Phase 0 set')
            continue
        rq = load(mp.replace('.meta.json', '.json'))
        eq = rq['raid']['parties'][0]['players'][0].get('equipment', {}).get('items') or []
        if gs in sets and [(i or {}).get('id') for i in eq] != [(i or {}).get('id') for i in sets[gs]]:
            flag(f'{bid}: its equipment is not {gs}')
    for op in sorted(glob.glob('tools/dps_report/req/races/*.json')):
        ov = load(op)
        if not ov.get('equipment'):
            continue
        bid, race = os.path.basename(op)[:-len('.json')].rsplit('.', 1)
        fac = faction_of(race)
        mp = f'tools/dps_report/req/{bid}.meta.json'
        spec = set_spec.get(load(mp).get('gearSet')) if os.path.exists(mp) else None
        for it in ov['equipment'].get('items') or []:
            if it and it.get('id'):
                item_ok(f'{bid} as {race}', it['id'], fac, spec)
    print(f'{bad} finding(s)')


if __name__ == '__main__':
    main()
