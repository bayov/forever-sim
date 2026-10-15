#!/usr/bin/env python3
"""build_data.py: turns out/*.json into the dashboard's dataset files in data/."""
import csv, glob, json, os, re, sys
from collections import defaultdict

sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
from dashpaths import D, TREE  # base dir and merged tree (env DASH_DIR / DASH_TREE)
OUT = os.path.join(D, 'data')
os.makedirs(OUT, exist_ok=True)
os.makedirs(OUT, exist_ok=True)
csv.field_size_limit(10**9)

RACE_NAMES = {'NightElf': 'Night Elf', 'SkyborneHighOrder': 'High Order', 'SkyborneWindshaper': 'Windshaper'}
CLASS_ORDER = ['Warrior', 'Rogue', 'Hunter', 'Paladin', 'Shaman', 'Druid', 'Mage', 'Warlock', 'Priest']
SLOTS = ['Head', 'Neck', 'Shoulders', 'Back', 'Chest', 'Wrist', 'Hands', 'Waist', 'Legs', 'Feet', 'Ring 1', 'Ring 2', 'Trinket 1', 'Trinket 2', 'Main hand', 'Off hand', 'Ranged']
OTHER = {'oOtherActionAttack': 'Melee', 'oOtherActionAttackt1': 'Melee (main hand)', 'oOtherActionAttackt2': 'Melee (off hand)',
         'oOtherActionShoot': 'Auto Shot', 'oOtherActionManaRegen': 'Mana regen', 'oOtherActionEnergyRegen': 'Energy regen',
         'oOtherActionFocusRegen': 'Focus regen', 'oOtherActionRefund': 'Refund', 'oOtherActionPotion': 'Potion', 'oOtherActionMove': 'Movement',
         'oOtherActionComboPoints': 'Combo points', 'oOtherActionManaGain': 'Mana gain', 'oOtherActionRageGain': 'Rage gain', 'oOtherActionDamageTaken': 'Damage taken', 'oOtherActionExplosives': 'Explosives', 'oOtherActionWait': 'Wait'}
# metrics keys use the enum's text; normalize to the log's keys
def norm_key(k):
    return k


DISPLAY = {'paladin__level_60': 'Ret twist', 'paladin__level_60_sub_holy': 'Ret sub-Holy', 'paladin__level_60_shock': 'Ret Shock',
           'priest__shadow': 'Shadow', 'priest__shadow-holy': 'Shadow + Holy', 'priest__smite': 'Smite',
           'shaman__elemental-level60': 'Elemental', 'shaman__elemental-level60-mana-tide': 'Elemental Mana Tide',
           'shaman__enhancement-level60': 'Enhancement', 'shaman__enhancement-imp-gw': 'Enh Imp GW',
           'shaman__enhancement-water-shield': 'Enh Water Shield', 'shaman__enhancement-mana-tide': 'Enh Mana Tide',
           'druid__feral-cat': 'Feral', 'hunter__survival': 'Survival', 'warrior__fury': 'Fury', 'warrior__fury_2h': 'Fury 2H'}


STAT_LABEL = {'AttackPower': 'Attack power', 'RangedAttackPower': 'Ranged attack power', 'MeleeHit': 'Melee hit %', 'MeleeCrit': 'Melee crit %',
              'MeleeHaste': 'Melee haste %', 'SpellPower': 'Spell power', 'SpellDamage': 'Spell damage', 'SpellHit': 'Spell hit %', 'SpellCrit': 'Spell crit %',
              'SpellHaste': 'Spell haste %', 'MP5': 'Mana per 5 s', 'FirePower': 'Fire power', 'FrostPower': 'Frost power', 'ShadowPower': 'Shadow power',
              'ArcanePower': 'Arcane power', 'NaturePower': 'Nature power', 'HolyPower': 'Holy power', 'ArmorPenetration': 'Armor penetration'}


RACE_ICON = {'Human': 'race_human_male', 'Dwarf': 'race_dwarf_male', 'Night Elf': 'race_nightelf_male', 'Gnome': 'race_gnome_male',
             'High Order': 'inv_feather_02', 'Orc': 'race_orc_male', 'Undead': 'race_scourge_male', 'Tauren': 'race_tauren_male',
             'Troll': 'race_troll_male', 'Windshaper': 'ability_druid_flightform'}
FACTION_ICON = {'Alliance': 'ui_allianceicon', 'Horde': 'ui_hordeicon'}
QUALITY = {0: 'poor', 1: 'common', 2: 'uncommon', 3: 'rare', 4: 'epic', 5: 'legendary'}
CLIENT_TREES = {c: json.load(open(os.path.join(TREE, f'tools/forever_talents/client/{c.lower()}.json')))['trees'] for c in CLASS_ORDER}
TREE_JSON = {c: json.load(open(os.path.join(TREE, f'ui/core/talents/trees/{c.lower()}.json'))) for c in CLASS_ORDER}


def race_name(r):
    return RACE_NAMES.get(r, r)


# ---- name sources ----
db = json.load(open(os.path.join(TREE, 'assets/database/db.json')))
item_names = {i['id']: i for i in db['items']}
enchant_names = {e['effectId']: e['name'] for e in db['enchants']}
spell_icons = {s['id']: s for s in db['spellIcons']}
# Forever spells the sim keys by a borrowed or renamed Classic/SoD spell id: show Forever's name and icon
FOREVER_SPELL = {425336: ('Rage of the Farseer', 'spell_nature_bloodlust'), 11713: ('Bane of Agony', None), 603: ('Bane of Doom', None)}
for _id, (_nm, _ic) in FOREVER_SPELL.items():
    _s = dict(spell_icons.get(_id) or {'id': _id})
    _s['name'] = _nm
    if _ic:
        _s['icon'] = _ic
    spell_icons[_id] = _s
wowhead = {}
wowhead_icon, wowhead_icon_by_name = {}, {}
with open(os.path.join(TREE, 'assets/db_inputs/wowhead_spell_tooltips.csv')) as f:
    for line in f:  # id,<json>: the JSON is not quoted, so split at the first comma
        row = line.rstrip('\n').split(',', 1)
        try:
            _j = json.loads(row[1])
            wowhead[int(row[0])] = _j['name']
            _ic = _j.get('icon')
            if _ic and 'temp' not in _ic:
                wowhead_icon[int(row[0])] = _ic
                wowhead_icon_by_name.setdefault(_j['name'].lower(), _ic)
        except Exception:
            pass
# Forever wiki ability icons, by name
wiki_icon = {}
for _cls, _lst in json.load(open(os.path.join(TREE, 'forever-wiki/data/abilities.json'))).items():
    for _a in _lst:
        if _a.get('icon') and _a.get('name'):
            wiki_icon.setdefault(_a['name'].lower(), _a['icon'])
talent_names = {}
for p in glob.glob(os.path.join(TREE, 'ui/core/talents/trees/*.json')):
    t = json.load(open(p))
    trees = t if isinstance(t, list) else t.get('trees', [])
    for tree in trees:
        for tal in tree.get('talents', []):
            for sid in tal.get('spellIds', []) or []:
                talent_names[sid] = tal.get('name')
