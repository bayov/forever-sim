// The `dash` object report.js reads its data through. The report was first written as a Claude dashboard artifact,
// whose host provides this API; here we provide it ourselves, over the datasets in data/*.json.
//
// dash.data(id) returns a dataset's rows (or a calculation's result) once loaded, and starts loading it on first use.
// dash.onData(fn) runs fn now and again whenever a dataset finishes loading or the report's width changes.

export interface DatasetInfo {
	title: string;
	description?: string;
	updated?: { at: string; by: string };
}

export type DatasetLoader = (id: string) => Promise<unknown>;

type Status = 'ok' | 'loading' | 'error' | 'missing';

interface Result {
	status: Status;
	data: any;
	message?: string;
	refreshing: boolean;
	meta: unknown[];
	columns: string[];
	rows: any;
	value: any;
}

interface Calc {
	inputs: string[];
	fn: (...inputs: any[]) => unknown;
	title?: string;
	description?: string;
}

// Series colours for charts, picked to read on the sim UI's dark slate panels.
const SERIES_COLORS = ['#7fb3f0', '#e0a335', '#74c46e', '#c48ae0', '#ef7aa4', '#4fc8c1', '#f08a62', '#a2a6ae'];

function columnsOf(data: unknown): string[] {
	if (!Array.isArray(data)) return [];
	const cols = new Set<string>();
	for (const row of data.slice(0, 50)) if (row && typeof row === 'object') Object.keys(row).forEach(k => cols.add(k));
	return [...cols];
}

const result = (status: Status, data: any, message?: string): Result => ({
	status,
	data,
	message,
	refreshing: false,
	meta: [],
	columns: columnsOf(data),
	rows: data,
	value: data,
});

export function createDash(manifest: Record<string, DatasetInfo>, load: DatasetLoader, root: HTMLElement) {
	const state = new Map<string, { status: Status; data: any; message?: string }>();
	const calcs = new Map<string, Calc>();
	const calcCache = new Map<string, { inputs: unknown[]; result: unknown }>();
	const listeners: Array<() => void> = [];
	let scheduled = false;

	const notify = () => {
		if (scheduled) return;
		scheduled = true;
		requestAnimationFrame(() => {
			scheduled = false;
			for (const fn of listeners) {
				try {
					fn();
				} catch (e) {
					console.error(e);
				}
			}
		});
	};

	const dataset = (id: string): Result => {
		if (!(id in manifest)) return result('missing', [], `No dataset named ${id}`);
		let s = state.get(id);
		if (!s) {
			const entry: { status: Status; data: any; message?: string } = { status: 'loading', data: [] };
			s = entry;
			state.set(id, entry);
			load(id)
				.then(
					data => {
						entry.status = 'ok';
						entry.data = data;
					},
					err => {
						console.error(`DPS report: failed to load ${id}`, err);
						entry.status = 'error';
						entry.message = String(err?.message ?? err);
					},
				)
				.then(notify);
		}
		return result(s.status, s.data, s.message);
	};

	const calc = (id: string): Result => {
		const c = calcs.get(id)!;
		const ins = c.inputs.map(i => data(i));
		const bad = ins.find(r => r.status !== 'ok');
		if (bad) return result(bad.status === 'missing' ? 'error' : bad.status, [], bad.message);
		const inData = ins.map(r => r.data);
		const cached = calcCache.get(id);
		if (cached && cached.inputs.every((d, i) => d === inData[i])) return result('ok', cached.result);
		try {
			const out = c.fn(...inData);
			calcCache.set(id, { inputs: inData, result: out });
			return result('ok', out);
		} catch (e) {
			console.error(`DPS report: calculation ${id} failed`, e);
			return result('error', [], String(e));
		}
	};

	const data = (id: string): Result => (calcs.has(id) ? calc(id) : dataset(id));

	let lastWidth = -1;
	new ResizeObserver(() => {
		if (root.clientWidth === lastWidth) return;
		lastWidth = root.clientWidth;
		notify();
	}).observe(root);

	return {
		data,
		onData(fn: () => void) {
			listeners.push(fn);
			fn();
		},
		calc(id: string, opts: Calc) {
			calcs.set(id, opts);
		},
		colors: SERIES_COLORS,
		// The page names its tab or preset in the address, so a link opens on it. push makes it a new history entry
		// (a new tab or preset, which Back returns from), and y is where the page being left was scrolled to, kept on
		// its entry so Back returns there.
		setLink(name: string, push = false, y?: number) {
			const url = name ? `#${name}` : location.pathname + location.search;
			try {
				if (push && location.hash !== (name ? `#${name}` : '')) {
					history.replaceState({ ...(history.state || {}), y }, '');
					history.pushState({}, '', url);
				} else {
					history.replaceState(history.state, '', url);
				}
			} catch {
				// A sandboxed frame may refuse history changes.
			}
		},
		// Back and Forward: the name the address now holds, and where that entry was scrolled to.
		onLink(fn: (name: string, y?: number) => void) {
			addEventListener('popstate', (e) => fn(location.hash.slice(1), e.state?.y));
		},
		manifest,
	};
}

export type Dash = ReturnType<typeof createDash>;
