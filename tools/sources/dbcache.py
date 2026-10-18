#!/usr/bin/env python3
"""Items our beta client was sent, read out of its hotfix cache.

    python3 tools/sources/dbcache.py <client folder> [--write]

The client keeps every database row the server sent it, on top of the tables
it shipped with, in Cache/ADB/enUS/DBCache.bin. Forever's new items reach us
that way. Some of them aren't in the client build's tables at all, and some
are there only as "not public" rows, which is how Forever keeps an item
secret until someone finds it. So this file knows items no table on
wago.tools has. It only knows what our own client was sent, though, which is
why it's one of several sources for the secret items (docs/data_sources.md).

With --write it writes assets/db_inputs/forever_client_cache.json, with a
meta block naming the build. Without --write it prints what it found. It runs
locally only, since the cloud has no client.

Each item there has its state and, for a row with data, its name, quality,
item level, required level, slot, set, class and subclass, and its stats as
the row stores them: [stat, share] pairs, where the share is ten thousandths
of the item's stat budget. That budget is RandPropPoints for the item level,
quality and slot, which gives Classic's items exactly (2,809 of them checked
against db.json at 70450). Forever's newer items don't follow it (Abomination
Bones has Strength 9 in game where the budget gives 8, and stat types like 54
that Classic never had), so nothing here turns shares into stats. What the
file is good for is which items exist and which the server still hides:
tools/sources/unknown.py reads it for the secret items, and stale.py for
reveals in our client.

The file is a header (the magic "XFTH", a version, the build) and then one
entry per row:

    "XFTH", int32 region, int32 push ID, uint32 unique ID, uint32 table hash,
    int32 record ID, uint32 data size, uint8 status, 3 pad bytes, data

Status 1 is a row with data, 2 a row the server removed, 3 a row it says is
invalid, and 4 one that isn't public yet. We read ItemSparse (table hash
0x919BE54E) and Item (0x50238EC2) rows. A row's data is the table's fields
in order, with strings inline and null-terminated, and without the ID, which
is the record ID. The field layouts are WoWDBDefs' for this client (ITEM_SPARSE
and ITEM below).

A layout is only right for the builds it was written for, so before we trust
a decode we check it against the rows the cache shares with wago.tools' copy
of the same build's table. A wrong layout shifts every field after the first
wrong one, so most fields of most rows differ, and then we stop rather than
write wrong stats. A row that differs in a field or two is a real difference:
the cache keeps rows from earlier pushes (Crystal Thread Cloak's item level 29
where build 70450 ships 30), so for an item the shipped table has, the shipped
table wins, and gen_db only takes the cache's items the shipped table lacks.
"""
import argparse
import collections
import json
import os
import struct
import sys

sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
import meta  # noqa: E402
import wago  # noqa: E402

ROOT = os.path.abspath(os.path.join(os.path.dirname(os.path.abspath(__file__)), '..', '..'))
OUT = os.path.join(ROOT, 'assets', 'db_inputs', 'forever_client_cache.json')
CACHE = os.path.join('Cache', 'ADB', 'enUS', 'DBCache.bin')

MAGIC = b'XFTH'
VALID, REMOVED, INVALID, NOT_PUBLIC = 1, 2, 3, 4
STATES = {VALID: 'valid', REMOVED: 'removed', INVALID: 'invalid', NOT_PUBLIC: 'notPublic'}
ITEM_SPARSE_HASH, ITEM_HASH = 0x919BE54E, 0x50238EC2
# A row that differs from wago.tools in this many fields, or this share of
# rows differing at all, means the layout is wrong.
LAYOUT_BROKEN_FIELDS, LAYOUT_BROKEN_SHARE = 3, 0.2

