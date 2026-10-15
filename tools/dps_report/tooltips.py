"""tooltips.py: game-style tooltip text for the dashboard (the `tooltips` dataset).

Rows: {key, kind, title, sub, icon, quality, lines, set, setSize, setBonuses}
  key   names keys (s<id>[t<tag>], i<id>, oOther...), gear items i<id>, ench:<name>, setup:<chip text>,
        race:<Race>, race:<Race>:<Class>, stat:<Stat>
  lines list of [style, text] or [style, left, right]:
        t plain, m muted, y gold (spell text), g green (Equip / Use / enchant), h gold heading with a muted right part,
        2 two columns, s spacer
All text is plain (HTML from the sources is stripped here); the page builds the tooltip with textContent.
"""
import glob, html, json, os, re
import whcache
from html.parser import HTMLParser

CLASS_ORDER = ['Warrior', 'Rogue', 'Hunter', 'Paladin', 'Shaman', 'Druid', 'Mage', 'Warlock', 'Priest']
RACE_NAMES = {'NightElf': 'Night Elf', 'SkyborneHighOrder': 'High Order', 'SkyborneWindshaper': 'Windshaper'}
HORDE = {'Orc', 'Undead', 'Tauren', 'Troll', 'Windshaper'}
QUALITY = {0: 'poor', 1: 'common', 2: 'uncommon', 3: 'rare', 4: 'epic', 5: 'legendary'}


def norm(s):
    return re.sub(r'[^a-z0-9]', '', str(s).lower())


# ---------- Wowhead tooltip HTML -> lines ----------
SKIP_CLASSES = ('whtt-sellprice', 'whtt-extra', 'wowhead-tooltip-requirements', 'whtt-name', 'wowhead-tooltip-item-classes', 'indent', 'whtt-ilvl')
COLOR = {'q': 'y', 'q0': 'm', 'q1': 't', 'q2': 'g', 'q3': 't', 'q4': 't', 'q5': 't'}
VOID = {'br', 'img', 'hr', 'input', 'meta', 'link'}


class _TT(HTMLParser):
    def __init__(self):
        super().__init__(convert_charrefs=True)
        self.lines, self.cur, self.stack = [], [], []
        self.tables = []  # [twocol, left, right, cell]

    def _skip(self):
        return any(s for _, s, _ in self.stack)

    def _color(self):
        for _, _, c in reversed(self.stack):
            if c:
                return c
        return None

    def flush(self):
        txt = re.sub(r'\s+', ' ', ''.join(t for t, _ in self.cur)).strip()
        col = next((c for t, c in self.cur if t.strip() and c), None)
        if txt:
            self.lines.append([col or 't', txt])
        self.cur = []

    def handle_starttag(self, tag, attrs):
        a = dict(attrs)
        cls = (a.get('class') or '').split()
        if tag == 'br':
            if self.tables and self.tables[-1][0] and self.tables[-1][3]:
                self.tables[-1][3].append((' ', None))
            else:
                self.flush()
            return
        if tag in VOID:
            return
        if tag in ('div', 'p', 'table', 'tr'):
            self.flush()
        if tag == 'table':
            self.tables.append([a.get('width') == '100%', [], [], None])
        if tag in ('td', 'th') and self.tables and self.tables[-1][0]:
            t = self.tables[-1]
            t[3] = t[1] if tag == 'td' else t[2]
        skip = any(c in SKIP_CLASSES for c in cls)
        col = next((COLOR[c] for c in cls if c in COLOR), None)
        self.stack.append((tag, skip, col))

    def handle_startendtag(self, tag, attrs):
        self.handle_starttag(tag, attrs)

    def handle_endtag(self, tag):
        if tag in VOID:
            return
        if tag == 'tr' and self.tables and self.tables[-1][0]:
            t = self.tables[-1]
            left = re.sub(r'\s+', ' ', ''.join(x for x, _ in t[1])).strip()
            right = re.sub(r'\s+', ' ', ''.join(x for x, _ in t[2])).strip()
            col = next((c for x, c in t[1] + t[2] if x.strip() and c), None)
            if left or right:
                self.lines.append(['2', left, right] if left and right else [col or 't', left or right])
            t[1], t[2], t[3] = [], [], None
        if tag in ('div', 'p', 'tr') and not (self.tables and self.tables[-1][0] and tag != 'tr'):
            self.flush()
        if tag == 'table' and self.tables:
            self.tables.pop()
            self.flush()
        for i in range(len(self.stack) - 1, -1, -1):
            if self.stack[i][0] == tag:
                del self.stack[i:]
                break

    def handle_data(self, data):
        if self._skip():
            return
        seg = (data, self._color())
        if self.tables and self.tables[-1][0] and self.tables[-1][3] is not None:
            self.tables[-1][3].append(seg)
        else:
            self.cur.append(seg)

    def close(self):
        super().close()
        self.flush()


def tidy_formula(s):
    """Forever wowhead leaves some numbers as formulas: "[48 + (151 * (1 * 1))]" -> "199", "(246 * 1)" -> "246",
    "[(430 + Spirit) * (1)]" -> "(430 + Spirit)"."""
    if '[' not in s and '*' not in s:
        return s
    prev = None
    while prev != s:
        prev = s
        s = re.sub(r'\(\s*(\d+(?:\.\d+)?)\s*\)', r'\1', s)
        s = re.sub(r'\s*\*\s*1(?![\d.])', '', s)

    def ev(m):
        body = m.group(1)
        if re.fullmatch(r'[\d\s+\-*/().]+', body):
            try:
                v = eval(compile(body, '<f>', 'eval'), {'__builtins__': {}}, {})
                return str(int(v)) if float(v).is_integer() else ('%.1f' % v).rstrip('0').rstrip('.')
            except Exception:
                pass
        return body
    return re.sub(r'\[([^\[\]]*)\]', ev, s)


def html_lines(h):
    p = _TT()
    p.feed(h or '')
    p.close()
    return [[l[0]] + [tidy_formula(x) for x in l[1:]] for l in p.lines]


DROP = re.compile(r'^(Item Effect|Rank \d+( Level \d+)?|Level \d+|Requires .*|Binds when .*|Unique.*|Durability .*|Item Level .*|Sell Price.*|Max Stack.*|'
                  r'Dropped by.*|Drop Chance.*|Classes: .*)$')


def clean_lines(lines, title):
    out = []
    for ln in lines:
        if ln[0] != '2':
            if DROP.match(ln[1]) or ln[1] == title:
                continue
        else:
            if re.match(r'^Rank \d+', ln[2]) or ln[1] == title:
                continue
        out.append(ln)
    return out


# ---------- sources ----------
def load_wowhead(tree):
    spells, items = {}, {}
    for line in open(os.path.join(tree, 'assets/db_inputs/wowhead_spell_tooltips.csv')):
        i, _, j = line.rstrip('\n').partition(',')
        try:
            spells[int(i)] = json.loads(j)
        except Exception:
            pass
    for line in open(os.path.join(tree, 'assets/db_inputs/wowhead_item_tooltips.csv')):
        i, _, j = line.rstrip('\n').partition(',')
        try:
            items[int(i)] = json.loads(j)
        except Exception:
            pass
    return spells, items


def spell_rank(j):
    m = re.search(r'Rank (\d+)', j.get('tooltip') or '')
    return int(m.group(1)) if m else 0


