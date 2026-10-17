#!/usr/bin/env python3

"""Generates client/<class>.json from the WoW Forever beta client's own game tables.

	tools/forever_talents/import_db2.py <build> [class ...] [--write]

Without --write it prints, per class, how the generated tree differs from the current
client/<class>.json: talents added or removed, moved, renamed or given another rank count
or prerequisite (structure), ranks whose numbers differ (numbers), and ranks whose text
differs with the same numbers (wording). With --write it rewrites the files.

The client/ files used to be put together by hand from hyjal.cc, foreverchanges.pro and
wago.tools tables at several builds. This reads all of it from the client's DB2 tables at
one build instead, through tools/sources/wago.py (the shared cache) and
tools/sources/spelltext.py (tooltips, cost lines, icons), and keeps the same JSON shape, so
import_client.py and import_talents.py read the result unchanged.

Forever's talents are trait trees, one TraitTree per class. Each talent is a TraitNode with
a position, and its TraitNodeEntry gives the rank count and the TraitDefinition, which names
the spell. Talents whose ranks scale one spell have a TraitDefinitionEffectPoints row per
effect, pointing to a curve in CurvePoint from rank to value, and we fill the spell's tooltip
once per rank with those values. TraitEdge gives the prerequisites.

The three talent trees of a class sit side by side in the one trait tree, so we find them by
grouping the node positions: the node columns fall into three groups far apart, and inside a
group the columns and rows are a fixed step apart. The leftmost group is TalentTab's
OrderIndex 0, which gives the background (bg is the tab's ID) and the icon. The tree's name is
the skill line most of its talent spells belong to (SkillLineAbility), since that is what the
files have always said ("Elemental Combat", "Shadow Magic") where TalentTab says "Elemental"
and "Shadow". A few nodes are parked far outside the grid, which is how the client hides a
talent it replaced, and we leave those out.

Two things the tables can't know are carried over from the current file: each tree's
`removed` list (talents that were in the tree in earlier builds), and the hand notes in
`tooltip.notes` of talents that still exist, minus the provenance notes this replaces. The
top-level "source" says so.
"""

import collections
import json
import os
import re
import sys

HERE = os.path.dirname(os.path.abspath(__file__))
sys.path.insert(0, os.path.join(HERE, '..', 'sources'))
import spelltext  # noqa: E402
import wago  # noqa: E402

CLIENT_DIR = os.path.join(HERE, 'client')
CLASSES = ['druid', 'hunter', 'mage', 'paladin', 'priest', 'rogue', 'shaman', 'warlock', 'warrior']

# The notes spelltext writes itself, and the one hyjal.cc wrote for a "?". A note in the
# current file that is one of these gets written again when it still applies.
GENERATED_NOTES = {spelltext.BONUS_NOTE, 'Some client values are unavailable (?).'}

# Hand notes that described where a hand-assembled tooltip came from. A generated tooltip
# comes from one place, so they no longer apply.
STALE_NOTES = re.compile(r'^Brought up to build |^icon is a placeholder')

NUMBER = re.compile(r'\d+(?:\.\d+)?')


def generic(note):
	return note in GENERATED_NOTES or note.startswith('Values shown at level')


def hand_note(note):
	"""Whether a note in the current file was written by hand and still applies."""
	return not generic(note) and not STALE_NOTES.search(note)


def rows_by(table, build, key):
	d = collections.defaultdict(list)
	for r in wago.rows(table, build):
		d[r[key]].append(r)
	return d


class Tables:
	"""The trait tables of one build, indexed the ways we look them up."""

	def __init__(self, build):
		self.build = build
		self.nodes = wago.rows('TraitNode', build)
		self.node_entry = {}
		for r in sorted(wago.rows('TraitNodeXTraitNodeEntry', build), key=lambda r: int(r['_Index'])):
			self.node_entry.setdefault(r['TraitNodeID'], r['TraitNodeEntryID'])
		self.entries = wago.by_id('TraitNodeEntry', build)
		self.definitions = wago.by_id('TraitDefinition', build)
		self.effect_points = rows_by('TraitDefinitionEffectPoints', build, 'TraitDefinitionID')
		self.curves = rows_by('CurvePoint', build, 'CurveID')
		self.edges = wago.rows('TraitEdge', build)
		self.tabs = wago.rows('TalentTab', build)
		self.classes = wago.rows('ChrClasses', build)
		self.class_sets = {r['SpellID']: r['SpellClassSet'] for r in wago.rows('SpellClassOptions', build)}
		self.skill_lines = wago.by_id('SkillLine', build)
		self.spell_skills = collections.defaultdict(set)
		for r in wago.rows('SkillLineAbility', build):
			self.spell_skills[r['Spell']].add(r['SkillLine'])

	def definition(self, node_id):
		entry = self.entries[self.node_entry[node_id]]
		return entry, self.definitions[entry['TraitDefinitionID']]

	def curve(self, curve_id, x):
		"""A curve's value at x, between its points when x falls between them."""
		points = sorted(((float(p['Pos_0']), float(p['Pos_1'])) for p in self.curves.get(curve_id, [])))
		if not points:
			return None
		if x <= points[0][0]:
			return points[0][1]
		for (x0, y0), (x1, y1) in zip(points, points[1:]):
			if x0 <= x <= x1:
				return y0 if x1 == x0 else y0 + (y1 - y0) * (x - x0) / (x1 - x0)
		return points[-1][1]


