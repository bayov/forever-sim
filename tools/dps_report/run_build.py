#!/usr/bin/env python3
"""run_build.py <id> : sims one build's request ($DASH_DIR/req/<id>.json) and writes $DASH_DIR/out/<id>.json.

Main run: 180 s, 20k iterations, seed 1, all values saved (distribution), first iteration logged.
Race sweep: every race the class can be, 10k iterations, seed 1.
Fight lengths: 60/120/180/300/480 s, 5k iterations, preset race.
Sample fights: seeds 1..3, one iteration each, logged and parsed into a compact timeline.
"""
import json, os, re, subprocess, sys, tempfile, copy, math

from dashpaths import D, BIN, TREE  # base dir, binaries, merged tree (env DASH_DIR / DASH_BIN / DASH_TREE)
CLI = os.path.join(BIN, 'wowsimcli')
STATS = os.path.join(BIN, 'dashstats')

CLASS_RACES = {
    'Warrior': ['Human', 'Dwarf', 'NightElf', 'Gnome', 'SkyborneHighOrder', 'Orc', 'Undead', 'Tauren', 'Troll', 'SkyborneWindshaper'],
    'Hunter': ['Human', 'Dwarf', 'NightElf', 'SkyborneHighOrder', 'Orc', 'Tauren', 'Troll', 'SkyborneWindshaper'],
    'Rogue': ['Human', 'Dwarf', 'NightElf', 'Gnome', 'SkyborneHighOrder', 'Orc', 'Undead', 'Troll', 'SkyborneWindshaper'],
    'Mage': ['Human', 'Gnome', 'SkyborneHighOrder', 'Orc', 'Undead', 'Troll'],
    'Warlock': ['Human', 'Gnome', 'Orc', 'Undead', 'Troll'],
    'Priest': ['Human', 'Dwarf', 'NightElf', 'Gnome', 'Undead', 'Troll'],
    'Paladin': ['Human', 'Dwarf', 'Undead'],
    'Shaman': ['Dwarf', 'Orc', 'Tauren', 'Troll', 'SkyborneWindshaper'],
    'Druid': ['NightElf', 'SkyborneHighOrder', 'Tauren', 'SkyborneWindshaper'],
}
HORDE = {'Orc', 'Undead', 'Tauren', 'Troll', 'SkyborneWindshaper'}
LENGTHS = [60, 120, 180, 300, 480]


def run(req, iters, seed, debug=False, save_all=False, race=None, duration=None, names=None):
    req = copy.deepcopy(req)
    if race:
        req['raid']['parties'][0]['players'][0]['race'] = 'Race' + race
    if duration:
        req['encounter']['duration'] = duration
    req['simOptions'] = {'iterations': iters, 'randomSeed': str(seed)}
    if debug:
        req['simOptions']['debugFirstIteration'] = True
    if save_all:
        req['simOptions']['saveAllValues'] = True
    with tempfile.TemporaryDirectory() as d:
        i, o = os.path.join(d, 'i.json'), os.path.join(d, 'o.json')
        json.dump(req, open(i, 'w'))
        env = dict(os.environ)
        if names:
            env['DASH_NAMES'] = names
        subprocess.run([CLI, 'sim', '--infile', i, '--outfile', o], check=True, capture_output=True, cwd=TREE, env=env)
        r = json.load(open(o))
    if r.get('error'):
        raise RuntimeError(r['error']['message'][:800])
    return r


def akey(aid):
    if 'spellId' in aid:
        k = 's%d' % aid['spellId']
    elif 'itemId' in aid:
        k = 'i%d' % aid['itemId']
    else:
        k = 'o%s' % aid.get('otherId', '0')
    if aid.get('tag'):
        k += 't%d' % aid['tag']
    return k


LOGID = re.compile(r'\{(SpellID|ItemID|OtherID): (\d+)(?:, Tag: (-?\d+))?\}')
OTHER_ENUM = {'1': 'OtherActionWait', '2': 'OtherActionManaRegen', '3': 'OtherActionEnergyRegen', '4': 'OtherActionFocusRegen',
              '7': 'OtherActionAttack', '8': 'OtherActionShoot', '10': 'OtherActionRefund', '13': 'OtherActionPotion', '14': 'OtherActionMove',
              '15': 'OtherActionComboPoints', '16': 'OtherActionExplosives'}


def lkey(m):
    kind, n, tag = m.group(1), m.group(2), m.group(3)
    if kind == 'SpellID':
        k = 's' + n
    elif kind == 'ItemID':
        k = 'i' + n
    else:
        k = 'o' + OTHER_ENUM.get(n, n)
    if tag and tag != '0':
        k += 't' + tag
    return k


