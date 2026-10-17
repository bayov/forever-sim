#!/usr/bin/env python3
"""Turn <data_dir>/report.json into the "Claude token audit" dashboard's data file,
<data_dir>/token_audit_sessions.json: one row per session.

usage: dashboard_rows.py <data_dir>"""
import json, statistics, sys

data = sys.argv[1]
LONG = ('Scheduled auditors', 'Orchestrator', 'Spawned workers')
rows = []
for x in json.load(open(f'{data}/report.json')):
    t = x['tokens']
    wake = None
    if x['kind'] == 'anthropic_cloud':
        res = json.load(open(f"{data}/cloud/{x['id']}.json")).get('results', [])
        gaps = [b['cost'] - a['cost'] for a, b in zip(res, res[1:]) if b['cost'] >= a['cost']]
        wake = round(statistics.mean(gaps), 2) if gaps else None
    rows.append(dict(
        session=x['id'], title=x['title'], type=x['type'],
        horizon='Long-running' if x['type'] in LONG or 'long-running' in x['type'] else 'One-off',
        repo=x['repo'] or 'none', where='Cloud' if x['kind'] == 'anthropic_cloud' else 'Local CLI',
        basis='exact' if x['exact'] else 'estimated', first=x['first'], last=x['last'],
        active_days=round(x['span_h'] / 24, 2), tokens=x['total'], cache_read=t['cr'], cache_write=t['cw'],
        input=t['in'], output=t['out'], non_cache=x['fresh'], usd=x['usd'],
        usd_per_day=round(x['usd'] / max(x['span_h'] / 24, 1), 2), recent_wake_usd=wake, api_calls=x['api_calls']))
rows.sort(key=lambda r: -r['usd'])
json.dump(rows, open(f'{data}/token_audit_sessions.json', 'w'), indent=0)
print(len(rows), 'rows')
