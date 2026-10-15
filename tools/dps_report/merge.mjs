// Merges a partial regeneration into the DPS Report's datasets.
//
// The pipeline can write datasets for only the builds it re-simmed, into a scratch folder with the same file names as
// ui/dps_report/data/. For each file found there:
//   - a dataset with a `build` column keeps its other builds' rows and takes the new rows for the builds present
//     (builds.json uses `id` for this);
//   - a lookup table (icons, names, tooltips, talent trees) is merged by its key, new entries winning;
//   - anything else (race_summary, race_notes: summaries across every build) is replaced whole, so the pipeline must
//     compute those from the full, merged data.
// The dataset's `updated` time in ui/dps_report/datasets.json is set to now.
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
		const map = new Map(current.map(r => [key(r), r]));
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

fs.writeFileSync(manifestPath, JSON.stringify(manifest, null, 1) + '\n');
