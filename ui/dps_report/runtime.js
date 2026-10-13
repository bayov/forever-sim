// A small stand-in for the `dash` API of the Claude dashboard artifact type, which report.html was written against.
// It lets the same report.html run in two hosts:
//   - the sim UI page (index.ts), which fetches each dataset from data/<id>.json on first use;
//   - the single-file artifact (tools/dps_report/build_standalone.mjs), which inlines every dataset.
// Only what report.html uses is implemented: data, onData, calc, colors and setLink (params are kept in memory).
(function () {
	const LIGHT = ['#2c7bd6', '#d9653b', '#3f9a3a', '#a052c2', '#c29512', '#22a09a', '#d0457a', '#6f7480'];
	const DARK = ['#6aa9f0', '#f08a62', '#74c46e', '#c48ae0', '#e6bb45', '#4fc8c1', '#ef7aa4', '#a2a6ae'];

	function isDark() {
		const t = document.documentElement.dataset.theme;
		if (t) return t === 'dark';
		return !!(window.matchMedia && window.matchMedia('(prefers-color-scheme: dark)').matches);
	}

	function columnsOf(data) {
		if (!Array.isArray(data)) return [];
		const cols = new Set();
		for (const row of data.slice(0, 50)) if (row && typeof row === 'object') Object.keys(row).forEach(k => cols.add(k));
		return [...cols];
	}

	// load(id) returns a promise of the dataset's parsed JSON; manifest maps dataset id to {title, description, updated}.
	function install({ load, manifest }) {
		const state = new Map(); // id -> {status, data, message}
		const calcs = new Map(); // id -> {inputs, fn, title, description}
		const calcCache = new Map(); // id -> {inputs: [data...], result}
		const listeners = [];
		let scheduled = false;
		let params = {};

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

		const result = (status, data, message) => ({ status, data, message, refreshing: false, meta: [], columns: columnsOf(data), rows: data, value: data });

		const dataset = id => {
			if (!(id in manifest)) return result('missing', [], `No dataset named ${id}`);
			let s = state.get(id);
			if (!s) {
				s = { status: 'loading', data: [] };
				state.set(id, s);
				Promise.resolve()
					.then(() => load(id))
					.then(
						data => {
							s.status = 'ok';
							s.data = data;
						},
						err => {
							console.error(`DPS report: failed to load ${id}`, err);
							s.status = 'error';
							s.message = String(err && err.message ? err.message : err);
						},
					)
					.then(notify);
			}
			return result(s.status, s.data, s.message);
		};

		const calc = id => {
			const c = calcs.get(id);
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

		const data = id => (calcs.has(id) ? calc(id) : dataset(id));

		let lastWidth = -1;
		const root = document.getElementById('dash-root');
		if (root && window.ResizeObserver) {
			new ResizeObserver(() => {
				const w = root.clientWidth;
				if (w !== lastWidth) {
					lastWidth = w;
					notify();
				}
			}).observe(root);
		}

		window.dash = {
			data,
			onData(fn) {
				listeners.push(fn);
				fn();
			},
			calc(id, opts) {
				calcs.set(id, opts);
			},
			loader() {},
			refresh() {},
			colors: isDark() ? DARK : LIGHT,
			params: () => ({ ...params }),
			setParams(p) {
				params = { ...params, ...p };
				notify();
			},
			resetParams() {
				params = {};
				notify();
			},
			setLink(name) {
				try {
					history.replaceState(null, '', name ? `#${name}` : location.pathname + location.search);
				} catch (e) {
					/* sandboxed frames may refuse history changes */
				}
			},
			manifest,
		};
	}

	// Puts report.html into root and runs its scripts in order (innerHTML alone doesn't run them).
	function mount(root, html) {
		root.innerHTML = html;
		for (const old of root.querySelectorAll('script')) {
			const s = document.createElement('script');
			s.textContent = old.textContent;
			old.replaceWith(s);
		}
	}

	window.DpsReport = { install, mount };
})();
