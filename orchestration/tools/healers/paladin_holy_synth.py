"""A synthetic Phase 1 Holy paladin set on standards 1b (pure caster) at the healer row's shares
(orchestration/gear_targets.md, Healers), shaman_restoration_synth.py with plate pieces and a caster main
hand and held off-hand in place of the staff (paladins have no staves)."""
import itertools, json, os
W = {'Sta': 1, 'Int': 1, 'Spi': 1, 'SP': 0.86, 'Heal': 0.45, 'Crit': 1, 'MP5': 2.4}
TARGET = {'Sta': 14.7, 'Int': 18.7, 'Spi': 10.0, 'SP': 21.9, 'Heal': 26.8, 'Crit': 2.8, 'MP5': 5.1}
L = lambda ilvl: 1.125 * ilvl + 2.0
F = {1: 1.053, 0.75: 1.026, 0.5625: 0.995}
SHAPE = lambda mod, ilvl: mod * (0.661 * ilvl + 2.92)
# slot: (mod, ilvl, lines)
PIECES = {
    'Head':      (1, 66, ['Sta', 'Int', 'Heal', 'Crit']),
    'Neck':      (0.5625, 71, ['Sta', 'Spi', 'Heal', 'MP5']),
    'Shoulder':  (0.75, 66, ['Sta', 'Int', 'Spi', 'Heal']),
    'Back':      (0.5625, 71, ['Sta', 'Int', 'SP', 'Heal']),
    'Chest':     (1, 66, ['Sta', 'Int', 'Spi', 'Heal']),
    'Wrist':     (0.5625, 71, ['Int', 'SP', 'Heal', 'MP5']),
    'Hands':     (0.75, 66, ['Sta', 'Int', 'SP', 'Heal']),
    'Waist':     (0.75, 71, ['Sta', 'Spi', 'SP', 'Heal']),
    'Legs':      (1, 66, ['Int', 'SP', 'Heal', 'MP5']),
    'Feet':      (0.75, 66, ['Sta', 'Int', 'Spi', 'SP']),
    'Finger':    (0.5625, 71, ['Sta', 'Int', 'SP', 'Heal']),
    'Finger II': (0.5625, 71, ['Int', 'SP', 'Heal', 'MP5']),
}
# The weapons: a caster main hand (71) at 1.82 * L(71) = 149.1 and a held off-hand (71) at 0.53 * L(71) =
# 43.4 (standards 1b), taken as one 192.5 here and split by their totals below.
STAFF = (149.1 + 43.4, ['Sta', 'Int', 'Spi', 'SP', 'Heal', 'Crit', 'MP5'])

def lin_target(mod, ilvl):
    return mod * L(ilvl) * F[mod]

def norm(vals):
    return sum(v ** 1.7 for v in vals) ** (1 / 1.7)

# Weighted amounts, IPF: stat factors to hit the set shares, each piece scaled to its linear target.
pieces = {k: {s: 1.0 for s in v[2]} for k, v in PIECES.items()}
pieces['Staff'] = {s: 1.0 for s in STAFF[1]}
targets = {k: lin_target(v[0], v[1]) for k, v in PIECES.items()}
targets['Staff'] = STAFF[0]
total = sum(targets.values())
want = {s: TARGET[s] / 100 * total for s in TARGET}
for _ in range(500):
    for k, p in pieces.items():
        t = sum(p.values())
        for s in p:
            p[s] *= targets[k] / t
    have = {s: sum(p.get(s, 0) for p in pieces.values()) for s in TARGET}
    for p in pieces.values():
        for s in p:
            p[s] *= (want[s] / have[s]) ** 0.5
for k, p in pieces.items():
    t = sum(p.values())
    for s in p:
        p[s] *= targets[k] / t

# The shape check: a piece too flat moves weight from its smallest line to its largest, one too peaked
# the other way, aiming at the middle of the band (the target itself). Its linear total stays.
for k, (mod, ilvl, _) in PIECES.items():
    p = pieces[k]
    sh = SHAPE(mod, ilvl)
    for _ in range(400):
        n = norm(p.values())
        if abs(n / sh - 1) < 0.02:
            break
        big, small = max(p, key=p.get), min(p, key=p.get)
        step = 0.2
        if n < sh:
            p[big] += step; p[small] -= step
        else:
            p[small] += step; p[big] -= step
# The staff takes what the set still needs of each stat, scaled back to its own total.
armor = {s: sum(p.get(s, 0) for k, p in pieces.items() if k != 'Staff') for s in TARGET}
need = {s: max(0.5, want[s] - armor[s]) for s in TARGET}
t = sum(need.values())
pieces['Staff'] = {s: v * targets['Staff'] / t for s, v in need.items()}

# Round to whole stat points, then nudge each piece back inside 1% of its target.
def amounts(p):
    return {s: max(1, round(v / W[s])) for s, v in p.items()}
def lin(a):
    return sum(W[s] * n for s, n in a.items())
out = {}
for k, p in pieces.items():
    a = amounts(p)
    t = targets[k]
    for _ in range(50):
        d = lin(a) - t
        if abs(d) <= 0.004 * t:
            break
        s = max(a, key=lambda s: p[s]) if abs(d) > 1.5 else min(a, key=lambda s: W[s])
        s = 'Heal' if 'Heal' in a and abs(d) < 3 else s
        a[s] += -1 if d > 0 else 1
    out[k] = a

print(f'{"piece":<10} {"target":>6} {"linear":>7} {"norm":>6} {"shape":>6}  stats')
for k, a in out.items():
    t = targets[k]
    l = lin(a)
    line = f'{k:<10} {t:6.1f} {l:6.1f} ({100*(l/t-1):+.1f}%)'
    if k != 'Staff':
        mod, ilvl, _ = PIECES[k]
        n, sh = norm([W[s] * v for s, v in a.items()]), SHAPE(mod, ilvl)
        line += f' {n:5.1f} / {sh:5.1f} ({100*(n/sh-1):+.1f}%)'
    else:
        line += ' ' * 24
    print(line + '  ' + ', '.join(f'{v} {s}' for s, v in a.items()))
tot = {s: sum(a.get(s, 0) for a in out.values()) for s in TARGET}
T = sum(W[s] * v for s, v in tot.items())
print(f'set total {T:.1f} (armor {sum(lin(a) for k,a in out.items() if k!="Staff"):.1f}, staff {lin(out["Staff"]):.1f})')
for s in TARGET:
    print(f'  {s:<5} {tot[s]:>4}  {100*W[s]*tot[s]/T:5.1f}%  target {TARGET[s]}')
json.dump(out, open(os.devnull, 'w'), indent=1)

# The main hand and the off-hand split the weapons' stats: the held off-hand 43.4 in Stamina, Intellect,
# Spirit and healing, the main hand the rest (149.1 within 1%).
w = out['Staff']
oh = {'Sta': 9, 'Int': 8, 'Spi': 6, 'Heal': 45}
mh = {k: v - oh.get(k, 0) for k, v in w.items()}
print('main hand', ', '.join(f'{v} {s}' for s, v in mh.items()), f'= {lin(mh):.1f} (149.1)')
print('off-hand', ', '.join(f'{v} {s}' for s, v in oh.items()), f'= {lin(oh):.1f} (43.4)')
