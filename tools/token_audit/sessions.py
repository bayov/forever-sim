#!/usr/bin/env python3
"""Turn saved list_sessions pages into <data_dir>/sessions.json, keeping sessions active since <since>.

usage: sessions.py <data_dir> <since YYYY-MM-DD> <saved_list_sessions_file>... [--skip <session_id>]"""
import json, sys

args = sys.argv[1:]
skip = set()
if '--skip' in args:
    i = args.index('--skip'); skip.add(args[i + 1]); del args[i:i + 2]
data, since, files = args[0], args[1], args[2:]
out = {}
for path in files:
    s = open(path).read()
    for x in json.loads(s[s.index('{"ccr"'):s.rindex('}') + 1])['ccr']['data']:
        if x['updated_at'] < since or x['id'] in skip:
            continue
        cu = (x.get('external_metadata') or {}).get('context_usage') or {}
        out[x['id']] = dict(
            id=x['id'], title=x['title'], created=x['created_at'], updated=x['updated_at'],
            kind=x.get('environment_kind'), origin=x.get('origin'), model=x.get('configured_model'),
            repo=','.join(g.get('git_repository', {}).get('url', '').split('/')[-1]
                          for g in (x.get('session_context') or {}).get('sources', [])),
            ctx=cu.get('used_tokens'))
json.dump(list(out.values()), open(f'{data}/sessions.json', 'w'), indent=1)
kinds = {}
for s in out.values():
    kinds.setdefault(s['kind'], []).append(s['id'])
for k, v in kinds.items():
    print(k, len(v)); print(' '.join(v))
