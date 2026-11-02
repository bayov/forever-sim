"""Usage: AUDIT_TREE=<repo checkout> [AUDIT_GIT_REF=origin/master] python3 audit1b.py
Gear audit (standards 1b + 5a + 5b) of the Level 60 synthetic gear sets, read with git show from AUDIT_GIT_REF.
Prints one line per set; every line should say "5a PASS | 1b PASS".

Bundled round audit (standards 1b + 5a + 5b) of the origin class heads.
5a: linear shares at 1b weights vs gear_targets_v3 (hit replaced by the set's actual hit share, the other
non-Sta/Spi stats rescaled to the room left; Sta +/-1.5, others +/-3; guardrail 55% of non-Sta).
1b: per-piece linear total = slot mod x L(ilvl) x f(archetype) +/-1%; 1.7-norm within +/-6% of slot mod x (0.661 ilvl + 2.92);
weapons / ranged / trinkets per 1b."""
import json, os, subprocess, sys
from collections import defaultdict
NAMES = "Str Agi Sta Int Spi SP Arc Fire Frost Holy Nat Shad MP5 SHit SCrit SHaste SPen AP MHit MCrit MHaste ArP Exp Mana En Rage Armor RAP Def Block BV Dodge Parry Res HP ArcR FirR FroR NatR ShaR BArm Heal SDmg FAP".split()
W = {"Str": 1, "Agi": 1, "Sta": 1, "Int": 1, "Spi": 1, "SP": .86, "SDmg": .86, "Arc": .70, "Fire": .70, "Frost": .70, "Holy": .70,
     "Nat": .70, "Shad": .70, "AP": .5, "RAP": .4, "MP5": 2.4, "Heal": .45, "FAP": .5}
G = {"SDmg": "SP", "Arc": "SP", "Fire": "SP", "Frost": "SP", "Holy": "SP", "Nat": "SP", "Shad": "SP", "RAP": "AP", "FAP": "AP"}
SETS = [  # (class, gear path, target name, archetype, weapon kind)
 ("warrior", "ui/warrior/gear_sets/level60_phase1_synthetic_fury.gear.json", "Warrior Fury", "phys", "1h"),
 ("warrior", "ui/warrior/gear_sets/level60_phase1_synthetic_arms.gear.json", "Warrior Arms / Fury 2H", "phys", "2h"),
 ("rogue", "ui/rogue/gear_sets/level60_phase1_synthetic_daggers.gear.json", "Rogue daggers", "phys", "1h"),
 ("rogue", "ui/rogue/gear_sets/level60_phase1_synthetic_swords.gear.json", "Rogue swords", "phys", "1h"),
 ("rogue", "ui/rogue/gear_sets/level60_phase1_synthetic_mutilate.gear.json", "Rogue daggers", "phys", "1h"),
 ("rogue", "ui/rogue/gear_sets/level60_phase1_synthetic_seal_fate_swords.gear.json", "Rogue swords", "phys", "1h"),
 ("hunter", "ui/hunter/gear_sets/level60_phase1_synthetic_ranged.gear.json", "Hunter ranged (BM / MM)", "hunter", "h2h"),
 ("hunter", "ui/hunter/gear_sets/level60_phase1_synthetic_survival.gear.json", "Hunter Survival", "hunter", "h2h"),
 ("paladin", "ui/retribution_paladin/gear_sets/level60_phase1_synthetic.gear.json", "Ret paladin", "hyb", "hyb2h"),
 ("shaman", "ui/enhancement_shaman/gear_sets/level60_phase1_synthetic.gear.json", "Enhancement shaman", "hyb", "hyb2h"),
 ("shaman", "ui/elemental_shaman/gear_sets/level60_phase1_synthetic.gear.json", "Elemental shaman", "cast", "cast"),
 ("druid", "ui/feral_druid/gear_sets/level60_phase1_synthetic.gear.json", "Feral druid", "phys", "2h"),
 ("druid", "ui/balance_druid/gear_sets/level60_phase1_synthetic.gear.json", "Balance druid (and Wrath)", "cast", "cast"),
 ("mage", "ui/mage/gear_sets/level60_phase1_synthetic.gear.json", "Mage (Frostfire / Frost / Fire)", "cast", "cast"),
 ("mage", "ui/mage/gear_sets/level60_phase1_synthetic_arcane.gear.json", "Mage Arcane", "cast", "cast"),
 ("warlock", "ui/warlock/gear_sets/level60_phase1_synthetic.gear.json", "Warlock (Destruction / Affliction / Fire)", "cast", "cast"),
 ("warlock", "ui/warlock/gear_sets/level60_phase1_synthetic_demonology.gear.json", "Warlock Demonology", "cast", "cast"),
 ("priest", "ui/shadow_priest/gear_sets/level60_phase1_synthetic.gear.json", "Shadow priest", "cast", "cast"),
 ("priest", "ui/shadow_priest/gear_sets/level60_phase1_smite.gear.json", "Smite priest", "cast", "cast"),
]
F = {"cast": (1.053, 1.026, .995), "phys": (.950, .959, .911), "hunter": (1.028, 1.056, .977), "hyb": (1.093, 1.103, 1.048)}
SLOTS = ["head","neck","shoulder","back","chest","wrist","hands","waist","legs","feet","fing1","fing2","trink1","trink2","mh","oh","ranged"]
TIER = {"head": 0, "chest": 0, "legs": 0, "shoulder": 1, "hands": 1, "waist": 1, "feet": 1, "neck": 2, "back": 2, "wrist": 2, "fing1": 2, "fing2": 2}
MOD = (1.0, .75, .5625)
L = lambda il: 1.125 * il + 2.0
N17 = lambda il: 0.661 * il + 2.92
TREE = os.environ["AUDIT_TREE"]
REF = os.environ.get("AUDIT_GIT_REF", "origin/master")
def show(cls, p): return subprocess.run(["git", "-C", TREE, "show", f"{REF}:{p}"], capture_output=True, text=True).stdout
T = {r["spec"]: r["target"] for r in json.loads(show(None, "orchestration/tools/gear_targets_v3.json"))}
def parts(it):
    st = it.get("stats", []) + [0] * 44; p = defaultdict(float)
    for i, v in enumerate(st[:44]):
        n = NAMES[i]
        if v and n in W and not (n == "RAP" and st[17] == v): p[G.get(n, n)] += W[n] * v
    h = it.get("hitRating", 0) + 10 * max(st[13], st[18]); c = it.get("critRating", 0) + 14 * max(st[14], st[19])
    if h: p["Hit"] += h
    if c: p["Crit"] += c
    return p
