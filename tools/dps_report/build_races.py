#!/usr/bin/env python3
"""build_races.py: the Races tab's datasets, from the race sweep (out/*.json) and the race
scenario sims (rv/results.jsonl, run_races.py). build_data.py calls build(); it can also run alone.

race_variants   build × race × scenario: DPS, error, gap to the preset's best race in that scenario
race_breakdown  build × race × part: base stats, each racial, and the weapon swap / target change
race_summary    scenario × scope × role × race: how often a race is best and how far behind it is
race_notes      what each scenario switches on, what each racial part is, a finding per class
"""
import json, math, os, statistics
from collections import defaultdict

from dashpaths import D
RACE_NAMES = {'NightElf': 'Night Elf', 'SkyborneHighOrder': 'High Order', 'SkyborneWindshaper': 'Windshaper'}
HORDE = {'Orc', 'Undead', 'Tauren', 'Troll', 'Windshaper'}
RACE_ICON = {'Human': 'race_human_male', 'Dwarf': 'race_dwarf_male', 'Night Elf': 'race_nightelf_male', 'Gnome': 'race_gnome_male',
             'High Order': 'inv_feather_02', 'Orc': 'race_orc_male', 'Undead': 'race_scourge_male', 'Tauren': 'race_tauren_male',
             'Troll': 'race_troll_male', 'Windshaper': 'ability_druid_flightform'}
VARIANTS = ['baseline', 'beast', 'elemental', 'humanoid', 'weapon', 'off']
TARGET_VARIANTS = ('beast', 'elemental', 'humanoid')
RACE_ORDER = ['Human', 'Dwarf', 'Night Elf', 'Gnome', 'High Order', 'Orc', 'Undead', 'Tauren', 'Troll', 'Windshaper']

# Racial parts the sim models, by race: (part key, racial name, what it does, when it counts)
PART_INFO = {
    'spirit': ('The Human Spirit', 'Spirit increased by 5%.'),
    'sword': ('Sword Specialization', '+2% critical strike chance with all spells and attacks while a sword is equipped.'),
    'mace': ('Mace Specialization', '+1% critical strike chance with all spells and attacks while a mace is equipped.'),
    'axe': ('Axe Specialization', '+1% critical strike chance with all spells and attacks while an axe is equipped.'),
    'elune': ("Elune's Light", '+10% critical strike chance for 15 sec, 3 min cooldown.'),
    'mind': ('Expansive Mind', 'Maximum Mana, Rage or Energy increased by 5%.'),
    'eureka': ('Eureka!', 'The next 3 damaging abilities cost 10% less and deal 10% more damage, 2 min cooldown.'),
    'bloodfury': ('Blood Fury', 'Attack power and spell power increased by 10% for 15 sec, 2 min cooldown.'),
    'touch': ('Touch of the Grave', 'Spells and attacks have a 5% chance (10% for Mages, Priests and Warlocks) to deal Shadow damage equal to 5% of your maximum health and heal you for it. It cannot crit and triggers at most once a second.'),
    'endurance': ('Endurance', 'Health increased by 5% and chance to hit increased by 1%.'),
    'berserking': ('Berserking', 'Attack and casting speed increased by 10% for 10 sec, 3 min cooldown.'),
    'windblessed': ('Wind Blessed', 'Spellcasting, melee and ranged haste increased by 1%.'),
    'beast_dwarf': ('Big Game Hunter', 'Damage dealt versus Beasts increased by 5%.'),
    'beast_troll': ('Beast Slaying', 'Damage dealt versus Beasts increased by 5%.'),
    'insight': ('Elemental Insight', 'Damage dealt versus Elementals increased by 5%.'),
    'chastise': ('Chastise', 'Dwarf Priests: an instant Holy hit on a Humanoid, 2 min cooldown.'),
    'starshards': ('Starshards', 'Night Elf Priests: Arcane damage over 6 sec, 30 sec cooldown.'),
    'darksacrifice': ('Dark Sacrifice', 'Undead Priests: trade health for mana over 15 sec, 10 min cooldown.'),
    'base': ('Base stats', "The race's own Strength, Agility, Stamina, Intellect and Spirit, with every racial left out."),
    'swap': ('Weapon swap', "A weapon racial only works with its weapon type. So for a preset that holds another weapon, we give the race that type of weapon with the same stats and damage. This piece is what the new weapon changes by itself, before its racial: the weapon's speed, and talents that give a different bonus per weapon type."),
    'target': ('Target type', 'What changes against this creature type: conditional racials, plus talents and spells that only work on some types.'),
}

