#!/usr/bin/env python3
"""Hit-trade gear steps (round 70753 phase 5, standards 4).

For a Level 60 synthetic set, writes the alternative pieces that move hit rating into another stat
in 1% steps (10 rating) at the 1b budget price, and the gear sets that wear them:
<set>_hit-1.gear.json ... _hit-N (10, 20, ... less gear hit) and <set>_hit+1 ... (more hit, paid
for by the trade stat). tools/rotopt -hit-steps picks the step per talent build.

Each step changes one more piece by 10 rating, so step -k wears k changed pieces, one unit each,
going round the pieces again (20 less on one piece) only when the set has fewer hit pieces than
steps. The pieces are copies of the set's own (same slot, item level, armor, set name) with a new
id from the spec's id block, written as literals between markers in the class's synthetic items
file, so tools/audit/gear (lin2.py) checks them like the rest.

Usage: python3 orchestration/tools/hit_steps.py <spec> [<spec> ...]   (or "all")
Then gofmt runs on the file; check with tools/audit/gear (extract.sh + lin2.py).
"""
import json, os, re, subprocess, sys

ROOT = os.path.abspath(os.path.join(os.path.dirname(__file__), '..', '..'))
DB = os.path.join(ROOT, 'tools', 'database')

# Budget price: how much of a stat 10 points of budget (1% hit, 10 rating) buys (standards 1b).
PER10 = {'crit': 10, 'Strength': 10, 'Agility': 10, 'Intellect': 10, 'Spirit': 10, 'Stamina': 10,
         'AttackPower': 20, 'RangedAttackPower': 25, 'SpellPower': 12, 'MP5': 4}

# Per spec: the class file, the base gear set (without .gear.json), the stat the freed budget buys
# (chosen from the sim's stat weights, see hit_sweep.md), how many steps down and up, and the first
# id to use (the spec's id block; the script checks the ids are free).
SPECS = json.load(open(os.path.join(os.path.dirname(__file__), 'hit_steps.json')))

# Variants already written this run, per file and trade stats: specs whose sets share a piece (a set's
# variant with another totem, say) share its variant too, written in the first spec's block.
SHARED = {}

BEGIN = '// BEGIN hit-trade steps: %s (orchestration/tools/hit_steps.py, do not edit by hand)'
END = '// END hit-trade steps: %s'


def used_ids():
    ids = set()
    for f in os.listdir(DB):
        if f.startswith('forever_synthetic_items') and f.endswith('.go'):
            ids |= {int(x) for x in re.findall(r'\bId:\s*(99\d{4})', open(os.path.join(DB, f)).read())}
    return ids


def item_blocks(src):
    """id -> (start, end, text) of each item literal `{ ... Id: 99xxxx ... },`."""
    out = {}
    for m in re.finditer(r'\bId:\s*(99\d{4})', src):
        start = src.rfind('{', 0, m.start())
        depth, i = 0, start
        while True:
            c = src[i]
            if c == '{':
                depth += 1
            elif c == '}':
                depth -= 1
                if depth == 0:
                    break
            i += 1
        out[int(m.group(1))] = (start, i + 1, src[start:i + 1])
    return out


def field(text, name):
    m = re.search(r'\b' + name + r':\s*([\d.]+)', text)
    return float(m.group(1)) if m else 0.0


def set_field(text, name, value):
    """Set a top-level numeric field (HitRating, CritRating), adding or dropping it."""
    if re.search(r'\b' + name + r':\s*[\d.]+', text):
        if value == 0:
            return re.sub(r'\s*\b' + name + r':\s*[\d.]+,', '', text, count=1)
        return re.sub(r'\b(' + name + r':\s*)[\d.]+', lambda m: m.group(1) + fmt(value), text, count=1)
    if value == 0:
        return text
    # after the Stats line
    return re.sub(r'(Stats:[^\n]*\n)', lambda m: m.group(1) + '\t\t%s: %s,\n' % (name, fmt(value)), text, count=1)


def stat_value(text, stat):
    m = re.search(r'stats\.Stats\{(.*?)\}', text)
    mm = re.search(r'stats\.' + stat + r':\s*([\d.]+)', m.group(1)) if m else None
    return float(mm.group(1)) if mm else 0.0


def set_stat(text, stat, value):
    m = re.search(r'stats\.Stats\{(.*?)\}', text)
    body = m.group(1)
    if re.search(r'stats\.' + stat + r':', body):
        if value == 0:
            body = re.sub(r'\s*stats\.' + stat + r':\s*[\d.]+,?', '', body).strip().rstrip(',')
        else:
            body = re.sub(r'(stats\.' + stat + r':\s*)[\d.]+', lambda x: x.group(1) + fmt(value), body)
    elif value:
        parts = [p.strip() for p in body.split(',') if p.strip()]
        # before Armor, so the spec's stats read first
        at = next((i for i, p in enumerate(parts) if p.startswith('stats.Armor')), len(parts))
        parts.insert(at, 'stats.%s: %s' % (stat, fmt(value)))
        body = ', '.join(parts)
    return text[:m.start(1)] + body + text[m.end(1):]


