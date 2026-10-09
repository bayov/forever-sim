#!/usr/bin/env python3
"""Print spells from the Forever client data on wago.tools.

Usage: tools/wago/spell.py [--build 1.60.1.70291] [--cache DIR] <name regex or spell ID>...

We download each DB2 table once as CSV (https://wago.tools/db2/<Table>/csv?build=<build>)
and keep it in the cache directory. Forever's SpellEffect has the real value in
EffectBasePointsF. Classic Era builds store EffectBasePoints as the value minus 1 instead.
"""
import argparse
import csv
import os
import re
import urllib.request

TABLES = ['SpellName', 'SpellEffect', 'SpellMisc', 'SpellDuration', 'SpellCastTimes', 'SpellCooldowns']


def load(cache, build, table):
    path = os.path.join(cache, f'{table}_{build.split(".")[-1]}.csv')
    if not os.path.exists(path):
        url = f'https://wago.tools/db2/{table}/csv?build={build}'
        req = urllib.request.Request(url, headers={'User-Agent': 'Mozilla/5.0'})
        with urllib.request.urlopen(req) as resp, open(path, 'wb') as out:
            out.write(resp.read())
    with open(path, newline='', encoding='utf-8') as f:
        return list(csv.DictReader(f))


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument('--build', default='1.60.1.70291')
    ap.add_argument('--cache', default=os.path.join(os.path.expanduser('~'), '.cache', 'wago'))
    ap.add_argument('queries', nargs='+')
    args = ap.parse_args()
    os.makedirs(args.cache, exist_ok=True)
    t = {name: load(args.cache, args.build, name) for name in TABLES}

    names = {int(r['ID']): r['Name_lang'] for r in t['SpellName']}
    effects = {}
    for r in t['SpellEffect']:
        effects.setdefault(int(r['SpellID']), []).append(r)
    misc = {int(r['SpellID']): r for r in t['SpellMisc']}
    durations = {r['ID']: r['Duration'] for r in t['SpellDuration']}
    casts = {r['ID']: r['Base'] for r in t['SpellCastTimes']}
    cooldowns = {int(r['SpellID']): r for r in t['SpellCooldowns']}

    for q in args.queries:
        ids = [int(q)] if q.isdigit() else [i for i, n in names.items() if re.fullmatch(q, n, re.I)]
        for i in sorted(ids):
            m = misc.get(i, {})
            c = cooldowns.get(i, {})
            print(f"{i} {names.get(i)!r} dur={durations.get(m.get('DurationIndex'))} "
                  f"cast={casts.get(m.get('CastingTimeIndex'))} cd={c.get('RecoveryTime')} "
                  f"cat={c.get('CategoryRecoveryTime')} gcd={c.get('StartRecoveryTime')}")
            for e in sorted(effects.get(i, []), key=lambda e: int(e['EffectIndex'])):
                print(f"   #{e['EffectIndex']} eff={e['Effect']} aura={e['EffectAura']} "
                      f"bp={e['EffectBasePointsF']} ppl={e['EffectRealPointsPerLevel']} "
                      f"misc={e['EffectMiscValue_0']} trig={e['EffectTriggerSpell']} "
                      f"coef={e['EffectBonusCoefficient']} mask={e['EffectSpellClassMask_0']},"
                      f"{e['EffectSpellClassMask_1']},{e['EffectSpellClassMask_2']}")


if __name__ == '__main__':
    main()
