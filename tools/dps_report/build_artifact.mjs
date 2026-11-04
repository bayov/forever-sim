// Builds the DPS Report for outside the sim UI, in two forms. The sim UI serves the same report at /forever/dps_report/
// from the same source (ui/dps_report).
//
//   - dist/dps_report.html: one self-contained HTML file that opens on a phone with no dev server, even from file://.
//     Every dataset and sample fight is inside it.
//   - dist/dps_report_site/: the site `make dps-report-deploy` publishes to Cloudflare Pages
//     (https://forever-dps.pages.dev). index.html is the same page without the data; the data is one file per report
//     mode (data/<mode>.<hash>.json: dps, tank, heal and common) and one per preset's sample fights
//     (data/events/<build>.<hash>.json.gz), which the page fetches when it first needs them (data_source.site.ts). So a
//     Tanking link loads only the Tanking and common data.
//
// Each runs vite with vite.dps_report.config.mts (the site with --mode site), then puts the script, the stylesheet and
// every image and font they use into the page (d3 comes from cdnjs, the fonts from Google Fonts). Cloudflare Pages
// takes a file of up to 25 MiB, so the build fails above that. That includes Font Awesome, which
// the sim UI's events view on the Specs page draws its buttons with. We download it once into dist/fontawesome/.
//
// Usage: node tools/dps_report/build_artifact.mjs [out.html]   (default: dist/dps_report.html; also `make dps-report-artifact`)
import fs from 'fs';
import path from 'path';
import { fileURLToPath } from 'url';
import { build } from 'vite';

const root = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '../..');
const out = path.resolve(process.argv[2] || path.join(root, 'dist/dps_report.html'));
const siteOut = path.join(root, 'dist/dps_report_site');
const configFile = path.join(root, 'vite.dps_report.config.mts');
const buildDir = path.join(root, 'dist/dps_report_artifact');
const siteBuildDir = path.join(root, 'dist/dps_report_site_build');
const MAX_BYTES = 25 * 1024 * 1024; // Cloudflare Pages' limit for one file

await build({ configFile, logLevel: 'error' });
await build({ configFile, logLevel: 'error', mode: 'site' });

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

// One built page (vite's output in dir) as the page's own tags, with everything it loads inside. extra goes before the
// report's script.
const assemble = (dir, extra = '') => {
	const html = fs.readFileSync(path.join(dir, 'dps_report/index.html'), 'utf8');
	const built = url => fs.readFileSync(path.join(dir, url.replace(/^\/forever\//, '')), 'utf8');
	const scriptUrl = html.match(/<script type="module"[^>]*src="([^"]+)"/)?.[1];
	const styleUrl = html.match(/<link rel="stylesheet"[^>]*href="(\/forever\/[^"]+\.css)"/)?.[1];
	if (!scriptUrl || !styleUrl) throw new Error('vite output has no report script or stylesheet');

	const title = html.match(/<title>([^<]*)<\/title>/)?.[1] ?? 'DPS Report';
	const cdnScripts = [...html.matchAll(/<script src="(https:\/\/cdnjs\.cloudflare\.com\/[^"]+)"><\/script>/g)].map(m => m[1]);
	// Inside an inline <script>, "</script" would end it early. In the bundle it only shows up inside strings.
	const script = inlineAssets(built(scriptUrl)).replace(/<\/script/gi, '<\\/script');
	const style = inlineAssets(built(styleUrl)).replace(/<\/style/gi, '<\\/style');
	const body = html.match(/<body>([\s\S]*)<\/body>/)?.[1].trim() ?? '';
	return `<title>${title}</title>
${cdnScripts.map(src => `<script src="${src}"></script>`).join('\n')}
<style>
${fontAwesome}
${style}
</style>
${body}
${extra}<script type="module">
${script}
</script>
`;
};

// The single file. The artifact host adds the doctype, <head> (charset, viewport) and <body>, so the file holds only the
// page's own tags; deploy_pages.py adds the same wrapper when it deploys a file like this one.
const page = assemble(buildDir);
if (missing.size) console.warn(`Not found in assets/, left as links: ${[...missing].join(', ')}`);
fs.mkdirSync(path.dirname(out), { recursive: true });
fs.writeFileSync(out, page);
const bytes = Buffer.byteLength(page);
console.log(`Wrote ${path.relative(process.cwd(), out)} (${(bytes / 1024 / 1024).toFixed(1)} MB)`);
if (bytes > MAX_BYTES) {
	console.error('Too big for Cloudflare Pages (25 MiB a file at most).');
	process.exit(1);
}

// The site: the page in a full document (as deploy_pages.py would wrap the single file), with the names of its data
// files, and the data files next to it.
const siteMap = fs.readFileSync(path.join(siteBuildDir, 'site.json'), 'utf8');
const ui = path.join(root, 'ui/dps_report');
const head =
	'<!doctype html><html lang="en"><head><meta charset="utf-8">' +
	'<meta name="viewport" content="width=device-width, initial-scale=1">' +
	`<link rel="icon" type="image/png" sizes="64x64" href="data:image/png;base64,${fs.readFileSync(path.join(ui, 'favicon.png')).toString('base64')}">` +
	`<link rel="icon" type="image/svg+xml" href="data:image/svg+xml;base64,${fs.readFileSync(path.join(ui, 'favicon.svg')).toString('base64')}"></head><body>`;
const sitePage = head + assemble(siteBuildDir, `<script>window.fvSite = ${siteMap};</script>\n`) + '</body></html>';
fs.rmSync(siteOut, { recursive: true, force: true });
fs.mkdirSync(siteOut, { recursive: true });
fs.cpSync(path.join(siteBuildDir, 'data'), path.join(siteOut, 'data'), { recursive: true });
fs.writeFileSync(path.join(siteOut, 'index.html'), sitePage);
const siteFiles = fs.readdirSync(siteOut, { recursive: true }).filter(f => fs.statSync(path.join(siteOut, f)).isFile());
const siteBytes = siteFiles.reduce((n, f) => n + fs.statSync(path.join(siteOut, f)).size, 0);
const biggest = Math.max(...siteFiles.map(f => fs.statSync(path.join(siteOut, f)).size));
console.log(
	`Wrote ${path.relative(process.cwd(), siteOut)}/: ${siteFiles.length} files, ${(siteBytes / 1024 / 1024).toFixed(1)} MB, ` +
		`index.html ${(Buffer.byteLength(sitePage) / 1024 / 1024).toFixed(1)} MB`,
);
if (biggest > MAX_BYTES) {
	console.error('A site file is too big for Cloudflare Pages (25 MiB a file at most).');
	process.exit(1);
}
