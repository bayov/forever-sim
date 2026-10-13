#!/usr/bin/env python3
"""build_harness.py [page.html] [prefix]: builds rd/<prefix>_plain.html and rd/<prefix>_str.html (prefix default h):
the page ($DASH_DIR/page/index.html, or argv[1]) inside a mock dashboard that serves every $DASH_DIR/data/*.json
(or $HDATA/*.json). _str stringifies nested values and numbers the way the real dashboard does; ASYNC=1 makes
datasets arrive late. Needs d3 v7 at $DASH_DIR/page/d3.min.js."""
import glob, json, os, sys

H = os.path.dirname(os.path.abspath(__file__))
sys.path.insert(0, os.path.join(H, '..'))
from dashpaths import D as BASE  # noqa: E402
P = os.path.join(BASE, 'page')
D = os.environ.get('HDATA') or os.path.join(BASE, 'data')
src = sys.argv[1] if len(sys.argv) > 1 and sys.argv[1] else os.path.join(P, 'index.html')
pre = sys.argv[2] if len(sys.argv) > 2 else 'h'

data = {os.path.basename(p)[:-5]: json.load(open(p)) for p in sorted(glob.glob(os.path.join(D, '*.json')))}
open(os.path.join(H, 'bundle.js'), 'w').write('window.__DATA=' + json.dumps(data) + ';')
if not os.path.exists(os.path.join(H, 'd3.min.js')):
    os.symlink(os.path.join(P, 'd3.min.js'), os.path.join(H, 'd3.min.js'))

page = open(src).read()
TOK_L = """--color-bg:#ffffff;--color-fg:#1f1e1d;--color-panel:#faf9f7;--color-border-line:#e6e3dd;--color-fg-muted:#6b6862;
--color-ok:#2e7d32;--color-warn:#a15c00;--color-bad:#c62828;
--cds-chart-axis:#bdb9b1;--cds-chart-grid:#ece9e3;--cds-chart-reference:#8a867e;--cds-chart-muted:#c9c5bd;--cds-chart-status-critical:#c62828;
--cds-surface-popover:#ffffff;--cds-border:#e0ddd6;--cds-shadow-popover:0 4px 16px rgba(0,0,0,.12);--cds-text-secondary:#6b6862;"""
TOK_D = """--color-bg:#1f1e1d;--color-fg:#f2f0ec;--color-panel:#262624;--color-border-line:#3a3936;--color-fg-muted:#a6a39c;
--color-ok:#6cc070;--color-warn:#e0a040;--color-bad:#ef6a5e;
--cds-chart-axis:#55534f;--cds-chart-grid:#33322f;--cds-chart-reference:#8f8b84;--cds-chart-muted:#5a5853;--cds-chart-status-critical:#ef6a5e;
--cds-surface-popover:#2c2b29;--cds-border:#45433f;--cds-shadow-popover:0 4px 16px rgba(0,0,0,.5);--cds-text-secondary:#a6a39c;color-scheme:dark;"""
TOK_C = """--cds-gap-xl:24px;--cds-gap-lg:16px;--cds-gap-md:12px;--cds-gap-sm:8px;--cds-gap-xs:4px;--cds-pad-xl:24px;--cds-pad-lg:16px;--cds-pad-md:12px;--cds-pad-sm:8px;--cds-pad-xs:4px;
--cds-radius:8px;--cds-font-size-caption:12px;--cds-font-size-body:14px;--cds-font-size-heading:16px;--cds-font-size-title:24px;--cds-font-weight-medium:500;
--cds-dur-slow:300ms;--cds-ease-out:cubic-bezier(.2,.8,.2,1);
--font-anthropic-sans:ui-sans-serif,system-ui,-apple-system,'Segoe UI',Roboto,sans-serif;--font-anthropic-serif:Georgia,'Times New Roman',serif;"""
COLORS = "['#2c7fb8','#e8853a','#3a9e5c','#c8424a','#8a63c7','#9a6b4b','#d0679f','#7d7a74']"

for mode in ('plain', 'str'):
    conv = "typeof v==='object'?JSON.stringify(v):String(v)" if mode == 'str' else 'v'
    html = f"""<!doctype html><html><head><meta charset="utf-8"><meta name="viewport" content="width=device-width, initial-scale=1, viewport-fit=cover"><style>
:root{{{TOK_L}{TOK_C}}}
@media (prefers-color-scheme: dark){{:root{{{TOK_D}}}}}
body{{margin:0;background:var(--color-bg);color:var(--color-fg);font:14px var(--font-anthropic-sans)}}
#toolbar{{height:44px;border-bottom:1px solid var(--color-border-line);display:flex;align-items:center;padding:0 16px;font-size:13px;color:var(--color-fg-muted)}}
#dash-root{{padding:16px var(--dash-gutter,16px) 40px;box-sizing:border-box}}
.dash-skeleton{{background:var(--color-border-line);border-radius:4px;color:transparent}}
</style></head><body>
<div id="toolbar">Dashboard toolbar (mock)</div>
<script>window.__errs=[];window.onerror=(m,s,l,c,e)=>{{__errs.push(m+' @'+l+':'+c+' '+(e&&e.stack||''))}};const _ce=console.error;console.error=(...a)=>{{__errs.push('console.error '+a.map(x=>x&&x.stack||String(x)).join(' '));_ce(...a)}};</script>
<script src="d3.min.js"></script><script src="bundle.js"></script>
<script>
const ASYNC={1 if os.environ.get('ASYNC') else 0}; const __loaded={{}}, __req={{}};
const cache={{}};
window.__links=[];
window.dash={{colors:{COLORS},_c:{{}},_cb:[],
 data(id){{ if(ASYNC && id in __DATA && !__loaded[id]){{ if(!__req[id]){{ __req[id]=1; const dl=/^timeline_/.test(id)?500+Math.random()*900:(['icons','treeart','tooltips'].includes(id)?300+Math.random()*500:60+Math.random()*400); setTimeout(()=>{{__loaded[id]=1; window.dash._cb.forEach(f=>f());}}, dl);}} return {{status:'loading',data:[],refreshing:false}}; }}
  if(id in __DATA){{ if(!cache['__s'+id]) cache['__s'+id]=__DATA[id].map(r=>Object.fromEntries(Object.entries(r).map(([k,v])=>[k, v==null?v: {conv}]))); return {{status:'ok',data:cache['__s'+id],columns:Object.keys(__DATA[id][0]||{{}})}}; }} const c=this._c[id]; if(c){{ const ins=c.inputs.map(i=>this.data(i)); if(ins.some(r=>r.status!=='ok')) return {{status:'loading',data:[]}}; if(!cache[id]) cache[id]=c.fn(...c.inputs.map(i=>this.data(i).data)); return {{status:'ok',data:cache[id]}};}} return {{status:'missing',data:[]}}; }},
 calc(id,o){{this._c[id]=o}}, onData(fn){{this._cb.push(fn); fn(); addEventListener('resize',()=>fn());}}, setLink(n){{__links.push(n)}}, params(){{return {{}}}}, setParams(){{}}}};
</script>
<div id="dash-root">{page}</div>
</body></html>"""
    open(os.path.join(H, f'{pre}_{mode}.html'), 'w').write(html)
print('ok', len(page))
