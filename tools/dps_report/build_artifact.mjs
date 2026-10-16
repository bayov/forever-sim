// Builds the DPS Report as one self-contained HTML file, for publishing as a Claude artifact that opens on a phone with
// no dev server. The sim UI serves the same report at /forever/dps_report/ from the same source (ui/dps_report).
//
// It runs vite with vite.dps_report.config.mts, then puts the script, the stylesheet and every image and font they use
// into the page. An artifact may only load scripts from a few CDNs (d3 comes from cdnjs) and stylesheets from Google
// Fonts, so everything else has to be inside the file, which must stay under 16 MB. That includes Font Awesome, which
// the sim UI's events view on the Specs page draws its buttons with. We download it once into dist/fontawesome/.
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
const built = url => fs.readFileSync(path.join(buildDir, url.replace(/^\/forever\//, '')), 'utf8');

// The sim UI's own images and fonts (/forever/assets/... is the repo's assets/ folder) become data URIs.
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
	text.replace(/\/forever\/assets\/[\w./-]+?\.(png|jpe?g|svg|woff2?|ico)\b/g, url => {
		const file = path.join(root, 'assets', url.slice('/forever/assets/'.length));
		if (!fs.existsSync(file)) {
			missing.add(url);
			return url;
		}
		return `data:${MIME[path.extname(file).toLowerCase()]};base64,${fs.readFileSync(file).toString('base64')}`;
	});

const scriptUrl = html.match(/<script type="module"[^>]*src="([^"]+)"/)?.[1];
const styleUrl = html.match(/<link rel="stylesheet"[^>]*href="(\/forever\/[^"]+\.css)"/)?.[1];
if (!scriptUrl || !styleUrl) throw new Error('vite output has no report script or stylesheet');

const title = html.match(/<title>([^<]*)<\/title>/)?.[1] ?? 'DPS Report';
const cdnScripts = [...html.matchAll(/<script src="(https:\/\/cdnjs\.cloudflare\.com\/[^"]+)"><\/script>/g)].map(m => m[1]);
// Inside an inline <script>, "</script" would end it early. In the bundle it only shows up inside strings.
const script = inlineAssets(built(scriptUrl)).replace(/<\/script/gi, '<\\/script');
const style = inlineAssets(built(styleUrl)).replace(/<\/style/gi, '<\\/style');
const body = html.match(/<body>([\s\S]*)<\/body>/)?.[1].trim() ?? '';

// Font Awesome isn't an allowed stylesheet in an artifact, so its stylesheet and its solid and regular fonts go inside
// the page. The page uses no brand icons, and every browser we care about reads woff2, so we leave out the rest.
const FA_VERSION = '6.0.0';
const FA_URL = `https://cdnjs.cloudflare.com/ajax/libs/font-awesome/${FA_VERSION}`;
const faDir = path.join(root, 'dist/fontawesome', FA_VERSION);
const faFile = async name => {
	const file = path.join(faDir, name);
	if (!fs.existsSync(file)) {
		const res = await fetch(`${FA_URL}/${name}`);
		if (!res.ok) throw new Error(`${res.status} for ${FA_URL}/${name}`);
		fs.mkdirSync(path.dirname(file), { recursive: true });
		fs.writeFileSync(file, Buffer.from(await res.arrayBuffer()));
	}
	return fs.readFileSync(file);
};
const faFonts = {
	400: `data:font/woff2;base64,${(await faFile('webfonts/fa-regular-400.woff2')).toString('base64')}`,
	900: `data:font/woff2;base64,${(await faFile('webfonts/fa-solid-900.woff2')).toString('base64')}`,
};
const fontAwesome = (await faFile('css/all.min.css'))
	.toString('utf8')
	.replace(/@font-face\{[^}]*\}/g, face =>
		face.includes('"Font Awesome 6 Free"') ? face.replace(/src:[^;}]*/, `src:url(${faFonts[face.match(/font-weight:(\d+)/)[1]]}) format("woff2")`) : '',
	);

// The artifact host adds the doctype, <head> (charset, viewport) and <body>, so the file holds only the page's own tags.
const page = `<title>${title}</title>
${cdnScripts.map(src => `<script src="${src}"></script>`).join('\n')}
<style>
${fontAwesome}
${style}
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
