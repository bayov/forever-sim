#!/usr/bin/env python3
"""Per-spec 5a stat-mix targets for the bundled gear round (standards 1b, 2026-10-13).

Reads the SoD level 60 BiS sets (wowsims/sod ui gear_sets, trinkets excluded, AP and RAP counted once, SoD weapon
feral AP excluded) and prices them with the Forever weights of standards 1b: spell power 0.86, school-only power 0.70,
AP 0.5, RAP 0.4, hit and crit rating 1.0 (10 / 14 per 1%), MP5 2.4, primaries 1.
Target per spec: Stamina = the spec's SoD average share; Spirit 10% for mana users (standards 5); the other stats in
the SoD average proportions, re-normalised to the rest. Rogue swords' Agility is clipped at the 45% guardrail.
Ret and Enhancement (standards 5b): Int and spell power keep their amounts relative to the set from the 5b targets
(spell power re-priced from 0.55 to 0.86), the rest in SoD proportions. Smite: shadow priest proportions without hit.
Usage: gear_targets_v3.py [SOD_UI_DIR] [SOD_DB]
"""
import glob, json, os, statistics, sys

SOD_UI = sys.argv[1] if len(sys.argv) > 1 else '/home/user/wowsims/sod/ui/'
SOD_DB = sys.argv[2] if len(sys.argv) > 2 else '/home/user/wowsims/sod/assets/database/db.json'
DB = {i['id']: i for i in json.load(open(SOD_DB))['items']}
W = {'Sta': 1.0, 'Str': 1.0, 'Agi': 1.0, 'Int': 1.0, 'Spi': 1.0, 'SP': 0.86, 'SchoolSP': 0.70, 'Hit': 1.0, 'Crit': 1.0, 'MP5': 2.4}
COLS = ['Sta', 'Str', 'Agi', 'Int', 'Spi', 'AP', 'SP', 'Hit', 'Crit', 'MP5']


def pts(it):
    s = (it.get('stats') or []) + [0] * 44
    p = dict.fromkeys(COLS, 0.0)
    p['Str'], p['Agi'], p['Sta'], p['Int'], p['Spi'] = s[0], s[1], s[2], s[3], s[4]
    p['SP'] = (s[5] + s[42]) * W['SP'] + sum(s[6:12]) * W['SchoolSP']  # weighted; school power folded into SP
    ap, rap = s[17], s[27]
    p['AP'] = max(ap * 0.5, rap * 0.4) if ap and rap else ap * 0.5 + rap * 0.4  # weighted, counted once
    p['Hit'] = it.get('hitRating') or 10 * max(s[18], s[13])
    p['Crit'] = it.get('critRating') or 14 * max(s[19], s[14])
    p['MP5'] = s[12] * W['MP5']
    return p


def shares(path):
    tot = dict.fromkeys(COLS, 0.0)
    for slot, e in enumerate(json.load(open(path)).get('items', [])):
        if not e or not e.get('id') or slot in (12, 13) or e['id'] not in DB:
            continue
        for c, v in pts(DB[e['id']]).items():
            tot[c] += v
    t = sum(tot.values()) or 1
    return {c: 100 * v / t for c, v in tot.items()}


def avg(files):
    rows = [shares(os.path.join(SOD_UI, f)) for f in files if os.path.exists(os.path.join(SOD_UI, f))]
    return {c: statistics.mean(r[c] for r in rows) for c in COLS}, len(rows)