# source: WoWDBDefs ItemSparse.dbd layout 6FCC3191 and Item.dbd layout 9A2A4834, client beta 1.60.1.70450
# Each field: (name, struct code, count). 's' is an inline string. A name
# matches wago.tools' column, with _0, _1 and so on for an array.
ITEM_SPARSE = [
    ('Description_lang', 's', 1), ('Display3_lang', 's', 1), ('Display2_lang', 's', 1),
    ('Display1_lang', 's', 1), ('Display_lang', 's', 1),
    ('ExpansionID', 'i', 1), ('DmgVariance', 'f', 1), ('LimitCategory', 'i', 1),
    ('DurationInInventory', 'I', 1), ('QualityModifier', 'f', 1), ('BagFamily', 'I', 1),
    ('StartQuestID', 'i', 1), ('LanguageID', 'i', 1), ('ItemRange', 'f', 1),
    ('StatPercentageOfSocket', 'f', 10), ('StatPercentEditor', 'i', 10),
    ('StatModifier_bonusStat', 'i', 10), ('Stackable', 'i', 1), ('MaxCount', 'i', 1),
    ('MinReputation', 'i', 1), ('RequiredAbility', 'I', 1), ('AllowableRace', 'i', 2),
    ('SellPrice', 'I', 1), ('BuyPrice', 'I', 1), ('VendorStackCount', 'I', 1),
    ('PriceVariance', 'f', 1), ('PriceRandomValue', 'f', 1), ('Flags', 'i', 5),
    ('OppositeFactionItemID', 'i', 1), ('ModifiedCraftingReagentItemID', 'i', 1),
    ('ContentTuningID', 'i', 1), ('PlayerLevelToItemLevelCurveID', 'i', 1),
    ('ItemLevelOffsetCurveID', 'i', 1), ('ItemLevelOffsetItemLevel', 'i', 1),
    ('ItemSquishEraID', 'i', 1), ('ItemNameDescriptionID', 'H', 1),
    ('RequiredTransmogHoliday', 'H', 1), ('RequiredHoliday', 'H', 1), ('Gem_properties', 'H', 1),
    ('Socket_match_enchantment_ID', 'H', 1), ('TotemCategoryID', 'H', 1),
    ('InstanceBound', 'H', 1), ('ZoneBound', 'H', 2), ('ItemSet', 'H', 1), ('LockID', 'H', 1),
    ('PageID', 'H', 1), ('ItemDelay', 'H', 1), ('MinFactionID', 'H', 1),
    ('RequiredSkillRank', 'H', 1), ('RequiredSkill', 'H', 1), ('ItemLevel', 'H', 1),
    ('AllowableClass', 'h', 1), ('ArtifactID', 'B', 1), ('SpellWeight', 'B', 1),
    ('SpellWeightCategory', 'B', 1), ('SocketType', 'B', 3), ('SheatheType', 'B', 1),
    ('Material', 'B', 1), ('PageMaterialID', 'B', 1), ('Bonding', 'B', 1), ('DamageType', 'B', 1),
    ('ContainerSlots', 'B', 1), ('RequiredPVPMedal', 'B', 1), ('RequiredPVPRank', 'b', 1),
    ('RequiredLevel', 'b', 1), ('InventoryType', 'b', 1), ('OverallQualityID', 'b', 1),
    ('AmmunitionType', 'B', 1),
]
ITEM = [
    ('ClassID', 'i', 1), ('SubclassID', 'B', 1), ('Material', 'B', 1), ('InventoryType', 'b', 1),
    ('SheatheType', 'B', 1), ('ItemPetFoodID', 'i', 1), ('Sound_override_subclassID', 'b', 1),
    ('IconFileDataID', 'i', 1), ('ItemGroupSoundsID', 'I', 1), ('ContentTuningID', 'i', 1),
    ('ModifiedCraftingReagentItemID', 'i', 1), ('AmmunitionType', 'B', 1),
    ('CraftingQualityID', 'i', 1), ('ItemSquishEraID', 'i', 1),
    ('RecraftReagentCountPercentage', 'f', 1), ('OrderSource', 'B', 1),
]


def entries(path):
    """(build, [(table hash, record ID, status, data)]) of a DBCache.bin."""
    with open(path, 'rb') as f:
        d = f.read()
    if d[:4] != MAGIC:
        raise ValueError('%s is not a hotfix cache' % path)
    _version, build = struct.unpack_from('<ii', d, 4)
    out = []
    i = d.find(MAGIC, 4)
    while i != -1 and i + 32 <= len(d):
        _region, _push, _unique, table, record, size = struct.unpack_from('<iiIIiI', d, i + 4)
        status = d[i + 28]
        out.append((table, record, status, d[i + 32:i + 32 + size]))
        i = d.find(MAGIC, i + 32 + size)
    return build, out


def decode(data, layout):
    """A row's data as {column: value}, in wago.tools' column names."""
    row, at = {}, 0
    for name, code, count in layout:
        for k in range(count):
            column = name if count == 1 else '%s_%d' % (name, k)
            if code == 's':
                end = data.index(b'\0', at)
                row[column] = data[at:end].decode('utf-8', 'replace')
                at = end + 1
            else:
                (row[column],) = struct.unpack_from('<' + code, data, at)
                at += struct.calcsize(code)
    if at != len(data):
        raise ValueError('row is %d bytes, the layout reads %d' % (len(data), at))
    return row


