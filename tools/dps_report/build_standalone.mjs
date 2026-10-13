// Builds the DPS Report as one self-contained HTML file (page, runtime, theme and every dataset inlined), for
// publishing as a Claude artifact that opens on a phone with no dev server. The sim UI serves the same report
// from ui/dps_report/ (see index.ts there).
//
// Usage: node tools/dps_report/build_standalone.mjs [out.html]   (default: dist/dps_report.html)
import fs from 'fs';
import path from 'path';
import { fileURLToPath } from 'url';

const root = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '../..');
const dir = path.join(root, 'ui/dps_report');
const out = path.resolve(process.argv[2] || path.join(root, 'dist/dps_report.html'));

const read = f => fs.readFileSync(path.join(dir, f), 'utf8');
// Safe inside an inline <script>: no "</script", no HTML comment openers.
const js = v => JSON.stringify(v).replace(/</g, '\\u003c');

const manifest = JSON.parse(read('datasets.json'));
const data = {};
for (const id of Object.keys(manifest)) data[id] = JSON.parse(read(`data/${id}.json`));

// The artifact host wraps the file in its own <!doctype>, <head> (charset, viewport) and <body>, so the file holds
// only the page's own tags.
const html = `<title>Forever DPS Report</title>
<style>
${read('theme.css')}
</style>
<script src="https://cdnjs.cloudflare.com/ajax/libs/d3/7.9.0/d3.min.js"></script>
<div id="dash-root"></div>
<script>
${read('runtime.js')}
</script>
<script>
(() => {
  const data = ${js(data)};
  DpsReport.install({ manifest: ${js(manifest)}, load: (id) => Promise.resolve(data[id]) });
  DpsReport.mount(document.getElementById('dash-root'), ${js(read('report.html'))});
})();
</script>
`;

fs.mkdirSync(path.dirname(out), { recursive: true });
fs.writeFileSync(out, html);
console.log(`Wrote ${path.relative(process.cwd(), out)} (${(html.length / 1e6).toFixed(1)} MB)`);