item_csv = {}
item_csv_icon = {}
item_by_name = {}
with open(os.path.join(TREE, 'assets/db_inputs/wowhead_item_tooltips.csv')) as f:
    for line in f:  # id,<json>: the JSON is not quoted, so split at the first comma
        row = line.rstrip('\n').split(',', 1)
        try:
            j = json.loads(row[1])
            item_csv[int(row[0])] = j['name']
            item_csv_icon[int(row[0])] = j.get('icon')
            item_by_name.setdefault(j['name'].lower(), j.get('icon'))
        except Exception:
            pass
spell_by_name = {}
for _s in db['spellIcons']:
    spell_by_name.setdefault(_s['name'].lower(), _s.get('icon'))
racial_names = {}
MANUAL = {'i10646': 'Goblin Sapper Charge', 'i12662': 'Demonic Rune', 'i13444': 'Major Mana Potion', 'i5513': 'Mana Jade', 'i5514': 'Mana Agate', 'i8007': 'Mana Citrine', 'i8008': 'Mana Ruby', 's15851': 'Dragonbreath Chili',
          's11281': 'Backstab', 's11356': 'Deadly Poison', 's11780': 'Lash of Pain', 's1293813': 'Incinerate', 's1310786': 'Sniper Shot',
          's1311598': "Light's Vigil", 's1329': 'Mutilate', 's1329t1': 'Mutilate (main hand)', 's1329t2': 'Mutilate (off hand)', 's1382': 'Shadow Oil',
          's17261': 'Bite', 's20947': 'Seal of Command', 's20966': 'Judgement of Command', 's3009': 'Claw', 'i4395': 'Goblin Land Mine', 'i7676': 'Thistle Tea',
          **{f's11300t{i}': f'Eviscerate ({i} CP)' for i in range(1, 6)}, 's11300': 'Eviscerate',
          's1293527': 'Summon Hawk', 's1293527t1': 'Hawk 1', 's1293527t2': 'Hawk 2', 's460534': 'Cutthroat'}
MANUAL_ICON = {'s1293527': 'ability_hunter_pet_owl', 's460534': 'ability_rogue_ambush'}
for race, lst in json.load(open(os.path.join(TREE, 'ui/core/forever/racials.json'))).items():
    for r in lst:
        for v in r.get('variants', []):
            racial_names[v['spellId']] = r['name']
            if v.get('icon'):
                MANUAL_ICON.setdefault('s%d' % v['spellId'], v['icon'])
# Rank arrays in the sim's Go code: var ArcaneBlastSpellId = [N]int32{0, 400574, ...}
go_ranks = {}
for gp in glob.glob(os.path.join(TREE, 'sim/**/*.go'), recursive=True):
    for mm in re.finditer(r'(\w+?)SpellIds?\s*=\s*\[[^\]]*\]int32\{([^}]*)\}', open(gp).read()):
        base = re.sub(r'(?<=[a-z])(?=[A-Z])', ' ', mm.group(1)[0].upper() + mm.group(1)[1:])
        ids = [int(x) for x in re.findall(r'\d+', mm.group(2))]
        for rank, sid in enumerate(ids):
            if sid:
                go_ranks.setdefault(sid, (base, rank))
aura_labels = {}
for p in glob.glob(os.path.join(D, 'names', '*.tsv')):
    for line in open(p):
        parts = line.rstrip('\n').split('\t')
        if len(parts) < 5:
            continue
        sid, iid, oid, tag, label = int(parts[0]), int(parts[1]), int(parts[2]), int(parts[3]), parts[4]
        k = ('s%d' % sid) if sid else ('i%d' % iid) if iid else None
        if not k:
            continue
        if tag:
            k += 't%d' % tag
        aura_labels.setdefault(k, label)


def clean_label(s):
    s = re.sub(r'-\d+$', '', s)
    s = re.sub(r'(?<=[a-z])\d+$', '', s)
    s = re.sub(r'(?<=[a-z])(?=[A-Z])', ' ', s) if ' ' not in s else s
    return s.strip()


def name_of(key):
    if key in MANUAL:
        return MANUAL[key], None
    if key in OTHER:
        return OTHER[key], None
    m = re.match(r'^([sio])(\w+?)(?:t(-?\d+))?$', key)
    if not m:
        return key, None
    kind, n, tag = m.group(1), m.group(2), m.group(3)
    if kind == 'o':
        base = OTHER.get('o' + n, n)
        return base + (' (main hand)' if tag == '1' else ' (off hand)' if tag == '2' else ' (extra attacks)' if tag == '3' else ''), None
    n = int(n)
    rank = None
    if kind == 'i':
        it = item_names.get(n)
        nm = it['name'] if it else item_csv.get(n) or (clean_label(aura_labels['i%d' % n]) if 'i%d' % n in aura_labels else None) or ('Item %d' % n)
    else:
        si = spell_icons.get(n)
        lab = aura_labels.get(key) or aura_labels.get('s%d' % n)
        if si:
            nm, rank = si['name'], si.get('rank')
        elif lab:
            nm = clean_label(lab)
            mr = re.search(r'\(Rank (\d+)\)', nm)
            if mr:
                rank = int(mr.group(1))
                nm = nm[:mr.start()].strip()
        elif n in racial_names:
            nm = racial_names[n]
        elif n in go_ranks:
            nm, rank = go_ranks[n]
        elif n in wowhead:
            nm = wowhead[n]
        elif n in talent_names:
            nm = talent_names[n]
        else:
            nm = 'Spell %d' % n
    nm = re.sub(r'\s*-?\s*\(#\d+\)\d*\s*$', '', nm).strip(' -')  # aura labels: "Winter's Chill -  (#1)" -> "Winter's Chill"
    nm = re.sub(r'\s{2,}', ' ', re.sub(r'-\s*$', '', nm)).strip()
    return nm, rank


def spec_icon(cls, talents):
    parts = talents.split('-') + ['', '', '']
    spent = [sum(int(c) for c in p if c.isdigit()) for p in parts[:3]]
    return CLIENT_TREES[cls][spent.index(max(spent))]['icon']