def same(a, b):
    if isinstance(a, float):
        return abs(a - float(b or 0)) < 1e-4 * max(1, abs(a))
    if isinstance(a, int):
        return a == int(float(b or 0))
    return a == (b or '')


def check(rows, table, build):
    """How many decoded rows we compared with wago.tools, and the ones that differ."""
    try:
        theirs = wago.by_id(table, build)
    except OSError:
        theirs = None
    if not theirs:
        return 0, []
    compared, bad = 0, []
    for rid, row in rows.items():
        t = theirs.get(str(rid))
        if t is None:
            continue
        compared += 1
        diff = [k for k, v in row.items() if k in t and not same(v, t[k])]
        if diff:
            bad.append((rid, diff[:5]))
    return compared, bad


def read(client):
    """{build, items: {id: entry}, states: Counter} from a client folder's cache."""
    build, rows = entries(os.path.join(client, CACHE))
    version = wago.version(str(build))
    sparse, item, state = {}, {}, {}
    for table, rid, status, data in rows:
        if table not in (ITEM_SPARSE_HASH, ITEM_HASH):
            continue
        if status == VALID and data:
            layout = ITEM_SPARSE if table == ITEM_SPARSE_HASH else ITEM
            (sparse if table == ITEM_SPARSE_HASH else item)[rid] = decode(data, layout)
        if table == ITEM_SPARSE_HASH:
            # A later entry for the same row wins, as in the client.
            state[rid] = status
    for table, rows_ in (('ItemSparse', sparse), ('Item', item)):
        compared, bad = check(rows_, table, version)
        broken = [b for b in bad if len(b[1]) >= LAYOUT_BROKEN_FIELDS]
        if broken or (compared and len(bad) > compared * LAYOUT_BROKEN_SHARE):
            raise SystemExit('%s rows decode differently from wago.tools for build %s, so the '
                             'layout changed (%d of %d compared, first: %s). Update ITEM_SPARSE '
                             'or ITEM from WoWDBDefs.' % (table, version, len(bad), compared, bad[:3]))
        print('%s: %d rows with data, %d of them checked against wago.tools %s, %d with a field '
              'or two that differ (older pushes): %s' % (
                  table, len(rows_), compared, version, len(bad), ', '.join(
                      '%d %s' % (rid, '/'.join(f)) for rid, f in bad[:5]) or '-'))
    items = {}
    for rid, status in sorted(state.items()):
        entry = {'state': STATES.get(status, str(status))}
        sp = sparse.get(rid) if status == VALID else None
        it = item.get(rid)
        if sp:
            entry.update(name=sp['Display_lang'], quality=sp['OverallQualityID'], ilvl=sp['ItemLevel'],
                         requiredLevel=sp['RequiredLevel'], inventoryType=sp['InventoryType'],
                         set=sp['ItemSet'], delay=sp['ItemDelay'],
                         statShares=[[sp['StatModifier_bonusStat_%d' % k], sp['StatPercentEditor_%d' % k]]
                                     for k in range(10) if sp['StatModifier_bonusStat_%d' % k] > 0])
        if it:
            entry.update(classID=it['ClassID'], subclassID=it['SubclassID'])
        items[rid] = {k: v for k, v in entry.items() if v not in (0, '', [], None) or k == 'state'}
    return version, items


def main():
    ap = argparse.ArgumentParser(description=__doc__.split('\n\n')[0])
    ap.add_argument('client', help='the client folder, like ".../World of Warcraft/_classic_beta_"')
    ap.add_argument('--write', action='store_true')
    a = ap.parse_args()
    version, items = read(a.client)
    shipped = wago.by_id('ItemSparse', version) or {}
    states = collections.Counter(e['state'] for e in items.values())
    new = [i for i, e in items.items() if 'name' in e and str(i) not in shipped]
    print('build %s: %d items (%s), %d with data that the shipped ItemSparse lacks' % (
        version, len(items), ', '.join('%d %s' % (n, s) for s, n in states.most_common()), len(new)))
    if a.write:
        m = meta.meta('client beta %s hotfix cache (%s)' % (version, CACHE.replace(os.sep, '/')),
                      'tools/sources/dbcache.py', build=version)
        with open(OUT, 'w', encoding='utf-8') as f:
            f.write('{\n' + meta.line(m) + '"items": {\n')
            f.write(',\n'.join('%s: %s' % (json.dumps(str(i)), json.dumps(e, sort_keys=True))
                               for i, e in sorted(items.items())))
            f.write('\n}}\n')
        print('wrote %s' % os.path.relpath(OUT, ROOT))
    return 0


if __name__ == '__main__':
    raise SystemExit(main())
