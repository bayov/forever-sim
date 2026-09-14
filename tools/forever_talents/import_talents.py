#!/usr/bin/python

# Generates the UI talent tree layout, e.g. 'ui/core/talents/trees/warlock.json', and the
# matching talents proto message from the Forever talent data in ./data.
#
# The proto field numbers and the order of the talents in the tree json have to agree,
# because FillTalentsProto maps the nth character of a talent string to field number n.
# Both are emitted in (row, col) order so they line up by construction.
#
# Usage:
#   tools/forever_talents/import_talents.py warlock            # print the proto message
#   tools/forever_talents/import_talents.py warlock --write    # also rewrite the tree json
#   tools/forever_talents/import_talents.py --unranked         # list talents whose per-rank
#                                                                scaling isn't known for any class

import json
import os
import re
import sys

DATA_DIR = os.path.join(os.path.dirname(__file__), 'data')
OVERRIDE_DIR = os.path.join(os.path.dirname(__file__), 'overrides')
TREE_DIR = os.path.join(os.path.dirname(__file__), '..', '..', 'ui', 'core', 'talents', 'trees')
SIM_DIR = os.path.join(os.path.dirname(__file__), '..', '..', 'sim')

# Talents that didn't exist in Classic have no spell id to give the picker. They carry their
# name, icon and tooltip through to the tree json instead, and the picker draws those.
PLACEHOLDER_SPELL_ID = 0


def camel_case(talent_id):
	head, *rest = talent_id.split('-')
	return head + ''.join(part.title() for part in rest)


def load_class(class_name):
	with open(os.path.join(DATA_DIR, class_name + '.json')) as f:
		data = json.load(f)

	override_path = os.path.join(OVERRIDE_DIR, class_name + '.json')
	if os.path.exists(override_path):
		with open(override_path) as f:
			overrides = json.load(f)['overrides']

		by_id = {talent['id']: talent for tree in data['trees'] for talent in tree['talents']}
		for override in overrides:
			talent = by_id.get(override['talent'])
			if talent is not None:
				talent.update(override.get('set', {}))

	return data


def simulated_talents(class_name):
	"""Field names the class's Go package actually reads.

	A talent the sim never looks at still draws a box the player can spend points in, and
	since the tooltips landed it describes an effect that isn't modelled. Rather than keep
	a hand-written list in step with the code, read it back off the source: anything whose
	Go field name appears nowhere outside the generated protobuf is not simulated.
	"""
	seen = set()
	class_dir = os.path.join(SIM_DIR, class_name)
	for root, _, files in os.walk(class_dir):
		if os.sep + 'proto' + os.sep in root + os.sep:
			continue
		for name in files:
			if not name.endswith('.go') or name.startswith('_'):
				continue
			with open(os.path.join(root, name)) as f:
				seen.update(re.findall(r'\b([A-Z][A-Za-z0-9_]*)\b', f.read()))
	return seen

def sorted_talents(tree):
	return sorted(tree['talents'], key=lambda talent: (talent['row'], talent['col']))


def spell_ids(talent, existing):
	# Keep whatever the tree already used so regenerating doesn't churn icons.
	if talent['name'] in existing:
		return existing[talent['name']]

	prior = talent.get('ranksPrior') or {}
	ids = prior.get('classicSpellIds') or []
	if ids:
		return (ids + [ids[-1]] * talent['maxRank'])[:talent['maxRank']]

	return [PLACEHOLDER_SPELL_ID] * talent['maxRank']