def class_trees(t):
	"""{class filename lowercased: TraitTreeID}, by the class of the talent spells in each tree."""
	by_set = {r['SpellClassSet']: r for r in t.classes}
	votes = collections.defaultdict(collections.Counter)
	for n in t.nodes:
		if n['ID'] not in t.node_entry:
			continue
		_, d = t.definition(n['ID'])
		cls = by_set.get(t.class_sets.get(d['SpellID']))
		if cls:
			votes[n['TraitTreeID']][cls['Filename'].lower()] += 1
	trees = {}
	for tree, v in votes.items():
		cls, n = v.most_common(1)[0]
		# Small trees of class spells (a class's other trait systems) lose to the talent tree.
		if cls not in trees or n > sum(votes[trees[cls]].values()):
			trees[cls] = tree
	return trees


def groups(values, gap):
	"""Sorted values split wherever two neighbours are more than gap apart."""
	out = []
	for v in sorted(values):
		if out and v - out[-1][-1] <= gap:
			out[-1].append(v)
		else:
			out.append([v])
	return out


def step(values):
	"""The grid step: the commonest distance between neighbouring positions."""
	diffs = collections.Counter(b - a for a, b in zip(values, values[1:]) if b - a > 50)
	return diffs.most_common(1)[0][0]


def layout(nodes):
	"""{node ID: (tab index, row, col)} for the nodes on the grid, and the IDs left off it.

	Positions sit on a grid with the same step both ways (600 in the builds so far), with a
	few nodes 10 off it. The three trees are the three groups of columns. A node whose row or
	column group is a stray far from the rest is parked off the grid.
	"""
	xs = sorted({int(n['PosX']) for n in nodes})
	ys = sorted({int(n['PosY']) for n in nodes})
	size = step(ys)
	col_groups = [g for g in groups(xs, size * 2) if len([n for n in nodes if g[0] <= int(n['PosX']) <= g[-1]]) > 2]
	row_group = max(groups(ys, size * 2), key=lambda g: len([n for n in nodes if g[0] <= int(n['PosY']) <= g[-1]]))
	top = row_group[0]
	placed, parked = {}, []
	for n in nodes:
		x, y = int(n['PosX']), int(n['PosY'])
		tab = next((i for i, g in enumerate(col_groups) if g[0] <= x <= g[-1]), None)
		if tab is None or not row_group[0] <= y <= row_group[-1]:
			parked.append(n['ID'])
			continue
		row = int(round((y - top) / size)) + 1
		col = int(round((x - col_groups[tab][0]) / size)) + 1
		placed[n['ID']] = (tab, row, col)
	return placed, parked


def rank_points(t, definition, rank):
	"""{effect index: value} for one rank of a talent, from its curves."""
	points = {}
	for p in t.effect_points.get(definition['ID'], []):
		v = t.curve(p['CurveID'], rank)
		if v is not None:
			points[int(p['EffectIndex'])] = v
	return points


def tree_name(t, spells, fallback):
	"""The skill line most of a tree's talent spells belong to."""
	votes = collections.Counter()
	for s in spells:
		for line in t.spell_skills.get(s, ()):
			sl = t.skill_lines.get(line)
			if sl and sl['DisplayName_lang']:
				votes[sl['DisplayName_lang']] += 1
	return votes.most_common(1)[0][0] if votes else fallback


