"""The Holy paladin's healbench requests and UI rotations (docs/healing_paladin.md).

    python3 -I orchestration/tools/healers/paladin_holy_requests.py            # writes the presets
    python3 -I orchestration/tools/healers/paladin_holy_requests.py --out DIR --role tank --set key=value ...

The presets: tools/healbench/req/paladin_holy.json (tank healer) and paladin_holy_raid.json (raid
healer), ui/holy_paladin/apls/tank_healer.apl.json and raid_healer.apl.json, and the gear set
ui/holy_paladin/gear_sets/level60_phase1_synthetic.gear.json. Each request is the
Restoration shaman's (tools/healbench/req/shaman_restoration*.json: the same raid buffs, debuffs,
consumables and encounter) with the paladin, its synthetic set and its rotation. --out writes one request
with the knobs below changed, for tuning on tools/healbench.
Run from the repo root.
"""
import argparse, copy, json, os

TANK = {'type': 'Tank', 'index': 0}
MOST = {'type': 'MostHurt'}
LOW = {'type': 'LowestHealth'}
SELF = {'type': 'Self'}

HL = {9: 25292, 8: 10329, 7: 10328, 6: 3472, 5: 1042}
FOL = {6: 19943, 5: 19942, 4: 19941, 3: 19940}
HOLY_SHOCK = 20930
DIVINE_FAVOR = 20216
OATH = 1328548
VIGIL, VIGIL_MARK = 1311595, 1311597
LAY_ON_HANDS = 10310

# The knobs, the presets' values (tuned on tools/healbench, docs/healing_paladin.md).
DEFAULTS = {
    'talents': '53320015002101251-5-052330002',
    'race': 'RaceHuman',
    'libram': 272436,      # Libram of Economy; 279248 Libram of Infusion
    'hl': '9,8,7',         # the Holy Light ranks bound, biggest first
    'fol': '6',            # the Flash of Light ranks bound
    'emergency': 0.35,     # Flash of Light (or Holy Shock) at the lowest health under this
    'shock_emergency': 0.5,  # Holy Shock at the lowest health under this (0 never)
    'shock_tank': 0,       # Holy Shock at the tank when it misses its heal (1) or never (0)
    'oath': 1,             # keep Oath of Light on the tank
    'favor': 1,            # Divine Favor before a Holy Light that what's missing covers
    'vigil': 0,            # Light's Vigil on ourselves and Holy Shock at it, for our party
    'vigil_missing': 1500,  # cast the Vigil once the most hurt misses this much
    'loh': 0.1,            # Lay on Hands at the lowest health under this (0 never)
    'fol_fill': 1,         # Flash of Light at the most hurt when it misses what it heals
    'hl_target': 'tank',   # where Holy Light goes: the tank, or the most hurt
    'shock_fill': 0,       # Holy Shock at the most hurt when it misses what Holy Shock heals
    'fit': 1.0,            # a rank goes out once what's missing is this share of what it heals
    'mana_hl': 0,          # Holy Light only above this mana share (Flash of Light below it)
    'aura': 'NoPaladinAura',  # our own aura (ConcentrationAura: the party resists 35% of the pushback)
}


def const(v):
    return {'const': {'val': str(v)}}


def cast(spell, target=None, cond=None, notes=None):
    c = {'spellId': {'spellId': spell}}
    if target:
        c['target'] = target
    a = {'action': {'castSpell': c}}
    if cond:
        a['action']['condition'] = cond
    if notes:
        a['notes'] = notes
    return a


def cmp(op, lhs, rhs):
    return {'cmp': {'op': op, 'lhs': lhs, 'rhs': rhs}}


def missing(t):
    return {'healthMissing': {'sourceUnit': t}}


def expected(spell, t):
    return {'spellExpectedHealing': {'spellId': {'spellId': spell}, 'targetUnit': t}}


def health_pct(t):
    return {'currentHealthPercent': {'sourceUnit': t}}


def aura_active(aura, t=None):
    v = {'auraIsActive': {'auraId': {'spellId': aura}}}
    if t:
        v['auraIsActive']['sourceUnit'] = t
    return v


def NOT(v):
    return {'not': {'val': v}}


def AND(*vs):
    return {'and': {'vals': list(vs)}}


FIT = [1.0]


def fits(spell, t):
    e = expected(spell, t)
    if FIT[0] != 1:
        e = {'math': {'op': 'OpMul', 'lhs': e, 'rhs': const(FIT[0])}}
    return cmp('OpGe', missing(t), e)


