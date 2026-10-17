#!/usr/bin/env python3
"""Record a cloud session's newest result event: its cumulative modelUsage and total_cost_usd.

usage: add_cloud.py <data_dir> <session_id> <saved_tool_result_file>
   or: add_cloud.py <data_dir> <session_id> --inline <created_at> <total_cost_usd> '<modelUsage json>'

A cloud session's result events carry running totals for the whole session (all models,
subagents included, kept across container restarts), so the newest one is all we need."""
import json, os, sys

data, sid = sys.argv[1], sys.argv[2]
os.makedirs(f'{data}/cloud', exist_ok=True)
if sys.argv[3] == '--inline':
    ts, cost, mu = sys.argv[4], float(sys.argv[5]), json.loads(sys.argv[6])
else:
    s = open(sys.argv[3]).read()
    d = json.loads(s[s.index('{"ccr"'):s.rindex('}') + 1])['ccr']
    res = [(e['created_at'], e['result'].get('internal_anthropic_catchall', e['result']))
           for e in d.get('data', []) if e.get('result')]
    if not res:
        print(json.dumps({'results': 0, 'has_more': d.get('has_more'), 'first_id': d.get('first_id')}))
        sys.exit()
    ts, r = max(res, key=lambda x: x[0])
    cost, mu = r.get('total_cost_usd'), r.get('modelUsage')
json.dump({'ts': ts, 'cost': cost, 'modelUsage': mu}, open(f'{data}/cloud/{sid}.json', 'w'), indent=1)
print(json.dumps({'ts': ts, 'cost': cost}))
