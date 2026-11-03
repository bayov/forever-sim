#!/usr/bin/env python3
"""Record a cloud session's newest result event: its cumulative modelUsage and total_cost_usd.

usage: add_cloud.py <data_dir> <session_id> <saved_tool_result_file>
   or: add_cloud.py <data_dir> <session_id> --inline <created_at> <total_cost_usd> '<modelUsage json>'

A cloud session's result events carry running totals for the whole session (all models,
subagents included, kept across container restarts), so the newest one is all we need.

--base (after the session id) records the result as the session's baseline instead
(<id>.base.json): for a session that started before the audit window, the newest result from
before the window. With a saved page, --base <ISO time> picks the newest result before that time.
get_session's external_metadata.usage carries the same running totals as the newest result."""
import json, os, sys

args = sys.argv[1:]
base = None
if '--base' in args:
    i = args.index('--base'); base = args[i + 1] if args[i + 1] != '--inline' else ''
    del args[i:i + (2 if base else 1)]
data, sid = args[0], args[1]
os.makedirs(f'{data}/cloud', exist_ok=True)
if args[2] == '--inline':
    ts, cost, mu = args[3], float(args[4]), json.loads(args[5])
else:
    s = open(args[2]).read()
    d = json.loads(s[s.index('{"ccr"'):s.rindex('}') + 1])['ccr']
    res = [(e['created_at'], e['result'].get('internal_anthropic_catchall', e['result']))
           for e in d.get('data', []) if e.get('result')]
    if base:
        res = [x for x in res if x[0] < base]
    if not res:
        print(json.dumps({'results': 0, 'has_more': d.get('has_more'), 'first_id': d.get('first_id')}))
        sys.exit()
    ts, r = max(res, key=lambda x: x[0])
    cost, mu = r.get('total_cost_usd'), r.get('modelUsage')
out = f"{data}/cloud/{sid}{'.base' if base is not None else ''}.json"
json.dump({'ts': ts, 'cost': cost, 'modelUsage': mu}, open(out, 'w'), indent=1)
print(json.dumps({'ts': ts, 'cost': cost}))
