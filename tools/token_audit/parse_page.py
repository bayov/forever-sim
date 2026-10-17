#!/usr/bin/env python3
"""Append per-message usage from one saved list_events page (kinds=["assistant"]) to
<data>/pages/<session>.jsonl, and print has_more / first_id for the next page.

usage: parse_page.py <data_dir> <session_id> <saved_tool_result_file>

Each content block of a message is its own event and repeats the message's usage, so
aggregate.py dedups rows by message id. The usage on these events is the stream's
message_start snapshot: input, cache write and cache read are exact, output_tokens is not
(aggregate.py estimates output from the visible text instead)."""
import json, os, sys

data, sid, path = sys.argv[1:4]
s = open(path).read()
d = json.loads(s[s.index('{"ccr"'):s.rindex('}') + 1])['ccr']
os.makedirs(f'{data}/pages', exist_ok=True)
n = 0
with open(f'{data}/pages/{sid}.jsonl', 'a') as f:
    for e in d.get('data', []):
        a = e.get('assistant')
        if not a:
            continue
        cat = a.get('internal_anthropic_catchall', a)
        m = cat.get('message') or {}
        u = m.get('usage')
        if not u:
            continue
        chars = {'text': 0, 'tool': 0}
        for b in m.get('content') or []:
            if b.get('type') == 'text':
                chars['text'] += len(b.get('text', ''))
            elif b.get('type') == 'tool_use':
                chars['tool'] += len(json.dumps(b.get('input', {})))
        f.write(json.dumps({'id': m.get('id'), 'ts': e.get('created_at'), 'model': m.get('model'),
                            'side': bool(cat.get('parent_tool_use_id')),
                            'in': u.get('input_tokens', 0), 'cw': u.get('cache_creation_input_tokens', 0),
                            'cr': u.get('cache_read_input_tokens', 0), 'chars': chars}) + '\n')
        n += 1
print(json.dumps({'rows': n, 'has_more': d.get('has_more'), 'first_id': d.get('first_id')}))
