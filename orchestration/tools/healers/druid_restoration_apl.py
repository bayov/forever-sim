"""The Restoration druid's healer rotations, with knobs, as APL JSON.

    python3 genapl.py tank|raid OUT.json [key=value ...]
"""
import json, sys

HT = {11: 25297, 10: 9889, 9: 9888, 8: 9758, 7: 8903, 6: 6778, 5: 9889 and 5189, 4: 5188}
RG = {9: 9858, 8: 9857, 7: 9856, 6: 9750, 5: 8941, 4: 8940}
RJ = {11: 25299, 10: 9841, 9: 9840, 8: 9839, 7: 8910, 6: 3627, 5: 2091}
TRANQ = 9863
WG = {3: 1238215, 2: 1238214, 1: 408120}
SWIFTMEND, NS, INNERVATE = 18562, 17116, 29166


def unit(t):
    if t == 'Tank':
        return {'type': 'Tank', 'index': 0}
    return {'type': t}


def cast(spell, target=None):
    c = {'spellId': {'spellId': spell}}
    if target:
        c['target'] = unit(target)
    return {'castSpell': c}


def const(v):
    return {'const': {'val': str(v)}}


def cmp(op, lhs, rhs):
    return {'cmp': {'op': op, 'lhs': lhs, 'rhs': rhs}}


def AND(*vals):
    vals = [v for v in vals if v]
    return vals[0] if len(vals) == 1 else {'and': {'vals': vals}}


def NOT(v):
    return {'not': {'val': v}}


def missing(t):
    return {'healthMissing': {'sourceUnit': unit(t)}}


def hp(t):
    return {'currentHealthPercent': {'sourceUnit': unit(t)}}


def expected(spell, t, scale=1.0):
    e = {'spellExpectedHealing': {'spellId': {'spellId': spell}, 'targetUnit': unit(t)}}
    if scale != 1.0:
        e = {'math': {'op': 'OpMul', 'lhs': e, 'rhs': const(scale)}}
    return e


def hot_active(spell, t):
    return {'dotIsActive': {'spellId': {'spellId': spell}, 'targetUnit': unit(t)}}


def hot_remaining(spell, t):
    return {'dotRemainingTime': {'spellId': {'spellId': spell}, 'targetUnit': unit(t)}}


def mana():
    return {'currentManaPercent': {}}


def line(action, cond=None, notes=''):
    a = {'action': dict(action)}
    if cond:
        a['action']['condition'] = cond
    if notes:
        a['notes'] = notes
    return a


def bound(spell, t, scale, notes):
    return line(cast(spell, t), cmp('OpGe', missing(t), expected(spell, t, scale)), notes)


