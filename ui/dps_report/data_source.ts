// Where the sim UI page gets its datasets: each data/<id>.json is its own asset, fetched the first time the report
// asks for it. The single-file build swaps this module for data_source.inline.ts (see vite.dps_report.config.mts).
import manifest from './datasets.json';

const urls = import.meta.glob<string>('./data/*.json', { query: '?url', import: 'default', eager: true });

export { manifest };

// True in the single-file artifact, which has no sim UI around it to link back to.
export const standalone = false;

export async function loadDataset(id: string): Promise<unknown> {
	const url = urls[`./data/${id}.json`];
	if (!url) throw new Error(`No data file for ${id}`);
	const res = await fetch(url);
	if (!res.ok) throw new Error(`${res.status} ${res.statusText} for ${url}`);
	return res.json();
}
