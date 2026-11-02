#!/usr/bin/env python3
"""Side by side, per ability: DPS, casts, hits, crits and misses an iteration, for the player and each pet, on the
first target. Works on either sim's result.

    breakdown.py <out dir>               res_ours.json against res_theirs.json
    breakdown.py <result.json> [...]     any results

Ids differ where convert.py remaps them (ranks, tags, the AQ ranks), and for auto attacks (ours: OtherActionAttack
tags 1, 2 and 3 for the main hand, off hand and extra attacks; theirs: tags 14 and 15, and the extra attack's own
spell): read the two rows of one ability together.
"""
import json, os, sys

def load(path):
    r = json.load(open(path))
    pl = r['raidMetrics']['parties'][0]['players'][0]
    n = r.get('iterationsDone') or 1
    dur = r.get('avgIterationDuration') or 180
    rows = {}
    def add(unit, who):
        for a in unit.get('actions', []):
            i = a['id']
            key = (who, i.get('spellId') or i.get('itemId') or i.get('otherId'), i.get('tag', 0))
            agg = rows.setdefault(key, [0.0] * 6)
            for t in a['targets'][:1]:
                for j, f in enumerate(('damage', 'casts', 'hits', 'crits', 'misses', 'dodges')):
                    agg[j] += t.get(f, 0)
    add(pl, 'player')
    for pet in pl.get('pets', []):
        add(pet, 'pet:' + pet.get('name', '?'))
    pets = sum(x['dps'].get('avg', 0) for x in pl.get('pets', []))
    return {'dps': pl['dps']['avg'], 'pets': pets, 'rows': rows, 'n': n, 'dur': dur}

paths = sys.argv[1:]
if len(paths) == 1 and os.path.isdir(paths[0]):
    paths = [os.path.join(paths[0], 'res_ours.json'), os.path.join(paths[0], 'res_theirs.json')]
res = [load(p) for p in paths]
for p, r in zip(paths, res):
    print(f'{p}: DPS {r["dps"]:.1f} (pets {r["pets"]:.1f}), {r["n"]} iterations of {r["dur"]:.1f} s')
keys = sorted(set().union(*[r['rows'] for r in res]), key=lambda k: -max(r['rows'].get(k, [0])[0] / r['n'] / r['dur'] for r in res))
cols = '%7s %6s %6s %6s %5s %5s' % ('dps', 'casts', 'hits', 'crits', 'miss', 'dodge')
print('%-18s %8s %3s ' % ('who', 'id', 'tag') + ' | '.join(cols for _ in res))
for k in keys:
    line = []
    for r in res:
        v = r['rows'].get(k, [0.0] * 6)
        line.append('%7.1f %6.1f %6.1f %6.1f %5.1f %5.1f' % (v[0] / r['n'] / r['dur'], *(x / r['n'] for x in v[1:])))
    print('%-18s %8s %3s ' % (k[0][:18], k[1], k[2]) + ' | '.join(line))