def build(role, k):
    pl = [line({'autocastOtherCooldowns': {}}, notes='Racials, trinkets, potions, Demonic Rune and Nature\'s Swiftness.')]
    who = 'Tank' if role == 'tank' else 'MostHurt'
    pl.append(line(cast(INNERVATE), cmp('OpLe', mana(), const(k['innervate'])),
                   f'Innervate on ourselves at {int(100 * float(k["innervate"]))}% mana.'))
    # Emergency: someone in the assignment low.
    low = float(k['low'])
    if k.get('ns', '0') != '0':
        ns_at = float(k['ns'])
        pl.append(line(cast(NS), cmp('OpLt', hp('LowestHealth'), const(ns_at)),
                       f'Nature\'s Swiftness when someone in my assignment is under {int(100 * ns_at)}% health...'))
        pl.append(line(cast(HT[11], 'LowestHealth'), {'auraIsActive': {'auraId': {'spellId': NS}}},
                       '...and an instant Healing Touch rank 11 at them.'))
    if k.get('swiftmend_low', '1') == '1':
        pl.append(line(cast(SWIFTMEND, 'LowestHealth'), cmp('OpLt', hp('LowestHealth'), const(low)),
                       f'Swiftmend whoever in my assignment is under {int(100 * low)}% health and has my HoT.'))
    pl.append(line(cast(RG[9], 'LowestHealth'), cmp('OpLt', hp('LowestHealth'), const(low)),
                   f'Regrowth at whoever is under {int(100 * low)}%: the fast heal.'))
    if role == 'tank':
        if k.get('tranq', '0') != '0':
            pass
        rj = RJ[int(k['rejuv'])]
        pl.append(line(cast(rj, 'Tank'), NOT(hot_active(rj, 'Tank')), 'Rejuvenation on the tank whenever it\'s off.'))
        if k.get('regrowth_hot', '1') == '1':
            rg = RG[int(k['regrowth'])]
            pl.append(line(cast(rg, 'Tank'), AND(NOT(hot_active(rg, 'Tank')), cmp('OpGe', missing('Tank'), expected(rg, 'Tank', float(k['rg_scale'])))),
                           'Regrowth on the tank when its HoT is off and what it misses covers the heal.'))
        if k.get('swiftmend', '1') == '1':
            pl.append(line(cast(SWIFTMEND, 'Tank'), cmp('OpGe', missing('Tank'), expected(SWIFTMEND, 'Tank', float(k['sm_scale']))),
                           'Swiftmend at the tank when what it misses covers it.'))
        for r in [int(x) for x in k['ht'].split(',')]:
            pl.append(bound(HT[r], 'Tank', float(k['ht_scale']), f'Healing Touch rank {r} at the tank, the biggest that what it misses covers.'))
        if k.get('filler', '1') == '1':
            # The tank topped: HoTs on the most hurt of the rest.
            pl.append(line(cast(rj, 'MostHurt'), AND(NOT(hot_active(rj, 'MostHurt')), cmp('OpGe', missing('MostHurt'), expected(rj, 'MostHurt'))),
                           'Rejuvenation at the most hurt of the rest when it fits.'))
    else:
        if 'wg' in k and k['wg'] != '0':
            wg = WG[int(k['wg'])]
            pl.append(line(cast(wg, 'MostHurt'), cmp('OpGe', missing('MostHurt'), expected(wg, 'MostHurt', float(k['wg_scale']))),
                           'Wild Growth at the most hurt: a HoT on its whole party.'))
        if k.get('tranq', '0') != '0':
            pl.append(line(cast(TRANQ), cmp('OpLt', hp('LowestHealth'), const(k['tranq'])),
                           'Tranquility when the party is low.'))
        rj = RJ[int(k['rejuv'])]
        rj2 = RJ[int(k['rejuv2'])] if k.get('rejuv2') else None
        none_up = AND(NOT(hot_active(rj, 'MostHurt')), NOT(hot_active(rj2, 'MostHurt')) if rj2 else None)
        pl.append(line(cast(rj, 'MostHurt'), AND(none_up, cmp('OpGe', missing('MostHurt'), expected(rj, 'MostHurt', float(k['rj_scale'])))),
                       f'Rejuvenation rank {k["rejuv"]} at the most hurt when it has none and it fits.'))
        if rj2:
            pl.append(line(cast(rj2, 'MostHurt'), AND(none_up, cmp('OpGe', missing('MostHurt'), expected(rj2, 'MostHurt', float(k['rj_scale'])))),
                           f'Rejuvenation rank {k["rejuv2"]} for a smaller hole.'))
        if k.get('swiftmend', '1') == '1':
            pl.append(line(cast(SWIFTMEND, 'MostHurt'), cmp('OpGe', missing('MostHurt'), expected(SWIFTMEND, 'MostHurt', float(k['sm_scale']))),
                           'Swiftmend at the most hurt when what it misses covers it.'))
        for r in [int(x) for x in k['rg'].split(',') if x]:
            rg = RG[r]
            pl.append(line(cast(rg, 'MostHurt'), AND(NOT(hot_active(rg, 'MostHurt')), cmp('OpGe', missing('MostHurt'), expected(rg, 'MostHurt', float(k['rg_scale'])))),
                           f'Regrowth rank {r} at the most hurt.'))
        for r in [int(x) for x in k['ht'].split(',') if x]:
            pl.append(bound(HT[r], 'MostHurt', float(k['ht_scale']), f'Healing Touch rank {r} at the most hurt.'))
    return {'type': 'TypeAPL', 'priorityList': pl}


DEFAULTS = {
    'tank': dict(innervate='0.35', low='0.35', rejuv='11', regrowth='9', rg_scale='1', sm_scale='1', ht='11,10,9', ht_scale='1'),
    'raid': dict(innervate='0.35', low='0.35', rejuv='11', rj_scale='1', sm_scale='1', rg='9', rg_scale='1', ht='11,9', ht_scale='1', wg='3', wg_scale='1'),
}

if __name__ == '__main__':
    role, out = sys.argv[1], sys.argv[2]
    k = dict(DEFAULTS[role])
    for a in sys.argv[3:]:
        key, val = a.split('=', 1)
        k[key] = val
    json.dump(build(role, k), open(out, 'w'), indent=1)
