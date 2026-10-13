"""Loaders for the level 60 BiS budget analysis (orchestration/bis_budget_rules.md).

Three item sets in one stat vocabulary (amounts in the item's own units, ratings in rating points):
  era60()    Classic Era items (Wowhead classic gear planner dump, versionNum 11300), with source zone / npc.
  sod60()    Season of Discovery items from wowsims/sod assets/database/db.json (ids >= 200000, i.e. SoD's own
             records, phase >= 4: the level 60 phases), deduplicated by name, ilvl and stats.
  forever30() Forever listing items of ilvl 5-30 (b30_data.forever) with the tank-stat vocabulary below.

Vocabulary (b30_data's, extended):
  str agi sta int spi           primaries
  sp heal school:xxx            spell damage+healing, healing on top, one school's damage
  ap rap fap                    attack power, ranged-only AP, feral-only AP
  crit hit                      melee / ranged crit and hit rating (14 per 1% crit, 10 per 1% hit)
  spcrit sphit                  spell crit and hit rating (Era items state them separately)
  acrit ahit                    SoD's "all spells and attacks" crit and hit (one stat covers both)
  mp5
  haste mhaste sphaste exp      SoD only: haste in % (all / melee / spell), expertise points
  def dodge parry block blockv  defense skill, dodge / parry / block in %, block value (tank stats)
Anything else (resistances, bonus armor, spell penetration, health regen, weapon skill)
marks the row 'other' and the fits leave it out.

Effects: every item whose Wowhead tooltip carries a Use:, Chance on hit: or a non-stat Equip: line has
'effect' True; 'use' holds its Use: texts (trinket analysis). SoD rows without a tooltip in the dump get
effect None (unknown) and the fits leave them out.
"""
import collections, json, os, re, sys

sys.path.insert(0, os.path.dirname(__file__))
from b30_data import normalize as _b30_normalize, forever as _b30_forever, INPUTS, SLOT, PRIM

SOD_DB = os.environ.get('SOD_DB', '/home/user/wowsims/sod/assets/database/db.json')

TANK = ['def', 'dodge', 'parry', 'block', 'blockv']
PHYS = {'str', 'agi', 'ap', 'rap', 'fap', 'crit', 'hit', 'exp', 'mhaste'}
SPELL = {'sp', 'heal', 'spcrit', 'sphit', 'mp5', 'sphaste'}

TIER = {'head': 'T1 head/chest/legs', 'chest': 'T1 head/chest/legs', 'legs': 'T1 head/chest/legs',
        'shoulder': 'T2 shoulder/hands/waist/feet', 'hands': 'T2 shoulder/hands/waist/feet',
        'waist': 'T2 shoulder/hands/waist/feet', 'feet': 'T2 shoulder/hands/waist/feet',
        'neck': 'T3 neck/back/wrist/finger', 'back': 'T3 neck/back/wrist/finger',
        'wrist': 'T3 neck/back/wrist/finger', 'finger': 'T3 neck/back/wrist/finger',
        'two-hand': 'two-hand', 'one-hand': 'one-hand', 'main-hand': 'one-hand', 'off-hand weapon': 'one-hand',
        'held off-hand': 'held off-hand', 'shield': 'shield', 'ranged': 'ranged', 'wand': 'wand', 'thrown': 'ranged',
        'relic': 'relic', 'trinket': 'trinket'}
TIERS = ['T1 head/chest/legs', 'T2 shoulder/hands/waist/feet', 'T3 neck/back/wrist/finger']
ARMOR = [s for s, t in TIER.items() if t in TIERS]


def kind(stat):
    """Weight group of a vocabulary key. Spell crit / hit keep their own weight here so Q3 can compare."""
    if stat in PRIM:
        return 'prim'
    if stat.startswith('school:'):
        return 'school'
    return {'acrit': 'crit', 'ahit': 'hit'}.get(stat, stat)


