"""Shared loader for the ilvl <= 30 budget analysis (orchestration/forever_item_budget.md, section
"ilvl <= 30 only").

Two item sets in one stat vocabulary:
  forever(): Wowhead's Forever listing (assets/db_inputs/forever_wowhead_items.json in w-dash).
  era():     Classic Era items (versionNum 11300) from Wowhead's classic gear planner dump
             (assets/db_inputs/wowhead_gearplannerdb.txt), the same jsonequip keys.
Note: wago_db2_items.csv is ItemSparse of build 1.15.3.55646 (the 2024 Era/SoD client, not
Forever): it agrees with era() on primaries and is used only as a cross-check.

Stat vocabulary (amounts in the item's own units, ratings in rating points):
  str agi sta int spi   primaries
  sp      spell damage and healing (Forever splpwr, or the spldmg/splheal pair's common part)
  heal    healing on top of sp (Forever: splheal beside splpwr; Era: healing-only items)
  school  one school's spell damage (firsplpwr, ...; one term per school)
  ap      attack power (melee and ranged), rap ranged-only attack power
  crit    crit rating (Era %: x14), hit hit rating (Era %: x10); spell crit/hit fold in
  mp5     mana per 5 sec
Anything else (resistances, defense, block, bonus armor, health regen, skill, haste,
expertise, ...) marks the item 'other' and the fits leave it out.
"""
import json, os, re

INPUTS = os.environ.get('FOREVER_INPUTS', '/home/user/w-dash/assets/db_inputs')

SLOT = {1: 'head', 2: 'neck', 3: 'shoulder', 5: 'chest', 20: 'chest', 6: 'waist', 7: 'legs', 8: 'feet',
        9: 'wrist', 10: 'hands', 11: 'finger', 12: 'trinket', 16: 'back',
        13: 'one-hand', 21: 'main-hand', 22: 'off-hand weapon', 17: 'two-hand', 14: 'shield',
        23: 'held off-hand', 15: 'ranged', 26: 'ranged', 25: 'thrown', 28: 'relic'}
ARMOR = ['head', 'chest', 'legs', 'shoulder', 'hands', 'waist', 'feet', 'neck', 'back', 'wrist', 'finger']
WEAPON = ['two-hand', 'one-hand', 'main-hand', 'off-hand weapon', 'shield', 'held off-hand', 'ranged', 'thrown']
STD_MOD = {'head': 1, 'chest': 1, 'legs': 1, 'shoulder': .77, 'hands': .77, 'waist': .77, 'feet': .77,
           'neck': .5625, 'back': .5625, 'wrist': .5625, 'finger': .5625, 'trinket': .5625}
PRIM = ['str', 'agi', 'sta', 'int', 'spi']

IGNORE = {'slotbak', 'reqlevel', 'armor', 'classes', 'itemset', 'races', 'reqfaction', 'reqrep', 'reqskill',
          'reqskillrank', 'reqspell', 'dmgrange', 'speed', 'dps', 'dmgmax1', 'dmgmin1', 'dmgtype1', 'mledps',
          'mlespeed', 'mledmgmax', 'mledmgmin', 'rgddmgmax', 'rgddmgmin', 'rgddps', 'rgdspeed', 'sellprice',
          'buyprice', 'avgbuyout', 'appearances', 'displayid', 'dura', 'sheathtype', 'maxcount',
          'damageMinAll', 'damageMaxAll', 'nsockets'}
SCHOOLS = ['arcsplpwr', 'firsplpwr', 'frosplpwr', 'holsplpwr', 'natsplpwr', 'shasplpwr']
KNOWN = set(PRIM) | set(SCHOOLS) | {'splpwr', 'spldmg', 'splheal', 'atkpwr', 'mleatkpwr', 'rgdatkpwr',
                                     'critstrkrtng', 'hitrtng', 'mlecritstrkpct', 'rgdcritstrkpct',
                                     'splcritstrkpct', 'mlehitpct', 'rgdhitpct', 'splhitpct', 'manargn'}


