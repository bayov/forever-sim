#!/usr/bin/env python3
"""Append usage rows by hand for an assistant page that came back inline (not saved to a file).

usage: add_rows.py <data_dir> <session_id> '<json list of {"id","in","cw","cr","text_chars","tool_chars","side"}>'"""
import json, os, sys

data, sid, rows = sys.argv[1], sys.argv[2], json.loads(sys.argv[3])
os.makedirs(f'{data}/pages', exist_ok=True)
with open(f'{data}/pages/{sid}.jsonl', 'a') as f:
    for r in rows:
        f.write(json.dumps({'id': r['id'], 'ts': r.get('ts'), 'model': r.get('model'), 'side': r.get('side', False),
                            'in': r.get('in', 0), 'cw': r.get('cw', 0), 'cr': r.get('cr', 0),
                            'chars': {'text': r.get('text_chars', 0), 'tool': r.get('tool_chars', 0)}}) + '\n')
print('appended', len(rows))