LINE = re.compile(r'^\[(-?[\d.]+)\] \[([^\]]*)\] (.*)$')
DMG = re.compile(r'^\[(Target \d+)\] (\{[^}]*\}) (tick )?([A-Za-z ]+?)(?: \((\d+)% Resist\))?(?: for ([\d.]+) damage| \(SpellSchool)')
CAST = re.compile(r'^Casting (\{[^}]*\}) \(Cost = ([\d.]+), Cast Time = ([\d.]+)(m?s), Effective Time = ([\d.]+)(m?s)\)')
RES = re.compile(r'^(Spent|Gained) ([\d.]+) ([a-z][a-z ]*?) from (\{[^}]*\}) \(([\d.]+) --> ([\d.]+)\)')
AURA = re.compile(r'^Aura (gained|faded): (\{[^}]*\})')
STACK = re.compile(r'^(\{[^}]*\}) stacks: (\d+) --> (\d+)')
MAJOR = re.compile(r'^Major cooldown used: (\{[^}]*\})')
OUTC = {'Hit': 'H', 'Crit': 'C', 'Miss': 'M', 'Dodge': 'D', 'Parry': 'P', 'Glance': 'G', 'Block': 'B', 'Crush': 'X', 'Blocked Crit': 'BC'}


def secs(v, unit):
    return float(v) / 1000 if unit == 'ms' else float(v)


def parse_log(text, dur):
    units = {}

    def uid(label):
        label = label.strip()
        if label.startswith('Target'):
            return 't'
        if label in ('(#1)', ''):
            return 'p'
        if label not in units:
            units[label] = 'u%d' % (len(units) + 1)
        return units[label]

    casts, dmg, cds = [], [], []
    open_auras, auras = {}, []
    res = {}
    for line in text.split('\n'):
        m = LINE.match(line)
        if not m:
            continue
        t, who, body = round(float(m.group(1)), 2), m.group(2), m.group(3)
        u = uid(who)
        if body.startswith('[Target'):
            d = DMG.match(body)
            if d:
                k = lkey(LOGID.search(d.group(2)))
                o = OUTC.get(d.group(4).strip(), d.group(4).strip())
                dmg.append([t, k, round(float(d.group(6) or 0)), o, u, 1 if d.group(3) else 0, int(d.group(5) or 0)])
            continue
        c = CAST.match(body)
        if c:
            k = lkey(LOGID.search(c.group(1)))
            casts.append([t, k, round(secs(c.group(3), c.group(4)), 2), round(secs(c.group(5), c.group(6)), 2), u, round(float(c.group(2)))])
            continue
        r = RES.match(body)
        if r:
            rt = r.group(3).strip()
            if rt in ('health',):
                continue
            res.setdefault(u + ':' + rt, []).append([t, round(float(r.group(6)), 1)])
            continue
        a = AURA.match(body)
        if a:
            k = lkey(LOGID.search(a.group(2)))
            key = (u, k)
            if a.group(1) == 'gained':
                open_auras[key] = [k, u, t, None, 1]
            else:
                rec = open_auras.pop(key, None)
                if rec:
                    rec[3] = t
                    auras.append(rec)
            continue
        s = STACK.match(body)
        if s:
            k = lkey(LOGID.search(s.group(1)))
            rec = open_auras.get((u, k))
            if rec:
                rec[4] = max(rec[4], int(s.group(3)))
            continue
        mj = MAJOR.match(body)
        if mj:
            cds.append([t, lkey(LOGID.search(mj.group(1))), u])
    for rec in open_auras.values():
        rec[3] = round(dur, 2)
        auras.append(rec)
    # A resource series keeps one point per change; thin very long ones (energy ticks) to every 0.5 s.
    for key, pts in res.items():
        if len(pts) > 1500:
            thin, last = [], -9
            for p in pts:
                if p[0] - last >= 0.5:
                    thin.append(p)
                    last = p[0]
            res[key] = thin
    return {'dur': round(dur, 2), 'units': {v: k for k, v in units.items()}, 'casts': casts, 'dmg': dmg,
            'auras': sorted(auras, key=lambda a: (a[1], a[2])), 'res': res, 'cds': cds}