def generate(t, cls, tree_id, old):
	"""The client/<cls>.json content for a class, with what we carry over from old.

	Answers it and a list of (tree, talent, note) for every tooltip construct left as "?".
	"""
	build = t.build
	nodes = [n for n in t.nodes if n['TraitTreeID'] == tree_id and n['ID'] in t.node_entry]
	placed, parked = layout(nodes)
	for node_id in parked:
		_, d = t.definition(node_id)
		print('%s: node %s (%s) is off the grid, left out' % (cls, node_id, spelltext.name(d['SpellID'], build)),
		      file=sys.stderr)

	class_row = next(r for r in t.classes if r['Filename'].lower() == cls)
	mask = 1 << (int(class_row['ID']) - 1)
	tabs = sorted((r for r in t.tabs if int(r['ClassMask']) & mask), key=lambda r: int(r['OrderIndex']))

	names = {}
	for node_id in placed:
		entry, d = t.definition(node_id)
		names[node_id] = d['OverrideName_lang'] or spelltext.name(d['SpellID'], build)

	# The prerequisite of a talent is the node on the other end of an edge into it. The client
	# lists some edges both ways, and then the prerequisite is the one higher up the tree.
	reqs = {}
	for e in t.edges:
		a, b = e['LeftTraitNodeID'], e['RightTraitNodeID']
		if a in placed and b in placed and placed[a][1] <= placed[b][1]:
			if placed[a][1] < placed[b][1] or b not in reqs:
				reqs[b] = a

	old_trees = old.get('trees', []) if old else []
	old_by_name = {x['name']: x for tr in old_trees for x in tr['talents']}
	old_by_id = {x['id']: x for tr in old_trees for x in tr['talents'] if x.get('id')}

	trees = []
	carried_notes = carried_removed = 0
	left = []
	for index, tab in enumerate(tabs):
		ids = sorted((i for i, p in placed.items() if p[0] == index), key=lambda i: placed[i][1:])
		spells = [t.definition(i)[1]['SpellID'] for i in ids]
		name = tree_name(t, spells, tab['Name_lang'])
		tree = {
			'name': name,
			'bg': int(tab['ID']),
			'icon': spelltext.icon_file(tab['SpellIconID'], build),
		}
		previous = next((o for o in old_trees if o['name'] == name), None)
		if previous is None and index < len(old_trees):
			previous = old_trees[index]
		# Kept as it is, even when a talent of the same name is back (Blood Frenzy was
		# Classic's talent before Forever gave the name to a new one).
		removed = (previous or {}).get('removed', [])
		if removed:
			tree['removed'] = removed
			carried_removed += len(removed)
		talents = []
		for node_id in ids:
			# By name first, since the hand-assembled files have a few node IDs swapped.
			before = old_by_name.get(names[node_id]) or old_by_id.get(int(node_id))
			talent, carried, unfilled = talent_json(t, node_id, placed[node_id], names, reqs, before)
			carried_notes += carried
			left.extend((name, talent['name'], n) for n in unfilled)
			talents.append(talent)
		tree['talents'] = talents
		trees.append(tree)

	source = ('wago.tools Trait, TalentTab, SkillLine and Spell tables of the WoW Forever beta client '
	          'build %s (TraitTree %s), generated by tools/forever_talents/import_db2.py.' % (t.build, tree_id))
	if carried_removed or carried_notes:
		kept = []
		if carried_removed:
			kept.append("the trees' removed lists")
		if carried_notes:
			kept.append('the hand notes on %d talent%s' % (carried_notes, '' if carried_notes == 1 else 's'))
		source += ' %s carried over from the hand-assembled files these replaced.' % ' and '.join(kept).capitalize()
	return {'source': source, 'class': cls, 'trees': trees}, left


def talent_json(t, node_id, place, names, reqs, before):
	"""One talent, whether it carried hand notes over from before, and the notes about
	constructs its tooltip left as "?"."""
	build = t.build
	entry, d = t.definition(node_id)
	spell = d['SpellID']
	max_rank = int(entry['MaxRanks'])
	notes = []
	desc = {}
	for rank in range(1, max_rank + 1):
		if d['OverrideDescription_lang']:
			desc[str(rank)] = d['OverrideDescription_lang']
		else:
			desc[str(rank)] = spelltext.description(spell, build, points=rank_points(t, d, rank), notes=notes)
	unfilled = [n for n in notes if spelltext.unresolved(n)]
	icon_id = d['OverrideIcon'] if d['OverrideIcon'] not in ('', '0') else None
	icon = spelltext.icon_file(icon_id, build) if icon_id else spelltext.icon(spell, build)
	if not icon:
		notes.append('The client icon has no name in ManifestInterfaceData.')
	is_passive = spelltext.passive(spell, build)

	talent = {
		'name': names[node_id],
		'max': max_rank,
		'row': place[1],
		'col': place[2],
		'passive': is_passive,
		'icon': icon or 'inv_misc_questionmark',
		'desc': desc,
	}
	if node_id in reqs:
		talent['req'] = names[reqs[node_id]]
	if not is_passive:
		needs = spelltext.requirements(spell, build)
		# A spell whose tooltip already starts with its requirement (Feral Charge, Berserk)
		# doesn't repeat it.
		if needs and not desc['1'].startswith('Requires'):
			talent['reqText'] = needs
		line = spelltext.cost(spell, build)
		if line:
			talent['cost'] = line
	talent['id'] = int(node_id)

	carried = False
	if before:
		for n in before.get('tooltip', {}).get('notes', []):
			if not hand_note(n):
				continue
			if n not in notes:
				notes.append(n)
				carried = True
	talent['tooltip'] = {
		'source': 'client-partial' if unfilled else 'client',
		'build': build,
		'spellId': int(spell),
		'notes': notes,
	}
	return talent, carried, unfilled


