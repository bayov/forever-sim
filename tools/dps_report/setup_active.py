#!/usr/bin/env python3
"""setup_active.py: decide, for every setup entry of every preset, whether it does anything in the sim.

Ground truth is abl/ablate.json (ablate.py): each preset simmed with one entry left out, same seed.
A bit-identical DPS means the entry has no effect. On top of that the sim's own rules decide which
entry of a non-stacking pair is the redundant one, and give the plain-language reason.

annotate(setup, builds) adds `active` (bool), `why` (reason, when inactive) and `note` (for active
entries whose 'improved' flag the sim ignores under Forever) to the setup rows in place.
"""
import json, os, re

from dashpaths import D


def words(k):
    return re.sub(r'(?<=[a-z0-9])(?=[A-Z])', ' ', k[0].upper() + k[1:])


def key_of(row):
    """The request key of a setup row (camelCase), from its item text."""
    if row['category'] == 'Professions':
        return 'prof:' + row['item']
    head = row['item'].split(':')[0] if row['category'] == 'Consumes' else re.sub(r' \(.*\)$', '', row['item'])
    w = head.split(' ')
    return w[0].lower() + ''.join(w[1:])


# what an entry gives, for the "this build doesn't use it" sentence
GIVES = {
    'arcaneBrilliance': 'Intellect', 'divineSpirit': 'Spirit', 'battleShout': 'attack power', 'blessingOfMight': 'attack power',
    'blessingOfKings': '10% to every attribute', 'blessingOfWisdom': 'mana per 5s', 'manaSpringTotem': 'mana per 5s',
    'giftOfTheWild': 'armor, attributes and resistances', 'strengthOfEarthTotem': 'Strength', 'graceOfAirTotem': 'Agility',
    'powerWordFortitude': 'Stamina', 'bloodPact': 'Stamina', 'judgementOfWisdom': 'mana when you hit the boss',
    'judgementOfTheCrusader': 'Holy damage taken by the boss', 'curseOfElements': 'spell damage taken by the boss',
    'flask': None, 'food': None,
}
CONSUME_GIVES = {
    'Elixir Of Frost Power': 'Frost spell damage', 'Elixir Of Firepower': 'Fire spell damage', 'Elixir Of Shadow Power': 'Shadow spell damage',
    'Elixir Of Nature Power': 'Nature spell damage', 'Elixir Of Greater Firepower': 'Holy spell damage (WoW Forever made it the Elixir of Holy Power)',
    'Greater Arcane Elixir': 'spell damage', 'Greater Mageblood Elixir': 'mana per 5s', 'Flask Of Distilled Wisdom': 'mana',
    'Flask Of Supreme Power': 'spell damage', 'Flask Of The Titans': 'health', 'Juju Might': 'attack power',
    'Elixir Of Greater Intellect': 'Intellect', 'Cerebral Cortex Compound': 'Intellect', 'Elixir Of Sages': 'Spirit and 2% crit',
    'Elixir Of Greater Spirit': 'Spirit', 'Elixir Of The Owl': 'Intellect and 2% crit',
}
SCHOOL_NONE = {'Frost spell damage': 'Frost', 'Fire spell damage': 'Fire', 'Shadow spell damage': 'Shadow', 'Nature spell damage': 'Nature'}

# raid buffs a class puts up itself in the sim (each class's AddRaidBuffs)
OWN = {('Mage', 'arcaneBrilliance'): "The mage already brings its own Arcane Brilliance",
       ('Priest', 'divineSpirit'): "The priest already brings its own Divine Spirit",
       ('Priest', 'shadowProtection'): "The priest already brings its own Shadow Protection",
       ('Priest', 'powerWordFortitude'): "The priest already brings its own Power Word: Fortitude",
       ('Warlock', 'bloodPact'): "The warlock's Imp already brings Blood Pact",
       ('Druid', 'moonkinAura'): "The druid already brings its own Moonkin Aura",
       ('Druid', 'leaderOfThePack'): "The druid already brings its own Leader of the Pack",
       ('Hunter', 'trueshotAura'): "The hunter already brings its own Trueshot Aura"}

