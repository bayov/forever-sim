"""Q4 of the level 60 BiS budget analysis (orchestration/bis_budget_rules.md): trinkets.

A. Stat-only trinkets (Era blues/epics ilvl 55+, SoD phase 4-6 epics): p-norm value of the passive stats against
   the armor budget V_q(ilvl) of the same quality (Era armor fit, measured weights), i.e. the trinket's slot modifier,
   at p 1.5 and 1.7, and the weighted linear total against the Era linear trend (1.125 ilvl + 2.0).
B. On-use trinkets whose Use: is a plain stat buff: amount x duration / cooldown (steady state) and the same over a
   180 sec fight used on cooldown from time 0, as weighted stat points, against
   - the passive budget a stat-only trinket of the same ilvl and quality would carry (mod x V), and
   - the trinket's own passive stats, when it has some.

Usage: python3 bis_trinkets.py
"""
import math, os, re, sys
import numpy as np
sys.path.insert(0, os.path.dirname(__file__))
from bis_data import era60, sod60, clean, ARMOR
from bis_model import PModel, W_MEASURED, value, linear, med

WM = dict(W_MEASURED, **{'def': 1.56, 'dodge': 12.4, 'parry': 20.5, 'block': 4.5})
FIGHT = 180

# Use: text -> stat terms. 'decay' marks buffs that shrink (or grow) per action; their average over the
# duration is given by the factor (Zandalarian Hero Charm: -17 per cast from 204 over ~8 casts in 20 sec -> ~0.67;
# Jom Gabbar: +65 every 2 sec for 20 sec from 65 -> average 5.5 steps -> x 5.5).
PATTERNS = [
    (r'(?:damage and healing done by magical spells and effects|damage done by magical spells and effects) by up to (\d+)', 'sp'),
    (r'spell damage by up to (\d+)', 'sp'),
    (r'(?<!damage and )healing done by (?:magical )?spells and effects by up to (\d+)', 'heal'),
    (r'(?:melee and ranged )?attack power by (\d+)', 'ap'),
    (r'Attack Power by (\d+)', 'ap'),
    (r'Strength by (\d+)', 'str'),
    (r'chance to hit(?: with (?:all )?spells(?: and attacks)?)? by (\d+)%', 'hit%'),
    (r'critical strike chance(?: with spells and melee attacks)? by (\d+)%', 'crit%'),
]
SPECIAL = {"Zandalarian Hero Charm": 0.67, "Jom Gabbar": 5.5}


def parse_use(text):
    m = re.search(r'\((?:(\d+) Min)?(?:, )?(?:(\d+) Sec)? Cooldown\)', text)
    if not m or not (m.group(1) or m.group(2)):
        return None
    cd = 60 * int(m.group(1) or 0) + int(m.group(2) or 0)
    d = re.search(r'(?:for|Lasts|lasts for) (\d+) (sec|min)', text)
    if not d:
        return None
    dur = int(d.group(1)) * (60 if d.group(2) == 'min' else 1)
    terms = {}
    body = text.split('(')[0]
    for pat, k in PATTERNS:
        for x in re.findall(pat, body):
            if k in terms:
                continue
            if k == 'hit%':
                terms['hit'] = 10 * int(x)
            elif k == 'crit%':
                terms['crit'] = 14 * int(x)
            else:
                terms[k] = int(x)
    if 'sp' in terms and 'heal' in terms and 'spell damage by up to' in body:
        terms['heal'] -= terms['sp']   # "spell damage by up to N and your healing by up to M": healing on top
    return (terms, dur, cd) if terms else None


def uptime(dur, cd, T=FIGHT):
    n = math.floor(T / cd) + 1 if cd else 1
    return min(T, sum(min(dur, T - i * cd) for i in range(n) if i * cd < T)) / T