def fmt(v):
    return str(int(v)) if float(v).is_integer() else str(v)


def trade_amount(text, stat):
    return field(text, 'CritRating') if stat == 'crit' else stat_value(text, stat)


def set_trade(text, stat, value):
    return set_field(text, 'CritRating', value) if stat == 'crit' else set_stat(text, stat, value)


WT = {'Strength': 1, 'Agility': 1, 'Stamina': 1, 'Intellect': 1, 'Spirit': 1, 'SpellPower': .86, 'SpellDamage': .86,
      'FirePower': .7, 'FrostPower': .7, 'ShadowPower': .7, 'ArcanePower': .7, 'NaturePower': .7, 'HolyPower': .7,
      'HealingPower': .45, 'AttackPower': .5, 'RangedAttackPower': .4, 'MP5': 2.4}
SLOT = {'Head': (1, 'T1'), 'Chest': (1, 'T1'), 'Legs': (1, 'T1'), 'Shoulder': (.75, 'T2'), 'Hands': (.75, 'T2'),
        'Waist': (.75, 'T2'), 'Feet': (.75, 'T2'), 'Neck': (.5625, 'T3'), 'Back': (.5625, 'T3'), 'Wrist': (.5625, 'T3'),
        'Finger': (.5625, 'T3')}
F = {'caster': {'T1': 1.053, 'T2': 1.026, 'T3': 0.995}, 'phys': {'T1': .95, 'T2': .959, 'T3': .911},
     'hunter': {'T1': 1.028, 'T2': 1.056, 'T3': .977}, 'hybrid': {'T1': 1.093, 'T2': 1.103, 'T3': 1.048}}


def shape(text, arch):
    """(linear total off its 1b target, 1.7-norm off its target), as fractions (tools/audit/gear/lin2.py)."""
    m = re.search(r'stats\.Stats\{(.*?)\}', text)
    st = {k: float(v) for k, v in re.findall(r'stats\.(\w+):\s*([\d.]+)', m.group(1))}
    comp = {}
    for k, v in st.items():
        if k not in WT or (k == 'RangedAttackPower' and st.get('AttackPower')):
            continue
        kk = 'AP' if k in ('AttackPower', 'RangedAttackPower') else ('SP' if WT[k] < 1 and k != 'MP5' else k)
        comp[kk] = comp.get(kk, 0) + v * WT[k]
    for k in ('HitRating', 'CritRating'):
        if field(text, k):
            comp[k] = field(text, k)
    ilvl = field(text, 'Ilvl')
    mod, tier = SLOT[re.search(r'Type:\s*proto\.ItemType_ItemType(\w+)', text).group(1)]
    lin = sum(comp.values()) / (mod * (1.125 * ilvl + 2.0) * F[arch][tier]) - 1
    norm = sum(v ** 1.7 for v in comp.values()) ** (1 / 1.7) / (mod * (0.661 * ilvl + 2.92)) - 1
    return lin, norm


def gear_ids(path):
    return [it.get('id', 0) for it in json.load(open(path))['items']]


