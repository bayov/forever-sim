import { SimRun } from './proto/ui';

// The last run's results stay in the browser, so a page refresh shows them again in the
// Results tab and the sidebar.
//
// We keep them in IndexedDB and not in localStorage. A run holds the logs of its typical,
// lucky and unlucky iterations, which are about 250 KB each, and localStorage only has about
// 5 MB for the whole site. Each sim keeps its own last run under its storage key.

const DB_NAME = 'forever-sim';
const STORE_NAME = 'lastSimRun';
// The results of each run in the sim history, for the Compare tab, see SimHistory.
const RUNS_STORE_NAME = 'simRuns';
// Version 3 kept the DPS and resources of each run over the fight in this store. We dropped
// them, so version 4 deletes it.
const OLD_SERIES_STORE_NAME = 'simRunSeries';

function openDb(): Promise<IDBDatabase> {
	return new Promise((resolve, reject) => {
		const request = indexedDB.open(DB_NAME, 4);
		request.onupgradeneeded = () => {
			const db = request.result;
			[STORE_NAME, RUNS_STORE_NAME].forEach(name => {
				if (!db.objectStoreNames.contains(name)) db.createObjectStore(name);
			});
			if (db.objectStoreNames.contains(OLD_SERIES_STORE_NAME)) db.deleteObjectStore(OLD_SERIES_STORE_NAME);
		};
		request.onsuccess = () => resolve(request.result);
		request.onerror = () => reject(request.error);
	});
}

async function withStore<T>(mode: IDBTransactionMode, use: (store: IDBObjectStore) => IDBRequest<T>, storeName = STORE_NAME): Promise<T> {
	const db = await openDb();
	try {
		return await new Promise<T>((resolve, reject) => {
			const request = use(db.transaction(storeName, mode).objectStore(storeName));
			request.onsuccess = () => resolve(request.result);
			request.onerror = () => reject(request.error);
		});
	} finally {
		db.close();
	}
}

export async function saveLastSimRun(key: string, run: SimRun): Promise<void> {
	try {
		await withStore('readwrite', store => store.put(SimRun.toBinary(run), key));
	} catch (e) {
		console.warn('Failed to save the last sim run: ' + e);
	}
}

export async function deleteLastSimRun(key: string): Promise<void> {
	try {
		await withStore('readwrite', store => store.delete(key));
	} catch (e) {
		console.warn('Failed to delete the last sim run: ' + e);
	}
}

// Returns null when there's no saved run, or when we can't read it. A run saved by an older
// version of the sim still reads, because protobuf skips the fields it doesn't know.
export async function loadLastSimRun(key: string): Promise<SimRun | null> {
	try {
		const data = await withStore<Uint8Array | undefined>('readonly', store => store.get(key));
		return data ? SimRun.fromBinary(data) : null;
	} catch (e) {
		console.warn('Failed to load the last sim run: ' + e);
		return null;
	}
}

// The results of each run in the sim history, so the Compare tab can compare them, and a
// Ctrl+click on a run shows its results again with its events.
//
// A run holds the logs of its typical, lucky and unlucky iterations (and in a leveling run the
// one with the most deaths), and they are most of its size. A 180s fight at level 60 has about
// 1.2 MB of logs, and a 20 minute leveling session about 8 MB. They are plain text, so we gzip
// them on their own, which makes them about 9 times smaller. The Compare tab doesn't show
// events, so it reads the results without them.
//
// Runs saved before we kept their logs are a bare SimRun, and they read without logs.
interface SavedHistoryRun {
	run: Uint8Array;
	// The IterationLogs, as gzipped JSON.
	logs?: Uint8Array;
}

// The logs of a run: the typical iteration's in the result, then the iterations it picked.
interface IterationLogs {
	result: string;
	typical: string;
	lucky: string;
	unlucky: string;
	died: string;
}

function iterationsOf(run: SimRun) {
	const result = run.result!;
	return {
		typical: result.typicalIteration,
		lucky: result.luckyIteration,
		unlucky: result.unluckyIteration,
		died: result.diedIteration,
	};
}

async function gzip(text: string): Promise<Uint8Array> {
	const stream = new Blob([text]).stream().pipeThrough(new CompressionStream('gzip'));
	return new Uint8Array(await new Response(stream).arrayBuffer());
}

async function gunzip(data: Uint8Array): Promise<string> {
	return new Response(new Blob([data]).stream().pipeThrough(new DecompressionStream('gzip'))).text();
}

export async function saveHistoryRun(key: string, run: SimRun): Promise<void> {
	try {
		const copy = SimRun.clone(run);
		let logs: Uint8Array | undefined;
		if (copy.result) {
			const iterations = iterationsOf(copy);
			const text: IterationLogs = {
				result: copy.result.logs,
				typical: iterations.typical?.logs ?? '',
				lucky: iterations.lucky?.logs ?? '',
				unlucky: iterations.unlucky?.logs ?? '',
				died: iterations.died?.logs ?? '',
			};
			copy.result.logs = '';
			Object.values(iterations).forEach(iteration => {
				if (iteration) iteration.logs = '';
			});
			if (Object.values(text).some(Boolean)) logs = await gzip(JSON.stringify(text));
		}
		const saved: SavedHistoryRun = { run: SimRun.toBinary(copy), ...(logs ? { logs } : {}) };
		await withStore('readwrite', store => store.put(saved, key), RUNS_STORE_NAME);
	} catch (e) {
		console.warn('Failed to save the results of a sim run: ' + e);
	}
}

export async function loadHistoryRun(key: string, withLogs = false): Promise<SimRun | null> {
	try {
		const data = await withStore<Uint8Array | SavedHistoryRun | undefined>('readonly', store => store.get(key), RUNS_STORE_NAME);
		if (!data) return null;
		if (data instanceof Uint8Array) return SimRun.fromBinary(data);
		const run = SimRun.fromBinary(data.run);
		if (withLogs && data.logs && run.result) {
			const text: IterationLogs = JSON.parse(await gunzip(data.logs));
			run.result.logs = text.result;
			const iterations = iterationsOf(run);
			(Object.keys(iterations) as (keyof typeof iterations)[]).forEach(kind => {
				const iteration = iterations[kind];
				if (iteration) iteration.logs = text[kind];
			});
		}
		return run;
	} catch (e) {
		console.warn('Failed to load the results of a sim run: ' + e);
		return null;
	}
}

export async function deleteHistoryRun(key: string): Promise<void> {
	try {
		await withStore('readwrite', store => store.delete(key), RUNS_STORE_NAME);
	} catch (e) {
		console.warn('Failed to delete the results of a sim run: ' + e);
	}
}

// The keys of every run we kept results for, across all sims.
export async function historyRunKeys(): Promise<string[]> {
	try {
		return (await withStore<IDBValidKey[]>('readonly', store => store.getAllKeys(), RUNS_STORE_NAME)).map(String);
	} catch (e) {
		console.warn('Failed to list the saved sim runs: ' + e);
		return [];
	}
}