# The tooltip's one line for the parts whose text above runs long; the others show their text as it is.
PART_SHORT = {
    'touch': '5% chance a hit (10% for Mages, Priests and Warlocks): Shadow damage of 5% of max health, at most once a second.',
    'swap': "What holding the racial's weapon type changes by itself, before the racial: the weapon's speed, and talents that give a different bonus per weapon type.",
}

# Each racial's game icon, from ui/core/forever/racials.json. The priest racials are from wowhead, and Dark Sacrifice
# has no known icon.
PART_ICON = {
    'spirit': 'inv_enchant_shardbrilliantsmall', 'sword': 'ability_meleedamage', 'mace': 'inv_hammer_05', 'beast_dwarf': 'inv_weapon_rifle_05',
    'elune': 'spell_nature_moonglow', 'mind': 'inv_enchant_essenceeternallarge', 'eureka': 'inv_gnometoy', 'axe': 'inv_axe_02',
    'bloodfury': 'racial_orc_berserkerstrength', 'touch': 'spell_shadow_fingerofdeath', 'endurance': 'spell_nature_unyeildingstamina',
    'berserking': 'racial_troll_berserk', 'beast_troll': 'inv_misc_pelt_bear_ruin_02', 'windblessed': 'inv_misc_volatileair',
    'insight': 'achievement_raidprimalist_windelemental', 'starshards': 'spell_arcane_starfire', 'chastise': 'spell_holy_chastise',
}

RACE_PARTS = {'Human': ['spirit', 'sword'], 'Dwarf': ['mace', 'beast_dwarf', 'chastise'], 'Night Elf': ['elune', 'starshards'],
              'Gnome': ['mind', 'eureka'], 'Orc': ['axe', 'bloodfury'], 'Undead': ['touch', 'darksacrifice'], 'Tauren': ['endurance'],
              'Troll': ['berserking', 'beast_troll'], 'High Order': ['windblessed', 'insight'], 'Windshaper': ['windblessed', 'insight']}
WEAPON_PART = {'Human': 'sword', 'Dwarf': 'mace', 'Orc': 'axe'}
WEAPON_WORD = {'sword': 'a sword', 'mace': 'a mace', 'axe': 'an axe'}

# Talents whose bonus depends on the weapon type, by class: (name, talent tree, index in the tree, icon, max points,
# what it does in a few words, and per weapon type what the bonus is and how much each point gives).
WEAPON_TALENTS = {
    'Warrior': ('Weaponmaster', 0, 13, 'garrison_weaponupgrade', 5,
                'A bonus that depends on the weapon: critical strike chance with an axe or polearm, armor ignored with a mace or staff, or extra attacks with a sword.', {
        'axe': ('% critical strike chance', 1), 'polearm': ('% critical strike chance', 1),
        'mace': ("% of the target's armor ignored", 3), 'staff': ("% of the target's armor ignored", 3),
        'sword': ('% chance of an extra attack', 1)}),
    'Rogue': ('Hack and Slash', 1, 13, 'inv_sword_27', 5,
              'A bonus that depends on the weapon: extra attacks with an axe or sword, critical strike chance with a dagger or fist weapon, or armor ignored with a mace.', {
        'axe': ('% chance of an extra attack', 1), 'sword': ('% chance of an extra attack', 1),
        'dagger': ('% critical strike chance', 1), 'fist': ('% critical strike chance', 1),
        'mace': ("% of the target's armor ignored", 3)}),
}
HAND_NAME = {14: 'main hand', 15: 'off hand'}


