"""Diff wago DB2 CSV dumps between consecutive builds, grouped by class and spell.

usage: python3 -I db2diff.py <db2dir> <oldbuild> <newbuild> <outfile>
  e.g. python3 -I db2diff.py $AUDIT_WORK/db2 70425 70450 $AUDIT_WORK/diffs/70425-70450.md
Covers the spell-keyed, talent and lookup tables (SpellItemEnchantment and ChrRaces among the
lookups); ItemSparse has its own script, items_diff.py.
"""
import csv, sys, os, collections

D, OLD, NEW, OUT = sys.argv[1], sys.argv[2], sys.argv[3], sys.argv[4]
csv.field_size_limit(10**9)

def load(t, b):
    p = os.path.join(D, f"{t}.{b}.csv")
    if not os.path.exists(p):
        return None, None
    with open(p, newline='', encoding='utf-8') as f:
        r = csv.reader(f)
        hdr = next(r)
        rows = {}
        idc = hdr.index('ID')
        for row in r:
            rows[row[idc]] = row
    return hdr, rows

CLASSSET = {'3': 'Mage', '4': 'Warrior', '5': 'Warlock', '6': 'Priest', '7': 'Druid', '8': 'Rogue', '9': 'Hunter', '10': 'Paladin', '11': 'Shaman'}
MASKBIT = {1: 'Warrior', 2: 'Paladin', 4: 'Hunter', 8: 'Rogue', 16: 'Priest', 64: 'Shaman', 128: 'Mage', 256: 'Warlock', 1024: 'Druid'}

# Name and class of every spell, from the new build (old build as fallback).
names = {}
for b in (OLD, NEW):
    h, rows = load('SpellName', b)
    if rows:
        ni = h.index('Name_lang')
        for k, r in rows.items():
            names[k] = r[ni]
spellclass = collections.defaultdict(set)
for b in (OLD, NEW):
    h, rows = load('SpellClassOptions', b)
    if rows:
        si, ci = h.index('SpellID'), h.index('SpellClassSet')
        for r in rows.values():
            if r[ci] in CLASSSET:
                spellclass[r[si]].add(CLASSSET[r[ci]])
    h, rows = load('SkillLineAbility', b)
    if rows:
        si, mi, li = h.index('Spell'), h.index('ClassMask'), h.index('SkillLine')
        for r in rows.values():
            m = int(r[mi] or 0)
            for bit, c in MASKBIT.items():
                if m & bit and m != 1535 and m & 1535 != 1535:
                    spellclass[r[si]].add(c)

# Talent spells: TraitDefinition.SpellID -> node -> tree; trees get a class by majority vote.
tdef_spell, tdef_of_entry, entry_node, node_tree = {}, {}, {}, {}
for b in (OLD, NEW):
    h, rows = load('TraitDefinition', b)
    if rows:
        si = h.index('SpellID')
        for k, r in rows.items():
            tdef_spell[k] = r[si]
    h, rows = load('TraitNodeEntry', b)
    if rows:
        di = h.index('TraitDefinitionID')
        for k, r in rows.items():
            tdef_of_entry[k] = r[di]
    h, rows = load('TraitNodeXTraitNodeEntry', b)
    if rows:
        ni, ei = h.index('TraitNodeID'), h.index('TraitNodeEntryID')
        for r in rows.values():
            entry_node[r[ei]] = r[ni]
    h, rows = load('TraitNode', b)
    if rows:
        ti = h.index('TraitTreeID')
        for k, r in rows.items():
            node_tree[k] = r[ti]
tree_votes = collections.defaultdict(collections.Counter)
tdef_tree = {}
for e, td in tdef_of_entry.items():
    n = entry_node.get(e)
    t = node_tree.get(n)
    if t:
        tdef_tree[td] = t
        for c in spellclass.get(tdef_spell.get(td, ''), ()):
            tree_votes[t][c] += 1
