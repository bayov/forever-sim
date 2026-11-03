#!/usr/bin/env python3
# usage: dead_lines.py REQ_DIR RUN_DIR [--min-rate 0.05]
# Screens every req preset against a finished run (one result JSON per req id, e.g. the audit loop's
# $AUDIT_WORK/runs/master): lists APL castSpell ids (priority list and pre-pull) that never or almost
# never cast (casts per iteration < --min-rate), plus resource health: seconds OOM, mana left, and
# overcapped rage/energy (gain - actualGain, as a share of gain). A listed line is a candidate, not a
# finding: end-of-fight, race-specific and talent-less lines are inert by design (passes/rotations.md).
import json, os, sys, glob

def ids(node, out):
    if isinstance(node, dict):
        cs = node.get('castSpell')
        if isinstance(cs, dict) and isinstance(cs.get('spellId'), dict):
            out.add(json.dumps(cs['spellId'], sort_keys=True))
        for v in node.values():
            ids(v, out)
    elif isinstance(node, list):
        for v in node:
            ids(v, out)

def key(i):
    d = {k: v for k, v in i.items() if k in ('spellId', 'itemId', 'otherId') and v}
    return d

def main():
    req_dir, run_dir = sys.argv[1], sys.argv[2]
    min_rate = float(sys.argv[sys.argv.index('--min-rate') + 1]) if '--min-rate' in sys.argv else 0.05
    for rf in sorted(glob.glob(os.path.join(req_dir, '*.json'))):
        if rf.endswith('.meta.json'):
            continue
        rid = os.path.basename(rf)[:-5]
        res = os.path.join(run_dir, rid + '.json')
        if not os.path.exists(res):
            print(f'{rid}: no result'); continue
        req = json.load(open(rf)); out = json.load(open(res))
        player = req['raid']['parties'][0]['players'][0]
        rot = player.get('rotation', {})
        want = set(); ids(rot, want)
        p = out['raidMetrics']['parties'][0]['players'][0]
        it = out.get('iterationsDone') or 1
        casts = {}
        for a in p['actions'] + [x for pet in p.get('pets', []) for x in pet['actions']]:
            k = json.dumps(key(a['id']), sort_keys=True)
            casts[k] = casts.get(k, 0) + sum(t.get('casts', 0) for t in a.get('targets', []))
        dead = []
        for w in sorted(want):
            wk = json.dumps(key(json.loads(w)), sort_keys=True)
            r = casts.get(wk, 0) / it
            if r < min_rate:
                dead.append(f'{wk}={r:.3f}')
        waste = []
        for r in p.get('resources', []):
            if r['type'] in ('ResourceTypeRage', 'ResourceTypeEnergy') and r['gain'] > 0:
                waste.append((r['type'][12:], r['gain'], r['actualGain']))
        agg = {}
        for t, g, a in waste:
            G, A = agg.get(t, (0, 0)); agg[t] = (G + g, A + a)
        ws = ' '.join(f'{t}waste={(g - a) / g * 100:.1f}%' for t, (g, a) in agg.items() if g)
        print(f"{rid}\tdps={p['dps']['avg']:.1f}\toom={p.get('secondsOomAvg', 0):.1f}s\tmanaEnd={p.get('manaEnd', {}).get('avg', 0):.0f}\t{ws}\tdead: {' '.join(dead) or '-'}")

main()
