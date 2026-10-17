#!/usr/bin/env python3
"""Every class ability of the Forever client, with its text per rank.

    python3 tools/sources/abilities.py <build>

It writes forever-wiki/data/abilities.json, which the DPS report reads for
ability icons and tooltips (tools/dps_report/build_data.py and tooltips.py).
That file used to be wowforevertalents.com's pre-beta transcription, where 316
of 442 abilities were Classic Era's text marked "not yet verified". Now it is
the client's, in the same shape, so the readers didn't change.

What we read:

- The abilities: SkillLineAbility rows on a class skill line (SkillLine
  category 7, without the pet, Beast Training and Engraving lines). Spells the
  client hides from the spellbook (Attributes_0 bit 0x80) are helpers and left
  out. So are Season of Discovery's rune spells the client still carries
  (Lava Lash, Decoy Totem, the Overload copies of Lightning Bolt): rows with
  AcquireMethod 3 and a spell ID in SoD's range, unless a talent teaches the
  spell on Forever. AcquireMethod 3 alone isn't enough, since Tiger's Fury
  (5217) has it too. The rogue's Poisons and Lockpicking lines are secondary
  skills (category 9) in the client, and count as class lines here.
- The class of a skill line: the SpellClassSet most of its spells have
  (SpellClassOptions), since two classes can have a line of the same name
  (Holy, Protection).
- An ability's ranks: the spells of the same name on that line, in the order
  of the level they're learned at (SpellLevels.BaseLevel). The rank title is
  the spell's subtext ("Rank 3").
- The text, icon and cost line of each rank: tools/sources/spelltext.py, the
  reader the talent and racial importers use.
- `status`: "new" when no rank's spell ID is in Classic 1.15.3's tables,
  "changed" when the top rank's tooltip template differs from Classic's, and
  "unchanged" otherwise. `fromTalent` is true for a spell a talent teaches
  (TraitDefinition).
"""
import collections
import json
import os
import re
import sys

sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
import meta  # noqa: E402
import spelltext  # noqa: E402
import wago  # noqa: E402

ROOT = os.path.abspath(os.path.join(os.path.dirname(os.path.abspath(__file__)), '..', '..'))
OUT = os.path.join(ROOT, 'forever-wiki', 'data', 'abilities.json')
META = os.path.join(ROOT, 'forever-wiki', 'data', 'abilities.meta.json')
CLASSIC = '1.15.3.55646'

# source: SpellClassOptions.SpellClassSet numbers, client beta 1.60.1.70450
FAMILY = {3: 'mage', 4: 'warrior', 5: 'warlock', 6: 'priest', 7: 'druid', 8: 'rogue', 9: 'hunter',
          10: 'paladin', 11: 'shaman'}
CLASS_ORDER = ['warrior', 'rogue', 'shaman', 'paladin', 'hunter', 'priest', 'mage', 'warlock', 'druid']
CLASS_CATEGORY = '7'
NOT_CLASS_LINES = ('Beast Training', 'Engraving')
EXTRA_CLASS_LINES = ('Poisons', 'Lockpicking')
RUNE = '3'
SOD_SPELLS = range(400000, 1000000)
HIDDEN = 0x80


def slug(text):
    return re.sub(r'[^a-z0-9]+', '-', text.lower().replace("'", '')).strip('-')


def meta_list(spell, build):
    """The cost line as the list the old file had (["150 Mana", "Instant"]), or ["Passive"]."""
    if spelltext.passive(spell, build):
        return ['Passive']
    line = spelltext.cost(spell, build)
    return [m for m in line.split('; ') if m] if line else []


