/// <reference types="vite/client" />
// The DPS Report page in the sim UI. report.html is shared with the single-file artifact
// (tools/dps_report/build_standalone.mjs); runtime.js gives it the `dash` API it expects.
import './runtime.js';

import manifest from './datasets.json';
import reportHtml from './report.html?raw';

// Each dataset is emitted as its own asset and fetched the first time the report asks for it.
const urls = import.meta.glob<string>('./data/*.json', { query: '?url', import: 'default', eager: true });

const runtime = (window as any).DpsReport;
runtime.install({
	manifest,
	load: async (id: string) => {
		const url = urls[`./data/${id}.json`];
		if (!url) throw new Error(`No data file for ${id}`);
		const res = await fetch(url);
		if (!res.ok) throw new Error(`${res.status} ${res.statusText} for ${url}`);
		return res.json();
	},
});
runtime.mount(document.getElementById('dash-root')!, reportHtml);
