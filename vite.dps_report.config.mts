// Builds the DPS Report (ui/dps_report) outside the sim UI. tools/dps_report/build_artifact.mjs runs it twice:
//   - the default mode, for one self-contained HTML file (dist/dps_report.html) that opens on a phone with no dev
//     server: the datasets are bundled into the script (data_source.inline.ts);
//   - `--mode site`, for the multi-file site deployed to Cloudflare Pages (dist/dps_report_site): the script carries no
//     data, and the siteData plugin below writes each report mode's datasets as one file (data/<mode>.<hash>.json) and
//     each preset's sample fights as another, which data_source.site.ts fetches when the page first needs them.
//
// Both differ from the sim UI build (vite.config.mts) in that the report's only entry is ui/dps_report/index.html and
// everything ends up in one script and one stylesheet, with no chunks to load later.
import fs from 'fs';
import path from 'path';
import crypto from 'crypto';
import { defineConfig, PluginOption } from 'vite';
import zlib from 'zlib';

// The sample fights of the Specs page (data/events/*.json.gz) go into the script as base64 strings, which
// data_source.inline.ts imports with ?base64.
//
// They are most of the file, so to keep it small we leave out two kinds of lines that only the log
// shows, which the timeline, the chart and the replay don't use. The sim UI page still has them.
//   - The rotation waiting: "No available actions!" and "Pausing GCD for 50ms due to rotation / CDs.". In a rogue
//     fight that's half the lines.
//   - The [DEBUG] lines with the damage formula of each hit. The log hides them unless we turn on Debug.
const LOG_ONLY = /^\[[-\d.]+\] \[[^\]]*\] (No available actions!$|Pausing GCD for \S+ due to rotation \/ CDs\.$|.*\[DEBUG\])/;

// We also leave out the hit for 0 damage and 0 threat when a DoT or debuff goes on (Rupture, Deep Wounds, Mind Flay).
// The sim no longer logs it (see dealDamageInternal in sim/core/spell_result.go), but most fights we carry are older
// than that.
const AURA_ONLY_HIT = /\} Hit for 0\.000 damage \(SpellSchool: \d+\)\. \(Threat: 0\.000\)$/;
const ITERATIONS = ['typicalIteration', 'unluckyIteration', 'luckyIteration'];

// A preset's sample fights (data/events/<build>.json.gz) without the log-only lines, gzipped again.
function trimFights(file: string): Buffer {
	const fights = JSON.parse(zlib.gunzipSync(fs.readFileSync(file)).toString('utf8'));
	for (const kind of ITERATIONS) {
		const iteration = fights.result[kind];
		if (iteration?.logs) {
			iteration.logs = iteration.logs
				.split('\n')
				.filter((line: string) => !LOG_ONLY.test(line) && !AURA_ONLY_HIT.test(line))
				.join('\n');
		}
	}
	return zlib.gzipSync(JSON.stringify(fights), { level: 9 });
}

function gzipAsBase64(): PluginOption {
	return {
		name: 'gzip-as-base64',
		enforce: 'pre',
		load(id) {
			const [file, query] = id.split('?');
			if (query != 'base64' || !file.endsWith('.json.gz')) return null;
			return `export default ${JSON.stringify(trimFights(file).toString('base64'))};`;
		},
	};
}

const DATA_DIR = path.join(__dirname, 'ui/dps_report/data');
const hashOf = (bytes: Buffer | string) => crypto.createHash('sha256').update(bytes).digest('hex').slice(0, 10);

// The site's data files, next to the page: one per report mode (datasets.json's "mode": dps when it has none, tank,
// heal, common), named by their content's hash so a new deploy never mixes old data with a new page, and each preset's
// sample fights, trimmed as in the single file. site.json maps each mode and preset to its file; build_artifact.mjs
// puts that map in the page (window.fvSite), where data_source.site.ts reads it.
function siteData(): PluginOption {
	return {
		name: 'fv-site-data',
		generateBundle() {
			const manifest = JSON.parse(fs.readFileSync(path.join(__dirname, 'ui/dps_report/datasets.json'), 'utf8'));
			const bundles: Record<string, Record<string, unknown>> = {};
			for (const [id, info] of Object.entries<{ mode?: string }>(manifest)) {
				const file = path.join(DATA_DIR, `${id}.json`);
				if (!fs.existsSync(file)) continue;
				(bundles[info.mode || 'dps'] ??= {})[id] = JSON.parse(fs.readFileSync(file, 'utf8'));
			}
			const map: { modes: Record<string, string>; events: Record<string, string> } = { modes: {}, events: {} };
			for (const [mode, sets] of Object.entries(bundles)) {
				const source = JSON.stringify(sets);
				const fileName = `data/${mode}.${hashOf(source)}.json`;
				this.emitFile({ type: 'asset', fileName, source });
				map.modes[mode] = fileName;
			}
			const eventDir = path.join(DATA_DIR, 'events');
			for (const name of fs.existsSync(eventDir) ? fs.readdirSync(eventDir).sort() : []) {
				if (!name.endsWith('.json.gz')) continue;
				const source = trimFights(path.join(eventDir, name));
				const build = name.slice(0, -'.json.gz'.length);
				const fileName = `data/events/${build}.${hashOf(source)}.json.gz`;
				this.emitFile({ type: 'asset', fileName, source });
				map.events[build] = fileName;
			}
			this.emitFile({ type: 'asset', fileName: 'site.json', source: JSON.stringify(map) });
		},
	};
}

export const ARTIFACT_OUT_DIR = path.join(__dirname, 'dist', 'dps_report_artifact');
export const SITE_OUT_DIR = path.join(__dirname, 'dist', 'dps_report_site_build');

export default defineConfig(({ mode }) => ({
	base: '/forever/',
	root: path.join(__dirname, 'ui'),
	publicDir: false,
	// JSON.parse of a string loads the 9 MB of data faster than the same data as an object literal.
	json: { stringify: true },
	plugins: mode == 'site' ? [siteData()] : [gzipAsBase64()],
	resolve: {
		alias: [
			{
				find: /^\.\/data_source$/,
				replacement: path.join(__dirname, `ui/dps_report/data_source.${mode == 'site' ? 'site' : 'inline'}.ts`),
			},
		],
	},
	esbuild: {
		jsxInject: "import { element, fragment } from 'tsx-vanilla';",
	},
	build: {
		outDir: mode == 'site' ? SITE_OUT_DIR : ARTIFACT_OUT_DIR,
		emptyOutDir: true,
		target: ['es2020'],
		minify: 'terser',
		cssCodeSplit: false,
		modulePreload: false,
		rollupOptions: {
			input: path.join(__dirname, 'ui/dps_report/index.html'),
			output: {
				inlineDynamicImports: true,
				entryFileNames: 'report.js',
				assetFileNames: '[name][extname]',
			},
		},
	},
}));