def swap_detail(RR, cls, req, wtype):
    """What a race's weapon swap changes, for the swap's tooltip.

    weapons has one [hand, old type, new type, old speed, new speed] per weapon we swap. A dagger Rogue keeps its main
    hand dagger, so only its off hand shows up. A caster gets its staff's stats on a one-hand weapon, and caster is
    true then, because casters don't attack with the weapon.

    talents has one [name, icon, points, max points, what it does, old type, old bonus, new type, new bonus] per talent
    in the preset whose bonus changes with the new weapon type. A bonus is '' when the type gets none.
    """
    r2, items = RR.swapped(req, cls, wtype)
    new = {it['id']: it for it in items}
    old, eq = RR.weapons(req), RR.player(r2)['equipment']['items']
    caster = cls in RR.CASTER or bool(old[0] and old[0]['weaponType'] == RR.STAFF and cls in ('Shaman', 'Druid'))
    weapons = []
    for k, slot in enumerate((14, 15)):
        it = new.get((eq[slot] or {}).get('id')) if len(eq) > slot and eq[slot] else None
        if it and old[k]:
            hand = 'two-hand' if old[k].get('handType') == 4 and it.get('handType') == 4 else HAND_NAME[slot]
            weapons.append([hand, RR.WT_NAME.get(old[k]['weaponType'], ''), RR.WT_NAME.get(wtype, ''), old[k].get('weaponSpeed'), it.get('weaponSpeed')])
    talents = []
    if cls in WEAPON_TALENTS and not caster:
        name, tree, idx, icon, mx, desc, bonus = WEAPON_TALENTS[cls]
        trees = (RR.player(req).get('talentsString') or '').split('-')
        pts = int(((trees[tree] if len(trees) > tree else '') + '0' * 32)[idx])
        for _, was, now, _, _ in weapons:
            if pts and bonus.get(was) != bonus.get(now):
                fmt = lambda t: f'{bonus[t][1] * pts}{bonus[t][0]}' if t in bonus else ''
                row = [name, icon, pts, mx, desc, was, fmt(was), now, fmt(now)]
                if row not in talents:
                    talents.append(row)
    return weapons, talents, caster


SCENARIOS = [
    ('baseline', 'Generic boss', 'Each preset with its own weapon type, against the generic level 63 boss (no creature type). Weapon racials count only where the preset already holds that weapon type: Human swords (Warrior, Hunter, Combat Swords and Seal Fate Swords rogues) and Orc axes (Enhancement).'),
    ('beast', 'Beast target', 'A Beast boss switches on Troll Beast Slaying and Dwarf Big Game Hunter: 5% more damage.'),
    ('elemental', 'Elemental target', 'An Elemental boss switches on High Order and Windshaper Elemental Insight: 5% more damage.'),
    ('humanoid', 'Humanoid target', 'Priests only: Dwarf Priests cast Chastise, which works on Humanoids only. No other racial cares about Humanoids, so the other presets have no Humanoid numbers here.'),
    ('weapon', 'Racial weapon', "Human, Dwarf and Orc each hold their racial's weapon type (sword, mace, axe) where the class can use it, at the same stats and weapon DPS. Casters move the staff's stats to a one-hand weapon; dagger Rogues change only the off hand. Other races keep the preset's weapon."),
    ('off', 'Racials off', 'Every race on its base stats alone, with all racial passives and abilities left out: what the races differ by before racials.'),
]


def rn(r):
    return RACE_NAMES.get(r, r)