def meta_lines(meta):
    """Forever wiki meta (['380 Mana', '30 yd range', '3 sec cast', '30 sec cooldown']) as game rows."""
    meta = [m for m in meta or [] if not re.match(r'^Level \d+$', m)]
    cost = [m for m in meta if re.search(r'\b(Mana|Rage|Energy|Focus|health|Health)\b', m) and 'cooldown' not in m]
    rng = [m for m in meta if re.search(r'range|Range', m)]
    cast = [m for m in meta if re.search(r'cast|Instant|Channeled|Passive|Next melee', m) and m not in cost]
    cd = [m for m in meta if 'cooldown' in m]
    rest = [m for m in meta if m not in cost + rng + cast + cd]
    out = []
    for a, b in ((cost, rng), (cast, cd)):
        if a and b:
            out.append(['2', a[0], b[0]])
        elif a or b:
            out.append(['t', (a or b)[0]])
    out += [['t', m] for m in rest]
    return out


OTHER_TEXT = {
    'oOtherActionAttack': 'Automatic melee swings with your weapon.',
    'oOtherActionAttackt1': 'Automatic melee swings with your main hand weapon.',
    'oOtherActionAttackt2': 'Automatic melee swings with your off hand weapon.',
    'oOtherActionAttackt3': 'Extra melee swings granted by procs such as Windfury, Sword Specialization or Hand of Justice.',
    'oOtherActionShoot': 'Automatically shoots the target with your ranged weapon.',
    'oOtherActionManaGain': 'Mana restored by procs, talents and buffs.',
    'oOtherActionRageGain': 'Rage generated by damage dealt and talents.',
    'oOtherActionMove': 'Time spent moving.',
    'oOtherActionManaRegen': 'Mana regenerated over the fight.',
    'oOtherActionEnergyRegen': 'Energy regenerated over the fight.',
    'oOtherActionFocusRegen': 'Focus regenerated over the fight.',
    'oOtherActionPotion': 'A potion used during the fight.',
}

# Sim-internal labels to the in-game names the sources use.
ALIAS = {'Natures Grace Proc': "Nature's Grace", 'Flurry Proc (16280)': 'Flurry', 'Elemental Devastation Proc': 'Elemental Devastation',
         'Nightfall Shadow Trance': 'Shadow Trance', 'Lights Vigil': "Light's Vigil", 'Shadow Word Death': 'Shadow Word: Death',
         'Crusader Enchant MH': 'Holy Strength', 'Crusader Enchant OH': 'Holy Strength', 'Extra Attacks (Main Hand)': 'Extra Attacks',
         'Windfury Totem Passive': 'Windfury Totem', 'Flametongue Totem Proc': 'Flametongue Totem', 'Attack': 'Searing Totem',
         'Shadow and Flame (Shadow)': 'Shadow and Flame'}


def display_name(n):
    s = str(n)
    s = re.sub(r'\s*-\s*\(#\d+\)\S*$', '', s)
    s = re.sub(r'\s*\(#\d+\)\S*$', '', s)
    s = re.sub(r'\s*\(Forever\)', '', s)
    s = re.sub(r'\s+', ' ', s).strip()
    return ALIAS.get(s, s)


SET_FILES = 'sim/*/item*sets_pve.go'


def set_bonuses(tree):
    """Synthetic set name -> [[pieces, text], ...] from the bonus comments in the sim's Go source."""
    out = {}
    for p in glob.glob(os.path.join(tree, SET_FILES)):
        src = open(p).read()
        bonus_vars = {}
        for m in re.finditer(r'var (\w+) = map\[int32\]core\.ApplyEffect\{(.*?)\n\}', src, re.S):
            body = m.group(2)
            items = []
            for mm in re.finditer(r'((?:\n\t//[^\n]*)+)\n\t(\d+): func', body):
                txt = ' '.join(l.strip()[2:].strip() for l in mm.group(1).strip().split('\n'))
                txt = re.sub(r'\s*\((?:[^()]*\b(?:aura|auras|spell|Spell|melee only|ItemSet)\b[^()]*)\)', '', txt)
                first = re.match(r'(.+?[.!])(\s|$)', txt)
                txt = first.group(1) if first else txt
                txt = txt[0].upper() + txt[1:] if txt else txt
                if not txt.endswith('.'):
                    txt += '.'
                items.append([int(mm.group(2)), txt])
            bonus_vars[m.group(1)] = items
        for m in re.finditer(r'core\.NewItemSet\(core\.ItemSet\{\s*Name:\s*"([^"]+)",\s*Bonuses:\s*(\w+)', src):
            if m.group(2) in bonus_vars:
                out[m.group(1)] = bonus_vars[m.group(2)]
    return out


STAT = ['Strength', 'Agility', 'Stamina', 'Intellect', 'Spirit', 'SpellPower', 'ArcanePower', 'FirePower', 'FrostPower', 'HolyPower', 'NaturePower',
        'ShadowPower', 'MP5', 'SpellHit', 'SpellCrit', 'SpellHaste', 'SpellPenetration', 'AttackPower', 'MeleeHit', 'MeleeCrit', 'MeleeHaste',
        'ArmorPenetration', 'Expertise', 'Mana', 'Energy', 'Rage', 'Armor', 'RangedAttackPower', 'Defense', 'Block', 'BlockValue', 'Dodge', 'Parry',
        'Resilience', 'Health', 'ArcaneResistance', 'FireResistance', 'FrostResistance', 'NatureResistance', 'ShadowResistance', 'BonusArmor',
        'HealingPower', 'SpellDamage', 'FeralAttackPower']


def fnum(v):
    return str(int(v)) if float(v).is_integer() else ('%.1f' % v).rstrip('0').rstrip('.')


