"""Parry haste check against Classic's rule (audit 6.2). Usage: python3 -I parry.py LOG [FROM] [TO]

Based on the audit session's enemy/fit.py, with each mob's swing speed taken from the most
common interval between its own swings (rounded to 0.05 sec), not forced to 2.0.
"""
import sys, os, collections
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
import cl
evs = cl.load(sys.argv[1]); evs.sort(key=lambda e: e.t)
lo = sys.argv[2] if len(sys.argv) > 2 else '00'
hi = sys.argv[3] if len(sys.argv) > 3 else '99'
ME = None
swings = collections.defaultdict(list)
parried = collections.defaultdict(list)  # the unit that parried gets the haste
names = {}
for e in evs:
    if e.src.startswith(cl.ME): ME = e.srcGUID
# File order breaks ties: a swing logged after the parry in the same ms came after it.
for i, e in enumerate(evs):
    e.f.append(i)
for e in evs:
    if e.ev in ('SWING_DAMAGE', 'SWING_MISSED'):
        if e.srcGUID == ME or (e.srcGUID.startswith('Creature') and e.dstGUID == ME):
            swings[e.srcGUID].append((e.t, e.f[-1])); names[e.srcGUID] = e.src
    if e.ev in ('SWING_MISSED', 'SPELL_MISSED') and cl.miss(e) == 'PARRY' and lo <= e.ts < hi:
        if e.srcGUID == ME and e.dstGUID.startswith('Creature'): parried[e.dstGUID].append(e)
        elif e.dstGUID == ME and e.srcGUID.startswith('Creature'): parried[ME].append(e)
# Each mob's speed: the most common interval of its own, by GUID, then by name.
def mode_interval(ts):
    ts = [t for t, _ in ts]
    iv = [round((b - a) / 0.05) * 0.05 for a, b in zip(ts, ts[1:]) if 0.8 <= b - a <= 4.0]
    return collections.Counter(round(x, 2) for x in iv).most_common(1)[0][0] if iv else None
by_name = collections.defaultdict(list)
for g, ts in swings.items():
    if g != ME: by_name[names[g]] += [b[0] - a[0] for a, b in zip(ts, ts[1:])]
name_speed = {}
for n, iv in by_name.items():
    c = collections.Counter(round(round(x / 0.05) * 0.05, 2) for x in iv if 0.8 <= x <= 4.0)
    if c: name_speed[n] = c.most_common(1)[0][0]
def predict(speed, since):
    if speed - since <= 0.2 * speed + 1e-9: return speed
    return max(0.6 * speed, since + 0.2 * speed)
res = {'mob': [], 'us': []}
for unit, ps in parried.items():
    ss = sorted(swings.get(unit, []))
    for p in ps:
        prev = [t for t, i in ss if i < p.f[-1]]; nxt = [t for t, i in ss if i > p.f[-1]]
        if len(prev) < 2 or not nxt: continue
        if unit == ME: speed = prev[-1] - prev[-2]
        else: speed = name_speed.get(names[unit]) or mode_interval(ss)
        gap = nxt[0] - prev[-1]; since = p.t - prev[-1]
        if not speed or speed > 4 or gap > 4.5 or since > speed + 0.3: continue
        pr = predict(speed, since)
        res['us' if unit == ME else 'mob'].append((round(gap - pr, 2), p.ts, names.get(unit, 'us')[:20], round(speed, 2), round(since, 2), round(gap, 2), round(pr, 2)))
for k, v in res.items():
    d = [abs(x[0]) for x in v]
    print(f'{k}: {len(v)} parries, within 0.1 sec {sum(x <= 0.1 for x in d)}, largest miss {max(d, default=0)}')
    for x in v:
        if abs(x[0]) > 0.1: print('   off', x)
print('mob speeds', {n: s for n, s in name_speed.items() if any(names[g] == n for g in parried)})