def build(spec, cfg, used):
    path = os.path.join(DB, cfg['file'])
    src = open(path).read()
    src = re.sub(r'\n?' + re.escape(BEGIN % spec) + r'.*?' + re.escape(END % spec) + r'\n', '\n', src, flags=re.S)
    blocks = item_blocks(src)
    base = os.path.join(ROOT, cfg['gear'])
    ids = gear_ids(base + '.gear.json')
    # The stat the freed budget buys, or a list in order of preference: each piece takes the first
    # that keeps the piece within the 1b budget (+-1%) and shape (+-6%) checks.
    stats_pref = cfg['stat'] if isinstance(cfg['stat'], list) else [cfg['stat']]
    # Steps up pay for their hit with the least valuable of the set's stats first ("give"), by default
    # the trade stats in reverse.
    give_pref = cfg.get('give') or stats_pref[::-1]

    def variant(i, direction, u):
        """The piece with u steps of hit traded, as (text, stat, what), or None when no listed stat
        fits: not enough of it to give up, or the piece would leave the 1b budget or shape check."""
        text = blocks[i][2]
        hit = field(text, 'HitRating') + direction * 10 * u
        if hit < 0:
            return None
        for rank, stat in enumerate(stats_pref if direction < 0 else give_pref):
            per10 = PER10[stat]
            trade = trade_amount(text, stat) - direction * per10 * u
            if trade < 0:
                continue
            t = set_trade(set_field(text, 'HitRating', hit), stat, trade)
            lin, norm = shape(t, cfg['arch'])
            if abs(lin) <= 0.01 and abs(norm) <= 0.06:
                what = '%d crit rating' % (per10 * u) if stat == 'crit' else '%d %s' % (per10 * u, stat)
                return t, rank, what
        return None

    used -= {i for i in used if cfg['id0'] <= i < cfg['id0'] + 100 and i not in blocks}
    next_id = [cfg['id0']]

    def new_id():
        while next_id[0] in used:
            next_id[0] += 1
        used.add(next_id[0])
        return next_id[0]

    def armor(i):
        t = re.search(r'Type:\s*proto\.ItemType_ItemType(\w+)', blocks[i][2]).group(1)
        return t not in ('Weapon', 'Ranged', 'Trinket')

    pieces = [i for i in dict.fromkeys(ids) if i in blocks and armor(i)]
    literals, sets = [], {}
    for direction, n in ((-1, cfg.get('down', 0)), (1, cfg.get('up', 0))):
        if direction < 0:
            # pieces with hit, those already carrying the trade stat first, most hit first
            cand = sorted([i for i in pieces if variant(i, -1, 1)],
                          key=lambda i: (variant(i, -1, 1)[1], -field(blocks[i][2], 'HitRating'), i))
        else:
            # pieces carrying the trade stat, those already carrying hit first
            cand = sorted([i for i in pieces if variant(i, 1, 1)],
                          key=lambda i: (variant(i, 1, 1)[1], field(blocks[i][2], 'HitRating') == 0, i))
        room = {}
        for i in cand:
            room[i] = 1
            while variant(i, direction, room[i] + 1):
                room[i] += 1
        units = []
        for r in range(1, max(room.values(), default=0) + 1):
            units += [i for i in cand if room[i] >= r]
        if len(units) < n:
            print('%s: only %d of %d steps %s fit on the set' % (spec, len(units), n, 'down' if direction < 0 else 'up'))
            n = len(units)
        variants = SHARED.setdefault((cfg['file'], tuple(stats_pref), tuple(give_pref)), {})
        for k in range(1, n + 1):
            taken = {}
            for i in units[:k]:
                taken[i] = taken.get(i, 0) + 1
            gear = list(ids)
            for i, u in taken.items():
                key = (i, direction * u)
                if key not in variants:
                    vid = new_id()
                    text = blocks[i][2]
                    name = re.search(r'Name:\s*"([^"]*)"', text).group(1)
                    t, _, what = variant(i, direction, u)
                    if direction < 0:
                        label, note = 'hit -%d' % (10 * u), '%d hit rating less, %s more' % (10 * u, what)
                    else:
                        label, note = 'hit +%d' % (10 * u), '%d hit rating more, %s less' % (10 * u, what)
                    t = re.sub(r'\bId:\s*\d+', 'Id: %d' % vid, t, count=1)
                    t = t.replace('Name: "%s"' % name, 'Name: "%s"' % re.sub(r'\)$', ', %s)' % label, name) if name.endswith(')') else name + ' (%s)' % label, 1)
                    literals.append('\t// %s, hit-trade step: %s (%d).\n\t%s,' % (name, note, i, t))
                    variants[key] = vid
                pos = gear.index(i)
                gear[pos] = variants[key]
            sets[direction * k] = gear
    for k, gear in sets.items():
        data = json.load(open(base + '.gear.json'))
        for it, gid in zip(data['items'], gear):
            if gid:
                it['id'] = gid
        lines = ',\n'.join('    ' + json.dumps(it, separators=(',', ':')) for it in data['items'])
        out = '%s_hit%+d.gear.json' % (base, k)
        open(out, 'w').write('{\n  "items": [\n' + lines + '\n  ]\n}\n')
    # stale step files past the new count
    d, b = os.path.split(base)
    for f in os.listdir(d):
        m = re.fullmatch(re.escape(b) + r'_hit([+-]\d+)\.gear\.json', f)
        if m and int(m.group(1)) not in sets:
            os.remove(os.path.join(d, f))
    var = 'hitSteps' + re.sub(r'[^A-Za-z0-9]', '', spec.title())
    block = '\n'.join([BEGIN % spec,
                       'func init() {'] +
                      # "init": Go statements on each step item, for a file whose base items get fields from a wrapper
                      # (the ret items' quality, level and class come from retribution()), which the copies miss
                      (['\tfor _, item := range %s {' % var] + ['\t\t' + l for l in SPECS[spec]['init']] + ['\t}']
                       if SPECS[spec].get('init') else []) +
                      ['\tForeverSyntheticItems = append(ForeverSyntheticItems, %s...)' % var, '}', '',
                       'var %s = []*proto.UIItem{' % var] + literals + ['}', END % spec, ''])
    src = src.rstrip('\n') + '\n\n' + block
    open(path, 'w').write(src)
    subprocess.run(['gofmt', '-w', path], check=True)
    return sorted(sets)


def main():
    names = sys.argv[1:]
    if names == ['all']:
        names = list(SPECS)
    # every spec of a file goes together, in hit_steps.json's order, so shared variants land once
    files = {SPECS[n]['file'] for n in names}
    names = [n for n in SPECS if SPECS[n]['file'] in files]
    used = used_ids()
    for spec in names:
        steps = build(spec, SPECS[spec], used)
        print(spec, 'steps', ' '.join('%+d' % s for s in steps))


if __name__ == '__main__':
    main()