def archetype(v):
    """caster / physical / hybrid / tank from the stats an item carries (amounts ignored).
    Hybrid: physical stats (Str, Agi, AP, melee/ranged crit or hit) together with spell stats (spell power,
    healing, MP5, spell crit or hit). Int and Spirit alone do not make a piece a caster or hybrid piece:
    Agi + Int (hunter-type mail) is physical, sub-archetype 'physical+int'."""
    ks = set(v)
    school = any(k.startswith('school:') for k in ks)
    if ks & set(TANK):
        return 'tank'
    ph = bool(ks & PHYS)
    sp = bool(ks & SPELL) or school
    if ph and sp:
        return 'hybrid'
    if ph:
        return 'physical'
    if sp or ks & {'int', 'spi'}:
        return 'caster'
    return 'stamina'


def subtype(v):
    a = archetype(v)
    ks = set(v)
    if a == 'physical':
        return 'physical+int' if ks & {'int', 'spi', 'mp5'} else 'physical'
    if a == 'caster':
        if 'heal' in ks:
            return 'healer'
        if 'sp' in ks or any(k.startswith('school:') for k in ks) or ks & {'spcrit', 'sphit'}:
            return 'caster dps'
        if 'mp5' in ks:
            return 'caster mp5'
        return 'caster int/spi'
    return a


# ---------------------------------------------------------------- effects (tooltips)
_STATLIKE = re.compile(
    r'Increases (damage and healing|healing done|damage done by (Arcane|Fire|Frost|Holy|Nature|Shadow))'
    r'|Improves your chance to (hit|get a critical)|Restores \d+ mana per|\+\d+ (Ranged )?Attack Power'
    r'|Increases your chance to hit|Increases your critical strike|Increased Defense|Increases your chance to (dodge|parry|block)'
    r'|Increases the block value|Increases attack power by \d+ in Cat, Bear'
    r'|Increases your (spell )?critical strike chance|Increases your hit chance')
_eff = None


def effects():
    """id -> {'use': [...], 'proc': [...], 'equip': [...]} for items whose tooltip has a non-stat line, plus the set
    of every id the tooltip dump covers."""
    global _eff
    if _eff is not None:
        return _eff
    out, seen = {}, set()
    for line in open(os.path.join(INPUTS, 'wowhead_item_tooltips.csv')):
        i = line.find(',')
        try:
            iid = int(line[:i])
            j = json.loads(line[i + 1:])
        except ValueError:
            continue
        seen.add(iid)
        tt = j.get('tooltip', '')
        tt = re.sub(r'<br\s*/?>', '\n', tt)
        tt = re.sub(r'<[^>]*>', ' ', tt).replace('&nbsp;', ' ')
        tt = re.sub(r'[ \t]+', ' ', tt)
        # set bonuses are not item effects
        tt = re.split(r'\(\d\) Set ?:', tt)[0]
        e = {'use': [], 'proc': [], 'equip': []}
        for m in re.finditer(r'(Equip:|Use:|Chance on hit:)\s*(.*?)(?=Equip:|Use:|Chance on hit:|Sell Price|Durability|Requires|Dropped by|$)', tt, re.S):
            k, text = m.group(1), ' '.join(m.group(2).split())
            if k == 'Use:':
                e['use'].append(text)
            elif k == 'Chance on hit:':
                e['proc'].append(text)
            elif not _STATLIKE.search(text):
                e['equip'].append(text)
        if any(e.values()):
            out[iid] = e
    _eff = (out, seen)
    return _eff


# ---------------------------------------------------------------- Era
_TANKMAP = {'def': 'def', 'dodgepct': 'dodge', 'parrypct': 'parry', 'blockpct': 'block', 'blockamount': 'blockv',
            'feratkpwr': 'fap'}


def _norm_era(eq):
    v, other = _b30_normalize(eq)
    for src, dst in _TANKMAP.items():
        if eq.get(src):
            v[dst] = eq[src]
            other.discard(src)
    return v, other


def _row(src, id_, name, ilvl, q, slot, v, other, **kw):
    eff, seen = effects()
    e = eff.get(id_)
    r = {'src': src, 'id': id_, 'name': name, 'ilvl': ilvl, 'q': q, 'slot': slot, 'tier': TIER.get(slot, slot),
         'v': v, 'other': other, 'effect': (bool(e) if id_ in seen else None), 'fx': e or {}}
    r.update(kw)
    r['arch'] = archetype(v)
    r['sub'] = subtype(v)
    return r


