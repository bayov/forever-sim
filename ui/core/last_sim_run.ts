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

// The results of a run in the sim history, without the logs of its typical, lucky and unlucky
// iterations. The Compare tab doesn't show events, and without the logs a run takes about
// 50 KB instead of 750 KB.
export async function saveHistoryRun(key: string, run: SimRun): Promise<void> {
	const copy = SimRun.clone(run);
	if (copy.result) {
		copy.result.logs = '';
		[copy.result.typicalIteration, copy.result.luckyIteration, copy.result.unluckyIteration].forEach(iteration => {
			if (iteration) iteration.logs = '';
		});
	}
	try {
		await withStore('readwrite', store => store.put(SimRun.toBinary(copy), key), RUNS_STORE_NAME);
	} catch (e) {
		console.warn('Failed to save the results of a sim run: ' + e);
	}
}

export async function loadHistoryRun(key: string): Promise<SimRun | null> {
	try {
		const data = await withStore<Uint8Array | undefined>('readonly', store => store.get(key), RUNS_STORE_NAME);
		return data ? SimRun.fromBinary(data) : null;
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