def stat_lines(stats, hit=0, crit=0):
    """Item / enchant stats as tooltip lines: white base stats, green Equip lines."""
    s = {STAT[i]: v for i, v in enumerate(stats or []) if v and i < len(STAT)}
    white, green = [], []
    if s.get('Armor'):
        white.append(['t', f"{fnum(s['Armor'])} Armor"])
    for k in ('Strength', 'Agility', 'Stamina', 'Intellect', 'Spirit'):
        if s.get(k):
            white.append(['t', f"+{fnum(s[k])} {k}"])
    for k in ('Arcane', 'Fire', 'Frost', 'Nature', 'Shadow'):
        if s.get(k + 'Resistance'):
            white.append(['t', f"+{fnum(s[k + 'Resistance'])} {k} Resistance"])
    hitp = (s.get('MeleeHit') or s.get('SpellHit') or 0) + hit / 10.0
    critp = (s.get('MeleeCrit') or s.get('SpellCrit') or 0) + crit / 14.0
    if hitp:
        green.append(f'Improves your chance to hit by {fnum(round(hitp, 2))}%.')
    if critp:
        green.append(f'Improves your chance to get a critical strike by {fnum(round(critp, 2))}%.')
    if s.get('AttackPower'):
        green.append(f"+{fnum(s['AttackPower'])} Attack Power.")
    if s.get('RangedAttackPower'):
        green.append(f"+{fnum(s['RangedAttackPower'])} ranged Attack Power.")
    if s.get('FeralAttackPower'):
        green.append(f"+{fnum(s['FeralAttackPower'])} Attack Power in Cat, Bear and Dire Bear forms only.")
    if s.get('SpellPower'):
        green.append(f"Increases damage and healing done by magical spells and effects by up to {fnum(s['SpellPower'])}.")
    if s.get('SpellDamage'):
        green.append(f"Increases damage done by magical spells and effects by up to {fnum(s['SpellDamage'])}.")
    for k in ('Arcane', 'Fire', 'Frost', 'Holy', 'Nature', 'Shadow'):
        if s.get(k + 'Power'):
            green.append(f"Increases damage done by {k} spells and effects by up to {fnum(s[k + 'Power'])}.")
    if s.get('HealingPower'):
        green.append(f"Increases healing done by spells and effects by up to {fnum(s['HealingPower'])}.")
    if s.get('MP5'):
        green.append(f"Restores {fnum(s['MP5'])} mana per 5 sec.")
    if s.get('SpellHaste') or s.get('MeleeHaste'):
        green.append(f"Increases your attack and casting speed by {fnum(s.get('SpellHaste') or s.get('MeleeHaste'))}%.")
    if s.get('SpellPenetration'):
        green.append(f"Decreases the magical resistances of your spell targets by {fnum(s['SpellPenetration'])}.")
    if s.get('ArmorPenetration'):
        green.append(f"Your attacks ignore {fnum(s['ArmorPenetration'])} of your opponent's armor.")
    if s.get('Defense'):
        green.append(f"Increased Defense +{fnum(s['Defense'])}.")
    if s.get('Dodge'):
        green.append(f"Increases your chance to dodge an attack by {fnum(s['Dodge'])}%.")
    if s.get('Parry'):
        green.append(f"Increases your chance to parry an attack by {fnum(s['Parry'])}%.")
    if s.get('Block'):
        green.append(f"Increases your chance to block attacks with a shield by {fnum(s['Block'])}%.")
    if s.get('BlockValue'):
        green.append(f"Increases the block value of your shield by {fnum(s['BlockValue'])}.")
    if s.get('Health'):
        white.append(['t', f"+{fnum(s['Health'])} Health"])
    if s.get('Mana'):
        white.append(['t', f"+{fnum(s['Mana'])} Mana"])
    if s.get('BonusArmor'):
        white.append(['t', f"+{fnum(s['BonusArmor'])} Armor"])
    return white, green


ITEM_SLOT = {1: 'Head', 2: 'Neck', 3: 'Shoulder', 4: 'Back', 5: 'Chest', 6: 'Wrist', 7: 'Hands', 8: 'Waist', 9: 'Legs', 10: 'Feet',
             11: 'Finger', 12: 'Trinket', 13: 'Weapon', 14: 'Ranged'}
ARMOR = {1: 'Cloth', 2: 'Leather', 3: 'Mail', 4: 'Plate'}
WEAPON = {1: 'Axe', 2: 'Dagger', 3: 'Fist Weapon', 4: 'Mace', 5: 'Held In Off-hand', 6: 'Polearm', 7: 'Shield', 8: 'Staff', 9: 'Sword'}
HAND = {1: 'Main Hand', 2: 'One-Hand', 3: 'Off Hand', 4: 'Two-Hand'}
RANGED = {1: 'Bow', 2: 'Crossbow', 3: 'Gun', 4: 'Idol', 5: 'Libram', 6: 'Thrown', 7: 'Totem', 8: 'Wand', 9: 'Sigil'}
# Forever relics: their only effect is an Equip line the sim implements in Go (sim/<class>/items.go comments).
RELIC_EQUIP = {272427: 'Reduces the cooldown of your Shifting Power ability by 1 sec.',
               272430: 'Increases the duration of your Insect Swarm spell by 2 sec.',
               272433: 'Increases the duration of your Flame Shock ability by 3 sec.',
               272435: 'Increases the damage of your Judgement ability by 4%.'}

# The synthetic on-use trinkets (standards 1b / 8a) carry no stats in the database; their effect is registered in Go
# (sim/<class>/forever_synthetic_trinkets.go, core.NewSimpleStatOffensiveTrinketEffect). Read it from there.
USE_STAT = {'SpellPower': 'spell power', 'AttackPower': 'attack power', 'RangedAttackPower': 'ranged attack power',
            'Strength': 'Strength', 'Agility': 'Agility', 'Intellect': 'Intellect'}


def synthetic_use(tree):
    out = {}
    for p in glob.glob(os.path.join(tree, 'sim', '*', 'forever_synthetic_trinkets.go')):
        src = open(p).read()
        ids = {k: int(v) for k, v in re.findall(r'(\w+)\s*=\s*(99\d{4})\b', src)}
        for m in re.finditer(r'NewSimpleStatOffensiveTrinketEffect\((\w+),\s*stats\.Stats\{([^}]*)\},\s*time\.Second\s*\*\s*(\d+),\s*time\.Minute\s*\*\s*(\d+)\)', src):
            iid = ids.get(m.group(1))
            if not iid:
                continue
            st = re.findall(r'stats\.(\w+):\s*(\d+)', m.group(2))
            d = dict(st)
            if d.get('AttackPower') and d.get('AttackPower') == d.get('RangedAttackPower'):
                what = f"attack power and ranged attack power by {d['AttackPower']}"
            else:
                what = ' and '.join(f"{USE_STAT.get(k, k)} by {v}" for k, v in st)
            out[iid] = f"Increases {what} for {m.group(3)} sec. ({m.group(4)} Min Cooldown)"
    return out


SETUP_ALIAS = {'goblin sapper': 'Goblin Sapper Charge', 'explosive goblin land mine': 'Goblin Land Mine', 'conjured demonic rune': 'Demonic Rune',
               'conjured rogue thistle tea': 'Thistle Tea', 'roids': 'R.O.I.D.S.', 'deadly poison': 'Deadly Poison V', 'instant poison': 'Instant Poison VI',
               'wizard oil': 'Wizard Oil', 'food flank au poivre': 'Flank au Poivre', 'flank au poivre': 'Flank au Poivre',
               'food grilled squid': 'Grilled Squid', 'food lobster stew': 'Lobster Stew', 'food smoked desert dumpling': 'Smoked Desert Dumplings',
               'dragon breath chili': 'Dragonbreath Chili', 'hunters mark': "Hunter's Mark", 'curse of elements': 'Curse of the Elements',
               'totem weapon buff (totemweaponbuffwindfury)': 'Windfury Totem', 'blood pact': 'Blood Pact', 'sanctity aura': 'Sanctity Aura',
               'leader of the pack': 'Leader of the Pack', 'moonkin aura': 'Moonkin Aura', 'judgement of the crusader': 'Judgement of the Crusader',
               'major frenzy potion': 'Major Frenzy Potion', 'major spellblasting potion': 'Major Spellblasting Potion'}
