#!/usr/bin/env python3

# Rewrites data/<class>.json from the beta client's talent tree in client/<class>.json.
#
# client/ holds the trees hyjal.cc reads out of the WoW Forever beta client (see README.md),
# one tooltip per rank with the numbers already filled in. data/ is what import_talents.py
# reads, one description with {n} placeholders and a list of numbers per rank. This script
# turns the former into the latter, so a class whose client tree is in, and the sim code has
# been checked against it, is regenerated with
#
#   tools/forever_talents/import_client.py rogue
#   tools/forever_talents/import_talents.py rogue --write
#
# The placeholders are found by comparing the ranks: every number in the tooltip becomes a
# placeholder when the text around the numbers is the same at every rank. The few talents
# whose wording changes between ranks keep a whole tooltip per rank instead, which the
# importer and the picker both understand.

import datetime
import json
import os
import re
import sys

CLIENT_DIR = os.path.join(os.path.dirname(__file__), 'client')
DATA_DIR = os.path.join(os.path.dirname(__file__), 'data')

NUMBER = re.compile(r'-?\d+(?:\.\d+)?')


def slug(name):
	return re.sub(r'[^a-z0-9]+', '-', name.lower().replace("'", '')).strip('-')


def clean(text):
	# The client wraps a few tooltips by hand (Hack and Slash, Venom). One line, one space.
	text = text.replace('\\n', '\n')
	text = re.sub(r'\s*\n\s*', ' ', text)
	text = re.sub(r'\s+', ' ', text)
	text = text.replace(' :', ':')
	return text.strip()


def parse_number(token):
	value = float(token)
	return int(value) if value.is_integer() else value


def template(descriptions):
	"""One description with placeholders plus the numbers per rank, or None when the
	wording differs between ranks."""
	skeletons = [NUMBER.sub('{}', d) for d in descriptions]
	if any(s != skeletons[0] for s in skeletons):
		return None

	ranks = [[parse_number(n) for n in NUMBER.findall(d)] for d in descriptions]
	index = iter(range(len(ranks[0])))
	description = NUMBER.sub(lambda _: '{%d}' % next(index), descriptions[0])
	return description, ranks


def convert_talent(talent, by_name):
	descriptions = [clean(talent['desc'][str(rank)]) for rank in range(1, talent['max'] + 1)]
	templated = template(descriptions)
	if templated is not None:
		description, ranks = templated
	else:
		description, ranks = '{0}', descriptions

	entry = {
		'id': slug(talent['name']),
		'name': talent['name'],
		'row': talent['row'] - 1,
		'col': talent['col'] - 1,
		'maxRank': talent['max'],
		'icon': talent['icon'],
		'iconSource': 'client',
		'description': description,
		'ranks': ranks,
		'ranksObserved': list(range(1, talent['max'] + 1)),
		'ranksSource': 'client',
	}
	if talent.get('cost'):
		entry['cost'] = talent['cost']
	if talent.get('req'):
		prereq = by_name[talent['req']]
		entry['requires'] = [{'talent': slug(prereq['name']), 'rank': prereq['max']}]
	spell_id = talent.get('tooltip', {}).get('spellId')
	if spell_id:
		entry['spellId'] = spell_id
	return entry


def convert(class_name):
	with open(os.path.join(CLIENT_DIR, class_name + '.json')) as f:
		client = json.load(f)

	trees = []
	for order, tree in enumerate(client['trees']):
		by_name = {t['name']: t for t in tree['talents']}
		trees.append({
			'id': slug(tree['name']),
			'name': tree['name'],
			'page': 'primary',
			'order': order,
			'icon': tree['icon'],
			'rows': 7,
			'cols': 4,
			'talents': [convert_talent(t, by_name) for t in tree['talents']],
		})

	return {
		'schemaVersion': 1,
		'class': class_name,
		'className': class_name.title(),
		'dataVersion': 2,
		'dataSource': 'client',
		'generatedAt': datetime.datetime.now(datetime.timezone.utc).strftime('%Y-%m-%dT%H:%M:%SZ'),
		'rules': {'pointsPerRow': 5, 'maxPoints': 51, 'firstPointLevel': 10, 'maxLevel': 60, 'rulesSource': 'assumed'},
		'pages': [{'id': 'primary', 'name': 'Primary'}],
		'trees': trees,
		'notes': [client['source']],
	}


def main():
	if len(sys.argv) != 2:
		print(__doc__)
		sys.exit(1)

	class_name = sys.argv[1]
	data = convert(class_name)
	path = os.path.join(DATA_DIR, class_name + '.json')
	with open(path, 'w') as f:
		json.dump(data, f, indent=2)
		f.write('\n')
	print('wrote ' + path)

	for tree in data['trees']:
		for talent in tree['talents']:
			if talent['description'] == '{0}':
				print('%s: %s keeps a whole tooltip per rank' % (tree['name'], talent['name']))


if __name__ == '__main__':
	main()