dbs = {}
for cls, path, name, arch, wk in SETS:
    if cls not in dbs: dbs[cls] = {i["id"]: i for i in json.loads(show(cls, "assets/database/db.json"))["items"]}
    db = dbs[cls]; gear = json.loads(show(cls, path))["items"]
    share = defaultdict(float); off = []; trk = []
    for i, e in enumerate(gear):
        if not e or "id" not in e or e["id"] not in db: continue
        it = db[e["id"]]; p = parts(it); lin = sum(p.values()); slot = SLOTS[i]
        if slot in ("trink1", "trink2"):
            trk.append(f"{it['name'][:28]} lin {lin:.1f}"); continue
        if not lin: continue
        for k, v in p.items(): share[k] += v
        il = it["ilvl"]
        if slot in TIER:
            t = TIER[slot]; want = MOD[t] * L(il) * F[arch][t]
            n17 = sum(v ** 1.7 for v in p.values()) ** (1 / 1.7); wn = MOD[t] * N17(il)
            if abs(lin / want - 1) > .01: off.append(f"{slot} lin {lin/want:.3f}")
            if abs(n17 / wn - 1) > .06: off.append(f"{slot} shape {n17/wn:.3f}")
        else:
            if slot == "ranged": want = (.26 if it.get("rangedWeaponType") == 5 or wk == "cast" else .24) * L(il)
            elif wk == "cast": want = (2.21 if it.get("handType") == 4 else (1.82 if slot == "mh" else .53)) * L(il)
            elif wk == "2h": want = .90 * L(il)
            elif wk == "h2h": want = (.98 if it.get("handType") == 4 else .42) * L(il)  # 1b a04b43a: hunter one-hander 0.42
            elif wk == "hyb2h": want = (1.035 if it.get("handType") == 4 else .39) * L(il)
            else: want = .39 * L(il)
            if abs(lin / want - 1) > .01: off.append(f"{slot} lin {lin/want:.3f}")
    tot = sum(share.values()); sh = {k: v / tot for k, v in share.items()}
    tg = {k: v / 100 for k, v in T[name].items()}
    held = ("SP", "Int") if arch == "hyb" else ()  # 5b: hybrids hold their SP and Int amounts
    hit = sh.get("Hit", 0); free = [k for k in tg if k not in ("Sta", "Spi", "Hit") + held]
    room = 1 - tg["Sta"] - tg.get("Spi", 0) - hit - sum(tg[k] for k in held); s = sum(tg[k] for k in free)
    adj = {k: tg[k] / s * room for k in free}; adj.update(Sta=tg["Sta"], Spi=tg.get("Spi", 0), Hit=hit, **{k: tg[k] for k in held})
    bad = [f"{k} {100*sh.get(k,0):.1f} vs {100*adj.get(k,0):.1f}" for k in set(adj) | set(sh)
           if abs(sh.get(k, 0) - adj.get(k, 0)) > (.015 if k == "Sta" else .03) + 1e-9]
    top = max((v for k, v in sh.items() if k != "Sta"), default=0) / (1 - sh.get("Sta", 0))
    if top > .55: bad.append(f"guardrail {100*top:.0f}%")
    print(f"{path.split('/')[-1][21:-10] or path.split('/')[1]:18s} {name[:20]:20s} 5a {'PASS' if not bad else 'FAIL ' + '; '.join(sorted(bad))} | 1b {'PASS' if not off else 'FAIL ' + ' '.join(off[:6]) + (' ...' if len(off) > 6 else '')} | trinkets: {', '.join(trk) or 'none'}")