def build_tree_json(data, existing_by_tree, simulated):
	trees = []
	for tree in sorted(data['trees'], key=lambda t: t['order']):
		existing = existing_by_tree.get(tree['name'], {})
		locations = {talent['id']: talent for talent in tree['talents']}

		talents = []
		for talent in sorted_talents(tree):
			entry = {
				'fieldName': camel_case(talent['id']),
				'location': {'rowIdx': talent['row'], 'colIdx': talent['col']},
			}

			requires = talent.get('requires') or []
			if requires:
				prereq = locations.get(requires[0]['talent'])
				if prereq is not None:
					entry['prereqLocation'] = {'rowIdx': prereq['row'], 'colIdx': prereq['col']}

			entry['spellIds'] = spell_ids(talent, existing)
			entry['maxPoints'] = talent['maxRank']

			# Every talent carries its Forever name, tooltip and per-rank values, so the picker
			# can describe what the talent does now rather than what its Classic spell id says
			# on Wowhead. The spell id is still used for the icon when there is one.
			entry['name'] = talent['name']
			if talent.get('description'):
				entry['description'] = talent['description']
			ranks = talent.get('ranks')
			if ranks:
				# A few talents carry whole tooltips per rank instead of values. Turn those
				# into a single placeholder so the picker shows the right rank's text.
				if not all(isinstance(rank, list) for rank in ranks):
					entry['description'] = '{0}'
					ranks = [[rank] if not isinstance(rank, list) else rank for rank in ranks]
				entry['ranks'] = ranks
			# Only rank 1 was ever on screen. Ranks copied from the Classic talent of the same
			# name are a fair bet, extrapolated ones are a guess, so say so in the tooltip.
			if talent['maxRank'] > 1 and talent.get('ranksSource') in ('extrapolated', 'manual'):
				entry['ranksGuessed'] = True
			# The icon name is used directly, so the picker never has to ask Wowhead about a
			# spell id that may not exist in Classic. An icon sourced from 'crop' is the name
			# of the screenshot the tooltip was read from, not a real icon, so it is left out
			# and the picker falls back to the spell id, then to its own placeholder.
			if talent.get('icon') and talent.get('iconSource') != 'crop':
				entry['icon'] = talent['icon']
			if entry['fieldName'][0].upper() + entry['fieldName'][1:] not in simulated:
				entry['notSimulated'] = True

			talents.append(entry)

		trees.append({
			'name': tree['name'],
			'backgroundUrl': existing_by_tree.get('backgroundUrl', {}).get(tree['name'], ''),
			'talents': talents,
		})

	return trees


def build_proto(data):
	lines = ['message %sTalents {' % data['className']]
	field = 1
	for tree in sorted(data['trees'], key=lambda t: t['order']):
		lines.append('\t// %s' % tree['name'])
		for talent in sorted_talents(tree):
			name = camel_case(talent['id'])
			snake = ''.join('_' + c.lower() if c.isupper() else c for c in name)
			kind = 'bool' if talent['maxRank'] == 1 else 'int32'
			lines.append('\t%s %s = %d;' % (kind, snake, field))
			field += 1
		lines.append('')
	lines.append('}')
	return '\n'.join(lines)


# Talents where the source data repeats rank 1's numbers for every rank, so the real
# per-rank scaling is unknown. Anything implemented from these is a guess.
def unranked_talents(data):
	unranked = []
	for tree in sorted(data['trees'], key=lambda t: t['order']):
		for talent in sorted_talents(tree):
			ranks = talent.get('ranks') or []
			if talent['maxRank'] > 1 and len(ranks) > 1 and all(r == ranks[0] for r in ranks):
				unranked.append((tree['name'], talent['name'], talent['maxRank']))
	return unranked


def report_unranked():
	total = 0
	for path in sorted(os.listdir(DATA_DIR)):
		if not path.endswith('.json'):
			continue

		data = load_class(path[:-len('.json')])
		unranked = unranked_talents(data)
		total += len(unranked)
		print('%s: %d' % (data['className'], len(unranked)))
		for tree_name, name, max_rank in unranked:
			print('\t%s, %s, %d ranks' % (tree_name, name, max_rank))

	print('%d talents in total.' % total)


def main():
	if len(sys.argv) < 2:
		print(__doc__)
		sys.exit(1)

	if sys.argv[1] == '--unranked':
		report_unranked()
		return

	class_name = sys.argv[1]
	write = '--write' in sys.argv
	data = load_class(class_name)

	tree_path = os.path.join(TREE_DIR, class_name + '.json')
	existing_by_tree = {}
	backgrounds = {}
	if os.path.exists(tree_path):
		with open(tree_path) as f:
			for tree in json.load(f):
				existing_by_tree[tree['name']] = {}
				backgrounds[tree['name']] = tree.get('backgroundUrl', '')

		# Index the current spell ids by talent name so they survive a regeneration.
		with open(tree_path) as f:
			current = json.load(f)
		by_field = {t['fieldName']: t['spellIds'] for tree in current for t in tree['talents']}
		for tree in data['trees']:
			existing_by_tree.setdefault(tree['name'], {})
			for talent in tree['talents']:
				field_name = camel_case(talent['id'])
				if field_name in by_field:
					existing_by_tree[tree['name']][talent['name']] = by_field[field_name]

	existing_by_tree['backgroundUrl'] = backgrounds
	trees = build_tree_json(data, existing_by_tree, simulated_talents(class_name))

	if write:
		with open(tree_path, 'w') as f:
			json.dump(trees, f, indent=2)
			f.write('\n')
		print('wrote ' + tree_path)

	print(build_proto(data))

	unranked = unranked_talents(data)
	if unranked:
		sys.stderr.write('%d talents have no known per-rank scaling:\n' % len(unranked))
		for tree_name, name, max_rank in unranked:
			sys.stderr.write('\t%s, %s, %d ranks\n' % (tree_name, name, max_rank))


if __name__ == '__main__':
	main()
