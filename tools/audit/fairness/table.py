"""Usage: AUDIT_TREE=<repo checkout> python3 table.py <new sweep.log> <previous sweep.log> [label-of-previous] > table.md
Builds the cross-spec ranking (fairness report section 1) from two sweep.py logs: best race, best Horde / Alliance,
the report race's DPS against meta.published (the gap), and the change from the previous sweep. Variants are listed
after the table. Flags a row with '<--' on stderr when |gap| > 1.5, the report race isn't the best race, or the build moved."""
import json, os, sys
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
from races import ALLIANCE
REQ = os.path.join(os.environ["AUDIT_TREE"], "tools/dps_report/req")
NAME = {"warlock__demonology": "Demonology", "warrior__fury": "Fury (dual wield)", "hunter__lone-wolf": "Lone Wolf (hunter)", "hunter__marksmanship": "Marksmanship", "mage__frostfire": "Frostfire", "warlock__destruction": "Destruction", "warlock__affliction": "Affliction", "paladin__level_60": "Retribution", "warlock__destruction_fire": "Destruction Fire", "hunter__beast-mastery": "Beast Mastery", "warrior__arms": "Arms", "mage__frost": "Frost", "warrior__fury_2h": "Fury (two-hand)", "hunter__survival": "Survival", "mage__fire": "Fire", "priest__shadow": "Shadow", "mage__arcane": "Arcane", "druid__balance": "Balance", "shaman__enhancement-level60": "Enhancement", "druid__feral-cat": "Feral", "priest__smite": "Smite", "rogue__combat-swords": "Rogue Combat Swords", "rogue__combat-daggers": "Rogue Combat Daggers", "shaman__elemental-level60": "Elemental", "rogue__mutilate": "Rogue Mutilate", "rogue__seal-fate-daggers": "Rogue Seal Fate Daggers", "rogue__seal-fate-swords": "Rogue Seal Fate Swords", "rogue__subtlety": "Rogue Subtlety"}
RN = {"SkyborneWindshaper": "Windshaper", "SkyborneHighOrder": "High Order", "NightElf": "Night Elf"}
rn = lambda r: RN.get(r, r)
def load(f):
    D = {}
    for l in open(f):
        p = l.split()
        if len(p) >= 3: D[p[0]] = {p[k]: float(p[k + 1]) for k in range(1, len(p) - 1, 2)}
    return D
D = load(sys.argv[1]); P = load(sys.argv[2]); lab = sys.argv[3] if len(sys.argv) > 3 else "before"
def row(i):
    d = D[i]; best = max(d, key=d.get)
    H = [r for r in d if r not in ALLIANCE]; A = [r for r in d if r in ALLIANCE]
    h = max(H, key=d.get) if H else None; a = max(A, key=d.get) if A else None
    ha = ((f"{rn(h)} {d[h]:.1f}" if h else "none") + f" / {rn(a)}") if best in ALLIANCE else (f"{rn(h)} / " + (f"{rn(a)} {d[a]:.1f}" if a else "none"))
    m = json.load(open(f"{REQ}/{i}.meta.json")); pub = m["published"]["dps"]; pr = m["race"]
    return best, d[best], ha, pub, pr, d.get(pr), max(P[i].values()) if i in P else float("nan")
rows = sorted([(NAME[i],) + row(i) for i in NAME if i in D], key=lambda r: -r[2])
print(f"| # | Spec | Best race | DPS | Best Horde / Alliance | Report race (DPS here) | Report's published | Gap | Before ({lab}) |\n|---|---|---|---|---|---|---|---|---|")
for k, (n, b, v, ha, pub, pr, pv, bef) in enumerate(rows, 1):
    print(f"| {k} | {n} | {rn(b)} | {v:.1f} | {ha} | {rn(pr)} ({pv:.1f}) | {pub:.1f} | {pv-pub:+.1f} | {bef:.1f} ({100*(v-bef)/bef:+.1f}%) |")
    if abs(pv - pub) > 1.5 or b != pr and v - pv > 1.0 or abs(v - bef) >= 0.3:
        print(f"<-- {n}: gap {pv-pub:+.1f}, best {rn(b)} vs report {rn(pr)} ({v-pv:+.1f}), moved {v-bef:+.1f}", file=sys.stderr)
print("\n| Variant | Best race | DPS | Report's published (at its race) | Before |\n|---|---|---|---|---|")
for i in D:
    if i in NAME: continue
    b, v, ha, pub, pr, pv, bef = row(i)
    print(f"| {i} | {rn(b)} | {v:.1f} | {pub:.1f} (at {rn(pr)} {pv:.1f}) | {bef:.1f} |")
print(f"\nTop over bottom: {rows[0][2]/rows[-1][2]:.2f}x")
