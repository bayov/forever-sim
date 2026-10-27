#!/usr/bin/env python3
"""The Forever racials, read out of the client's tables.

    python3 tools/sources/racials.py <build> [--write]

It writes ui/core/forever/racials.json, which the sim's UI (ui/core/forever/
racials.ts) and the DPS report read, and racials.meta.json next to it with
the build. Without --write it only prints what would change. The racials used
to be typed in from hyjal.cc at build 69876. Now a new client build is one run
of this script.

What we read:

- Which spells are racials: SkillLineAbility rows on a racial skill line
  (SkillLine category 9 with "Racial" in its name, which leaves out the
  riding skills in the same category), for the race in RaceMasks. The races
  past bit 31 (the two Skyborne races, 32 and 33) are in RaceMasks_1. Spells
  the client hides from the spellbook (Attributes_0 bit 0x80), like Quickness
  Passive, are helpers and not listed.
- Which classes get which: ClassMask, or every class the race can be
  (CharBaseInfo) when it is 0 or -1. A racial with class versions, like
  Eureka! or Touch of the Grave, gets one variant per version, and a version
  for every class takes the classes no other version claims.
- The text, icon, cost line and whether it's passive: tools/sources/
  spelltext.py, the same reader as the talent importer.

The order of the racials in each race is kept from the current file, so a
regeneration only shows real changes. A racial the file didn't have goes last.
Where the client has no number for a tooltip (Berserking's duration), the
current file's text stays (keep_known).

A server change the client doesn't have yet (the 2026-10-26 racial pass) is
carried by hand in tools/sources/racials_hotfix.json and applied last
(hotfix): it rewrites a racial's text, or swaps the racial for a new one in
its place, and marks the variant with source 'hotfix' and a note citing the
change. Once a client build carries it, the script says so and the entry
can go.
"""
import argparse
import collections
import json
import os
import sys

sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
import meta  # noqa: E402
import spelltext  # noqa: E402
import wago  # noqa: E402

ROOT = os.path.abspath(os.path.join(os.path.dirname(os.path.abspath(__file__)), '..', '..'))
OUT = os.path.join(ROOT, 'ui', 'core', 'forever', 'racials.json')
META = os.path.join(ROOT, 'ui', 'core', 'forever', 'racials.meta.json')
HOTFIX = os.path.join(ROOT, 'tools', 'sources', 'racials_hotfix.json')

# source: ChrRaces and ChrClasses IDs, client beta 1.60.1.70450
# The sim's names for the client's playable races and classes (proto/common.proto).
RACES = {1: 'RaceHuman', 2: 'RaceOrc', 3: 'RaceDwarf', 4: 'RaceNightElf', 5: 'RaceUndead',
         6: 'RaceTauren', 7: 'RaceGnome', 8: 'RaceTroll', 95: 'RaceSkyborneHighOrder',
         96: 'RaceSkyborneWindshaper'}
CLASSES = {1: 'ClassWarrior', 2: 'ClassPaladin', 3: 'ClassHunter', 4: 'ClassRogue', 5: 'ClassPriest',
           7: 'ClassShaman', 8: 'ClassMage', 9: 'ClassWarlock', 11: 'ClassDruid'}
# The order the classes are listed in, which is the order the hand-typed file had.
CLASS_ORDER = [1, 3, 8, 4, 5, 9, 2, 11, 7]
RACIAL_CATEGORY = '9'
HIDDEN = 0x80


def race_bits(build):
    """{race ID: its bit in RaceMasks}, from ChrRaces.PlayableRaceBit."""
    return {int(r['ID']): int(r['PlayableRaceBit']) for r in wago.rows('ChrRaces', build)
            if int(r['ID']) in RACES}


def has_race(row, bit):
    lo, hi = int(row['RaceMasks_0']), int(row['RaceMasks_1'])
    return bool((lo if bit < 32 else hi) >> (bit % 32) & 1)


def read(build):
    """{race name: [racial]} in the file's shape, in skill line order."""
    lines = {r['ID'] for r in wago.rows('SkillLine', build)
             if r['CategoryID'] == RACIAL_CATEGORY and 'Racial' in r['DisplayName_lang']}
    misc = {r['SpellID']: r for r in wago.rows('SpellMisc', build)}
    classes = collections.defaultdict(list)
    for r in wago.rows('CharBaseInfo', build):
        if int(r['RaceID']) in RACES and int(r['ClassID']) in CLASSES:
            classes[int(r['RaceID'])].append(int(r['ClassID']))
    bits = race_bits(build)
    rows = [r for r in wago.rows('SkillLineAbility', build) if r['SkillLine'] in lines
            and not int(misc.get(r['Spell'], {}).get('Attributes_0', 0)) & HIDDEN]

    out = {}
    for race, bit in bits.items():
        mine = sorted((r for r in rows if has_race(r, bit)), key=lambda r: int(r['ID']))
        can = sorted(classes[race], key=CLASS_ORDER.index)
        by_name = collections.OrderedDict()
        for r in mine:
            spell = int(r['Spell'])
            by_name.setdefault(spelltext.name(spell, build), []).append((spell, int(r['ClassMask'])))
        racials = []
        for name, versions in by_name.items():
            claimed = {c for _, mask in versions if mask not in (0, -1) for c in can if mask >> (c - 1) & 1}
            variants = []
            for spell, mask in versions:
                if mask in (0, -1):
                    got = [c for c in can if c not in claimed]
                else:
                    got = [c for c in can if mask >> (c - 1) & 1]
                if not got:
                    continue
                passive = spelltext.passive(spell, build)
                variants.append({
                    'spellId': spell,
                    'classes': [CLASSES[c] for c in got],
                    'description': spelltext.description(spell, build),
                    'icon': spelltext.icon(spell, build),
                    'passive': passive,
                    'cost': None if passive else (spelltext.cost(spell, build) or None),
                })
            if variants:
                racials.append({'name': name, 'variants': variants})
        out[RACES[race]] = racials
    return out


