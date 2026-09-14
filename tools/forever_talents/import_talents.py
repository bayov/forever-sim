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

import json
import os
import sys

DATA_DIR = os.path.join(os.path.dirname(__file__), 'data')
OVERRIDE_DIR = os.path.join(os.path.dirname(__file__), 'overrides')
TREE_DIR = os.path.join(os.path.dirname(__file__), '..', '..', 'ui', 'core', 'talents', 'trees')

# Placeholder icons for talents that didn't exist in Classic, so the picker has something to
# draw. These are ids from later expansions and show the wrong tooltip until the beta client
# is datamined.
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


def build_tree_json(data, existing_by_tree):
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


def main():
	if len(sys.argv) < 2:
		print(__doc__)
		sys.exit(1)

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
	trees = build_tree_json(data, existing_by_tree)

	if write:
		with open(tree_path, 'w') as f:
			json.dump(trees, f, indent=2)
			f.write('\n')
		print('wrote ' + tree_path)

	print(build_proto(data))


if __name__ == '__main__':
	main()