PHASE1_NPCS = {10184: 'Onyxia', 6109: 'Azuregos', 12397: 'Lord Kazzak'}
PRERAID_ZONES = {1583: 'Blackrock Spire', 1584: 'Blackrock Depths', 2017: 'Stratholme', 2057: 'Scholomance'}


def era60(min_ilvl=45, max_ilvl=100, quals=(3, 4)):
    """Era rows (quality 3-4) with 'zone', 'npc' and 'p1': 'MC', 'Onyxia', 'world boss', 'pre-raid dungeon' or None
    (Era Phase 1 content: Molten Core, Onyxia, Azuregos / Kazzak, and BRD / BRS / Strat / Scholo at ilvl 60-66)."""
    pat = re.compile(r'^(?:WH\.setPageData\("wow\.gearPlanner\.classic\.item", \{)?"(\d+)":(\{.*\}),?\s*(\}\);)?\s*$')
    rows = []
    for line in open(os.path.join(INPUTS, 'wowhead_gearplannerdb.txt')):
        m = pat.match(line)
        if not m:
            continue
        it = json.loads(m.group(2))
        if it.get('versionNum') != 11300 or it.get('quality') not in quals or not min_ilvl <= it.get('itemLevel', 0) <= max_ilvl:
            continue
        st = it.get('stats', {})
        slot = SLOT.get(st.get('slotbak', it.get('inventoryType')))
        if slot is None:
            continue
        if slot == 'ranged' and it.get('subclass') == 19:
            slot = 'wand'
        v, other = _norm_era(st)
        sm = it.get('sourcemore') or [{}]
        zone = next((s.get('z') for s in sm if s.get('z')), None)
        npc = next((s.get('ti') for s in sm if s.get('t') == 1 and s.get('ti')), None)
        p1 = None
        if zone == 2717:
            p1 = 'Molten Core'
        elif npc in PHASE1_NPCS:
            p1 = 'Onyxia' if npc == 10184 else 'world boss'
        elif zone in PRERAID_ZONES and 60 <= it['itemLevel'] <= 66:
            p1 = 'pre-raid dungeon'
        rows.append(_row('era', it['id'], it['name'], it['itemLevel'], it['quality'], slot, v, other,
                         zone=zone, npc=npc, p1=p1, sub_class=it.get('subclass'), classes=st.get('classes')))
    return rows


# ---------------------------------------------------------------- SoD
SOD_TYPE = {1: 'head', 2: 'neck', 3: 'shoulder', 4: 'back', 5: 'chest', 6: 'wrist', 7: 'hands', 8: 'waist', 9: 'legs',
            10: 'feet', 11: 'finger', 12: 'trinket'}
SOD_HAND = {1: 'main-hand', 2: 'one-hand', 3: 'off-hand weapon', 4: 'two-hand'}
SCH = ['arc', 'fir', 'fro', 'hol', 'nat', 'sha']


def _norm_sod(s):
    v, other = {}, set()
    for i, k in enumerate(PRIM_ORDER):
        if s[i]:
            v[k] = s[i]
    if s[5] or s[42]:
        v['sp'] = s[5] + s[42]
    if s[41]:
        v['heal'] = s[41]
    for j, sc in enumerate(SCH):
        if s[6 + j]:
            v['school:' + sc] = s[6 + j]
    if s[12]:
        v['mp5'] = s[12]
    # hit / crit: SoD's "+1% hit with all spells and attacks" sets both; equal values count once (ahit / acrit)
    for mel, spl, both, mk, sk, f in ((18, 13, 'ahit', 'hit', 'sphit', 10), (19, 14, 'acrit', 'crit', 'spcrit', 14)):
        a, b = s[mel], s[spl]
        if a and a == b:
            v[both] = f * a
        else:
            if a:
                v[mk] = f * a
            if b:
                v[sk] = f * b
    if s[17]:
        v['ap'] = s[17]
    if s[27] and s[27] > s[17]:
        v['rap'] = s[27] - s[17]
    if s[43]:
        v['fap'] = s[43]
    for i, k in ((28, 'def'), (29, 'block'), (30, 'blockv'), (31, 'dodge'), (32, 'parry')):
        if s[i]:
            v[k] = s[i]
    # haste: SoD's "attack and casting speed" sets both; equal values count once (in %)
    if s[20] and s[20] == s[15]:
        v['haste'] = s[20]
    else:
        if s[20]:
            v['mhaste'] = s[20]
        if s[15]:
            v['sphaste'] = s[15]
    if s[22]:
        v['exp'] = s[22]
    for i, k in ((16, 'splpen'), (21, 'arpen'), (23, 'mana'),
                 (33, 'resil'), (34, 'health'), (35, 'res'), (36, 'res'), (37, 'res'), (38, 'res'), (39, 'res'), (40, 'bonusarmor')):
        if len(s) > i and s[i]:
            other.add(k)
    if any(x < 0 for x in v.values()):
        other.add('negative')
    return v, other