# The id behind each consumes / buffs / debuffs chip, as the sim UI's pickers link it (ui/core/components/inputs/*.ts),
# by the chip's value with any "(improved)" dropped. Forever's own ids where the UI uses them (Curse of the Elements, the 25xxxx consumables),
# and the rank the sim casts where Forever's launch lacks the top one (Battle Shout 6, Strength of Earth 4, Grace of Air 2: Ahn'Qiraj books).
SETUP_IDS = {
    'Scroll Of Imbue Greater Flame': ('item', 277500), 'Scroll Of Imbue Greater Frost': ('item', 277501), 'Scroll Of Imbue Precision': ('item', 277502),
    'Elixir Of Greater Agility': ('item', 9187), 'Elixir Of The Mongoose': ('item', 13452), 'Juju Might': ('item', 12460),
    'Conjured Demonic Rune': ('item', 12662), 'Conjured Rogue Thistle Tea': ('item', 7676), 'Major Frenzy Potion': ('item', 250943),
    'Major Mana Potion': ('item', 13444), 'Major Spellblasting Potion': ('item', 250937), 'Mighty Rage Potion': ('item', 13442),
    'Dragon Breath Chili': ('item', 12217), 'Explosive Goblin Land Mine': ('item', 4395), 'Elixir Of Firepower': ('item', 6373),
    'Elixir Of Greater Firepower': ('item', 21546), 'Flask Of Distilled Wisdom': ('item', 13511), 'Flask Of Supreme Power': ('item', 13512),
    'Flask Of The Titans': ('item', 13510), 'Food Flank Au Poivre': ('item', 250069), 'Food Grilled Squid': ('item', 13928),
    'Food Lobster Stew': ('item', 13933), 'Food Smoked Desert Dumpling': ('item', 20452), 'Elixir Of Frost Power': ('item', 17708),
    'Elixir Of Greater Intellect': ('item', 9179), 'Elixir Of The Owl': ('item', 250337), 'Deadly Poison': ('item', 20844),
    'Dense Sharpening Stone': ('item', 12404), 'Dense Weightstone': ('item', 12643), 'Elixir Of Cunning': ('item', 250328), 'Shadow Oil': ('item', 3824), 'Wizard Oil': ('item', 20750),
    'Greater Mageblood Elixir': ('item', 250341), 'Elixir Of Nature Power': ('item', 250343), 'Instant Poison': ('item', 8928),
    'Sapper Goblin Sapper': ('item', 10646), 'Elixir Of Shadow Power': ('item', 9264), 'Greater Arcane Elixir': ('item', 13454),
    'Elixir Of Greater Spirit': ('item', 250347), 'Elixir Of Sages': ('item', 250338), 'Elixir Of Ferocity': ('item', 250350),
    'Juju Power': ('item', 12451), 'Cerebral Cortex Compound': ('item', 8423), 'Ground Scorpok Assay': ('item', 8412), 'ROIDS': ('item', 8410),
    'Curse Of Elements': ('spell', 1311680), 'Curse Of Recklessness': ('spell', 11717), 'Curse Of Shadow': ('spell', 17937),
    'Curse Of Weakness': ('spell', 11708), 'Demoralizing Roar': ('spell', 9898), 'Demoralizing Shout': ('spell', 11556),
    'Expose Armor': ('spell', 11198), 'Faerie Fire': ('spell', 9907), 'Hunters Mark': ('spell', 14325), 'Insect Swarm': ('spell', 24977),
    'Judgement Of Light': ('spell', 20346), 'Judgement Of The Crusader': ('spell', 20303), 'Judgement Of Wisdom': ('spell', 20355),
    'Scorpid Sting': ('spell', 3043), 'Sunder Armor': ('spell', 11597), 'Thunder Clap': ('spell', 6343),
    'Arcane Brilliance': ('spell', 23028), 'Battle Shout': ('spell', 11551), 'Blood Pact': ('spell', 11767), 'Devotion Aura': ('spell', 10293),
    'Divine Spirit': ('spell', 27841), 'Fire Resistance Aura': ('spell', 19900), 'Frost Resistance Aura': ('spell', 19898),
    'Gift Of The Wild': ('spell', 21850), 'Grace Of Air Totem': ('spell', 10627), 'Leader Of The Pack': ('spell', 24932),
    'Mana Spring Totem': ('spell', 10497), 'Moonkin Aura': ('spell', 24907), 'Power Word Fortitude': ('spell', 10938),
    'Retribution Aura': ('spell', 10301), 'Sanctity Aura': ('spell', 20218), 'Shadow Protection': ('spell', 10958),
    'Strength Of Earth Totem': ('spell', 10442), 'Thorns': ('spell', 9910), 'Totem Weapon Buff (TotemWeaponBuffWindfury)': ('spell', 10614),
    'Trueshot Aura': ('spell', 20906), 'Blessing Of Kings': ('spell', 20217), 'Blessing Of Might': ('spell', 25291),
    'Blessing Of Wisdom': ('spell', 25290),
}


def setup_value(txt, cat):
    val = txt.split(': ', 1)[1] if ': ' in txt and cat == 'Consumes' else txt
    val = re.sub(r'\s*\((improved|\d+)\)$', '', val)
    if val == 'True':
        val = txt.split(': ', 1)[0]
    return val


PROFESSION_TEXT = {'Engineering': 'Engineering: lets the preset use explosives such as Goblin Sapper Charges.',
                   'Alchemy': 'Alchemy.', 'Blacksmithing': 'Blacksmithing.', 'Leatherworking': 'Leatherworking.', 'Tailoring': 'Tailoring.',
                   'Enchanting': 'Enchanting: lets the preset use ring enchants.'}

# Forever-only auras and consumables that none of the sources describe; the text follows the sim's code and
# the wiki's class notes (forever-wiki/classes/*.md, sim/core/consumes.go).
# Forever versions of spells whose wiki / Wowhead entries are the Classic or SoD spell (title -> note, text, meta)
FOREVER_TEXT = {
    'Rage of the Farseer': ('New in Forever (Enhancement talent)', 'Increases your attack speed by 30% for 25 sec. The beta build of 2026-09-24 removed its cast speed half.', ['Instant', '3 min cooldown']),
    'Bane of Agony': ('Renamed from Curse of Agony in Forever', 'Banes the target with agony, causing 552 Shadow damage over 24 sec. This damage is dealt slowly at first, and builds up as the Bane reaches its full duration. Only one Bane per Warlock can be active on any one target.'),
    'Bane of Doom': ('Renamed from Curse of Doom in Forever', 'Banes the target with impending doom, causing 1742 Shadow damage after 1 min. Only one Bane per Warlock can be active on any one target.'),
}
MANUAL_SPELL = {
    'Quick Shots': ('Deadly Aspects', 'Ranged attack speed increased by 30% for 12 sec. Auto Shot can grant it while Aspect of the Hawk is active.'),
    'Quick Strikes': ('Deadly Aspects', 'Melee attack speed increased by 30% for 12 sec. Melee auto attacks can grant it while Aspect of the Beast is active.'),
    'Dark Sacrifice': ('Undead priest racial', 'Gives up 1600 health over 15 sec for 1600 mana plus your Spirit. 10 min cooldown.'),
    'Echo of Command': ('Twist of Light', "Your next melee attack applies Seal of Command's effect, consuming the Echo."),
    'Echo of Righteousness': ('Twist of Light', "Your next melee attack applies Seal of Righteousness's effect, consuming the Echo."),
    'Burning Shadow': ('Demonic Sacrifice', 'Shadow damage increased by 15%. Granted by sacrificing the Imp.'),
    'Lacerate': ('New in Forever', 'Wounds the target, causing it to bleed for 406 damage over 21 sec.', ['95 Mana', 'Melee Range', 'Instant']),
    'Frostfire Bolt': ('New in Forever', 'Launches a bolt of frostfire at the enemy, causing Fire and Frost damage, more damage over 9 sec, and chilling the target. '
                       'Rank 3: 270 to 314 damage plus 57 over 9 sec.', ['370 Mana', '30 yd range', '3 sec cast']),
    'Shadow Word: Death': ('New in Forever', 'A word of dark binding that inflicts 444 to 472 Shadow damage to the target. If the target is not killed, '
                           'you take damage equal to 10% of your maximum health.', ['340 Mana', '30 yd range', 'Instant', '15 sec cooldown']),
}
# Spells Forever gave new ids, where the sim still keys the cast by the Classic id (as the sim UI's Curse of the Elements picker)
FOREVER_ID = {11722: 1311680}
TALENT_ALIAS = {'Shadow Vulnerability': 'Improved Shadow Bolt', 'Fire Vulnerability': 'Improved Scorch'}
MANUAL_ITEM = {
    'major frenzy potion': 'Use: Increases attack power and ranged attack power by 80.',
    'major spellblasting potion': 'Use: Increases damage done by magical spells and effects by up to 47.',
    'flank au poivre': 'Use: Well Fed: increases Agility by 20.',
    'elixir of the owl': 'Use: Increases Intellect by 25 and critical strike chance by 2%.',
    'elixir of sages': 'Use: Increases Spirit by 25 and critical strike chance by 2%.',
    'elixir of greater spirit': 'Use: Increases Spirit by 18.',
    'greater mageblood elixir': 'Use: Restores 20 mana per 5 sec.',
    'elixir of nature power': 'Use: Increases damage done by Nature spells and effects by up to 40.',
    'elixir of ferocity': 'Use: Increases Strength and Agility by 18.',
}

