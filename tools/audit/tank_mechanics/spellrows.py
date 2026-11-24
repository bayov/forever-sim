#!/usr/bin/env python3
"""Print a spell's client rows (name, effects, class options, cooldowns, misc) from $AUDIT_DB.
Usage: spellrows.py <spellid>... ; masks: spellrows.py --mask <spellid> lists which given ids an effect's mask covers."""
import csv, os, sys
DB = os.environ["AUDIT_DB"]
def load(t):
    with open(os.path.join(DB, t + ".csv"), newline="") as f:
        return list(csv.DictReader(f))
T = {t: load(t) for t in ["SpellName", "SpellEffect", "SpellClassOptions", "SpellCooldowns", "SpellMisc", "SpellAuraOptions", "SpellPower", "SpellCategories"]}
def rows(t, sid, key="SpellID"):
    return [r for r in T[t] if r.get(key) == str(sid)]
def short(r, keep):
    return {k: v for k, v in r.items() if k in keep or (v not in ("0", "", "0.0") and k not in ("ID",))}
def mask_of(r):
    return [int(r.get(f"EffectSpellClassMask_{i}", 0) or 0) for i in range(4)]
def cls(sid):
    c = rows("SpellClassOptions", sid)
    return (int(c[0]["SpellClassSet"]), [int(c[0][f"SpellClassMask_{i}"]) for i in range(4)]) if c else None
args = sys.argv[1:]
if args and args[0] == "--mask":
    eff_sid, targets = args[1], args[2:]
    for e in sorted(rows("SpellEffect", eff_sid), key=lambda r: int(r["EffectIndex"])):
        m = mask_of(e)
        hits = []
        for t in targets:
            c = cls(t)
            if c and any(a & b for a, b in zip(m, c[1])):
                hits.append(t)
        print(f"eff {e['EffectIndex']} aura {e['EffectAura']} misc {e['EffectMiscValue_0']} bp {e['EffectBasePointsF']} mask {m} covers {hits}")
    sys.exit()
for sid in args:
    n = rows("SpellName", sid, "ID")
    print(f"== {sid} {n[0]['Name_lang'] if n else '?'} class={cls(sid)}")
    for e in sorted(rows("SpellEffect", sid), key=lambda r: int(r["EffectIndex"])):
        print("  eff", e["EffectIndex"], "type", e["Effect"], "aura", e["EffectAura"], "bp", e["EffectBasePointsF"], "misc", e["EffectMiscValue_0"], "trig", e["EffectTriggerSpell"], "mask", mask_of(e), "coef", e.get("EffectBonusCoefficient"), "apcoef", e.get("BonusCoefficientFromAP"))
    for t in ["SpellCooldowns", "SpellAuraOptions", "SpellPower", "SpellCategories"]:
        for r in rows(t, sid):
            print("  ", t, short(r, ()))
    for r in rows("SpellMisc", sid):
        print("   SpellMisc dur", r.get("DurationIndex"), "attr0-3", [r.get(f"Attributes_{i}") for i in range(4)])
