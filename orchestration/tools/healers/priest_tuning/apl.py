"""A small builder for the priest's healing APLs (JSON, as the UI's apls/*.apl.json)."""
import json

TANK = {'type': 'Tank'}
MOSTHURT = {'type': 'MostHurt'}
LOWEST = {'type': 'LowestHealth'}
SELF = {'type': 'Self'}

# Spell ids
GH = {5: 25314, 4: 10965, 3: 10964, 2: 10963, 1: 2060}
HEAL = {4: 6064, 3: 6063, 2: 2055, 1: 2054}
FH = {7: 10917, 6: 10916, 5: 10915, 4: 9474, 3: 9473}
BH = {6: 1240774, 5: 1240773, 4: 1240772}
POH = {5: 25316, 4: 10961, 3: 10960, 2: 996, 1: 596}
RENEW = {10: 25315, 9: 10929, 8: 10928, 7: 10927, 6: 6078, 5: 6077}
PWS = {10: 10901, 9: 10900, 8: 10899, 7: 10898}
HN = {6: 27801, 5: 27800, 4: 27799}
POM = {3: 1240827, 2: 1240826, 1: 401859}
PEN = {4: 1316995, 3: 1240721}
INNER_FOCUS = 14751
PI = 10060


def const(v):
    return {'const': {'val': str(v)}}


def cmp(op, lhs, rhs):
    return {'cmp': {'op': op, 'lhs': lhs, 'rhs': rhs}}


def AND(*vals):
    return {'and': {'vals': list(vals)}}


def OR(*vals):
    return {'or': {'vals': list(vals)}}


def NOT(v):
    return {'not': {'val': v}}


def missing(unit):
    return {'healthMissing': {'sourceUnit': unit}}


def health_pct(unit):
    return {'currentHealthPercent': {'sourceUnit': unit}}


def mana_pct():
    return {'currentManaPercent': {}}


def expected(spell, unit):
    return {'spellExpectedHealing': {'spellId': {'spellId': spell}, 'targetUnit': unit}}


def dot_active(spell, unit):
    return {'dotIsActive': {'spellId': {'spellId': spell}, 'targetUnit': unit}}


def aura_active(aura, unit=None):
    a = {'auraIsActive': {'auraId': {'spellId': aura}}}
    if unit:
        a['auraIsActive']['sourceUnit'] = unit
    return a


def ready(spell):
    return {'spellIsReady': {'spellId': {'spellId': spell}}}


def cast(spell, target=None, cond=None, notes=None):
    a = {'castSpell': {'spellId': {'spellId': spell}}}
    if target:
        a['castSpell']['target'] = target
    item = {'action': a}
    if cond:
        item['action']['condition'] = cond
    if notes:
        item['notes'] = notes
    return item


def covers(spell, unit, share=1.0):
    """What's missing covers the heal's Expected Healing (share of it)."""
    rhs = expected(spell, unit)
    if share != 1.0:
        rhs = {'math': {'op': 'OpMul', 'lhs': rhs, 'rhs': const(share)}}
    return cmp('OpGe', missing(unit), rhs)


def autocast():
    return {'action': {'autocastOtherCooldowns': {}}, 'notes': 'Trinkets, potions and Demonic Rune.'}


def rotation(items, prepull=None):
    r = {'type': 'TypeAPL', 'priorityList': items}
    if prepull:
        r['prepullActions'] = prepull
    return r


def prepull_cast(spell, at, target=None):
    a = {'castSpell': {'spellId': {'spellId': spell}}}
    if target:
        a['castSpell']['target'] = target
    return {'action': a, 'doAtValue': const(at)}


def dump(r, path):
    json.dump(r, open(path, 'w'), indent=2)
