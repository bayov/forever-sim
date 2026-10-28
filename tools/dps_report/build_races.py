#!/usr/bin/env python3
"""build_races.py: the Races tab's datasets, from the race sweep (out/*.json) and the race
scenario sims (rv/results.jsonl, run_races.py). build_data.py calls build(); it can also run alone.

race_variants   build × race × scenario: DPS, error, gap to the preset's best race in that scenario, and `tuned`: the
                note of the race's own rotation or gear where it has one (req/races/<id>.<Race>.json, race_override.py)
race_breakdown  build × race × part: base stats, each racial, and the target change
race_summary    scenario × scope × role × race: how often a race is best and how far behind it is
race_notes      what each scenario switches on, what each racial part is, a finding per class
"""
import json, math, os, re, statistics
from collections import defaultdict

from dashpaths import D, TREE
from race_override import tuned_kind  # noqa: E402
RACE_NAMES = {'NightElf': 'Night Elf', 'SkyborneHighOrder': 'High Order', 'SkyborneWindshaper': 'Windshaper'}
HORDE = {'Orc', 'Undead', 'Tauren', 'Troll', 'Windshaper'}
RACE_ICON = {'Human': 'race_human_male', 'Dwarf': 'race_dwarf_male', 'Night Elf': 'race_nightelf_male', 'Gnome': 'race_gnome_male',
             'High Order': 'inv_feather_02', 'Orc': 'race_orc_male', 'Undead': 'race_scourge_male', 'Tauren': 'race_tauren_male',
             'Troll': 'race_troll_male', 'Windshaper': 'ability_druid_flightform'}
VARIANTS = ['baseline', 'beast', 'elemental', 'humanoid', 'off']
TARGET_VARIANTS = ('beast', 'elemental', 'humanoid')
RACE_ORDER = ['Human', 'Dwarf', 'Night Elf', 'Gnome', 'High Order', 'Orc', 'Undead', 'Tauren', 'Troll', 'Windshaper']

# Racial parts the sim models, by race: (part key, racial name, what it does). The text follows client 1.60.1.70425,
# and the development notes of 2026-10-26 for the racial pass (Human, Dwarf, Orc, Tauren, Skyborne).
PART_INFO = {
    'spirit': ('The Human Spirit', 'Spirit increased by 10%.'),
    'ruthless': ('Ruthless', '+1% critical strike chance with all spells and attacks, with any weapon.'),
    'precision': ("Artisan's Precision", '+1% hit chance with spells and attacks.'),
    'command': ('Command', 'Damage dealt by Hunter and Warlock pets increased by 5%.'),
    'elune': ("Elune's Light", '+10% critical strike chance for 15s, 3 min cooldown.'),
    'shadowmeld': ('Shadowmeld', 'Slip into the shadows, 2 min cooldown in combat. The next stealth ability (Ravage, Ambush, Garrote) can follow it.'),
    'leyline': ('Read Leyline', 'A 2s cast that raises health and mana regeneration by 50% for 15s away from a Ley Line, 2 min cooldown.'),
    'mind': ('Expansive Mind', 'Maximum Mana, Rage or Energy increased by 5%.'),
    'eureka': ('Eureka!', 'For 8s, direct damage abilities cost 10% less and deal 10% more damage (DoTs too for Warlocks, heals too for Priests), 2 min cooldown.'),
    'bloodfury': ('Blood Fury', 'Attack power and spell power increased by 10% for 15s, 2 min cooldown.'),
    'touch': ('Touch of the Grave', 'Spells and attacks have a 20% chance (30% for Mages, Priests and Warlocks) to deal Shadow damage equal to 5% of your maximum health and heal you for it. At most once every 15s, and it cannot crit.'),
    'endurance': ('Endurance', 'Health increased by 5%.'),
    'berserking': ('Berserking', 'Attack and casting speed increased by 10% for 15s, 3 min cooldown.'),
    'windblessed': ('Wind Blessed', 'Spellcasting, melee and ranged haste increased by 2%.'),
    'beast_dwarf': ('Big Game Hunter', 'Damage dealt versus Beasts increased by 5%.'),
    'beast_troll': ('Beast Slaying', 'Damage dealt versus Beasts increased by 5%.'),
    'insight': ('Elemental Insight', 'Damage dealt versus Elementals increased by 5%.'),
    'chastise': ('Chastise', 'Dwarf Priests: an instant Holy hit on a Humanoid, 2 min cooldown.'),
    'starshards': ('Starshards', 'Night Elf Priests: Arcane damage over 6s, 30s cooldown.'),
    'darksacrifice': ('Dark Sacrifice', 'Undead Priests: trade health for mana over 15s, 10 min cooldown.'),
    'base': ('Base stats', "The race's own stats, with no racials."),
    'target': ('Target type', 'Racials, talents and spells that work only on this creature type.'),
}