def rotation(k, role):
    hl = [int(x) for x in str(k['hl']).split(',') if x]
    fol = [int(x) for x in str(k['fol']).split(',') if x]
    hl_target = TANK if role == 'tank' and k['hl_target'] == 'tank' else MOST
    FIT[0] = k['fit']
    prepull = []
    items = [{'action': {'autocastOtherCooldowns': {}}, 'notes': 'Trinkets, potions and Demonic Rune.'}]
    if k['loh']:
        items.append(cast(LAY_ON_HANDS, LOW, cmp('OpLt', health_pct(LOW), const(k['loh'])),
                          f"Lay on Hands at whoever in my assignment is under {k['loh']:.0%} health: all my mana for a full heal."))
    if k['oath'] and role == 'tank':
        prepull.append({'action': cast(OATH, TANK)['action'], 'doAtValue': const('-1.5s')})
        items.append(cast(OATH, TANK, NOT(aura_active(OATH, TANK)), 'Oath of Light on the tank whenever it is off.'))
    if k['shock_emergency']:
        items.append(cast(HOLY_SHOCK, LOW, cmp('OpLt', health_pct(LOW), const(k['shock_emergency'])),
                          f"Holy Shock at whoever is under {k['shock_emergency']:.0%} health: an instant heal."))
    if k['emergency'] and fol:
        items.append(cast(FOL[fol[0]], LOW, cmp('OpLt', health_pct(LOW), const(k['emergency'])),
                          f"Flash of Light at whoever is under {k['emergency']:.0%} health: the fast heal."))
    if k['vigil']:
        items.append(cast(VIGIL, SELF, AND(NOT(aura_active(VIGIL_MARK, SELF)), cmp('OpGe', missing(MOST), const(k['vigil_missing']))),
                          "Light's Vigil on myself, for my party, once someone misses enough."))
        items.append(cast(HOLY_SHOCK, SELF, aura_active(VIGIL_MARK, SELF), 'Holy Shock at my Vigil: no cooldown, and it heals my party.'))
    if k['favor'] and hl:
        items.append(cast(DIVINE_FAVOR, None, fits(HL[hl[0]], hl_target),
                          f'Divine Favor for the Holy Light rank {hl[0]} that comes next.'))
    if k['shock_tank'] and role == 'tank':
        items.append(cast(HOLY_SHOCK, TANK, fits(HOLY_SHOCK, TANK), 'Holy Shock at the tank when it misses its heal.'))
    if k['shock_fill']:
        items.append(cast(HOLY_SHOCK, MOST, fits(HOLY_SHOCK, MOST), 'Holy Shock at the most hurt when it misses what Holy Shock heals.'))
    for r in hl:
        c = fits(HL[r], hl_target)
        if k['mana_hl']:
            c = AND(c, cmp('OpGt', {'currentManaPercent': {}}, const(k['mana_hl'])))
        items.append(cast(HL[r], hl_target, c, f'Holy Light rank {r}, the biggest that what is missing covers.'))
    if k['fol_fill']:
        for r in fol:
            items.append(cast(FOL[r], MOST, fits(FOL[r], MOST), f'Flash of Light rank {r} at the most hurt.'))
    return {'type': 'TypeAPL', 'prepullActions': prepull, 'priorityList': items}


def request(k, role):
    base = json.load(open('tools/healbench/req/shaman_restoration%s.json' % ('' if role == 'tank' else '_raid')))
    req = copy.deepcopy(base)
    sham = base['raid']['parties'][0]['players'][0]
    ench = {i: it.get('enchant') for i, it in enumerate(sham['equipment']['items'])}
    items = [{'id': 990300 + n} for n in range(1, 13)] + [{'id': 992014}, {'id': 992015}, {'id': 990313}, {'id': 990314}, {'id': k['libram']}]
    # The shaman's enchants slot for slot (head, neck, back, chest, wrist, legs, feet), and the caster
    # one-hand's and off-hand's: Mighty Intellect (Thorium Brotherhood) and Off-Hand - Wisdom.
    for i in range(12):
        if ench.get(i):
            items[i]['enchant'] = ench[i]
    items[14]['enchant'] = 2568
    items[15]['enchant'] = 1219579
    p = {
        'race': k['race'],
        'class': 'ClassPaladin',
        'equipment': {'items': items},
        'consumes': sham['consumes'],
        'buffs': sham['buffs'],
        'talentsString': k['talents'],
        'level': 60,
        'profession1': 'Engineering',
        'profession2': 'Alchemy',
        'rotation': rotation(k, role),
        'reactionTimeMs': sham.get('reactionTimeMs', 200),
        'distanceFromTarget': 20,
        'holyPaladin': {'options': {'aura': k['aura']} if k['aura'] != 'NoPaladinAura' else {}},
    }
    req['raid']['parties'][0]['players'][0] = p
    return req


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument('--out')
    ap.add_argument('--role', default='tank')
    ap.add_argument('--set', action='append', default=[])
    a = ap.parse_args()
    k = dict(DEFAULTS)
    for kv in a.set:
        key, v = kv.split('=', 1)
        t = type(DEFAULTS[key])
        k[key] = t(v) if t is not str else v
    if a.out:
        json.dump(request(k, a.role), open(a.out, 'w'), indent=2)
        return
    for role, name in (('tank', 'paladin_holy'), ('raid', 'paladin_holy_raid')):
        json.dump(request(k, role), open(f'tools/healbench/req/{name}.json', 'w'), indent=2)
        os.makedirs('ui/holy_paladin/apls', exist_ok=True)
        apl = {'tank': 'tank_healer', 'raid': 'raid_healer'}[role]
        json.dump(rotation(k, role), open(f'ui/holy_paladin/apls/{apl}.apl.json', 'w'), indent=2)
    gear = request(k, 'tank')['raid']['parties'][0]['players'][0]['equipment']
    json.dump(gear, open('ui/holy_paladin/gear_sets/level60_phase1_synthetic.gear.json', 'w'))


if __name__ == '__main__':
    main()
