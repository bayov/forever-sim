"""usage: AUDIT_WORK=<scratch> python3 -I trait_tree.py <TraitTreeID> <build>...
Every talent-tree diff line (from db2diff.py outputs in $AUDIT_WORK/diffs/*.md) that touches the
given class tree: its nodes, entries, definitions and effect points across the builds given.
Find a class's TraitTreeID in the build's TraitTree table (`tools/sources/wago.py path <build>
TraitTree`) or from a talent spell's TraitDefinition."""
import csv, glob, os, re, sys
W = os.environ['AUDIT_WORK']
sys.path.insert(0, os.path.join(os.path.dirname(os.path.abspath(__file__)), '../../sources'))
import wago  # noqa: E402
tree, builds = sys.argv[1], sys.argv[2:]
def rows(t, b):
    try:
        return wago.rows(t, b)
    except OSError:
        return []
nodes, entries, defs, ep = set(), set(), set(), set()
for b in builds:
    nodes |= {r['ID'] for r in rows('TraitNode', b) if r['TraitTreeID'] == tree}
    entries |= {r['TraitNodeEntryID'] for r in rows('TraitNodeXTraitNodeEntry', b) if r['TraitNodeID'] in nodes}
    defs |= {r['TraitDefinitionID'] for r in rows('TraitNodeEntry', b) if r['ID'] in entries}
    ep |= {r['ID'] for r in rows('TraitDefinitionEffectPoints', b) if r['TraitDefinitionID'] in defs}
print('nodes', len(nodes), 'entries', len(entries), 'definitions', len(defs), 'effect points', len(ep))
own = {'TraitNode': nodes, 'TraitNodeEntry': entries, 'TraitDefinition': defs, 'TraitDefinitionEffectPoints': ep}
for f in sorted(glob.glob(os.path.join(W, 'diffs', '*.md'))):
    for line in open(f, encoding='utf-8'):
        m = re.match(r'- (\w+) row (\d+)', line)
        if not m: continue
        t, i = m.groups()
        if (t in own and i in own[t]) or (t in ('TraitNodeXTraitNodeEntry', 'TraitEdge') and any(f"'{n}'" in line for n in nodes)):
            print(os.path.basename(f), line.strip()[:250])
