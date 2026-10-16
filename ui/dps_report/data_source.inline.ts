// Where the single-file artifact gets its datasets: every data/<id>.json is bundled into the page itself.
import manifest from './datasets.json';

const datasets = import.meta.glob('./data/*.json', { import: 'default', eager: true });
// The sample fights in base64 (see vite.dps_report.config.mts), decoded only when the page opens a preset.
const events = import.meta.glob<string>('./data/events/*.json.gz', { query: '?base64', import: 'default', eager: true });

export { manifest };

// True in the single-file artifact, which has no sim UI around it to link back to.
export const standalone = true;

export async function loadDataset(id: string): Promise<unknown> {
	const data = datasets[`./data/${id}.json`];
	if (data === undefined) throw new Error(`No data file for ${id}`);
	return data;
}

// A preset's sample fights (see events.ts), still gzipped, or null when it has none.
export async function loadEvents(build: string): Promise<Uint8Array | null> {
	const base64 = events[`./data/events/${build}.json.gz`];
	if (!base64) return null;
	const binary = atob(base64);
	const bytes = new Uint8Array(binary.length);
	for (let i = 0; i < binary.length; i++) bytes[i] = binary.charCodeAt(i);
	return bytes;
}
