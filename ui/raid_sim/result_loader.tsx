// Where a comp search's result comes from: a file we pick, a file dropped anywhere on the page, or a
// ?result=<url> parameter, which is how a guild shares a search as a static page.
import { CompSearchResult } from '../core/proto/api';

export const RESULT_PARAM = 'result';

export const parseCompSearchResult = (text: string): CompSearchResult => {
	let json: unknown;
	try {
		json = JSON.parse(text);
	} catch (e) {
		throw new Error(`It isn't JSON: ${e instanceof Error ? e.message : e}`);
	}
	if (!json || typeof json != 'object' || !('scenarios' in json || 'request' in json || 'error' in json)) {
		throw new Error('It isn’t a comp search result (a CompSearchResult, like tools/compsearch’s <out>.json).');
	}
	return CompSearchResult.fromJson(json as any, { ignoreUnknownFields: true });
};

export class ResultLoader {
	readonly rootElem: HTMLElement;
	private readonly onLoad: (result: CompSearchResult, source: string) => void;
	private readonly fileInput: HTMLInputElement;
	private readonly statusElem: HTMLElement;

	constructor(parent: HTMLElement, onLoad: (result: CompSearchResult, source: string) => void) {
		this.onLoad = onLoad;
		this.fileInput = (<input type="file" accept=".json,application/json" className="d-none" />) as HTMLInputElement;
		this.statusElem = (<div className="rs-loader-status" />) as HTMLElement;
		const pick = (
			<button className="btn btn-primary" type="button">
				<i className="fas fa-file-import me-1" />
				Open a result
			</button>
		) as HTMLButtonElement;
		pick.addEventListener('click', () => this.fileInput.click());
		this.fileInput.addEventListener('change', () => {
			const file = this.fileInput.files?.[0];
			if (file) this.loadFile(file);
			this.fileInput.value = '';
		});

		this.rootElem = (
			<div className="rs-loader">
				{pick}
				{this.fileInput}
				<span className="rs-loader-hint">or drop a search’s JSON anywhere on the page</span>
				{this.statusElem}
			</div>
		) as HTMLElement;
		parent.appendChild(this.rootElem);

		// Dropping a file anywhere on the page loads it.
		let depth = 0;
		const hasFiles = (e: DragEvent) => !!e.dataTransfer && [...e.dataTransfer.types].includes('Files');
		document.addEventListener('dragenter', e => {
			if (!hasFiles(e)) return;
			depth++;
			document.body.classList.add('rs-dropping');
		});
		document.addEventListener('dragleave', e => {
			if (!hasFiles(e)) return;
			depth = Math.max(0, depth - 1);
			if (depth == 0) document.body.classList.remove('rs-dropping');
		});
		document.addEventListener('dragover', e => {
			if (hasFiles(e)) e.preventDefault();
		});
		document.addEventListener('drop', e => {
			if (!hasFiles(e)) return;
			e.preventDefault();
			depth = 0;
			document.body.classList.remove('rs-dropping');
			const file = e.dataTransfer!.files[0];
			if (file) this.loadFile(file);
		});
	}

	// What the tab shows before a result is open.
	emptyState() {
		return (
			<div className="rs-empty-state">
				<i className="fas fa-people-group rs-empty-state-icon" />
				<h5>No search result open</h5>
				<p>
					Open the JSON a comp search wrote (<code>go run -tags with_db ./tools/compsearch -out name</code> writes <code>name.json</code>), drop it on
					the page, or link to one with <code>?result=&lt;url&gt;</code>. The Search tab runs a search when the native server serves this page.
				</p>
			</div>
		);
	}

	async loadFromUrlParameter() {
		const url = new URLSearchParams(window.location.search).get(RESULT_PARAM);
		if (!url) return;
		this.setStatus(`Loading ${url}…`);
		try {
			const res = await fetch(url);
			if (!res.ok) throw new Error(`${res.status} ${res.statusText}`);
			this.load(await res.text(), url);
		} catch (e) {
			this.setStatus(`Couldn’t load ${url}: ${e instanceof Error ? e.message : e}`, true);
		}
	}

	private async loadFile(file: File) {
		try {
			this.load(await file.text(), file.name);
		} catch (e) {
			this.setStatus(`Couldn’t read ${file.name}: ${e instanceof Error ? e.message : e}`, true);
		}
	}

	private load(text: string, source: string) {
		try {
			const result = parseCompSearchResult(text);
			this.setStatus('');
			this.onLoad(result, source);
		} catch (e) {
			this.setStatus(`Couldn’t open ${source}: ${e instanceof Error ? e.message : e}`, true);
		}
	}

	private setStatus(text: string, error = false) {
		this.statusElem.textContent = text;
		this.statusElem.classList.toggle('text-danger', error);
	}
}