SETUP_ICON = {'manaSpringTotem': 'spell_nature_manaregentotem', 'windfuryTotem': 'spell_nature_windfury', 'graceOfAirTotem': 'spell_nature_invisibilitytotem',
              'strengthOfEarthTotem': 'spell_nature_earthbindtotem', 'leaderOfThePack': 'spell_nature_unyeildingstamina', 'moonkinAura': 'spell_nature_moonglow',
              'trueshotAura': 'ability_trueshot', 'bloodPact': 'spell_shadow_bloodboil', 'battleShout': 'ability_warrior_battleshout',
              'devotionAura': 'spell_holy_devotionaura', 'retributionAura': 'spell_holy_auraoflight', 'sanctityAura': 'spell_holy_mindvision',
              'judgementOfWisdom': 'spell_holy_righteousnessaura', 'judgementOfLight': 'spell_holy_healingaura', 'judgementOfTheCrusader': 'spell_holy_holysmite',
              'curseOfElements': 'spell_shadow_chilltouch', 'curseOfRecklessness': 'spell_shadow_unholystrength', 'curseOfShadow': 'spell_shadow_curseofachimonde',
              'curseOfWeakness': 'spell_shadow_curseofmannoroth', 'sunderArmor': 'ability_warrior_sunder', 'exposeArmor': 'ability_warrior_riposte',
              'faerieFire': 'spell_nature_faeriefire', 'demoralizingShout': 'ability_warrior_warcry', 'demoralizingRoar': 'ability_druid_demoralizingroar',
              'thunderClap': 'spell_nature_thunderclap', 'insectSwarm': 'spell_nature_insectswarm', 'scorpidSting': 'ability_hunter_criticalshot',
              'shadowProtection': 'spell_shadow_antishadow', 'fireResistanceAura': 'spell_fire_sealoffire', 'frostResistanceAura': 'spell_frost_wizardmark',
              'blessingOfKings': 'spell_magic_magearmor', 'blessingOfMight': 'spell_holy_fistofjustice', 'blessingOfWisdom': 'spell_holy_sealofwisdom',
              'arcaneBrilliance': 'spell_holy_magicalsentry', 'divineSpirit': 'spell_holy_divinespirit', 'giftOfTheWild': 'spell_nature_regeneration',
              'powerWordFortitude': 'spell_holy_wordfortitude', 'thorns': 'spell_nature_thorns', 'improvedScorch': 'spell_fire_soulburn',
              'wintersChill': 'spell_frost_chillingblast', 'shadowWeaving': 'spell_shadow_blackplague', 'stormstrike': 'spell_holy_sealofmight'}


CONSUME_ICON = {'scroll of imbue greater flame': 'inv_10_inscription2_repcontracts_scroll_02_uprez_color2',
                'scroll of imbue greater frost': 'inv_10_inscription2_repcontracts_scroll_02_uprez',
                'scroll of imbue precision': 'inv_10_inscription2_repcontracts_scroll_02_uprez_color4', 'flask of supreme power': 'inv_potion_41', 'flask of the titans': 'inv_potion_62', 'flask of distilled wisdom': 'inv_potion_97',
                'elixir of the mongoose': 'inv_potion_32', 'elixir of greater agility': 'inv_potion_94', 'juju might': 'inv_misc_monsterscales_07',
                'juju power': 'inv_misc_monsterscales_11', 'elixir of ferocity': 'inv_potion_61', 'greater arcane elixir': 'inv_potion_25',
                'elixir of shadow power': 'inv_potion_46', 'elixir of greater firepower': 'inv_potion_60', 'elixir of firepower': 'inv_potion_33',
                'elixir of frost power': 'inv_potion_03', 'elixir of nature power': 'inv_potion_06', 'greater mageblood elixir': 'inv_potion_45',
                'elixir of the owl': 'inv_potion_10', 'elixir of greater intellect': 'inv_potion_10', 'elixir of sages': 'inv_potion_29',
                'elixir of greater spirit': 'inv_potion_09', 'major mana potion': 'inv_potion_76', 'major frenzy potion': 'inv_potion_41',
                'mighty rage potion': 'inv_potion_41', 'major spellblasting potion': 'inv_potion_25', 'conjured demonic rune': 'inv_misc_rune_04',
                'conjured rogue thistle tea': 'inv_drink_milk_05', 'goblin sapper': 'spell_fire_selfdestruct', 'explosive goblin land mine': 'inv_misc_bomb_05',
                'lobster stew': 'inv_misc_food_86', 'grilled squid': 'inv_misc_fish_13', 'smoked desert dumpling': 'inv_misc_food_64',
                'flank au poivre': 'inv_misc_food_65', 'wizard oil': 'inv_potion_105', 'shadow oil': 'inv_potion_23',
                'dense sharpening stone': 'inv_stone_sharpeningstone_05', 'instant poison': 'ability_poisons', 'deadly poison': 'ability_rogue_dualweild',
                'cerebral cortex compound': 'inv_potion_32', 'ground scorpok assay': 'inv_misc_dust_02', 'roids': 'inv_stone_15',
                'alchemy': 'trade_alchemy', 'blacksmithing': 'trade_blacksmithing', 'engineering': 'trade_engineering', 'leatherworking': 'inv_misc_armorkit_17',
                'tailoring': 'trade_tailoring', 'enchanting': 'trade_engraving', 'hunters mark': 'ability_hunter_snipershot',
                'totem weapon buff windfury': 'spell_nature_cyclone', 'dragon breath chili': 'inv_misc_food_49'}


def setup_icon(k, v, cat):
    if k in SETUP_ICON:
        return SETUP_ICON[k]
    for cand in ((v if isinstance(v, str) else ''), k):
        w = re.sub(r'(?<=[a-z0-9])(?=[A-Z])', ' ', cand).lower()
        for pre in ('food ', 'sapper ', 'conjured rogue ', 'tristate effect ', 'totem weapon buff '):
            if w.startswith(pre) and pre not in ('conjured rogue ',):
                w = w[len(pre):]
        w = w.replace('totemweaponbuff', 'totem weapon buff ')
        if w in CONSUME_ICON:
            return CONSUME_ICON[w]
        w2 = re.sub(r'(?<=[a-z0-9])(?=[A-Z])', ' ', cand).lower()
        if w2 in CONSUME_ICON:
            return CONSUME_ICON[w2]
    cands = []
    if isinstance(v, str):
        w = re.sub(r'(?<=[a-z0-9])(?=[A-Z])', ' ', v)
        for pre in ('Food ', 'Sapper ', 'Conjured ', 'Tristate Effect '):
            if w.startswith(pre):
                w = w[len(pre):]
        cands += [w, w.replace(' Of ', ' of ').replace(' The ', ' the ')]
    cands.append(re.sub(r'(?<=[a-z0-9])(?=[A-Z])', ' ', k))
    for c in cands:
        c = c.lower()
        if c in item_by_name and item_by_name[c]:
            return item_by_name[c]
        if c in spell_by_name and spell_by_name[c]:
            return spell_by_name[c]
    return None


# the sim UI's consumes and buffs sections; consumables in the order flask, elixirs, food, potion, weapon, explosives, runes
SETUP_GROUP = {'Consumes': 'Consumables', 'Self buffs': 'Self buffs', 'Raid buffs': 'Raid buffs', 'Party buffs': 'Party buffs', 'Debuffs': 'Boss debuffs'}
CONSUME_SLOT = {'flask': ('Flask', 1), 'agilityElixir': ('Elixir', 2), 'strengthBuff': ('Elixir', 2), 'intellectElixir': ('Elixir', 2),
                'spiritElixir': ('Elixir', 2), 'attackPowerBuff': ('Elixir', 2), 'firePowerBuff': ('Elixir', 2), 'frostPowerBuff': ('Elixir', 2),
                'naturePowerBuff': ('Elixir', 2), 'shadowPowerBuff': ('Elixir', 2), 'spellPowerBuff': ('Elixir', 2), 'manaRegenElixir': ('Elixir', 2),
                'zanzaBuff': ('Elixir', 2), 'food': ('Food', 3), 'dragonBreathChili': ('Food', 3), 'defaultPotion': ('Potion', 4),
                'mainHandImbue': ('Weapon', 5), 'offHandImbue': ('Weapon', 5), 'staffImbue': ('Weapon', 5), 'sapperExplosive': ('Explosive', 6), 'fillerExplosive': ('Explosive', 6),
                'defaultConjured': ('Rune', 7)}