def unit_actions(unit, iters, avgdur, owner):
    rows = []
    for a in unit.get('actions', []):
        tot = {}
        for tg in a.get('targets', []):
            for f, v in tg.items():
                if isinstance(v, (int, float)) and f != 'unitIndex':
                    tot[f] = tot.get(f, 0) + v
        dmg = tot.get('damage', 0)
        casts = tot.get('casts', 0)
        if dmg == 0 and casts == 0:
            continue
        hits = tot.get('hits', 0)
        crits = tot.get('crits', 0)
        ticks = tot.get('ticks', 0)
        crit_ticks = tot.get('critTicks', 0)
        landed = hits + crits
        rows.append({
            'key': akey(a['id']), 'owner': owner, 'melee': bool(a.get('isMelee')),
            'dps': dmg / iters / avgdur, 'damage': dmg / iters,
            'casts': casts / iters, 'hits': hits / iters, 'crits': crits / iters,
            'ticks': ticks / iters, 'critTicks': crit_ticks / iters,
            'misses': tot.get('misses', 0) / iters, 'dodges': tot.get('dodges', 0) / iters, 'parries': tot.get('parries', 0) / iters,
            'glances': tot.get('glances', 0) / iters, 'blocks': tot.get('blocks', 0) / iters,
            'resisted': (tot.get('resistedHits', 0) + tot.get('resistedCrits', 0) + tot.get('resistedTicks', 0) + tot.get('resistedCritTicks', 0)) / iters,
            'avgHit': (tot.get('damage', 0) - tot.get('critDamage', 0) - tot.get('tickDamage', 0) - tot.get('glanceDamage', 0) - tot.get('blockDamage', 0)) / max(1, hits - ticks + crit_ticks) if hits else 0,
            'avgCrit': (tot.get('critDamage', 0) - tot.get('critTickDamage', 0)) / max(1, crits - crit_ticks) if crits > crit_ticks else 0,
            'avgTick': (tot.get('tickDamage', 0) + tot.get('critTickDamage', 0)) / max(1, ticks + crit_ticks) if ticks + crit_ticks else 0,
            'critPct': 100 * crits / landed if landed else 0,
        })
    return rows


