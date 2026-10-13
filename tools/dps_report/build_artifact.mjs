// Builds the DPS Report as one self-contained HTML file, for publishing as a Claude artifact that opens on a phone with
// no dev server. The sim UI serves the same report at /classic/dps_report/ from the same source (ui/dps_report).
//
// It runs vite with vite.dps_report.config.mts, then puts the script, the stylesheet and every image and font they use
// into the page. An artifact may only load scripts from a few CDNs (d3 comes from cdnjs) and stylesheets from Google
// Fonts, so everything else has to be inside the file, which must stay under 16 MB.
//
// Usage: node tools/dps_report/build_artifact.mjs [out.html]   (default: dist/dps_report.html; also `make dps-report-artifact`)
import fs from 'fs';
import path from 'path';
import { fileURLToPath } from 'url';
import { build } from 'vite';

const root = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '../..');
const out = path.resolve(process.argv[2] || path.join(root, 'dist/dps_report.html'));
const configFile = path.join(root, 'vite.dps_report.config.mts');
const buildDir = path.join(root, 'dist/dps_report_artifact');
const MAX_BYTES = 16 * 1024 * 1024;

await build({ configFile, logLevel: 'error' });

const html = fs.readFileSync(path.join(buildDir, 'dps_report/index.html'), 'utf8');
const built = url => fs.readFileSync(path.join(buildDir, url.replace(/^\/classic\//, '')), 'utf8');

// The sim UI's own images and fonts (/classic/assets/... is the repo's assets/ folder) become data URIs.
const MIME = {
	'.png': 'image/png',
	'.jpg': 'image/jpeg',
	'.jpeg': 'image/jpeg',
	'.svg': 'image/svg+xml',
	'.woff2': 'font/woff2',
	'.woff': 'font/woff',
	'.ico': 'image/x-icon',
};
const missing = new Set();
const inlineAssets = text =>
	text.replace(/\/classic\/assets\/[\w./-]+?\.(png|jpe?g|svg|woff2?|ico)\b/g, url => {
		const file = path.join(root, 'assets', url.slice('/classic/assets/'.length));
		if (!fs.existsSync(file)) {
			missing.add(url);
			return url;
		}
		return `data:${MIME[path.extname(file).toLowerCase()]};base64,${fs.readFileSync(file).toString('base64')}`;
	});

const scriptUrl = html.match(/<script type="module"[^>]*src="([^"]+)"/)?.[1];
const styleUrl = html.match(/<link rel="stylesheet"[^>]*href="(\/classic\/[^"]+\.css)"/)?.[1];
if (!scriptUrl || !styleUrl) throw new Error('vite output has no report script or stylesheet');

const title = html.match(/<title>([^<]*)<\/title>/)?.[1] ?? 'DPS Report';
const cdnScripts = [...html.matchAll(/<script src="(https:\/\/cdnjs\.cloudflare\.com\/[^"]+)"><\/script>/g)].map(m => m[1]);
// Inside an inline <script>, "</script" would end it early. In the bundle it only shows up inside strings.
const script = inlineAssets(built(scriptUrl)).replace(/<\/script/gi, '<\\/script');
const style = inlineAssets(built(styleUrl)).replace(/<\/style/gi, '<\\/style');
const body = html.match(/<body>([\s\S]*)<\/body>/)?.[1].trim() ?? '';

// Font Awesome isn't an allowed stylesheet in an artifact. The report draws no icons with it; only the close button of
// a pinned tooltip does, so that one gets a plain cross.
const noFontAwesome = '.sticky-tooltip-close .fa-times::before { content: "\\2715"; font-style: normal; }';

// The artifact host adds the doctype, <head> (charset, viewport) and <body>, so the file holds only the page's own tags.
const page = `<title>${title}</title>
${cdnScripts.map(src => `<script src="${src}"></script>`).join('\n')}
<style>
${style}
${noFontAwesome}
</style>
${body}
<script type="module">
${script}
</script>
`;

if (missing.size) console.warn(`Not found in assets/, left as links: ${[...missing].join(', ')}`);
fs.mkdirSync(path.dirname(out), { recursive: true });
fs.writeFileSync(out, page);
const bytes = Buffer.byteLength(page);
console.log(`Wrote ${path.relative(process.cwd(), out)} (${(bytes / 1024 / 1024).toFixed(1)} MB)`);
if (bytes > MAX_BYTES) {
	console.error('Too big for an artifact (16 MB at most).');
	process.exit(1);
}