def tree_points(talents):
    parts = talents.split('-')
    return '/'.join(str(sum(int(c) for c in p)) for p in parts) if talents else ''


outs = [json.load(open(p)) for p in sorted(glob.glob(os.path.join(D, 'out', '*.json')))]
for _o in outs:  # the request's meta file is the source of truth for text (it can change without a re-run)
    _mp = os.path.join(D, 'req', _o['id'] + '.meta.json')
    if os.path.exists(_mp):
        _o['meta'] = json.load(open(_mp))
outs.sort(key=lambda o: (CLASS_ORDER.index(o['meta']['class']), -o['main']['dps']))

builds, races, lengths, abilities, auras, hist, stats, gear, setup, notes = [], [], [], [], [], [], [], [], [], []
# the previous round's main numbers (out_pre1b), for the before -> after note
PREV = {}
for _p in glob.glob(os.path.join(D, 'out_pre1b', '*.json')):
    try:
        _x = json.load(open(_p))
        PREV[_x['id']] = _x['main']['dps']
    except Exception:
        pass
# The gear rules of the bundled gear round (orchestration/standards.md 1b, 5a, 8a), shown on the Method tab.
GEAR_RULES = [
    "Every synthetic piece is priced at Forever's own budget for its archetype (standards 1b): its weighted linear stat total is the slot modifier x (1.125 x item level + 2.0) x the archetype's factor, within 1%, and its shape (a 1.7-norm) is within 6% of the standard.",
    "Weights: spell power 0.86, attack power 0.5 (a hunter's ranged attack power 0.4), every other stat (Strength, Agility, Stamina, Intellect, Spirit, hit and crit rating) 1. Caster weapons use Forever's caster weapon rate.",
    "Item levels follow the shared table: 66 for head, shoulders, chest, hands, legs and feet; 71 for neck, back, wrist, waist, rings, one-handers and ranged; 76 for two-handers.",
    "Stamina is each spec's own share in its Season of Discovery level 60 best-in-slot sets. Hit sits at the cap, and the other stats keep the spec's target proportions in the room left (standards 5a).",
    "Every spec wears two synthetic item level 71 trinkets: a stat trinket (about 46 weighted points) and an on-use trinket, 140 spell power, 335 attack power, 415 ranged attack power or 165 Strength for 20 s every 2 min, used in the opening seconds and again on cooldown. Forever's launch trinkets aren't simulated.",
    "Hybrids keep hybrid stats: Retribution and Enhancement carry spell power and Intellect alongside their melee stats, at the hybrid budget factor.",
    "Only the tier set's 2-piece bonus is live, and enchants are the best Phase 1 ones from outside the raids. Talents, rotations and races did not change in this round, apart from a few small rotation retunes and Destruction Fire's race (now Undead), noted on each preset.",
]
for _i, _t in enumerate(GEAR_RULES):
    notes.append({'build': '_round', 'kind': 'gear', 'order': _i, 'text': _t})
names = {}
timelines = defaultdict(list)


def icon_of(key):
    m = re.match(r'^([si])(\d+)', key)
    if not m:
        return 'inv_sword_04' if key.startswith('oOtherActionAttack') else 'ability_marksmanship' if key.startswith('oOtherActionShoot') else None
    n = int(m.group(2))
    if m.group(1) == 'i':
        ic = (item_names.get(n) or {}).get('icon') or item_csv_icon.get(n)
    else:
        ic = (spell_icons.get(n) or {}).get('icon') or wowhead_icon.get(n)
    if ic:
        return ic
    # by name: the Forever wiki, the sim's spell list, Wowhead's items and spells
    nm0 = re.sub(r'\s*\((main hand|off hand|\d CP|#\d+|\d+)\)\s*$', '', name_of(key)[0]).strip()
    for cand in (nm0, nm0.replace('Lights ', "Light's "), re.sub(r' Proc$', '', nm0)):
        c = cand.lower()
        for src in (wiki_icon, spell_by_name, item_by_name, wowhead_icon_by_name):
            if src.get(c):
                return src[c]
    tm = re.match(r'^(s\d+)t\d+$', key)  # a tagged copy (a second hawk, an off-hand hit) shares its spell's icon
    if tm:
        return icon_of(tm.group(1))
    return MANUAL_ICON.get(key)


def nm(key):
    if key not in names:
        n, r = name_of(key)
        names[key] = {'key': key, 'name': n, 'rank': r, 'icon': icon_of(key)}
    return names[key]['name']