if __name__ == '__main__':
    E = era60()
    EA = [r for r in clean(E, ARMOR) if ((r['q'] == 3 and 45 <= r['ilvl'] <= 70) or (r['q'] == 4 and 55 <= r['ilvl'] <= 80)) and r['arch'] != 'stamina']
    fits = {}
    for p in (1.5, 1.7):
        M = PModel(EA, p, w=WM)
        th, mask = M.robust()
        fits[p] = M.unpack(th)[2]
    V = lambda q, L, p: fits[p][q][0] * L + fits[p][q][1]
    LIN = lambda L: 1.125 * L + 2.0
    S = sod60()
    TR = [r for r in E if r['slot'] == 'trinket' and r['ilvl'] >= 55] + [r for r in S if r['slot'] == 'trinket' and r['phase'] <= 6]

    print('### A. Stat-only trinkets (passive stats, no Use: or proc)')
    print('| source | item | q | ilvl | stats | n stats | value / V (p 1.5) | value / V (p 1.7) | linear / (1.125 ilvl + 2) |')
    print('|---|---|---|---|---|---|---|---|---|')
    mods = {1.5: [], 1.7: []}
    lins = []
    for r in sorted(TR, key=lambda r: (r['src'], r['ilvl'])):
        if not r['v'] or r['other'] - {'res'} or r['fx'].get('use') or r['fx'].get('proc') or r['fx'].get('equip'):
            continue
        q = 4 if r['q'] == 4 else 3
        m15 = value(r['v'], WM, 1.5) / V(q, r['ilvl'], 1.5)
        m17 = value(r['v'], WM, 1.7) / V(q, r['ilvl'], 1.7)
        li = linear(r['v'], WM) / LIN(r['ilvl'])
        if r['src'] == 'era' or True:
            mods[1.5].append(m15); mods[1.7].append(m17)
            if r['q'] == 4:
                lins.append(li)
        print(f'| {r["src"]} | {r["name"]} | {r["q"]} | {r["ilvl"]} | {r["v"]} | {len(r["v"])} | {m15:.2f} | {m17:.2f} | {li:.2f} |')
    print(f'\nmedian trinket modifier: p 1.5 {med(mods[1.5]):.3f} (n={len(mods[1.5])}), p 1.7 {med(mods[1.7]):.3f}; '
          f'epic linear total / trend median {med(lins):.2f} (n={len(lins)})')
    tmod = {p: med(mods[p]) for p in mods}

    print('\n### B. On-use stat trinkets')
    print('| source | item | q | ilvl | on-use | duration / cooldown | steady uptime | uptime in 180 s from 0 | average weighted points (180 s) | / passive budget (p 1.7) | passive stats | passive / budget |')
    print('|---|---|---|---|---|---|---|---|---|---|---|---|')
    ratios = []
    for r in sorted(TR, key=lambda r: (r['src'], r['ilvl'])):
        for u in r['fx'].get('use', []):
            pu = parse_use(u)
            if not pu:
                continue
            terms, dur, cd = pu
            f = SPECIAL.get(r['name'], 1.0)
            up_s, up_f = dur / cd, uptime(dur, cd)
            avg = f * up_f * linear(terms, WM)
            q = 4 if r['q'] == 4 else 3
            bud = tmod[1.7] * V(q, r['ilvl'], 1.7)
            pas = value(r['v'], WM, 1.7) / bud if r['v'] else 0.0
            ratios.append((r, avg / bud, pas, terms))
            print(f'| {r["src"]} | {r["name"]} | {r["q"]} | {r["ilvl"]} | {terms}{" x%.2f avg" % f if f != 1 else ""} | {dur} / {cd} s | {100*up_s:.0f}% | {100*up_f:.0f}% | '
                  f'{avg:.1f} | {avg/bud:.2f} | {r["v"] or "-"} | {pas:.2f} |')
    # Eye of Moam's resistance reduction is not priced, so it is left out of the medians
    ratios = [x for x in ratios if x[0]['name'] != 'Eye of Moam']
    only = [x[1] for x in ratios if x[2] == 0 and x[0]['q'] == 4]
    both = [x[1] + x[2] for x in ratios if x[2] > 0]
    print(f'\non-use average / passive budget, epics with no passive stats: median {med(only):.2f} (n={len(only)}, range {min(only):.2f}-{max(only):.2f}); '
          f'items with passive stats too, passive + on-use: median {med(both):.2f} (n={len(both)})')
    for kind_ in ('sp', 'ap'):
        x = [v for r, v, pas, t in ratios if kind_ in t and pas == 0]
        if x:
            print(f'   {kind_} on-use only: median {med(x):.2f} (n={len(x)})')