def build(builds, races_rows, stats_rows, dump):
    import run_races as RR
    bmap = {b['id']: b for b in builds}
    # role per preset: ranged when ranged attack power beats attack power, caster when spell power does, melee otherwise
    st = defaultdict(dict)
    for s in stats_rows:
        st[s['build']][s['stat']] = s['value']
    role = {}
    for bid, s in st.items():
        role[bid] = 'Ranged' if s.get('RangedAttackPower', 0) > s.get('AttackPower', 0) else 'Caster' if s.get('SpellPower', 0) > s.get('AttackPower', 0) else 'Melee'

    res = defaultdict(dict)  # (build, race display) -> variant -> (dps, se)
    for line in open(os.path.join(D, 'rv', 'results.jsonl')):
        x = json.loads(line)
        res[(x['build'], rn(x['race']))][x['variant']] = (x['dps'], x['se'])
    for o in sorted(os.listdir(os.path.join(D, 'out'))):
        if not o.endswith('.json'):
            continue
        out = json.load(open(os.path.join(D, 'out', o)))
        for r in out['races']:
            res[(out['id'], rn(r['race']))]['baseline'] = (r['dps'], r['se'])

    variants, breakdown, notes = [], [], []
    for bid, b in list(bmap.items()):
        req = json.load(open(os.path.join(D, 'req', bid + '.json')))
        cls = b['class']
        if b.get('encounter'):
            del bmap[bid]  # the encounter presets have only the race sweep; the scenarios are on the generic boss
            continue
        if os.environ.get('RACES_PARTIAL') and any(v not in res[(bid, rn(rc))] for rc in RR.CLASS_RACES[cls] for v in ('off', 'beast', 'elemental')):
            del bmap[bid]  # a test build on part of the sims
            continue
        held = {w['weaponType'] for w in RR.weapons(req) if w}
        held_names = {RR.WT_NAME.get(t, '') for t in held}
        rows = []
        for rc in RR.CLASS_RACES[cls]:
            race = rn(rc)
            v = res[(bid, race)]
            base = v['baseline']
            for var in VARIANTS:
                note = ''
                if var == 'weapon':
                    wp = WEAPON_PART.get(race)
                    if 'weapon' in v:
                        val, sim = v['weapon'], True
                    else:
                        val, sim = base, False
                        note = ('no weapon racial' if not wp else 'already holds ' + WEAPON_WORD[wp][2:].split()[-1] if RR.RACE_WEAPON[rc][0] in held
                                else f"{cls}s can't use {wp}s")
                elif var == 'baseline':
                    val, sim = base, True
                else:
                    val, sim = v.get(var, (None, None)), var in v
                if var == 'humanoid' and not sim:
                    continue  # only the classes with a Humanoid racial (Priests) have this scenario
                if val[0] is None:
                    raise SystemExit(f'missing {bid} {race} {var}')
                rows.append({'build': bid, 'class': cls, 'race': race, 'faction': 'Horde' if race in HORDE else 'Alliance', 'variant': var,
                             'dps': round(val[0], 1), 'se': round(val[1], 2), 'isPreset': race == b['race'], 'simmed': sim, 'note': note,
                             'raceIcon': RACE_ICON.get(race)})
            # breakdown: base stats, then each racial part (Shapley split when two parts count, so the parts add up exactly)
            off = v['off']
            ps = [p for p in RR.parts_for(cls, rc, req)]
            contrib = {}
            if len(ps) == 1:
                contrib[ps[0]] = (base[0] - off[0], math.hypot(base[1], off[1]))
            elif len(ps) == 2:
                a, c = ps
                oa, oc = v['off:' + a], v['off:' + c]
                contrib[a] = (0.5 * ((base[0] - oa[0]) + (oc[0] - off[0])), math.hypot(base[1], off[1]))
                contrib[c] = (0.5 * ((base[0] - oc[0]) + (oa[0] - off[0])), math.hypot(base[1], off[1]))
            elif len(ps) > 2:
                raise SystemExit(f'three parts {bid} {race}')
            breakdown.append({'build': bid, 'race': race, 'scenario': 'all', 'part': 'base', 'label': PART_INFO['base'][0], 'dps': round(off[0], 2),
                              'se': round(off[1], 2), 'active': True, 'note': ''})
            sim_part = {'sword': 'sword', 'mace': 'mace', 'axe': 'axe'}
            for p in RACE_PARTS[race]:
                key = {'beast_dwarf': 'beast', 'beast_troll': 'beast'}.get(p, p)
                if p in ('chastise', 'starshards', 'darksacrifice') and cls != 'Priest':
                    continue
                if key in contrib:
                    d, s = contrib[key]
                    breakdown.append({'build': bid, 'race': race, 'scenario': 'all', 'part': p, 'label': PART_INFO[p][0], 'dps': round(d, 2), 'se': round(s, 2), 'active': True, 'note': ''})
                else:
                    if p in sim_part:
                        wt = RR.WEAPON_PART[p]
                        nt = f"{cls}s can't use {p}s" if wt not in RR.LEGAL[cls] else f'needs {WEAPON_WORD[p]}; the preset holds ' + (' and '.join(sorted(held_names)) or 'no weapon')
                    elif p.startswith('beast'):
                        nt = 'only against Beasts'
                    elif p == 'insight':
                        nt = 'only against Elementals'
                    else:
                        nt = 'no effect on this preset'
                    breakdown.append({'build': bid, 'race': race, 'scenario': 'all', 'part': p, 'label': PART_INFO[p][0], 'dps': 0, 'se': 0, 'active': False, 'note': nt})
            if 'weapon' in v:
                wn, w = v['weaponNoRacial'], v['weapon']
                wp = WEAPON_PART[race]
                sw, st, sc = swap_detail(RR, cls, req, RR.WEAPON_PART[wp])
                breakdown.append({'build': bid, 'race': race, 'scenario': 'weapon', 'part': 'swap', 'label': PART_INFO['swap'][0], 'dps': round(wn[0] - base[0], 2),
                                  'se': round(math.hypot(wn[1], base[1]), 2), 'active': True, 'note': f'{WEAPON_WORD[wp]} in place of the preset weapon',
                                  'weapons': sw, 'talents': st, 'caster': sc})
                breakdown.append({'build': bid, 'race': race, 'scenario': 'weapon', 'part': wp, 'label': PART_INFO[wp][0], 'dps': round(w[0] - wn[0], 2),
                                  'se': round(math.hypot(w[1], wn[1]), 2), 'active': True, 'note': f'with {WEAPON_WORD[wp]}'})
            for t in TARGET_VARIANTS:
                if t not in v:
                    continue
                tv = v[t]
                breakdown.append({'build': bid, 'race': race, 'scenario': t, 'part': 'target', 'label': {'beast': 'Beast target', 'elemental': 'Elemental target', 'humanoid': 'Humanoid target'}[t],
                                  'dps': round(tv[0] - base[0], 2), 'se': round(math.hypot(tv[1], base[1]), 2), 'active': True, 'note': ''})
        # gap to the best race within each scenario, and the rank
        for var in VARIANTS:
            vr = sorted([r for r in rows if r['variant'] == var], key=lambda r: -r['dps'])
            if not vr:
                continue
            for i, r in enumerate(vr):
                r['vsBest'] = round(100 * (r['dps'] / vr[0]['dps'] - 1), 2)
                r['rank'] = i + 1
                r['of'] = len(vr)
                fr = [q for q in vr if q['faction'] == r['faction']]
                r['vsFactionBest'] = round(100 * (r['dps'] / fr[0]['dps'] - 1), 2)
                r['factionRank'] = fr.index(r) + 1
        variants += rows

    # summary: per scenario, scope (all races, or within the race's faction), role and race
    summary = []
    for var in VARIANTS:
        for scope in ('All', 'Faction'):
            for rl in ('All', 'Melee', 'Ranged', 'Caster'):
                for race in RACE_ORDER:
                    rs = [r for r in variants if r['variant'] == var and r['race'] == race and (rl == 'All' or role.get(r['build']) == rl)]
                    if not rs:
                        continue
                    gk, rk = ('vsBest', 'rank') if scope == 'All' else ('vsFactionBest', 'factionRank')
                    best = [r for r in rs if r[rk] == 1]
                    near = sorted(rs, key=lambda r: -r[gk])
                    summary.append({'variant': var, 'scope': scope, 'role': rl, 'race': race, 'faction': 'Horde' if race in HORDE else 'Alliance',
                                    'presets': len(rs), 'timesBest': len(best), 'avgGap': round(-statistics.mean(r[gk] for r in rs), 2),
                                    'worstGap': round(-min(r[gk] for r in rs), 2), 'medianRank': statistics.median(r[rk] for r in rs),
                                    'bestFor': [bmap[r['build']]['label'] for r in near[:3]], 'bestForGap': [round(-r[gk], 2) for r in near[:3]],
                                    'raceIcon': RACE_ICON.get(race)})

    for key, title, text in SCENARIOS:
        notes.append({'kind': 'scenario', 'key': key, 'title': title, 'text': text})
    for p, (title, text) in PART_INFO.items():
        notes.append({'kind': 'part', 'key': p, 'title': title, 'text': text, 'short': PART_SHORT.get(p, text), 'icon': PART_ICON.get(p)})
    for cls in ['Warrior', 'Rogue', 'Hunter', 'Paladin', 'Shaman', 'Druid', 'Mage', 'Warlock', 'Priest']:
        notes.append({'kind': 'class', 'key': cls, 'title': cls, 'text': finding(cls, bmap, variants, breakdown)})
    dump('race_variants', variants)
    dump('race_breakdown', breakdown)
    dump('race_summary', summary)
    dump('race_notes', notes)
    return variants, breakdown, summary, notes