def read(build):
    lines = {r['ID']: r['DisplayName_lang'] for r in wago.rows('SkillLine', build)
             if (r['CategoryID'] == CLASS_CATEGORY and not r['DisplayName_lang'].startswith('Pet - ')
                 and r['DisplayName_lang'] not in NOT_CLASS_LINES)
             or r['DisplayName_lang'] in EXTRA_CLASS_LINES}
    misc = {r['SpellID']: r for r in wago.rows('SpellMisc', build)}
    talent = {r['SpellID'] for r in wago.rows('TraitDefinition', build)}
    family = {r['SpellID']: int(r['SpellClassSet']) for r in wago.rows('SpellClassOptions', build)}
    level = {r['SpellID']: int(r['BaseLevel']) for r in wago.rows('SpellLevels', build)}
    subtext = {r['ID']: r['NameSubtext_lang'] for r in wago.rows('Spell', build)}
    template = {r['ID']: r['Description_lang'] for r in wago.rows('Spell', build)}
    classic = {r['ID']: r['Description_lang'] for r in wago.rows('Spell', CLASSIC)}

    rows = [r for r in wago.rows('SkillLineAbility', build) if r['SkillLine'] in lines
            and not int(misc.get(r['Spell'], {}).get('Attributes_0', 0)) & HIDDEN
            and not (r['AcquireMethod'] == RUNE and int(r['Spell']) in SOD_SPELLS
                     and r['Spell'] not in talent)]
    votes = collections.defaultdict(collections.Counter)
    for r in rows:
        f = family.get(r['Spell'])
        if f in FAMILY:
            votes[r['SkillLine']][FAMILY[f]] += 1

    groups = collections.OrderedDict()
    for r in sorted(rows, key=lambda r: int(r['ID'])):
        if not votes[r['SkillLine']]:
            continue
        cls = votes[r['SkillLine']].most_common(1)[0][0]
        spell = int(r['Spell'])
        name = spelltext.name(spell, build)
        if not name:
            continue
        ranks = groups.setdefault((cls, r['SkillLine'], name), [])
        if spell not in ranks:
            ranks.append(spell)

    out = {c: [] for c in CLASS_ORDER}
    for (cls, line, name), spells in groups.items():
        spells.sort(key=lambda s: (level.get(str(s), 0), s))
        tab = slug(lines[line])
        ranks = [{'t': subtext.get(str(s), ''), 'l': level.get(str(s), 0),
                  'd': spelltext.description(s, build), 'm': meta_list(s, build)} for s in spells]
        top = spells[-1]
        if not any(str(s) in classic for s in spells):
            status = 'new'
        elif classic.get(str(top)) != template.get(str(top)):
            status = 'changed'
        else:
            status = 'unchanged'
        out[cls].append({
            'key': '%s__%s' % (tab, slug(name)),
            'tab': tab,
            'name': name,
            'icon': spelltext.icon(top, build),
            'status': status,
            'fromTalent': any(str(s) in talent for s in spells),
            'desc': ranks[-1]['d'],
            'meta': (['Level %d' % ranks[0]['l']] if ranks[0]['l'] else []) + ranks[-1]['m'],
            'ranks': ranks,
        })
    return out


def main(argv):
    if len(argv) != 2:
        print(__doc__.split('\n\n')[1].strip(), file=sys.stderr)
        return 2
    build = wago.version(argv[1])
    out = read(build)
    with open(OUT, 'w', encoding='utf-8') as f:
        json.dump(out, f, indent=1, ensure_ascii=False)
        f.write('\n')
    m = meta.meta('client beta %s (wago.tools)' % build, 'tools/sources/abilities.py', build=build)
    with open(META, 'w', encoding='utf-8') as f:
        json.dump({'meta': m}, f, indent=2, sort_keys=True)
        f.write('\n')
    counts = collections.Counter(a['status'] for v in out.values() for a in v)
    print('%d abilities from build %s (%s) into %s' % (
        sum(counts.values()), build, ', '.join('%d %s' % (n, s) for s, n in counts.most_common()),
        os.path.relpath(OUT, ROOT)))
    return 0


if __name__ == '__main__':
    raise SystemExit(main(sys.argv))