DEFENSIVE = {
    'thorns': "It only hurts attackers, and the boss attacks the tank, not you.",
    'retributionAura': "It only hurts attackers, and the boss attacks the tank, not you.",
    'devotionAura': "Armor only, and the boss attacks the tank, not you.",
    'shadowProtection': "Shadow resistance only. Nothing in this fight deals Shadow damage to you.",
    'fireResistanceAura': "Fire resistance only. Nothing in this fight deals Fire damage to you.",
    'frostResistanceAura': "Frost resistance only. Nothing in this fight deals Frost damage to you.",
    'powerWordFortitude': "Stamina only. Health doesn't change damage.",
    'bloodPact': "Stamina only. Health doesn't change damage.",
    'curseOfWeakness': "It weakens the boss's attacks, and the boss attacks the tank, not you.",
    'demoralizingShout': "It lowers the boss's attack power, and the boss attacks the tank, not you.",
    'demoralizingRoar': "It lowers the boss's attack power, and the boss attacks the tank, not you.",
    'thunderClap': "It slows the boss's attacks, and the boss attacks the tank, not you.",
    'insectSwarm': "It makes the boss miss more, and the boss attacks the tank, not you.",
    'scorpidSting': "Defensive only. The sim gives it no effect.",
    'judgementOfLight': "It heals whoever hits the boss. Healing doesn't add damage.",
}
GONE = {
    'curseOfShadow': "Not in WoW Forever: its Curse of the Elements already covers Shadow and Arcane.",
    'sanctityAura': "Not in WoW Forever: Sanctity Aura isn't in its Retribution tree.",
}
# under Forever the sim takes these 'improved' flags as regular (the talents that improved them are gone)
IMPROVED_IGNORED = {'giftOfTheWild', 'powerWordFortitude', 'bloodPact', 'blessingOfMight', 'blessingOfWisdom', 'battleShout', 'exposeArmor',
                    'judgementOfTheCrusader', 'retributionAura', 'thorns', 'strengthOfEarthTotem', 'graceOfAirTotem'}
NO_MANA = {'Warrior', 'Rogue'}
PAIRS = {  # (a, b): what they share, the one kept when both are redundant
    ('sunderArmor', 'exposeArmor'): ('the same 2250 armor', 'sunderArmor'),
    ('faerieFire', 'curseOfRecklessness'): ('the same 505 armor', 'faerieFire'),
    ('leaderOfThePack', 'moonkinAura'): ('the same 3% crit', None),
    ('demoralizingShout', 'demoralizingRoar'): ("the boss's attack power", 'demoralizingShout'),
}
NAME = {'sunderArmor': 'Sunder Armor', 'exposeArmor': 'Expose Armor', 'faerieFire': 'Faerie Fire', 'curseOfRecklessness': 'Curse of Recklessness',
        'leaderOfThePack': 'Leader of the Pack', 'moonkinAura': 'Moonkin Aura', 'demoralizingShout': 'Demoralizing Shout', 'demoralizingRoar': 'Demoralizing Roar'}
ARMOR = {'sunderArmor', 'exposeArmor', 'faerieFire', 'curseOfRecklessness'}


def mana_reason(g, cls):
    if cls in NO_MANA:
        return f"It gives {g}, and {cls}s have no mana."
    return f"It gives {g}, and this build never runs short of mana."


def reason(k, row, cls, role, abl_same, pair_same):
    """The plain sentence for an entry that changes nothing."""
    item = row['item']
    if k in GONE:
        return GONE[k]
    if (cls, k) in OWN:
        return OWN[(cls, k)] + '.'
    if k in ARMOR and pair_same:
        return "It lowers armor, and armor doesn't reduce this build's spell damage."
    if k in DEFENSIVE:
        return DEFENSIVE[k]
    if k == 'trueshotAura':
        return "In WoW Forever it gives only ranged attack power, and this build doesn't attack at range."
    if row['category'] == 'Professions':
        return "No effect on this build's damage."
    if row['category'] == 'Consumes':
        name = item.split(': ', 1)[1]
        g = CONSUME_GIVES.get(name)
        if g in SCHOOL_NONE:
            return f"It adds {g}, and this build deals no {SCHOOL_NONE[g]} damage."
        if k in ('defaultPotion', 'defaultConjured', 'sapperExplosive', 'fillerExplosive'):
            return "The rotation never uses it."
        if k == 'offHandImbue':
            return "There is no off-hand weapon to put it on."
        if g and ('mana' in g or g in ('Intellect', 'Spirit')):
            return mana_reason(g, cls)
        if g == 'health':
            return "Health only. Health doesn't change damage."
        if g:
            return f"It gives {g}, which this build doesn't use."
        return "No effect on this build's damage."
    g = GIVES.get(k)
    if g and ('mana' in g or g in ('Intellect', 'Spirit')):
        return mana_reason(g, cls)
    if k in ('battleShout', 'blessingOfMight') and role == 'Ranged':
        return "It gives melee attack power only, and this build attacks at range."
    if k in ('battleShout', 'blessingOfMight') and role == 'Caster':
        return "It gives attack power, which this build's spells don't use."
    if k == 'judgementOfTheCrusader':
        return "It raises Holy damage taken, but this build deals no Holy damage."
    if g:
        return f"It gives {g}, which this build doesn't use."
    return "No effect on this build's damage."


