// Merges a partial regeneration into the DPS Report's datasets.
//
// The pipeline can write datasets for only the builds it re-simmed, into a scratch folder with the same file names as
// ui/dps_report/data/. For each file found there:
//   - a dataset with a `build` column keeps its other builds' rows and takes the new rows for the builds present
//     (builds.json uses `id` for this);
//   - a lookup table (icons, names, tooltips, talent trees) is merged by its key, new entries winning (a class in the
//     incoming talenttrees or treeart replaces all of that class's rows);
//   - anything else (race_summary, race_notes: summaries across every build) is replaced whole, so the pipeline must
//     compute those from the full, merged data.
// The dataset's `updated` time in ui/dps_report/datasets.json is set to now.
//
// The sample fights in <partial-dir>/events/ (events.py) replace the report's files of the same builds.
//
// A build with no request left (no tools/dps_report/req/<id>.meta.json: a removed preset) is pruned from every
// per-build dataset, from data/events/ and from ui/dps_report/provenance.json, so its old rows don't outlive it.
//
// Usage: node tools/dps_report/merge.mjs <partial-dir> [--by "<who>"]
import fs from 'fs';
import path from 'path';
import { fileURLToPath } from 'url';

const root = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '../..');
const dataDir = path.join(root, 'ui/dps_report/data');
const manifestPath = path.join(root, 'ui/dps_report/datasets.json');

const KEYS = {
	icons: ['icon'],
	icons_more: ['icon'],
	event_names: ['key'],
	event_icons: ['icon'],
	names: ['key'],
	tooltips: ['key', 'kind'],
	talenttrees: ['class', 'tree', 'idx'],
	treeart: ['class', 'tree'],
};

const [from, ...rest] = process.argv.slice(2);
if (!from) {
	console.error('Usage: node tools/dps_report/merge.mjs <partial-dir> [--by "<who>"]');
	process.exit(1);
}
const byIdx = rest.indexOf('--by');
const by = byIdx >= 0 ? rest[byIdx + 1] : 'Claude (forever-sim)';

const manifest = JSON.parse(fs.readFileSync(manifestPath, 'utf8'));
const readJson = p => JSON.parse(fs.readFileSync(p, 'utf8'));

for (const file of fs.readdirSync(from).filter(f => f.endsWith('.json'))) {
	const id = file.slice(0, -5);
	if (!(id in manifest)) {
		console.warn(`Skipping ${file}: not a dataset in datasets.json (add it there first)`);
		continue;
	}
	const incoming = readJson(path.join(from, file));
	const target = path.join(dataDir, file);
	const current = fs.existsSync(target) ? readJson(target) : [];
	let merged;
	let how;

	const buildCol = id === 'builds' ? 'id' : 'build';
	if (Array.isArray(incoming) && incoming.length && buildCol in incoming[0]) {
		const fresh = new Set(incoming.map(r => r[buildCol]));
		merged = [...current.filter(r => !fresh.has(r[buildCol])), ...incoming];
		how = `${fresh.size} build(s) replaced`;
	} else if (KEYS[id] && Array.isArray(incoming)) {
		const key = r => KEYS[id].map(k => r[k]).join('\u0000');
		// a class's talent trees come whole: drop its old rows, so a talent the tree lost doesn't linger
		const classes = new Set(id === 'talenttrees' || id === 'treeart' ? incoming.map(r => r.class) : []);
		const map = new Map(current.filter(r => !classes.has(r.class)).map(r => [key(r), r]));
		for (const r of incoming) map.set(key(r), r);
		merged = [...map.values()];
		how = `${incoming.length} entr${incoming.length === 1 ? 'y' : 'ies'} merged by ${KEYS[id].join('+')}`;
	} else {
		merged = incoming;
		how = 'replaced whole';
	}

	fs.writeFileSync(target, JSON.stringify(merged));
	manifest[id].updated = { at: new Date().toISOString(), by };
	console.log(`${id}: ${how} (${merged.length} rows)`);
}

const eventsFrom = path.join(from, 'events');
if (fs.existsSync(eventsFrom)) {
	const files = fs.readdirSync(eventsFrom).filter(f => f.endsWith('.json.gz'));
	fs.mkdirSync(path.join(dataDir, 'events'), { recursive: true });
	for (const file of files) fs.copyFileSync(path.join(eventsFrom, file), path.join(dataDir, 'events', file));
	console.log(`events: ${files.length} build(s) replaced`);
}

// prune the builds whose request is gone
const reqDir = path.join(root, 'tools/dps_report/req');
const live = new Set(
	fs
		.readdirSync(reqDir)
		.filter(f => f.endsWith('.meta.json'))
		.map(f => f.slice(0, -'.meta.json'.length)),
);
const removed = b => typeof b === 'string' && b.includes('__') && !live.has(b); // not pseudo rows such as notes' "_round"
for (const file of fs.readdirSync(dataDir).filter(f => f.endsWith('.json'))) {
	const id = file.slice(0, -5);
	const target = path.join(dataDir, file);
	const rows = readJson(target);
	const buildCol = id === 'builds' ? 'id' : 'build';
	if (!Array.isArray(rows) || !rows.length || !(buildCol in rows[0])) continue;
	const kept = rows.filter(r => !removed(r[buildCol]));
	if (kept.length === rows.length) continue;
	const gone = [...new Set(rows.filter(r => removed(r[buildCol])).map(r => r[buildCol]))];
	fs.writeFileSync(target, JSON.stringify(kept));
	manifest[id].updated = { at: new Date().toISOString(), by };
	console.log(`${id}: pruned ${rows.length - kept.length} row(s) of removed build(s) ${gone.join(', ')}`);
}
const eventsDir = path.join(dataDir, 'events');
if (fs.existsSync(eventsDir)) {
	for (const file of fs.readdirSync(eventsDir).filter(f => f.endsWith('.json.gz'))) {
		if (removed(file.slice(0, -'.json.gz'.length))) {
			fs.unlinkSync(path.join(eventsDir, file));
			console.log(`events: pruned ${file}`);
		}
	}
}
const provPath = path.join(root, 'ui/dps_report/provenance.json');
if (fs.existsSync(provPath)) {
	const prov = readJson(provPath);
	const gone = Object.keys(prov.builds || {}).filter(removed);
	if (gone.length) {
		for (const b of gone) delete prov.builds[b];
		fs.writeFileSync(provPath, JSON.stringify(prov, null, '\t') + '\n');
		console.log(`provenance: pruned ${gone.join(', ')}`);
	}
}

fs.writeFileSync(manifestPath, JSON.stringify(manifest, null, 1) + '\n');
