#!/usr/bin/env python3
# usage: talent_check.py REQ_DIR [class ...]: checks each req's talentsString against ui/core/talents/trees/<class>.json
# (ranks <= max, 5 points a row in that tree, prerequisite maxed, 51 points), and the tree's ranks against
# tools/forever_talents/notes/<class>.json "max" fields. Prints only problems and a point total per req.
import json, glob, os, sys
root = os.path.dirname(os.path.dirname(os.path.dirname(os.path.dirname(os.path.abspath(__file__)))))
req_dir = sys.argv[1]; classes = sys.argv[2:] or ['rogue', 'warrior', 'shaman']
for cls in classes:
    trees = json.load(open(f'{root}/ui/core/talents/trees/{cls}.json'))
    notes = f'{root}/tools/forever_talents/notes/{cls}.json'
    if os.path.exists(notes):
        byname = {t['name']: (ti, t) for ti, tr in enumerate(trees) for t in tr['talents']}
        for e in json.load(open(notes))['entries']:
            f = e.get('fields') or e.get('talent') or {}
            name = f.get('name') or e.get('name')  # a set with fields.name is a rename
            if e.get('op') == 'remove':
                if name in byname: print(f'{cls} tree: removed {name} still in ui tree')
                continue
            if name not in byname:
                print(f'{cls} tree: notes entry {name} ({e.get("op")}) not in ui tree'); continue
            t = byname[name][1]
            if 'max' in f and t['maxPoints'] != f['max']: print(f'{cls} tree: {name} max {t["maxPoints"]} vs notes {f["max"]}')
            if 'row' in f and (t['location']['rowIdx'] + 1, t['location']['colIdx'] + 1) != (f['row'], f['col']):
                print(f'{cls} tree: {name} at {t["location"]} vs notes row {f["row"]} col {f["col"]} (1-based)')
            if 'req' in f:
                pl = t.get('prereqLocation'); pn = None
                if pl:
                    pn = [x['name'] for x in trees[byname[name][0]]['talents'] if (x['location']['rowIdx'], x['location']['colIdx']) == (pl['rowIdx'], pl['colIdx'])]
                if (pn[0] if pn else None) != f['req']: print(f'{cls} tree: {name} prereq {pn} vs notes {f["req"]}')
    for rf in sorted(glob.glob(f'{req_dir}/{cls}__*.json')):
        if rf.endswith('.meta.json'): continue
        ts = json.load(open(rf))['raid']['parties'][0]['players'][0].get('talentsString', '')
        parts = ts.split('-'); probs = []; total = 0
        for ti, tr in enumerate(trees):
            s = parts[ti] if ti < len(parts) else ''
            tal = tr['talents']
            if len(s) > len(tal): probs.append(f'tree{ti} string longer than tree')
            pts = {}
            for i, t in enumerate(tal):
                p = int(s[i]) if i < len(s) else 0
                loc = (t['location']['rowIdx'], t['location']['colIdx']); pts[loc] = (p, t)
                if p > t['maxPoints']: probs.append(f'{t["name"]} {p}>{t["maxPoints"]}')
            for loc, (p, t) in pts.items():
                if not p: continue
                below = sum(q for (r, c), (q, _) in pts.items() if r < loc[0])
                if below < 5 * loc[0]: probs.append(f'{t["name"]} row {loc[0]} with {below} points above')
                pl = t.get('prereqLocation')
                if pl:
                    q, pt = pts.get((pl['rowIdx'], pl['colIdx']), (0, None))
                    if pt is None or q < pt['maxPoints']: probs.append(f'{t["name"]} needs {pt and pt["name"]} maxed ({q})')
            total += sum(p for p, _ in pts.values())
        if total != 51: probs.append(f'{total} points')
        print(os.path.basename(rf)[:-5], ts, '; '.join(probs) or 'ok')