# The tooltip's one line for the parts whose text above runs long; the others show their text as it is.
PART_SHORT = {
    'touch': '20% chance a hit (30% for Mages, Priests and Warlocks) to deal Shadow damage equal to 5% of max health. At most once every 15s.',
}

# Each racial's game icon, from ui/core/forever/racials.json. The priest racials are from wowhead, and Dark Sacrifice
# has no known icon.
PART_ICON = {
    'spirit': 'inv_enchant_shardbrilliantsmall', 'ruthless': 'ability_meleedamage', 'precision': 'ability_marksmanship', 'beast_dwarf': 'inv_weapon_rifle_05',
    'elune': 'spell_nature_moonglow', 'shadowmeld': 'ability_ambush', 'leyline': 'ability_mage_incantersabsorbtion', 'mind': 'inv_enchant_essenceeternallarge', 'eureka': 'inv_gnometoy', 'command': 'ability_warrior_warcry',
    'bloodfury': 'racial_orc_berserkerstrength', 'touch': 'spell_shadow_fingerofdeath', 'endurance': 'spell_nature_unyeildingstamina',
    'berserking': 'racial_troll_berserk', 'beast_troll': 'inv_misc_pelt_bear_ruin_02', 'windblessed': 'inv_misc_volatileair',
    'insight': 'achievement_raidprimalist_windelemental', 'starshards': 'spell_arcane_starfire', 'chastise': 'spell_holy_chastise',
}

RACE_PARTS = {'Human': ['spirit', 'ruthless'], 'Dwarf': ['precision', 'beast_dwarf', 'chastise'], 'Night Elf': ['elune', 'shadowmeld', 'starshards'],
              'Gnome': ['mind', 'eureka'], 'Orc': ['command', 'bloodfury'], 'Undead': ['touch', 'darksacrifice'], 'Tauren': ['endurance'],
              'Troll': ['berserking', 'beast_troll'], 'High Order': ['windblessed', 'insight', 'leyline'], 'Windshaper': ['windblessed', 'insight']}