tree_class = {t: v.most_common(1)[0][0] for t, v in tree_votes.items() if v}
talent_spells = set()
for td, t in tdef_tree.items():
    s = tdef_spell.get(td)
    if s and s != '0':
        talent_spells.add(s)
        if t in tree_class:
            spellclass[s].add(tree_class[t] + ' (talent)')

# Racial spells (SkillLineAbility with a race mask and no class) and item spells.
for b in (OLD, NEW):
    h, rows = load('SkillLineAbility', b)
    if rows:
        si, mi, ri = h.index('Spell'), h.index('ClassMask'), h.index('RaceMasks_0')
        li = h.index('SkillLine')
        for r in rows.values():
            if int(r[mi] or 0) == 0 and int(r[ri] or 0) != 0 and r[si] not in spellclass:
                spellclass[r[si]].add('Racial')
    h, rows = load('ItemEffect', b)
    if rows:
        si = h.index('SpellID')
        for r in rows.values():
            if r[si] not in spellclass:
                spellclass[r[si]].add('Item')

# Triggered spells inherit their parent's class; spells sharing a name with a class spell inherit it too.
trig = collections.defaultdict(set)
for b in (OLD, NEW):
    h, rows = load('SpellEffect', b)
    if rows:
        si, ti = h.index('SpellID'), h.index('EffectTriggerSpell')
        for r in rows.values():
            if r[ti] not in ('0', ''):
                trig[r[si]].add(r[ti])
for _ in range(3):
    for parent, kids in trig.items():
        pc = {c.replace(' (talent)', '') for c in spellclass.get(parent, ())}
        if pc:
            for k in kids:
                if k not in spellclass:
                    spellclass[k] = set(pc)
byname = collections.defaultdict(set)
for s, c in list(spellclass.items()):
    if s in names:
        byname[names[s]] |= {x.replace(' (talent)', '') for x in c}
for s, n in names.items():
    if s not in spellclass and n in byname and len(byname[n]) == 1:
        spellclass[s] = set(byname[n])

def cls(spell):
    c = spellclass.get(spell)
    if c:
        c = {x.replace(' (talent)', '') for x in c}
        return ', '.join(sorted(c))
    return ''

def _unused(spell):
    c = spellclass.get(spell)
    return ', '.join(sorted(c)) if c else ''

SPELLKEYED = ['Spell', 'SpellName', 'SpellEffect', 'SpellAuraOptions', 'SpellCooldowns', 'SpellCategories',
              'SpellEquippedItems', 'SpellLevels', 'SpellMisc', 'SpellPower', 'SpellClassOptions',
              'SpellAuraRestrictions', 'SpellTargetRestrictions', 'SpellInterrupts', 'SpellShapeshift',
              'SpellReagents', 'SpellScaling', 'SpellLabel']
LOOKUP = ['SpellDuration', 'SpellCastTimes', 'SpellCategory', 'SpellProcsPerMinute', 'SpellRadius', 'SpellRange',
          'SpellItemEnchantment', 'ChrRaces', 'TraitNode', 'TraitNodeEntry', 'TraitCond', 'TraitEdge',
          'TraitNodeGroup', 'TraitTree', 'TraitNodeXTraitNodeEntry', 'ItemEffect']
SKIPCOLS = {'SpellXSpellVisual', 'SpellVisualScript', 'ActiveSpellVisualScript', 'SpellIconFileDataID', 'ActiveIconFileDataID'}

by_spell = collections.defaultdict(list)
other = []

def rowdiff(hdr_o, ro, hdr_n, rn):
    out = []
    od = dict(zip(hdr_o, ro))
    nd = dict(zip(hdr_n, rn))
    for c in hdr_n:
        if c in SKIPCOLS:
            continue
        if od.get(c) != nd.get(c):
            out.append(f"{c}: {od.get(c)} -> {nd.get(c)}")
    return out