def main(bid):
    req = json.load(open(os.path.join(D, 'req', bid + '.json')))
    meta = json.load(open(os.path.join(D, 'req', bid + '.meta.json')))
    names = os.path.join(D, 'names', bid + '.tsv')
    os.makedirs(os.path.dirname(names), exist_ok=True)
    if os.path.exists(names):
        os.remove(names)
    out = {'id': bid, 'meta': meta}
    race = meta['race']

    r = run(req, 20000, 1, save_all=True, names=names)
    p = r['raidMetrics']['parties'][0]['players'][0]
    vals = sorted(p['dps'].get('allValues', []))
    n = len(vals)
    pct = {str(q): vals[min(n - 1, int(q / 100 * n))] for q in (1, 5, 10, 25, 50, 75, 90, 95, 99)}
    lo, hi = math.floor(vals[0] / 5) * 5, math.ceil(vals[-1] / 5) * 5
    bins = {}
    for v in vals:
        b = int(math.floor(v / 5) * 5)
        bins[b] = bins.get(b, 0) + 1
    avgdur = r['avgIterationDuration']
    actions = unit_actions(p, 20000, avgdur, 'player')
    pets = []
    for pet in p.get('pets', []):
        pets.append({'name': pet['name'], 'dps': pet['dps']['avg']})
        actions += unit_actions(pet, 20000, avgdur, pet['name'])
    auras = [{'key': akey(a['id']), 'uptime': 100 * a['uptimeSecondsAvg'] / avgdur, 'procs': a.get('procsAvg', 0), 'owner': 'player'} for a in p.get('auras', [])]
    for t in r['encounterMetrics']['targets'][:1]:
        auras += [{'key': akey(a['id']), 'uptime': 100 * a['uptimeSecondsAvg'] / avgdur, 'procs': a.get('procsAvg', 0), 'owner': 'target'} for a in t.get('auras', [])]
    resources = [{'key': akey(x['id']), 'type': x['type'].replace('ResourceType', ''), 'gain': x['gain'] / 20000, 'actual': x['actualGain'] / 20000} for x in p.get('resources', []) if x['type'] != 'ResourceTypeHealth']
    out['main'] = {
        'iters': 20000, 'avgDur': avgdur, 'dps': p['dps']['avg'], 'stdev': p['dps']['stdev'], 'min': p['dps']['min'], 'max': p['dps']['max'],
        'se': p['dps']['stdev'] / math.sqrt(20000), 'pct': pct, 'hist': sorted([[k, v] for k, v in bins.items()]),
        'oomSecs': p.get('secondsOomAvg', 0), 'manaEnd': p.get('manaEnd', {}).get('avg', 0) if isinstance(p.get('manaEnd'), dict) else 0,
        'tps': p['threat']['avg'], 'petDps': sum(x['dps'] for x in pets), 'pets': pets,
        'actions': actions, 'auras': auras, 'resources': resources,
    }
    # Sample fights: one-iteration runs; pick the seeds nearest the 10th, 50th and 90th percentile.
    rolls = []
    for seed in range(1, 31):
        rs = run(req, 1, seed)
        rolls.append((rs['raidMetrics']['parties'][0]['players'][0]['dps']['avg'], seed))
    samples, used = [], set()
    for label, q in (('Typical fight', '50'), ('Unlucky fight', '10'), ('Lucky fight', '90')):
        dps_s, seed = min((x for x in rolls if x[1] not in used), key=lambda x: abs(x[0] - pct[q]))
        used.add(seed)
        rs = run(req, 1, seed, debug=True)
        samples.append(parse_log(rs['logs'], rs['firstIterationDuration']) | {'seed': seed, 'dps': dps_s, 'label': label, 'pct': int(q)})
    out['samples'] = samples

    races = []
    for rc in CLASS_RACES[meta['class']]:
        rr = run(req, 10000, 1, race=rc)
        pp = rr['raidMetrics']['parties'][0]['players'][0]
        races.append({'race': rc, 'faction': 'Horde' if rc in HORDE else 'Alliance', 'dps': pp['dps']['avg'], 'se': pp['dps']['stdev'] / 100})
    out['races'] = races

    lengths = []
    for L in LENGTHS:
        rr = run(req, 5000, 1, duration=L)
        pp = rr['raidMetrics']['parties'][0]['players'][0]
        lengths.append({'secs': L, 'dps': pp['dps']['avg'], 'se': pp['dps']['stdev'] / math.sqrt(5000), 'oomSecs': pp.get('secondsOomAvg', 0)})
    out['lengths'] = lengths

    st = subprocess.run([STATS, os.path.join(D, 'req', bid + '.json')], capture_output=True, text=True, cwd=TREE)
    out['stats'] = json.loads(st.stdout) if st.returncode == 0 else {}
    pl = req['raid']['parties'][0]['players'][0]
    out['setup'] = {'equipment': pl.get('equipment', {}).get('items', []), 'consumes': pl.get('consumes', {}), 'playerBuffs': pl.get('buffs', {}),
                    'raidBuffs': req['raid'].get('buffs', {}), 'partyBuffs': req['raid']['parties'][0].get('buffs', {}),
                    'debuffs': req['raid'].get('debuffs', {}), 'professions': [pl.get('profession1', ''), pl.get('profession2', '')],
                    'encounter': {k: v for k, v in req['encounter'].items() if k != 'targets'} | {'target': {k: v for k, v in req['encounter']['targets'][0].items() if k != 'stats'}},
                    'level': pl.get('level', 60)}
    os.makedirs(os.path.join(D, 'out'), exist_ok=True)
    json.dump(out, open(os.path.join(D, 'out', bid + '.json'), 'w'), separators=(',', ':'))
    print(f"{bid}: {out['main']['dps']:.1f} (published {meta['published']['dps']})")


def resample(bid):
    req = json.load(open(os.path.join(D, 'req', bid + '.json')))
    path = os.path.join(D, 'out', bid + '.json')
    out = json.load(open(path))
    pct = out['main']['pct']
    rolls = []
    for seed in range(1, 31):
        rs = run(req, 1, seed)
        rolls.append((rs['raidMetrics']['parties'][0]['players'][0]['dps']['avg'], seed))
    samples, used = [], set()
    for label, q in (('Typical fight', '50'), ('Unlucky fight', '10'), ('Lucky fight', '90')):
        dps_s, seed = min((x for x in rolls if x[1] not in used), key=lambda x: abs(x[0] - pct[q]))
        used.add(seed)
        rs = run(req, 1, seed, debug=True)
        samples.append(parse_log(rs['logs'], rs['firstIterationDuration']) | {'seed': seed, 'dps': dps_s, 'label': label, 'pct': int(q)})
    out['samples'] = samples
    json.dump(out, open(path, 'w'), separators=(',', ':'))
    print(bid, 'resampled')


if __name__ == '__main__':
    if sys.argv[1] == '--resample':
        for b in sys.argv[2:]:
            resample(b)
    else:
        for b in sys.argv[1:]:
            main(b)
