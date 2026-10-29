#!/usr/bin/env python3
"""M0 of the raid sim (docs/raid_sim.md): a rough, additive estimate of what each buff role is worth to a
raid, from data the DPS Report already has. No sims.

Reads each build's one-buff-at-a-time runs (abl.json in tools/dps_report/raw/<id>.tgz), the published DPS and
the duty variants' DPS (req/*.meta.json). Writes docs/raid_sim_m0.md.

It's rough on purpose: the runs are 300 iterations at 180 s under the report's full set, with no camps, and they
skip buffs a request never had. M1's own role sims replace it.

    python3 tools/compsearch/m0.py
"""

import glob
import json
import os
import tarfile

ROOT = os.path.dirname(os.path.dirname(os.path.dirname(os.path.abspath(__file__))))
REQ = os.path.join(ROOT, 'tools', 'dps_report', 'req')
RAW = os.path.join(ROOT, 'tools', 'dps_report', 'raw')
OUT = os.path.join(ROOT, 'docs', 'raid_sim_m0.md')

# Role -> the ablation entries that remove it. A pair entry removes both slot holders at once.
ROLES = {
    'Sunder Armor': [('Debuffs', 'Sunder Armor')],
    'Minor armor slot': [('pair', 'faerieFire+curseOfRecklessness')],
    'Battle Shout': [('Raid buffs', 'Battle Shout (improved)')],
    'Curse of the Elements': [('Debuffs', 'Curse Of Elements')],
    'Judgement of Wisdom': [('Debuffs', 'Judgement Of Wisdom')],
    'Judgement of the Crusader': [('Debuffs', 'Judgement Of The Crusader (improved)')],
    'Kings': [('Self buffs', 'Blessing Of Kings')],
    'Might': [('Self buffs', 'Blessing Of Might (improved)')],
    'Wisdom': [('Self buffs', 'Blessing Of Wisdom (improved)')],
    'Gift of the Wild': [('Raid buffs', 'Gift Of The Wild (improved)')],
    'Arcane Brilliance': [('Raid buffs', 'Arcane Brilliance')],
    'Prayer of Spirit': [('Raid buffs', 'Divine Spirit')],
    'Strength of Earth': [('Raid buffs', 'Strength Of Earth Totem')],
    'Grace of Air': [('Raid buffs', 'Grace Of Air Totem')],
    'Weapon totem': [('Raid buffs', 'Totem Weapon Buff (TotemWeaponBuffWindfury)'),
                     ('Raid buffs', 'Totem Weapon Buff (TotemWeaponBuffFlametongue)')],
    'Trueshot Aura': [('Raid buffs', 'Trueshot Aura')],
    'Leader of the Pack': [('Raid buffs', 'Leader Of The Pack')],
    'Moonkin Aura': [('Raid buffs', 'Moonkin Aura')],
    'Mana Spring (party)': [('Raid buffs', 'Mana Spring Totem')],
}

# Who gives a role without a duty (their base rotation or a held aura), and which duty variants hold it.
FREE_GIVERS = {
    'Moonkin Aura': ['druid__balance', 'druid__balance-wrath'],
    'Gift of the Wild': ['druid__balance', 'druid__balance-wrath', 'druid__feral-cat'],
    'Leader of the Pack': ['druid__feral-cat'],
    'Trueshot Aura': ['hunter__marksmanship'],
    'Arcane Brilliance': ['mage__arcane', 'mage__fire', 'mage__frost', 'mage__frostfire'],
    'Prayer of Spirit': ['priest__shadow', 'priest__shadow-holy', 'priest__smite'],
    'Strength of Earth': ['shaman__enhancement-level60', 'shaman__enhancement-imp-gw',
                          'shaman__enhancement-water-shield', 'shaman__enhancement-mana-tide'],
    'Weapon totem': ['shaman__enhancement-level60', 'shaman__enhancement-imp-gw',
                     'shaman__enhancement-water-shield', 'shaman__enhancement-mana-tide'],
    'Kings': ['paladin__level_60', 'paladin__level_60_shock', 'paladin__level_60_shock_ret',
              'paladin__level_60_sub_holy'],
    'Might': ['paladin__level_60', 'paladin__level_60_shock', 'paladin__level_60_shock_ret',
              'paladin__level_60_sub_holy'],
    'Wisdom': ['paladin__level_60', 'paladin__level_60_shock', 'paladin__level_60_shock_ret',
               'paladin__level_60_sub_holy'],
}
NO_PROVIDER_IN_M1 = {'Grace of Air', 'Judgement of Wisdom', 'Sunder Armor'}  # placeholders only (or M1b builds)

SLOTS = {10: 7, 20: 14, 40: 29}  # DPS slots at the default tank and healer counts


def load():
    builds, duties = {}, {}
    for f in sorted(glob.glob(os.path.join(REQ, '*.meta.json'))):
        m = json.load(open(f))
        dps = m['published']['dps']
        if 'duty' in m:
            duties[m['id']] = (m['baseBuild'], m['duty'], dps)
        elif 'baseBuild' not in m:
            builds[m['id']] = {'dps': dps, 'class': m['class'], 'spec': m['spec'], 'cost': {}, 'measured': set()}
    for bid, b in builds.items():
        with tarfile.open(os.path.join(RAW, bid + '.tgz')) as t:
            abl = json.load(t.extractfile('abl.json'))['entry']
        drops = {(d['category'], d['item']): -d['delta'] for d in abl['drop'] if 'delta' in d}
        drops.update({('pair', p['pair']): -p['delta'] for p in abl['pairs'] if 'delta' in p})
        for role, keys in ROLES.items():
            hit = [drops[k] for k in keys if k in drops]
            if hit:
                b['cost'][role] = max(hit)
                b['measured'].add(role)
            else:
                b['cost'][role] = 0.0
    return builds, duties


