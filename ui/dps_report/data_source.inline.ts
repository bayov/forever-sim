// Where the single-file artifact gets its datasets: every data/<id>.json is bundled into the page itself.
import manifest from './datasets.json';

const datasets = import.meta.glob('./data/*.json', { import: 'default', eager: true });

export { manifest };

// True in the single-file artifact, which has no sim UI around it to link back to.
export const standalone = true;

export async function loadDataset(id: string): Promise<unknown> {
	const data = datasets[`./data/${id}.json`];
	if (data === undefined) throw new Error(`No data file for ${id}`);
	return data;
}