for o in outs:
    m, mn, bid = o['meta'], o['main'], o['id']
    rs = sorted(o['races'], key=lambda r: -r['dps'])
    best = rs[0]
    horde = max((r for r in rs if r['faction'] == 'Horde'), key=lambda r: r['dps'], default=None)
    alli = max((r for r in rs if r['faction'] == 'Alliance'), key=lambda r: r['dps'], default=None)
    pub = m.get('published') or {}
    _prev = PREV.get(bid)
    encn = m.get('encounter')  # an encounter preset: the same boss as a creature type (shown only with "Special encounters" on)
    base_id = m.get('baseBuild')
    duty = m.get('duty')  # a duty preset (standards 23): it keeps this raid buff or debuff up itself (shown only with "Duty" on)
    base_name = DISPLAY.get(base_id, json.load(open(os.path.join(D, 'req', base_id + '.meta.json')))['build']) if base_id else None
    dname = f"{base_name} · vs {encn}" if encn else f"{base_name} · {duty} duty" if duty else DISPLAY.get(bid, m['build'])
    builds.append({
        'id': bid, 'class': m['class'], 'presetName': m['build'], 'spec': 'Seal Fate' if bid.startswith('rogue__seal-fate') else m['spec'], 'build': dname, 'isDefault': bool(m.get('isDefault')),
        'label': f"{m['class']} · {dname}", 'race': race_name(m['race']),
        'encounter': encn or None, 'baseBuild': base_id or None, 'duty': duty or None,
        'bossType': encn or 'Generic boss', 'bossText': f"Level 63 {encn} boss" if encn else 'Level 63 boss with no creature type (generic)',
        'dps': round(mn['dps'], 1), 'se': round(mn['se'], 2), 'stdev': round(mn['stdev'], 1),
        'p1': round(mn['pct']['1'], 1), 'p10': round(mn['pct']['10'], 1), 'p25': round(mn['pct']['25'], 1), 'p50': round(mn['pct']['50'], 1),
        'p75': round(mn['pct']['75'], 1), 'p90': round(mn['pct']['90'], 1), 'p99': round(mn['pct']['99'], 1),
        'min': round(mn['min'], 1), 'max': round(mn['max'], 1), 'cv': round(100 * mn['stdev'] / mn['dps'], 2),
        'petDps': round(mn['petDps'], 1), 'petShare': round(100 * mn['petDps'] / mn['dps'], 1) if mn['dps'] else 0,
        'oomSecs': round(mn['oomSecs'], 1), 'avgDur': round(mn['avgDur'], 1), 'tps': round(mn['tps'], 1), 'iters': mn['iters'],
        'talents': m.get('talents', ''), 'treePoints': m.get('treePoints') or tree_points(m.get('talents', '')),
        'trees': ' / '.join(m.get('treeNames') or []), 'gearSet': m.get('gearSet', ''), 'apl': m.get('apl', ''),
        'summary': m.get('summary', ''),
        # Why an encounter preset beats its normal build on its creature type of boss, for the Advantaged badge.
        'advantage': m.get('advantage') or None,
        # The abilities and talents the advantage line names, as game tooltip refs (k:<spell key>, tal:<class>:<tree>:<idx>).
        'advantageRefs': m.get('advantageRefs') or None,
        'advantageShort': m.get('advantageShort') or m.get('advantage') or None,  # the badge tooltip's one line
        'bestRace': race_name(best['race']), 'bestRaceDps': round(best['dps'], 1),
        'hordeRace': race_name(horde['race']) if horde else None, 'hordeDps': round(horde['dps'], 1) if horde else None,
        'allianceRace': race_name(alli['race']) if alli else None, 'allianceDps': round(alli['dps'], 1) if alli else None,
        'dps120': round(next(x['dps'] for x in o['lengths'] if x['secs'] == 120), 1),
        'dps300': round(next(x['dps'] for x in o['lengths'] if x['secs'] == 300), 1),
        'fightSecs': o['setup']['encounter'].get('duration'), 'fightVar': o['setup']['encounter'].get('durationVariation'),
        'targetLevel': o['setup']['encounter']['target'].get('level'), 'targetArmor': 3731,
        'raceIters': 10000, 'lengthIters': 5000, 'sampleSeeds': 30,
        'classIcon': 'classicon_' + m['class'].lower(), 'raceIcon': RACE_ICON.get(race_name(m['race'])),
        'bestRaceIcon': RACE_ICON.get(race_name(best['race'])), 'hordeRaceIcon': RACE_ICON.get(race_name(horde['race'])) if horde else None,
        'allianceRaceIcon': RACE_ICON.get(race_name(alli['race'])) if alli else None,
        'specIcon': spec_icon(m['class'], m.get('talents', '')),
        'faction': 'Horde' if m['race'] in ('Orc', 'Undead', 'Tauren', 'Troll', 'SkyborneWindshaper') else 'Alliance',
    })
    for x in o['lengths']:
        lengths.append({'build': bid, 'secs': x['secs'], 'dps': round(x['dps'], 1), 'se': round(x['se'], 2), 'oomSecs': round(x['oomSecs'], 1)})
    if duty:
        # The page shows a duty preset in the Ranking only, so it keeps its numbers and leaves the detail to its base.
        continue
    for r in o['races']:
        races.append({'build': bid, 'class': m['class'], 'race': race_name(r['race']), 'faction': r['faction'],
                      'dps': round(r['dps'], 1), 'se': round(r['se'], 2), 'vsBest': round(100 * (r['dps'] / best['dps'] - 1), 2),
                      'isPreset': r['race'] == m['race'], 'raceIcon': RACE_ICON.get(race_name(r['race']))})
    total = sum(a['dps'] for a in mn['actions'])
    seen = defaultdict(int)
    for a in mn['actions']:
        seen[nm(a['key'])] += 1
    for a in sorted(mn['actions'], key=lambda a: -a['dps']):
        if a['key'].startswith('oOtherAction') and a['dps'] == 0 and a['key'] not in ('oOtherActionAttack', 'oOtherActionShoot'):
            continue
        name = nm(a['key'])
        rank = names[a['key']]['rank']
        label = name + (f' r{rank}' if rank and seen[name] > 1 else '') + ('' if a['owner'] == 'player' else f" ({a['owner']})")

        abilities.append({'build': bid, 'key': a['key'], 'ability': label, 'owner': 'Player' if a['owner'] == 'player' else a['owner'],
                          'dps': round(a['dps'], 1), 'share': round(100 * a['dps'] / total, 2) if total else 0,
                          'casts': round(a['casts'], 1), 'hits': round(a['hits'], 1), 'crits': round(a['crits'], 1), 'ticks': round(a['ticks'], 1),
                          'critPct': round(a['critPct'], 1), 'misses': round(a['misses'], 1), 'dodges': round(a['dodges'], 1), 'parries': round(a['parries'], 1),
                          'glances': round(a['glances'], 1), 'avgHit': round(a['avgHit']), 'avgCrit': round(a['avgCrit']), 'avgTick': round(a['avgTick']),
                          'damage': round(a['damage'])})
    mine = [r for r in abilities if r['build'] == bid]
    cnt = defaultdict(int)
    for r in mine:
        cnt[r['ability']] += 1
    for r in mine:
        tm = re.search(r't(\d+)$', r['key'])
        if cnt[r['ability']] > 1 and tm:
            r['ability'] += f" #{tm.group(1)}"
    seen2 = defaultdict(int)
    for r in mine:
        seen2[r['ability']] += 1
        if seen2[r['ability']] > 1:
            r['ability'] += f" ({seen2[r['ability']]})"
    for a in mn['auras']:
        if a['uptime'] <= 0.05:
            continue
        nm(a['key'])
        auras.append({'build': bid, 'key': a['key'], 'aura': nm(a['key']), 'icon': names[a['key']].get('icon'), 'owner': 'Self' if a['owner'] == 'player' else 'Target',
                      'uptime': round(min(100, a['uptime']), 1), 'procs': round(a['procs'], 2)})
    for lo, cnt in mn['hist']:
        hist.append({'build': bid, 'lo': lo, 'count': cnt})
    fin = (o.get('stats') or {}).get('final', {})
    for st in ['Strength', 'Agility', 'Stamina', 'Intellect', 'Spirit', 'AttackPower', 'RangedAttackPower', 'MeleeHit', 'MeleeCrit', 'MeleeHaste',
               'SpellPower', 'SpellDamage', 'FirePower', 'FrostPower', 'ShadowPower', 'ArcanePower', 'NaturePower', 'HolyPower', 'SpellHit', 'SpellCrit', 'SpellHaste',
               'MP5', 'Mana', 'Health', 'Armor', 'ArmorPenetration', 'Expertise']:
        if fin.get(st):
            stats.append({'build': bid, 'stat': st, 'label': STAT_LABEL.get(st, st), 'value': round(fin[st], 2)})
    for i, it in enumerate(o['setup']['equipment'][:len(SLOTS)]):
        if not it or not it.get('id'):
            continue
        info = item_names.get(it['id'], {})
        gear.append({'build': bid, 'slot': SLOTS[i], 'item': info.get('name', 'Item %d' % it['id']), 'ilvl': info.get('ilvl'),
                     'enchant': enchant_names.get(it.get('enchant'), '') if it.get('enchant') else '', 'itemId': it['id'],
                     'icon': info.get('icon'), 'quality': QUALITY.get(info.get('quality'), 'common')})

    def words(k):
        s = re.sub(r'(?<=[a-z0-9])(?=[A-Z])', ' ', k[0].upper() + k[1:])
        return s

    # a rogue's poisons are class imbues in its options (beside a stone in the consumes' imbue slot)
    _ro = (json.load(open(os.path.join(D, 'req', bid + '.json')))['raid']['parties'][0]['players'][0].get('rogue') or {}).get('options') or {}
    _poisons = [('mainHandImbue', _ro['mainHandPoison']) for _ in [0] if _ro.get('mainHandPoison')] + [('offHandImbue', _ro['offHandPoison']) for _ in [0] if _ro.get('offHandPoison')]
    # a mage's staff imbue scroll is a class option too, on top of the oil
    _mo = (json.load(open(os.path.join(D, 'req', bid + '.json')))['raid']['parties'][0]['players'][0].get('mage') or {}).get('options') or {}
    _poisons += [('staffImbue', 'ScrollOf' + _mo['staffImbue'])] if _mo.get('staffImbue') else []
    for cat, obj in (('Consumes', o['setup']['consumes']), ('Consumes', dict(_poisons)), ('Self buffs', o['setup']['playerBuffs']), ('Raid buffs', o['setup']['raidBuffs']),
                     ('Party buffs', o['setup']['partyBuffs']), ('Debuffs', o['setup']['debuffs'])):
        for k, v in (obj or {}).items():
            if v in (False, 0, '', None) or v == 'TristateEffectMissing':
                continue
            if cat == 'Consumes':
                txt = words(k) + ': ' + (words(str(v)) if isinstance(v, str) else str(v))
            else:
                txt = words(k) + (' (improved)' if v == 'TristateEffectImproved' else '' if v in (True, 'TristateEffectRegular') else f' ({v})')
            grp, slot, so = SETUP_GROUP[cat], None, 0
            if cat == 'Consumes':
                slot, so = CONSUME_SLOT.get(k, ('Other', 90))
            setup.append({'build': bid, 'category': cat, 'item': txt, 'icon': setup_icon(k, v, cat), 'group': grp, 'slot': slot,
                          'slotOrder': so, 'hand': 'MH' if k == 'mainHandImbue' else 'OH' if k == 'offHandImbue' else None,
                          'level': 'Improved' if v == 'TristateEffectImproved' else (str(v) if isinstance(v, int) and not isinstance(v, bool) and v > 1 else None)})
    for p in o['setup']['professions']:
        if p:
            setup.append({'build': bid, 'category': 'Professions', 'item': words(p), 'icon': setup_icon(p, p, 'Professions'), 'group': 'Professions',
                          'slot': None, 'slotOrder': 0, 'hand': None, 'level': None})
    for kind in ('keyTalents', 'rotation', 'consumes', 'caveats'):
        for i, t in enumerate(m.get(kind) or []):
            notes.append({'build': bid, 'kind': kind, 'order': i, 'text': t})
    for s in o['samples']:
        for c in s['casts']:
            nm(c[1])
        for d in s['dmg']:
            nm(d[1])
        for a in s['auras']:
            nm(a[0])
        for c in s['cds']:
            nm(c[1])
        timelines[m['class'].lower()].append({'build': bid, 'label': s['label'], 'pct': s['pct'], 'seed': s['seed'], 'dps': round(s['dps'], 1),
                                              'dur': s['dur'], 'units': s['units'], 'casts': s['casts'], 'dmg': s['dmg'], 'auras': s['auras'],
                                              'res': s['res'], 'cds': s['cds']})