PRIM_ORDER = ['str', 'agi', 'sta', 'int', 'spi']


def sod60(min_phase=4, quals=(4,), min_req=58):
    d = json.load(open(SOD_DB))['items']
    rows, seen = [], set()
    for it in d:
        if it['id'] < 200000 or it.get('phase', 0) < min_phase or it.get('quality') not in quals or it.get('requiresLevel', 0) < min_req:
            continue
        t = it['type']
        if t in SOD_TYPE:
            slot = SOD_TYPE[t]
        elif t == 13:
            wt = it.get('weaponType')
            slot = 'shield' if wt == 7 else 'held off-hand' if wt == 5 else SOD_HAND.get(it.get('handType'))
        elif t == 14:
            rw = it.get('rangedWeaponType')
            slot = 'wand' if rw == 8 else 'relic' if rw in (4, 5, 7, 9) else 'thrown' if rw == 6 else 'ranged'
        else:
            slot = None
        if slot is None:
            continue
        v, other = _norm_sod(it['stats'])
        key = (it['name'], it['ilvl'], tuple(sorted(v.items())))
        if key in seen:
            continue
        seen.add(key)
        rows.append(_row('sod', it['id'], it['name'], it['ilvl'], it['quality'], slot, v, other, phase=it['phase'],
                         classes=tuple(it.get('classAllowlist', ())), set=it.get('setName'), armor_type=it.get('armorType'),
                         weapon_type=it.get('weaponType'), ranged_type=it.get('rangedWeaponType')))
    return rows


# ---------------------------------------------------------------- Forever ilvl <= 30
def forever30(min_ilvl=5, max_ilvl=30):
    d = json.load(open(os.path.join(INPUTS, 'forever_wowhead_items.json')))['items']
    rows = []
    for it in d:
        if not min_ilvl <= it.get('level', 0) <= max_ilvl or 200000 <= it['id'] < 250000:
            continue
        slot = SLOT.get(it.get('slot'))
        if slot is None or it.get('quality') not in (2, 3, 4):
            continue
        v, other = _norm_era(it.get('eq', {}))
        r = _row('forever-new' if it['id'] >= 250000 else 'classic-id', it['id'], it['name'], it['level'],
                 it.get('quality'), slot, v, other)
        r['effect'] = bool(r['fx'])  # Forever-new ids have no Classic tooltip: no known effect (as in b30_data)
        rows.append(r)
    return rows


def clean(rows, slots=None, effects_ok=False):
    """Usable for budget fits: some stat, nothing outside the vocabulary, no known effect (unknown counts as
    not clean unless effects_ok), slot in slots."""
    return [r for r in rows if r['v'] and not r['other'] and (effects_ok or r['effect'] is False)
            and (slots is None or r['slot'] in slots)]


if __name__ == '__main__':
    E = era60()
    S = sod60()
    for lab, R in (('Era q3-4 ilvl 45+', E), ('SoD level 60 epics (phase 4+)', S)):
        print(lab, len(R), 'clean', len(clean(R)), 'effect', sum(r['effect'] is True for r in R), 'unknown', sum(r['effect'] is None for r in R))
        print('  archetypes (clean):', collections.Counter(r['arch'] for r in clean(R)))
        print('  sub (clean):', collections.Counter(r['sub'] for r in clean(R)))
        print('  other keys:', collections.Counter(k for r in R for k in r['other']).most_common(12))
    print('Era P1:', collections.Counter((r['p1'], r['q']) for r in E if r['p1']))