def ordered(new, old):
    """new, with races and racials in old's order where old has them.

    A racial is found in old by any of its spell IDs, so a renamed one (Sword
    Specialization became Ruthless) keeps its place.
    """
    def rank(keys, key):
        return keys.index(key) if key in keys else len(keys)
    races = sorted(new, key=lambda r: rank(list(old), r))
    out = {}
    for race in races:
        # A hotfixed racial keeps the place of the client racial it replaces.
        spots = {k: i for i, x in enumerate(old.get(race, [])) for v in x['variants']
                 for k in (v['spellId'], v.get('replaces')) if k is not None}
        out[race] = sorted(new[race], key=lambda x: min(
            (spots[v['spellId']] for v in x['variants'] if v['spellId'] in spots), default=len(spots)))
    return out


def keep_known(new, old):
    """Keep old's text for a variant whose new text has a value the client lacks.

    spelltext leaves a ? where the client tables have no number. Berserking's
    duration is one: its tooltip reads a spell with no duration, and the old
    file has Classic's 10 sec. We'd rather show that than a ?, and say so.
    """
    was = {v['spellId']: v for x in old.values() for r in x for v in r['variants']}
    for race, racials in new.items():
        for r in racials:
            for v in r['variants']:
                o = was.get(v['spellId'])
                if '?' in v['description'] and o and '?' not in o['description']:
                    print('%s: %s (%d) has a ? in the client, kept the old text: %s' % (
                        race, r['name'], v['spellId'], o['description']))
                    v['description'] = o['description']


def hotfix(new):
    """Apply tools/sources/racials_hotfix.json to new, in place."""
    with open(HOTFIX, encoding='utf-8') as f:
        entries = json.load(f)['entries']
    for e in entries:
        racials = new.get(e['race'], [])
        at = next(((i, v) for i, x in enumerate(racials) for v in x['variants'] if v['spellId'] == e['spellId']), None)
        if at is None:
            print('hotfix: %s has no racial %d in the client, entry not applied' % (e['race'], e['spellId']))
            continue
        i, v = at
        if 'replace' in e:
            r = e['replace']
            racials[i] = {'name': r['name'], 'variants': [{
                'spellId': r['spellId'], 'classes': v['classes'], 'description': r['description'],
                'icon': r['icon'], 'passive': v['passive'], 'cost': v['cost'],
                'source': 'hotfix', 'replaces': e['spellId'], 'note': e['note']}]}
            continue
        if v['description'] == e['description']:
            print('hotfix: %s %s (%d) already reads so in the client, the entry can go' % (
                e['race'], racials[i]['name'], e['spellId']))
            continue
        v['description'] = e['description']
        v['source'] = 'hotfix'
        v['note'] = e['note']


def changes(old, new):
    lines = []
    for race in sorted(set(old) | set(new)):
        o = {(x['name'], v['spellId']): v for x in old.get(race, []) for v in x['variants']}
        n = {(x['name'], v['spellId']): v for x in new.get(race, []) for v in x['variants']}
        for k in sorted(set(o) - set(n)):
            lines.append('%s: removed %s (%d)' % (race, k[0], k[1]))
        for k in sorted(set(n) - set(o)):
            lines.append('%s: added %s (%d)' % (race, k[0], k[1]))
        for k in sorted(set(o) & set(n)):
            for field in ('classes', 'description', 'icon', 'passive', 'cost', 'source'):
                if o[k].get(field) != n[k].get(field):
                    lines.append('%s: %s (%d) %s\n    was: %s\n    now: %s' % (
                        race, k[0], k[1], field, o[k].get(field), n[k].get(field)))
    return lines


def main():
    ap = argparse.ArgumentParser(description=__doc__.split('\n\n')[0])
    ap.add_argument('build')
    ap.add_argument('--write', action='store_true')
    a = ap.parse_args()
    build = wago.version(a.build)
    with open(OUT, encoding='utf-8') as f:
        old = json.load(f)
    new = ordered(read(build), old)
    keep_known(new, old)
    hotfix(new)
    for line in changes(old, new):
        print(line)
    if a.write:
        with open(OUT, 'w', encoding='utf-8') as f:
            json.dump(new, f, indent=2, ensure_ascii=False)
            f.write('\n')
        m = meta.meta('client beta %s (wago.tools), with tools/sources/racials_hotfix.json' % build,
                      'tools/sources/racials.py', build=build)
        with open(META, 'w', encoding='utf-8') as f:
            json.dump({'meta': m}, f, indent=2, sort_keys=True)
            f.write('\n')
        print('wrote %s' % os.path.relpath(OUT, ROOT))
    return 0


if __name__ == '__main__':
    raise SystemExit(main())