SCENARIOS = [
    ('baseline', 'Generic boss', 'Each preset against a level 63 boss with no creature type. In client 70425 no racial depends on the weapon. '
     "A race plays its own rotation or gear where that beats the preset's by 2 standard errors, in every scenario; its row says what changed."),
    ('beast', 'Beast target', 'A Beast boss turns on Troll Beast Slaying and Dwarf Big Game Hunter (5% more damage).'),
    ('elemental', 'Elemental target', 'An Elemental boss turns on High Order and Windshaper Elemental Insight (5% more damage).'),
    ('humanoid', 'Humanoid target', 'Priests only. Dwarf Priests cast Chastise, which works only on Humanoids. No other racial cares about Humanoids.'),
    ('off', 'Racials off', 'Every race on its base stats, with all racials left out.'),
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
    tuned = {}  # (build, race display) -> the note of the race's own rotation or gear (req/races/, race_override.py)
    for o in sorted(os.listdir(os.path.join(D, 'out'))):
        if not o.endswith('.json'):
            continue
        out = json.load(open(os.path.join(D, 'out', o)))
        for r in out['races']:
            res[(out['id'], rn(r['race']))]['baseline'] = (r['dps'], r['se'])
            if r.get('override'):
                tuned[(out['id'], rn(r['race']))] = (r['override'].get('note', ''), tuned_kind(r['override']))

    variants, breakdown, notes = [], [], []
    for bid, b in list(bmap.items()):
        req = json.load(open(os.path.join(D, 'req', bid + '.json')))
        cls = b['class']
        if b.get('encounter') or b.get('duty'):
            del bmap[bid]  # encounter and duty presets have only the race sweep; the scenarios are on the generic boss's base presets
            continue
        if os.environ.get('RACES_PARTIAL') and any(v not in res[(bid, rn(rc))] for rc in RR.CLASS_RACES[cls] for v in ('off', 'beast', 'elemental')):
            del bmap[bid]  # a test build on part of the sims
            continue
        rows = []
        for rc in RR.CLASS_RACES[cls]:
            race = rn(rc)
            v = res[(bid, race)]
            base = v['baseline']
            for var in VARIANTS:
                note = ''
                if var == 'baseline':
                    val, sim = base, True
                else:
                    val, sim = v.get(var, (None, None)), var in v
                if var == 'humanoid' and not sim:
                    continue  # only the classes with a Humanoid racial (Priests) have this scenario
                if val[0] is None:
                    raise SystemExit(f'missing {bid} {race} {var}')
                rows.append({'build': bid, 'class': cls, 'race': race, 'faction': 'Horde' if race in HORDE else 'Alliance', 'variant': var,
                             'dps': round(val[0], 1), 'se': round(val[1], 2), 'isPreset': race == b['race'], 'simmed': sim, 'note': note,
                             'tuned': tuned.get((bid, race), ('', ''))[0], 'tunedKind': tuned.get((bid, race), ('', ''))[1], 'raceIcon': RACE_ICON.get(race)})
            # breakdown: base stats, then each racial part (Shapley split when two parts count, so the parts add up exactly)
            off = v['off']
            ps = [p for p in RR.parts_for(cls, rc, req)]
            if len(ps) > 2:  # a part whose own sim matches the baseline exactly (same seed) does nothing on this preset
                ps = [p for p in ps if v['off:' + p][0] != base[0]]
            contrib = {}
            if len(ps) == 1:
                contrib[ps[0]] = (base[0] - off[0], math.hypot(base[1], off[1]))
            elif len(ps) == 2:
                a, c = ps
                oa, oc = v['off:' + a], v['off:' + c]
                contrib[a] = (0.5 * ((base[0] - oa[0]) + (oc[0] - off[0])), math.hypot(base[1], off[1]))
                contrib[c] = (0.5 * ((base[0] - oc[0]) + (oa[0] - off[0])), math.hypot(base[1], off[1]))
            elif len(ps) > 2:
                # three or more live parts: each part's leave-one-out loss, scaled so the parts add up to all of them
                loo = {p: base[0] - v['off:' + p][0] for p in ps}
                scale = (base[0] - off[0]) / sum(loo.values()) if sum(loo.values()) else 0
                for p in ps:
                    contrib[p] = (loo[p] * scale, math.hypot(base[1], off[1]))
            breakdown.append({'build': bid, 'race': race, 'scenario': 'all', 'part': 'base', 'label': PART_INFO['base'][0], 'dps': round(off[0], 2),
                              'se': round(off[1], 2), 'active': True, 'note': ''})
            for p in RACE_PARTS[race]:
                key = {'beast_dwarf': 'beast', 'beast_troll': 'beast'}.get(p, p)
                if p in ('chastise', 'starshards', 'darksacrifice') and cls != 'Priest':
                    continue
                if key in contrib:
                    d, s = contrib[key]
                    breakdown.append({'build': bid, 'race': race, 'scenario': 'all', 'part': p, 'label': PART_INFO[p][0], 'dps': round(d, 2), 'se': round(s, 2), 'active': True, 'note': ''})
                else:
                    if p.startswith('beast'):
                        nt = 'only against Beasts'
                    elif p == 'insight':
                        nt = 'only against Elementals'
                    else:
                        nt = 'no effect on this preset'
                    breakdown.append({'build': bid, 'race': race, 'scenario': 'all', 'part': p, 'label': PART_INFO[p][0], 'dps': 0, 'se': 0, 'active': False, 'note': nt})
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
    for race, offs in race_offsets().items():
        notes.append({'kind': 'stats', 'key': race, 'stats': offs})
    for cls in ['Warrior', 'Rogue', 'Hunter', 'Paladin', 'Shaman', 'Druid', 'Mage', 'Warlock', 'Priest']:
        notes.append({'kind': 'class', 'key': cls, 'title': cls, 'text': finding(cls, bmap, variants, breakdown)})
    dump('race_variants', variants)
    dump('race_breakdown', breakdown)
    dump('race_summary', summary)
    dump('race_notes', notes)
    return variants, breakdown, summary, notes


# Each race's level 60 attribute offsets from the sim (RaceOffsets in sim/core/base_stats.go). Every race of a class
# starts from the same class row, so these are the whole stat gap between two races before racials.
STAT_ORDER = ['Strength', 'Agility', 'Stamina', 'Intellect', 'Spirit']


def race_offsets():
    src = open(os.path.join(TREE, 'sim/core/base_stats.go')).read()
    body = src[src.index('var RaceOffsets'):]
    body = body[:body.index('\n}\n')]
    out = {}
    for m in re.finditer(r'proto\.Race_Race(\w+):\s*\{([^}]*)\}', body):
        if m.group(1) == 'Unknown':
            continue
        vals = dict((k, float(v)) for k, v in re.findall(r'stats\.(\w+):\s*(-?[\d.]+)', m.group(2)))
        out[RACE_NAMES.get(m.group(1), m.group(1))] = {k: vals.get(k, 0) for k in STAT_ORDER}
    return out


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
        out.append(f"{race} is best on {'every' if len(bs) > 1 else 'the'} {cls} preset.")
        if second:
            out.append(f"{second['race']} is next, {pct(second['vsBest'])} behind on {bests[0][0]['build']}.")
    else:
        by = {}
        for b, rs in bests:
            by.setdefault(rs[0]['race'], []).append(b['build'])
        def andjoin(xs):
            return xs[0] if len(xs) == 1 else ', '.join(xs[:-1]) + ' and ' + xs[-1]
        out.extend(f"{r} is best on {andjoin(bl)}." for r, bl in by.items())
    # closest other-faction pick on the top preset
    top = bests[0][1]
    fac0 = top[0]['faction']
    other = next((r for r in top if r['faction'] != fac0), None)
    if other and not (len(top) > 1 and top[1] is other):
        out.append(f"The best {other['faction']} race on {bests[0][0]['build']} is {other['race']}, {pct(other['vsBest'])} behind.")
    # the racial worth most on the top preset, and any racial that costs DPS
    b0 = bests[0][0]
    mine = [x for x in breakdown if x['scenario'] == 'all' and x['part'] != 'base' and x['active'] and bmap.get(x['build'], {}).get('class') == cls]
    top_part = max((x for x in mine if x['build'] == b0['id']), key=lambda x: x['dps'], default=None)
    if top_part and top_part['dps'] > 2 * top_part['se']:
        dps = next(r['dps'] for r in bests[0][1] if r['race'] == top_part['race'])
        out.append(f"{top_part['race']} {top_part['label']} is the biggest racial on {b0['build']}, +{top_part['dps']:.1f} DPS ({100 * top_part['dps'] / dps:.1f}%).")
    worst = min(mine, key=lambda x: x['dps'] / max(x['se'], 0.01), default=None)
    if worst and worst['dps'] < -2 * worst['se'] and worst['dps'] < -1:
        why = ', for the time spent casting it' if worst['part'] in ('chastise', 'starshards', 'darksacrifice', 'eureka', 'elune', 'shadowmeld', 'leyline', 'bloodfury', 'berserking') else ''
        out.append(f"{worst['race']} {worst['label']} costs {abs(worst['dps']):.1f} DPS on {bmap[worst['build']]['build']}{why}.")
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
            s = f"Against {word}, {lead} {b0['build']} ({'+' if tr[0]['dps'] >= base['dps'] else '-'}{pct(100 * (tr[0]['dps'] / base['dps'] - 1))} vs the generic boss)"
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
