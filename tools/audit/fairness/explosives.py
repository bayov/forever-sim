"""Usage: AUDIT_TREE=<checkout> AUDIT_WORK=<scratch> [ITERS=10000] python3 explosives.py <request id>...
Paired check of the filler explosive: sims each report request with fillerExplosive none / Goblin Land Mine /
Dense Dynamite (only that changed), seeds 1 and 1000003, and prints the mean DPS and the delta over none.
Needs $AUDIT_WORK/wowsimcli (build.sh)."""
import json, os, subprocess, sys
TREE = os.environ["AUDIT_TREE"]; WORK = os.environ["AUDIT_WORK"]; ITERS = int(os.environ.get("ITERS", "10000"))
BIN = os.path.join(WORK, "wowsimcli"); OUT = os.path.join(WORK, "explosives"); os.makedirs(OUT, exist_ok=True)

def find_consumes(p):
    for k in ("consumables", "consumes"):
        if k in p: return p[k]
    p["consumables"] = {}; return p["consumables"]

def sim(rid, expl, seed):
    req = json.load(open(os.path.join(TREE, "tools/dps_report/req", rid + ".json")))
    p = req["raid"]["parties"][0]["players"][0]
    c = find_consumes(p)
    if expl: c["fillerExplosive"] = expl
    else: c.pop("fillerExplosive", None)
    req["simOptions"] = {"iterations": ITERS, "randomSeed": str(seed)}
    base = os.path.join(OUT, f"{rid}.{expl or 'none'}.{seed}")
    json.dump(req, open(base + ".in.json", "w"))
    subprocess.run([BIN, "sim", "--infile", base + ".in.json", "--outfile", base + ".out.json"], check=True,
                   stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)
    r = json.load(open(base + ".out.json"))
    return r["raidMetrics"]["dps"]["avg"]

for rid in sys.argv[1:]:
    res = {}
    for e in ("", "ExplosiveGoblinLandMine", "ExplosiveDenseDynamite"):
        res[e] = sum(sim(rid, e, s) for s in (1, 1000003)) / 2
    n = res[""]
    print(f"{rid:40s} none {n:7.1f}  mine {res['ExplosiveGoblinLandMine']-n:+5.1f}  dynamite {res['ExplosiveDenseDynamite']-n:+5.1f}", flush=True)