G = lambda d, fs: [f'{d}/gear_sets/{f}' for f in fs]
SPECS = [
    ('Warrior Fury', G('warrior', ['phase_4_dw.gear.json', 'phase_5_dw_t1.gear.json', 'phase_5_dw_t2.gear.json', 'phase_6_dw.gear.json']), False),
    ('Warrior Arms / Fury 2H', G('warrior', ['phase_4_2h.gear.json', 'phase_5_2h_t1.gear.json', 'phase_5_2h_t2.gear.json', 'phase_6_2h.gear.json']), False),
    ('Rogue daggers', G('rogue', ['p4_muti.gear.json', 'p5_backstab.gear.json', 'p5_mutilate.gear.json', 'p6_backstab.gear.json', 'p6_mutilate.gear.json']), False),
    ('Rogue swords', G('rogue', ['p4_saber.gear.json', 'p5_saber.gear.json', 'p6_saber.gear.json']), False),
    ('Hunter ranged (BM / MM)', G('hunter', ['p4_ranged.gear.json', 'p5_ranged_sv.gear.json', 'p6_ranged_killshot.gear.json', 'p6_ranged_draconic.gear.json', 'p7_ranged_killshot.gear.json', 'p7_ranged_lnl.gear.json']), True),
    ('Hunter Survival', G('hunter', ['p5_melee_sv.gear.json', 'p6_melee_sv.gear.json', 'p7_melee_2h.gear.json', 'p7_melee_dw.gear.json']), True),
    ('Ret paladin', G('retribution_paladin', ['p4-twist.gear.json', 'p5-twisting.gear.json', 'p5-shockadin.gear.json', 'p6-twisting.gear.json', 'p7-twisting-naxx.gear.json', 'p8-twist.gear.json', 'p8-shockadin-2h.gear.json']), True),
    ('Enhancement shaman', G('enhancement_shaman', ['phase_4_2h.gear.json', 'phase_4_dw.gear.json', 'phase_5_2h.gear.json', 'phase_5_dw.gear.json', 'phase_6_2h.gear.json', 'phase_6_dw.gear.json']), True),
    ('Elemental shaman', G('elemental_shaman', ['phase_4.gear.json', 'phase_5.gear.json', 'phase_6.gear.json']), True),
    ('Feral druid', G('feral_druid', ['phase_4.gear.json', 'phase_5.gear.json']), False),
    ('Balance druid (and Wrath)', G('balance_druid', ['phase_4.gear.json', 'phase_5.gear.json', 'phase_6.gear.json']), True),
    ('Mage (Frostfire / Frost / Fire)', G('mage', ['p4_fire.gear.json', 'p4_frost.gear.json', 'p5_fire.gear.json', 'p5_frost.gear.json']), True),
    ('Mage Arcane', G('mage', ['p4_arcane.gear.json', 'p5_arcane.gear.json']), True),
    ('Warlock (Destruction / Affliction / Fire)', G('warlock', ['p4/destruction.gear.json', 'p4/affliction.gear.json', 'p5/destruction.gear.json', 'p5/affliction.gear.json', 'p6/destruction.gear.json', 'p6/affliction.gear.json']), True),
    ('Warlock Demonology', G('warlock', ['p5/demonology.gear.json', 'p6/demonology.gear.json']), True),
    ('Shadow priest', G('shadow_priest', ['phase_4.gear.json', 'phase_5_t1.gear.json', 'phase_5_t2.gear.json', 'phase_6.gear.json']), True),
    ('Smite priest', G('shadow_priest', ['phase_4.gear.json', 'phase_5_t1.gear.json', 'phase_5_t2.gear.json', 'phase_6.gear.json']), True),
]
# standards 5b hybrid targets (spell power share at 0.55, set total at 0.55): Int, SP
HYBRID_5B = {'Ret paladin': (11.1, 20.2), 'Enhancement shaman': (11.7, 22.1)}

out = []
for name, files, mana in SPECS:
    a, n = avg(files)
    if name == 'Smite priest':
        a['Hit'] = 0.0
    sta = a['Sta']
    spi = 10.0 if mana else 0.0
    fixed = {'Sta': sta, 'Spi': spi}
    if name in HYBRID_5B:
        i55, sp55 = HYBRID_5B[name]
        # same amounts, re-priced: SP points x 0.86/0.55; the set total grows by the same extra points
        grow = 1 + sp55 * (0.86 / 0.55 - 1) / 100
        fixed['Int'] = i55 / grow
        fixed['SP'] = sp55 * 0.86 / 0.55 / grow
    rest = [c for c in COLS if c not in fixed]
    room = 100 - sum(fixed.values())
    base = sum(a[c] for c in rest) or 1
    t = dict(fixed, **{c: a[c] * room / base for c in rest})
    if name == 'Rogue swords':
        cap = 0.45 * (100 - sta)
        if t['Agi'] > cap:
            extra = t['Agi'] - cap
            t['Agi'] = cap
            others = [c for c in rest if c != 'Agi' and t[c] > 0]
            s = sum(t[c] for c in others)
            for c in others:
                t[c] += extra * t[c] / s
    out.append({'spec': name, 'lists': n, 'sod': a, 'target': t})
    print(f"{name:42s} n={n} " + ' '.join(f"{c}={t[c]:.1f}" for c in COLS))
json.dump(out, open(os.path.join(os.path.dirname(os.path.abspath(__file__)), 'gear_targets_v3.json'), 'w'), indent=1)