def duty_givers(duties):
    g = {}
    names = {'Battle Shout': 'Battle Shout', 'Faerie Fire': 'Minor armor slot',
             'Curse of the Elements': 'Curse of the Elements',
             'Judgement of the Crusader': 'Judgement of the Crusader'}
    for did, (base, duty, dps) in duties.items():
        g.setdefault(names[duty], []).append((did, base, dps))
    return g


def dominated(builds):
    """The survivors of the dominance rule (docs/raid_sim.md, The search, step 1), at a 1.5 DPS margin."""
    ids = list(builds)
    gives = {r: set(v) for r, v in FREE_GIVERS.items()}
    out = []
    for b in ids:
        bb = builds[b]
        for a in ids:
            if a == b:
                continue
            aa = builds[a]
            # a must be able to fill wherever b can: a gives nothing b doesn't.
            if any(a in v and b not in v for v in gives.values()):
                continue
            need = sum(max(0.0, aa['cost'][r] - bb['cost'][r]) for r in ROLES)
            if aa['dps'] - bb['dps'] >= need + 1.5:
                out.append((b, a))
                break
    dead = {b for b, _ in out}
    return [b for b in ids if b not in dead], out


def dps_under(b, missing):
    return b['dps'] - sum(b['cost'][r] for r in missing)


def best_fill(builds, missing, n):
    """The best build under a set (repeats allowed), among builds whose gifts are all up."""
    ok = [bid for bid in builds if not any(bid in FREE_GIVERS.get(r, ()) for r in missing)]
    best = max(ok, key=lambda bid: dps_under(builds[bid], missing))
    return best, n * dps_under(builds[best], missing)


def role_table(builds, duties, n):
    """Each role's gain at the all-buffs set, its cheapest provider's cost and the net, with a raid of n DPS."""
    dg = duty_givers(duties)
    top, _ = best_fill(builds, set(), n)
    rows = []
    for role in ROLES:
        gain_one = builds[top]['cost'][role]
        gain = gain_one * (n - 1)
        prov = []
        for g in FREE_GIVERS.get(role, []):
            prov.append((builds[top]['dps'] - builds[g]['dps'], g))
        for did, base, dps in dg.get(role, []):
            prov.append((builds[top]['dps'] - dps, did))
        cost, who = min(prov) if prov else (None, 'placeholders only' if role in NO_PROVIDER_IN_M1 else '-')
        rows.append((role, gain_one, gain, cost, who, None if cost is None else gain - cost))
    return top, rows


def main():
    builds, duties = load()
    surv, dom = dominated(builds)
    lines = ['# Raid sim M0: what each role is worth (rough)', '',
             'Written by `tools/compsearch/m0.py` from the DPS Report\'s one-buff-at-a-time runs (300 iterations, 180 s,',
             'the report\'s full set, no camps). Additive, no sims: M1\'s own role sims replace it. A role a build\'s',
             'request never had shows 0 for it (marked *).', '']
    lines += ['## The fillers, by published DPS', '', '| Build | DPS | Dominated by |', '|---|---|---|']
    domby = dict(dom)
    for bid in sorted(builds, key=lambda x: -builds[x]['dps']):
        lines.append(f"| {bid} | {builds[bid]['dps']:.1f} | {domby.get(bid, '**survives**')} |")
    lines += ['', f'**{len(surv)} of {len(builds)} survive** the dominance rule at a 1.5 DPS margin: '
              + ', '.join(sorted(surv)) + '.', '']
    for n_size, n in SLOTS.items():
        top, rows = role_table(builds, duties, n)
        lines += [f'## {n_size} players ({n} DPS slots)', '',
                  f'The fill under the full set is {n} × {top}. Gain is what the role is worth to the other '
                  f'{n - 1}; cost is its cheapest provider\'s DPS below the filler (duty included).', '',
                  '| Role | Per member | Gain | Provider | Cost | Net |', '|---|---|---|---|---|---|']
        for role, one, gain, cost, who, net in rows:
            star = '' if role in builds[top]['measured'] else '*'
            lines.append(f'| {role} | {one:.1f}{star} | {gain:.0f} | {who} | '
                         + ('-' if cost is None else f'{cost:.1f}') + ' | '
                         + ('-' if net is None else f'{net:.0f}') + ' |')
        lines.append('')
    lines += ['## Per build, per role (DPS lost when the role is missing)', '',
              '| Build | ' + ' | '.join(ROLES) + ' |', '|---|' + '---|' * len(ROLES)]
    for bid in sorted(builds):
        b = builds[bid]
        cells = [f"{b['cost'][r]:.1f}" + ('' if r in b['measured'] else '*') for r in ROLES]
        lines.append(f'| {bid} | ' + ' | '.join(cells) + ' |')
    open(OUT, 'w').write('\n'.join(lines) + '\n')
    print('wrote', OUT, f'({len(surv)} survivors)')


if __name__ == '__main__':
    main()