def normalize(eq):
    """jsonequip dict -> (stats, other) where stats maps vocabulary keys to amounts and
    other is the set of keys the vocabulary has no place for."""
    g = lambda k: eq.get(k, 0) or 0
    other = {k for k, x in eq.items() if k not in IGNORE and k not in KNOWN and x}
    v = {k: g(k) for k in PRIM if g(k)}
    both = min(g('spldmg'), g('splheal'))
    sp = g('splpwr') + both + (g('spldmg') - both)  # damage-only spell power counts as sp
    heal = g('splheal') - both
    if sp:
        v['sp'] = sp
    if heal:
        v['heal'] = heal
    for s in SCHOOLS:
        if g(s):
            v['school:' + s[:3]] = g(s)
    ap = g('atkpwr') or g('mleatkpwr')
    rap = g('rgdatkpwr') - min(g('rgdatkpwr'), ap) if g('rgdatkpwr') else 0
    if ap:
        v['ap'] = ap
    if rap:
        v['rap'] = rap
    crit = g('critstrkrtng') + 14 * max(g('mlecritstrkpct'), g('rgdcritstrkpct'))
    hit = g('hitrtng') + 10 * max(g('mlehitpct'), g('rgdhitpct'))
    if crit:
        v['crit'] = crit
    if hit:
        v['hit'] = hit
    if g('splcritstrkpct'):
        v['spcrit'] = 14 * g('splcritstrkpct')
    if g('splhitpct'):
        v['sphit'] = 10 * g('splhitpct')
    if g('manargn'):
        v['mp5'] = g('manargn')
    if any(x < 0 for x in v.values()):
        other.add('negative')
    return v, other


def kind_of(stat):
    """Weight group of a vocabulary key (school:fir -> school, spcrit -> crit, sphit -> hit)."""
    if stat.startswith('school:'):
        return 'school'
    return {'spcrit': 'crit', 'sphit': 'hit'}.get(stat, stat)


_era_effects = None


def era_effects():
    """Item ids whose Classic tooltip carries a Use:, Chance on hit: or a non-stat Equip: line
    (procs and on-use effects the stat listing does not show)."""
    global _era_effects
    if _era_effects is not None:
        return _era_effects
    statlike = re.compile(r'Increases (damage and healing|healing done|damage done by (Arcane|Fire|Frost|Holy|Nature|Shadow))'
                          r'|Improves your chance to (hit|get a critical)|Restores \d+ mana per|\+\d+ (Ranged )?Attack Power'
                          r'|Increases your chance to hit|Increases your critical strike')
    out = set()
    for line in open(os.path.join(INPUTS, 'wowhead_item_tooltips.csv')):
        i = line.find(',')
        try:
            j = json.loads(line[i + 1:])
        except ValueError:
            continue
        tt = re.sub(r'<[^>]*>', ' ', j.get('tooltip', ''))
        for kind, text in re.findall(r'(Equip:|Use:|Chance on hit:)\s*([^.]*)', tt):
            if kind != 'Equip:' or not statlike.search(text):
                out.add(int(line[:i]))
    _era_effects = out
    return out


def _row(id_, name, ilvl, q, wslot, eq, src):
    slot = SLOT.get(wslot)
    if slot is None or q not in (2, 3, 4):
        return None
    v, other = normalize(eq)
    return {'id': id_, 'name': name, 'ilvl': ilvl, 'q': q, 'slot': slot, 'v': v, 'other': other, 'src': src,
            'effect': id_ in era_effects()}


def forever(max_ilvl=30, min_ilvl=5, sod_ids=False):
    """Forever listing rows, ilvl in [min_ilvl, max_ilvl]. src: 'classic-id' (an Era item id that
    Forever kept or reworked), 'forever-new' (id >= 250000) or 'sod-id' (200000-249999, Season of
    Discovery records the listing carries; left out unless sod_ids)."""
    d = json.load(open(os.path.join(INPUTS, 'forever_wowhead_items.json')))['items']
    rows = []
    for it in d:
        if not min_ilvl <= it.get('level', 0) <= max_ilvl:
            continue
        src = 'forever-new' if it['id'] >= 250000 else ('sod-id' if it['id'] >= 200000 else 'classic-id')
        if src == 'sod-id' and not sod_ids:
            continue
        r = _row(it['id'], it['name'], it['level'], it.get('quality'), it.get('slot'), it.get('eq', {}), src)
        if r:
            rows.append(r)
    return rows


def era(max_ilvl=200, min_ilvl=1):
    """Classic Era rows (gear planner versionNum 11300)."""
    rows = []
    pat = re.compile(r'^(?:WH\.setPageData\("wow\.gearPlanner\.classic\.item", \{)?"(\d+)":(\{.*\}),?\s*(\}\);)?\s*$')
    for line in open(os.path.join(INPUTS, 'wowhead_gearplannerdb.txt')):
        m = pat.match(line)
        if not m:
            continue
        it = json.loads(m.group(2))
        if it.get('versionNum') != 11300 or not min_ilvl <= it.get('itemLevel', 0) <= max_ilvl:
            continue
        st = it.get('stats', {})
        r = _row(it['id'], it['name'], it['itemLevel'], it.get('quality'), st.get('slotbak', it.get('inventoryType')), st, 'era')
        if r:
            rows.append(r)
    return rows


def clean(rows, slots=None, effects=False):
    """Rows usable for budget fits: some stat, nothing outside the vocabulary, (optionally) no
    known proc or on-use effect, slot in slots."""
    return [r for r in rows if r['v'] and not r['other'] and (effects or not r['effect'])
            and (slots is None or r['slot'] in slots)]
