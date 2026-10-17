#!/usr/bin/env python3
"""What each permanent enchant gives, read out of the Forever client's tables.

    python3 tools/sources/enchants.py <build>

It writes assets/db_inputs/forever_enchants.json, and gen_db takes the stats of
every enchant in tools/database/forever_enchants.go from it. The Go list still
says which enchants we offer, their names, slots and level rules, because that
is our choice (which formulas have a source, what the beta has). Only the
numbers come from here, so a build that changes an enchant changes the sim the
next time gen_db runs, and nobody has to retype a value.

The file has one entry per spell that puts a permanent enchant on an item:
every Enchanting recipe, every item whose use enchants (the armor kits), and
every recipe that makes such an item (with the item in `item`). The key is the
spell ID, which is also the sim's SpellId for the enchant.

    "13419": {"enchant": 247, "name": "+3 Agility", "stats": {"Agility": 3}}

`stats` uses the sim's stat names (sim/core/stats StatName). `damage` is
weapon damage ("+4 Weapon Damage"). `other` lists what we couldn't turn into a
stat (a proc, a percent, threat, a skill), which the sim models by hand when it
models it at all, like the Impact and Revelation weapon effects in sim/common.

How we read an enchant is the same as bayovui's build/enchants.py, which reads
the same tables for BayovCore's enchant advisor. SpellEffect effect 53 on a
spell names its SpellItemEnchantment row. That row has up to three effects:

    1  a proc
    2  weapon damage
    3  a passive spell while worn, whose auras we turn into stats
    4  a resistance, the school in EffectArg (0 is armor)
    5  a stat in the item stat numbering (ITEM_MOD), Forever's newer enchants

A healing enchant ("Healing +11, Damage +4") gives HealingPower and
SpellDamage separately, and one that heals and damages by the same amount
gives SpellPower, the way the sim's item stats do.
"""
import collections
import json
import os
import sys

sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
import meta  # noqa: E402
import wago  # noqa: E402

ROOT = os.path.abspath(os.path.join(os.path.dirname(os.path.abspath(__file__)), '..', '..'))
OUT = os.path.join(ROOT, 'assets', 'db_inputs', 'forever_enchants.json')

# source: wago SpellEffect, ItemEffect and SpellItemEnchantment numbers 1.60.1.70450
ENCHANTING = 333
ENCHANT_ITEM = 53
CREATE_ITEM = 24
TRIGGER_USE = 0
PROC, DAMAGE, EQUIP_SPELL, RESISTANCE, STAT = 1, 2, 3, 4, 5

# The item stat numbering of effect 5. 84 to 89 are Forever's own school
# damage stats. 41 and 42 are the healing and the damage of a healing enchant.
ITEM_MOD = {
    0: ['Mana'], 1: ['Health'], 3: ['Agility'], 4: ['Strength'], 5: ['Intellect'],
    6: ['Spirit'], 7: ['Stamina'], 12: ['Defense'],
    38: ['AttackPower', 'RangedAttackPower'], 39: ['RangedAttackPower'],
    43: ['MP5'], 45: ['SpellPower'], 47: ['SpellPenetration'], 48: ['BlockValue'],
    84: ['HolyPower'], 85: ['FirePower'], 86: ['NaturePower'], 87: ['FrostPower'],
    88: ['ShadowPower'], 89: ['ArcanePower'],
}
HEALING, DAMAGE_DONE = 41, 42
RESIST_INDEX = {0: 'BonusArmor', 2: 'FireResistance', 3: 'NatureResistance',
                4: 'FrostResistance', 5: 'ShadowResistance', 6: 'ArcaneResistance'}
STAT_INDEX = {0: 'Strength', 1: 'Agility', 2: 'Stamina', 3: 'Intellect', 4: 'Spirit'}
SCHOOL_MASK = {2: 'Holy', 4: 'Fire', 8: 'Nature', 16: 'Frost', 32: 'Shadow', 64: 'Arcane'}
ALL_MAGIC = 126
DEFENSE_SKILL = 95

# Aura numbers of the passive spells (effect 3) we turn into stats.
MOD_DAMAGE_DONE, MOD_RESISTANCE, MOD_STAT, MOD_SKILL = 13, 22, 29, 30
MOD_HEALTH, MOD_MANA, MOD_POWER_REGEN = 34, 35, 85
MOD_AP, MOD_RAP, MOD_HEALING_DONE = 99, 124, 135
MOD_BONUS_ARMOR, MOD_BLOCK_VALUE = 674, 274


def num(v):
    f = float(v or 0)
    return int(f) if f.is_integer() else f