def roles(stat_rows):
    """Ranged, Caster or Melee, as the page's roles calc reads the character sheet."""
    by = {}
    for s in stat_rows:
        by.setdefault(s['build'], {})[s['stat']] = float(s['value'] or 0)
    return {b: ('Ranged' if s.get('RangedAttackPower', 0) > s.get('AttackPower', 0) else 'Caster' if s.get('SpellPower', 0) > s.get('AttackPower', 0) else 'Melee')
            for b, s in by.items()}


def annotate(setup, builds, stat_rows, path=os.path.join(D, 'abl', 'ablate.json')):
    abl = json.load(open(path))['builds']
    binfo = {b['id']: dict(b) for b in builds}
    for bid, role in roles(stat_rows).items():
        binfo.setdefault(bid, {})['role'] = role
    stats = {}
    for r in setup:
        bid = r['build']
        b = binfo.get(bid, {})
        cls, role = b.get('class'), b.get('role')
        a = abl.get(bid)
        k = key_of(r)
        r['active'], r['why'], r['note'] = True, None, None
        if not a:
            continue
        d = next((x for x in a['drop'] if x['category'] == r['category'] and x['item'] == r['item']), None)
        same = bool(d and d.get('same'))
        # the pair this entry belongs to, if both are set
        pair = next(((p, v) for p, v in PAIRS.items() if k in p), None)
        pair_same = None
        partner_row = None
        if pair:
            other = pair[0][1] if pair[0][0] == k else pair[0][0]
            partner_row = next((x for x in a['drop'] if key_of({'category': x['category'], 'item': x['item']}) == other), None)
            pr = next((x for x in a['pairs'] if x['pair'] in (k + '+' + other, other + '+' + k)), None)
            pair_same = bool(pr and pr.get('same'))
        if k in GONE:
            r['active'], r['why'] = False, GONE[k]
        elif pair and partner_row is not None and not pair_same and (same or partner_row.get('same')):
            # a non-stacking pair that matters: the redundant one is the one whose removal changes nothing;
            # when neither changes anything alone, keep the sim's first choice
            other = pair[0][1] if pair[0][0] == k else pair[0][0]
            other_same = bool(partner_row.get('same'))
            keep = pair[1][1]
            if keep is None:  # the 3% crit pair: casters keep Moonkin Aura, everyone else Leader of the Pack
                keep = 'moonkinAura' if role == 'Caster' else 'leaderOfThePack'
            redundant = same and (not other_same or k != keep)
            if redundant:
                what = pair[1][0]
                tail = " In WoW Forever the two share one slot." if k in ('faerieFire', 'curseOfRecklessness', 'leaderOfThePack', 'moonkinAura') else " The two don't stack."
                if k == 'exposeArmor':  # the sim's aura log shows Expose Armor up from 3s, as it takes the slot over
                    tail += " Its uptime still shows because the sim swaps it in after 3s."
                r['active'], r['why'] = False, f"{NAME[other]} already gives {what}.{tail}"
        elif same:
            r['active'], r['why'] = False, reason(k, r, cls, role, same, pair_same)
        if r['active'] and '(improved)' in r['item'] and k in IMPROVED_IGNORED:
            r['note'] = "WoW Forever has no Improved version of it, so the sim uses the regular one."
        stats.setdefault(bid, [0, 0])
        stats[bid][0] += 1
        stats[bid][1] += 0 if r['active'] else 1
    return stats


if __name__ == '__main__':
    setup = json.load(open(os.path.join(D, 'data', 'setup.json')))
    builds = json.load(open(os.path.join(D, 'data', 'builds.json')))
    st = annotate(setup, builds, json.load(open(os.path.join(D, 'data', 'stats.json'))))
    for bid, (n, off) in st.items():
        print(f'{bid:40s} {off:2d} of {n} inactive')
    for r in setup:
        if not r['active']:
            print(' ', r['build'], '|', r['item'], '|', r['why'])
