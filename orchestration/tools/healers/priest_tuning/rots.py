"""The priest's three healing rotations, with knobs. python3 rots.py NAME OUT [knob=value ...]"""
import sys, json
from apl import *

def k(knobs, name, default):
    v = knobs.get(name, default)
    if isinstance(default, bool):
        return v in (True, 'true', '1', 'yes') if isinstance(v, str) else v
    if isinstance(default, (int, float)) and isinstance(v, str):
        return type(default)(float(v))
    if isinstance(default, list) and isinstance(v, str):
        return [int(x) for x in v.split(',') if x]
    return v

def emergency(knobs, target):
    line = k(knobs, 'emergency', 0.35)
    return cast(FH[7], LOWEST, cmp('OpLt', health_pct(LOWEST), const(line)),
                f'Someone in my assignment is under {line:.0%} health: the fast heal.')

def inner_focus(knobs, spell, unit, share=1.0):
    return cast(INNER_FOCUS, None, covers(spell, unit, share), 'Inner Focus right before a big heal: free, and 25% more crit.')

def power_infusion(knobs):
    return cast(PI, SELF, cmp('OpGe', mana_pct(), const(0)) if False else None, 'Power Infusion on myself: 20% more healing for 15 sec.')

NAMES = {id(GH): 'Greater Heal', id(FH): 'Flash Heal', id(POH): 'Prayer of Healing'}
UNITS = {'Tank': 'the tank', 'MostHurt': 'the most hurt', 'LowestHealth': 'the lowest'}

def bound(spells, ranks, unit, share=1.0):
    # The ranks bound for a heal (at most 4), biggest first: each when what's missing covers it.
    name = NAMES.get(id(spells), 'the heal')
    what = 'what is missing covers its Expected Healing' if share >= 1 else f'what is missing covers {share:.0%} of its Expected Healing'
    note = f"{name} {', '.join(map(str, ranks))} at {UNITS[unit['type']]}: each rank, biggest first, once {what}."
    return [cast(spells[r], unit, covers(spells[r], unit, share), note if i == 0 else None) for i, r in enumerate(ranks)]

def holy_tank(knobs):
    items = [autocast(), emergency(knobs, TANK)]
    if k(knobs, 'pi', True):
        items.append(cast(PI, SELF, cmp('OpGe', missing(TANK), expected(GH[5], TANK)), 'Power Infusion on myself when a big heal is due: 20% more healing for 15 sec.'))
    gate = k(knobs, 'gate', 0.0)
    rich = cmp('OpGe', mana_pct(), const(gate)) if gate > 0 else None
    if k(knobs, 'pws', False):
        items.append(cast(PWS[10], TANK, rich, 'Power Word: Shield on the tank whenever Weakened Soul is off it, while mana lasts.'))
    if k(knobs, 'pom', True):
        items.append(cast(POM[3], TANK, None, 'Prayer of Mending on the tank whenever it is ready.'))
    if k(knobs, 'renew', True):
        items.append(cast(RENEW[10], TANK, NOT(dot_active(RENEW[10], TANK)), 'Renew on the tank whenever it is off.'))
    gh = k(knobs, 'gh', [5, 4, 2])
    fh = k(knobs, 'fh', [7])
    # share: the heal gate, the part of a heal's Expected Healing the tank must miss before it's cast
    # (healer_fairness-20261104-2).
    share = k(knobs, 'share', 1.0)
    if k(knobs, 'fhfirst', False):
        for r in fh:
            cond = covers(FH[r], TANK, share)
            items.append(cast(FH[r], TANK, AND(rich, cond) if rich else cond,
                              f'Flash Heal first while mana is over {gate:.0%}: more healing a second, if less a mana.' if r == fh[0] else ''))
    items.append(inner_focus(knobs, GH[gh[0]], TANK, k(knobs, 'ghshare', share)))
    items += bound(GH, gh, TANK, k(knobs, 'ghshare', share))
    items += bound(FH, fh, TANK, share)
    return rotation(items)