def difftable(t):
    ho, ro = load(t, OLD)
    hn, rn = load(t, NEW)
    if ro is None or rn is None:
        other.append(f"[{t}] missing in one build")
        return []
    res = []
    for k in sorted(set(ro) | set(rn), key=lambda x: int(x) if x.lstrip('-').isdigit() else 0):
        a, b = ro.get(k), rn.get(k)
        if a == b:
            continue
        if a is None:
            res.append((k, 'ADDED', dict(zip(hn, b))))
        elif b is None:
            res.append((k, 'REMOVED', dict(zip(ho, a))))
        else:
            d = rowdiff(ho, a, hn, b)
            if d:
                res.append((k, 'CHANGED', d, dict(zip(hn, b))))
    return res

for t in SPELLKEYED:
    for item in difftable(t):
        k, kind = item[0], item[1]
        row = item[-1]
        spell = k if t in ('Spell', 'SpellName') else row.get('SpellID', '?')
        if kind == 'CHANGED':
            extra = ''
            if t == 'SpellEffect':
                extra = f" (EffectIndex {row.get('EffectIndex')}, Effect {row.get('Effect')}, Aura {row.get('EffectAura')})"
            by_spell[spell].append(f"{t} row {k}{extra}: " + '; '.join(item[2]))
        else:
            keep = {c: v for c, v in row.items() if v not in ('0', '', '0.0') and c not in SKIPCOLS}
            if t in ('Spell',):
                keep = {c: v[:300] for c, v in keep.items()}
            by_spell[spell].append(f"{t} row {k} {kind}: {keep}")

# Trait effect points belong to a talent's spell.
for item in difftable('TraitDefinitionEffectPoints'):
    k, kind, row = item[0], item[1], item[-1]
    spell = tdef_spell.get(row.get('TraitDefinitionID'), '?')
    msg = '; '.join(item[2]) if kind == 'CHANGED' else str(row)
    by_spell[spell].append(f"TraitDefinitionEffectPoints row {k} {kind}: {msg}")
for item in difftable('TraitDefinition'):
    k, kind, row = item[0], item[1], item[-1]
    spell = row.get('SpellID', '?')
    msg = '; '.join(item[2]) if kind == 'CHANGED' else str({c: v for c, v in row.items() if v not in ('0', '')})
    by_spell[spell].append(f"TraitDefinition row {k} {kind}: {msg}")

for t in LOOKUP:
    for item in difftable(t):
        k, kind, row = item[0], item[1], item[-1]
        msg = '; '.join(item[2]) if kind == 'CHANGED' else str({c: v for c, v in row.items() if v not in ('0', '')})[:400]
        other.append(f"[{t}] row {k} {kind}: {msg}")

groups = collections.defaultdict(list)
for spell, lines in by_spell.items():
    c = cls(spell) or 'No class'
    groups[c].append((spell, lines))

with open(OUT, 'w') as f:
    f.write(f"# DB2 diff 1.60.1.{OLD} -> 1.60.1.{NEW}\n\n")
    f.write(f"Spells changed: {len(by_spell)}\n\n")
    import re
    for c in sorted(groups, key=lambda x: (x == 'No class', x)):
        f.write(f"\n## {c} ({len(groups[c])} spells)\n")
        bc_only = [s for s, lines in groups[c]
                   if all(re.search(r': EffectBonusCoefficient: 1 -> 0$', l) for l in lines)]
        if len(bc_only) > 5:
            f.write(f"\n### {len(bc_only)} spells whose only change is EffectBonusCoefficient 1 -> 0\n")
            f.write('- ' + '; '.join(f"{names.get(s, '?')} ({s})" for s in sorted(bc_only, key=lambda s: names.get(s, ''))) + '\n')
        else:
            bc_only = []
        for spell, lines in sorted(groups[c], key=lambda x: names.get(x[0], '')):
            if spell in bc_only:
                continue
            f.write(f"\n### {names.get(spell, '?')} ({spell})\n")
            for l in lines:
                f.write(f"- {l}\n")
    f.write("\n## Lookup and talent-tree tables\n")
    for l in other:
        f.write(f"- {l}\n")
print(OUT, len(by_spell), 'spells', len(other), 'lookup rows')