# neutral wording: the page never speaks of what was "shipped"
_SHIP = [(r"\bShipped as (a|an|the|its) ", lambda m: m.group(1)[0].upper() + m.group(1)[1:] + ' '), (r"\bShips as its own playstyle", "Its own playstyle"),
         (r"\bit ships as the pet alternative", "it is the pet alternative"), (r"\bIt ships as ", "It is "),
         (r"\bso it ships as its own preset", "so it is its own preset"), (r"\bso Major Mana Potion ships\b", "so the preset uses Major Mana Potion")]


def unship(t):
    for a, b in _SHIP:
        t = re.sub(a, b, t)
    return t


for b in builds:
    b['summary'] = unship(b['summary'])

# What each duty costs: its base's DPS minus its own, at 180 s.
_by_id = {b['id']: b for b in builds}
for b in builds:
    if b['duty']:
        _base = _by_id[b['baseBuild']]
        b['dutyCost'] = round(_base['dps'] - b['dps'], 1)
        b['dutyCostPct'] = round(100 * (1 - b['dps'] / _base['dps']), 2)

# rank among the default presets (an encounter or duty preset gets the place it would take among them), and among all
_defaults = [b for b in builds if not b['encounter'] and not b['duty']]
for b in builds:
    b['rank'] = 1 + sum(1 for x in _defaults if x['dps'] > b['dps'] and x is not b)
    b['rankAll'] = 1 + sum(1 for x in builds if x['dps'] > b['dps'])


def dump(name, rows):
    p = os.path.join(OUT, name + '.json')
    json.dump(rows, open(p, 'w'), separators=(',', ':'))
    print(f'{name}: {len(rows)} rows, {os.path.getsize(p) / 1024:.0f} KB')


dump('builds', builds)
dump('races', races)
dump('lengths', lengths)
dump('abilities', abilities)
dump('auras', auras)
dump('hist', hist)
dump('stats', stats)
dump('gear', gear)
# which entries do anything in the sim, and why not (setup_active.py, from abl/ablate.json by ablate.py)
if os.path.exists(os.path.join(D, 'abl', 'ablate.json')):
    sys.path.insert(0, D)
    import setup_active
    setup_active.annotate(setup, builds, stats)
dump('setup', setup)
# the dashboard is a snapshot of the final data: drop notes that describe earlier rounds or the
# project's process, and strip process wording from the rest
_HIST = re.compile(r"earlier gear|earlier set|first set|before the 5a and 1b rounds|gear before the|The audit found", re.I)
_SUBS = [(r"\s*\(standards \d+[a-z]?\)", ""), (r"Late Forever rulings cost", "Forever launch rules cost"),
         (r"one 2% crit elixir per character by ruling", "only one 2% crit elixir per character"),
         (r"by the one-crit-elixir ruling", "by the one-crit-elixir rule"),
         (r"\s*Talents, rotations and races did not change in this round[^.]*\.", "")]


def _clean_note(t):
    for a, b in _SUBS:
        t = re.sub(a, b, t)
    return unship(t).strip()


notes = [dict(n, text=_clean_note(n['text'])) for n in notes if not _HIST.search(n['text'])]
dump('notes', notes)

# race scenarios (build_races.py, from run_races.py's sims): race_variants, race_breakdown, race_summary, race_notes
sys.path.insert(0, D)
import build_races
build_races.build(builds, races, stats, dump)

