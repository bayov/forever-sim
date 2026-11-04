// Says which DPS Report builds are out of date, so a regeneration only re-sims what changed.
//
// Each build (a level 60 preset, id "<class>__<preset>") depends on:
//   - shared inputs every build uses: the sim's own packages (sim/core, sim/common, sim/encounters, sim/lib), the
//     protos the sim compiles (not ui.proto or test.proto, which nothing in the sim imports), the CLI and go.mod, the
//     report's local patches and dashstats, and the report's step scripts (their sims, iterations and post-processing);
//   - its class's sim package sim/<class>/;
//   - its own inputs: its request tools/dps_report/req/<id>.json (what is simmed), the fields of its meta file the
//     step scripts read (class, spec, race, baseBuild, duty, pet, encounter), its base build's request (a variant's race
//     overrides depend on it), its races' own rotations and gear tools/dps_report/req/races/<id>.<Race>.json (and its
//     base build's, race_override.py), and its assumption scenario rotations tools/dps_report/asm_apls/<id>.*
//     (run_assumptions.py);
//   - the item database entries it wears: the items, enchants and random suffixes in its request, from
//     assets/database/db.json (the sim reads nothing else of the database at run time: sim/core/database_load.go).
// Test files, golden results, testdata and docs inside those folders are left out of every hash: the sim binary never
// reads them. tools/rotopt and the class's ui/<spec> folders are not inputs either: they generate requests (mksettings),
// and the requests are hashed themselves.
//
// ui/dps_report/provenance.json records, per build, the commit its data is current at and the hash of each input there.
// A build is stale when any hash differs at the commit being checked, or when it has no record. A record from before
// this input set (no `format`, or an older one) is compared by computing this input set at its recorded commit (git
// fetch --shallow-since=<date> when that commit is missing from a shallow clone).
//
// Stale says "may have moved", not "moved": pipeline.sh probe (probe.py) re-runs each stale build's sims at a tenth of
// their iterations and compares them with the probe stored in its raw pack, and only the builds that differ are
// re-simmed. Builds it clears are recorded with --by probe.
//
// Usage:
//   node tools/dps_report/stale.mjs [--commit <rev>] [--json]       list stale builds (default rev: HEAD)
//   node tools/dps_report/stale.mjs --record <id...|all> [--commit <rev>] [--iterations N] [--by sim|probe|hand]
//                                                                   record builds as current at <rev>
import { execFileSync } from 'child_process';
import crypto from 'crypto';
import fs from 'fs';
import path from 'path';
import { fileURLToPath } from 'url';

const root = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '../..');
const dataDir = path.join(root, 'ui/dps_report/data');
const provPath = path.join(root, 'ui/dps_report/provenance.json');

export const FORMAT = 2;
const STEP_SCRIPTS = ['run_build', 'run_races', 'weights', 'mana_curves', 'ablate', 'run_assumptions', 'assumptions_def', 'race_override', 'dashpaths'].map(
	s => `tools/dps_report/${s}.py`,
);
export const SHARED_INPUTS = [
	'sim/core',
	'sim/common',
	'sim/encounters',
	'sim/lib',
	'proto',
	'cmd/wowsimcli',
	'go.mod',
	'go.sum',
	'tools/dps_report/patches',
	'tools/dps_report/dashstats',
	...STEP_SCRIPTS,
];
// files inside the inputs that the sim binary and the step scripts never read
const IGNORED = /(_test\.go|\.results|\.results\.tmp|\.md)$|\/testdata\/|^proto\/(ui|test)\.proto$/;
const META_FIELDS = ['class', 'spec', 'race', 'baseBuild', 'duty', 'pet', 'encounter'];

const args = process.argv.slice(2);
const opt = name => {
	const i = args.indexOf(name);
	return i >= 0 ? args[i + 1] : undefined;
};
const rev = opt('--commit') || 'HEAD';

const git = (...a) => execFileSync('git', a, { cwd: root, encoding: 'utf8', maxBuffer: 1 << 30 });
const sha = s => crypto.createHash('sha256').update(s).digest('hex');
const memo = f => {
	const m = new Map();
	return (...a) => {
		const k = a.join('\0');
		if (!m.has(k)) m.set(k, f(...a));
		return m.get(k);
	};
};
const hasCommit = memo(r => {
	try {
		git('cat-file', '-e', `${r}^{commit}`);
		return true;
	} catch {
		return false;
	}
});
const show = memo((r, p) => {
	try {
		return git('show', `${r}:${p}`);
	} catch {
		return null; // not present at this revision
	}
});
const showJson = (r, p) => {
	const s = show(r, p);
	return s === null ? null : JSON.parse(s);
};
const lsTree = memo((r, dir) =>
	git('ls-tree', '-r', r, '--', dir)
		.split('\n')
		.filter(Boolean)
		.map(l => ({ blob: l.split(/\s+/)[2], path: l.split('\t')[1] })),
);
// a folder's or file's hash with the ignored files left out; null when it doesn't exist at the revision
const inputHash = memo((r, p) => {
	const files = lsTree(r, p).filter(f => !IGNORED.test(f.path));
	return files.length ? sha(files.map(f => `${f.blob} ${f.path}`).join('\n')) : null;
});

