"""whdump.py: Forever vs Classic tooltip text side by side for the setup chips' spells and items (for the SIM_USES table)."""
import json, os, sys
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
import whcache, tooltips

for name, (kind, i) in tooltips.SETUP_IDS.items():
    f, c = whcache.cached('forever', kind, i), whcache.cached('classic', kind, i)
    fl = [l for l in tooltips.html_lines((f or {}).get('tooltip')) if l[0] in ('y', 'g')] if whcache.ok(f) else None
    cl = [l for l in tooltips.html_lines((c or {}).get('tooltip')) if l[0] in ('y', 'g')] if whcache.ok(c) else None
    if kind == 'spell' and whcache.ok(f) and not fl:
        fl = [l for l in tooltips.html_lines(f.get('buff')) if l[0] != '2']
    print(f'== {name} [{kind} {i}] forever={"yes" if fl is not None else "NO"} {(f or {}).get("name", "")}')
    print('  F:', ' | '.join(l[1] for l in fl or [])[:400])
    if cl is not None and fl is None:
        print('  C:', ' | '.join(l[1] for l in cl)[:400])
