import { SimRun } from './proto/ui';

// The last run's results stay in the browser, so a page refresh shows them again in the
// Results tab and the sidebar.
//
// We keep them in IndexedDB and not in localStorage. A run holds the logs of its typical,
// lucky and unlucky iterations, which are about 250 KB each, and localStorage only has about
// 5 MB for the whole site. Each sim keeps its own last run under its storage key.

const DB_NAME = 'forever-sim';
const STORE_NAME = 'lastSimRun';

function openDb(): Promise<IDBDatabase> {
	return new Promise((resolve, reject) => {
		const request = indexedDB.open(DB_NAME, 1);
		request.onupgradeneeded = () => request.result.createObjectStore(STORE_NAME);
		request.onsuccess = () => resolve(request.result);
		request.onerror = () => reject(request.error);
	});
}

async function withStore<T>(mode: IDBTransactionMode, use: (store: IDBObjectStore) => IDBRequest<T>): Promise<T> {
	const db = await openDb();
	try {
		return await new Promise<T>((resolve, reject) => {
			const request = use(db.transaction(STORE_NAME, mode).objectStore(STORE_NAME));
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
