"""Usage: AUDIT_TREE=<repo checkout> AUDIT_WORK=<scratch dir> [ITERS=5000] python3 sweep.py [request ids...] > sweep.log
Sims the DPS report's own base requests (AUDIT_TREE/tools/dps_report/req/<id>.json) for every race the class can
play, changing only the race (seed 1, ITERS iterations). Duty, vs-type (humanoid / undead / dragonkin) requests are
skipped. Needs $AUDIT_WORK/wowsimcli (build.sh). Results are cached in $AUDIT_WORK/runs keyed on request + binary.
Output: one line per request: '<id>  Race dps  Race dps ...' sorted best first."""
import json, glob, os, sys, hashlib, subprocess
from concurrent.futures import ThreadPoolExecutor
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
from races import RACES
TREE = os.environ["AUDIT_TREE"]; WORK = os.environ["AUDIT_WORK"]
REQ = os.path.join(TREE, "tools/dps_report/req"); OUT = os.path.join(WORK, "runs"); os.makedirs(OUT, exist_ok=True)
BIN = os.path.join(WORK, "wowsimcli"); ITERS = int(os.environ.get("ITERS", "5000"))
def ids():
    out = []
    for f in sorted(glob.glob(f"{REQ}/*.meta.json")):
        i = json.load(open(f))["id"]
        if "duty" in i or i.endswith("humanoid") or "undead" in i or "dragonkin" in i: continue
        out.append(i)
    return out
def run(i, race):
    req = json.load(open(f"{REQ}/{i}.json")); p = req["raid"]["parties"][0]["players"][0]
    p["race"] = "Race" + race; req["simOptions"] = {"iterations": ITERS, "randomSeed": "1"}
    h = hashlib.md5((json.dumps(req, sort_keys=True) + str(os.path.getmtime(BIN))).encode()).hexdigest()[:12]
    inf = f"{OUT}/{i}_{h}.in.json"; outf = f"{OUT}/{i}_{h}.out.json"
    if not os.path.exists(outf):
        json.dump(req, open(inf, "w"))
        if subprocess.run([BIN, "sim", "--infile", inf, "--outfile", outf], capture_output=True).returncode: return race, None
    return race, json.load(open(outf))["raidMetrics"]["dps"]["avg"]
if __name__ == "__main__":
    for i in sys.argv[1:] or ids():
        with ThreadPoolExecutor(4) as ex: res = list(ex.map(lambda r: run(i, r), RACES[i.split("__")[0]].split()))
        d = {r: v for r, v in res if v}
        print(f"{i:38s} " + "  ".join(f"{r} {d[r]:.1f}" for r in sorted(d, key=lambda r: -d[r])), flush=True)