def pct(x):
    return f'{abs(x):.1f}%'


def finding(cls, bmap, variants, breakdown):
    """A few plain sentences on a class's races, from the numbers."""
    bs = sorted([b for b in bmap.values() if b['class'] == cls], key=lambda b: -b['dps'])
    V = defaultdict(list)
    for r in variants:
        if r['class'] == cls:
            V[(r['build'], r['variant'])].append(r)
    for k in V:
        V[k].sort(key=lambda r: -r['dps'])
    out = []
    # who is best with the presets' own weapons
    bests = [(b, V[(b['id'], 'baseline')]) for b in bs]
    if not bests:
        return ''
    names = {rs[0]['race'] for _, rs in bests}
    if len(names) == 1:
        race = names.pop()
        gaps = [rs[1]['vsBest'] for _, rs in bests if len(rs) > 1]
        second = bests[0][1][1] if len(bests[0][1]) > 1 else None
        s = f"{race} is the best race on {'every' if len(bs) > 1 else 'the'} {cls} preset"
        if second:
            s += f"; {second['race']} is next, {pct(second['vsBest'])} behind on {bests[0][0]['build']}"
        out.append(s + '.')
    else:
        by = {}
        for b, rs in bests:
            by.setdefault(rs[0]['race'], []).append(b['build'])
        def andjoin(xs):
            return xs[0] if len(xs) == 1 else ', '.join(xs[:-1]) + ' and ' + xs[-1]
        out.append('Best race: ' + '; '.join(f"{r} on {andjoin(bl)}" for r, bl in by.items()) + '.')
    # closest other-faction pick on the top preset
    top = bests[0][1]
    fac0 = top[0]['faction']
    other = next((r for r in top if r['faction'] != fac0), None)
    if other and not (len(top) > 1 and top[1] is other):
        out.append(f"Best {other['faction']} pick on {bests[0][0]['build']}: {other['race']}, {pct(other['vsBest'])} behind.")
    # the racial worth most on the top preset, and any racial that costs DPS
    b0 = bests[0][0]
    mine = [x for x in breakdown if x['scenario'] == 'all' and x['part'] != 'base' and x['active'] and bmap.get(x['build'], {}).get('class') == cls]
    top_part = max((x for x in mine if x['build'] == b0['id']), key=lambda x: x['dps'], default=None)
    if top_part and top_part['dps'] > 2 * top_part['se']:
        dps = next(r['dps'] for r in bests[0][1] if r['race'] == top_part['race'])
        out.append(f"The biggest racial on {b0['build']} is {top_part['race']} {top_part['label']}, +{top_part['dps']:.1f} DPS ({100 * top_part['dps'] / dps:.1f}%).")
    worst = min(mine, key=lambda x: x['dps'] / max(x['se'], 0.01), default=None)
    if worst and worst['dps'] < -2 * worst['se'] and worst['dps'] < -1:
        why = ', the time spent casting it' if worst['part'] in ('chastise', 'starshards', 'darksacrifice', 'eureka', 'elune', 'bloodfury', 'berserking') else ''
        out.append(f"{worst['race']} {worst['label']} costs {abs(worst['dps']):.1f} DPS on {bmap[worst['build']]['build']} in the sim{why}.")
    # racial weapons
    wl = []
    for b, rs in bests:
        best_dps = rs[0]['dps']
        for w in V[(b['id'], 'weapon')]:
            if w['simmed']:
                wl.append((w['dps'] / best_dps - 1, b, w))
    if wl:
        g, b, w = max(wl, key=lambda t: t[0])
        wp = WEAPON_PART[w['race']]
        if abs(g) < 0.0005:
            out.append(f"With {WEAPON_WORD[wp]}, {w['race']} ties for the lead on {b['build']}.")
        elif g > 0:
            out.append(f"With {WEAPON_WORD[wp]}, {w['race']} would lead {b['build']} by {pct(100 * g)}.")
        else:
            out.append(f"With {WEAPON_WORD[wp]}, {w['race']} comes within {pct(100 * g)} on {b['build']}.")
    # target types that change the leader
    for t, word in (('beast', 'Beasts'), ('elemental', 'Elementals')):
        b0, rs0 = bests[0]
        tr = V[(b0['id'], t)]
        if tr and tr[0]['race'] != rs0[0]['race']:
            sky = {'High Order', 'Windshaper'}
            lead = tr[0]['race']
            nxt = next((r for r in tr[1:] if not (r['race'] in sky and lead in sky)), None)
            if lead in sky and len(tr) > 1 and tr[1]['race'] in sky:
                lead = 'High Order and Windshaper lead'
            else:
                lead += ' leads'
            base = next(r for r in rs0 if r['race'] == tr[0]['race'])
            s = f"Against {word} {lead} {b0['build']} ({'+' if tr[0]['dps'] >= base['dps'] else '-'}{pct(100 * (tr[0]['dps'] / base['dps'] - 1))} over its generic-boss DPS)"
            if nxt and abs(nxt['vsBest']) < 0.05:
                s += f", level with {nxt['race']}"
            elif nxt:
                s += f", {pct(nxt['vsBest'])} ahead of {nxt['race']}"
            out.append(s + '.')
    return ' '.join(out)


if __name__ == '__main__':
    import sys
    sys.path.insert(0, D)
    builds = json.load(open(os.path.join(D, 'data', 'builds.json')))
    races = json.load(open(os.path.join(D, 'data', 'races.json')))
    stats = json.load(open(os.path.join(D, 'data', 'stats.json')))

    def dump(name, rows):
        p = os.path.join(D, 'data', name + '.json')
        json.dump(rows, open(p, 'w'), separators=(',', ':'))
        print(f'{name}: {len(rows)} rows, {os.path.getsize(p) / 1024:.0f} KB')
    build(builds, races, stats, dump)