def stats_of(e, effects):
    """(stats, weapon damage, what we couldn't read) of a SpellItemEnchantment row."""
    stats = collections.Counter()
    damage, other = 0, []
    healing = dmg = 0

    def add(keys, v):
        for k in keys:
            stats[k] += v

    for i in range(3):
        kind = num(e['Effect_%d' % i])
        points = num(e['EffectPointsMin_%d' % i])
        arg = num(e['EffectArg_%d' % i])
        if kind == 0:
            continue
        if kind == STAT:
            if arg == HEALING:
                healing += points
            elif arg == DAMAGE_DONE:
                dmg += points
            elif arg in ITEM_MOD:
                add(ITEM_MOD[arg], points)
            else:
                other.append('item stat %d' % arg)
        elif kind == RESISTANCE:
            if arg in RESIST_INDEX:
                add([RESIST_INDEX[arg]], points)
            else:
                other.append('resistance %d' % arg)
        elif kind == DAMAGE:
            damage += points
        elif kind == EQUIP_SPELL:
            for se in effects.get(arg, ()):
                aura, misc = num(se['EffectAura']), num(se['EffectMiscValue_0'])
                v = num(se['EffectBasePointsF'])
                if aura == 0:
                    continue
                if aura == MOD_STAT and misc in STAT_INDEX:
                    add([STAT_INDEX[misc]], v)
                elif aura == MOD_STAT and misc == -1:
                    add(STAT_INDEX.values(), v)
                elif aura == MOD_DAMAGE_DONE and misc & ALL_MAGIC == ALL_MAGIC:
                    dmg += v
                elif aura == MOD_DAMAGE_DONE and misc in SCHOOL_MASK:
                    add([SCHOOL_MASK[misc] + 'Power'], v)
                elif aura == MOD_HEALING_DONE:
                    healing += v
                elif aura == MOD_RESISTANCE:
                    if misc & 1:
                        add(['BonusArmor'], v)
                    for bit, school in SCHOOL_MASK.items():
                        if misc & bit and school != 'Holy':
                            add([school + 'Resistance'], v)
                elif aura == MOD_AP:
                    add(['AttackPower', 'RangedAttackPower'], v)
                elif aura == MOD_RAP:
                    add(['RangedAttackPower'], v)
                elif aura == MOD_POWER_REGEN and misc == 0:
                    add(['MP5'], v)
                elif aura == MOD_BONUS_ARMOR:
                    add(['BonusArmor'], v)
                elif aura == MOD_BLOCK_VALUE:
                    add(['BlockValue'], v)
                elif aura == MOD_HEALTH:
                    add(['Health'], v)
                elif aura == MOD_MANA:
                    add(['Mana'], v)
                elif aura == MOD_SKILL and misc == DEFENSE_SKILL:
                    add(['Defense'], v)
                else:
                    other.append('spell %d aura %d' % (arg, aura))
        elif kind == PROC:
            other.append('proc spell %d' % arg)
        else:
            other.append('effect %d' % kind)
    if healing == dmg:
        add(['SpellPower'], dmg)
    else:
        add(['HealingPower'], healing)
        add(['SpellDamage'], dmg)
    return {k: v for k, v in sorted(stats.items()) if v}, damage, other


def read(build):
    """{spell: entry} for every spell that puts a permanent enchant on an item."""
    t = {name: wago.rows(name, build) for name in (
        'SpellItemEnchantment', 'SpellEffect', 'SkillLineAbility', 'ItemEffect', 'ItemXItemEffect')}
    effects = collections.defaultdict(list)
    for r in t['SpellEffect']:
        effects[int(r['SpellID'])].append(r)
    sie = {int(r['ID']): r for r in t['SpellItemEnchantment']}

    def enchant_of(spell):
        for e in effects.get(spell, ()):
            if num(e['Effect']) == ENCHANT_ITEM and num(e['EffectMiscValue_0']) > 0:
                return num(e['EffectMiscValue_0'])
        return None

    spells = {int(r['Spell']) for r in t['SkillLineAbility'] if num(r['SkillLine']) == ENCHANTING}
    item_effects = {int(r['ID']): r for r in t['ItemEffect']}
    use = {}
    for link in t['ItemXItemEffect']:
        e = item_effects.get(num(link['ItemEffectID']))
        if e is not None and num(e['TriggerType']) == TRIGGER_USE:
            use[num(link['ItemID'])] = num(e['SpellID'])
            spells.add(num(e['SpellID']))

    # A recipe that makes an enchanting item (Leatherworking's Forceful and
    # Mystic kits) counts as that item's enchant too, because the sim lists
    # those kits by the recipe's spell.
    made = {}
    for r in t['SkillLineAbility']:
        spell = int(r['Spell'])
        for e in effects.get(spell, ()):
            if num(e['Effect']) == CREATE_ITEM and use.get(num(e['EffectItemType'])):
                made[spell] = num(e['EffectItemType'])

    out = {}
    for spell in sorted(spells | set(made)):
        ench = enchant_of(use[made[spell]] if spell in made else spell)
        row = sie.get(ench) if ench else None
        if row is None or num(row['Duration']) > 0:
            continue
        stats, damage, other = stats_of(row, effects)
        entry = {'enchant': ench, 'name': row['Name_lang'], 'stats': stats}
        if spell in made:
            entry['item'] = made[spell]
        if damage:
            entry['damage'] = damage
        if other:
            entry['other'] = other
        out[str(spell)] = entry
    return out


def main(argv):
    if len(argv) != 2:
        print(__doc__.split('\n\n')[1].strip(), file=sys.stderr)
        return 2
    build = wago.version(argv[1])
    enchants = read(build)
    with open(OUT, 'w', encoding='utf-8') as f:
        f.write('{\n' + meta.line(meta.meta('client beta %s (wago.tools)' % build,
                                            'tools/sources/enchants.py', build=build)))
        f.write('"enchants": {\n')
        f.write(',\n'.join('%s: %s' % (json.dumps(k), json.dumps(v, sort_keys=True))
                           for k, v in enchants.items()))
        f.write('\n}}\n')
    print('%d enchants from build %s into %s' % (len(enchants), build, os.path.relpath(OUT, ROOT)))
    return 0


if __name__ == '__main__':
    raise SystemExit(main(sys.argv))
