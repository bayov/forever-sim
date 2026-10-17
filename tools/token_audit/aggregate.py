#!/usr/bin/env python3
"""Combine <data_dir>/cloud/*.json and <data_dir>/pages/*.jsonl into per-session usage.
Writes <data_dir>/report.json and prints a summary by session type.

usage: aggregate.py <data_dir>

Cloud sessions: exact token counts and the harness's own list-price cost (all models, subagents
included). Local CLI (bridge) sessions: exact input, cache write and cache read for the main
thread, deduped by message id; output tokens and cost are estimated (see prices.json)."""
import json, os, re, sys
from collections import defaultdict
from datetime import datetime

HERE = os.path.dirname(os.path.abspath(__file__))
data = sys.argv[1]
sessions = json.load(open(f'{data}/sessions.json'))
P = json.load(open(f'{HERE}/prices.json'))
C = json.load(open(f'{HERE}/categories.json'))


def ts(x):
    return datetime.fromisoformat(x.replace('Z', '+00:00'))


def classify(s, span_h):
    for r in C['rules']:
        if r.get('orchestrator') and s['id'] not in C['orchestrators']:
            continue
        if 'origin' in r and s['origin'] != r['origin']:
            continue
        if 'title_re' in r and not re.search(r['title_re'], s['title']):
            continue
        return r['type']
    where = 'Cloud' if s['kind'] == 'anthropic_cloud' else 'Local CLI'
    return f"{where}, {'long-running' if span_h >= C['long_running_min_span_h'] else 'one-off'}"


def price(model, t):
    p = P['models'].get(model) or P['models']['default']
    return sum(t[k] * p[k] for k in ('in', 'cw', 'cr', 'out')) / 1e6


def cloud(s, path):
    d = json.load(open(path))
    if 'results' in d:  # older shape: a list of results, newest last
        d = [r for r in d['results'] if r.get('modelUsage')][-1]
    by_model = {}
    for m, u in d['modelUsage'].items():
        by_model[m] = {'in': u.get('inputTokens', 0), 'cw': u.get('cacheCreationInputTokens', 0),
                       'cr': u.get('cacheReadInputTokens', 0), 'out': u.get('outputTokens', 0),
                       'usd': u.get('costUSD', 0)}
    span = (ts(d['ts']) - ts(s['created'])).total_seconds() / 3600
    return by_model, d.get('cost'), True, None, span, s['created'], d['ts']


def local(s, path):
    rows = {}
    for line in open(path):
        r = json.loads(line)
        k = r['id'] or line
        if k in rows:  # another content block of the same message
            for c in ('text', 'tool'):
                rows[k]['chars'][c] += r['chars'][c]
        else:
            rows[k] = r
    by_model = {}
    for r in rows.values():
        m = r.get('model') or 'default'
        out = (r['chars']['text'] + r['chars']['tool']) / 3.5 * P['out_mult'] + P['out_per_call']
        t = {'in': r['in'], 'cw': r['cw'], 'cr': r['cr'], 'out': out}
        b = by_model.setdefault(m, {'in': 0, 'cw': 0, 'cr': 0, 'out': 0, 'usd': 0.0})
        for k in t:
            b[k] += t[k]
        b['usd'] += price(m, t)
    tss = sorted(r['ts'] for r in rows.values() if r.get('ts'))
    first, last = (tss[0], tss[-1]) if tss else (s['created'], s['created'])
    span = (ts(last) - ts(first)).total_seconds() / 3600
    return by_model, None, False, len(rows), span, first, last


out = []
for s in sessions:
    cf, pf = f"{data}/cloud/{s['id']}.json", f"{data}/pages/{s['id']}.jsonl"
    if s['kind'] == 'anthropic_cloud' and os.path.exists(cf):
        by_model, usd, exact, calls, span, first, last = cloud(s, cf)
    elif os.path.exists(pf):
        by_model, usd, exact, calls, span, first, last = local(s, pf)
    else:
        print('no data:', s['id'], s['title'], file=sys.stderr)
        continue
    tok = {k: sum(b[k] for b in by_model.values()) for k in ('in', 'cw', 'cr', 'out')}
    usd = usd if usd is not None else sum(b['usd'] for b in by_model.values())
    out.append(dict(id=s['id'], title=s['title'], repo=s['repo'], kind=s['kind'], origin=s['origin'],
                    type=classify(s, span), exact=exact, api_calls=calls, first=first, last=last,
                    span_h=round(span, 1), tokens={k: round(v) for k, v in tok.items()},
                    total=round(sum(tok.values())), fresh=round(tok['in'] + tok['cw'] + tok['out']),
                    usd=round(usd, 2), by_model=by_model))
json.dump(out, open(f'{data}/report.json', 'w'), indent=1)

g = defaultdict(lambda: defaultdict(float))
for r in out:
    x = g[r['type']]
    x['n'] += 1; x['total'] += r['total']; x['fresh'] += r['fresh']; x['out'] += r['tokens']['out']; x['usd'] += r['usd']
T = sum(x['usd'] for x in g.values())
print(f"{'type':30s} {'n':>3s} {'tokens':>9s} {'non-cache':>9s} {'output':>7s} {'list $':>8s} {'share':>6s}")
for k, x in sorted(g.items(), key=lambda kv: -kv[1]['usd']):
    print(f"{k:30s} {x['n']:3.0f} {x['total']/1e6:8.0f}M {x['fresh']/1e6:8.1f}M {x['out']/1e6:6.1f}M {x['usd']:8.0f} {x['usd']/T:6.1%}")
print(f"{'total':30s} {sum(x['n'] for x in g.values()):3.0f} {sum(x['total'] for x in g.values())/1e6:8.0f}M "
      f"{sum(x['fresh'] for x in g.values())/1e6:8.1f}M {sum(x['out'] for x in g.values())/1e6:6.1f}M {T:8.0f}")