CLASSIC_NOTE = 'Classic text; Forever may differ'

# Where the sim's number for a buff or debuff chip differs from Forever wowhead's text: chip -> (number the text should carry,
# what the sim uses). From sim/core/buffs.go and debuffs.go at level 60 under Forever; a chip whose text lacks the number gets a
# muted "Sim uses: ..." line under wowhead's text. Filled in after comparing every chip (whdump.py).
SIM_USES = {
    'Demoralizing Roar': (193, '193 attack power off the boss (138, plus 40% for the improved setting).'),
    'Demoralizing Shout': (204, '204 attack power off the boss (146 at rank 5, plus 40% for the improved setting).'),
    'Blood Pact': (42, '42 Stamina.'),
    'Thunder Clap': (15, 'attacks 15% slower (the improved setting).'),
}


def sim_uses_lines(key, lines):
    if key not in SIM_USES:
        return []
    want, txt = SIM_USES[key]
    body = ' '.join(l[1] for l in lines if l[0] in ('y', 'g'))
    if re.search(r'(?<![\d.])' + re.escape(str(want)) + r'(?![\d])', body):
        return []
    return [['m', 'Sim uses: ' + txt]]


STAT_TEXT = {
    'Strength': 'Increases attack power: 2 per point for warriors, paladins, shamans and druids, 1 for rogues and hunters.',
    'Agility': 'Increases attack power, critical strike chance, armor and dodge chance.',
    'Intellect': 'Increases maximum mana and the chance to land a critical strike with spells.',
    'Spirit': 'Increases mana regeneration while not casting.',
    'AttackPower': 'Increases the damage of melee attacks and abilities that scale with attack power.',
    'RangedAttackPower': 'Increases the damage of ranged attacks.',
    'SpellPower': 'Increases damage and healing done by spells and effects.',
    'MP5': 'Mana restored every 5 seconds, even while casting.',
    'MeleeHit': 'Lowers the chance for attacks and spells to miss. Forever gear gives hit as a flat percent: 10 hit rating is 1%.',
    'MeleeHitDown': 'Losing hit: how much DPS this preset would give up with 1% less hit.',
    'MeleeCrit': 'Raises the chance to land a critical strike with attacks and spells. 14 crit rating is 1%.',
}


