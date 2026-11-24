#!/usr/bin/env python3
"""The gear pass's mechanical checks on the Phase 0 lists (orchestration/standards.md 43, passes/gear.md).

Run from the repo root: python3 tools/audit/gear/phase0_check.py [--noise 1.0] [--max-gpd 30]

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


def load(p):
    with open(p) as f:
        return json.load(f)


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument('--noise', type=float, default=1.0, help='DPS an open item may sit above the pick (default 1.0)')
    ap.add_argument('--max-gpd', type=float, default=float(os.environ.get('MAX_GOLD_PER_DPS', 30)))
    a = ap.parse_args()
    specs = load('tools/rotopt/phase0_specs.json')['specs']
    content = load('tools/rotopt/phase0_content.json')
    enchants = load('tools/rotopt/phase0_enchants.json')
    bad = 0

    def flag(msg):
        nonlocal bad
        bad += 1
        print('! ' + msg)

    def item_ok(where, iid, faction):
        c = content.get(str(iid))
        if c is None:
            return flag(f'{where}: item {iid} is not in phase0_content.json')
        if c.get('avail') in ('out', 'infeasible'):
            flag(f"{where}: {c.get('name')} ({iid}) is {c['avail']}: {c.get('note', '')}")
        if (c.get('goldPerDps') or 0) > a.max_gpd:
            flag(f"{where}: {c.get('name')} ({iid}) costs {c['goldPerDps']}g per DPS, over the {a.max_gpd:g}g line")
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
                item_ok(f'{spec} {slot}', r['id'], fac)
            best = min((r['loss'] for r in pk), default=0)
            for r in rows:
                c = content.get(str(r['id'])) or {}  # today's tag, and the one the list copied when it ran
                # an item tagged infeasible (Exalted reputation, or over the gold line when the list ran) ranks but is
                # never a pick, by design (bis_phase0.sh -gear-no-pick): sitting above the pick is not a finding
                if (r.get('bis') or c.get('avail') in ('infeasible', 'out') or r.get('avail') in ('infeasible', 'out')
                        or (c.get('goldPerDps') or 0) > a.max_gpd):
                    continue
                if r['loss'] < best - a.noise:
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
        for it in ov['equipment'].get('items') or []:
            if it and it.get('id'):
                item_ok(f'{bid} as {race}', it['id'], fac)
    print(f'{bad} finding(s)')


if __name__ == '__main__':
    main()
