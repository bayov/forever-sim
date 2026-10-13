// Builds the DPS Report (ui/dps_report) for one self-contained HTML file, a Claude artifact that opens on a phone with
// no dev server. tools/dps_report/build_artifact.mjs runs it and then inlines the output into dist/dps_report.html.
//
// It differs from the sim UI build (vite.config.mts) in three ways: the report's only entry is ui/dps_report/index.html;
// the datasets are bundled into the script (data_source.inline.ts) instead of fetched; and everything ends up in one
// script and one stylesheet, with no chunks to load later.
import path from 'path';
import { defineConfig } from 'vite';

export const ARTIFACT_OUT_DIR = path.join(__dirname, 'dist', 'dps_report_artifact');

export default defineConfig({
	base: '/forever/',
	root: path.join(__dirname, 'ui'),
	publicDir: false,
	// JSON.parse of a string loads the 9 MB of data faster than the same data as an object literal.
	json: { stringify: true },
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