def build(tree, names, gear, setup, outs, abilities, auras, timelines, builds, item_db, enchants, talent_trees, racials_path, slots):
    spells, items = load_wowhead(tree)
    by_name = {}
    for sid, j in spells.items():
        nm = norm(j.get('name'))
        if not nm:
            continue
        cur = by_name.get(nm)
        key = ('class="q"' in (j.get('tooltip') or '') or bool(j.get('buff')), spell_rank(j), -sid)
        if cur is None or key > cur[0]:
            by_name[nm] = (key, sid)
    items_by_name = {}
    for iid, j in items.items():
        items_by_name.setdefault(norm(j.get('name')), iid)
    wiki = json.load(open(os.path.join(tree, 'forever-wiki/data/abilities.json')))
    wiki_by_name = {}
    for cls, lst in wiki.items():
        for e in lst:
            if e.get('status') == 'removed':
                continue
            wiki_by_name.setdefault(norm(e['name']), []).append((cls.capitalize(), e))
    tal_by_name = {}
    for cls, trees in talent_trees.items():
        for tr in trees:
            for t in tr['talents']:
                tal_by_name.setdefault(norm(t['name']), []).append((cls, tr['name'], t))
    racials = json.load(open(racials_path))
    racial_by_spell, racial_by_name = {}, {}
    for race, lst in racials.items():
        for r in lst:
            for v in r.get('variants', []):
                racial_by_spell.setdefault(v['spellId'], []).append((r['name'], v))
                racial_by_name.setdefault(r['name'], []).append((r['name'], v))

    # which classes use each key, to pick the right class's wiki / talent entry
    build_class = {b['id']: b['class'] for b in builds}
    key_classes = {}
    for r in abilities + auras:
        key_classes.setdefault(r['key'], set()).add(build_class[r['build']])
    for cls, rows in timelines.items():
        for s in rows:
            for c in s['casts']:
                key_classes.setdefault(c[1], set()).add(cls.capitalize())
            for d in s['dmg']:
                key_classes.setdefault(d[1], set()).add(cls.capitalize())
            for a in s['auras']:
                key_classes.setdefault(a[0], set()).add(cls.capitalize())

    rows = []
    missing = []
    extra = []
    renamed = []
    item_diffs = []
    ench_diffs = []

    def pick(cands, classes):
        for c, e in cands:
            if c in classes:
                return e
        return cands[0][1]

    def talent_text(t, rank=None):
        vals = t.get('ranks') or []
        r = (rank or len(vals) or 1) - 1
        v = vals[r] if r < len(vals) else (vals[-1] if vals else [])
        return re.sub(r'\{(\d+)\}', lambda m: str(v[int(m.group(1))]) if int(m.group(1)) < len(v) else m.group(0), t.get('description', ''))

    def wowhead_spell_lines(j, title):
        ls = clean_lines(html_lines(j.get('tooltip')), title)
        ls = [l for l in ls if not (l[0] != '2' and (re.match(r'^\(Proc chance|^Modifies ', l[1]) or l[1] in ('100 yd range', 'Unlimited range')
                                                      or (l[0] == 'y' and l[1].count(',') > 12)))]
        if [l for l in ls if l[0] != 'y'] == [['t', 'Instant']]:
            ls = [l for l in ls if l[0] == 'y']
        if not any(l[0] == 'y' for l in ls) and j.get('buff'):
            ls += [['y', l[1]] for l in clean_lines(html_lines(j.get('buff')), title)
                   if l[0] != '2' and not re.search(r'remaining$', l[1]) and l[1] not in (title, 'Magic')]
        return ls

    def item_lines_wowhead(j, title):
        ls = clean_lines(html_lines(j.get('tooltip')), title)
        return [l for l in ls if not (l[0] == 'm' and re.match(r'^\(\d\) Set', l[1]))]

    # ---- wowhead Forever first, Classic only where Forever's database lacks the id ----
    def fwh(kind, i):
        j = whcache.cached('forever', kind, i)
        return j if whcache.ok(j) else None

    def cwh(kind, i):
        j = whcache.cached('classic', kind, i)
        return j if whcache.ok(j) else None

    def set_link(row, kind, i):
        if i and fwh(kind, i):
            row['wh'], row['whd'] = f'{kind}={i}', 'forever'
        elif i and cwh(kind, i):
            row['wh'], row['whd'] = f'{kind}={i}', 'classic'

    def has_text(ls):
        return any(l[0] in ('y', 'g') for l in ls)

    # ---- spells, consumables used in fights, other actions ----
    for key, n in sorted(names.items()):
        nm = n['name']
        title = display_name(nm)
        rank = n.get('rank')
        classes = key_classes.get(key, set())
        row = {'key': key, 'kind': 'spell', 'title': title, 'sub': f'Rank {rank}' if rank else '', 'icon': n.get('icon')}
        lines, note, src = [], None, None
        m = re.match(r'^([sio])(\w+?)(?:t(-?\d+))?$', key)
        kind, num = (m.group(1), m.group(2)) if m else ('o', key)
        if kind == 'o':
            base = re.sub(r't\d+$', '', key)
            txt = OTHER_TEXT.get(key) or OTHER_TEXT.get(base)
            if txt:
                lines.append(['y', txt])
                src = 'sim'
        elif kind == 'i':
            iid = int(num)
            fj, cj = fwh('item', iid), cwh('item', iid)
            j = fj or cj
            if j:
                row['kind'] = 'item'
                row['quality'] = QUALITY.get(j.get('quality'), 'common')
                row['title'] = j.get('name') or title
                row['icon'] = row['icon'] or j.get('icon')
                lines = item_lines_wowhead(j, row['title'])
                src = 'forever' if fj else 'classic'
            elif norm(title) in {norm(k) for k in MANUAL_ITEM}:
                row['kind'] = 'item'
                row['quality'] = 'common'
                lines = [['g', next(v for k, v in MANUAL_ITEM.items() if norm(k) == norm(title))]]
                src = 'sim'
            set_link(row, 'item', iid)
        else:
            sid = int(num)
            wn = norm(title)
            w = wiki_by_name.get(wn)
            fj, cj = fwh('spell', FOREVER_ID.get(sid, sid)), cwh('spell', sid)
            if sid in FOREVER_ID and fj:
                sid = FOREVER_ID[sid]
            set_link(row, 'spell', sid)
            if fj and norm(fj.get('name')) != wn:
                renamed.append(f"{key} {title} -> Forever wowhead '{fj.get('name')}'")
            if title in FOREVER_TEXT and not (fj and norm(fj.get('name')) == wn):
                # the sim keys this Forever spell by a borrowed Classic / SoD id that wowhead files under the old spell
                note, txt, *meta = FOREVER_TEXT[title]
                lines = (meta_lines(meta[0]) if meta else []) + [['y', txt]]
                src = 'sim'
                row.pop('wh', None), row.pop('whd', None)
            if not lines and sid in racial_by_spell:
                # the sim can use one variant's spell id for every class (Eureka!), so match the racial by name and class
                vs = racial_by_name[racial_by_spell[sid][0][0]]
                rn, v = next(((n_, v_) for n_, v_ in vs if any('Class' + c in v_.get('classes', []) for c in classes)), vs[0])
                vj = fwh('spell', v['spellId'])
                vl = wowhead_spell_lines(vj, vj.get('name')) if vj else []
                if has_text(vl):
                    lines, src = vl, 'forever'
                else:
                    lines = ([['t', v['cost'].replace('; ', ', ')]] if v.get('cost') else []) + [['y', v['description']]]
                    src = 'client'
                row.pop('wh', None), row.pop('whd', None)
                set_link(row, 'spell', v['spellId'])
                note = 'Racial'
                row['icon'] = row['icon'] or v.get('icon')
            if not lines and fj:
                fl = wowhead_spell_lines(fj, fj.get('name'))
                if has_text(fl):
                    lines, src = fl, 'forever'
                    row['icon'] = row['icon'] or fj.get('icon')
            if w and not lines:
                e = pick(w, classes)
                rk = next((r for r in e.get('ranks') or [] if rank and r.get('t') == f'Rank {rank}'), None)
                desc = (rk or {}).get('d') or e.get('desc') or ''
                meta = (rk or {}).get('m') if rk else e.get('meta')
                if desc:
                    lines = meta_lines(meta) + [['y', desc]]
                    note = {'changed': 'Changed in Forever', 'new': 'New in Forever'}.get(e.get('status'))
                    row['icon'] = row['icon'] or e.get('icon')
                    src = 'wiki'
            if not lines and wn in tal_by_name:
                cands = tal_by_name[wn]
                cls, tree_name, t = next(((c, tn, t) for c, tn, t in cands if c in classes), cands[0])
                lines = [['y', talent_text(t)]]
                note = f'{tree_name} talent'
                row['icon'] = row['icon'] or t.get('icon')
                src = 'client'
                if 'wh' not in row:
                    set_link(row, 'spell', (t.get('spellIds') or [0])[-1])
            if not lines and title in MANUAL_SPELL:
                note, txt, *meta = MANUAL_SPELL[title]
                lines = (meta_lines(meta[0]) if meta else []) + [['y', txt]]
                src = 'sim'
            tal_alias = norm(TALENT_ALIAS.get(title, ''))
            if not lines and tal_alias in tal_by_name:
                cls, tree_name, t = tal_by_name[tal_alias][0]
                lines = [['y', talent_text(t)]]
                note = f"{t['name']} ({tree_name} talent)"
                row['icon'] = row['icon'] or t.get('icon')
                src = 'client'
            if not lines and title.startswith('Judgement of '):
                seal = norm('Seal of ' + title[len('Judgement of '):])
                if seal in wiki_by_name:
                    e = pick(wiki_by_name[seal], classes)
                    lines = [['y', e.get('desc', '')]]
                    note = 'Judging the seal: ' + e['name']
                    src = 'wiki'
            if not lines and title in racial_by_name:
                vs = racial_by_name[title]
                rn, v = next(((n_, v_) for n_, v_ in vs if any('Class' + c in v_.get('classes', []) for c in classes)), vs[0])
                vj = fwh('spell', v['spellId'])
                vl = wowhead_spell_lines(vj, vj.get('name')) if vj else []
                lines, src = (vl, 'forever') if has_text(vl) else (([['t', v['cost'].replace('; ', ', ')]] if v.get('cost') else []) + [['y', v['description']]], 'client')
                note = 'Racial'
                row.pop('wh', None), row.pop('whd', None)
                set_link(row, 'spell', v['spellId'])
            # the same spell or item under another id Forever's database does carry
            if not lines and wn in by_name and fwh('spell', by_name[wn][1]):
                cid = by_name[wn][1]
                fl = wowhead_spell_lines(fwh('spell', cid), fwh('spell', cid).get('name'))
                if has_text(fl):
                    lines, src = fl, 'forever'
                    row.pop('wh', None), row.pop('whd', None)
                    set_link(row, 'spell', cid)
            if not lines and wn in items_by_name and fwh('item', items_by_name[wn]):
                ciid = items_by_name[wn]
                lines, src = item_lines_wowhead(fwh('item', ciid), fwh('item', ciid).get('name')), 'forever'
                row.pop('wh', None), row.pop('whd', None)
                set_link(row, 'item', ciid)
            # Classic's text only where nothing from Forever has it
            if not lines and cj:
                cl = wowhead_spell_lines(cj, cj.get('name'))
                if has_text(cl):
                    lines, src, note = cl, 'classic', CLASSIC_NOTE
                    row['icon'] = row['icon'] or cj.get('icon')
            if not lines and wn in by_name:
                cid = by_name[wn][1]
                cl = wowhead_spell_lines(spells[cid], spells[cid].get('name'))
                if has_text(cl):
                    lines, src, note = cl, 'classic', CLASSIC_NOTE
                    row['icon'] = row['icon'] or spells[cid].get('icon')
                    if 'wh' not in row:
                        set_link(row, 'spell', cid)
            if not lines and wn in items_by_name:
                ciid = items_by_name[wn]
                lines, src, note = item_lines_wowhead(items[ciid], items[ciid].get('name')), 'classic', CLASSIC_NOTE
                if 'wh' not in row:
                    set_link(row, 'item', ciid)
            # a Forever page filed under another spell (the sim keys Searing Totem's damage by its "Attack" spell) is no link for this one
            if fj and src != 'forever' and row.get('wh') == f'spell={sid}' and norm(fj.get('name')) != wn:
                row.pop('wh', None), row.pop('whd', None)
        if note:
            lines = [['m', note]] + lines
        # a racial whose text differs by class gets one row per class as well (<key>:<Class>)
        rvs = racial_by_name.get(racial_by_spell[int(num)][0][0]) if kind == 's' and int(num) in racial_by_spell else None
        if rvs and len({v['description'] for _, v in rvs}) > 1:
            for cls in sorted(classes):
                v = next((v for _, v in rvs if 'Class' + cls in v.get('classes', [])), None)
                if v:
                    vj = fwh('spell', v['spellId'])
                    vl = wowhead_spell_lines(vj, vj.get('name')) if vj else []
                    if not has_text(vl):
                        vl = ([['t', v['cost'].replace('; ', ', ')]] if v.get('cost') else []) + [['y', v['description']]]
                    xr = {**row, 'key': f'{key}:{cls}', 'lines': [['m', 'Racial']] + vl, 'src': 'forever' if vj and has_text(wowhead_spell_lines(vj, vj.get('name'))) else 'client'}
                    xr.pop('wh', None), xr.pop('whd', None)
                    set_link(xr, 'spell', v['spellId'])
                    extra.append(xr)
        if not lines:
            missing.append(f'{key} {title}')
        row['lines'] = lines
        row['src'] = src if lines else None
        rows.append(row)

    rows += extra

    # ---- gear ----
    synth_use = synthetic_use(tree)
    setb = set_bonuses(tree)
    set_size = {}
    for it in item_db.values():
        if it.get('setName'):
            set_size[it['setName']] = set_size.get(it['setName'], 0) + 1
    seen = set()
    for g in gear:
        iid = g['itemId']
        if iid in seen:
            continue
        seen.add(iid)
        it = item_db.get(iid, {})
        lines = []
        if it.get('ilvl'):
            lines.append(['y', f"Item Level {it['ilvl']}"])
        typ = it.get('type')
        if typ == 13:
            lines.append(['2', HAND.get(it.get('handType'), 'Weapon'), WEAPON.get(it.get('weaponType'), '')])
        elif typ == 14:
            rt = it.get('rangedWeaponType')
            lines.append(['2', 'Relic' if rt in (4, 5, 7, 9) else 'Ranged', RANGED.get(rt, '')])
        elif typ in ITEM_SLOT:
            at = ARMOR.get(it.get('armorType'))
            lines.append(['2', ITEM_SLOT[typ], at] if at else ['t', ITEM_SLOT[typ]])
        if it.get('weaponDamageMax'):
            lo, hi, sp = it['weaponDamageMin'], it['weaponDamageMax'], it.get('weaponSpeed') or 0
            lines.append(['2', f'{fnum(lo)} - {fnum(hi)} Damage', f'Speed {sp:.2f}'])
            if sp:
                lines.append(['t', f'({(lo + hi) / 2 / sp:.1f} damage per second)'])
        white, green = stat_lines(it.get('stats'), it.get('hitRating') or 0, it.get('critRating') or 0)
        fj, cj = (fwh('item', iid), cwh('item', iid)) if iid < 990000 else (None, None)
        src = 'db'
        if fj or cj:
            # a real item: wowhead Forever's text (Classic's where Forever lacks it), and the sim's numbers where they differ
            j = fj or cj
            wl = item_lines_wowhead(j, j.get('name'))
            shown = ' '.join(' '.join(l[1:]) for l in wl)
            diff = [t[1] for t in white if t[1] not in shown] + [t for t in green if t.rstrip('.') not in shown]
            il = re.search(r'Item Level (?:<!--ilvl-->)?(\d+)', j.get('tooltip') or '')
            lines = ([['m', CLASSIC_NOTE]] if not fj else []) + ([['y', f'Item Level {il.group(1)}']] if il else lines[:1]) + wl
            if il and it.get('ilvl') and int(il.group(1)) != it['ilvl']:
                diff.append(f"item level {it['ilvl']}")
            if diff:
                lines.append(['m', 'Sim uses: ' + '; '.join(diff)])
                item_diffs.append(f"i{iid} {it.get('name')}: {'; '.join(diff)}")
            src = 'forever' if fj else 'classic'
        else:
            lines += white
            if iid in RELIC_EQUIP:
                green.append(RELIC_EQUIP[iid])
            if not white and not green and iid in items:
                lines += [l for l in item_lines_wowhead(items[iid], it.get('name')) if l not in lines]
            lines += [['g', 'Equip: ' + t] for t in green]
        if iid in synth_use:
            lines.append(['g', 'Use: ' + synth_use[iid]])
        row = {'key': f'i{iid}', 'kind': 'item', 'title': it.get('name') or g['item'], 'sub': '', 'icon': it.get('icon') or g.get('icon'),
               'quality': QUALITY.get(it.get('quality'), 'common'), 'lines': lines, 'src': src}
        if iid < 990000:
            set_link(row, 'item', iid)
        if it.get('setName'):
            row['set'] = it['setName']
            row['setSize'] = set_size.get(it['setName'])
            row['setBonuses'] = setb.get(it['setName'], [])
        rows = [r for r in rows if r['key'] != row['key']] + [row]

    # ---- enchants on the gear ----
    ench_text = {}
    for eid in sorted({it['enchant'] for o in outs for it in o['setup']['equipment'] if it and it.get('enchant')}):
        e = enchants.get(eid)
        if not e:
            continue
        white, green = stat_lines(e.get('stats'))
        lines = [['g', l[1]] for l in white] + [['g', t] for t in green]
        esrc = 'db' if lines else None
        if lines:
            # the stat lines are the sim's; they count as Forever's text when wowhead Forever's enchant text carries the same numbers
            fj = fwh('item', e['itemId']) if e.get('itemId') and fwh('item', e['itemId']) else fwh('spell', e.get('spellId'))
            ft = ' '.join(l[1] for l in html_lines((fj or {}).get('tooltip')) if l[0] in ('y', 'g'))
            nums = {n for l in lines for n in re.findall(r'\d+', l[1])}
            if ft and not re.search(r' by 0\b', ft) and all(re.search(r'(?<!\d)' + n + r'(?!\d)', ft) for n in nums):
                esrc = 'forever'
            elif ft:
                ench_diffs.append(f"{e['name']}: sim {'; '.join(l[1] for l in lines)} / wowhead Forever: {ft[:120]}")
        if not lines and e.get('spellId'):
            fj, cj = fwh('spell', e['spellId']), cwh('spell', e['spellId'])
            for j, sr in ((fj, 'forever'), (cj, 'classic')):
                # Forever's text for a few reworked kits lost its numbers ("by 0"), like the sim UI we skip those
                if j and not re.search(r' by 0\b', j.get('tooltip') or ''):
                    desc = [l for l in clean_lines(html_lines(j.get('tooltip')), j.get('name')) if l[0] == 'y']
                    if desc:
                        lines, esrc = [['g', d[1]] for d in desc], sr
                        break
        ench_text[eid] = (e['name'], lines, esrc, e)
        if not lines:
            missing.append('enchant ' + e['name'])
    # one row per enchanted gear slot (ench:<build>:<slot>): several enchants share a name, like the three Lesser Arcanum of Voracity
    for o in outs:
        for i, it in enumerate(o['setup']['equipment'][:len(slots)]):
            if it and it.get('id') and it.get('enchant') in ench_text:
                nm, ls, esrc, e = ench_text[it['enchant']]
                er = {'key': f"ench:{o['id']}:{slots[i]}", 'kind': 'enchant', 'title': nm, 'sub': '', 'icon': None, 'lines': ls, 'src': esrc}
                if e.get('itemId') and fwh('item', e['itemId']):
                    set_link(er, 'item', e['itemId'])
                else:
                    set_link(er, 'spell', e.get('spellId'))
                rows.append(er)

    # ---- consumes, buffs and debuffs chips ----
    for txt in sorted({s['item'] for s in setup}):
        cat = next(s['category'] for s in setup if s['item'] == txt)
        val = setup_value(txt, cat)
        improved = txt.endswith('(improved)')
        stripped = re.sub(r'^(Food|Sapper|Conjured|Explosive) ', '', val)
        cands = [SETUP_ALIAS.get(val.lower()), SETUP_ALIAS.get(stripped.lower()), val, stripped]
        row = {'key': 'setup:' + txt, 'kind': 'setup', 'title': val, 'sub': cat if cat != 'Consumes' else '', 'icon': None, 'lines': [], 'src': None}
        if cat == 'Professions':
            row['lines'] = [['y', PROFESSION_TEXT.get(val, val + '.')]]
            row['title'] = val
            row['src'] = 'sim'
            rows.append(row)
            continue
        sk, si = SETUP_IDS.get(val, (None, None))
        if sk:
            set_link(row, sk, si)
            fj = fwh(sk, si)
            if fj and sk == 'item':
                row.update(kind='item', title=fj['name'], quality=QUALITY.get(fj.get('quality'), 'common'), icon=fj.get('icon'))
                row['lines'], row['src'] = item_lines_wowhead(fj, fj['name']), 'forever'
            elif fj:
                ls = [l for l in wowhead_spell_lines(fj, fj.get('name')) if l[0] == 'y']
                if ls:
                    row.update(title=fj['name'], icon=fj.get('icon'))
                    row['lines'], row['src'] = ls, 'forever'
        for c in cands if not row['lines'] else []:
            if not c:
                continue
            nc = norm(c)
            if nc in wiki_by_name and not (sk == 'item'):
                e = wiki_by_name[nc][0][1]
                row.update(title=e['name'], icon=e.get('icon'))
                row['lines'], row['src'] = [['y', e.get('desc', '')]], 'wiki'
                break
            cj = cwh(sk, si) if sk else None
            if cj and sk == 'item':
                row.update(kind='item', title=cj['name'], quality=QUALITY.get(cj.get('quality'), 'common'), icon=cj.get('icon'))
                row['lines'], row['src'] = [['m', CLASSIC_NOTE]] + item_lines_wowhead(cj, cj['name']), 'classic'
                break
            if cj:
                ls = [l for l in wowhead_spell_lines(cj, cj.get('name')) if l[0] == 'y']
                if ls:
                    row.update(title=cj['name'], icon=cj.get('icon'))
                    row['lines'], row['src'] = [['m', CLASSIC_NOTE]] + ls, 'classic'
                    break
            if cat == 'Consumes' and nc in items_by_name:
                j = items[items_by_name[nc]]
                row.update(kind='item', title=j['name'], quality=QUALITY.get(j.get('quality'), 'common'), icon=j.get('icon'))
                row['lines'], row['src'] = [['m', CLASSIC_NOTE]] + item_lines_wowhead(j, j['name']), 'classic'
                break
            if nc in by_name:
                j = spells[by_name[nc][1]]
                ls = wowhead_spell_lines(j, j.get('name'))
                if any(l[0] == 'y' for l in ls):
                    row.update(title=j['name'], icon=j.get('icon'))
                    row['lines'], row['src'] = [['m', CLASSIC_NOTE]] + [l for l in ls if l[0] == 'y'], 'classic'
                    break
        if not row['lines']:
            mi = next((v for k, v in MANUAL_ITEM.items() if norm(k) in (norm(val), norm(stripped))), None)
            if mi:
                row.update(kind='item', quality='common', title=next(SETUP_ALIAS.get(x.lower()) or x for x in (stripped,)))
                row['lines'], row['src'] = [['g', mi]], 'sim'
        row['lines'] += sim_uses_lines(val, row['lines'])
        if improved:
            row['lines'].append(['m', 'Improved by talents'])
        if cat != 'Consumes' and row['sub'] == '':
            row['sub'] = cat
        if not row['lines']:
            missing.append('setup ' + txt)
        rows.append(row)

    # ---- races ----
    race_classes = {}
    for b in outs:
        for r in b['races']:
            race_classes.setdefault(RACE_NAMES.get(r['race'], r['race']), set()).add(b['meta']['class'])
    for rkey, lst in racials.items():
        race = RACE_NAMES.get(rkey[4:], rkey[4:])
        fac = 'Horde' if race in HORDE else 'Alliance'
        generic = []
        for r in lst:
            vs = r.get('variants', [])
            descs = {v['description'] for v in vs}
            for v in vs:
                cls = [c.replace('Class', '') for c in v.get('classes', [])]
                right = 'Passive' if v.get('passive') else (v.get('cost') or '').replace('; ', ', ')
                generic.append(['h', r['name'], right])
                if len(descs) > 1:
                    generic.append(['m', ', '.join(cls)])
                generic.append(['y', v['description']])
                if len(descs) == 1:
                    break
        rows.append({'key': f'race:{race}', 'kind': 'race', 'title': race, 'sub': fac, 'icon': None, 'lines': generic})
        for cls in CLASS_ORDER:
            if cls not in race_classes.get(race, set()):
                continue
            ls = []
            for r in lst:
                v = next((v for v in r.get('variants', []) if 'Class' + cls in v.get('classes', [])), None)
                if not v:
                    continue
                right = 'Passive' if v.get('passive') else (v.get('cost') or '').replace('; ', ', ')
                ls += [['h', r['name'], right], ['y', v['description']]]
            rows.append({'key': f'race:{race}:{cls}', 'kind': 'race', 'title': race, 'sub': f'{fac} {cls}', 'icon': None, 'lines': ls})

    # ---- stat weights ----
    for k, t in STAT_TEXT.items():
        rows.append({'key': 'stat:' + k, 'kind': 'stat', 'title': k, 'sub': '', 'icon': None, 'lines': [['m', t]]})

    cols = ['key', 'kind', 'title', 'sub', 'icon', 'quality', 'lines', 'set', 'setSize', 'setBonuses', 'wh', 'whd', 'src']
    report = {'renamed': renamed, 'itemDiffs': item_diffs, 'enchantDiffs': ench_diffs}
    return [{k: (r.get(k) if r.get(k) != '' else None) for k in cols} for r in rows], missing, report