# Comparing a generated class with the current file.

def numbers(text):
	return NUMBER.findall(text or '')


def same_numbers(a, b):
	"""Whether two texts carry the same numbers, in any order, once rounded to whole numbers.

	The hand-assembled files print some values unrounded (Vindication's 66.833), where the
	game prints a whole number unless the template asks for decimals.
	"""
	def whole(text):
		return sorted(spelltext.round_half_up(float(n)) for n in numbers(text))
	return whole(a) == whole(b)


def words(text):
	return re.sub(r'\s+', ' ', (text or '').replace('\\n', '\n')).strip()


def compare(old, new):
	"""(structure, numbers, wording) lists of differences between two class files."""
	structure, nums, wording = [], [], []
	old_by, new_by = {}, {}
	for side, by in ((old, old_by), (new, new_by)):
		for tr in side['trees']:
			for x in tr['talents']:
				by[x.get('id') or x['name']] = (tr['name'], x)
	# By name first, so a talent the old file gave the wrong node ID still pairs with
	# itself, then by node ID, which pairs a renamed talent.
	new_names = {x['name']: k for k, (_, x) in new_by.items()}
	matched = {}
	for k, v in old_by.items():
		if v[1]['name'] in new_names:
			matched[k] = new_names[v[1]['name']]
	for k in old_by:
		if k not in matched and k in new_by and k not in matched.values():
			matched[k] = k
	for k in old_by:
		if k not in matched:
			structure.append('removed: %s (%s)' % (old_by[k][1]['name'], old_by[k][0]))
	for k in new_by:
		if k not in matched.values():
			tr, x = new_by[k]
			structure.append('added: %s (%s, row %d col %d, %d rank%s)' % (
				x['name'], tr, x['row'], x['col'], x['max'], '' if x['max'] == 1 else 's'))
	for ko, kn in matched.items():
		(otree, o), (ntree, n) = old_by[ko], new_by[kn]
		label = n['name']
		if otree != ntree:
			structure.append('%s: tree %s -> %s' % (label, otree, ntree))
		if o['name'] != n['name']:
			structure.append('%s: renamed from %s' % (label, o['name']))
		for f in ('id', 'row', 'col', 'max', 'req', 'passive'):
			if o.get(f) != n.get(f):
				structure.append('%s: %s %s -> %s' % (label, f, o.get(f), n.get(f)))
		for rank in sorted(set(o['desc']) | set(n['desc']), key=int):
			a, b = o['desc'].get(rank), n['desc'].get(rank)
			if a is None or b is None:
				continue
			if not same_numbers(a, b):
				nums.append('%s rank %s: %s -> %s' % (label, rank, '/'.join(numbers(a)), '/'.join(numbers(b))))
			elif words(a) != words(b):
				wording.append('%s rank %s:\n      old: %s\n      new: %s' % (label, rank, words(a), words(b)))
		for f in ('cost', 'reqText', 'icon'):
			a, b = o.get(f), n.get(f)
			if a == b:
				continue
			if f == 'cost' and not same_numbers(a, b):
				nums.append('%s cost: %s -> %s' % (label, a, b))
			else:
				wording.append('%s %s: %s -> %s' % (label, f, a, b))
	return structure, nums, wording


def main():
	args = [a for a in sys.argv[1:] if not a.startswith('--')]
	write = '--write' in sys.argv
	if not args:
		print(__doc__)
		return 1
	build = wago.version(args[0])
	classes = args[1:] or CLASSES
	t = Tables(build)
	trees = class_trees(t)
	for cls in classes:
		path = os.path.join(CLIENT_DIR, cls + '.json')
		old = None
		if os.path.exists(path):
			with open(path, encoding='utf-8') as f:
				old = json.load(f)
		new, left = generate(t, cls, trees[cls], old)
		if old:
			structure, nums, wording = compare(old, new)
			print('== %s: %d structure, %d numbers, %d wording' % (cls, len(structure), len(nums), len(wording)))
			for kind, lines in (('structure', structure), ('numbers', nums), ('wording', wording)):
				for line in lines:
					print('  %s  %s' % (kind, line))
		for tr, name, n in left:
			print('  ?  %s: %s: %s' % (tr, name, n))
		if write:
			with open(path, 'w', encoding='utf-8') as f:
				json.dump(new, f, indent=1, ensure_ascii=False)
				f.write('\n')
			print('wrote ' + path)
	return 0


if __name__ == '__main__':
	sys.exit(main())