# stat weights (weights.py)
wrows = []
ABBR = {'AttackPower': 'AP', 'RangedAttackPower': 'RAP', 'SpellPower': 'SP'}
for w in sorted(glob.glob(os.path.join(D, 'weights', '*.json'))):
    w = json.load(open(w))
    for r in w['rows']:
        pct = r['stat'] in ('MeleeHit', 'MeleeHitDown', 'MeleeCrit')
        wrows.append({'build': w['id'], 'stat': r['stat'], 'label': r['label'], 'primary': w['primaryLabel'], 'primaryAbbr': ABBR[w['primary']],
                      'unit': '1%' if pct else 'point', 'dpsPer': round(r['dpsPer'], 3), 'sePer': round(r['sePer'], 3),
                      'worth': round(r['vsPrimary'], 2) if r['vsPrimary'] is not None else None,
                      'ratingWorth': round(r['perRatingVsPrimary'], 2) if r.get('perRatingVsPrimary') is not None else None,
                      'iters': w['iters']})
if wrows:
    dump('weights', wrows)

# mana over a 480 s fight (mana_curves.py)
import mana_curves
mana_rows, mana_summary = mana_curves.datasets()
if mana_summary:
    dump('mana', mana_rows)
    dump('mana_summary', mana_summary)

# assumption scenarios (run_assumptions.py): assumptions, assumption_dps
if os.path.exists(os.path.join(D, 'asm', 'results.jsonl')):
    import asm_data
    asm_data.build(dump, builds)

# rotation: each preset's real APL as plain-English priority rows (rotation.py)
import rotation
RACIAL_RACE = {}
for _race, _lst in json.load(open(os.path.join(TREE, 'ui/core/forever/racials.json'))).items():
    for _r in _lst:
        for _v in _r.get('variants', []):
            RACIAL_RACE.setdefault(_v['spellId'], race_name(_race.replace('Race', '')))
        RACIAL_RACE.setdefault(_r['name'], race_name(_race.replace('Race', '')))
for _k, _v in list(RACIAL_RACE.items()):
    RACIAL_RACE[_k] = {'SkyborneHighOrder': 'High Order', 'SkyborneWindshaper': 'Windshaper'}.get(_v, _v)
_outs_r = [dict(o, meta=dict(o['meta'], race=race_name(o['meta']['race']))) for o in outs if not o['meta'].get('duty')]
rot_rows, rot_fallback = rotation.build(os.path.join(D, 'req'), _outs_r, nm, names, abilities, RACIAL_RACE)
for r in rot_rows:
    if r['key'] and not r['icon']:
        r['icon'] = names.get(r['key'], {}).get('icon')
    if r['key'] == 'oOtherActionPotion':  # the APL's "potion" is the preset's default potion
        pot = next((x for x in setup if x['build'] == r['build'] and x.get('slot') == 'Potion'), None)
        if pot:
            r['name'] = pot['item'].split(': ', 1)[-1]
            r['icon'] = pot['icon']
            _norm = lambda t: re.sub(r'[^a-z]', '', t.lower())
            pk = next((k for k, n in names.items() if _norm(n['name']) == _norm(r['name'])), None)
            ab = next((a for a in abilities if a['build'] == r['build'] and a['key'] == pk), None)
            if ab:
                r['casts'] = ab['casts']
    r['notes'] = _clean_note(r['notes'])
dump('rotation', rot_rows)
print('rotation conditions shown in compact form:', rot_fallback)

# tooltips: game-style text for spells, items, enchants, setup chips, races and stats (tooltips.py)
sys.path.insert(0, D)
import tooltips
import whcache
tt_rows, tt_missing, tt_report = tooltips.build(TREE, names, gear, setup, [o for o in outs if not o['meta'].get('duty')], abilities, auras, timelines, builds, item_names,
                                                {e['effectId']: e for e in db['enchants']}, TREE_JSON, os.path.join(TREE, 'ui/core/forever/racials.json'), SLOTS)
dump('tooltips', tt_rows)
print('tooltips without text:', tt_missing)

# talent trees, laid out as the sim UI draws them
trees = []
for cls in CLASS_ORDER:
    t = json.load(open(os.path.join(TREE, f'ui/core/talents/trees/{cls.lower()}.json')))
    for ti, tr in enumerate(t):
        for i, tal in enumerate(tr['talents']):
            pre = tal.get('prereqLocation')
            trees.append({'class': cls, 'tree': ti, 'treeName': tr['name'], 'idx': i, 'name': tal['name'], 'row': tal['location']['rowIdx'],
                          'col': tal['location']['colIdx'], 'maxPoints': tal['maxPoints'], 'preRow': pre['rowIdx'] if pre else None,
                          'preCol': pre['colIdx'] if pre else None, 'icon': tal.get('icon') or 'inv_misc_questionmark',
                          'description': tal.get('description', ''), 'ranks': tal.get('ranks') or [],
                          # wowhead link per rank: "<domain>:<spell id>", Forever's where its database knows the id, else Classic's, else ""
                          'wh': [next((f'{dm}:{sid}' for dm in ('forever', 'classic') if whcache.ok(whcache.cached(dm, 'spell', sid))), '')
                                 for sid in tal.get('spellIds') or []]})
dump('talenttrees', trees)

# wowhead coverage of the tooltips: which source each one's text came from, and the ids behind them
import collections
cov = {'rows': dict(collections.Counter(r['kind'] + ':' + (r['src'] or 'none') for r in tt_rows if r['kind'] in ('spell', 'item', 'setup', 'enchant')))}
ids = {}
for r in tt_rows:
    if r['kind'] not in ('spell', 'item', 'setup', 'enchant'):
        continue
    m = re.match(r'^[si](\d+)', r['key'])
    ident = r['wh'] or (('spell=' if r['key'][0] == 's' else 'item=') + m.group(1) if m and r['key'][1:].split('t')[0].isdigit() and int(m.group(1)) < 990000 else None)
    if not ident:
        continue
    prev = ids.get(ident)
    order = {'forever': 0, 'classic': 2, None: 3}
    src = r['src'] if r['src'] in ('forever', 'classic') else ('other' if r['src'] else None)
    if prev is None or order.get(src, 1) < order.get(prev, 1):
        ids[ident] = src
cov['ids'] = dict(collections.Counter(v or 'none' for v in ids.values()))
cov['idsBySource'] = {k: sorted(i for i, v in ids.items() if (v or 'none') == k) for k in set(v or 'none' for v in ids.values())}
cov['talentLinks'] = dict(collections.Counter((w.split(':')[0] if w else 'none') for t in trees for w in t['wh']))
cov.update(tt_report)
json.dump(cov, open(os.path.join(D, 'fortt_coverage.json'), 'w'), indent=1)
print('wowhead coverage:', cov['rows'], cov['ids'], cov['talentLinks'])

# talent tree art: each tree's icon and background
art = []
for cls in CLASS_ORDER:
    for ti, tr in enumerate(TREE_JSON[cls]):
        art.append({'class': cls, 'tree': ti, 'treeName': tr['name'], 'icon': CLIENT_TREES[cls][ti]['icon'], 'bgUrl': tr.get('backgroundUrl')})

