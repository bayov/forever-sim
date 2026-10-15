"""asm_data.py: the `assumptions` and `assumption_dps` datasets from asm/results.jsonl (run_assumptions.py).
build_data.py calls build(); `python3 asm_data.py` prints the sensitivities and the combination error."""
import json, math, os, re
from collections import defaultdict

sys_path = os.path.dirname(os.path.abspath(__file__))
import sys
sys.path.insert(0, sys_path)
from dashpaths import D  # noqa: E402
RES = os.path.join(D, 'asm', 'results.jsonl')
ROT_NAME = {'il': 'Ice Lance on Fingers of Frost', 'noil': 'never casts Ice Lance'}
CLASS_ORDER = ['All', 'Warrior', 'Rogue', 'Hunter', 'Paladin', 'Shaman', 'Druid', 'Mage', 'Warlock', 'Priest']


def load():
    """(build, combo) -> {secs: (dps, se)}, keeping the better rotation where an Ice Lance setting ran two."""
    best = {}
    for line in open(RES) if os.path.exists(RES) else []:
        x = json.loads(line)
        best.setdefault((x['build'], x['combo']), {}).setdefault(x['rot'], {})[x['secs']] = (x['dps'], x['se'], x['kind'])
    out = {}
    for k, by_rot in best.items():
        rot = max(by_rot, key=lambda r: by_rot[r].get(180, (-1,))[0])
        out[k] = (by_rot[rot], rot)
    return out


def predict(res, bid, combo, secs=180):
    base = res.get((bid, ''), ({},))[0].get(secs)
    if not base:
        return None
    f = 1.0
    for part in combo.split('|'):
        s = res.get((bid, part), ({},))[0].get(secs)
        if not s:
            return None
        f *= s[0] / base[0]
    return base[0] * f


def build(dump, builds=None):
    from assumptions_def import A
    res = load()
    names = {b['id']: b['build'] for b in (builds or [])}
    rows = []
    for (bid, combo), (by_secs, rot) in sorted(res.items()):
        r = {'build': bid, 'combo': combo, 'kind': next(iter(by_secs.values()))[2], 'rotation': ROT_NAME.get(rot, 'preset'),
             'dps180': by_secs.get(180, (None,))[0], 'se180': by_secs.get(180, (None, None))[1],
             'dps120': by_secs.get(120, (None,))[0], 'dps300': by_secs.get(300, (None,))[0]}
        rows.append(r)
    # sensitivities: each single setting against the build's base, at 180 s
    sens = defaultdict(list)
    for (bid, combo), (by_secs, rot) in res.items():
        if combo and '|' not in combo and 180 in by_secs:
            base = res.get((bid, ''), ({},))[0].get(180)
            if base:
                k, sid = combo.split('=')
                sens[k].append({'build': bid, 'name': names.get(bid, bid), 'setting': sid, 'pct': round(100 * (by_secs[180][0] / base[0] - 1), 2),
                                'dps': round(by_secs[180][0] - base[0], 1)})
    arows = []
    for i, a in enumerate(A):
        arows.append({'order': i, 'class': a['cls'], 'key': a['key'], 'label': a['label'], 'short': a.get('short') or a['desc'], 'description': a['desc'], 'source': a.get('src', ''), 'builds': a['builds'],
                      'settings': [{k: s[k] for k in ('id', 'label', 'value', 'rationale')} for s in a['settings']],
                      'defaultSetting': a['default'], 'expectedSetting': a['expected'], 'status': a.get('status', 'live'),
                      'sens': sorted(sens.get(a['key'], []), key=lambda x: (x['setting'], x['build']))})
    arows.sort(key=lambda r: (CLASS_ORDER.index(r['class']), r['order']))
    dump('assumptions', arows)
    dump('assumption_dps', rows)
    return arows, rows, res


def report():
    from assumptions_def import A
    res = load()
    errs = []
    for (bid, combo), (by_secs, rot) in res.items():
        if '|' in combo and 180 in by_secs:
            p = predict(res, bid, combo)
            if p:
                errs.append((100 * (p / by_secs[180][0] - 1), bid, combo, by_secs[180][2]))
    for kind in ('grid', 'check'):
        e = [abs(x[0]) for x in errs if x[3] == kind]
        if e:
            print(f'{kind}: {len(e)} combos, mean |error| {sum(e) / len(e):.2f}%, max {max(e):.2f}%')
    for x in sorted(errs, key=lambda x: -abs(x[0]))[:5]:
        print(f'  worst {x[0]:+.2f}% {x[1]} {x[2]}')


if __name__ == '__main__':
    def dump(name, rows):
        print(name, len(rows))
    arows, rows, res = build(dump)
    for a in arows:
        if a['sens']:
            lo = [s for s in a['sens'] if s['setting'] == 'low']
            hi = [s for s in a['sens'] if s['setting'] == 'high']
            ex = [s for s in a['sens'] if s['setting'] == 'expected']
            f = lambda xs: (f"{min(s['pct'] for s in xs):+.1f}..{max(s['pct'] for s in xs):+.1f}%" if xs else '-')
            print(f"{a['key']:24} low {f(lo):16} expected {f(ex):16} high {f(hi)}")
    report()