def disc_tank(knobs):
    items = [autocast(), emergency(knobs, TANK)]
    gate = k(knobs, 'gate', 0.0)
    rich = cmp('OpGe', mana_pct(), const(gate)) if gate > 0 else None
    if k(knobs, 'pi', True):
        items.append(cast(PI, SELF, cmp('OpGe', missing(TANK), expected(GH[5], TANK)), 'Power Infusion on myself when a big heal is due.'))
    pwsgate = k(knobs, 'pwsgate', 0.0)
    pwscond = None
    if pwsgate > 0:
        # Below the gate, a shield only when the tank is hurt enough to use it soon.
        pwscond = OR(cmp('OpGe', mana_pct(), const(pwsgate)), covers(GH[5], TANK, k(knobs, 'pwsneed', 0.5)))
    pwsnote = 'Power Word: Shield on the tank whenever Weakened Soul is off it (Renewed Hope cuts it short)'
    if pwsgate > 0:
        pwsnote += f"; under {pwsgate:.0%} mana only once what is missing covers {k(knobs, 'pwsneed', 0.5):.0%} of a Greater Heal 5"
    items.append(cast(PWS[10], TANK, pwscond, pwsnote + '.'))
    if k(knobs, 'penance', True):
        items.append(cast(PEN[4], TANK, covers(PEN[4], TANK, k(knobs, 'penshare', 1.0)), 'Penance once what is missing covers its 3 bolts.'))
    if k(knobs, 'pom', False):
        items.append(cast(POM[3], TANK, None, 'Prayer of Mending on the tank whenever it is ready.'))
    if k(knobs, 'renew', True):
        rg = k(knobs, 'renewgate', 0.0)
        rc = NOT(dot_active(RENEW[10], TANK))
        items.append(cast(RENEW[10], TANK, AND(rc, cmp('OpGe', mana_pct(), const(rg))) if rg > 0 else rc,
                          'Renew on the tank whenever it is off' + (f' and mana is over {rg:.0%}.' if rg > 0 else '.')))
    gh = k(knobs, 'gh', [5, 4, 2])
    fh = k(knobs, 'fh', [7])
    # share: the heal gate, the part of a heal's Expected Healing the tank must miss before it's cast
    # (healer_fairness-20261104-2).
    share = k(knobs, 'share', 1.0)
    if k(knobs, 'fhfirst', False):
        for r in fh:
            cond = covers(FH[r], TANK, share)
            items.append(cast(FH[r], TANK, AND(rich, cond) if rich else cond,
                              f'Flash Heal first while mana is over {gate:.0%}: more healing a second, if less a mana.' if r == fh[0] else ''))
    items.append(inner_focus(knobs, GH[gh[0]], TANK, k(knobs, 'ghshare', share)))
    items += bound(GH, gh, TANK, k(knobs, 'ghshare', share))
    items += bound(FH, fh, TANK, share)
    return rotation(items)

def holy_raid(knobs):
    items = [autocast(), emergency(knobs, LOWEST)]
    if k(knobs, 'pi', True):
        items.append(cast(PI, SELF, cmp('OpGe', missing(MOSTHURT), expected(POH[5], MOSTHURT)), 'Power Infusion on myself when the raid is hurt.'))
    pws = k(knobs, 'pws', 0.0)
    if pws > 0:
        items.append(cast(PWS[10], LOWEST, cmp('OpLt', health_pct(LOWEST), const(pws)), f'Power Word: Shield on whoever drops under {pws:.0%} health, for the next hit.'))
    if k(knobs, 'pom', True):
        items.append(cast(POM[3], MOSTHURT, None, 'Prayer of Mending at the most hurt whenever it is ready.'))
    first = k(knobs, 'first', 'poh')
    poh = k(knobs, 'poh', [5])
    fh = k(knobs, 'fh', [7, 5])
    gh = k(knobs, 'gh', [5, 2])
    blocks = {
        'poh': bound(POH, poh, MOSTHURT, k(knobs, 'pohshare', 1.0)) if poh else [],
        'fh': bound(FH, fh, MOSTHURT, k(knobs, 'share', 1.0)),
        'gh': ([inner_focus(knobs, GH[gh[0]], MOSTHURT)] + bound(GH, gh, MOSTHURT)) if gh else [],
        'renew': [cast(RENEW[k(knobs, 'renewrank', 10)], MOSTHURT, covers(RENEW[k(knobs, 'renewrank', 10)], MOSTHURT, k(knobs, 'renewshare', 1.0)), f"Renew at the most hurt once what is missing covers {k(knobs, 'renewshare', 1.0):.0%} of its Expected Healing.")] if k(knobs, 'renew', True) else [],
    }
    for name in k(knobs, 'order', 'poh,gh,renew,fh').split(','):
        items += blocks[name]
    return rotation(items)

ROTS = {'holy_tank': holy_tank, 'disc_tank': disc_tank, 'holy_raid': holy_raid}
if __name__ == '__main__':
    name, out = sys.argv[1], sys.argv[2]
    knobs = dict(a.split('=', 1) for a in sys.argv[3:])
    dump(ROTS[name](knobs), out)