# icons as data URLs (36 px, from Wowhead's icon CDN, cached locally)
import base64, urllib.request
from concurrent.futures import ThreadPoolExecutor
os.makedirs(os.path.join(D, 'icons'), exist_ok=True)
extra = set(RACE_ICON.values()) | set(FACTION_ICON.values()) | {'classicon_' + c.lower() for c in CLASS_ORDER} | {a['icon'] for a in art}
extra |= set(build_races.PART_ICON.values())
extra |= {g['icon'] for g in gear if g.get('icon')} | {x['icon'] for x in setup if x.get('icon')} | {x['icon'] for x in auras if x.get('icon')} | {b['specIcon'] for b in builds}
want = sorted({t['icon'] for t in trees} | {n['icon'] for n in names.values() if n.get('icon')} | {'inv_misc_questionmark'} | extra)
def fetch(ic):
    path = os.path.join(D, 'icons', ic + '.jpg')
    if not os.path.exists(path):
        try:
            with urllib.request.urlopen(f'https://wow.zamimg.com/images/wow/icons/medium/{ic}.jpg', timeout=20) as r:
                open(path, 'wb').write(r.read())
        except Exception as e:
            return ic, None
    return ic, base64.b64encode(open(path, 'rb').read()).decode()
with ThreadPoolExecutor(16) as ex:
    got = list(ex.map(fetch, want))
# large (56 px) versions for headers
def fetch_large(ic):
    path = os.path.join(D, 'icons', ic + '.large.jpg')
    if not os.path.exists(path):
        try:
            with urllib.request.urlopen(f'https://wow.zamimg.com/images/wow/icons/large/{ic}.jpg', timeout=20) as r:
                open(path, 'wb').write(r.read())
        except Exception:
            return ic, None
    return ic, base64.b64encode(open(path, 'rb').read()).decode()
large_want = sorted({b['specIcon'] for b in builds} | {'classicon_' + c.lower() for c in CLASS_ORDER} | set(RACE_ICON.values()))
with ThreadPoolExecutor(16) as ex:
    got_l = list(ex.map(fetch_large, large_want))
# empty gear slot placeholders, from the sim UI
SLOT_FILE = {'Head': 'head', 'Neck': 'neck', 'Shoulders': 'shoulders', 'Back': 'shirt', 'Chest': 'chest', 'Wrist': 'wrists', 'Hands': 'hands', 'Waist': 'waist',
             'Legs': 'legs', 'Feet': 'feet', 'Ring 1': 'finger', 'Ring 2': 'finger', 'Trinket 1': 'trinket', 'Trinket 2': 'trinket', 'Main hand': 'mainhand',
             'Off hand': 'offhand', 'Ranged': 'ranged'}
slot_icons = [{'icon': 'slot_' + f, 'url': 'data:image/jpeg;base64,' + base64.b64encode(open(os.path.join(TREE, 'assets/item_slots', f + '.jpg'), 'rb').read()).decode()}
              for f in sorted(set(SLOT_FILE.values()))]
# The Ranking tab, which the page opens on, needs only the class, race, faction and preset icons, so those load first.
# The spell, item and gear-slot icons the other tabs show are a separate dataset that loads later.
core_icons = set(RACE_ICON.values()) | set(FACTION_ICON.values()) | {'classicon_' + c.lower() for c in CLASS_ORDER} | {b['specIcon'] for b in builds} | {'inv_misc_questionmark'}
all_icons = [{'icon': ic, 'url': 'data:image/jpeg;base64,' + b} for ic, b in got if b] + [{'icon': ic + ':l', 'url': 'data:image/jpeg;base64,' + b} for ic, b in got_l if b] + slot_icons
dump('icons', [r for r in all_icons if r['icon'] in core_icons or r['icon'].endswith(':l')])
dump('icons_more', [r for r in all_icons if not (r['icon'] in core_icons or r['icon'].endswith(':l'))])

# each spell's and item's colour, from its icon: the dominant saturated hue, at a lightness that reads on light and dark
from PIL import Image
import colorsys
_have = {ic for ic, b in got if b}


def icon_color(ic):
    try:
        im = Image.open(os.path.join(D, 'icons', ic + '.jpg')).convert('RGB').resize((32, 32))
    except Exception:
        return None
    px = [im.getpixel((x, y)) for x in range(5, 27) for y in range(5, 27)]
    bins = [[0.0, 0.0, 0.0, 0.0] for _ in range(24)]
    tot = [0.0, 0.0, 0.0]
    for r, g, b in px:
        h, l, sat = colorsys.rgb_to_hls(r / 255, g / 255, b / 255)
        tot[0] += r; tot[1] += g; tot[2] += b
        if sat < 0.3 or l < 0.12 or l > 0.92:
            continue
        w = sat * (1 - abs(l - 0.5) * 1.6)
        i = int(h * 24) % 24
        bins[i][0] += w; bins[i][1] += r * w; bins[i][2] += g * w; bins[i][3] += b * w
    sw = [bins[i][0] + 0.5 * (bins[i - 1][0] + bins[(i + 1) % 24][0]) for i in range(24)]
    best = max(range(24), key=lambda i: sw[i])
    if sw[best] < 6:  # a grey or dark icon: a neutral grey with a hint of its tint
        r, g, b = (t / len(px) / 255 for t in tot)
        h, l, sat = colorsys.rgb_to_hls(r, g, b)
        r, g, b = colorsys.hls_to_rgb(h, 0.55, min(sat, 0.18))
    else:
        ws = sum(bins[j][0] for j in (best - 1, best, (best + 1) % 24))
        r, g, b = (sum(bins[j][k] for j in (best - 1, best, (best + 1) % 24)) / ws / 255 for k in (1, 2, 3))
        h, l, sat = colorsys.rgb_to_hls(r, g, b)
        # yellows and greens look darker at the same lightness: lift less; keep every colour mid-tone
        r, g, b = colorsys.hls_to_rgb(h, min(0.58, max(0.44, l)), min(0.85, max(0.45, sat)))
    return '#%02x%02x%02x' % tuple(round(c * 255) for c in (r, g, b))


for n in names.values():
    if n.get('icon') and n['icon'] not in _have:
        n['icon'] = None
    n['color'] = icon_color(n['icon']) if n.get('icon') else None
dump('names', sorted(names.values(), key=lambda x: x['key']))
print('names without an icon:', sorted(n['name'] for n in names.values() if not n.get('icon')))
# tree backgrounds as data URLs
def fetch_bg(a):
    if not a['bgUrl']:
        return a, None
    path = os.path.join(D, 'icons', f"bg_{a['class']}_{a['tree']}.jpg")
    if not os.path.exists(path):
        try:
            with urllib.request.urlopen(a['bgUrl'], timeout=20) as r:
                open(path, 'wb').write(r.read())
        except Exception:
            return a, None
    return a, base64.b64encode(open(path, 'rb').read()).decode()
with ThreadPoolExecutor(9) as ex:
    got_bg = list(ex.map(fetch_bg, art))
dump('treeart', [{'class': a['class'], 'tree': a['tree'], 'treeName': a['treeName'], 'icon': a['icon'], 'bg': ('data:image/jpeg;base64,' + b) if b else None} for a, b in got_bg])
print('missing icons:', [ic for ic, b in got if not b])
for c, rows in timelines.items():
    dump('timeline_' + c, rows)
