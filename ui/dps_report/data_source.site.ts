// Where the multi-file site (https://forever-dps.pages.dev, dist/dps_report_site) gets its datasets: one file per report
// mode (DPS, Tanking, Healing, and the common one they share), fetched the first time the page asks for any dataset in
// it, and one file per preset's sample fights. build_artifact.mjs writes the files and puts their names in the page as
// window.fvSite (see siteData in vite.dps_report.config.mts).
import manifest from './datasets.json';

export { manifest };

// True on the site too: it has no sim UI around it to link back to.
export const standalone = true;

interface SiteMap {
	modes: Record<string, string>;
	events: Record<string, string>;
}

const site = (): SiteMap => (window as unknown as { fvSite?: SiteMap }).fvSite ?? { modes: {}, events: {} };
const bundles = new Map<string, Promise<Record<string, unknown>>>();

const fetchOk = async (url: string) => {
	const res = await fetch(url);
	if (!res.ok) throw new Error(`${res.status} ${res.statusText} for ${url}`);
	return res;
};

const bundle = (mode: string) => {
	let p = bundles.get(mode);
	if (!p) {
		const url = site().modes[mode];
		p = url ? fetchOk(url).then(res => res.json()) : Promise.reject(new Error(`No data file for mode ${mode}`));
		// A failed fetch can be tried again when the page next asks.
		p.catch(() => bundles.delete(mode));
		bundles.set(mode, p);
	}
	return p;
};

export async function loadDataset(id: string): Promise<unknown> {
	const info = (manifest as Record<string, { mode?: string }>)[id];
	const sets = await bundle(info?.mode || 'dps');
	if (!(id in sets)) throw new Error(`No data file for ${id}`);
	return sets[id];
}

// A preset's sample fights (see events.ts), still gzipped, or null when it has none.
export async function loadEvents(build: string): Promise<Uint8Array | null> {
	const url = site().events[build];
	if (!url) return null;
	return new Uint8Array(await (await fetchOk(url)).arrayBuffer());
}