// the database the sim loads, by kind and id
const db = memo(r => {
	const d = showJson(r, 'assets/database/db.json') || {};
	return {
		item: new Map((d.items || []).map(x => [x.id, x])),
		enchant: new Map((d.enchants || []).map(x => [x.effectId, x])),
		suffix: new Map((d.randomSuffixes || []).map(x => [x.id, x])),
	};
});
// the database entries a request wears: every equipment item, its enchant and its random suffix
const dbSlice = (r, req) => {
	const d = db(r);
	const rows = [];
	for (const party of req?.raid?.parties || [])
		for (const pl of party.players || [])
			for (const it of pl.equipment?.items || []) {
				if (it.id) rows.push(['item', it.id, d.item.get(it.id) ?? null]);
				if (it.enchant) rows.push(['enchant', it.enchant, d.enchant.get(it.enchant) ?? null]);
				if (it.randomSuffix) rows.push(['suffix', it.randomSuffix, d.suffix.get(it.randomSuffix) ?? null]);
			}
	rows.sort((a, b) => (a[0] + a[1] < b[0] + b[1] ? -1 : 1));
	return sha(JSON.stringify(rows));
};

const metaAt = (r, id) => showJson(r, `tools/dps_report/req/${id}.meta.json`);

// every input of a build at a revision, as {name: hash}
export const fingerprint = memo((r, id) => {
	const cls = id.split('__')[0];
	const fp = {};
	for (const p of [...SHARED_INPUTS, `sim/${cls}`]) fp[p] = inputHash(r, p);
	const reqPath = `tools/dps_report/req/${id}.json`;
	fp[reqPath] = inputHash(r, reqPath);
	const meta = metaAt(r, id);
	fp[`${reqPath.slice(0, -5)}.meta.json (sim fields)`] = meta ? sha(JSON.stringify(META_FIELDS.map(k => meta[k] ?? null))) : null;
	const base = meta?.baseBuild;
	if (base) fp[`tools/dps_report/req/${base}.json`] = inputHash(r, `tools/dps_report/req/${base}.json`);
	const ids = [id, base].filter(Boolean);
	for (const f of lsTree(r, 'tools/dps_report/req/races/'))
		if (ids.some(i => path.basename(f.path).startsWith(i + '.'))) fp[f.path] = f.blob;
	for (const f of lsTree(r, 'tools/dps_report/asm_apls/')) if (path.basename(f.path).startsWith(id + '.')) fp[f.path] = f.blob;
	fp['assets/database (the entries it wears)'] = dbSlice(r, showJson(r, reqPath));
	return fp;
});

const diff = (a, b) => [...new Set([...Object.keys(a), ...Object.keys(b)])].filter(k => (a[k] ?? null) !== (b[k] ?? null)).sort();

// every build in the report, and every request not in it yet (a new preset)
const published = JSON.parse(fs.readFileSync(path.join(dataDir, 'builds.json'), 'utf8')).map(b => b.id);
const requested = fs
	.readdirSync(path.join(root, 'tools/dps_report/req'))
	.filter(f => f.endsWith('.meta.json'))
	.map(f => f.slice(0, -'.meta.json'.length));
const builds = [...new Set([...published, ...requested])].filter(id => requested.includes(id)).sort();
const removed = published.filter(id => !requested.includes(id));
const prov = fs.existsSync(provPath) ? JSON.parse(fs.readFileSync(provPath, 'utf8')) : { builds: {} };

if (args.includes('--record')) {
	const i = args.indexOf('--record');
	let ids = [];
	for (const a of args.slice(i + 1)) {
		if (a.startsWith('--')) break;
		ids.push(a);
	}
	if (ids.includes('all')) ids = builds;
	const commit = git('rev-parse', rev).trim();
	const iterations = Number(opt('--iterations')) || undefined;
	const by = opt('--by') || 'sim';
	if (!['sim', 'probe', 'hand'].includes(by)) throw new Error(`--by ${by}: sim, probe or hand`);
	for (const id of ids) {
		if (!builds.includes(id)) throw new Error(`Unknown build ${id}`);
		const old = prov.builds[id];
		prov.builds[id] = {
			commit,
			recordedAt: new Date().toISOString(),
			// a probe or hand record keeps the sims it vouches for: their commit and iterations
			iterations: by === 'sim' ? iterations : old?.iterations,
			by,
			simCommit: by === 'sim' ? commit : old?.simCommit || old?.commit,
			format: FORMAT,
			inputs: fingerprint(commit, id),
		};
	}
	fs.writeFileSync(provPath, JSON.stringify(prov, null, '\t') + '\n');
	console.log(`Recorded ${ids.length} build(s) at ${commit.slice(0, 10)} (by ${by})`);
	process.exit(0);
}

const stale = [];
const missing = new Set();
for (const id of builds) {
	const rec = prov.builds[id];
	if (!rec) {
		stale.push({ id, reasons: ['no provenance record'] });
		continue;
	}
	let then = rec.inputs;
	if (rec.format !== FORMAT) {
		// an older record: this input set at the commit it was recorded at
		if (!hasCommit(rec.commit)) {
			missing.add(rec.commit);
			stale.push({ id, reasons: [`recorded at ${rec.commit.slice(0, 10)}, which this clone doesn't have`] });
			continue;
		}
		then = fingerprint(rec.commit, id);
	}
	const changed = diff(then, fingerprint(rev, id));
	if (changed.length) stale.push({ id, reasons: changed.map(p => `${p} changed`) });
}

if (args.includes('--json')) {
	console.log(JSON.stringify({ rev, total: builds.length, stale, removed }, null, 2));
} else {
	console.log(`${stale.length} of ${builds.length} builds are stale at ${rev}.`);
	for (const s of stale) console.log(`  ${s.id}: ${s.reasons.join(', ')}`);
	if (removed.length) console.log(`In the report with no request (removed presets; pipeline.sh publish drops them): ${removed.join(' ')}`);
	if (missing.size)
		console.log(`Fetch the recorded commits to compare them (git fetch --shallow-since=<a date before them> origin master): ${[...missing].join(' ')}`);
}
