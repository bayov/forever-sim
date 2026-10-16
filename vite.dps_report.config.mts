// Builds the DPS Report (ui/dps_report) for one self-contained HTML file, a Claude artifact that opens on a phone with
// no dev server. tools/dps_report/build_artifact.mjs runs it and then inlines the output into dist/dps_report.html.
//
// It differs from the sim UI build (vite.config.mts) in three ways: the report's only entry is ui/dps_report/index.html;
// the datasets are bundled into the script (data_source.inline.ts) instead of fetched; and everything ends up in one
// script and one stylesheet, with no chunks to load later.
import fs from 'fs';
import path from 'path';
import { defineConfig, PluginOption } from 'vite';
import zlib from 'zlib';

// The sample fights of the Specs page (data/events/*.json.gz) go into the script as base64 strings, which
// data_source.inline.ts imports with ?base64.
//
// They are most of the file, and the file must stay under 16 MB. So we leave out two kinds of lines that only the log
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

function gzipAsBase64(): PluginOption {
	return {
		name: 'gzip-as-base64',
		enforce: 'pre',
		load(id) {
			const [file, query] = id.split('?');
			if (query != 'base64' || !file.endsWith('.json.gz')) return null;
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
			const packed = zlib.gzipSync(JSON.stringify(fights), { level: 9 });
			return `export default ${JSON.stringify(packed.toString('base64'))};`;
		},
	};
}

export const ARTIFACT_OUT_DIR = path.join(__dirname, 'dist', 'dps_report_artifact');

export default defineConfig({
	base: '/forever/',
	root: path.join(__dirname, 'ui'),
	publicDir: false,
	// JSON.parse of a string loads the 9 MB of data faster than the same data as an object literal.
	json: { stringify: true },
	plugins: [gzipAsBase64()],
	resolve: {
		alias: [{ find: /^\.\/data_source$/, replacement: path.join(__dirname, 'ui/dps_report/data_source.inline.ts') }],
	},
	esbuild: {
		jsxInject: "import { element, fragment } from 'tsx-vanilla';",
	},
	build: {
		outDir: ARTIFACT_OUT_DIR,
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
});
