#!/usr/bin/env python3
"""Stochastic talent climb with hit-trade gear steps (round 70753 phase 5, rogue).

Each round: N random valid neighbours (1 to 5 points moved; a share of them always touch
Precision), scored by rotopt -talents-file at 3000 iterations over 120/180/300 s combined with
the hit-step flags; the best 5 plus the current build are scored again at 10000, and the best
is kept when it beats the current one by 2 standard errors of the difference. Stops after two
rounds in a row keep nothing (or MAXR rounds).

usage: climb.py <id> <template> <knobs> <gearbase> <pins comma> <start> <outdir>
"""
import json, math, os, random, re, subprocess, sys

ROT = os.path.join(os.path.dirname(os.path.abspath(__file__)), "rotopt")
bid, template, knobs, gearbase, pins, start, outdir = sys.argv[1:8]
pins = [p for p in pins.split(",") if p]
N = int(os.environ.get("N", "80"))
MAXR = int(os.environ.get("MAXR", "10"))
random.seed(int(os.environ.get("SEED", "1")))

trees = json.load(open("ui/core/talents/trees/rogue.json"))
maxp = [[t["maxPoints"] for t in tr["talents"]] for tr in trees]
names = [[t["fieldName"] for t in tr["talents"]] for tr in trees]
rows = [[t["location"]["rowIdx"] for t in tr["talents"]] for tr in trees]
nosim = [[bool(t.get("notSimulated")) for t in tr["talents"]] for tr in trees]
prq = []
for tr in trees:
    loc = {(t["location"]["rowIdx"], t["location"]["colIdx"]): j for j, t in enumerate(tr["talents"])}
    prq.append([loc[(t["prereqLocation"]["rowIdx"], t["prereqLocation"]["colIdx"])] if t.get("prereqLocation") else -1
                for t in tr["talents"]])
flat = [(ti, i) for ti in range(3) for i in range(len(maxp[ti]))]
pinned = {(ti, i) for (ti, i) in flat if names[ti][i] in pins}
prec = next((ti, i) for (ti, i) in flat if names[ti][i] == "precision")


def parse(s):
    parts = s.split("-") + [""] * 3
    return [[int(parts[ti][i]) if i < len(parts[ti]) else 0 for i in range(len(maxp[ti]))] for ti in range(3)]


def valid(b):
    if sum(map(sum, b)) > 51:
        return False
    for ti in range(3):
        rp = [0] * 10
        for i, p in enumerate(b[ti]):
            if p < 0 or p > maxp[ti][i]:
                return False
            if (ti, i) in pinned and p < maxp[ti][i]:
                return False
            if p > 0 and prq[ti][i] >= 0 and b[ti][prq[ti][i]] < maxp[ti][prq[ti][i]]:
                return False
            rp[rows[ti][i]] += p
        above = 0
        for r, p in enumerate(rp):
            if p > 0 and above < 5 * r:
                return False
            above += p
    return True


def fmt(b):
    out = ["".join(map(str, t)).rstrip("0") for t in b]
    while out and out[-1] == "":
        out.pop()
    return "-".join(out)


def neighbour(b, touch_prec):
    c = [t[:] for t in b]
    k = random.randint(1, 5)
    for step in range(k):
        srcs = [(ti, i) for (ti, i) in flat if c[ti][i] > 0 and (ti, i) not in pinned]
        dsts = [(ti, i) for (ti, i) in flat if c[ti][i] < maxp[ti][i] and not nosim[ti][i]]
        if touch_prec and step == 0 and random.random() < 0.5 and c[prec[0]][prec[1]] > 0:
            src = prec
        else:
            src = random.choice(srcs)
        if touch_prec and step == 0 and src != prec and c[prec[0]][prec[1]] < 3:
            dst = prec
        else:
            dst = random.choice(dsts)
        if src == dst:
            continue
        c[src[0]][src[1]] -= 1
        c[dst[0]][dst[1]] += 1
    return c


def score(strings, iters, tag):
    path = os.path.join(outdir, f"{bid}_{tag}.txt")
    with open(path, "w") as f:
        f.write("\n".join(strings) + "\n")
    cmd = [ROT, "-template", template, "-settings", f"tools/dps_report/req/{bid}.json", "-knobs", knobs,
           "-combined", "-durs", "120,180,300", "-iters", str(iters), "-talents-file", path,
           "-hit-steps", gearbase, "-hit-cap", "melee:8", "-hit-span", "2"]
    out = subprocess.run(cmd, capture_output=True, text=True).stdout
    res = {}
    for line in out.splitlines():
        m = re.match(r"^([\d.]+)\(([\d.]+)\)\s+(\S+)$", line.strip())
        if m:
            res[m.group(3)] = (float(m.group(1)), float(m.group(2)))
    return res


log = open(os.path.join(outdir, f"{bid}_climb.log"), "a")


def say(*a):
    print(*a, file=log, flush=True)


cur = fmt(parse(start))
say(f"start {cur} pins={pins} N={N}")
fails, rnd = 0, 0
seen = {cur}
while fails < 2 and rnd < MAXR:
    rnd += 1
    b = parse(cur)
    cands = set()
    tries = 0
    while len(cands) < N and tries < N * 2000:
        tries += 1
        nb = neighbour(b, random.random() < 0.3)
        s = fmt(nb)
        if valid(nb) and s not in seen:
            cands.add(s)
    r = score(sorted(cands), 3000, f"r{rnd}")
    seen |= cands
    top = sorted(r.items(), key=lambda kv: -kv[1][0])[:5]
    say(f"round {rnd}: {len(r)} valid of {len(cands)}; top3000 " + "; ".join(f"{s} {d:.1f}" for s, (d, e) in top))
    r2 = score([cur] + [s for s, _ in top], 10000, f"r{rnd}c")
    c0 = r2.get(cur)
    best = max(((s, v) for s, v in r2.items() if s != cur), key=lambda kv: kv[1][0], default=None)
    say(f"  confirm: current {cur} {c0}; best {best}")
    if best and c0 and best[1][0] - c0[0] > 2 * math.hypot(best[1][1], c0[1]):
        cur = best[0]
        fails = 0
        say(f"  KEEP {cur} (+{best[1][0]-c0[0]:.1f})")
    else:
        fails += 1
say(f"final {cur}")
print(cur)
